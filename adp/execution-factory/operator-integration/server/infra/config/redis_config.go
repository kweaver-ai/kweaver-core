package config

import (
	"context"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/redis/go-redis/v9"
)

const (
	// RedisTypeSentinel redis哨兵模式
	RedisTypeSentinel = "sentinel"
	// RedisTypeMasterSlave redis主从模式
	RedisTypeMasterSlave = "master-slave"
	// RedisTypeStandalone redis单机模式
	RedisTypeStandalone = "standalone"
	commaSep            = ","
)

// RedisConfig redis配置
type RedisConfig struct {
	ConnectType string           `yaml:"connectType"` // sentinel/master-slave/standalone 对应哨兵、主从、单机三种连接方式
	ConnectInfo RedisConnectInfo `yaml:"connectInfo"`
	EnableSSL   bool             `yaml:"enableSSL"`
	SecretName  string           `yaml:"secretName"` // 当 enableSSL 为 true 需要
	CaName      string           `yaml:"caName"`     // 当 enableSSL 为 true 需要，表示secret里 ca 证书的名字
	CertName    string           `yaml:"certName"`   // 当 enableSSL 为 true 需要，表示secret里 cert 证书的名字
	KeyName     string           `yaml:"keyName"`    // 当 enableSSL 为 true 需要，表示secret里 key 密钥的名字
}

// RedisConnectInfo redis连接配置
type RedisConnectInfo struct {
	Username         string `yaml:"username"`
	Password         string `yaml:"password"`
	Host             string `yaml:"host"`
	Port             int    `yaml:"port"`
	MasterlHost      string `yaml:"masterHost"`
	MasterlPort      int    `yaml:"masterPort"`
	SlavelHost       string `yaml:"slaveHost"`
	SlavelPort       int    `yaml:"slavePort"`
	SentinelHost     string `yaml:"sentinelHost"`
	SentinelPort     int    `yaml:"sentinelPort"`
	SentinelUsername string `yaml:"sentinelUsername"`
	SentinelPassword string `yaml:"sentinelPassword"`
	MasterGroupName  string `yaml:"masterGroupName"`
	PoolSize         int    `yaml:"poolSize" default:"10"` // 连接池大小
}

var (
	globalCli, globalReadCli *redis.Client
	redisOnce                sync.Once
)

func getCertPool(caName string) (*x509.CertPool, error) {
	pool := x509.NewCertPool()
	cas := strings.Split(caName, commaSep)
	for _, ca := range cas {
		caCrt, err := os.ReadFile(strings.TrimSpace(ca))
		if err != nil {
			return nil, err
		}
		if !pool.AppendCertsFromPEM(caCrt) {
			err = errors.New("add ca to pool err")
			return nil, err
		}
	}
	return pool, nil
}

// DERToPrivateKey der 转为 私钥
func DERToPrivateKey(der []byte) (key interface{}, err error) {
	if key, err = x509.ParsePKCS1PrivateKey(der); err == nil {
		return key, nil
	}

	if key, err = x509.ParsePKCS8PrivateKey(der); err == nil {
		switch key.(type) {
		case *rsa.PrivateKey, *ecdsa.PrivateKey:
			return
		default:
			return nil, errors.New("found unknown private key type in PKCS#8 wrapping")
		}
	}

	if key, err = x509.ParseECPrivateKey(der); err == nil {
		return
	}

	return nil, errors.New("invalid key type. The DER must contain an rsa.PrivateKey or ecdsa.PrivateKey")
}

// DecryptPEM 带密码per解密
func DecryptPEM(pemRaw, passwd []byte) (pemDer []byte, err error) {
	block, _ := pem.Decode(pemRaw)
	if block == nil {
		return nil, fmt.Errorf("failed decoding PEM. Block must be different from nil. [% x]", pemRaw)
	}

	if !x509.IsEncryptedPEMBlock(block) { //nolint:staticcheck
		return nil, fmt.Errorf("failed decryptPEM PEM. it's not a decryped PEM [%s]", pemRaw)
	}

	der, err := x509.DecryptPEMBlock(block, passwd) //nolint:staticcheck
	if err != nil {
		return nil, fmt.Errorf("failed PEM decryption [%s]", err)
	}

	privateKey, err := DERToPrivateKey(der)
	if err != nil {
		return nil, err
	}

	var raw []byte
	switch k := privateKey.(type) {
	case *ecdsa.PrivateKey:
		raw, err = x509.MarshalECPrivateKey(k)
		if err != nil {
			return
		}
	case *rsa.PrivateKey:
		raw = x509.MarshalPKCS1PrivateKey(k)
	default:
		err = errors.New("invalid key type. It must be *ecdsa.PrivateKey or *rsa.PrivateKey")
		return
	}

	rawBase64 := base64.StdEncoding.EncodeToString(raw)
	derBase64 := base64.StdEncoding.EncodeToString(der)
	if rawBase64 != derBase64 {
		err = errors.New("invalid decrypt PEM: raw does not match with der")
		return
	}

	block = &pem.Block{
		Type:  block.Type,
		Bytes: der,
	}

	pemDer = pem.EncodeToMemory(block)
	return
}

func (conf *RedisConfig) getTLSConfig() (tlsConf *tls.Config, err error) {
	if !conf.EnableSSL {
		return
	}
	var cer tls.Certificate
	if conf.SecretName == "" {
		cer, err = tls.LoadX509KeyPair(conf.CertName, conf.KeyName)
	} else {
		var certBlock, keyBlock []byte
		certBlock, err = os.ReadFile(conf.CertName)
		if err != nil {
			return
		}
		keyBlock, err = os.ReadFile(conf.KeyName)
		if err != nil {
			return
		}
		keyBlock, err = DecryptPEM(keyBlock, []byte(conf.SecretName))
		if err != nil {
			return
		}
		cer, err = tls.X509KeyPair(certBlock, keyBlock)
	}
	if err != nil {
		return
	}
	tlsConf = &tls.Config{
		Certificates: []tls.Certificate{cer},
	}
	if conf.CaName != "" {
		tlsConf.RootCAs, err = getCertPool(conf.CaName)
	}
	return
}

// GetClient 获取Redis客户端
func (conf *RedisConfig) GetClient() (cli, readCli *redis.Client, err error) {
	redisOnce.Do(func() {
		globalCli, globalReadCli, err = conf.getClient()
	})
	return globalCli, globalReadCli, err
}
func (conf *RedisConfig) getClient() (cli, readCli *redis.Client, err error) {
	tlsConf, err := conf.getTLSConfig()
	if err != nil {
		return
	}
	switch conf.ConnectType {
	case RedisTypeSentinel:
		cli = redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:       conf.ConnectInfo.MasterGroupName,
			SentinelAddrs:    []string{fmt.Sprintf("%s:%d", conf.ConnectInfo.SentinelHost, conf.ConnectInfo.SentinelPort)},
			Username:         conf.ConnectInfo.Username,
			Password:         conf.ConnectInfo.Password,
			SentinelUsername: conf.ConnectInfo.SentinelUsername,
			SentinelPassword: conf.ConnectInfo.SentinelPassword,
			TLSConfig:        tlsConf,
			PoolSize:         conf.ConnectInfo.PoolSize,
		})
	case RedisTypeMasterSlave:
		cli = redis.NewClient(&redis.Options{
			Addr:      fmt.Sprintf("%s:%d", conf.ConnectInfo.MasterlHost, conf.ConnectInfo.MasterlPort),
			Username:  conf.ConnectInfo.Username,
			Password:  conf.ConnectInfo.Password,
			TLSConfig: tlsConf,
			PoolSize:  conf.ConnectInfo.PoolSize,
		})
		if conf.ConnectInfo.SlavelHost != "" {
			readCli = redis.NewClient(&redis.Options{
				Addr:      fmt.Sprintf("%s:%d", conf.ConnectInfo.SlavelHost, conf.ConnectInfo.SlavelPort),
				Username:  conf.ConnectInfo.Username,
				Password:  conf.ConnectInfo.Password,
				TLSConfig: tlsConf,
				PoolSize:  conf.ConnectInfo.PoolSize,
			})
		}
	case RedisTypeStandalone:
		cli = redis.NewClient(&redis.Options{
			Addr:      fmt.Sprintf("%s:%d", conf.ConnectInfo.Host, conf.ConnectInfo.Port),
			Username:  conf.ConnectInfo.Username,
			Password:  conf.ConnectInfo.Password,
			TLSConfig: tlsConf,
			PoolSize:  conf.ConnectInfo.PoolSize,
		})
	default:
		err = fmt.Errorf("redis connect type shouldbe one of %s, %s, %s", RedisTypeSentinel, RedisTypeMasterSlave, RedisTypeStandalone)
		return
	}

	ctx := context.Background()
	_, err = cli.Ping(ctx).Result()
	if err != nil {
		return
	}
	if readCli != nil {
		_, err = readCli.Ping(ctx).Result()
	}
	return
}

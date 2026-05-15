// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package common

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/fsnotify/fsnotify"
	libdb "github.com/kweaver-ai/kweaver-go-lib/db"
	"github.com/kweaver-ai/kweaver-go-lib/hydra"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	libmq "github.com/kweaver-ai/kweaver-go-lib/mq"
	"github.com/kweaver-ai/kweaver-go-lib/otel"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/spf13/viper"

	"vega-backend/version"
)

// ServerSetting server配置项
type ServerSetting struct {
	RunMode      string        `mapstructure:"runMode"`
	HttpPort     int           `mapstructure:"httpPort"`
	Language     string        `mapstructure:"language"`
	ReadTimeOut  time.Duration `mapstructure:"readTimeOut"`
	WriteTimeout time.Duration `mapstructure:"writeTimeOut"`
}

// CryptoSetting RSA 密钥配置项
type CryptoSetting struct {
	Enabled        bool   `mapstructure:"enabled"`
	PrivateKey     string `mapstructure:"-"`              // RSA 私钥 (PEM 格式) - 从文件读取
	PublicKey      string `mapstructure:"-"`              // RSA 公钥 (PEM 格式) - 从文件读取
	PrivateKeyPath string `mapstructure:"privateKeyPath"` // RSA 私钥文件路径
	PublicKeyPath  string `mapstructure:"publicKeyPath"`  // RSA 公钥文件路径
}

// RedisSetting Redis 配置项
type RedisSetting struct {
	ConnectType      string
	Username         string
	Password         string
	Host             string
	Port             int
	MasterHost       string
	MasterPort       int
	SlaveHost        string
	SlavePort        int
	SentinelHost     string
	SentinelPort     int
	SentinelUsername string
	SentinelPassword string
	MasterGroupName  string
}

// KafkaConnectSetting Kafka Connect 配置项
type KafkaConnectSetting struct {
	Host     string
	Port     int
	Protocol string
}

// RateLimitingConfig rate limiting 配置项
type RateLimitingConfig struct {
	Concurrency ConcurrencyConfig `mapstructure:"concurrency"`
}

// ConcurrencyConfig 并发控制配置项
type ConcurrencyConfig struct {
	Enabled bool                    `mapstructure:"enabled"`
	Global  GlobalConcurrencyConfig `mapstructure:"global"`
}

// GlobalConcurrencyConfig 全局并发配置
type GlobalConcurrencyConfig struct {
	MaxConcurrentQueries int `mapstructure:"max_concurrent_queries"`
}

// AppSetting app配置项
type AppSetting struct {
	ServerSetting       ServerSetting             `mapstructure:"server"`
	LogSetting          logger.LogSetting         `mapstructure:"log"`
	OtelSetting         otel.OtelConfig           `mapstructure:"otel"`
	CryptoSetting       CryptoSetting             `mapstructure:"crypto"`
	DepServices         map[string]map[string]any `mapstructure:"depServices"`
	RateLimitingSetting RateLimitingConfig        `mapstructure:"rateLimiting"`

	DBSetting           libdb.DBSetting
	MQSetting           libmq.MQSetting
	OpenSearchSetting   rest.OpenSearchClientConfig
	HydraAdminSetting   hydra.HydraAdminSetting
	RedisSetting        RedisSetting
	KafkaConnectSetting KafkaConnectSetting

	PermissionUrl     string
	UserMgmtUrl       string
	MfModelManagerUrl string
	MfModelApiUrl     string
}

const (
	// ConfigFile 配置文件信息
	configPath string = "./config/"
	configName string = "vega-backend-config"
	configType string = "yaml"

	rdsServiceName            string = "rds"
	mqServiceName             string = "mq"
	opensearchServiceName     string = "opensearch"
	redisServiceName          string = "redis"
	permissionServiceName     string = "authorization-private"
	userMgmtServiceName       string = "user-management"
	hydraAdminServiceName     string = "hydra-admin"
	kafkaConnectServiceName   string = "kafka-connect"
	mfModelManagerServiceName string = "mf-model-manager"
	mfModelApiServiceName     string = "mf-model-api"

	DATA_BASE_NAME string = "kweaver"
)

var (
	appSetting *AppSetting
	vp         *viper.Viper

	settingOnce sync.Once

	// 当前系统时区
	APP_LOCATION *time.Location
)

// NewSetting 读取服务配置
func NewSetting() *AppSetting {
	settingOnce.Do(func() {
		appSetting = &AppSetting{}
		vp = viper.New()
		initSetting(vp)
	})

	return appSetting
}

// 初始化配置
func initSetting(vp *viper.Viper) {
	logger.Infof("Init Setting From File %s%s.%s", configPath, configName, configType)

	vp.AddConfigPath(configPath)
	vp.SetConfigName(configName)
	vp.SetConfigType(configType)

	loadSetting(vp)

	vp.WatchConfig()
	vp.OnConfigChange(func(e fsnotify.Event) {
		logger.Infof("Config file changed:%s", e)
		loadSetting(vp)
	})
}

// 读取配置文件
func loadSetting(vp *viper.Viper) {
	logger.Infof("Load Setting File %s%s.%s", configPath, configName, configType)

	if err := vp.ReadInConfig(); err != nil {
		logger.Fatalf("err:%s\n", err)
	}

	if err := vp.Unmarshal(appSetting); err != nil {
		logger.Fatalf("err:%s\n", err)
	}

	// 联调/CI：允许用环境变量覆盖监听端口，避免与本地已占用端口冲突
	if hp := strings.TrimSpace(os.Getenv("VEGA_HTTP_PORT")); hp != "" {
		if v, err := strconv.Atoi(hp); err == nil && v > 0 && v < 65536 {
			appSetting.ServerSetting.HttpPort = v
			logger.Infof("HttpPort overridden by VEGA_HTTP_PORT=%d", v)
		}
	}

	// 联调脚本（如 issue382）：无挂载密钥时关闭 crypto，避免读取 /opt/... 失败
	if v := strings.TrimSpace(os.Getenv("VEGA_CRYPTO_DISABLED")); strings.EqualFold(v, "1") || strings.EqualFold(v, "true") {
		appSetting.CryptoSetting.Enabled = false
		logger.Info("Crypto disabled via VEGA_CRYPTO_DISABLED env")
	}

	// 加载时区
	loc, err := time.LoadLocation(os.Getenv("TZ"))
	if err != nil {
		loc = time.Local
		logger.Warnf("WARNING: Failed to load timezone from env, using Local[%v] as default. Error: %v\n", time.Local, err)
	}
	APP_LOCATION = loc

	if err := loadCryptoKeys(); err != nil {
		logger.Fatalf("Failed to load crypto keys: %s\n", err)
	}

	SetLogSetting(appSetting.LogSetting)

	SetDBSetting()
	overrideDBSettingFromEnv()

	SetMQSetting()

	SetOpenSearchSetting()

	SetRedisSetting()

	SetKafkaConnectSetting()

	SetHydraAdminSetting()

	SetPermissionSetting()

	SetUserMgmtSetting()

	SetMfModelManagerSetting()
	SetMfModelApiSetting()

	appSetting.OtelSetting.ServiceName = version.ServerName
	appSetting.OtelSetting.ServiceVersion = version.ServerVersion
	logger.Infof("ServerName: %s, ServerVersion: %s, Language: %s, GoVersion: %s, GoArch: %s",
		version.ServerName, version.ServerVersion, version.LanguageGo,
		version.GoVersion, version.GoArch)

	s, _ := sonic.MarshalString(appSetting)
	logger.Debug(s)
}

func SetDBSetting() {
	setting, ok := appSetting.DepServices[rdsServiceName]
	if !ok {
		logger.Fatalf("service %s not found in depServices", rdsServiceName)
	}

	appSetting.DBSetting = libdb.DBSetting{
		Host:     setting["host"].(string),
		Port:     setting["port"].(int),
		Username: setting["user"].(string),
		Password: setting["password"].(string),
		DBName:   DATA_BASE_NAME,
	}
}

// overrideDBSettingFromEnv 联调/脚本用：覆盖 depServices 解析出的 DB 连接（如本地 127.0.0.1:3306）
func overrideDBSettingFromEnv() {
	if h := strings.TrimSpace(os.Getenv("VEGA_DB_HOST")); h != "" {
		appSetting.DBSetting.Host = h
		logger.Infof("DB Host overridden by VEGA_DB_HOST=%s", h)
	}
	if p := strings.TrimSpace(os.Getenv("VEGA_DB_PORT")); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 && v < 65536 {
			appSetting.DBSetting.Port = v
			logger.Infof("DB Port overridden by VEGA_DB_PORT=%d", v)
		}
	}
	if u := strings.TrimSpace(os.Getenv("VEGA_DB_USER")); u != "" {
		appSetting.DBSetting.Username = u
		logger.Infof("DB Username overridden by VEGA_DB_USER")
	}
	if pw, ok := os.LookupEnv("VEGA_DB_PASSWORD"); ok {
		appSetting.DBSetting.Password = pw
		logger.Info("DB Password overridden by VEGA_DB_PASSWORD")
	}
	if db := strings.TrimSpace(os.Getenv("VEGA_DB_NAME")); db != "" {
		appSetting.DBSetting.DBName = db
		logger.Infof("DB Name overridden by VEGA_DB_NAME=%s", db)
	}
}

func SetMQSetting() {
	setting, ok := appSetting.DepServices[mqServiceName]
	if !ok {
		logger.Fatalf("service %s not found in depServices", mqServiceName)
	}
	authSetting, ok := setting["auth"].(map[string]any)
	if !ok {
		logger.Fatalf("service %s auth not found in depServices", mqServiceName)
	}

	appSetting.MQSetting = libmq.MQSetting{
		MQType: setting["mqtype"].(string),
		MQHost: setting["mqhost"].(string),
		MQPort: setting["mqport"].(int),
		Tenant: setting["tenant"].(string),
		Auth: libmq.MQAuthSetting{
			Username:  authSetting["username"].(string),
			Password:  authSetting["password"].(string),
			Mechanism: authSetting["mechanism"].(string),
		},
	}
}

func SetOpenSearchSetting() {
	setting, ok := appSetting.DepServices[opensearchServiceName]
	if !ok {
		logger.Fatalf("service %s not found in depServices", opensearchServiceName)
	}

	appSetting.OpenSearchSetting = rest.OpenSearchClientConfig{
		Host:     setting["host"].(string),
		Port:     setting["port"].(int),
		Protocol: setting["protocol"].(string),
		Username: setting["user"].(string),
		Password: setting["password"].(string),
	}
}

// GetAuthEnabled 获取认证开关状态
// 通过环境变量 AUTH_ENABLED 控制，默认 true（安全优先）
func GetAuthEnabled() bool {
	envVal := os.Getenv("AUTH_ENABLED")
	// 仅当显式设置为 false 或 0 时禁用认证
	return envVal != "false" && envVal != "0"
}

func SetHydraAdminSetting() {
	if !GetAuthEnabled() {
		logger.Info("ISF authentication disabled via AUTH_ENABLED env, skipping hydra-admin configuration")
		return
	}
	setting, ok := appSetting.DepServices[hydraAdminServiceName]
	if !ok {
		logger.Fatalf("service %s not found in depServices", hydraAdminServiceName)
	}
	appSetting.HydraAdminSetting = hydra.HydraAdminSetting{
		HydraAdminProcotol: setting["protocol"].(string),
		HydraAdminHost:     setting["host"].(string),
		HydraAdminPort:     setting["port"].(int),
	}
}

func SetPermissionSetting() {
	if !GetAuthEnabled() {
		logger.Info("ISF authentication disabled via AUTH_ENABLED env, skipping authorization configuration")
		return
	}
	setting, ok := appSetting.DepServices[permissionServiceName]
	if !ok {
		logger.Fatalf("service %s not found in depServices", permissionServiceName)
	}

	protocol := setting["protocol"].(string)
	host := setting["host"].(string)
	port := setting["port"].(int)

	appSetting.PermissionUrl = fmt.Sprintf("%s://%s:%d/api/authorization/v1", protocol, host, port)
}

func SetUserMgmtSetting() {
	if !GetAuthEnabled() {
		logger.Info("ISF authentication disabled via AUTH_ENABLED env, skipping user-management configuration")
		return
	}
	setting, ok := appSetting.DepServices[userMgmtServiceName]
	if !ok {
		logger.Fatalf("service %s not found in depServices", userMgmtServiceName)
	}

	protocol := setting["protocol"].(string)
	host := setting["host"].(string)
	port := setting["port"].(int)

	appSetting.UserMgmtUrl = fmt.Sprintf("%s://%s:%d", protocol, host, port)
}

// SetRedisSetting 设置 Redis 配置
func SetRedisSetting() {
	setting, ok := appSetting.DepServices[redisServiceName]
	if !ok {
		logger.Fatalf("service %s not found in depServices", redisServiceName)
	}

	connectInfo, ok := setting["connectinfo"].(map[string]any)
	if !ok {
		logger.Fatalf("service %s connectInfo not found in depServices", redisServiceName)
	}

	appSetting.RedisSetting = RedisSetting{
		ConnectType:      setting["connecttype"].(string),
		Username:         connectInfo["username"].(string),
		Password:         connectInfo["password"].(string),
		SentinelHost:     connectInfo["sentinelhost"].(string),
		SentinelPort:     connectInfo["sentinelport"].(int),
		SentinelUsername: connectInfo["sentinelusername"].(string),
		SentinelPassword: connectInfo["sentinelpassword"].(string),
		MasterGroupName:  connectInfo["mastergroupname"].(string),

		Host:       connectInfo["host"].(string),
		Port:       connectInfo["port"].(int),
		MasterHost: connectInfo["masterhost"].(string),
		MasterPort: connectInfo["masterport"].(int),
		SlaveHost:  connectInfo["slavehost"].(string),
		SlavePort:  connectInfo["slaveport"].(int),
	}
}

// loadCryptoKeys 从文件加载 RSA 密钥
func loadCryptoKeys() error {
	if !appSetting.CryptoSetting.Enabled {
		return nil
	}

	if appSetting.CryptoSetting.PrivateKeyPath == "" {
		return fmt.Errorf("privateKeyPath is required when crypto is enabled")
	}
	if appSetting.CryptoSetting.PublicKeyPath == "" {
		return fmt.Errorf("publicKeyPath is required when crypto is enabled")
	}

	privateKeyContent, err := os.ReadFile(appSetting.CryptoSetting.PrivateKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read private key file: %w", err)
	}
	appSetting.CryptoSetting.PrivateKey = string(privateKeyContent)

	publicKeyContent, err := os.ReadFile(appSetting.CryptoSetting.PublicKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read public key file: %w", err)
	}
	appSetting.CryptoSetting.PublicKey = string(publicKeyContent)

	return nil
}

func SetKafkaConnectSetting() {
	setting, ok := appSetting.DepServices[kafkaConnectServiceName]
	if !ok {
		logger.Fatalf("service %s not found in depServices", kafkaConnectServiceName)
	}

	appSetting.KafkaConnectSetting = KafkaConnectSetting{
		Host:     setting["host"].(string),
		Port:     setting["port"].(int),
		Protocol: setting["protocol"].(string),
	}
}

func SetMfModelManagerSetting() {
	setting, ok := appSetting.DepServices[mfModelManagerServiceName]
	if !ok {
		logger.Fatalf("service %s not found in depServices", mfModelManagerServiceName)
	}

	protocol := setting["protocol"].(string)
	host := setting["host"].(string)
	port := setting["port"].(int)

	appSetting.MfModelManagerUrl = fmt.Sprintf("%s://%s:%d", protocol, host, port)
}

func SetMfModelApiSetting() {
	setting, ok := appSetting.DepServices[mfModelApiServiceName]
	if !ok {
		logger.Fatalf("service %s not found in depServices", mfModelApiServiceName)
	}

	protocol := setting["protocol"].(string)
	host := setting["host"].(string)
	port := setting["port"].(int)

	appSetting.MfModelApiUrl = fmt.Sprintf("%s://%s:%d", protocol, host, port)
}

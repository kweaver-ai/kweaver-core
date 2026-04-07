package anysharedshandler

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sync"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driver/ihandlerportdriver"
	"github.com/kweaver-ai/kweaver-go-lib/logger"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

const RSA_PRIVATE_KEY = `-----BEGIN PRIVATE KEY-----
MIIEwAIBADANBgkqhkiG9w0BAQEFAASCBKowggSmAgEAAoIBAQDbYY5JDWN4OGl3
PGEl11J/5TXQXgi+63uCJiFyEUAgmBGicxqYNPzoxKpBRzwN1MV08YubszcZQpyw
bWpsLdKqThcn02OZBduzsgxbJYRZCs3si/tLZCMNgkj6mm8g3TcjQqbdYqyaeuV7
ZsvNP9Ubx7vZhfr119cB6Jfq4OGG3W9r8nFlUUwhXekdroBBD2CflIBRNfZqQSFF
97Vs8mN9XKgz51822+B19qtqpXWs6FUMSp3Q2U5a3ZEfARNDBTyAgmJvZtqPYy3I
gBj59DAqUpbFSzUZwVsQKeFU+JU3oESFcPs6FG31AT9y5iXm/hrrKm6AZowErpUo
6oUcO9+bAgMBAAECggEBAMXsiwlfeemBw60enWsdi8H1koqN/Af7vi9apXwbEicV
63sLq+e8jpyWqiBA226DEy6BqfnsQ36XuXP3EzfMU67wyzVUIxxwy5mgvkMRYwlO
lSCf3jVTf8h1TdBCupYE3vUB8jf0CVNKI3Yk9SQVPfhVSCZlGVjpxYJkTYNMJkyc
GMYAdZFCEV43mIm+ev4GaepR+d/syeXL/SZfFa0uEy8SFChrehRDhdVVkn+dRzeg
O6tbDkTFYtOpi+UI5obcGsVXEN3ZAZzaOKrB2TPwU1Ei5sIcWZhvKEfJkpiKIdpe
eLztYSaRB6gjCqhYQ3wzaJQCnoNVz+XqVaRTcPdZfBECgYEA+34Xo43WjhSdWR/P
laqleXfcwCmsF4Za+2qZjXLW//D2SXQylRv6hMAcVg7qCJM5a95X5VTr5H7pQNHN
ungE5Oi9lvYlZYb+pmG2wRn+/ufBs6OjwR6aDw/bsqeDHVjPeFIrxFPeXNllEHe1
xtZuhXvxFjDXqIwzQa2WijT8hjMCgYEA31Ag3lj9bAF7dBTJn8yRPhZX4v3I5N3x
H7G5XVj75cwMk4RB1s4WN/uLsuVDzXmG7NXjZ2c6kMYk658TTPKznQwhDx3Jq7Mh
HSJklWDtcPOioFZzFkikfqHseAWGf9s/HxBgieLa3IuGR9hEJ4EjDHa43UDXGQD2
90QGX7qlSPkCgYEAh9FL8N8LzQVjCJu+XqSe4t+RjxGyR64eeoLSVGp9pBE84ORo
4NAQVhrt8qfxShpAO3oDW+2ly2uiiogDo71nXzw2D031WkQySCajLNveM0lz+ZDZ
QdVF+/ZjfrMqgvHQcblmu4tTni8lfmQ3/h8V5u7Nf193SCYXFFQr5Y3CBrMCgYEA
wBgWXg3g2WKhBqbHFd4L5oOj0FAM2ssMGv5vfJwJ+4++FbtEQ3n95ORONHJBE+SB
KwOGXTGQUG8R3Vl2ac+wr9x6J52xGDC7wGsQaOr69RmvAAu9biLI1WGGn2vpWdyI
fLlCwfnR2LtwpCal4fGU66jItxKKtSh+SQ9MCFbuzUkCgYEAvUZaQDKmjdSbtR7J
yRXWfXPf0DpUqYKDzP40VoPcoQVGBmZAmq92yl1DMqFBfYCueCv1aA7Ozt+RFgyV
bMdUcJ0qzhKdCnEaonlpJPlnZkfATj5vOLs+nwfsmyO0iwcjA2zjJHmBZM+Xg+tl
enZgox36xuiZZrGQd0jXRt134QM=
-----END PRIVATE KEY-----`

type anysharedsHandler struct {
	logger icmp.Logger
}

func (h *anysharedsHandler) RegPubRouter(router *gin.RouterGroup) {
	router.POST("/anyshare7ds/getinfobypath", h.getInfoByPath)
	router.POST("/anyshare7ds/dir/list", h.dirList)
}

func (a *anysharedsHandler) RegPriRouter(router *gin.RouterGroup) {
	// 私有路由注册
}

var (
	handlerOnce sync.Once
	_handler    ihandlerportdriver.IHTTPRouter
)

func NewAnysharedsHandler() ihandlerportdriver.IHTTPRouter {
	handlerOnce.Do(func() {
		_handler = &anysharedsHandler{
			logger: logger.GetLogger(),
		}
	})

	return _handler
}

func getAnyshareOauth2Token(_ context.Context, targetURL string, username string, passwordRSABase64 string) (string, error) {
	passwordRSA, err := base64.StdEncoding.DecodeString(passwordRSABase64)
	if err != nil {
		return "", errors.Wrapf(err, "Base64解码失败")
	}

	// 创建请求数据（form）
	formData := url.Values{}
	formData.Add("grant_type", "client_credentials")
	formData.Add("scope", "all")
	// 创建 HTTP 请求
	req, err := http.NewRequest("POST", targetURL, bytes.NewBuffer([]byte(formData.Encode())))
	if err != nil {
		return "", errors.Wrapf(err, "创建请求失败")
	}

	// 设置 Basic Auth 认证头
	//  解析私钥
	privateKey, err := parseRSAPrivateKey([]byte(RSA_PRIVATE_KEY))
	if err != nil {
		return "", errors.Wrapf(err, "解析私钥失败")
	}
	//  选择填充方式解密
	// 方式1：PKCS1 v1.5（兼容大部分场景）
	decryptedPassword, err := rsaPrivateDecryptPKCS1(privateKey, []byte(passwordRSA))
	if err != nil {
		return "", errors.Wrapf(err, "RSA私钥解密失败")
	}

	// 方式2：OAEP（更安全，需加密端也用OAEP）
	// decryptedPassword, err := rsaPrivateDecryptOAEP(privateKey, []byte(passwordRSA))
	// if err != nil {
	// 	return "", errors.Wrapf(err, "OAEP解密失败")
	// }
	req.SetBasicAuth(username, string(decryptedPassword))

	// 设置内容类型
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// 发送请求
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("请求发送失败:", err)
	}

	defer resp.Body.Close()

	// 处理响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrapf(err, "读取响应失败")
	}

	type response struct {
		AccessToken string `json:"access_token"`
	}

	var rt response

	err = json.Unmarshal(respBody, &rt)
	if err != nil {
		return "", errors.Wrapf(err, "解析响应失败")
	}

	return rt.AccessToken, nil
}

// 解析PEM格式的RSA私钥
func parseRSAPrivateKey(privateKeyPEM []byte) (*rsa.PrivateKey, error) {
	// 解码PEM块
	block, _ := pem.Decode(privateKeyPEM)
	if block == nil {
		return nil, errors.New("解析PEM私钥失败：无效的PEM格式")
	}

	// 解析PKCS1格式私钥（传统RSA私钥）
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		// 尝试解析PKCS8格式私钥（通用格式，如Java生成的私钥）
		pkcs8Key, err2 := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err2 != nil {
			return nil, fmt.Errorf("解析私钥失败：PKCS1错误=%v, PKCS8错误=%v", err, err2)
		}
		// 类型断言为RSA私钥
		rsaKey, ok := pkcs8Key.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("私钥不是RSA类型")
		}

		privateKey = rsaKey
	}

	return privateKey, nil
}

// RSA私钥解密（PKCS1 v1.5填充）
func rsaPrivateDecryptPKCS1(privateKey *rsa.PrivateKey, encryptedData []byte) ([]byte, error) {
	// 解密：PKCS1 v1.5填充
	decrypted, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, encryptedData)
	if err != nil {
		return nil, fmt.Errorf("PKCS1解密失败：%v", err)
	}

	return decrypted, nil
}

// RSA私钥解密（OAEP填充，更安全）
func rsaPrivateDecryptOAEP(privateKey *rsa.PrivateKey, encryptedData []byte) ([]byte, error) {
	// OAEP填充需要指定哈希算法（常用SHA256）和随机数源
	hash := crypto.SHA256.New()

	decrypted, err := rsa.DecryptOAEP(hash, rand.Reader, privateKey, encryptedData, nil)
	if err != nil {
		return nil, fmt.Errorf("OAEP解密失败：%v", err)
	}

	return decrypted, nil
}

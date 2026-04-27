package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"

	"github.com/pkg/errors"
)

// RSA_PRIVATE_KEY RSA私钥用于密码解密
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

// DecryptPassword 解密密码（先base64解码，再RSA解密）
func DecryptPassword(passwordRSABase64 string) (string, error) {
	// 先进行base64解码
	passwordRSA, err := base64.StdEncoding.DecodeString(passwordRSABase64)
	if err != nil {
		return "", errors.Wrapf(err, "Base64解码失败")
	}

	// 解析私钥
	privateKey, err := ParseRSAPrivateKey([]byte(RSA_PRIVATE_KEY))
	if err != nil {
		return "", errors.Wrapf(err, "解析私钥失败")
	}

	// RSA私钥解密（PKCS1 v1.5填充）
	decryptedPassword, err := RSAPrivateDecryptPKCS1(privateKey, passwordRSA)
	if err != nil {
		return "", errors.Wrapf(err, "RSA私钥解密失败")
	}

	return string(decryptedPassword), nil
}

// ParseRSAPrivateKey 解析PEM格式的RSA私钥
func ParseRSAPrivateKey(privateKeyPEM []byte) (*rsa.PrivateKey, error) {
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

// RSAPrivateDecryptPKCS1 RSA私钥解密（PKCS1 v1.5填充）
func RSAPrivateDecryptPKCS1(privateKey *rsa.PrivateKey, encryptedData []byte) ([]byte, error) {
	// 解密：PKCS1 v1.5填充
	decrypted, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, encryptedData)
	if err != nil {
		return nil, fmt.Errorf("PKCS1解密失败：%v", err)
	}

	return decrypted, nil
}


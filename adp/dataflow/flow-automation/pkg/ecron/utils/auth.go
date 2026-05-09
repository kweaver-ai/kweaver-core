package utils

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/ecron/common"
)

//go:generate mockgen -package mock -source ../utils/auth.go -destination ../mock/mock_auth.go

// OAuthClient 授权服务客户端接口
type OAuthClient interface {
	RequestToken(id string, secret string, scope []string) (token string, duration time.Duration, err error)
	VerifyToken(token string) (visitor common.Visitor, ecronErr *common.ECronError)
	Release()
	GetSecret() string
	GetCode(secret string) (string, error)
	VerifyCode(secret, code string) (bool, *common.ECronError)
	VerifyHydraVersion() (string, bool)
}

type oauth struct {
	publicAddress string
	adminAddress  string
	httpClient    HTTPClient
	timer         *time.Timer
}

var (
	oauthConfig  = NewConfiger()
	oauthLog     = NewLogger()
	cacheHandler = common.NewCacheConfig()

	authEnabled    = oauthConfig.Config().AuthEnabled
	publicAddr     = oauthConfig.Config().OAuthPublicAddr
	publicPort     = oauthConfig.Config().OAuthPublicPort
	publicProtocol = oauthConfig.Config().OAuthPublicProtocol
	adminAddr      = oauthConfig.Config().OAuthAdminAddr
	adminPort      = oauthConfig.Config().OAuthAdminPort
	adminProtocol  = oauthConfig.Config().OAuthAdminProtocol
)

// OAuth2 OPEN API
var (
	requestTokenPATH = "/oauth2/token"
)

const (
	mockHydraAccessToken   = "mock-access-token"
	mockHydraClientID      = "mock-client-id"
	mockHydraVisitorName   = "mock-user-id"
	mockHydraTokenDuration = time.Hour
)

// NewOAuthClient 创建授权服务客户端
func NewOAuthClient() OAuthClient {
	o := &oauth{
		publicAddress: common.GetHTTPAccess(publicAddr, publicPort, publicProtocol == common.HTTPS),
		adminAddress:  common.GetHTTPAccess(adminAddr, adminPort, adminProtocol == common.HTTPS),
		httpClient:    NewHTTPClient(),
		timer:         nil,
	}

	return o
}

func authDisabled() bool {
	return strings.EqualFold(strings.TrimSpace(authEnabled), "false")
}

// 该函数未被使用
func (o *oauth) RequestToken(id string, secret string, scope []string) (token string, duration time.Duration, err error) {
	if authDisabled() {
		return mockHydraAccessToken, mockHydraTokenDuration, nil
	}

	if nil == o.httpClient {
		err = errors.New(common.ErrHTTPClientUnavailable)
		return
	}

	headers := map[string]string{
		common.ContentType:   common.ApplicationFormUrlencoded,
		common.Authorization: fmt.Sprintf("%v %v", common.Basic, base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%v:%v", id, secret)))),
	}
	req := map[string]interface{}{
		"scope":      strings.Join(scope, " "),
		"grant_type": "client_credentials",
	}
	resp := make(map[string]interface{})
	url := fmt.Sprintf("%s%s", o.publicAddress, requestTokenPATH)
	err = o.httpClient.Post(url, headers, req, &resp)
	if nil != err {
		oauthLog.Errorln(err)
		return
	}

	if v, ok := resp["access_token"]; ok {
		if reflect.TypeOf(v).Kind() == reflect.String {
			token = v.(string)
		}
	}

	if v, ok := resp["expires_in"]; ok {
		expires := 0
		switch exp := v.(type) {
		case float64:
			expires = int(exp)
		}

		if expires <= 0 {
			err = errors.New(common.ErrInvalidExpiresTime)
			oauthLog.Errorln(err)
		}

		//token过期时间（支持过期时间不小于30秒），提前10秒重新申请
		duration = time.Duration(1e9 * (common.GetIntMoreThanLowerLimit(expires, 30) - 10))
	}

	return
}

func (o *oauth) VerifyToken(token string) (visitor common.Visitor, ecronErr *common.ECronError) {
	if authDisabled() {
		return common.Visitor{Admin: true, ClientID: mockHydraClientID, Name: mockHydraVisitorName}, nil
	}

	if 0 == len(token) {
		return common.Visitor{}, NewECronError(common.ErrTokenEmpty, common.InternalError, nil)
	}

	if nil == o.httpClient {
		return common.Visitor{}, NewECronError(common.ErrHTTPClientUnavailable, common.InternalError, nil)
	}

	pos := strings.Index(token, " ")
	if pos < 0 {
		return common.Visitor{}, NewECronError(common.ErrInvalidToken, common.Unauthorized, map[string]interface{}{
			common.DetailParameters: []string{
				0: common.Authorization,
			},
		})
	}
	token = strings.TrimLeft(string([]byte(token)[pos:]), " ")

	// 自测数据
	// if token == "123456" {
	// 	return common.Visitor{Admin: false, ClientID: "123456", Name: ""}, nil
	// }

	headers := map[string]string{
		common.ContentType: common.ApplicationFormUrlencoded,
	}
	req := map[string]interface{}{"token": token}
	resp := make(map[string]interface{})
	url := fmt.Sprintf("%s%s", o.adminAddress, cacheHandler.GetHydraConfig().VerifyTokenPath)
	err := o.httpClient.Post(url, headers, req, &resp)
	if nil != err {
		oauthLog.Errorf("VerifyToken failed, url: %v, error: %v", url, err)
		return common.Visitor{}, NewECronError(err.Error(), common.InternalError, nil)
	}

	switch active := resp["active"].(type) {
	case bool:
		if !active {
			return common.Visitor{}, NewECronError(common.ErrTokenExpired, common.Unauthorized, map[string]interface{}{
				common.DetailParameters: []string{
					0: common.Authorization,
				},
			})
		}
	}

	switch clientID := resp["client_id"].(type) {
	case string:
		return common.Visitor{Admin: false, ClientID: clientID, Name: ""}, nil
	}

	return common.Visitor{}, NewECronError(common.ErrInvalidToken, common.Unauthorized, map[string]interface{}{
		common.DetailParameters: []string{
			0: common.Authorization,
		},
	})
}

func (o *oauth) Release() {
	if nil != o.timer {
		o.timer.Stop()
	}
}

// 每30秒更换一次
func (o *oauth) un() int64 {
	return time.Now().UnixNano() / 1000 / 30
}

func (this *oauth) hmacSha1(key, data []byte) []byte {
	h := hmac.New(sha1.New, key)
	if total := len(data); total > 0 {
		h.Write(data)
	}
	return h.Sum(nil)
}

func (this *oauth) base32encode(src []byte) string {
	return base32.StdEncoding.EncodeToString(src)
}

func (this *oauth) base32decode(s string) ([]byte, error) {
	return base32.StdEncoding.DecodeString(s)
}

func (this *oauth) toBytes(value int64) []byte {
	var result []byte
	mask := int64(0xFF)
	shifts := [8]uint16{56, 48, 40, 32, 24, 16, 8, 0}
	for _, shift := range shifts {
		result = append(result, byte((value>>shift)&mask))
	}
	return result
}

func (this *oauth) toUint32(bts []byte) uint32 {
	return (uint32(bts[0]) << 24) + (uint32(bts[1]) << 16) +
		(uint32(bts[2]) << 8) + uint32(bts[3])
}

func (this *oauth) oneTimePassword(key []byte, data []byte) uint32 {
	hash := this.hmacSha1(key, data)
	offset := hash[len(hash)-1] & 0x0F
	hashParts := hash[offset : offset+4]
	hashParts[0] = hashParts[0] & 0x7F
	number := this.toUint32(hashParts)
	return number % 1000000
}

// 获取秘钥
func (this *oauth) GetSecret() string {
	var buf bytes.Buffer
	err := binary.Write(&buf, binary.BigEndian, this.un())
	if err != nil {
		oauthLog.Infof("GetSecret failed, buf:%v error: %v", buf, err)
	}
	return strings.ToUpper(this.base32encode(this.hmacSha1(buf.Bytes(), nil)))
}

// 获取动态码
func (this *oauth) GetCode(secret string) (string, error) {
	secretUpper := strings.ToUpper(secret)
	secretKey, err := this.base32decode(secretUpper)
	if err != nil {
		oauthLog.Infof("GetCode failed, secret :%v,secretKey: %v  error: %v", secret, secretKey, err)
		return "", err
	}
	number := this.oneTimePassword(secretKey, this.toBytes(time.Now().Unix()/30))
	return fmt.Sprintf("%06d", number), nil
}

// 验证动态码
func (this *oauth) VerifyCode(secret, code string) (bool, *common.ECronError) {
	_code, err := this.GetCode(secret)
	if err != nil {
		return false, NewECronError(err.Error(), common.InternalError, nil)
	}
	return _code == code, nil
}

var (
	oldVerifyTokenPath = "/oauth2/introspect"
	newVerifyTokenPath = "/admin/oauth2/introspect"
)

func (o *oauth) VerifyHydraVersion() (string, bool) {
	if authDisabled() {
		return newVerifyTokenPath, true
	}

	if nil == o.httpClient {
		return "", false
	}

	headers := map[string]string{
		common.ContentType: common.ApplicationFormUrlencoded,
	}
	req := map[string]interface{}{"token": ""}
	oldUrl := fmt.Sprintf("%s%s", o.adminAddress, oldVerifyTokenPath)
	newUrl := fmt.Sprintf("%s%s", o.adminAddress, newVerifyTokenPath)

	code1, _, err1 := o.httpClient.PostV2(newUrl, headers, req)
	if code1 == http.StatusOK {
		return newVerifyTokenPath, true
	}
	oauthLog.Infof("verifyHydraVersion failed, code: %v, newUrl: %v, error: %v", code1, newUrl, err1)

	code2, _, err2 := o.httpClient.PostV2(oldUrl, headers, req)
	if code2 == http.StatusOK {
		return oldVerifyTokenPath, true
	}
	oauthLog.Infof("verifyHydraVersion failed, code: %v, oldUrl: %v, error: %v", code2, oldUrl, err2)

	return "", false
}

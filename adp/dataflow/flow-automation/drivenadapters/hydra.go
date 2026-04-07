package drivenadapters

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"

	jsoniter "github.com/json-iterator/go"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	otelHttp "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/http"
	commonLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/log"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
)

//go:generate mockgen -package mock_drivenadapters -source ../drivenadapters/hydra.go -destination ../tests/mock_drivenadapters/hydra_mock.go

// TokenIntrospectInfo 令牌内省结果
type TokenIntrospectInfo struct {
	Active      bool   // 令牌状态
	UserID      string // 用户ID
	UdID        string // 设备ID
	LoginIP     string // 登录IP
	VisitorType string // 访问者类型
	ClientType  string // 客户端类型
	ClientID    string // 客户端ID
	ExpiresIn   int64  // 过期时间
}

type ext struct {
	ClientType  string `json:"client_type"` // 客户端类型
	Udid        string `json:"udid"`        // 设备编码
	LoginIP     string `json:"login_ip"`
	VisitorType string `json:"visitor_type"` // 访问者类型
}

// HydraAdmin 授权服务接口
type HydraAdmin interface {
	// Introspect token内省
	Introspect(ctx context.Context, token string) (info TokenIntrospectInfo, err error)

	// UpdateClient 更新客户端
	UpdateClient(ctx context.Context, id, name, secret, redirectURI, logoutRedirectURI string) (code int, err error)
}

type hydraAdmin struct {
	adminAddress string
	client       *http.Client
	httpClient   otelHttp.HTTPClient
}

var (
	hOnce sync.Once
	h     HydraAdmin
)

// NewHydraAdmin 创建授权服务
func NewHydraAdmin() HydraAdmin {
	hOnce.Do(func() {
		config := common.NewConfig()
		if !config.IsAuthEnabled() {
			h = &mockHydraAdmin{}
		} else {
			h = &hydraAdmin{
				adminAddress: fmt.Sprintf("http://%s:%v%s", config.OAuth.AdminHost, config.OAuth.AdminPort, config.OAuth.AdminPrefix),
				client:       NewOtelRawHTTPClient(),
				httpClient:   NewOtelHTTPClient(),
			}
		}
	})

	return h
}

// Introspect token内省
func (h *hydraAdmin) Introspect(ctx context.Context, token string) (info TokenIntrospectInfo, err error) {
	target := fmt.Sprintf("%v/oauth2/introspect", h.adminAddress)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target, bytes.NewReader([]byte(fmt.Sprintf("token=%v", token))))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := h.client.Do(req)
	if err != nil {
		traceLog.WithContext(ctx).Warnln(err)
		return
	}

	defer func() {
		closeErr := resp.Body.Close()
		if closeErr != nil {
			traceLog.WithContext(ctx).Warnln(closeErr)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if (resp.StatusCode < http.StatusOK) || (resp.StatusCode >= http.StatusMultipleChoices) {
		err = errors.New(string(body))
		return
	}

	if len(body) != 0 {
		respParam := make(map[string]interface{})
		err = jsoniter.Unmarshal(body, &respParam)
		if err != nil {
			return
		}

		info.Active = respParam["active"].(bool)
		if info.Active {
			info.UserID = respParam["sub"].(string)
			info.ClientID = respParam["client_id"].(string)
			info.ExpiresIn = int64(respParam["exp"].(float64)) - int64(respParam["iat"].(float64))
			extInfo, ok := respParam["ext"].(map[string]interface{})
			if ok {
				var e ext
				extBytes, _ := json.Marshal(extInfo)
				err = jsoniter.Unmarshal(extBytes, &e)
				if err != nil {
					return
				}

				info.UdID = e.Udid
				info.LoginIP = e.LoginIP
				info.ClientType = e.ClientType
				info.VisitorType = e.VisitorType
			}
		}
	}

	return
}

// UpdateClient 更新客户端
func (h *hydraAdmin) UpdateClient(ctx context.Context, id, name, secret, redirectURI, logoutRedirectURI string) (code int, err error) {
	urlStr := fmt.Sprintf("%s/clients/%s", h.adminAddress, id)
	paras := map[string]interface{}{
		"client_name":               name,
		"client_secret":             secret,
		"grant_types":               []string{"authorization_code", "implicit", "refresh_token"},
		"response_types":            []string{"token id_token", "code", "token"},
		"scope":                     "offline openid all",
		"redirect_uris":             []string{redirectURI},
		"post_logout_redirect_uris": []string{logoutRedirectURI},
		"metadata": map[string]map[string]string{
			"device": {
				"name":        "内容自动化",
				"client_type": "web",
				"description": "内容自动化",
			},
		},
	}
	code, _, err = h.httpClient.Put(ctx, urlStr, map[string]string{"Content-Type": "application/json;charset=UTF-8"}, paras)

	if err != nil {
		traceLog.WithContext(ctx).Errorf("update oauth2 client failed: %v", err)
		return
	}
	return
}

// HydraPublic 授权服务客户端
type HydraPublic interface {
	RequestTokenWithCredential(id, secret string, scope []string) (tokenInfo TokenInfo, code int, err error)
	RequestTokenWithRefreshToken(id, secret, token string) (tokenInfo TokenInfo, code int, err error)
	RequestTokenWithCode(id, secret, code, redirect string) (tokenInfo TokenInfo, rescode int, err error)
	RequestTokenWithAsserts(id, secret, assertion string) (tokenInfo TokenInfo, code int, err error)
	RegisterClient(name, redirectURI, logoutRedirectURI string) (clientID, clientSecret string, code int, err error)
}

// TokenInfo 令牌信息
type TokenInfo struct {
	Token        string
	ExpiresIn    int
	RefreshToken string
}

type hydraPublic struct {
	publicAddress             string
	authPublicAddress         string
	contentType               string
	applicationFormUrlencoded string
	authorization             string
	basic                     string
	logger                    commonLog.Logger
	client                    *http.Client
	httpClient                HTTPClient
	useHTTPS                  bool
}

// NewHydraPublic 创建一个本文档域的public授权服务客户端对象
func NewHydraPublic() (hydraP HydraPublic) {
	config := common.NewConfig()
	if !config.IsAuthEnabled() {
		return &mockHydraPublic{}
	}

	logger := commonLog.NewLogger()
	client := NewRawHTTPClient()
	publicAddress := fmt.Sprintf("%s:%v", config.OAuth.PublicHost, config.OAuth.PublicPort)
	authPublicAddress := fmt.Sprintf("%s:%v", config.Authentication.PublicHost, config.Authentication.PublicPort)

	hydraP = &hydraPublic{
		publicAddress:             publicAddress,
		authPublicAddress:         authPublicAddress,
		contentType:               "Content-Type",
		applicationFormUrlencoded: "application/x-www-form-urlencoded",
		authorization:             "Authorization",
		basic:                     "Basic",
		logger:                    logger,
		client:                    client,
		httpClient:                NewHTTPClient(),
		useHTTPS:                  false,
	}
	return hydraP
}

// NewHydraPublicWithAddress 创建一个public授权服务客户端对象
func NewHydraPublicWithAddress(publicAddress string, useHTTPS bool) HydraPublic {
	logger := commonLog.NewLogger()
	client := NewRawHTTPClient()
	if !common.NewConfig().IsAuthEnabled() {
		return &mockHydraPublic{}
	}

	o := &hydraPublic{
		publicAddress:             publicAddress,
		contentType:               "Content-Type",
		applicationFormUrlencoded: "application/x-www-form-urlencoded",
		authorization:             "Authorization",
		basic:                     "Basic",
		logger:                    logger,
		client:                    client,
		useHTTPS:                  useHTTPS,
	}
	return o
}

// RequestTokenWithAsserts 断言方式申请令牌
func (o *hydraPublic) RequestTokenWithAsserts(id, secret, assertion string) (tokenInfo TokenInfo, code int, err error) {
	reqParam := make([]string, 3) //nolint
	reqParam[0] = fmt.Sprintf("%v=%v", "grant_type", "urn:ietf:params:oauth:grant-type:jwt-bearer")
	reqParam[1] = fmt.Sprintf("%v=%v", "scope", "all")
	reqParam[2] = fmt.Sprintf("%v=%v", "assertion", assertion)
	return o.requestToken(id, secret, reqParam)
}

// RequestToken client_credentials申请令牌
func (o *hydraPublic) RequestTokenWithCredential(id, secret string, scope []string) (tokenInfo TokenInfo, code int, err error) {
	reqParam := make([]string, 2) //nolint
	reqParam[0] = fmt.Sprintf("%v=%v", "grant_type", "client_credentials")
	reqParam[1] = fmt.Sprintf("%v=%v", "scope", strings.Join(scope, " "))
	return o.requestToken(id, secret, reqParam)
}

// RequestTokenWithRefreshToken 申请令牌
func (o *hydraPublic) RequestTokenWithRefreshToken(id, secret, token string) (tokenInfo TokenInfo, code int, err error) {
	reqParam := make([]string, 2) //nolint
	reqParam[0] = fmt.Sprintf("%v=%v", "grant_type", "refresh_token")
	reqParam[1] = fmt.Sprintf("%v=%v", "refresh_token", token)

	return o.requestToken(id, secret, reqParam)
}

// RequestTokenWithCode 申请令牌
func (o *hydraPublic) RequestTokenWithCode(id, secret, code, redirect string) (tokenInfo TokenInfo, rescode int, err error) {
	reqParam := make([]string, 3) //nolint
	reqParam[0] = fmt.Sprintf("%v=%v", "grant_type", "authorization_code")
	reqParam[1] = fmt.Sprintf("%v=%v", "code", code)
	reqParam[2] = fmt.Sprintf("%v=%v", "redirect_uri", redirect)

	return o.requestToken(id, secret, reqParam)
}

func (o *hydraPublic) requestToken(id, secret string, reqParam []string) (tokenInfo TokenInfo, code int, err error) {
	urlStr := ""
	if o.useHTTPS {
		urlStr = fmt.Sprintf("https://%s/oauth2/token", o.publicAddress)
	} else {
		urlStr = fmt.Sprintf("http://%s/oauth2/token", o.publicAddress)
	}

	body := bytes.NewReader([]byte(strings.Join(reqParam, "&")))
	req, err := http.NewRequest("POST", urlStr, body)
	if err != nil {
		return
	}

	headers := make([]string, 2) //nolint
	headers[0] = fmt.Sprintf("%v:%v", o.contentType, o.applicationFormUrlencoded)
	headers[1] = fmt.Sprintf("%v:%v %v", o.authorization, o.basic, base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%v:%v", url.QueryEscape(id), url.QueryEscape(secret)))))
	for _, header := range headers {
		i := strings.IndexRune(header, ':')
		if i == -1 {
			err = fmt.Errorf("invalid header: %v", header)
			return
		}
		req.Header.Set(header[:i], strings.TrimSpace(header[i+1:]))
	}
	resp, err := o.client.Do(req)
	if err != nil {
		o.logger.Errorf("RequestToken failed: %v", err)
		return
	}

	defer func() {
		closeErr := resp.Body.Close()
		if closeErr != nil {
			o.logger.Errorln(closeErr)
		}
	}()

	code = resp.StatusCode

	respBody, _ := io.ReadAll(resp.Body)
	if (resp.StatusCode < http.StatusOK) || (resp.StatusCode >= http.StatusMultipleChoices) {
		err = fmt.Errorf("%v", string(respBody))
		return
	}
	var respParam interface{}
	if len(respBody) != 0 {
		err = json.Unmarshal(respBody, &respParam)
		if err != nil {
			return
		}
		if token, ok := respParam.(map[string]interface{})["access_token"].(string); ok {
			tokenInfo.Token = token
		}
		if refreshToken, ok := respParam.(map[string]interface{})["refresh_token"].(string); ok {
			tokenInfo.RefreshToken = refreshToken
		}
		if expiresIn, ok := respParam.(map[string]interface{})["expires_in"].(float64); ok {
			tokenInfo.ExpiresIn = int(expiresIn)
		}
	}
	return
}

// RegisterClient 注册客户端
func (o *hydraPublic) RegisterClient(name, redirectURI, logoutRedirectURI string) (clientID, clientSecret string, code int, err error) {
	urlStr := fmt.Sprintf("http://%s/oauth2/clients", o.authPublicAddress)
	paras := map[string]interface{}{
		"client_name":               name,
		"grant_types":               []string{"authorization_code", "implicit", "refresh_token"},
		"response_types":            []string{"token id_token", "code", "token"},
		"scope":                     "offline openid all",
		"redirect_uris":             []string{redirectURI},
		"post_logout_redirect_uris": []string{logoutRedirectURI},
		"metadata": map[string]map[string]string{
			"device": {
				"name":        "内容自动化",
				"client_type": "web",
				"description": "内容自动化",
			},
		},
	}
	code, respParam, err := o.httpClient.Post(urlStr, map[string]string{"Content-Type": "application/json;charset=UTF-8"}, paras)

	if err != nil {
		o.logger.Errorf("register oauth2 client failed: %v", err)
		return
	}

	clientID = respParam.(map[string]interface{})["client_id"].(string)
	clientSecret = respParam.(map[string]interface{})["client_secret"].(string)
	return
}

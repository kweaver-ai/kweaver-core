// Package drivenadapters 定义驱动适配器
// @file hydra.go
// @description: 实现授权服务接口
package drivenadapters

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/rest"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
)

type hydraService struct {
	adminAddress string
	logger       interfaces.Logger
	httpClient   interfaces.HTTPClient
}

type noopHydra struct{}

var (
	once sync.Once
	h    interfaces.Hydra
)

// Extend 解析拓展信息
type Extend struct {
	AccountType string `json:"account_type"`
	ClientType  string `json:"client_type"`
	LoginIP     string `json:"login_ip"`
	UdID        string `json:"udid"`
	VisitorType string `json:"visitor_type"`
	PhoneNumber string `json:"phone_number"`
	VisitorName string `json:"visitor_name"`
}

// IntrospectInfo 内省信息
type IntrospectInfo struct {
	Active    bool   `json:"active"`
	Scope     string `json:"scope"`
	ClientID  string `json:"client_id"`
	SubID     string `json:"sub"`
	TokenType string `json:"token_type"`
	Ext       Extend `json:"ext"`
}

const introspectURI = "/oauth2/introspect"

// NewHydra 创建授权服务对象
func NewHydra() interfaces.Hydra {
	if !config.GetAuthEnabled() {
		return &noopHydra{}
	}
	once.Do(func() {
		config := config.NewConfigLoader()
		h = &hydraService{
			adminAddress: fmt.Sprintf("http://%s:%d%s", config.OAuth.AdminHost, config.OAuth.AdminPort, config.OAuth.AdminPrefix),
			logger:       config.GetLogger(),
			httpClient:   rest.NewHTTPClient(),
		}
	})
	return h
}

// 获取通用的认证信息
// 从Header中获取X-Account-Type和X-Account-ID，构建TokenInfo对象
// 如果X-Account-Type为空，默认设置为AccessorTypeAnonymous
// 如果X-Account-ID为空，默认设置为空字符串

func (n *noopHydra) GenerateVisitor(c *gin.Context) (info *interfaces.TokenInfo, err error) {
	xAccountType := c.GetHeader(string(interfaces.HeaderXAccountType))
	xAccountID := c.GetHeader(string(interfaces.HeaderXAccountID))
	info = &interfaces.TokenInfo{
		Active:     true,
		VisitorID:  xAccountID,
		VisitorTyp: interfaces.AccessorType(xAccountType).ToVisitorType(),
		LoginIP:    c.ClientIP(),
		MAC:        c.GetHeader("X-Request-MAC"),
		UserAgent:  c.GetHeader("User-Agent"),
	}
	if info.VisitorID == "" {
		// 如果用户未登录，默认设置为匿名用户
		info.VisitorID = interfaces.UnknownUser
		info.VisitorTyp = interfaces.Anonymous
	}
	return info, nil
}

func (n *noopHydra) Introspect(c *gin.Context) (info *interfaces.TokenInfo, err error) {
	info, err = n.GenerateVisitor(c)
	return
}

func (h *hydraService) GenerateVisitor(c *gin.Context) (info *interfaces.TokenInfo, err error) {
	xAccountType := c.GetHeader(string(interfaces.HeaderXAccountType))
	xAccountID := c.GetHeader(string(interfaces.HeaderXAccountID))
	info = &interfaces.TokenInfo{
		Active:     true,
		VisitorID:  xAccountID,
		VisitorTyp: interfaces.AccessorType(xAccountType).ToVisitorType(),
		LoginIP:    c.ClientIP(),
		MAC:        c.GetHeader("X-Request-MAC"),
		UserAgent:  c.GetHeader("User-Agent"),
	}
	return info, nil
}

// Introspect token内省
func (h *hydraService) Introspect(c *gin.Context) (info *interfaces.TokenInfo, err error) {
	ctx := c.Request.Context()
	token := getToken(c)
	target := fmt.Sprintf("%s%s", h.adminAddress, introspectURI)
	header := map[string]string{"Content-Type": "application/x-www-form-urlencoded"}
	_, resp, err := h.httpClient.Post(ctx, target, header, []byte(fmt.Sprintf("token=%v", token)))
	if err != nil {
		h.logger.WithContext(ctx).Error(err)
		return
	}
	introspectInfo := &IntrospectInfo{}
	respByt := utils.ObjectToByte(resp)
	if err = jsoniter.Unmarshal(respByt, introspectInfo); err != nil {
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		h.logger.WithContext(ctx).Warnf("Get introspect object to struct failed:%+v, resp:%+v", err, resp)
		return
	}
	info = &interfaces.TokenInfo{}
	// 令牌状态
	info.Active = introspectInfo.Active
	if !info.Active {
		err = errors.DefaultHTTPError(ctx, http.StatusUnauthorized, "token is invalid")
		return
	}
	// 访问者ID
	info.VisitorID = introspectInfo.SubID
	// Scope 权限范围
	info.Scope = introspectInfo.Scope
	// 客户端ID
	info.ClientID = introspectInfo.ClientID
	// 客户端凭据模式
	if info.VisitorID == info.ClientID {
		info.VisitorTyp = interfaces.Business
		return
	}
	// 以下字段 只在非客户端凭据模式时才存在
	// 访问者类型
	info.VisitorTyp = interfaces.VisitorType(introspectInfo.Ext.VisitorType)

	// 匿名用户
	if info.VisitorTyp == interfaces.Anonymous {
		info.PhoneNumber = introspectInfo.Ext.PhoneNumber
		info.VisitorName = introspectInfo.Ext.VisitorName
		return
	}
	// 实名用户
	if info.VisitorTyp == interfaces.RealName {
		// 登陆IP
		info.LoginIP = introspectInfo.Ext.LoginIP
		// 用户名
		info.VisitorName = introspectInfo.Ext.VisitorName
		// 设备ID
		info.Udid = introspectInfo.Ext.UdID
		// 登录账号类型
		info.AccountTyp = interfaces.ReverseAccountTypeMap[introspectInfo.Ext.AccountType]
		// 设备类型
		info.ClientTyp = interfaces.ReverseClientTypeMap[introspectInfo.Ext.ClientType]
	}
	if info.LoginIP == "" {
		// 若返回IP为空则使用clientIP
		info.LoginIP = c.ClientIP()
	}
	info.MAC = c.GetHeader("X-Request-MAC")
	info.UserAgent = c.GetHeader("User-Agent")
	return
}

func getToken(c *gin.Context) (token string) {
	tokenID := c.GetHeader("Authorization")
	if tokenID == "" {
		tokenID = c.GetHeader("X-Authorization")
	}
	if tokenID == "" {
		token, _ = c.GetQuery("token")
	} else {
		token = strings.TrimPrefix(tokenID, "Bearer ")
	}
	return token
}

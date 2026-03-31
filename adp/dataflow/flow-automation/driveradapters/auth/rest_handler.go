// Package auth check auth
package auth

import (
	"encoding/base64"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/driveradapters/middleware"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/errors"
	commonLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/log"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/logics/auth"
)

// RESTHandler 公共RESTful api Handler接口
type RESTHandler interface {
	// 注册开放API
	RegisterAPI(engine *gin.RouterGroup)
}

var (
	once       sync.Once
	rh         RESTHandler
	authSchema = "base/auth.json"
)

type restHandler struct {
	auth       auth.AuthHandler
	hydra      drivenadapters.HydraPublic
	hydraAdmin drivenadapters.HydraAdmin
	logger     commonLog.Logger
}

// CheckAuthReq CheckAuthReq struct
type CheckAuthReq struct {
	RedirectURI string `json:"redirect_uri"`
}

// NewRESTHandler 创建公共RESTful api handler对象
func NewRESTHandler() RESTHandler {
	once.Do(func() {
		rh = &restHandler{
			auth:       auth.NewAuth(),
			hydra:      drivenadapters.NewHydraPublic(),
			hydraAdmin: drivenadapters.NewHydraAdmin(),
			logger:     commonLog.NewLogger(),
		}
	})

	return rh
}

// 注册开放API
func (h *restHandler) RegisterAPI(engine *gin.RouterGroup) {
	engine.GET("/oauth2/callback", h.hydraCallback)
	engine.POST("/oauth2/auth", middleware.TokenAuth(), h.saveAuth)
}

// 回调处理
func (h *restHandler) hydraCallback(c *gin.Context) {
	ip := c.ClientIP()
	code := c.Query("code")
	state := c.Query("state")
	if code == "" {
		err := errors.NewIError(errors.InvalidParameter, "", map[string]string{"code": "empty is not allowed"})
		errors.ReplyError(c, err)
		return
	}
	if state == "" {
		err := errors.NewIError(errors.InvalidParameter, "", map[string]string{"state": "empty is not allowed"})
		errors.ReplyError(c, err)
		return
	}

	sDec, err := base64.StdEncoding.DecodeString(state)
	if err != nil {
		h.logger.Errorf("base64 decode failure, error=%v", err.Error())
		err := errors.NewIError(errors.InvalidParameter, "", map[string]string{"state": "base64 decode failure"})
		errors.ReplyError(c, err)
		return
	}

	err = h.auth.RequestToken(c.Request.Context(), code, ip)
	if err != nil {
		errors.ReplyError(c, err)
		return
	}

	c.Redirect(http.StatusFound, string(sDec))
}

func (h *restHandler) saveAuth(c *gin.Context) {
	c.PureJSON(http.StatusOK, gin.H{
		"status": true,
		"url":    "",
	})
}

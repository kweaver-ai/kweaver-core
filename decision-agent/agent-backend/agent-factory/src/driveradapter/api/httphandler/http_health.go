package httphandler

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driver/ihandlerportdriver"
)

// 健康检查
type httpHealthHandler struct{}

var (
	httpHealthOnce sync.Once
	httpHealthHand ihandlerportdriver.IHTTPRouter
)

func NewHTTPHealthHandler() ihandlerportdriver.IHTTPRouter {
	httpHealthOnce.Do(func() {
		httpHealthHand = &httpHealthHandler{}
	})

	return httpHealthHand
}

// RegisterRouter 注册路由
func (h *httpHealthHandler) RegPriRouter(router *gin.RouterGroup) {
	router.GET("/ready", h.getReady)
	router.GET("/alive", h.getAlive)
}

func (h *httpHealthHandler) RegPubRouter(router *gin.RouterGroup) {
}

func (h *httpHealthHandler) getReady(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.String(http.StatusOK, "ready")
}

func (h *httpHealthHandler) getAlive(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.String(http.StatusOK, "alive")
}

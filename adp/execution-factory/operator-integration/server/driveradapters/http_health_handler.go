// Package driveradapters 定义驱动适配器
// @file http_health_handler.go
// @description: 定义HTTP健康检查适配器
package driveradapters

import (
	"net/http"
	"sync"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/gin-gonic/gin"
)

// 健康检查
type httpHealthHandler struct{}

var (
	httpHealthOnce sync.Once
	httpHealthHand interfaces.HTTPRouterInterface
)

func NewHTTPHealthHandler() interfaces.HTTPRouterInterface {
	httpHealthOnce.Do(func() {
		httpHealthHand = &httpHealthHandler{}
	})

	return httpHealthHand
}

// RegisterRouter 注册路由
func (h *httpHealthHandler) RegisterRouter(router *gin.RouterGroup) {
	router.GET("/ready", h.getReady)
	router.GET("/alive", h.getAlive)
}

func (h *httpHealthHandler) getReady(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.String(http.StatusOK, "ready")
}

func (h *httpHealthHandler) getAlive(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.String(http.StatusOK, "alive")
}

// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
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

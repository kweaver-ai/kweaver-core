// Package worker worker driveradapters
package worker

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

// RESTHandler 公共RESTful api Handler接口
type RESTHandler interface {
	// 注册开放API
	RegisterAPI(engine *gin.RouterGroup)
}

var (
	once sync.Once
	rh   RESTHandler
)

type restHandler struct {
}

// NewRESTHandler 创建公共RESTful api handler对象
func NewRESTHandler() RESTHandler {
	once.Do(func() {
		rh = &restHandler{}
	})

	return rh
}

// 注册开放API
func (h *restHandler) RegisterAPI(engine *gin.RouterGroup) {
	engine.GET("/worker", h.worker)
}

func (h *restHandler) worker(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.String(http.StatusOK, "ready")
}

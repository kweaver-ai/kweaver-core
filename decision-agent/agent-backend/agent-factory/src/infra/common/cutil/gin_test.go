package cutil

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestUpdateGinReqCtx(t *testing.T) {
	t.Parallel()

	router := gin.Default()

	var updatedCtx context.Context

	router.GET("/test", func(c *gin.Context) {
		ctx := context.WithValue(context.Background(), "key", "value") //nolint:staticcheck
		UpdateGinReqCtx(c, ctx)

		updatedCtx = c.Request.Context()
	})

	// 模拟请求
	performRequest(router, "GET", "/test")

	// 检查更新的上下文是否具有期望的值
	if assert.NotNil(t, updatedCtx) {
		assert.Equal(t, "value", updatedCtx.Value("key"))
	}
}

// 辅助函数用于执行HTTP请求
func performRequest(r *gin.Engine, method, path string) {
	req, _ := http.NewRequest(method, path, nil)
	r.ServeHTTP(httptest.NewRecorder(), req)
}

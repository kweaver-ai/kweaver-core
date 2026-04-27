package capimiddleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cenum"
)

func TestSetInternalAPIFlag(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	t.Run("sets flag in gin context and request context", func(t *testing.T) {
		t.Parallel()

		var capturedGinFlag interface{}

		var capturedReqCtxFlag interface{}

		router := gin.New()
		router.Use(SetInternalAPIFlag())
		router.GET("/test", func(c *gin.Context) {
			capturedGinFlag, _ = c.Get(cenum.InternalAPIFlagCtxKey.String())
			capturedReqCtxFlag = c.Request.Context().Value(cenum.InternalAPIFlagCtxKey.String())
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, true, capturedGinFlag)
		assert.Equal(t, true, capturedReqCtxFlag)
	})

	t.Run("calls Next", func(t *testing.T) {
		t.Parallel()

		nextCalled := false
		router := gin.New()
		router.Use(SetInternalAPIFlag())
		router.GET("/test", func(c *gin.Context) {
			nextCalled = true

			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.True(t, nextCalled)
	})
}

func TestIsInternalAPI(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	t.Run("returns true when flag set via middleware", func(t *testing.T) {
		t.Parallel()

		var result bool

		router := gin.New()
		router.Use(SetInternalAPIFlag())
		router.GET("/test", func(c *gin.Context) {
			result = IsInternalAPI(c)
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.True(t, result)
	})

	t.Run("returns false when flag not set", func(t *testing.T) {
		t.Parallel()

		var result bool

		router := gin.New()
		router.GET("/test", func(c *gin.Context) {
			ctx := context.WithValue(c.Request.Context(), cenum.InternalAPIFlagCtxKey.String(), false) //nolint:staticcheck // SA1029
			c.Request = c.Request.WithContext(ctx)
			result = IsInternalAPI(c)
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.False(t, result)
	})
}

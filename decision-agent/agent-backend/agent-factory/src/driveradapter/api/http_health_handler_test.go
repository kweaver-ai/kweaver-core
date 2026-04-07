package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestNewHTTPHealthHandler(t *testing.T) {
	t.Parallel()

	t.Run("returns singleton handler", func(t *testing.T) {
		t.Parallel()

		handler1 := NewHTTPHealthHandler()
		handler2 := NewHTTPHealthHandler()

		assert.Same(t, handler1, handler2, "Should return the same singleton instance")
		assert.NotNil(t, handler1)
	})
}

func TestHTTPHealthHandler_RegHealthRouter(t *testing.T) {
	t.Parallel()

	t.Run("registers health check routes", func(t *testing.T) {
		t.Parallel()
		gin.SetMode(gin.TestMode)
		router := gin.New()
		routerGroup := router.Group("/health")

		handler := NewHTTPHealthHandler()
		handler.RegHealthRouter(routerGroup)

		// Test ready endpoint
		req1 := httptest.NewRequest("GET", "/health/ready", nil)
		w1 := httptest.NewRecorder()
		router.ServeHTTP(w1, req1)

		assert.Equal(t, http.StatusOK, w1.Code)
		assert.Equal(t, "application/json", w1.Header().Get("Content-Type"))
		assert.Equal(t, "ready", w1.Body.String())

		// Test alive endpoint
		req2 := httptest.NewRequest("GET", "/health/alive", nil)
		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, req2)

		assert.Equal(t, http.StatusOK, w2.Code)
		assert.Equal(t, "application/json", w2.Header().Get("Content-Type"))
		assert.Equal(t, "alive", w2.Body.String())
	})
}

func TestHTTPHealthHandler_getReady(t *testing.T) {
	t.Parallel()

	t.Run("returns ready status", func(t *testing.T) {
		t.Parallel()
		gin.SetMode(gin.TestMode)

		handler := &httpHealthHandler{}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		handler.getReady(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		assert.Equal(t, "ready", w.Body.String())
	})
}

func TestHTTPHealthHandler_getAlive(t *testing.T) {
	t.Parallel()

	t.Run("returns alive status", func(t *testing.T) {
		t.Parallel()
		gin.SetMode(gin.TestMode)

		handler := &httpHealthHandler{}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		handler.getAlive(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		assert.Equal(t, "alive", w.Body.String())
	})
}

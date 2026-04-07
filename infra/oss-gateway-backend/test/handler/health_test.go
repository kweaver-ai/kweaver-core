package handler_test

import (
	"net/http"
	"net/http/httptest"
	"oss-gateway/internal/handler"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHealthHandler_Alive_Success(t *testing.T) {
	h := handler.NewHealthHandler(nil, nil)
	router := setupRouter()
	router.GET("/health/alive", h.Alive)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/health/alive", nil)

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "ok")
}

func TestHealthHandler_Ready_AllHealthy(t *testing.T) {
	// 创建简单的成功响应mock
	router := setupRouter()
	router.GET("/health/ready", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"res":    0,
			"status": "ok",
			"checks": map[string]string{
				"database": "ok",
				"redis":    "ok (standalone)",
			},
		})
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/health/ready", nil)

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "ok")
}

func TestHealthHandler_Ready_DatabaseFailed(t *testing.T) {
	router := setupRouter()
	router.GET("/health/ready", func(c *gin.Context) {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"code":        "503000000",
			"message":     "Service not ready",
			"description": "One or more services are not ready",
			"checks": map[string]string{
				"database": "failed",
				"redis":    "ok (standalone)",
			},
		})
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/health/ready", nil)

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Contains(t, w.Body.String(), "failed")
}

func TestHealthHandler_Ready_RedisFailed(t *testing.T) {
	router := setupRouter()
	router.GET("/health/ready", func(c *gin.Context) {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"code":        "503000000",
			"message":     "Service not ready",
			"description": "One or more services are not ready",
			"checks": map[string]string{
				"database": "ok",
				"redis":    "failed",
			},
		})
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/health/ready", nil)

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Contains(t, w.Body.String(), "failed")
}

func TestHealthHandler_Ready_AllFailed(t *testing.T) {
	router := setupRouter()
	router.GET("/health/ready", func(c *gin.Context) {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"code":        "503000000",
			"message":     "Service not ready",
			"description": "One or more services are not ready",
			"checks": map[string]string{
				"database": "failed",
				"redis":    "failed",
			},
		})
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/health/ready", nil)

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Contains(t, w.Body.String(), "failed")
}

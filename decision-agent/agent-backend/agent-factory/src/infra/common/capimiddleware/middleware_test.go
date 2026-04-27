package capimiddleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cenum"
)

// ==================== SetInternalAPIUserInfo ====================

func TestSetInternalAPIUserInfo_NoAccountID(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	nextCalled := false
	router := gin.New()
	router.Use(SetInternalAPIFlag())
	router.Use(SetInternalAPIUserInfo(true))
	router.GET("/test", func(c *gin.Context) {
		nextCalled = true

		c.Status(http.StatusOK)
	})

	// 没有 x-account-id header
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.True(t, nextCalled, "should continue when no account-id")
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSetInternalAPIUserInfo_WithAccountID_NoType(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(SetInternalAPIFlag())
	router.Use(SetInternalAPIUserInfo(true))
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("x-account-id", "user-1")
	// 没有设置 x-account-type
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestSetInternalAPIUserInfo_UnsupportedAccountType(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(SetInternalAPIFlag())
	router.Use(SetInternalAPIUserInfo(true, cenum.AccountTypeUser))
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("x-account-id", "user-1")
	req.Header.Set("x-account-type", "app")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestSetInternalAPIUserInfo_SupportedAccountType(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	var capturedVisitorExists bool

	router := gin.New()
	router.Use(SetInternalAPIFlag())
	router.Use(SetInternalAPIUserInfo(false))
	router.GET("/test", func(c *gin.Context) {
		_, capturedVisitorExists = c.Get(cenum.VisitUserInfoCtxKey.String())
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("x-account-id", "user-1")
	req.Header.Set("x-account-type", string(cenum.AccountTypeUser))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, capturedVisitorExists)
}

func TestSetInternalAPIUserInfo_NoCheckAccountType(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(SetInternalAPIFlag())
	router.Use(SetInternalAPIUserInfo(false))
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("x-account-id", "user-1")
	req.Header.Set("x-account-type", "app")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// ==================== ErrorHandler ====================

func TestErrorHandler_NoErrors(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(ErrorHandler())
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestErrorHandler_WithHTTPError(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(ErrorHandler())
	router.GET("/test", func(c *gin.Context) {
		httpErr := capierr.New400Err(c, "test error")
		_ = c.Error(httpErr)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.NotEqual(t, http.StatusOK, w.Code)
}

func TestErrorHandler_WithGenericError(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(ErrorHandler())
	router.GET("/test", func(c *gin.Context) {
		_ = c.Error(errors.New("generic error"))
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 期望 ErrorHandler 处理了错误
	assert.NotEqual(t, http.StatusOK, w.Code)
}

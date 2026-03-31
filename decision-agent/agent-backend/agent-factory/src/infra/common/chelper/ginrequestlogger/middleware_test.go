package ginrequestlogger

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/httprequesthelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestLogger_Middleware_Disabled(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	config := &httprequesthelper.Config{
		Enabled:    false,
		OutputMode: httprequesthelper.OutputModeConsole,
	}

	logger, err := NewRequestLogger(config)
	require.NoError(t, err)
	require.NotNil(t, logger)

	router := gin.New()
	router.Use(logger.Middleware())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "test response")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "test response", w.Body.String())
}

func TestRequestLogger_Middleware_Enabled(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	config := &httprequesthelper.Config{
		Enabled:     true,
		OutputMode:  httprequesthelper.OutputModeConsole,
		MaxBodySize: 1024,
	}

	logger, err := NewRequestLogger(config)
	require.NoError(t, err)
	require.NotNil(t, logger)

	defer logger.Close()

	router := gin.New()
	router.Use(logger.Middleware())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "test response")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "test response", w.Body.String())
}

func TestRequestLogger_Middleware_WithRequestBody(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	config := &httprequesthelper.Config{
		Enabled:     true,
		OutputMode:  httprequesthelper.OutputModeConsole,
		MaxBodySize: 1024,
	}

	logger, err := NewRequestLogger(config)
	require.NoError(t, err)
	require.NotNil(t, logger)

	defer logger.Close()

	router := gin.New()
	router.Use(logger.Middleware())
	router.POST("/test", func(c *gin.Context) {
		body := c.Request.Body
		data := make([]byte, 100)
		n, _ := body.Read(data)
		c.String(http.StatusOK, "received: "+string(data[:n]))
	})

	req := httptest.NewRequest("POST", "/test", httptest.NewRequest("POST", "/test", nil).Body)
	req.Body = httptest.NewRequest("POST", "/test", nil).Body
	req = httptest.NewRequest("POST", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDefaultMiddleware(t *testing.T) {
	// 不使用 t.Parallel(): 修改 singleton defaultRequestLogger
	gin.SetMode(gin.TestMode)

	// Reset and initialize default logger
	defaultRequestLogger = nil
	defaultRequestLoggerOnce = *(new(sync.Once))

	config := &httprequesthelper.Config{
		Enabled:    true,
		OutputMode: httprequesthelper.OutputModeConsole,
	}

	err := InitDefaultRequestLogger(config)
	require.NoError(t, err)

	router := gin.New()
	router.Use(DefaultMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "test response")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "test response", w.Body.String())
}

func TestDefaultMiddleware_NotInitialized(t *testing.T) {
	// 不使用 t.Parallel(): 修改 singleton defaultRequestLogger
	gin.SetMode(gin.TestMode)

	// Reset default logger
	defaultRequestLogger = nil
	defaultRequestLoggerOnce = *(new(sync.Once))

	router := gin.New()
	router.Use(DefaultMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "test response")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// Should panic
	assert.Panics(t, func() {
		router.ServeHTTP(w, req)
	})
}

func TestRequestLogger_Middleware_WithErrorResponse(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	config := &httprequesthelper.Config{
		Enabled:    true,
		OutputMode: httprequesthelper.OutputModeConsole,
	}

	logger, err := NewRequestLogger(config)
	require.NoError(t, err)
	require.NotNil(t, logger)

	defer logger.Close()

	router := gin.New()
	router.Use(logger.Middleware())
	router.GET("/error", func(c *gin.Context) {
		c.String(http.StatusInternalServerError, "error occurred")
	})

	req := httptest.NewRequest("GET", "/error", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "error occurred", w.Body.String())
}

func TestRequestLogger_Middleware_WithHeaders(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	config := &httprequesthelper.Config{
		Enabled:        true,
		OutputMode:     httprequesthelper.OutputModeConsole,
		IncludeHeaders: true,
	}

	logger, err := NewRequestLogger(config)
	require.NoError(t, err)
	require.NotNil(t, logger)

	defer logger.Close()

	router := gin.New()
	router.Use(logger.Middleware())
	router.GET("/test", func(c *gin.Context) {
		c.Header("X-Custom-Header", "custom-value")
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("User-Agent", "test-agent")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "custom-value", w.Header().Get("X-Custom-Header"))
}

package apimiddleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCheckAgentUsePms_ReturnsHandler(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	// Just verify the function returns a handler
	handler := CheckAgentUsePms()
	assert.NotNil(t, handler)
}

func TestCheckAgentUsePmsInternal_ReturnsHandler(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	// Just verify the function returns a handler
	handler := CheckAgentUsePmsInternal()
	assert.NotNil(t, handler)
}

func TestCheckAgentUsePms_MissingAgentID(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/test", CheckAgentUsePms(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{})
	})

	// Create request without agent_id or agent_key
	req := httptest.NewRequest("POST", "/test", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return an error status
	assert.NotEqual(t, http.StatusOK, w.Code)
}

func TestCheckAgentUsePmsInternal_MissingAgentID(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/test", CheckAgentUsePmsInternal(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{})
	})

	// Create request without agent_id or agent_key
	req := httptest.NewRequest("POST", "/test", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return an error status
	assert.NotEqual(t, http.StatusOK, w.Code)
}

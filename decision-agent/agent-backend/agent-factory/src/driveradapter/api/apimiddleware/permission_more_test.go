package apimiddleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/decision-agent/agent-factory/conf"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/global"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setDisablePmsCheckForMiddlewareTest(t *testing.T, disable bool) {
	t.Helper()

	oldCfg := global.GConfig
	global.GConfig = &conf.Config{
		SwitchFields: conf.NewSwitchFields(),
	}
	global.GConfig.SwitchFields.DisablePmsCheck = disable

	t.Cleanup(func() {
		global.GConfig = oldCfg
	})
}

// ==================== CheckAgentUsePms — additional error paths ====================

func TestCheckAgentUsePms_BadJSON(t *testing.T) {
	t.Parallel()

	router := gin.New()
	router.POST("/test", CheckAgentUsePms(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{})
	})

	req := httptest.NewRequest("POST", "/test", strings.NewReader("not valid json"))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.NotEqual(t, http.StatusOK, w.Code)
}

func TestCheckAgentUsePms_AgentIDFromBody(t *testing.T) {
	t.Parallel()

	router := gin.New()
	router.POST("/test", CheckAgentUsePms(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{})
	})

	// Request with agent_id in body — will fail at visitor check (no visitor in context)
	req := httptest.NewRequest("POST", "/test", strings.NewReader(`{"agent_id":"agent-1"}`))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	// Should return 401 (no visitor)
	assert.NotEqual(t, http.StatusOK, w.Code)
}

func TestCheckAgentUsePms_AgentIDFromParam(t *testing.T) {
	t.Parallel()

	router := gin.New()
	router.POST("/test/:agent_id", CheckAgentUsePms(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{})
	})

	// agent_id not in body but in path param — will fail at visitor check
	req := httptest.NewRequest("POST", "/test/agent-1", strings.NewReader(`{"other":"val"}`))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.NotEqual(t, http.StatusOK, w.Code)
}

// ==================== CheckAgentUsePmsInternal — additional error paths ====================

func TestCheckAgentUsePmsInternal_BadJSON(t *testing.T) {
	t.Parallel()

	router := gin.New()
	router.POST("/test", CheckAgentUsePmsInternal(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{})
	})

	req := httptest.NewRequest("POST", "/test", strings.NewReader("bad json"))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.NotEqual(t, http.StatusOK, w.Code)
}

func TestCheckAgentUsePmsInternal_AgentIDFromBody_MissingUser(t *testing.T) {
	t.Parallel()

	router := gin.New()
	router.POST("/test", CheckAgentUsePmsInternal(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{})
	})

	// agent_id present but no x-account-id header
	req := httptest.NewRequest("POST", "/test", strings.NewReader(`{"agent_id":"agent-1"}`))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	// Should return 401 (user not found)
	assert.NotEqual(t, http.StatusOK, w.Code)
}

func TestCheckAgentUsePmsInternal_AgentIDFromParam_MissingUser(t *testing.T) {
	t.Parallel()

	router := gin.New()
	router.POST("/test/:agent_id", CheckAgentUsePmsInternal(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{})
	})

	req := httptest.NewRequest("POST", "/test/agent-1", strings.NewReader(`{"other":"val"}`))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.NotEqual(t, http.StatusOK, w.Code)
}

func TestCheckAgentUsePmsInternal_BadAccountType(t *testing.T) {
	t.Parallel()

	router := gin.New()
	router.POST("/test", CheckAgentUsePmsInternal(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{})
	})

	req := httptest.NewRequest("POST", "/test", strings.NewReader(`{"agent_id":"agent-1"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-account-id", "user-1")
	req.Header.Set("x-account-type", "invalid_type")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.NotEqual(t, http.StatusOK, w.Code)
}

func TestCheckAgentUsePmsInternal_DisablePmsCheck_AllowsMissingUser(t *testing.T) {
	setDisablePmsCheckForMiddlewareTest(t, true)

	router := gin.New()
	router.POST("/test", CheckAgentUsePmsInternal(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest("POST", "/test", strings.NewReader(`{"agent_id":"agent-1"}`))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

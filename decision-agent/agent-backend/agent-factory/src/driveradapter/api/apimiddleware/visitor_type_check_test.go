package apimiddleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/stretchr/testify/assert"
)

// ==================== IsUserType ====================

func TestIsUserType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		vType    rest.VisitorType
		expected bool
	}{
		{"VisitorType_User", rest.VisitorType_User, true},
		{"VisitorType_RealName", rest.VisitorType_RealName, true},
		{"VisitorType_App", rest.VisitorType_App, false},
		{"VisitorType_Anonymous", rest.VisitorType_Anonymous, false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := IsUserType(tt.vType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// ==================== VisitorTypeCheck ====================

func newVisitorTypeTestCtx(path string) (*gin.Context, *httptest.ResponseRecorder) { //nolint:unused
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)

	req := httptest.NewRequest(http.MethodPost, path, nil)
	c.Request = req

	return c, recorder
}

func setVisitorInCtx(c *gin.Context, id string, visitorType rest.VisitorType) {
	visitor := &rest.Visitor{
		ID:   id,
		Type: visitorType,
	}
	ctxKey := cenum.VisitUserInfoCtxKey.String()
	c.Set(ctxKey, visitor)

	_ctx := context.WithValue(c.Request.Context(), ctxKey, visitor) //nolint:staticcheck
	c.Request = c.Request.WithContext(_ctx)
}

func TestVisitorTypeCheck_AllowedPath(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	nextCalled := false
	router := gin.New()
	router.POST("/api/agent-factory/v3/agent-permission/execute", VisitorTypeCheck(), func(c *gin.Context) {
		nextCalled = true

		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/api/agent-factory/v3/agent-permission/execute", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.True(t, nextCalled)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestVisitorTypeCheck_UserType_Allowed(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	nextCalled := false
	router := gin.New()
	// 先设置 visitor middleware
	router.Use(func(c *gin.Context) {
		setVisitorInCtx(c, "user-1", rest.VisitorType_User)
		c.Next()
	})
	router.POST("/test", VisitorTypeCheck(), func(c *gin.Context) {
		nextCalled = true

		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.True(t, nextCalled)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestVisitorTypeCheck_AppType_Forbidden(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		setVisitorInCtx(c, "app-1", rest.VisitorType_App)
		c.Next()
	})
	router.POST("/test", VisitorTypeCheck(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.NotEqual(t, http.StatusOK, w.Code)
}

func TestVisitorTypeCheck_NilVisitor_Allowed(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	nextCalled := false
	router := gin.New()
	router.POST("/test", VisitorTypeCheck(), func(c *gin.Context) {
		nextCalled = true

		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.True(t, nextCalled)
}

// ==================== CheckAgentUsePms Error Branches ====================

func TestCheckAgentUsePms_InvalidJSON(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/test", CheckAgentUsePms(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{})
	})

	req := httptest.NewRequest("POST", "/test", strings.NewReader("not-json"))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.NotEqual(t, http.StatusOK, w.Code)
}

func TestCheckAgentUsePmsInternal_InvalidJSON(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/test", CheckAgentUsePmsInternal(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{})
	})

	req := httptest.NewRequest("POST", "/test", strings.NewReader("not-json"))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.NotEqual(t, http.StatusOK, w.Code)
}

func TestCheckAgentUsePmsInternal_NoUserID(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/test", CheckAgentUsePmsInternal(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{})
	})

	body := `{"agent_id":"agent-1"}`
	req := httptest.NewRequest("POST", "/test", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	// 没有 x-account-id header

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCheckAgentUsePmsInternal_InvalidAccountType(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/test", CheckAgentUsePmsInternal(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{})
	})

	body := `{"agent_id":"agent-1"}`
	req := httptest.NewRequest("POST", "/test", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-account-id", "user-1")
	req.Header.Set("x-account-type", "invalid_type")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

package capimiddleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/cconf"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/conf"
	"github.com/stretchr/testify/assert"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdapmsenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/global"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// ==================== CheckPmsReq.IsAgentUseCheck ====================

func TestCheckPmsReq_IsAgentUseCheck_True(t *testing.T) {
	t.Parallel()

	req := &CheckPmsReq{
		ResourceType: cdaenum.ResourceTypeDataAgent,
		Operator:     cdapmsenum.AgentUse,
	}
	assert.True(t, req.IsAgentUseCheck())
}

func TestCheckPmsReq_IsAgentUseCheck_WrongType(t *testing.T) {
	t.Parallel()

	req := &CheckPmsReq{
		ResourceType: cdaenum.ResourceTypeDataAgentTpl,
		Operator:     cdapmsenum.AgentUse,
	}
	assert.False(t, req.IsAgentUseCheck())
}

func TestCheckPmsReq_IsAgentUseCheck_WrongOperator(t *testing.T) {
	t.Parallel()

	req := &CheckPmsReq{
		ResourceType: cdaenum.ResourceTypeDataAgent,
		Operator:     cdapmsenum.AgentPublish,
	}
	assert.False(t, req.IsAgentUseCheck())
}

// ==================== CheckPmsReq.ReqCheck ====================

func TestCheckPmsReq_ReqCheck_Valid(t *testing.T) {
	t.Parallel()

	req := &CheckPmsReq{
		ResourceType: cdaenum.ResourceTypeDataAgent,
		ResourceID:   "agent-1",
		Operator:     cdapmsenum.AgentUse,
		UserID:       "user-1",
	}
	assert.NoError(t, req.ReqCheck())
}

func TestCheckPmsReq_ReqCheck_InvalidResourceType(t *testing.T) {
	t.Parallel()

	req := &CheckPmsReq{
		ResourceType: cdaenum.ResourceType("invalid"),
		ResourceID:   "agent-1",
		Operator:     cdapmsenum.AgentUse,
		UserID:       "user-1",
	}
	assert.Error(t, req.ReqCheck())
}

func TestCheckPmsReq_ReqCheck_InvalidOperator(t *testing.T) {
	t.Parallel()

	req := &CheckPmsReq{
		ResourceType: cdaenum.ResourceTypeDataAgent,
		ResourceID:   "agent-1",
		Operator:     cdapmsenum.Operator("invalid"),
		UserID:       "user-1",
	}
	assert.Error(t, req.ReqCheck())
}

func TestCheckPmsReq_ReqCheck_EmptyUserAndApp(t *testing.T) {
	t.Parallel()

	req := &CheckPmsReq{
		ResourceType: cdaenum.ResourceTypeDataAgent,
		ResourceID:   "agent-1",
		Operator:     cdapmsenum.AgentUse,
		UserID:       "",
		AppAccountID: "",
	}
	assert.Error(t, req.ReqCheck())
}

func TestCheckPmsReq_ReqCheck_EmptyResourceID(t *testing.T) {
	t.Parallel()

	req := &CheckPmsReq{
		ResourceType: cdaenum.ResourceTypeDataAgent,
		ResourceID:   "",
		Operator:     cdapmsenum.AgentUse,
		UserID:       "user-1",
	}
	assert.Error(t, req.ReqCheck())
}

func TestCheckPmsReq_ReqCheck_AppAccountOnly(t *testing.T) {
	t.Parallel()

	req := &CheckPmsReq{
		ResourceType: cdaenum.ResourceTypeDataAgent,
		ResourceID:   "agent-1",
		Operator:     cdapmsenum.AgentUse,
		AppAccountID: "app-1",
	}
	assert.NoError(t, req.ReqCheck())
}

// ==================== NewCheckAgentUsePmsReq ====================

func TestNewCheckAgentUsePmsReq(t *testing.T) {
	t.Parallel()

	req := NewCheckAgentUsePmsReq("agent-1", "user-1", "app-1")
	assert.Equal(t, cdaenum.ResourceTypeDataAgent, req.ResourceType)
	assert.Equal(t, "agent-1", req.ResourceID)
	assert.Equal(t, cdapmsenum.AgentUse, req.Operator)
	assert.Equal(t, "user-1", req.UserID)
	assert.Equal(t, "app-1", req.AppAccountID)
}

// ==================== HandleBizDomain ====================

func newBizDomainCtx(method, path string, headers map[string]string) (*gin.Context, *httptest.ResponseRecorder) {
	recorder := httptest.NewRecorder()

	req, _ := http.NewRequest(method, path, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	c, _ := gin.CreateTestContext(recorder)
	c.Request = req

	return c, recorder
}

func TestHandleBizDomain_WithBizDomainID(t *testing.T) {
	t.Parallel()

	handler := HandleBizDomain(false)
	c, recorder := newBizDomainCtx(http.MethodGet, "/", map[string]string{
		"x-business-domain": "bd-123",
	})

	handler(c)
	assert.Equal(t, http.StatusOK, recorder.Code)

	ctxKey := cenum.BizDomainIDCtxKey.String()
	val, exists := c.Get(ctxKey)
	assert.True(t, exists)
	assert.Equal(t, "bd-123", val)
}

func TestHandleBizDomain_MissingBizDomain_NoDefault(t *testing.T) {
	t.Parallel()

	handler := HandleBizDomain(false)
	c, recorder := newBizDomainCtx(http.MethodGet, "/", map[string]string{})

	handler(c)
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestHandleBizDomain_MissingBizDomain_UseDefault(t *testing.T) {
	t.Parallel()

	handler := HandleBizDomain(true)
	c, recorder := newBizDomainCtx(http.MethodGet, "/", map[string]string{})

	handler(c)
	assert.Equal(t, http.StatusOK, recorder.Code)

	ctxKey := cenum.BizDomainIDCtxKey.String()
	val, exists := c.Get(ctxKey)
	assert.True(t, exists)
	assert.Equal(t, cenum.BizDomainPublic.ToString(), val)
}

func TestHandleBizDomain_DisabledWithoutHeader(t *testing.T) {
	oldCfg := global.GConfig
	global.GConfig = &conf.Config{
		Config:       cconf.BaseDefConfig(),
		SwitchFields: conf.NewSwitchFields(),
	}
	global.GConfig.SwitchFields.DisableBizDomain = true

	t.Cleanup(func() {
		global.GConfig = oldCfg
	})

	handler := HandleBizDomain(false)
	c, recorder := newBizDomainCtx(http.MethodGet, "/", map[string]string{})

	handler(c)
	assert.Equal(t, http.StatusOK, recorder.Code)

	ctxKey := cenum.BizDomainIDCtxKey.String()
	_, exists := c.Get(ctxKey)
	assert.False(t, exists)
	assert.Empty(t, c.Request.Context().Value(ctxKey))
}

func TestHandleBizDomain_DisabledWithHeader(t *testing.T) {
	oldCfg := global.GConfig
	global.GConfig = &conf.Config{
		Config:       cconf.BaseDefConfig(),
		SwitchFields: conf.NewSwitchFields(),
	}
	global.GConfig.SwitchFields.DisableBizDomain = true

	t.Cleanup(func() {
		global.GConfig = oldCfg
	})

	handler := HandleBizDomain(true)
	c, recorder := newBizDomainCtx(http.MethodGet, "/", map[string]string{
		"x-business-domain": "bd-123",
	})

	handler(c)
	assert.Equal(t, http.StatusOK, recorder.Code)

	ctxKey := cenum.BizDomainIDCtxKey.String()
	_, exists := c.Get(ctxKey)
	assert.False(t, exists)
	assert.Empty(t, c.Request.Context().Value(ctxKey))
}

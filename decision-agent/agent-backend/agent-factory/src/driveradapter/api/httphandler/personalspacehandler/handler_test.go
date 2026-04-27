package personalspacehandler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/personal_space/personalspaceresp"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driver/iv3portdriver/v3portdrivermock"
)

// testLogger 实现 icmp.Logger, 用于测试
type testLogger struct{}

func (testLogger) Infof(string, ...interface{})  {}
func (testLogger) Infoln(...interface{})         {}
func (testLogger) Debugf(string, ...interface{}) {}
func (testLogger) Debugln(...interface{})        {}
func (testLogger) Errorf(string, ...interface{}) {}
func (testLogger) Errorln(...interface{})        {}
func (testLogger) Warnf(string, ...interface{})  {}
func (testLogger) Warnln(...interface{})         {}
func (testLogger) Panicf(string, ...interface{}) {}
func (testLogger) Panicln(...interface{})        {}
func (testLogger) Fatalf(string, ...interface{}) {}
func (testLogger) Fatalln(...interface{})        {}

func newTestCtx(method, target string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)

	req := httptest.NewRequest(method, target, nil)
	c.Request = req

	return c, recorder
}

// ==================== RegPubRouter ====================

func TestPersonalSpaceHandler_RegPubRouter(t *testing.T) {
	t.Parallel()

	h := &PersonalSpaceHTTPHandler{}
	r := gin.New()
	h.RegPubRouter(r.Group("/v1"))
	routes := r.Routes()

	foundAgentList := false
	foundTplList := false

	for _, route := range routes {
		if route.Method == http.MethodGet && route.Path == "/v1/personal-space/agent-list" {
			foundAgentList = true
		}

		if route.Method == http.MethodGet && route.Path == "/v1/personal-space/agent-tpl-list" {
			foundTplList = true
		}
	}

	assert.True(t, foundAgentList, "should have agent-list route")
	assert.True(t, foundTplList, "should have agent-tpl-list route")
}

// ==================== AgentList ====================

func TestPersonalSpaceHandler_AgentList(t *testing.T) {
	t.Parallel()

	t.Run("bind error - size exceeds max", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockSvc := v3portdrivermock.NewMockIPersonalSpaceService(ctrl)
		h := &PersonalSpaceHTTPHandler{personalSpaceService: mockSvc, logger: testLogger{}}

		c, _ := newTestCtx(http.MethodGet, "/personal-space/agent-list?size=9999")
		h.AgentList(c)
		assert.True(t, len(c.Errors) > 0, "should have errors in context")
	})

	t.Run("custom check error - invalid publish_status", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockSvc := v3portdrivermock.NewMockIPersonalSpaceService(ctrl)
		h := &PersonalSpaceHTTPHandler{personalSpaceService: mockSvc, logger: testLogger{}}

		c, _ := newTestCtx(http.MethodGet, "/personal-space/agent-list?size=10&publish_status=invalid_status")
		h.AgentList(c)
		assert.True(t, len(c.Errors) > 0, "should have errors in context")
	})

	t.Run("load marker error - invalid marker", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockSvc := v3portdrivermock.NewMockIPersonalSpaceService(ctrl)
		h := &PersonalSpaceHTTPHandler{personalSpaceService: mockSvc, logger: testLogger{}}

		c, recorder := newTestCtx(http.MethodGet, "/personal-space/agent-list?size=10&pagination_marker_str=not-valid-base64")
		h.AgentList(c)
		assert.NotEqual(t, http.StatusOK, recorder.Code)
	})

	t.Run("service error", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockSvc := v3portdrivermock.NewMockIPersonalSpaceService(ctrl)
		h := &PersonalSpaceHTTPHandler{personalSpaceService: mockSvc, logger: testLogger{}}

		mockSvc.EXPECT().AgentList(gomock.Any(), gomock.Any()).
			Return(nil, errors.New("service failed"))

		c, _ := newTestCtx(http.MethodGet, "/personal-space/agent-list?size=10")
		h.AgentList(c)
		assert.True(t, len(c.Errors) > 0, "should have errors in context")
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockSvc := v3portdrivermock.NewMockIPersonalSpaceService(ctrl)
		h := &PersonalSpaceHTTPHandler{personalSpaceService: mockSvc, logger: testLogger{}}

		mockSvc.EXPECT().AgentList(gomock.Any(), gomock.Any()).
			Return(&personalspaceresp.AgentListResp{}, nil)

		c, recorder := newTestCtx(http.MethodGet, "/personal-space/agent-list?size=10")
		h.AgentList(c)
		assert.Equal(t, http.StatusOK, recorder.Code)
	})
}

// ==================== AgentTplList ====================

func TestPersonalSpaceHandler_AgentTplList(t *testing.T) {
	t.Parallel()

	t.Run("bind error - size exceeds max", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockSvc := v3portdrivermock.NewMockIPersonalSpaceService(ctrl)
		h := &PersonalSpaceHTTPHandler{personalSpaceService: mockSvc, logger: testLogger{}}

		c, recorder := newTestCtx(http.MethodGet, "/personal-space/agent-tpl-list?size=9999")
		h.AgentTplList(c)
		assert.NotEqual(t, http.StatusOK, recorder.Code)
	})

	t.Run("custom check error - invalid publish_status", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockSvc := v3portdrivermock.NewMockIPersonalSpaceService(ctrl)
		h := &PersonalSpaceHTTPHandler{personalSpaceService: mockSvc, logger: testLogger{}}

		c, recorder := newTestCtx(http.MethodGet, "/personal-space/agent-tpl-list?size=10&publish_status=invalid_status")
		h.AgentTplList(c)
		assert.NotEqual(t, http.StatusOK, recorder.Code)
	})

	t.Run("load marker error - invalid marker", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockSvc := v3portdrivermock.NewMockIPersonalSpaceService(ctrl)
		h := &PersonalSpaceHTTPHandler{personalSpaceService: mockSvc, logger: testLogger{}}

		c, recorder := newTestCtx(http.MethodGet, "/personal-space/agent-tpl-list?size=10&pagination_marker_str=not-valid-base64")
		h.AgentTplList(c)
		assert.NotEqual(t, http.StatusOK, recorder.Code)
	})

	t.Run("service error", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockSvc := v3portdrivermock.NewMockIPersonalSpaceService(ctrl)
		h := &PersonalSpaceHTTPHandler{personalSpaceService: mockSvc, logger: testLogger{}}

		mockSvc.EXPECT().AgentTplList(gomock.Any(), gomock.Any()).
			Return(nil, errors.New("service failed"))

		c, recorder := newTestCtx(http.MethodGet, "/personal-space/agent-tpl-list?size=10")
		h.AgentTplList(c)
		assert.NotEqual(t, http.StatusOK, recorder.Code)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockSvc := v3portdrivermock.NewMockIPersonalSpaceService(ctrl)
		h := &PersonalSpaceHTTPHandler{personalSpaceService: mockSvc, logger: testLogger{}}

		mockSvc.EXPECT().AgentTplList(gomock.Any(), gomock.Any()).
			Return(&personalspaceresp.AgentTplListResp{}, nil)

		c, recorder := newTestCtx(http.MethodGet, "/personal-space/agent-tpl-list?size=10")
		h.AgentTplList(c)
		assert.Equal(t, http.StatusOK, recorder.Code)
	})
}

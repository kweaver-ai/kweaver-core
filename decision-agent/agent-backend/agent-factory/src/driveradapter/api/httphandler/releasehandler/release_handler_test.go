package releasehandler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/auditlogdto"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/release/releasereq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/release/releaseresp"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driver/iv3portdriver/v3portdrivermock"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

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

func newTestContext(method, target, body string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)

	req := httptest.NewRequest(method, target, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	c.Request = req

	return c, recorder
}

func setInternalAPI(c *gin.Context, isInternal bool) {
	c.Set(cenum.InternalAPIFlagCtxKey.String(), isInternal)
	ctx := context.WithValue(c.Request.Context(), cenum.InternalAPIFlagCtxKey.String(), isInternal) //nolint:staticcheck // SA1029
	c.Request = c.Request.WithContext(ctx)
}

func setVisitor(c *gin.Context, userID string, withRequestContext bool) {
	visitor := &rest.Visitor{ID: userID}
	c.Set(cenum.VisitUserInfoCtxKey.String(), visitor)

	if withRequestContext {
		ctx := context.WithValue(c.Request.Context(), cenum.VisitUserInfoCtxKey.String(), visitor) //nolint:staticcheck // SA1029
		c.Request = c.Request.WithContext(ctx)
	}
}

func routeExists(routes []gin.RouteInfo, method, path string) bool {
	for _, r := range routes {
		if r.Method == method && r.Path == path {
			return true
		}
	}

	return false
}

func TestSetIsPrivate2Req(t *testing.T) {
	t.Parallel()

	c, _ := newTestContext(http.MethodPost, "/", "")
	req := releasereq.NewPublishReq()

	setInternalAPI(c, true)
	setIsPrivate2Req(c, req)
	assert.True(t, req.IsInternalAPI)

	setInternalAPI(c, false)
	setIsPrivate2Req(c, req)
	assert.False(t, req.IsInternalAPI)
}

func TestReleaseHandler_RegRouters(t *testing.T) {
	t.Parallel()

	h := &releaseHandler{}

	pubRouter := gin.New()
	h.RegPubRouter(pubRouter.Group("/pub"))
	pubRoutes := pubRouter.Routes()
	assert.True(t, routeExists(pubRoutes, http.MethodPost, "/pub/agent/:agent_id/publish"))
	assert.True(t, routeExists(pubRoutes, http.MethodPut, "/pub/agent/:agent_id/unpublish"))
	assert.True(t, routeExists(pubRoutes, http.MethodGet, "/pub/agent/:agent_id/release-history"))
	assert.True(t, routeExists(pubRoutes, http.MethodGet, "/pub/agent/:agent_id/publish-info"))
	assert.True(t, routeExists(pubRoutes, http.MethodPut, "/pub/agent/:agent_id/publish-info"))

	priRouter := gin.New()
	h.RegPriRouter(priRouter.Group("/pri"))
	priRoutes := priRouter.Routes()
	assert.True(t, routeExists(priRoutes, http.MethodPost, "/pri/agent/:agent_id/publish"))
	assert.True(t, routeExists(priRoutes, http.MethodGet, "/pri/agent/:agent_id/publish-info"))
}

func TestReleaseHandler_Publish(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSvc := v3portdrivermock.NewMockIReleaseSvc(ctrl)
		h := &releaseHandler{releaseSvc: mockSvc, logger: testLogger{}}

		mockSvc.EXPECT().Publish(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, req *releasereq.PublishReq) (*releaseresp.PublishUpsertResp, auditlogdto.AgentPublishAuditLogInfo, error) {
			assert.Equal(t, "agent-1", req.AgentID)
			assert.Equal(t, "user-1", req.UserID)
			assert.True(t, req.IsInternalAPI)

			return &releaseresp.PublishUpsertResp{ReleaseId: "r-1"}, auditlogdto.AgentPublishAuditLogInfo{ID: "agent-1", Name: "name"}, nil
		})

		c, recorder := newTestContext(http.MethodPost, "/", `{"category_ids":["cat-1"],"publish_to_where":["square"],"publish_to_bes":["api_agent"]}`)
		c.Params = gin.Params{{Key: "agent_id", Value: "agent-1"}}
		setInternalAPI(c, true)
		setVisitor(c, "user-1", false)

		h.Publish(c)
		assert.Equal(t, http.StatusCreated, recorder.Code)
	})

	t.Run("user id empty", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSvc := v3portdrivermock.NewMockIReleaseSvc(ctrl)
		h := &releaseHandler{releaseSvc: mockSvc, logger: testLogger{}}

		c, _ := newTestContext(http.MethodPost, "/", "{}")
		c.Params = gin.Params{{Key: "agent_id", Value: "agent-1"}}
		setInternalAPI(c, true)
		setVisitor(c, "", false)

		h.Publish(c)
		assert.NotEmpty(t, c.Errors)
	})

	t.Run("bind error", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSvc := v3portdrivermock.NewMockIReleaseSvc(ctrl)
		h := &releaseHandler{releaseSvc: mockSvc, logger: testLogger{}}

		c, _ := newTestContext(http.MethodPost, "/", "{")
		c.Params = gin.Params{{Key: "agent_id", Value: "agent-1"}}
		setInternalAPI(c, true)
		setVisitor(c, "user-1", false)

		h.Publish(c)
		assert.NotEmpty(t, c.Errors)
	})

	t.Run("custom check error", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSvc := v3portdrivermock.NewMockIReleaseSvc(ctrl)
		h := &releaseHandler{releaseSvc: mockSvc, logger: testLogger{}}

		c, recorder := newTestContext(http.MethodPost, "/", `{"category_ids":["cat-1"],"publish_to_where":["invalid"]}`)
		c.Params = gin.Params{{Key: "agent_id", Value: "agent-1"}}
		setInternalAPI(c, true)
		setVisitor(c, "user-1", false)

		h.Publish(c)
		assert.NotEqual(t, http.StatusCreated, recorder.Code)
	})

	t.Run("missing category ids does not panic", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSvc := v3portdrivermock.NewMockIReleaseSvc(ctrl)
		h := &releaseHandler{releaseSvc: mockSvc, logger: testLogger{}}

		mockSvc.EXPECT().Publish(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, req *releasereq.PublishReq) (*releaseresp.PublishUpsertResp, auditlogdto.AgentPublishAuditLogInfo, error) {
			assert.Empty(t, req.CategoryIDs)
			assert.Empty(t, req.PublishToWhere)

			return &releaseresp.PublishUpsertResp{ReleaseId: "r-1"}, auditlogdto.AgentPublishAuditLogInfo{ID: "agent-1", Name: "name"}, nil
		})

		c, _ := newTestContext(http.MethodPost, "/", `{}`)
		c.Params = gin.Params{{Key: "agent_id", Value: "agent-1"}}
		setInternalAPI(c, true)
		setVisitor(c, "user-1", false)

		assert.NotPanics(t, func() { h.Publish(c) })
		assert.Empty(t, c.Errors)
	})

	t.Run("optional fields can be omitted", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSvc := v3portdrivermock.NewMockIReleaseSvc(ctrl)
		h := &releaseHandler{releaseSvc: mockSvc, logger: testLogger{}}

		mockSvc.EXPECT().Publish(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, req *releasereq.PublishReq) (*releaseresp.PublishUpsertResp, auditlogdto.AgentPublishAuditLogInfo, error) {
			assert.Equal(t, []string{"cat-1"}, req.CategoryIDs)
			assert.Empty(t, req.PublishToWhere)
			assert.Nil(t, req.PmsControl)
			assert.Empty(t, req.PublishToBes)

			return &releaseresp.PublishUpsertResp{ReleaseId: "r-1"}, auditlogdto.AgentPublishAuditLogInfo{ID: "agent-1", Name: "name"}, nil
		})

		c, recorder := newTestContext(http.MethodPost, "/", `{"category_ids":["cat-1"]}`)
		c.Params = gin.Params{{Key: "agent_id", Value: "agent-1"}}
		setInternalAPI(c, true)
		setVisitor(c, "user-1", false)

		h.Publish(c)
		assert.Equal(t, http.StatusCreated, recorder.Code)
	})

	t.Run("custom space publish target is rejected", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSvc := v3portdrivermock.NewMockIReleaseSvc(ctrl)
		h := &releaseHandler{releaseSvc: mockSvc, logger: testLogger{}}

		c, recorder := newTestContext(http.MethodPost, "/", `{"publish_to_where":["custom_space"]}`)
		c.Params = gin.Params{{Key: "agent_id", Value: "agent-1"}}
		setInternalAPI(c, true)
		setVisitor(c, "user-1", false)

		h.Publish(c)
		assert.NotEqual(t, http.StatusCreated, recorder.Code)
		assert.NotEmpty(t, c.Errors)
	})

	t.Run("service error", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSvc := v3portdrivermock.NewMockIReleaseSvc(ctrl)
		h := &releaseHandler{releaseSvc: mockSvc, logger: testLogger{}}

		mockSvc.EXPECT().Publish(gomock.Any(), gomock.Any()).Return(nil, auditlogdto.AgentPublishAuditLogInfo{}, errors.New("publish failed"))

		c, _ := newTestContext(http.MethodPost, "/", `{"category_ids":["cat-1"],"publish_to_where":["square"],"publish_to_bes":["api_agent"]}`)
		c.Params = gin.Params{{Key: "agent_id", Value: "agent-1"}}
		setInternalAPI(c, true)
		setVisitor(c, "user-1", false)

		h.Publish(c)
		assert.NotEmpty(t, c.Errors)
	})

	t.Run("missing category ids enters service", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSvc := v3portdrivermock.NewMockIReleaseSvc(ctrl)
		h := &releaseHandler{releaseSvc: mockSvc, logger: testLogger{}}

		mockSvc.EXPECT().UpdatePublishInfo(gomock.Any(), "agent-1", gomock.Any()).DoAndReturn(func(ctx context.Context, agentID string, req *releasereq.UpdatePublishInfoReq) (*releaseresp.PublishUpsertResp, auditlogdto.AgentModifyPublishAuditLogInfo, error) {
			assert.Empty(t, req.CategoryIDs)
			assert.Empty(t, req.PublishToWhere)

			return &releaseresp.PublishUpsertResp{ReleaseId: "r-1"}, auditlogdto.AgentModifyPublishAuditLogInfo{ID: "agent-1"}, nil
		})

		c, recorder := newTestContext(http.MethodPut, "/", `{}`)
		c.Params = gin.Params{{Key: "agent_id", Value: "agent-1"}}
		setInternalAPI(c, true)
		h.UpdatePublishInfo(c)
		assert.Equal(t, http.StatusOK, recorder.Code)
	})

	t.Run("custom space publish target is rejected", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSvc := v3portdrivermock.NewMockIReleaseSvc(ctrl)
		h := &releaseHandler{releaseSvc: mockSvc, logger: testLogger{}}

		c, _ := newTestContext(http.MethodPut, "/", `{"publish_to_where":["custom_space"]}`)
		c.Params = gin.Params{{Key: "agent_id", Value: "agent-1"}}
		setInternalAPI(c, true)
		h.UpdatePublishInfo(c)
		assert.NotEmpty(t, c.Errors)
	})
}

func TestReleaseHandler_UnPublish(t *testing.T) {
	t.Parallel()

	t.Run("agent id empty", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSvc := v3portdrivermock.NewMockIReleaseSvc(ctrl)
		h := &releaseHandler{releaseSvc: mockSvc, logger: testLogger{}}

		c, _ := newTestContext(http.MethodPut, "/", "")
		setInternalAPI(c, true)

		h.UnPublish(c)
		assert.NotEmpty(t, c.Errors)
	})

	t.Run("service error", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSvc := v3portdrivermock.NewMockIReleaseSvc(ctrl)
		h := &releaseHandler{releaseSvc: mockSvc, logger: testLogger{}}

		mockSvc.EXPECT().UnPublish(gomock.Any(), "agent-1").Return(auditlogdto.AgentUnPublishAuditLogInfo{}, errors.New("unpublish failed"))

		c, _ := newTestContext(http.MethodPut, "/", "")
		c.Params = gin.Params{{Key: "agent_id", Value: "agent-1"}}
		setInternalAPI(c, true)

		h.UnPublish(c)
		assert.NotEmpty(t, c.Errors)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSvc := v3portdrivermock.NewMockIReleaseSvc(ctrl)
		h := &releaseHandler{releaseSvc: mockSvc, logger: testLogger{}}

		mockSvc.EXPECT().UnPublish(gomock.Any(), "agent-1").Return(auditlogdto.AgentUnPublishAuditLogInfo{ID: "agent-1"}, nil)

		c, recorder := newTestContext(http.MethodPut, "/", "")
		c.Params = gin.Params{{Key: "agent_id", Value: "agent-1"}}
		setInternalAPI(c, true)

		h.UnPublish(c)
		assert.Equal(t, http.StatusNoContent, recorder.Code)
	})
}

func TestReleaseHandler_GetPublishInfo(t *testing.T) {
	t.Parallel()

	t.Run("agent id empty", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSvc := v3portdrivermock.NewMockIReleaseSvc(ctrl)
		h := &releaseHandler{releaseSvc: mockSvc, logger: testLogger{}}

		c, _ := newTestContext(http.MethodGet, "/", "")
		h.GetPublishInfo(c)
		assert.NotEmpty(t, c.Errors)
	})

	t.Run("service error", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSvc := v3portdrivermock.NewMockIReleaseSvc(ctrl)
		h := &releaseHandler{releaseSvc: mockSvc, logger: testLogger{}}

		mockSvc.EXPECT().GetPublishInfo(gomock.Any(), "agent-1").Return(nil, errors.New("get publish info failed"))

		c, _ := newTestContext(http.MethodGet, "/", "")
		c.Params = gin.Params{{Key: "agent_id", Value: "agent-1"}}
		h.GetPublishInfo(c)
		assert.NotEmpty(t, c.Errors)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSvc := v3portdrivermock.NewMockIReleaseSvc(ctrl)
		h := &releaseHandler{releaseSvc: mockSvc, logger: testLogger{}}

		mockSvc.EXPECT().GetPublishInfo(gomock.Any(), "agent-1").Return(&releaseresp.PublishInfoResp{}, nil)

		c, recorder := newTestContext(http.MethodGet, "/", "")
		c.Params = gin.Params{{Key: "agent_id", Value: "agent-1"}}
		h.GetPublishInfo(c)
		assert.Equal(t, http.StatusOK, recorder.Code)
	})
}

func TestReleaseHandler_UpdatePublishInfo(t *testing.T) {
	t.Parallel()

	t.Run("agent id empty", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSvc := v3portdrivermock.NewMockIReleaseSvc(ctrl)
		h := &releaseHandler{releaseSvc: mockSvc, logger: testLogger{}}

		c, _ := newTestContext(http.MethodPut, "/", "{}")
		setInternalAPI(c, true)
		h.UpdatePublishInfo(c)
		assert.NotEmpty(t, c.Errors)
	})

	t.Run("bind json error", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSvc := v3portdrivermock.NewMockIReleaseSvc(ctrl)
		h := &releaseHandler{releaseSvc: mockSvc, logger: testLogger{}}

		c, _ := newTestContext(http.MethodPut, "/", "{")
		c.Params = gin.Params{{Key: "agent_id", Value: "agent-1"}}
		setInternalAPI(c, true)
		h.UpdatePublishInfo(c)
		assert.NotEmpty(t, c.Errors)
	})

	t.Run("custom check error", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSvc := v3portdrivermock.NewMockIReleaseSvc(ctrl)
		h := &releaseHandler{releaseSvc: mockSvc, logger: testLogger{}}

		c, _ := newTestContext(http.MethodPut, "/", `{"category_ids":["cat-1"],"publish_to_bes":["invalid"]}`)
		c.Params = gin.Params{{Key: "agent_id", Value: "agent-1"}}
		setInternalAPI(c, true)
		h.UpdatePublishInfo(c)
		assert.NotEmpty(t, c.Errors)
	})

	t.Run("service error", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSvc := v3portdrivermock.NewMockIReleaseSvc(ctrl)
		h := &releaseHandler{releaseSvc: mockSvc, logger: testLogger{}}

		mockSvc.EXPECT().UpdatePublishInfo(gomock.Any(), "agent-1", gomock.Any()).Return(nil, auditlogdto.AgentModifyPublishAuditLogInfo{}, errors.New("update publish info failed"))

		c, _ := newTestContext(http.MethodPut, "/", `{"category_ids":["cat-1"]}`)
		c.Params = gin.Params{{Key: "agent_id", Value: "agent-1"}}
		setInternalAPI(c, true)
		h.UpdatePublishInfo(c)
		assert.NotEmpty(t, c.Errors)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSvc := v3portdrivermock.NewMockIReleaseSvc(ctrl)
		h := &releaseHandler{releaseSvc: mockSvc, logger: testLogger{}}

		mockSvc.EXPECT().UpdatePublishInfo(gomock.Any(), "agent-1", gomock.Any()).Return(&releaseresp.PublishUpsertResp{ReleaseId: "r-1"}, auditlogdto.AgentModifyPublishAuditLogInfo{ID: "agent-1"}, nil)

		c, recorder := newTestContext(http.MethodPut, "/", `{"category_ids":["cat-1"]}`)
		c.Params = gin.Params{{Key: "agent_id", Value: "agent-1"}}
		setInternalAPI(c, true)
		h.UpdatePublishInfo(c)
		assert.Equal(t, http.StatusOK, recorder.Code)
	})
}

func TestReleaseHandler_HistoryListAndInfo(t *testing.T) {
	t.Parallel()

	t.Run("history list agent id empty", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSvc := v3portdrivermock.NewMockIReleaseSvc(ctrl)
		h := &releaseHandler{releaseSvc: mockSvc, logger: testLogger{}}

		c, recorder := newTestContext(http.MethodGet, "/", "")
		h.HistoryList(c)
		assert.NotEqual(t, http.StatusOK, recorder.Code)
	})

	t.Run("history list service error", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSvc := v3portdrivermock.NewMockIReleaseSvc(ctrl)
		h := &releaseHandler{releaseSvc: mockSvc, logger: testLogger{}}

		mockSvc.EXPECT().GetPublishHistoryList(gomock.Any(), "agent-1").Return(nil, int64(0), errors.New("history failed"))

		c, recorder := newTestContext(http.MethodGet, "/", "")
		c.Params = gin.Params{{Key: "agent_id", Value: "agent-1"}}
		h.HistoryList(c)
		assert.NotEqual(t, http.StatusOK, recorder.Code)
	})

	t.Run("history list success", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSvc := v3portdrivermock.NewMockIReleaseSvc(ctrl)
		h := &releaseHandler{releaseSvc: mockSvc, logger: testLogger{}}

		mockSvc.EXPECT().GetPublishHistoryList(gomock.Any(), "agent-1").Return(releaseresp.HistoryListResp{{HistoryId: "h1"}}, int64(1), nil)

		c, recorder := newTestContext(http.MethodGet, "/", "")
		c.Params = gin.Params{{Key: "agent_id", Value: "agent-1"}}
		h.HistoryList(c)
		assert.Equal(t, http.StatusOK, recorder.Code)
	})

	t.Run("history info success", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockSvc := v3portdrivermock.NewMockIReleaseSvc(ctrl)
		h := &releaseHandler{releaseSvc: mockSvc, logger: testLogger{}}

		c, recorder := newTestContext(http.MethodGet, "/", "")
		h.HistoryInfo(c)
		assert.Equal(t, http.StatusOK, recorder.Code)
	})
}

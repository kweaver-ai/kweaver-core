package observabilityhandler

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

	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/conversation/conversationresp"
	observabilityresp "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/observability/resp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driver/iportdriver/iportdrivermock"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

type obsTestLogger struct{}

func (obsTestLogger) Infof(string, ...interface{})  {}
func (obsTestLogger) Infoln(...interface{})         {}
func (obsTestLogger) Debugf(string, ...interface{}) {}
func (obsTestLogger) Debugln(...interface{})        {}
func (obsTestLogger) Errorf(string, ...interface{}) {}
func (obsTestLogger) Errorln(...interface{})        {}
func (obsTestLogger) Warnf(string, ...interface{})  {}
func (obsTestLogger) Warnln(...interface{})         {}
func (obsTestLogger) Panicf(string, ...interface{}) {}
func (obsTestLogger) Panicln(...interface{})        {}
func (obsTestLogger) Fatalf(string, ...interface{}) {}
func (obsTestLogger) Fatalln(...interface{})        {}

func newObsCtx(method, target, body string) (*gin.Context, *httptest.ResponseRecorder) {
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

func setObsVisitor(c *gin.Context, userID string) {
	visitor := &rest.Visitor{ID: userID, Type: rest.VisitorType_User}
	c.Set(cenum.VisitUserInfoCtxKey.String(), visitor)
	ctx := context.WithValue(c.Request.Context(), cenum.VisitUserInfoCtxKey.String(), visitor) //nolint:staticcheck // SA1029
	c.Request = c.Request.WithContext(ctx)
}

// ==================== RegPubRouter / RegPriRouter ====================

func TestObservabilityHandler_RegPubRouter(t *testing.T) {
	t.Parallel()

	h := &observabilityHTTPHandler{}
	r := gin.New()
	h.RegPubRouter(r.Group("/v1"))
	routes := r.Routes()

	assert.NotEmpty(t, routes)
}

func TestObservabilityHandler_RegPriRouter(t *testing.T) {
	t.Parallel()

	h := &observabilityHTTPHandler{}
	r := gin.New()
	h.RegPriRouter(r.Group("/v1"))
	routes := r.Routes()

	assert.NotEmpty(t, routes)
}

// ==================== AgentDetail ====================

func TestObservabilityHandler_AgentDetail(t *testing.T) {
	t.Parallel()

	t.Run("agent_id empty", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockObsSvc := iportdrivermock.NewMockIObservability(ctrl)
		h := &observabilityHTTPHandler{observabilitySvc: mockObsSvc, logger: obsTestLogger{}}

		c, recorder := newObsCtx(http.MethodPost, "/", `{"start_time":1,"end_time":2}`)
		h.AgentDetail(c)
		assert.NotEqual(t, http.StatusOK, recorder.Code)
	})

	t.Run("bind json error", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockObsSvc := iportdrivermock.NewMockIObservability(ctrl)
		h := &observabilityHTTPHandler{observabilitySvc: mockObsSvc, logger: obsTestLogger{}}

		c, recorder := newObsCtx(http.MethodPost, "/", "{")
		c.Params = gin.Params{{Key: "agent_id", Value: "a1"}}
		h.AgentDetail(c)
		assert.NotEqual(t, http.StatusOK, recorder.Code)
	})

	t.Run("time range missing", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockObsSvc := iportdrivermock.NewMockIObservability(ctrl)
		h := &observabilityHTTPHandler{observabilitySvc: mockObsSvc, logger: obsTestLogger{}}

		c, recorder := newObsCtx(http.MethodPost, "/", `{}`)
		c.Params = gin.Params{{Key: "agent_id", Value: "a1"}}
		h.AgentDetail(c)
		assert.NotEqual(t, http.StatusOK, recorder.Code)
	})

	t.Run("start_time > end_time", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockObsSvc := iportdrivermock.NewMockIObservability(ctrl)
		h := &observabilityHTTPHandler{observabilitySvc: mockObsSvc, logger: obsTestLogger{}}

		c, recorder := newObsCtx(http.MethodPost, "/", `{"start_time":100,"end_time":50}`)
		c.Params = gin.Params{{Key: "agent_id", Value: "a1"}}
		h.AgentDetail(c)
		assert.NotEqual(t, http.StatusOK, recorder.Code)
	})

	t.Run("user not found", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockObsSvc := iportdrivermock.NewMockIObservability(ctrl)
		h := &observabilityHTTPHandler{observabilitySvc: mockObsSvc, logger: obsTestLogger{}}

		c, recorder := newObsCtx(http.MethodPost, "/", `{"start_time":1,"end_time":100}`)
		c.Params = gin.Params{{Key: "agent_id", Value: "a1"}}
		// 不设置 visitor
		h.AgentDetail(c)
		assert.NotEqual(t, http.StatusOK, recorder.Code)
	})

	t.Run("service error", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockObsSvc := iportdrivermock.NewMockIObservability(ctrl)
		h := &observabilityHTTPHandler{observabilitySvc: mockObsSvc, logger: obsTestLogger{}}

		mockObsSvc.EXPECT().AgentDetail(gomock.Any(), gomock.Any()).Return(nil, errors.New("svc failed"))

		c, recorder := newObsCtx(http.MethodPost, "/", `{"start_time":1,"end_time":100}`)
		c.Params = gin.Params{{Key: "agent_id", Value: "a1"}}
		setObsVisitor(c, "u1")
		h.AgentDetail(c)
		assert.NotEqual(t, http.StatusOK, recorder.Code)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockObsSvc := iportdrivermock.NewMockIObservability(ctrl)
		h := &observabilityHTTPHandler{observabilitySvc: mockObsSvc, logger: obsTestLogger{}}

		mockObsSvc.EXPECT().AgentDetail(gomock.Any(), gomock.Any()).Return(&observabilityresp.AgentResp{}, nil)

		c, recorder := newObsCtx(http.MethodPost, "/", `{"start_time":1,"end_time":100}`)
		c.Params = gin.Params{{Key: "agent_id", Value: "a1"}}
		setObsVisitor(c, "u1")
		h.AgentDetail(c)
		assert.Equal(t, http.StatusOK, recorder.Code)
	})
}

// ==================== ConversationList ====================

func TestObservabilityHandler_ConversationList(t *testing.T) {
	t.Parallel()

	t.Run("agent_id empty", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockObsSvc := iportdrivermock.NewMockIObservability(ctrl)
		mockConvSvc := iportdrivermock.NewMockIConversationSvc(ctrl)
		h := &observabilityHTTPHandler{observabilitySvc: mockObsSvc, conversationSvc: mockConvSvc, logger: obsTestLogger{}}

		c, recorder := newObsCtx(http.MethodPost, "/", `{}`)
		h.ConversationList(c)
		assert.NotEqual(t, http.StatusOK, recorder.Code)
	})

	t.Run("bind json error", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockObsSvc := iportdrivermock.NewMockIObservability(ctrl)
		mockConvSvc := iportdrivermock.NewMockIConversationSvc(ctrl)
		h := &observabilityHTTPHandler{observabilitySvc: mockObsSvc, conversationSvc: mockConvSvc, logger: obsTestLogger{}}

		c, recorder := newObsCtx(http.MethodPost, "/", "{")
		c.Params = gin.Params{{Key: "agent_id", Value: "a1"}}
		h.ConversationList(c)
		assert.NotEqual(t, http.StatusOK, recorder.Code)
	})

	t.Run("start_time > end_time", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockObsSvc := iportdrivermock.NewMockIObservability(ctrl)
		mockConvSvc := iportdrivermock.NewMockIConversationSvc(ctrl)
		h := &observabilityHTTPHandler{observabilitySvc: mockObsSvc, conversationSvc: mockConvSvc, logger: obsTestLogger{}}

		c, recorder := newObsCtx(http.MethodPost, "/", `{"start_time":100,"end_time":50}`)
		c.Params = gin.Params{{Key: "agent_id", Value: "a1"}}
		h.ConversationList(c)
		assert.NotEqual(t, http.StatusOK, recorder.Code)
	})

	t.Run("user not found", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockObsSvc := iportdrivermock.NewMockIObservability(ctrl)
		mockConvSvc := iportdrivermock.NewMockIConversationSvc(ctrl)
		h := &observabilityHTTPHandler{observabilitySvc: mockObsSvc, conversationSvc: mockConvSvc, logger: obsTestLogger{}}

		c, recorder := newObsCtx(http.MethodPost, "/", `{"start_time":1,"end_time":100}`)
		c.Params = gin.Params{{Key: "agent_id", Value: "a1"}}
		h.ConversationList(c)
		assert.NotEqual(t, http.StatusOK, recorder.Code)
	})

	t.Run("conversation svc error", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockObsSvc := iportdrivermock.NewMockIObservability(ctrl)
		mockConvSvc := iportdrivermock.NewMockIConversationSvc(ctrl)
		h := &observabilityHTTPHandler{observabilitySvc: mockObsSvc, conversationSvc: mockConvSvc, logger: obsTestLogger{}}

		mockConvSvc.EXPECT().ListByAgentID(gomock.Any(), "a1", gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil, int64(0), errors.New("list failed"))

		c, recorder := newObsCtx(http.MethodPost, "/", `{"start_time":1,"end_time":100}`)
		c.Params = gin.Params{{Key: "agent_id", Value: "a1"}}
		setObsVisitor(c, "u1")
		h.ConversationList(c)
		assert.NotEqual(t, http.StatusOK, recorder.Code)
	})

	t.Run("success empty list", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockObsSvc := iportdrivermock.NewMockIObservability(ctrl)
		mockConvSvc := iportdrivermock.NewMockIConversationSvc(ctrl)
		h := &observabilityHTTPHandler{observabilitySvc: mockObsSvc, conversationSvc: mockConvSvc, logger: obsTestLogger{}}

		mockConvSvc.EXPECT().ListByAgentID(gomock.Any(), "a1", gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(conversationresp.ListConversationResp{}, int64(0), nil)

		c, recorder := newObsCtx(http.MethodPost, "/", `{"start_time":1,"end_time":100}`)
		c.Params = gin.Params{{Key: "agent_id", Value: "a1"}}
		setObsVisitor(c, "u1")
		h.ConversationList(c)
		assert.Equal(t, http.StatusOK, recorder.Code)
	})

	t.Run("success default params", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockObsSvc := iportdrivermock.NewMockIObservability(ctrl)
		mockConvSvc := iportdrivermock.NewMockIConversationSvc(ctrl)
		h := &observabilityHTTPHandler{observabilitySvc: mockObsSvc, conversationSvc: mockConvSvc, logger: obsTestLogger{}}

		// 返回空列表，这样不会调用 GetSessionCountsByConversationIDs
		mockConvSvc.EXPECT().ListByAgentID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return([]conversationresp.ConversationDetail{}, int64(0), nil)

		c, recorder := newObsCtx(http.MethodPost, "/", `{"start_time":1,"end_time":100}`)
		c.Params = gin.Params{{Key: "agent_id", Value: "a1"}}
		setObsVisitor(c, "u1")
		h.ConversationList(c)
		assert.Equal(t, http.StatusOK, recorder.Code)
	})
}

// ==================== SessionList ====================

func TestObservabilityHandler_SessionList(t *testing.T) {
	t.Parallel()

	t.Run("agent_id empty", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockObsSvc := iportdrivermock.NewMockIObservability(ctrl)
		h := &observabilityHTTPHandler{observabilitySvc: mockObsSvc, logger: obsTestLogger{}}

		c, recorder := newObsCtx(http.MethodPost, "/", `{}`)
		h.SessionList(c)
		assert.NotEqual(t, http.StatusOK, recorder.Code)
	})

	t.Run("conversation_id empty", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockObsSvc := iportdrivermock.NewMockIObservability(ctrl)
		h := &observabilityHTTPHandler{observabilitySvc: mockObsSvc, logger: obsTestLogger{}}

		c, recorder := newObsCtx(http.MethodPost, "/", `{}`)
		c.Params = gin.Params{{Key: "agent_id", Value: "a1"}}
		h.SessionList(c)
		assert.NotEqual(t, http.StatusOK, recorder.Code)
	})

	t.Run("time range missing", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockObsSvc := iportdrivermock.NewMockIObservability(ctrl)
		h := &observabilityHTTPHandler{observabilitySvc: mockObsSvc, logger: obsTestLogger{}}

		c, recorder := newObsCtx(http.MethodPost, "/", `{}`)
		c.Params = gin.Params{{Key: "agent_id", Value: "a1"}, {Key: "conversation_id", Value: "c1"}}
		h.SessionList(c)
		assert.NotEqual(t, http.StatusOK, recorder.Code)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockObsSvc := iportdrivermock.NewMockIObservability(ctrl)
		h := &observabilityHTTPHandler{observabilitySvc: mockObsSvc, logger: obsTestLogger{}}

		mockObsSvc.EXPECT().SessionList(gomock.Any(), gomock.Any()).Return(&observabilityresp.SessionListResp{}, nil)

		c, recorder := newObsCtx(http.MethodPost, "/", `{"start_time":1,"end_time":100}`)
		c.Params = gin.Params{{Key: "agent_id", Value: "a1"}, {Key: "conversation_id", Value: "c1"}}
		setObsVisitor(c, "u1")
		h.SessionList(c)
		assert.Equal(t, http.StatusOK, recorder.Code)
	})
}

// ==================== SessionDetail ====================

func TestObservabilityHandler_SessionDetail(t *testing.T) {
	t.Parallel()

	t.Run("agent_id empty", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockObsSvc := iportdrivermock.NewMockIObservability(ctrl)
		h := &observabilityHTTPHandler{observabilitySvc: mockObsSvc, logger: obsTestLogger{}}

		c, recorder := newObsCtx(http.MethodPost, "/", `{}`)
		h.SessionDetail(c)
		assert.NotEqual(t, http.StatusOK, recorder.Code)
	})

	t.Run("session_id empty", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockObsSvc := iportdrivermock.NewMockIObservability(ctrl)
		h := &observabilityHTTPHandler{observabilitySvc: mockObsSvc, logger: obsTestLogger{}}

		c, recorder := newObsCtx(http.MethodPost, "/", `{}`)
		c.Params = gin.Params{{Key: "agent_id", Value: "a1"}, {Key: "conversation_id", Value: "c1"}}
		h.SessionDetail(c)
		assert.NotEqual(t, http.StatusOK, recorder.Code)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockObsSvc := iportdrivermock.NewMockIObservability(ctrl)
		h := &observabilityHTTPHandler{observabilitySvc: mockObsSvc, logger: obsTestLogger{}}

		mockObsSvc.EXPECT().SessionDetail(gomock.Any(), gomock.Any()).Return(&observabilityresp.SessionDetailResp{}, nil)

		c, recorder := newObsCtx(http.MethodPost, "/", `{"start_time":1,"end_time":100}`)
		c.Params = gin.Params{{Key: "agent_id", Value: "a1"}, {Key: "conversation_id", Value: "c1"}, {Key: "session_id", Value: "s1"}}
		setObsVisitor(c, "u1")
		h.SessionDetail(c)
		assert.Equal(t, http.StatusOK, recorder.Code)
	})
}

// ==================== RunList ====================

func TestObservabilityHandler_RunList(t *testing.T) {
	t.Parallel()

	t.Run("agent_id empty", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockObsSvc := iportdrivermock.NewMockIObservability(ctrl)
		h := &observabilityHTTPHandler{observabilitySvc: mockObsSvc, logger: obsTestLogger{}}

		c, recorder := newObsCtx(http.MethodPost, "/", `{}`)
		h.RunList(c)
		assert.NotEqual(t, http.StatusOK, recorder.Code)
	})

	t.Run("session_id empty", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockObsSvc := iportdrivermock.NewMockIObservability(ctrl)
		h := &observabilityHTTPHandler{observabilitySvc: mockObsSvc, logger: obsTestLogger{}}

		c, recorder := newObsCtx(http.MethodPost, "/", `{}`)
		c.Params = gin.Params{{Key: "agent_id", Value: "a1"}, {Key: "conversation_id", Value: "c1"}}
		h.RunList(c)
		assert.NotEqual(t, http.StatusOK, recorder.Code)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockObsSvc := iportdrivermock.NewMockIObservability(ctrl)
		h := &observabilityHTTPHandler{observabilitySvc: mockObsSvc, logger: obsTestLogger{}}

		mockObsSvc.EXPECT().RunList(gomock.Any(), gomock.Any()).Return(&observabilityresp.RunListResp{}, nil)

		c, recorder := newObsCtx(http.MethodPost, "/", `{"start_time":1,"end_time":100}`)
		c.Params = gin.Params{{Key: "agent_id", Value: "a1"}, {Key: "conversation_id", Value: "c1"}, {Key: "session_id", Value: "s1"}}
		setObsVisitor(c, "u1")
		h.RunList(c)
		assert.Equal(t, http.StatusOK, recorder.Code)
	})
}

// ==================== RunDetail ====================

func TestObservabilityHandler_RunDetail(t *testing.T) {
	t.Parallel()

	t.Run("agent_id empty", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockObsSvc := iportdrivermock.NewMockIObservability(ctrl)
		h := &observabilityHTTPHandler{observabilitySvc: mockObsSvc, logger: obsTestLogger{}}

		c, recorder := newObsCtx(http.MethodPost, "/", `{}`)
		h.RunDetail(c)
		assert.NotEqual(t, http.StatusOK, recorder.Code)
	})

	t.Run("run_id empty", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockObsSvc := iportdrivermock.NewMockIObservability(ctrl)
		h := &observabilityHTTPHandler{observabilitySvc: mockObsSvc, logger: obsTestLogger{}}

		c, recorder := newObsCtx(http.MethodPost, "/", `{}`)
		c.Params = gin.Params{{Key: "agent_id", Value: "a1"}, {Key: "conversation_id", Value: "c1"}, {Key: "session_id", Value: "s1"}}
		h.RunDetail(c)
		assert.NotEqual(t, http.StatusOK, recorder.Code)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockObsSvc := iportdrivermock.NewMockIObservability(ctrl)
		h := &observabilityHTTPHandler{observabilitySvc: mockObsSvc, logger: obsTestLogger{}}

		mockObsSvc.EXPECT().RunDetail(gomock.Any(), gomock.Any()).Return(&observabilityresp.RunDetailResp{}, nil)

		c, recorder := newObsCtx(http.MethodPost, "/", `{"start_time":1,"end_time":100}`)
		c.Params = gin.Params{{Key: "agent_id", Value: "a1"}, {Key: "conversation_id", Value: "c1"}, {Key: "session_id", Value: "s1"}, {Key: "run_id", Value: "r1"}}
		setObsVisitor(c, "u1")
		h.RunDetail(c)
		assert.Equal(t, http.StatusOK, recorder.Code)
	})
}

// ==================== AnalyticsQuery ====================

func TestObservabilityHandler_AnalyticsQuery(t *testing.T) {
	t.Parallel()

	t.Run("bind json error", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockObsSvc := iportdrivermock.NewMockIObservability(ctrl)
		h := &observabilityHTTPHandler{observabilitySvc: mockObsSvc, logger: obsTestLogger{}}

		c, recorder := newObsCtx(http.MethodPost, "/", "{")
		h.AnalyticsQuery(c)
		assert.NotEqual(t, http.StatusOK, recorder.Code)
	})

	t.Run("analysis_level empty", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockObsSvc := iportdrivermock.NewMockIObservability(ctrl)
		h := &observabilityHTTPHandler{observabilitySvc: mockObsSvc, logger: obsTestLogger{}}

		c, recorder := newObsCtx(http.MethodPost, "/", `{"id":"x","start_time":1,"end_time":100}`)
		h.AnalyticsQuery(c)
		assert.NotEqual(t, http.StatusOK, recorder.Code)
	})

	t.Run("id empty", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockObsSvc := iportdrivermock.NewMockIObservability(ctrl)
		h := &observabilityHTTPHandler{observabilitySvc: mockObsSvc, logger: obsTestLogger{}}

		c, recorder := newObsCtx(http.MethodPost, "/", `{"analysis_level":"agent","start_time":1,"end_time":100}`)
		h.AnalyticsQuery(c)
		assert.NotEqual(t, http.StatusOK, recorder.Code)
	})

	t.Run("time range missing", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockObsSvc := iportdrivermock.NewMockIObservability(ctrl)
		h := &observabilityHTTPHandler{observabilitySvc: mockObsSvc, logger: obsTestLogger{}}

		c, recorder := newObsCtx(http.MethodPost, "/", `{"analysis_level":"agent","id":"x"}`)
		h.AnalyticsQuery(c)
		assert.NotEqual(t, http.StatusOK, recorder.Code)
	})

	t.Run("start_time > end_time", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockObsSvc := iportdrivermock.NewMockIObservability(ctrl)
		h := &observabilityHTTPHandler{observabilitySvc: mockObsSvc, logger: obsTestLogger{}}

		c, recorder := newObsCtx(http.MethodPost, "/", `{"analysis_level":"agent","id":"x","start_time":100,"end_time":50}`)
		h.AnalyticsQuery(c)
		assert.NotEqual(t, http.StatusOK, recorder.Code)
	})

	t.Run("service error", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockObsSvc := iportdrivermock.NewMockIObservability(ctrl)
		h := &observabilityHTTPHandler{observabilitySvc: mockObsSvc, logger: obsTestLogger{}}

		mockObsSvc.EXPECT().AnalyticsQuery(gomock.Any(), gomock.Any()).Return(nil, errors.New("svc failed"))

		c, recorder := newObsCtx(http.MethodPost, "/", `{"analysis_level":"agent","id":"x","start_time":1,"end_time":100}`)
		h.AnalyticsQuery(c)
		assert.NotEqual(t, http.StatusOK, recorder.Code)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockObsSvc := iportdrivermock.NewMockIObservability(ctrl)
		h := &observabilityHTTPHandler{observabilitySvc: mockObsSvc, logger: obsTestLogger{}}

		mockObsSvc.EXPECT().AnalyticsQuery(gomock.Any(), gomock.Any()).Return(&observabilityresp.AnalyticsQueryResp{}, nil)

		c, recorder := newObsCtx(http.MethodPost, "/", `{"analysis_level":"agent","id":"x","start_time":1,"end_time":100}`)
		h.AnalyticsQuery(c)
		assert.Equal(t, http.StatusOK, recorder.Code)
	})
}

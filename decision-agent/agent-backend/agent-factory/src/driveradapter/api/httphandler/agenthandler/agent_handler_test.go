package agenthandler

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

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/constant"
	agentrunsvc "github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service/agentrunsvc"
	agentreq "github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/agent/req"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driver/iportdriver/iportdrivermock"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

type agentTestLogger struct{}

func (agentTestLogger) Infof(string, ...interface{})  {}
func (agentTestLogger) Infoln(...interface{})         {}
func (agentTestLogger) Debugf(string, ...interface{}) {}
func (agentTestLogger) Debugln(...interface{})        {}
func (agentTestLogger) Errorf(string, ...interface{}) {}
func (agentTestLogger) Errorln(...interface{})        {}
func (agentTestLogger) Warnf(string, ...interface{})  {}
func (agentTestLogger) Warnln(...interface{})         {}
func (agentTestLogger) Panicf(string, ...interface{}) {}
func (agentTestLogger) Panicln(...interface{})        {}
func (agentTestLogger) Fatalf(string, ...interface{}) {}
func (agentTestLogger) Fatalln(...interface{})        {}

func newAgentCtx(method, target, body string) (*gin.Context, *httptest.ResponseRecorder) {
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

func setAgentVisitor(c *gin.Context, id, token string, visitorType rest.VisitorType) {
	visitor := &rest.Visitor{ID: id, TokenID: token, Type: visitorType}
	c.Set(cenum.VisitUserInfoCtxKey.String(), visitor)
	ctx := context.WithValue(c.Request.Context(), cenum.VisitUserInfoCtxKey.String(), visitor) //nolint:staticcheck // SA1029
	c.Request = c.Request.WithContext(ctx)
}

func jsonChannel(payload string) chan []byte {
	ch := make(chan []byte, 1)
	ch <- []byte(payload)
	close(ch)

	return ch
}

func hasAgentRoute(routes []gin.RouteInfo, method, path string) bool {
	for _, r := range routes {
		if r.Method == method && r.Path == path {
			return true
		}
	}

	return false
}

func TestAgentHandler_RegRouters(t *testing.T) {
	t.Parallel()

	h := &agentHTTPHandler{}

	pubRouter := gin.New()
	h.RegPubRouter(pubRouter.Group("/v1"))
	pubRoutes := pubRouter.Routes()
	assert.True(t, hasAgentRoute(pubRoutes, http.MethodPost, "/v1/app/:app_key/chat/resume"))
	assert.True(t, hasAgentRoute(pubRoutes, http.MethodPost, "/v1/app/:app_key/chat/termination"))
	assert.True(t, hasAgentRoute(pubRoutes, http.MethodPost, "/v1/app/:app_key/chat/completion"))
	assert.True(t, hasAgentRoute(pubRoutes, http.MethodPost, "/v1/app/:app_key/debug/completion"))
	assert.True(t, hasAgentRoute(pubRoutes, http.MethodPost, "/v1/app/:app_key/api/chat/completion"))
	assert.True(t, hasAgentRoute(pubRoutes, http.MethodPost, "/v1/api/chat/completion"))
	assert.True(t, hasAgentRoute(pubRoutes, http.MethodPost, "/v1/app/:app_key/api/doc"))

	priRouter := gin.New()
	h.RegPriRouter(priRouter.Group("/v1"))
	priRoutes := priRouter.Routes()
	assert.True(t, hasAgentRoute(priRoutes, http.MethodPost, "/v1/app/:app_key/chat/completion"))
	assert.True(t, hasAgentRoute(priRoutes, http.MethodPost, "/v1/app/:app_key/api/chat/completion"))
}

func TestAgentHandler_GetAPIDoc(t *testing.T) {
	t.Parallel()

	t.Run("bind json error", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockAgent := iportdrivermock.NewMockIAgent(ctrl)
		h := &agentHTTPHandler{agentSvc: mockAgent, logger: agentTestLogger{}}

		c, recorder := newAgentCtx(http.MethodPost, "/", "{")
		h.GetAPIDoc(c)
		assert.NotEqual(t, http.StatusOK, recorder.Code)
	})

	t.Run("app key empty", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockAgent := iportdrivermock.NewMockIAgent(ctrl)
		h := &agentHTTPHandler{agentSvc: mockAgent, logger: agentTestLogger{}}

		c, recorder := newAgentCtx(http.MethodPost, "/", "{}")
		h.GetAPIDoc(c)
		assert.NotEqual(t, http.StatusOK, recorder.Code)
	})

	t.Run("service error", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockAgent := iportdrivermock.NewMockIAgent(ctrl)
		h := &agentHTTPHandler{agentSvc: mockAgent, logger: agentTestLogger{}}

		mockAgent.EXPECT().GetAPIDoc(gomock.Any(), gomock.Any()).Return(nil, errors.New("get api doc failed"))

		c, recorder := newAgentCtx(http.MethodPost, "/", "{}")
		c.Params = gin.Params{{Key: "app_key", Value: "app-1"}}
		h.GetAPIDoc(c)
		assert.NotEqual(t, http.StatusOK, recorder.Code)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockAgent := iportdrivermock.NewMockIAgent(ctrl)
		h := &agentHTTPHandler{agentSvc: mockAgent, logger: agentTestLogger{}}

		mockAgent.EXPECT().GetAPIDoc(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, req *agentreq.GetAPIDocReq) (interface{}, error) {
			assert.Equal(t, "app-1", ctx.Value(constant.AppKey))
			return map[string]string{"openapi": "3.0"}, nil
		})

		c, recorder := newAgentCtx(http.MethodPost, "/", "{}")
		c.Params = gin.Params{{Key: "app_key", Value: "app-1"}}
		h.GetAPIDoc(c)
		assert.Equal(t, http.StatusOK, recorder.Code)
	})
}

func TestAgentHandler_Chat_EarlyAndNonStreamBranches(t *testing.T) {
	t.Parallel()

	t.Run("user not found", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockAgent := iportdrivermock.NewMockIAgent(ctrl)
		h := &agentHTTPHandler{agentSvc: mockAgent, logger: agentTestLogger{}}

		c, recorder := newAgentCtx(http.MethodPost, "/", `{"agent_id":"a1"}`)
		c.Params = gin.Params{{Key: "app_key", Value: "app-1"}}
		h.Chat(c)
		assert.NotEqual(t, http.StatusOK, recorder.Code)
	})

	t.Run("service error", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockAgent := iportdrivermock.NewMockIAgent(ctrl)
		h := &agentHTTPHandler{agentSvc: mockAgent, logger: agentTestLogger{}}

		mockAgent.EXPECT().Chat(gomock.Any(), gomock.Any()).Return(nil, errors.New("chat failed"))

		c, recorder := newAgentCtx(http.MethodPost, "/", `{"agent_id":"a1"}`)
		c.Params = gin.Params{{Key: "app_key", Value: "app-1"}}
		setAgentVisitor(c, "u1", "Bearer tk", rest.VisitorType_User)
		h.Chat(c)
		assert.NotEqual(t, http.StatusOK, recorder.Code)
	})

	t.Run("non stream success", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockAgent := iportdrivermock.NewMockIAgent(ctrl)
		h := &agentHTTPHandler{agentSvc: mockAgent, logger: agentTestLogger{}}

		mockAgent.EXPECT().Chat(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, req *agentreq.ChatReq) (chan []byte, error) {
			assert.Equal(t, constant.Chat, req.CallType)
			assert.Equal(t, "u1", req.UserID)
			assert.False(t, req.Stream)

			return jsonChannel(`{"content":"ok"}`), nil
		})

		c, recorder := newAgentCtx(http.MethodPost, "/", `{"agent_id":"a1","stream":false}`)
		c.Params = gin.Params{{Key: "app_key", Value: "app-1"}}
		setAgentVisitor(c, "u1", "Bearer tk", rest.VisitorType_User)
		h.Chat(c)
		assert.Equal(t, http.StatusOK, recorder.Code)
	})
}

func TestAgentHandler_APIChat_EarlyAndNonStreamBranches(t *testing.T) {
	t.Parallel()

	t.Run("agent_key required", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockAgent := iportdrivermock.NewMockIAgent(ctrl)
		h := &agentHTTPHandler{agentSvc: mockAgent, logger: agentTestLogger{}}

		c, recorder := newAgentCtx(http.MethodPost, "/", `{"stream":false}`)
		c.Params = gin.Params{{Key: "app_key", Value: "app-1"}}
		setAgentVisitor(c, "u1", "Bearer tk", rest.VisitorType_App)
		h.APIChat(c)
		assert.NotEqual(t, http.StatusOK, recorder.Code)
	})

	t.Run("service error", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockAgent := iportdrivermock.NewMockIAgent(ctrl)
		h := &agentHTTPHandler{agentSvc: mockAgent, logger: agentTestLogger{}}

		mockAgent.EXPECT().Chat(gomock.Any(), gomock.Any()).Return(nil, errors.New("api chat failed"))

		c, recorder := newAgentCtx(http.MethodPost, "/", `{"agent_key":"k1","stream":false}`)
		c.Params = gin.Params{{Key: "app_key", Value: "app-1"}}
		setAgentVisitor(c, "u1", "Bearer tk", rest.VisitorType_App)
		h.APIChat(c)
		assert.NotEqual(t, http.StatusOK, recorder.Code)
	})

	t.Run("non stream success", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockAgent := iportdrivermock.NewMockIAgent(ctrl)
		h := &agentHTTPHandler{agentSvc: mockAgent, logger: agentTestLogger{}}

		mockAgent.EXPECT().Chat(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, req *agentreq.ChatReq) (chan []byte, error) {
			assert.Equal(t, constant.APIChat, req.CallType)
			assert.Equal(t, "k1", req.AgentID)
			assert.Equal(t, "app-1", req.AgentAPPKey)

			return jsonChannel(`{"message":"ok"}`), nil
		})

		c, recorder := newAgentCtx(http.MethodPost, "/", `{"agent_key":"k1","stream":false}`)
		c.Params = gin.Params{{Key: "app_key", Value: "app-1"}}
		setAgentVisitor(c, "u1", "Bearer tk", rest.VisitorType_App)
		h.APIChat(c)
		assert.Equal(t, http.StatusOK, recorder.Code)
	})

	t.Run("non stream success without path app_key uses agent_key as app key", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockAgent := iportdrivermock.NewMockIAgent(ctrl)
		h := &agentHTTPHandler{agentSvc: mockAgent, logger: agentTestLogger{}}

		mockAgent.EXPECT().Chat(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, req *agentreq.ChatReq) (chan []byte, error) {
			assert.Equal(t, constant.APIChat, req.CallType)
			assert.Equal(t, "k1", req.AgentID)
			assert.Equal(t, "k1", req.AgentAPPKey)

			return jsonChannel(`{"message":"ok"}`), nil
		})

		c, recorder := newAgentCtx(http.MethodPost, "/", `{"agent_key":"k1","stream":false}`)
		setAgentVisitor(c, "u1", "Bearer tk", rest.VisitorType_App)
		h.APIChat(c)
		assert.Equal(t, http.StatusOK, recorder.Code)
	})
}

func TestAgentHandler_InternalChatAndInternalAPIChat_NonStream(t *testing.T) {
	t.Parallel()

	t.Run("internal chat success", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockAgent := iportdrivermock.NewMockIAgent(ctrl)
		h := &agentHTTPHandler{agentSvc: mockAgent, logger: agentTestLogger{}}

		mockAgent.EXPECT().Chat(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, req *agentreq.ChatReq) (chan []byte, error) {
			assert.Equal(t, constant.InternalChat, req.CallType)
			assert.Equal(t, "inner-user", req.UserID)

			return jsonChannel(`{"ok":true}`), nil
		})

		c, recorder := newAgentCtx(http.MethodPost, "/", `{"agent_id":"a1","stream":false}`)
		c.Params = gin.Params{{Key: "app_key", Value: "app-1"}}
		c.Request.Header.Set("x-user", "inner-user")
		c.Request.Header.Set("x-account-id", "acc-1")
		c.Request.Header.Set("x-account-type", "user")
		h.InternalChat(c)
		assert.Equal(t, http.StatusOK, recorder.Code)
	})

	t.Run("internal api chat agent_key required", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockAgent := iportdrivermock.NewMockIAgent(ctrl)
		h := &agentHTTPHandler{agentSvc: mockAgent, logger: agentTestLogger{}}

		c, recorder := newAgentCtx(http.MethodPost, "/", `{"stream":false}`)
		c.Params = gin.Params{{Key: "app_key", Value: "app-1"}}
		h.InternalAPIChat(c)
		assert.NotEqual(t, http.StatusOK, recorder.Code)
	})

	t.Run("internal api chat success", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockAgent := iportdrivermock.NewMockIAgent(ctrl)
		h := &agentHTTPHandler{agentSvc: mockAgent, logger: agentTestLogger{}}

		mockAgent.EXPECT().Chat(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, req *agentreq.ChatReq) (chan []byte, error) {
			assert.Equal(t, constant.APIChat, req.CallType)
			assert.Equal(t, "k1", req.AgentID)

			return jsonChannel(`{"ok":true}`), nil
		})

		c, recorder := newAgentCtx(http.MethodPost, "/", `{"agent_key":"k1","stream":false}`)
		c.Params = gin.Params{{Key: "app_key", Value: "app-1"}}
		c.Request.Header.Set("x-user", "inner-user")
		c.Request.Header.Set("x-account-id", "acc-1")
		c.Request.Header.Set("x-account-type", "user")
		h.InternalAPIChat(c)
		assert.Equal(t, http.StatusOK, recorder.Code)
	})
}

func TestAgentHandler_Debug_NonStream(t *testing.T) {
	t.Parallel()

	t.Run("user not found", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockAgent := iportdrivermock.NewMockIAgent(ctrl)
		h := &agentHTTPHandler{agentSvc: mockAgent, logger: agentTestLogger{}}

		c, recorder := newAgentCtx(http.MethodPost, "/", `{"stream":false,"input":{"query":"q"}}`)
		c.Params = gin.Params{{Key: "app_key", Value: "app-1"}}
		h.Debug(c)
		assert.NotEqual(t, http.StatusOK, recorder.Code)
	})

	t.Run("service error", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockAgent := iportdrivermock.NewMockIAgent(ctrl)
		h := &agentHTTPHandler{agentSvc: mockAgent, logger: agentTestLogger{}}

		mockAgent.EXPECT().Chat(gomock.Any(), gomock.Any()).Return(nil, errors.New("debug chat failed"))

		c, recorder := newAgentCtx(http.MethodPost, "/", `{"stream":false,"input":{"query":"q"}}`)
		c.Params = gin.Params{{Key: "app_key", Value: "app-1"}}
		setAgentVisitor(c, "u1", "Bearer tk", rest.VisitorType_User)
		h.Debug(c)
		assert.NotEqual(t, http.StatusOK, recorder.Code)
	})

	t.Run("non stream success", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockAgent := iportdrivermock.NewMockIAgent(ctrl)
		h := &agentHTTPHandler{agentSvc: mockAgent, logger: agentTestLogger{}}

		mockAgent.EXPECT().Chat(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, req *agentreq.ChatReq) (chan []byte, error) {
			assert.Equal(t, constant.DebugChat, req.CallType)
			assert.Equal(t, "q", req.Query)

			return jsonChannel(`{"ok":true}`), nil
		})

		c, recorder := newAgentCtx(http.MethodPost, "/", `{"stream":false,"input":{"query":"q"}}`)
		c.Params = gin.Params{{Key: "app_key", Value: "app-1"}}
		setAgentVisitor(c, "u1", "Bearer tk", rest.VisitorType_User)
		h.Debug(c)
		assert.Equal(t, http.StatusOK, recorder.Code)
	})
}

func TestAgentHandler_TerminateChat(t *testing.T) {
	t.Parallel()

	t.Run("bind json error", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockAgent := iportdrivermock.NewMockIAgent(ctrl)
		h := &agentHTTPHandler{agentSvc: mockAgent, logger: agentTestLogger{}}

		c, recorder := newAgentCtx(http.MethodPost, "/", "{")
		h.TerminateChat(c)
		assert.NotEqual(t, http.StatusNoContent, recorder.Code)
	})

	t.Run("conversation id empty", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockAgent := iportdrivermock.NewMockIAgent(ctrl)
		h := &agentHTTPHandler{agentSvc: mockAgent, logger: agentTestLogger{}}

		c, recorder := newAgentCtx(http.MethodPost, "/", `{"conversation_id":""}`)
		h.TerminateChat(c)
		assert.NotEqual(t, http.StatusNoContent, recorder.Code)
	})

	t.Run("service error", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockAgent := iportdrivermock.NewMockIAgent(ctrl)
		h := &agentHTTPHandler{agentSvc: mockAgent, logger: agentTestLogger{}}

		mockAgent.EXPECT().TerminateChat(gomock.Any(), "conv-1", "run-1", "asst-1").Return(errors.New("terminate failed"))

		c, recorder := newAgentCtx(http.MethodPost, "/", `{"conversation_id":"conv-1","agent_run_id":"run-1","interrupted_assistant_message_id":"asst-1"}`)
		h.TerminateChat(c)
		assert.NotEqual(t, http.StatusNoContent, recorder.Code)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockAgent := iportdrivermock.NewMockIAgent(ctrl)
		h := &agentHTTPHandler{agentSvc: mockAgent, logger: agentTestLogger{}}

		mockAgent.EXPECT().TerminateChat(gomock.Any(), "conv-1", "run-1", "asst-1").Return(nil)

		c, recorder := newAgentCtx(http.MethodPost, "/", `{"conversation_id":"conv-1","agent_run_id":"run-1","interrupted_assistant_message_id":"asst-1"}`)
		h.TerminateChat(c)
		assert.Equal(t, http.StatusNoContent, recorder.Code)
	})
}

func TestAgentHandler_ResumeChat(t *testing.T) {
	t.Parallel()

	t.Run("bind json error", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockAgent := iportdrivermock.NewMockIAgent(ctrl)
		h := &agentHTTPHandler{agentSvc: mockAgent, logger: agentTestLogger{}}

		c, recorder := newAgentCtx(http.MethodPost, "/", "{")
		h.ResumeChat(c)
		assert.NotEqual(t, http.StatusOK, recorder.Code)
	})

	t.Run("service error", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockAgent := iportdrivermock.NewMockIAgent(ctrl)
		h := &agentHTTPHandler{agentSvc: mockAgent, logger: agentTestLogger{}}

		mockAgent.EXPECT().ResumeChat(gomock.Any(), "conv-1").Return(nil, errors.New("resume failed"))

		c, recorder := newAgentCtx(http.MethodPost, "/", `{"conversation_id":"conv-1"}`)
		h.ResumeChat(c)
		assert.NotEqual(t, http.StatusOK, recorder.Code)
	})

	t.Run("success stream and reset session flag", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockAgent := iportdrivermock.NewMockIAgent(ctrl)
		h := &agentHTTPHandler{agentSvc: mockAgent, logger: agentTestLogger{}}

		ch := make(chan []byte, 2)
		ch <- []byte("data: partial")
		ch <- []byte(constant.DataEventEndStr)
		close(ch)
		mockAgent.EXPECT().ResumeChat(gomock.Any(), "conv-1").Return(ch, nil)

		session := &agentrunsvc.Session{ConversationID: "conv-1", IsResuming: true}
		agentrunsvc.SessionMap.Store("conv-1", session)

		defer agentrunsvc.SessionMap.Delete("conv-1")

		c, recorder := newAgentCtx(http.MethodPost, "/", `{"conversation_id":"conv-1"}`)
		h.ResumeChat(c)
		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.False(t, session.GetIsResuming())
	})
}

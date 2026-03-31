package agenthandler

import (
	"context"
	"net/http"
	"testing"

	agentreq "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent/req"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driver/iportdriver/iportdrivermock"
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// ==================== Chat — non-stream additional paths ====================

func TestChat_NonStream_NilResult(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockAgent := iportdrivermock.NewMockIAgent(ctrl)
	h := &agentHTTPHandler{agentSvc: mockAgent, logger: agentTestLogger{}}

	// Return empty closed channel → res will be nil
	ch := make(chan []byte)
	close(ch)
	mockAgent.EXPECT().Chat(gomock.Any(), gomock.Any()).Return(ch, nil)

	c, recorder := newAgentCtx(http.MethodPost, "/", `{"agent_id":"a1","stream":false}`)
	c.Params = append(c.Params, paramKV("app_key", "app-1"))
	setAgentVisitor(c, "u1", "Bearer tk", rest.VisitorType_User)
	h.Chat(c)
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestChat_NonStream_BaseErrorResult(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockAgent := iportdrivermock.NewMockIAgent(ctrl)
	h := &agentHTTPHandler{agentSvc: mockAgent, logger: agentTestLogger{}}

	// Return channel with BaseError JSON
	mockAgent.EXPECT().Chat(gomock.Any(), gomock.Any()).Return(
		jsonChannel(`{"BaseError":{"Code":500,"Message":"internal error"}}`), nil)

	c, recorder := newAgentCtx(http.MethodPost, "/", `{"agent_id":"a1","stream":false}`)
	c.Params = append(c.Params, paramKV("app_key", "app-1"))
	setAgentVisitor(c, "u1", "Bearer tk", rest.VisitorType_User)
	h.Chat(c)
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestChat_NonStream_UnmarshalError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockAgent := iportdrivermock.NewMockIAgent(ctrl)
	h := &agentHTTPHandler{agentSvc: mockAgent, logger: agentTestLogger{}}

	// Return channel with invalid JSON
	mockAgent.EXPECT().Chat(gomock.Any(), gomock.Any()).Return(
		jsonChannel(`not valid json`), nil)

	c, recorder := newAgentCtx(http.MethodPost, "/", `{"agent_id":"a1","stream":false}`)
	c.Params = append(c.Params, paramKV("app_key", "app-1"))
	setAgentVisitor(c, "u1", "Bearer tk", rest.VisitorType_User)
	h.Chat(c)
	assert.NotEqual(t, http.StatusOK, recorder.Code)
}

func TestChat_EmptyAppKey(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockAgent := iportdrivermock.NewMockIAgent(ctrl)
	h := &agentHTTPHandler{agentSvc: mockAgent, logger: agentTestLogger{}}

	c, recorder := newAgentCtx(http.MethodPost, "/", `{"agent_id":"a1"}`)
	// No app_key param
	h.Chat(c)
	assert.NotEqual(t, http.StatusOK, recorder.Code)
}

func TestChat_BindJsonError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockAgent := iportdrivermock.NewMockIAgent(ctrl)
	h := &agentHTTPHandler{agentSvc: mockAgent, logger: agentTestLogger{}}

	c, recorder := newAgentCtx(http.MethodPost, "/", `{invalid`)
	c.Params = append(c.Params, paramKV("app_key", "app-1"))
	h.Chat(c)
	assert.NotEqual(t, http.StatusOK, recorder.Code)
}

// ==================== Debug — non-stream nil result (previously panic, now fixed) ====================

func TestDebug_NonStream_NilResult(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockAgent := iportdrivermock.NewMockIAgent(ctrl)
	h := &agentHTTPHandler{agentSvc: mockAgent, logger: agentTestLogger{}}

	ch := make(chan []byte)
	close(ch)
	mockAgent.EXPECT().Chat(gomock.Any(), gomock.Any()).Return(ch, nil)

	c, recorder := newAgentCtx(http.MethodPost, "/", `{"stream":false,"input":{"query":"q"}}`)
	c.Params = append(c.Params, paramKV("app_key", "app-1"))
	setAgentVisitor(c, "u1", "Bearer tk", rest.VisitorType_User)
	h.Debug(c)
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
}

// ==================== APIChat — non-stream nil result (previously panic, now fixed) ====================

func TestAPIChat_NonStream_NilResult(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockAgent := iportdrivermock.NewMockIAgent(ctrl)
	h := &agentHTTPHandler{agentSvc: mockAgent, logger: agentTestLogger{}}

	ch := make(chan []byte)
	close(ch)
	mockAgent.EXPECT().Chat(gomock.Any(), gomock.Any()).Return(ch, nil)

	c, recorder := newAgentCtx(http.MethodPost, "/", `{"agent_key":"k1","stream":false}`)
	c.Params = append(c.Params, paramKV("app_key", "app-1"))
	setAgentVisitor(c, "u1", "Bearer tk", rest.VisitorType_App)
	h.APIChat(c)
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
}

// ==================== InternalChat — non-stream nil result ====================

func TestInternalChat_NonStream_NilResult(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockAgent := iportdrivermock.NewMockIAgent(ctrl)
	h := &agentHTTPHandler{agentSvc: mockAgent, logger: agentTestLogger{}}

	ch := make(chan []byte)
	close(ch)
	mockAgent.EXPECT().Chat(gomock.Any(), gomock.Any()).Return(ch, nil)

	c, recorder := newAgentCtx(http.MethodPost, "/", `{"agent_id":"a1","stream":false}`)
	c.Params = append(c.Params, paramKV("app_key", "app-1"))
	c.Request.Header.Set("x-user", "inner-user")
	c.Request.Header.Set("x-account-id", "acc-1")
	c.Request.Header.Set("x-account-type", "user")
	h.InternalChat(c)
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
}

// ==================== InternalAPIChat — non-stream nil result ====================

func TestInternalAPIChat_NonStream_NilResult(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockAgent := iportdrivermock.NewMockIAgent(ctrl)
	h := &agentHTTPHandler{agentSvc: mockAgent, logger: agentTestLogger{}}

	ch := make(chan []byte)
	close(ch)
	mockAgent.EXPECT().Chat(gomock.Any(), gomock.Any()).Return(ch, nil)

	c, recorder := newAgentCtx(http.MethodPost, "/", `{"agent_key":"k1","stream":false}`)
	c.Params = append(c.Params, paramKV("app_key", "app-1"))
	c.Request.Header.Set("x-user", "inner-user")
	c.Request.Header.Set("x-account-id", "acc-1")
	c.Request.Header.Set("x-account-type", "user")
	h.InternalAPIChat(c)
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
}

// ==================== InternalChat — additional paths ====================

func TestInternalChat_ServiceError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockAgent := iportdrivermock.NewMockIAgent(ctrl)
	h := &agentHTTPHandler{agentSvc: mockAgent, logger: agentTestLogger{}}

	mockAgent.EXPECT().Chat(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, req *agentreq.ChatReq) (chan []byte, error) {
		return nil, assert.AnError
	})

	c, recorder := newAgentCtx(http.MethodPost, "/", `{"agent_id":"a1","stream":false}`)
	c.Params = append(c.Params, paramKV("app_key", "app-1"))
	c.Request.Header.Set("x-user", "inner-user")
	c.Request.Header.Set("x-account-id", "acc-1")
	c.Request.Header.Set("x-account-type", "user")
	h.InternalChat(c)
	assert.NotEqual(t, http.StatusOK, recorder.Code)
}

// ==================== InternalAPIChat — additional paths ====================

func TestInternalAPIChat_ServiceError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockAgent := iportdrivermock.NewMockIAgent(ctrl)
	h := &agentHTTPHandler{agentSvc: mockAgent, logger: agentTestLogger{}}

	mockAgent.EXPECT().Chat(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, req *agentreq.ChatReq) (chan []byte, error) {
		return nil, assert.AnError
	})

	c, recorder := newAgentCtx(http.MethodPost, "/", `{"agent_key":"k1","stream":false}`)
	c.Params = append(c.Params, paramKV("app_key", "app-1"))
	c.Request.Header.Set("x-user", "inner-user")
	c.Request.Header.Set("x-account-id", "acc-1")
	c.Request.Header.Set("x-account-type", "user")
	h.InternalAPIChat(c)
	assert.NotEqual(t, http.StatusOK, recorder.Code)
}

// Helper
func paramKV(key, value string) struct {
	Key   string
	Value string
} {
	return struct {
		Key   string
		Value string
	}{Key: key, Value: value}
}

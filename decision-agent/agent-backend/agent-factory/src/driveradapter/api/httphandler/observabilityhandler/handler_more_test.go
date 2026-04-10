package observabilityhandler

import (
	"errors"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driver/iportdriver/iportdrivermock"
)

// ==================== RunDetail — deeper paths ====================

func TestRunDetail_ConversationIDEmpty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSvc := iportdrivermock.NewMockIObservability(ctrl)
	h := &observabilityHTTPHandler{observabilitySvc: mockSvc, logger: obsTestLogger{}}

	c, recorder := newObsCtx(http.MethodPost, "/", `{}`)
	c.Params = gin.Params{{Key: "agent_id", Value: "a1"}}
	h.RunDetail(c)
	assert.NotEqual(t, http.StatusOK, recorder.Code)
}

func TestRunDetail_SessionIDEmpty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSvc := iportdrivermock.NewMockIObservability(ctrl)
	h := &observabilityHTTPHandler{observabilitySvc: mockSvc, logger: obsTestLogger{}}

	c, recorder := newObsCtx(http.MethodPost, "/", `{}`)
	c.Params = gin.Params{{Key: "agent_id", Value: "a1"}, {Key: "conversation_id", Value: "c1"}}
	h.RunDetail(c)
	assert.NotEqual(t, http.StatusOK, recorder.Code)
}

func TestRunDetail_BindJsonError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSvc := iportdrivermock.NewMockIObservability(ctrl)
	h := &observabilityHTTPHandler{observabilitySvc: mockSvc, logger: obsTestLogger{}}

	c, recorder := newObsCtx(http.MethodPost, "/", "{")
	c.Params = gin.Params{
		{Key: "agent_id", Value: "a1"},
		{Key: "conversation_id", Value: "c1"},
		{Key: "session_id", Value: "s1"},
		{Key: "run_id", Value: "r1"},
	}
	h.RunDetail(c)
	assert.NotEqual(t, http.StatusOK, recorder.Code)
}

func TestRunDetail_TimeRangeMissing(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSvc := iportdrivermock.NewMockIObservability(ctrl)
	h := &observabilityHTTPHandler{observabilitySvc: mockSvc, logger: obsTestLogger{}}

	c, recorder := newObsCtx(http.MethodPost, "/", `{}`)
	c.Params = gin.Params{
		{Key: "agent_id", Value: "a1"},
		{Key: "conversation_id", Value: "c1"},
		{Key: "session_id", Value: "s1"},
		{Key: "run_id", Value: "r1"},
	}
	h.RunDetail(c)
	assert.NotEqual(t, http.StatusOK, recorder.Code)
}

func TestRunDetail_StartTimeGtEndTime(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSvc := iportdrivermock.NewMockIObservability(ctrl)
	h := &observabilityHTTPHandler{observabilitySvc: mockSvc, logger: obsTestLogger{}}

	c, recorder := newObsCtx(http.MethodPost, "/", `{"start_time":100,"end_time":50}`)
	c.Params = gin.Params{
		{Key: "agent_id", Value: "a1"},
		{Key: "conversation_id", Value: "c1"},
		{Key: "session_id", Value: "s1"},
		{Key: "run_id", Value: "r1"},
	}
	h.RunDetail(c)
	assert.NotEqual(t, http.StatusOK, recorder.Code)
}

func TestRunDetail_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSvc := iportdrivermock.NewMockIObservability(ctrl)
	h := &observabilityHTTPHandler{observabilitySvc: mockSvc, logger: obsTestLogger{}}

	c, recorder := newObsCtx(http.MethodPost, "/", `{"start_time":1,"end_time":100}`)
	c.Params = gin.Params{
		{Key: "agent_id", Value: "a1"},
		{Key: "conversation_id", Value: "c1"},
		{Key: "session_id", Value: "s1"},
		{Key: "run_id", Value: "r1"},
	}
	h.RunDetail(c)
	assert.NotEqual(t, http.StatusOK, recorder.Code)
}

func TestRunDetail_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSvc := iportdrivermock.NewMockIObservability(ctrl)
	h := &observabilityHTTPHandler{observabilitySvc: mockSvc, logger: obsTestLogger{}}

	mockSvc.EXPECT().RunDetail(gomock.Any(), gomock.Any()).Return(nil, errors.New("svc failed"))

	c, recorder := newObsCtx(http.MethodPost, "/", `{"start_time":1,"end_time":100}`)
	c.Params = gin.Params{
		{Key: "agent_id", Value: "a1"},
		{Key: "conversation_id", Value: "c1"},
		{Key: "session_id", Value: "s1"},
		{Key: "run_id", Value: "r1"},
	}
	setObsVisitor(c, "u1")
	h.RunDetail(c)
	assert.NotEqual(t, http.StatusOK, recorder.Code)
}

// ==================== SessionDetail — deeper paths ====================

func TestSessionDetail_BindJsonError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSvc := iportdrivermock.NewMockIObservability(ctrl)
	h := &observabilityHTTPHandler{observabilitySvc: mockSvc, logger: obsTestLogger{}}

	c, recorder := newObsCtx(http.MethodPost, "/", "{")
	c.Params = gin.Params{
		{Key: "agent_id", Value: "a1"},
		{Key: "conversation_id", Value: "c1"},
		{Key: "session_id", Value: "s1"},
	}
	h.SessionDetail(c)
	assert.NotEqual(t, http.StatusOK, recorder.Code)
}

func TestSessionDetail_TimeRangeMissing(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSvc := iportdrivermock.NewMockIObservability(ctrl)
	h := &observabilityHTTPHandler{observabilitySvc: mockSvc, logger: obsTestLogger{}}

	c, recorder := newObsCtx(http.MethodPost, "/", `{}`)
	c.Params = gin.Params{
		{Key: "agent_id", Value: "a1"},
		{Key: "conversation_id", Value: "c1"},
		{Key: "session_id", Value: "s1"},
	}
	h.SessionDetail(c)
	assert.NotEqual(t, http.StatusOK, recorder.Code)
}

func TestSessionDetail_StartTimeGtEndTime(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSvc := iportdrivermock.NewMockIObservability(ctrl)
	h := &observabilityHTTPHandler{observabilitySvc: mockSvc, logger: obsTestLogger{}}

	c, recorder := newObsCtx(http.MethodPost, "/", `{"start_time":100,"end_time":50}`)
	c.Params = gin.Params{
		{Key: "agent_id", Value: "a1"},
		{Key: "conversation_id", Value: "c1"},
		{Key: "session_id", Value: "s1"},
	}
	h.SessionDetail(c)
	assert.NotEqual(t, http.StatusOK, recorder.Code)
}

func TestSessionDetail_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSvc := iportdrivermock.NewMockIObservability(ctrl)
	h := &observabilityHTTPHandler{observabilitySvc: mockSvc, logger: obsTestLogger{}}

	c, recorder := newObsCtx(http.MethodPost, "/", `{"start_time":1,"end_time":100}`)
	c.Params = gin.Params{
		{Key: "agent_id", Value: "a1"},
		{Key: "conversation_id", Value: "c1"},
		{Key: "session_id", Value: "s1"},
	}
	h.SessionDetail(c)
	assert.NotEqual(t, http.StatusOK, recorder.Code)
}

func TestSessionDetail_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSvc := iportdrivermock.NewMockIObservability(ctrl)
	h := &observabilityHTTPHandler{observabilitySvc: mockSvc, logger: obsTestLogger{}}

	mockSvc.EXPECT().SessionDetail(gomock.Any(), gomock.Any()).Return(nil, errors.New("svc failed"))

	c, recorder := newObsCtx(http.MethodPost, "/", `{"start_time":1,"end_time":100}`)
	c.Params = gin.Params{
		{Key: "agent_id", Value: "a1"},
		{Key: "conversation_id", Value: "c1"},
		{Key: "session_id", Value: "s1"},
	}
	setObsVisitor(c, "u1")
	h.SessionDetail(c)
	assert.NotEqual(t, http.StatusOK, recorder.Code)
}

// ==================== RunList — deeper paths ====================

func TestRunList_ConversationIDEmpty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSvc := iportdrivermock.NewMockIObservability(ctrl)
	h := &observabilityHTTPHandler{observabilitySvc: mockSvc, logger: obsTestLogger{}}

	c, recorder := newObsCtx(http.MethodPost, "/", `{}`)
	c.Params = gin.Params{{Key: "agent_id", Value: "a1"}}
	h.RunList(c)
	assert.NotEqual(t, http.StatusOK, recorder.Code)
}

func TestRunList_BindJsonError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSvc := iportdrivermock.NewMockIObservability(ctrl)
	h := &observabilityHTTPHandler{observabilitySvc: mockSvc, logger: obsTestLogger{}}

	c, recorder := newObsCtx(http.MethodPost, "/", "{")
	c.Params = gin.Params{
		{Key: "agent_id", Value: "a1"},
		{Key: "conversation_id", Value: "c1"},
		{Key: "session_id", Value: "s1"},
	}
	h.RunList(c)
	assert.NotEqual(t, http.StatusOK, recorder.Code)
}

func TestRunList_TimeRangeMissing(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSvc := iportdrivermock.NewMockIObservability(ctrl)
	h := &observabilityHTTPHandler{observabilitySvc: mockSvc, logger: obsTestLogger{}}

	c, recorder := newObsCtx(http.MethodPost, "/", `{}`)
	c.Params = gin.Params{
		{Key: "agent_id", Value: "a1"},
		{Key: "conversation_id", Value: "c1"},
		{Key: "session_id", Value: "s1"},
	}
	h.RunList(c)
	assert.NotEqual(t, http.StatusOK, recorder.Code)
}

func TestRunList_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSvc := iportdrivermock.NewMockIObservability(ctrl)
	h := &observabilityHTTPHandler{observabilitySvc: mockSvc, logger: obsTestLogger{}}

	c, recorder := newObsCtx(http.MethodPost, "/", `{"start_time":1,"end_time":100}`)
	c.Params = gin.Params{
		{Key: "agent_id", Value: "a1"},
		{Key: "conversation_id", Value: "c1"},
		{Key: "session_id", Value: "s1"},
	}
	h.RunList(c)
	assert.NotEqual(t, http.StatusOK, recorder.Code)
}

func TestRunList_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSvc := iportdrivermock.NewMockIObservability(ctrl)
	h := &observabilityHTTPHandler{observabilitySvc: mockSvc, logger: obsTestLogger{}}

	mockSvc.EXPECT().RunList(gomock.Any(), gomock.Any()).Return(nil, errors.New("svc failed"))

	c, recorder := newObsCtx(http.MethodPost, "/", `{"start_time":1,"end_time":100}`)
	c.Params = gin.Params{
		{Key: "agent_id", Value: "a1"},
		{Key: "conversation_id", Value: "c1"},
		{Key: "session_id", Value: "s1"},
	}
	setObsVisitor(c, "u1")
	h.RunList(c)
	assert.NotEqual(t, http.StatusOK, recorder.Code)
}

// ==================== SessionList — deeper paths ====================

func TestSessionList_BindJsonError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSvc := iportdrivermock.NewMockIObservability(ctrl)
	h := &observabilityHTTPHandler{observabilitySvc: mockSvc, logger: obsTestLogger{}}

	c, recorder := newObsCtx(http.MethodPost, "/", "{")
	c.Params = gin.Params{{Key: "agent_id", Value: "a1"}, {Key: "conversation_id", Value: "c1"}}
	h.SessionList(c)
	assert.NotEqual(t, http.StatusOK, recorder.Code)
}

func TestSessionList_StartTimeGtEndTime(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSvc := iportdrivermock.NewMockIObservability(ctrl)
	h := &observabilityHTTPHandler{observabilitySvc: mockSvc, logger: obsTestLogger{}}

	c, recorder := newObsCtx(http.MethodPost, "/", `{"start_time":100,"end_time":50}`)
	c.Params = gin.Params{{Key: "agent_id", Value: "a1"}, {Key: "conversation_id", Value: "c1"}}
	h.SessionList(c)
	assert.NotEqual(t, http.StatusOK, recorder.Code)
}

func TestSessionList_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSvc := iportdrivermock.NewMockIObservability(ctrl)
	h := &observabilityHTTPHandler{observabilitySvc: mockSvc, logger: obsTestLogger{}}

	c, recorder := newObsCtx(http.MethodPost, "/", `{"start_time":1,"end_time":100}`)
	c.Params = gin.Params{{Key: "agent_id", Value: "a1"}, {Key: "conversation_id", Value: "c1"}}
	h.SessionList(c)
	assert.NotEqual(t, http.StatusOK, recorder.Code)
}

func TestSessionList_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSvc := iportdrivermock.NewMockIObservability(ctrl)
	h := &observabilityHTTPHandler{observabilitySvc: mockSvc, logger: obsTestLogger{}}

	mockSvc.EXPECT().SessionList(gomock.Any(), gomock.Any()).Return(nil, errors.New("svc failed"))

	c, recorder := newObsCtx(http.MethodPost, "/", `{"start_time":1,"end_time":100}`)
	c.Params = gin.Params{{Key: "agent_id", Value: "a1"}, {Key: "conversation_id", Value: "c1"}}
	setObsVisitor(c, "u1")
	h.SessionList(c)
	assert.NotEqual(t, http.StatusOK, recorder.Code)
}

// ==================== ConversationList — time range zero ====================

func TestConversationList_TimeRangeMissing(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSvc := iportdrivermock.NewMockIObservability(ctrl)
	h := &observabilityHTTPHandler{observabilitySvc: mockSvc, logger: obsTestLogger{}}

	c, recorder := newObsCtx(http.MethodPost, "/", `{}`)
	c.Params = gin.Params{{Key: "agent_id", Value: "a1"}}
	h.ConversationList(c)
	assert.NotEqual(t, http.StatusOK, recorder.Code)
}

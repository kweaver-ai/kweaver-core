package agentsvc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/daconfvalobj"
	agentreq "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent/req"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent/req/chatopt"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp/cmpmock"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess/idbaccessmock"
)

// AfterProcess: IsNeedDocRetrivalPostProcess=true branch
func TestAfterProcess_IsNeedDocRetrivalPostProcess(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
	mockConvRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)
	allowAnyLoggerCalls(mockLogger)

	svc := &agentSvc{
		SvcBase:             service.NewSvcBase(),
		conversationMsgRepo: mockMsgRepo,
		conversationRepo:    mockConvRepo,
		logger:              mockLogger,
	}

	agent := newTestAgent()
	agent.Config.Output = &daconfvalobj.Output{
		Variables: &daconfvalobj.VariablesS{AnswerVar: "answer"},
	}

	req := &agentreq.ChatReq{
		AgentID:        "a1",
		ConversationID: "conv-ap-doc",
		InternalParam:  agentreq.InternalParam{UserID: "u1", AssistantMessageID: "asst-ap-doc"},
		ChatOption:     chatopt.ChatOption{IsNeedDocRetrivalPostProcess: true},
	}
	callResult := []byte(`{"status":"False","answer":{"final_answer":"partial"}}`)

	result, isEnd, err := svc.AfterProcess(context.Background(), callResult, req, agent, 0)
	assert.NoError(t, err)
	assert.False(t, isEnd)
	assert.NotNil(t, result)
}

// AfterProcess: IsNeedProgress=false → progressAns cleared
func TestAfterProcess_IsNeedProgressFalse(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
	mockConvRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)
	allowAnyLoggerCalls(mockLogger)

	svc := &agentSvc{
		SvcBase:             service.NewSvcBase(),
		conversationMsgRepo: mockMsgRepo,
		conversationRepo:    mockConvRepo,
		logger:              mockLogger,
	}

	agent := newTestAgent()
	agent.Config.Output = &daconfvalobj.Output{
		Variables: &daconfvalobj.VariablesS{AnswerVar: "answer"},
	}

	req := &agentreq.ChatReq{
		AgentID:        "a1",
		ConversationID: "conv-ap-noprog",
		InternalParam:  agentreq.InternalParam{UserID: "u1", AssistantMessageID: "asst-ap-noprog"},
		ChatOption:     chatopt.ChatOption{IsNeedProgress: false},
	}
	callResult := []byte(`{"status":"False","answer":{"final_answer":"partial"}}`)

	result, isEnd, err := svc.AfterProcess(context.Background(), callResult, req, agent, 0)
	assert.NoError(t, err)
	assert.False(t, isEnd)
	assert.NotNil(t, result)
}

// AfterProcess: status=True with TTFT already set (skip CalculateTTFT)
func TestAfterProcess_StatusTrue_TTFTAlreadySet(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
	mockConvRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)
	allowAnyLoggerCalls(mockLogger)

	mockMsgRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockConvRepo.EXPECT().GetByID(gomock.Any(), gomock.Any()).Return(&dapo.ConversationPO{ID: "conv-ap-ttft"}, nil).AnyTimes()
	mockConvRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	svc := &agentSvc{
		SvcBase:             service.NewSvcBase(),
		conversationMsgRepo: mockMsgRepo,
		conversationRepo:    mockConvRepo,
		logger:              mockLogger,
	}

	agent := newTestAgent()
	agent.Config.Output = &daconfvalobj.Output{
		Variables: &daconfvalobj.VariablesS{AnswerVar: "answer"},
	}

	req := &agentreq.ChatReq{
		AgentID:        "a1",
		ConversationID: "conv-ap-ttft",
		InternalParam:  agentreq.InternalParam{UserID: "u1", AssistantMessageID: "asst-ap-ttft", TTFT: 100}, // TTFT already set → skip CalculateTTFT
	}
	callResult := []byte(`{"status":"True","answer":{"final_answer":"done"}}`)

	result, isEnd, err := svc.AfterProcess(context.Background(), callResult, req, agent, 0)
	assert.NoError(t, err)
	assert.True(t, isEnd)
	assert.NotNil(t, result)
}

// handleMessageAndTempArea: conversationRepo.GetByID fails
func TestAfterProcess_HandleMsgArea_ConvGetByIDFails(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
	mockConvRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)
	allowAnyLoggerCalls(mockLogger)

	mockMsgRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockConvRepo.EXPECT().GetByID(gomock.Any(), gomock.Any()).Return(nil, assert.AnError).AnyTimes()

	svc := &agentSvc{
		SvcBase:             service.NewSvcBase(),
		conversationMsgRepo: mockMsgRepo,
		conversationRepo:    mockConvRepo,
		logger:              mockLogger,
	}

	agent := newTestAgent()
	agent.Config.Output = &daconfvalobj.Output{
		Variables: &daconfvalobj.VariablesS{AnswerVar: "answer"},
	}

	req := &agentreq.ChatReq{
		AgentID:        "a1",
		ConversationID: "conv-ap-convfail",
		InternalParam:  agentreq.InternalParam{UserID: "u1", AssistantMessageID: "asst-ap-convfail"},
	}
	callResult := []byte(`{"status":"True","answer":{"final_answer":"done"}}`)

	result, isEnd, err := svc.AfterProcess(context.Background(), callResult, req, agent, 0)
	assert.Error(t, err)
	assert.False(t, isEnd)
	assert.NotNil(t, result)
}

// handleMessageAndTempArea: conversationRepo.Update fails
func TestAfterProcess_HandleMsgArea_ConvUpdateFails(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
	mockConvRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)
	allowAnyLoggerCalls(mockLogger)

	mockMsgRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockConvRepo.EXPECT().GetByID(gomock.Any(), gomock.Any()).Return(&dapo.ConversationPO{ID: "conv-ap-cupd"}, nil).AnyTimes()
	mockConvRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(assert.AnError).AnyTimes()

	svc := &agentSvc{
		SvcBase:             service.NewSvcBase(),
		conversationMsgRepo: mockMsgRepo,
		conversationRepo:    mockConvRepo,
		logger:              mockLogger,
	}

	agent := newTestAgent()
	agent.Config.Output = &daconfvalobj.Output{
		Variables: &daconfvalobj.VariablesS{AnswerVar: "answer"},
	}

	req := &agentreq.ChatReq{
		AgentID:        "a1",
		ConversationID: "conv-ap-cupd",
		InternalParam:  agentreq.InternalParam{UserID: "u1", AssistantMessageID: "asst-ap-cupd"},
	}
	callResult := []byte(`{"status":"True","answer":{"final_answer":"done"}}`)

	result, isEnd, err := svc.AfterProcess(context.Background(), callResult, req, agent, 0)
	assert.Error(t, err)
	assert.False(t, isEnd)
	assert.NotNil(t, result)
}

// AfterProcess: status=True with IsNeedProgress=true (progressAns kept)
func TestAfterProcess_StatusTrue_IsNeedProgress(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
	mockConvRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)
	allowAnyLoggerCalls(mockLogger)

	mockMsgRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockConvRepo.EXPECT().GetByID(gomock.Any(), gomock.Any()).Return(&dapo.ConversationPO{ID: "conv-ap-prog"}, nil).AnyTimes()
	mockConvRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	svc := &agentSvc{
		SvcBase:             service.NewSvcBase(),
		conversationMsgRepo: mockMsgRepo,
		conversationRepo:    mockConvRepo,
		logger:              mockLogger,
	}

	agent := newTestAgent()
	agent.Config.Output = &daconfvalobj.Output{
		Variables: &daconfvalobj.VariablesS{AnswerVar: "answer"},
	}

	req := &agentreq.ChatReq{
		AgentID:        "a1",
		ConversationID: "conv-ap-prog",
		InternalParam:  agentreq.InternalParam{UserID: "u1", AssistantMessageID: "asst-ap-prog"},
		ChatOption:     chatopt.ChatOption{IsNeedProgress: true},
	}
	callResult := []byte(`{"status":"True","answer":{"final_answer":"done"}}`)

	result, isEnd, err := svc.AfterProcess(context.Background(), callResult, req, agent, 0)
	assert.NoError(t, err)
	assert.True(t, isEnd)
	assert.NotNil(t, result)
}

// AfterProcess: status=Error with error map containing error_code
func TestAfterProcess_StatusError_WithErrorCode(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
	mockConvRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)
	allowAnyLoggerCalls(mockLogger)

	mockMsgRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockConvRepo.EXPECT().GetByID(gomock.Any(), gomock.Any()).Return(&dapo.ConversationPO{ID: "conv-ap-errcode"}, nil).AnyTimes()
	mockConvRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	svc := &agentSvc{
		SvcBase:             service.NewSvcBase(),
		conversationMsgRepo: mockMsgRepo,
		conversationRepo:    mockConvRepo,
		logger:              mockLogger,
	}

	agent := newTestAgent()
	agent.Config.Output = &daconfvalobj.Output{
		Variables: &daconfvalobj.VariablesS{AnswerVar: "answer"},
	}

	req := &agentreq.ChatReq{
		AgentID:        "a1",
		ConversationID: "conv-ap-errcode",
		InternalParam:  agentreq.InternalParam{UserID: "u1", AssistantMessageID: "asst-ap-errcode"},
	}
	callResult := []byte(`{"status":"Error","answer":{"final_answer":""},"error":{"error_code":"AgentExecutor.DolphinSDKException.ModelExecption","error_details":"model error"}}`)

	result, _, err := svc.AfterProcess(context.Background(), callResult, req, agent, 0)
	assert.Error(t, err)
	assert.NotNil(t, result)
}

// TransformErrorToHTTPError: non-map error
func TestTransformErrorToHTTPError_NonMap(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	httpErr := TransformErrorToHTTPError(ctx, "plain string error")
	assert.NotNil(t, httpErr)
}

// TransformErrorToHTTPError: map without error_code
func TestTransformErrorToHTTPError_MapNoErrorCode(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	httpErr := TransformErrorToHTTPError(ctx, map[string]interface{}{"other": "value"})
	assert.NotNil(t, httpErr)
}

// TransformErrorToHTTPError: map with non-string error_code
func TestTransformErrorToHTTPError_MapNonStringErrorCode(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	httpErr := TransformErrorToHTTPError(ctx, map[string]interface{}{"error_code": 123})
	assert.NotNil(t, httpErr)
}

// TransformErrorToHTTPError: SkillExecption
func TestTransformErrorToHTTPError_SkillExecption(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	httpErr := TransformErrorToHTTPError(ctx, map[string]interface{}{
		"error_code":    "AgentExecutor.DolphinSDKException.SkillExecption",
		"error_details": "skill error",
	})
	assert.NotNil(t, httpErr)
}

// TransformErrorToHTTPError: BaseExecption
func TestTransformErrorToHTTPError_BaseExecption(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	httpErr := TransformErrorToHTTPError(ctx, map[string]interface{}{
		"error_code":    "AgentExecutor.DolphinSDKException.BaseExecption",
		"error_details": "base error",
	})
	assert.NotNil(t, httpErr)
}

// TransformErrorToHTTPError: default error code
func TestTransformErrorToHTTPError_DefaultCode(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	httpErr := TransformErrorToHTTPError(ctx, map[string]interface{}{
		"error_code":    "SomeOtherCode",
		"error_details": "other error",
	})
	assert.NotNil(t, httpErr)
}

// TransformErrorToHTTPError: nil map
func TestTransformErrorToHTTPError_NilMap(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	var nilMap map[string]interface{}
	httpErr := TransformErrorToHTTPError(ctx, nilMap)
	assert.Nil(t, httpErr)
}

package agentsvc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/daconfvalobj"
	agentreq "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent/req"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/square/squareresp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp/cmpmock"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess/idbaccessmock"
)

func newAfterProcessSvc(ctrl *gomock.Controller, mockMsgRepo *idbaccessmock.MockIConversationMsgRepo, mockConvRepo *idbaccessmock.MockIConversationRepo, mockLogger *cmpmock.MockLogger) *agentSvc {
	return &agentSvc{
		SvcBase:             service.NewSvcBase(),
		conversationMsgRepo: mockMsgRepo,
		conversationRepo:    mockConvRepo,
		logger:              mockLogger,
	}
}

func newAfterProcessAgent() *squareresp.AgentMarketAgentInfoResp { //nolint:unused
	agent := newTestAgent()
	agent.Config.Output = &daconfvalobj.Output{
		Variables: &daconfvalobj.VariablesS{AnswerVar: "answer"},
	}

	return agent
}

// AfterProcess: status=False, no handleMessageAndTempArea → success
func TestAfterProcess_StatusFalse_Full(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
	mockConvRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)
	allowAnyLoggerCalls(mockLogger)

	svc := newAfterProcessSvc(ctrl, mockMsgRepo, mockConvRepo, mockLogger)

	agent := newTestAgent()
	agent.Config.Output = &daconfvalobj.Output{
		Variables: &daconfvalobj.VariablesS{AnswerVar: "answer"},
	}

	req := &agentreq.ChatReq{
		AgentID:        "a1",
		ConversationID: "conv-ap-1",
		InternalParam:  agentreq.InternalParam{UserID: "u1", AssistantMessageID: "asst-ap-1"},
	}
	callResult := []byte(`{"status":"False","answer":{"final_answer":"partial answer"}}`)

	result, isEnd, err := svc.AfterProcess(context.Background(), callResult, req, agent, 0)
	assert.NoError(t, err)
	assert.False(t, isEnd)
	assert.NotNil(t, result)
}

// AfterProcess: status=True → calls handleMessageAndTempArea
func TestAfterProcess_StatusTrue_Full(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
	mockConvRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)
	allowAnyLoggerCalls(mockLogger)

	// handleMessageAndTempArea calls Update + GetByID + Update
	mockMsgRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockConvRepo.EXPECT().GetByID(gomock.Any(), gomock.Any()).Return(&dapo.ConversationPO{ID: "conv-ap-2"}, nil).AnyTimes()
	mockConvRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	svc := newAfterProcessSvc(ctrl, mockMsgRepo, mockConvRepo, mockLogger)

	agent := newTestAgent()
	agent.Config.Output = &daconfvalobj.Output{
		Variables: &daconfvalobj.VariablesS{AnswerVar: "answer"},
	}

	req := &agentreq.ChatReq{
		AgentID:        "a1",
		ConversationID: "conv-ap-2",
		InternalParam:  agentreq.InternalParam{UserID: "u1", AssistantMessageID: "asst-ap-2"},
	}
	callResult := []byte(`{"status":"True","answer":{"final_answer":"final answer"}}`)

	result, isEnd, err := svc.AfterProcess(context.Background(), callResult, req, agent, 0)
	assert.NoError(t, err)
	assert.True(t, isEnd)
	assert.NotNil(t, result)
}

// AfterProcess: status=Error → calls handleMessageAndTempArea then returns error
func TestAfterProcess_StatusError_Full(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
	mockConvRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)
	allowAnyLoggerCalls(mockLogger)

	// status=Error → isEnd=true → handleMessageAndTempArea is called first
	mockMsgRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockConvRepo.EXPECT().GetByID(gomock.Any(), gomock.Any()).Return(&dapo.ConversationPO{ID: "conv-ap-3"}, nil).AnyTimes()
	mockConvRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	svc := newAfterProcessSvc(ctrl, mockMsgRepo, mockConvRepo, mockLogger)

	agent := newTestAgent()
	agent.Config.Output = &daconfvalobj.Output{
		Variables: &daconfvalobj.VariablesS{AnswerVar: "answer"},
	}

	req := &agentreq.ChatReq{
		AgentID:        "a1",
		ConversationID: "conv-ap-3",
		InternalParam:  agentreq.InternalParam{UserID: "u1", AssistantMessageID: "asst-ap-3"},
	}
	callResult := []byte(`{"status":"Error","answer":{"final_answer":""},"error":"something went wrong"}`)

	result, _, err := svc.AfterProcess(context.Background(), callResult, req, agent, 0)
	assert.Error(t, err)
	assert.NotNil(t, result)
}

// AfterProcess: handleMessageAndTempArea fails
func TestAfterProcess_HandleMsgAreaError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
	mockConvRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)
	allowAnyLoggerCalls(mockLogger)

	// handleMessageAndTempArea → Update fails
	mockMsgRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(assert.AnError).AnyTimes()
	mockConvRepo.EXPECT().GetByID(gomock.Any(), gomock.Any()).Return(&dapo.ConversationPO{}, nil).AnyTimes()
	mockConvRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	svc := newAfterProcessSvc(ctrl, mockMsgRepo, mockConvRepo, mockLogger)

	agent := newTestAgent()
	agent.Config.Output = &daconfvalobj.Output{
		Variables: &daconfvalobj.VariablesS{AnswerVar: "answer"},
	}

	req := &agentreq.ChatReq{
		AgentID:        "a1",
		ConversationID: "conv-ap-4",
		InternalParam:  agentreq.InternalParam{UserID: "u1", AssistantMessageID: "asst-ap-4"},
	}
	callResult := []byte(`{"status":"True","answer":{"final_answer":"final"}}`)

	result, isEnd, err := svc.AfterProcess(context.Background(), callResult, req, agent, 0)
	assert.Error(t, err)
	assert.False(t, isEnd)
	assert.NotNil(t, result)
}

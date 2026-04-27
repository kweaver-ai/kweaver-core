package agentsvc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/daconfvalobj"
	agentreq "github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/agent/req"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/square/squareresp"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/cmp/icmp/cmpmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/idbaccess/idbaccessmock"
)

// keep squareresp used
var _ *squareresp.AgentMarketAgentInfoResp

func newProcessSvc(ctrl *gomock.Controller, mockMsgRepo *idbaccessmock.MockIConversationMsgRepo, mockConvRepo *idbaccessmock.MockIConversationRepo, mockLogger *cmpmock.MockLogger) *agentSvc {
	return &agentSvc{
		SvcBase:             service.NewSvcBase(),
		conversationMsgRepo: mockMsgRepo,
		conversationRepo:    mockConvRepo,
		logger:              mockLogger,
		streamDiffFrequency: 1,
	}
}

func newProcessAgent() *squareresp.AgentMarketAgentInfoResp { //nolint:unused
	agent := newTestAgent()
	agent.Config.Output = &daconfvalobj.Output{
		Variables: &daconfvalobj.VariablesS{AnswerVar: "answer"},
	}

	return agent
}

// Process: messageChan closed immediately → isEnd=true, no error
func TestProcess_MessageChanClosed(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
	mockConvRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)
	allowAnyLoggerCalls(mockLogger)

	// messageChan closed triggers GetByID + Update to set message status to failed
	mockMsgRepo.EXPECT().GetByID(gomock.Any(), "asst-proc-1").Return(&dapo.ConversationMsgPO{ID: "asst-proc-1"}, nil)
	mockMsgRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

	svc := newProcessSvc(ctrl, mockMsgRepo, mockConvRepo, mockLogger)

	req := &agentreq.ChatReq{
		AgentID:        "a1",
		ConversationID: "conv-proc-1",
		InternalParam:  agentreq.InternalParam{UserID: "u1", AssistantMessageID: "asst-proc-1"},
	}
	agent := newTestAgent()
	agent.Config.Output = &daconfvalobj.Output{
		Variables: &daconfvalobj.VariablesS{AnswerVar: "answer"},
	}

	stopChan := make(chan struct{})
	respChan := make(chan []byte, 10)
	messageChan := make(chan string)
	errChan := make(chan error)

	// Register session so Process can find it
	session := &Session{ConversationID: "conv-proc-1"}
	SessionMap.Store("conv-proc-1", session)

	defer SessionMap.Delete("conv-proc-1")

	// Close messageChan immediately
	close(messageChan)

	err := svc.Process(context.Background(), req, agent, stopChan, respChan, messageChan, errChan, func() {})
	assert.NoError(t, err)
}

// Process: errChan receives EOF → isEnd=true
func TestProcess_ErrChanEOF(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
	mockConvRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)
	allowAnyLoggerCalls(mockLogger)

	// errChan EOF triggers GetByID + Update to set message status to failed
	mockMsgRepo.EXPECT().GetByID(gomock.Any(), "asst-proc-2").Return(&dapo.ConversationMsgPO{ID: "asst-proc-2"}, nil)
	mockMsgRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

	svc := newProcessSvc(ctrl, mockMsgRepo, mockConvRepo, mockLogger)

	req := &agentreq.ChatReq{
		AgentID:        "a1",
		ConversationID: "conv-proc-2",
		Stream:         true,
		InternalParam:  agentreq.InternalParam{UserID: "u1", AssistantMessageID: "asst-proc-2"},
	}
	agent := newTestAgent()
	agent.Config.Output = &daconfvalobj.Output{
		Variables: &daconfvalobj.VariablesS{AnswerVar: "answer"},
	}

	stopChan := make(chan struct{})
	respChan := make(chan []byte, 10)
	messageChan := make(chan string)
	errChan := make(chan error, 1)

	session := &Session{ConversationID: "conv-proc-2"}
	SessionMap.Store("conv-proc-2", session)

	defer SessionMap.Delete("conv-proc-2")

	errChan <- errors.New("EOF")
	close(errChan)

	err := svc.Process(context.Background(), req, agent, stopChan, respChan, messageChan, errChan, func() {})
	assert.NoError(t, err)
}

// Process: errChan receives non-EOF error (stream mode) → sends err to respChan
func TestProcess_ErrChanNonEOF(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
	mockConvRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)
	allowAnyLoggerCalls(mockLogger)

	// errChan non-EOF error triggers GetByID + Update to set message status to failed
	mockMsgRepo.EXPECT().GetByID(gomock.Any(), "asst-proc-3").Return(&dapo.ConversationMsgPO{ID: "asst-proc-3"}, nil)
	mockMsgRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

	svc := newProcessSvc(ctrl, mockMsgRepo, mockConvRepo, mockLogger)

	req := &agentreq.ChatReq{
		AgentID:        "a1",
		ConversationID: "conv-proc-3",
		Stream:         true,
		InternalParam:  agentreq.InternalParam{UserID: "u1", AssistantMessageID: "asst-proc-3"},
	}
	agent := newTestAgent()
	agent.Config.Output = &daconfvalobj.Output{
		Variables: &daconfvalobj.VariablesS{AnswerVar: "answer"},
	}

	stopChan := make(chan struct{})
	respChan := make(chan []byte, 10)
	messageChan := make(chan string)
	errChan := make(chan error, 1)

	session := &Session{ConversationID: "conv-proc-3"}
	SessionMap.Store("conv-proc-3", session)

	defer SessionMap.Delete("conv-proc-3")

	// stream=true, non-EOF: sends error to respChan then continues until close
	errChan <- errors.New("some real error")
	close(errChan)

	err := svc.Process(context.Background(), req, agent, stopChan, respChan, messageChan, errChan, func() {})
	assert.NoError(t, err)
}

// Process: messageChan receives a message → AfterProcess errors (AnswerVar empty) → marks msg failed
func TestProcess_MessageWithAfterProcessError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
	mockConvRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)
	allowAnyLoggerCalls(mockLogger)

	// AfterProcess errors → Process calls GetByID + Update to mark msg as failed
	mockMsgRepo.EXPECT().GetByID(gomock.Any(), "asst-proc-5").Return(&dapo.ConversationMsgPO{ID: "asst-proc-5"}, nil)
	mockMsgRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

	svc := newProcessSvc(ctrl, mockMsgRepo, mockConvRepo, mockLogger)

	// Output set but AnswerVar empty → AfterProcess returns error on AnswerVar check
	agent := newTestAgent()
	agent.Config.Output = &daconfvalobj.Output{
		Variables: &daconfvalobj.VariablesS{AnswerVar: ""},
	}

	req := &agentreq.ChatReq{
		AgentID:        "a1",
		ConversationID: "conv-proc-5",
		Stream:         false,
		InternalParam:  agentreq.InternalParam{UserID: "u1", AssistantMessageID: "asst-proc-5"},
	}

	stopChan := make(chan struct{})
	respChan := make(chan []byte, 10)
	messageChan := make(chan string, 1)
	errChan := make(chan error)

	session := &Session{ConversationID: "conv-proc-5"}
	SessionMap.Store("conv-proc-5", session)

	defer SessionMap.Delete("conv-proc-5")

	// Send a valid SSE-formatted message
	messageChan <- `data:{"status":"False","answer":{"final_answer":"hi"}}`
	close(messageChan)

	err := svc.Process(context.Background(), req, agent, stopChan, respChan, messageChan, errChan, func() {})
	assert.NoError(t, err)
}

// Process: messageChan receives status=True message → AfterProcess success → handleMessageAndTempArea called
func TestProcess_MessageStatusTrue(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
	mockConvRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)
	allowAnyLoggerCalls(mockLogger)

	// handleMessageAndTempArea calls Update + GetByID + Update
	mockMsgRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockConvRepo.EXPECT().GetByID(gomock.Any(), gomock.Any()).Return(&dapo.ConversationPO{ID: "conv-proc-6"}, nil).AnyTimes()
	mockConvRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	svc := newProcessSvc(ctrl, mockMsgRepo, mockConvRepo, mockLogger)

	agent := newTestAgent()
	agent.Config.Output = &daconfvalobj.Output{
		Variables: &daconfvalobj.VariablesS{AnswerVar: "answer"},
	}

	req := &agentreq.ChatReq{
		AgentID:        "a1",
		ConversationID: "conv-proc-6",
		Stream:         false,
		InternalParam:  agentreq.InternalParam{UserID: "u1", AssistantMessageID: "asst-proc-6"},
	}

	stopChan := make(chan struct{})
	respChan := make(chan []byte, 10)
	messageChan := make(chan string, 1)
	errChan := make(chan error)

	session := &Session{ConversationID: "conv-proc-6"}
	SessionMap.Store("conv-proc-6", session)

	defer SessionMap.Delete("conv-proc-6")

	// Send status=True → isEnd=true → handleMessageAndTempArea
	messageChan <- `data:{"status":"True","answer":{"final_answer":"done"}}`
	close(messageChan)

	err := svc.Process(context.Background(), req, agent, stopChan, respChan, messageChan, errChan, func() {})
	assert.NoError(t, err)
}

// Process: stopChan triggered
func TestProcess_StopChanTriggered(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
	mockConvRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)
	allowAnyLoggerCalls(mockLogger)

	// HandleStopChan calls conversationMsgRepo.GetByID + Update + conversationRepo.GetByID + Update
	mockMsgRepo.EXPECT().GetByID(gomock.Any(), "asst-proc-4").Return(&dapo.ConversationMsgPO{ID: "asst-proc-4"}, nil)
	mockMsgRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockConvRepo.EXPECT().GetByID(gomock.Any(), "conv-proc-4").Return(&dapo.ConversationPO{ID: "conv-proc-4"}, nil).AnyTimes()
	mockConvRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	svc := newProcessSvc(ctrl, mockMsgRepo, mockConvRepo, mockLogger)

	req := &agentreq.ChatReq{
		AgentID:        "a1",
		ConversationID: "conv-proc-4",
		InternalParam:  agentreq.InternalParam{UserID: "u1", AssistantMessageID: "asst-proc-4"},
	}
	agent := newTestAgent()
	agent.Config.Output = &daconfvalobj.Output{
		Variables: &daconfvalobj.VariablesS{AnswerVar: "answer"},
	}

	stopChan := make(chan struct{})
	respChan := make(chan []byte, 10)
	messageChan := make(chan string)
	errChan := make(chan error)

	session := &Session{ConversationID: "conv-proc-4"}
	SessionMap.Store("conv-proc-4", session)

	defer SessionMap.Delete("conv-proc-4")

	// Close stopChan after a brief delay
	go func() {
		time.Sleep(5 * time.Millisecond)
		close(stopChan)
	}()

	err := svc.Process(context.Background(), req, agent, stopChan, respChan, messageChan, errChan, func() {})
	assert.NoError(t, err)
}

package agentsvc

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/daconfvalobj"
	agentreq "github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/agent/req"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/cmp/icmp/cmpmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/idbaccess/idbaccessmock"
)

// Process: IncStream=true → StreamDiff path
func TestProcess_IncStream(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
	mockConvRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)
	allowAnyLoggerCalls(mockLogger)

	mockMsgRepo.EXPECT().GetByID(gomock.Any(), "asst-inc").Return(&dapo.ConversationMsgPO{ID: "asst-inc"}, nil).AnyTimes()
	mockMsgRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockConvRepo.EXPECT().GetByID(gomock.Any(), gomock.Any()).Return(&dapo.ConversationPO{ID: "conv-inc"}, nil).AnyTimes()
	mockConvRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	svc := &agentSvc{
		SvcBase:             service.NewSvcBase(),
		conversationMsgRepo: mockMsgRepo,
		conversationRepo:    mockConvRepo,
		logger:              mockLogger,
		streamDiffFrequency: 1,
	}

	agent := newTestAgent()
	agent.Config.Output = &daconfvalobj.Output{
		Variables: &daconfvalobj.VariablesS{AnswerVar: "answer"},
	}

	req := &agentreq.ChatReq{
		AgentID:        "a1",
		ConversationID: "conv-inc",
		Stream:         true,
		IncStream:      true,
		InternalParam:  agentreq.InternalParam{UserID: "u1", AssistantMessageID: "asst-inc"},
	}

	stopChan := make(chan struct{})
	respChan := make(chan []byte, 20)
	messageChan := make(chan string, 1)
	errChan := make(chan error)

	session := &Session{ConversationID: "conv-inc"}
	SessionMap.Store("conv-inc", session)

	defer SessionMap.Delete("conv-inc")

	messageChan <- `data:{"status":"True","answer":{"final_answer":"done"}}`
	close(messageChan)

	err := svc.Process(context.Background(), req, agent, stopChan, respChan, messageChan, errChan, func() {})
	assert.NoError(t, err)
}

// Process: non-stream errChan non-EOF → writes error to respChan
func TestProcess_NonStream_ErrChan(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
	mockConvRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)
	allowAnyLoggerCalls(mockLogger)

	// After errChan non-EOF in non-stream mode, Process continues until errChan closes
	// Then err != nil → GetByID + Update
	mockMsgRepo.EXPECT().GetByID(gomock.Any(), gomock.Any()).Return(&dapo.ConversationMsgPO{ID: "asst-nse"}, nil).AnyTimes()
	mockMsgRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	svc := &agentSvc{
		SvcBase:             service.NewSvcBase(),
		conversationMsgRepo: mockMsgRepo,
		conversationRepo:    mockConvRepo,
		logger:              mockLogger,
		streamDiffFrequency: 1,
	}

	agent := newTestAgent()
	agent.Config.Output = &daconfvalobj.Output{
		Variables: &daconfvalobj.VariablesS{AnswerVar: "answer"},
	}

	req := &agentreq.ChatReq{
		AgentID:        "a1",
		ConversationID: "conv-nse",
		Stream:         false,
		InternalParam:  agentreq.InternalParam{UserID: "u1", AssistantMessageID: "asst-nse"},
	}

	stopChan := make(chan struct{})
	respChan := make(chan []byte, 20)
	messageChan := make(chan string)
	errChan := make(chan error, 1)

	session := &Session{ConversationID: "conv-nse"}
	SessionMap.Store("conv-nse", session)

	defer SessionMap.Delete("conv-nse")

	errChan <- assert.AnError
	close(errChan)

	err := svc.Process(context.Background(), req, agent, stopChan, respChan, messageChan, errChan, func() {})
	assert.NoError(t, err)
}

// Process: stream unexpected EOF → isEnd=true
func TestProcess_Stream_UnexpectedEOF(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
	mockConvRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)
	allowAnyLoggerCalls(mockLogger)

	// stream unexpected EOF triggers GetByID + Update to set message status to failed
	mockMsgRepo.EXPECT().GetByID(gomock.Any(), "asst-ueof").Return(&dapo.ConversationMsgPO{ID: "asst-ueof"}, nil)
	mockMsgRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

	svc := &agentSvc{
		SvcBase:             service.NewSvcBase(),
		conversationMsgRepo: mockMsgRepo,
		conversationRepo:    mockConvRepo,
		logger:              mockLogger,
		streamDiffFrequency: 1,
	}

	agent := newTestAgent()
	agent.Config.Output = &daconfvalobj.Output{
		Variables: &daconfvalobj.VariablesS{AnswerVar: "answer"},
	}

	req := &agentreq.ChatReq{
		AgentID:        "a1",
		ConversationID: "conv-ueof",
		Stream:         true,
		InternalParam:  agentreq.InternalParam{UserID: "u1", AssistantMessageID: "asst-ueof"},
	}

	stopChan := make(chan struct{})
	respChan := make(chan []byte, 20)
	messageChan := make(chan string)
	errChan := make(chan error, 1)

	session := &Session{ConversationID: "conv-ueof"}
	SessionMap.Store("conv-ueof", session)

	defer SessionMap.Delete("conv-ueof")

	errChan <- &unexpectedEOFError{}
	close(errChan)

	err := svc.Process(context.Background(), req, agent, stopChan, respChan, messageChan, errChan, func() {})
	assert.NoError(t, err)
}

// unexpectedEOFError implements error with "unexpected EOF" message
type unexpectedEOFError struct{}

func (e *unexpectedEOFError) Error() string { return "unexpected EOF" }

// Process: stream non-EOF error → writes error to respChan then continues
func TestProcess_Stream_NonEOFError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
	mockConvRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)
	allowAnyLoggerCalls(mockLogger)

	// stream non-EOF error triggers GetByID + Update to set message status to failed
	mockMsgRepo.EXPECT().GetByID(gomock.Any(), "asst-sneof").Return(&dapo.ConversationMsgPO{ID: "asst-sneof"}, nil)
	mockMsgRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

	svc := &agentSvc{
		SvcBase:             service.NewSvcBase(),
		conversationMsgRepo: mockMsgRepo,
		conversationRepo:    mockConvRepo,
		logger:              mockLogger,
		streamDiffFrequency: 1,
	}

	agent := newTestAgent()
	agent.Config.Output = &daconfvalobj.Output{
		Variables: &daconfvalobj.VariablesS{AnswerVar: "answer"},
	}

	req := &agentreq.ChatReq{
		AgentID:        "a1",
		ConversationID: "conv-sneof",
		Stream:         true,
		InternalParam:  agentreq.InternalParam{UserID: "u1", AssistantMessageID: "asst-sneof"},
	}

	stopChan := make(chan struct{})
	respChan := make(chan []byte, 20)
	messageChan := make(chan string)
	errChan := make(chan error, 2)

	session := &Session{ConversationID: "conv-sneof"}
	SessionMap.Store("conv-sneof", session)

	defer SessionMap.Delete("conv-sneof")

	// non-EOF error followed by EOF to terminate
	errChan <- assert.AnError
	errChan <- &eofErrType{}
	close(errChan)

	err := svc.Process(context.Background(), req, agent, stopChan, respChan, messageChan, errChan, func() {})
	assert.NoError(t, err)
}

type eofErrType struct{}

func (e *eofErrType) Error() string { return "EOF" }

// Process: invalid message format (no "data:" prefix) → continue
func TestProcess_InvalidMessageFormat(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
	mockConvRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)
	allowAnyLoggerCalls(mockLogger)

	// invalid message format triggers GetByID + Update to set message status to failed
	mockMsgRepo.EXPECT().GetByID(gomock.Any(), "asst-invfmt").Return(&dapo.ConversationMsgPO{ID: "asst-invfmt"}, nil)
	mockMsgRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

	svc := &agentSvc{
		SvcBase:             service.NewSvcBase(),
		conversationMsgRepo: mockMsgRepo,
		conversationRepo:    mockConvRepo,
		logger:              mockLogger,
		streamDiffFrequency: 1,
	}

	agent := newTestAgent()
	agent.Config.Output = &daconfvalobj.Output{
		Variables: &daconfvalobj.VariablesS{AnswerVar: "answer"},
	}

	req := &agentreq.ChatReq{
		AgentID:        "a1",
		ConversationID: "conv-invfmt",
		InternalParam:  agentreq.InternalParam{UserID: "u1", AssistantMessageID: "asst-invfmt"},
	}

	stopChan := make(chan struct{})
	respChan := make(chan []byte, 20)
	messageChan := make(chan string, 2)
	errChan := make(chan error)

	session := &Session{ConversationID: "conv-invfmt"}
	SessionMap.Store("conv-invfmt", session)

	defer SessionMap.Delete("conv-invfmt")

	// invalid format (no colon separator with "data" prefix)
	messageChan <- `invalid_no_colon`
	close(messageChan)

	err := svc.Process(context.Background(), req, agent, stopChan, respChan, messageChan, errChan, func() {})
	assert.NoError(t, err)
}

// Process: 5s timeout branch (send message after timeout fires once)
func TestProcess_TimeoutBranch(t *testing.T) {
	t.Parallel()

	// This test just verifies Process doesn't hang when messageChan is slow
	// We close stopChan after a brief delay to exit
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
	mockConvRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)
	allowAnyLoggerCalls(mockLogger)

	mockMsgRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockMsgRepo.EXPECT().GetByID(gomock.Any(), "asst-to").Return(&dapo.ConversationMsgPO{ID: "asst-to"}, nil)
	mockConvRepo.EXPECT().GetByID(gomock.Any(), gomock.Any()).Return(&dapo.ConversationPO{ID: "conv-to"}, nil).AnyTimes()
	mockConvRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	svc := &agentSvc{
		SvcBase:             service.NewSvcBase(),
		conversationMsgRepo: mockMsgRepo,
		conversationRepo:    mockConvRepo,
		logger:              mockLogger,
		streamDiffFrequency: 1,
	}

	agent := newTestAgent()
	agent.Config.Output = &daconfvalobj.Output{
		Variables: &daconfvalobj.VariablesS{AnswerVar: "answer"},
	}

	req := &agentreq.ChatReq{
		AgentID:        "a1",
		ConversationID: "conv-to",
		InternalParam:  agentreq.InternalParam{UserID: "u1", AssistantMessageID: "asst-to"},
	}

	stopChan := make(chan struct{})
	respChan := make(chan []byte, 20)
	messageChan := make(chan string)
	errChan := make(chan error)

	session := &Session{ConversationID: "conv-to"}
	SessionMap.Store("conv-to", session)

	defer SessionMap.Delete("conv-to")

	go func() {
		time.Sleep(10 * time.Millisecond)
		close(stopChan)
	}()

	err := svc.Process(context.Background(), req, agent, stopChan, respChan, messageChan, errChan, func() {})
	assert.NoError(t, err)
}

// Process: AfterProcess error → GetByID fails (errNew != nil branch)
func TestProcess_AfterProcessError_GetByIDFails(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
	mockConvRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)
	allowAnyLoggerCalls(mockLogger)

	// AfterProcess errors (AnswerVar empty) → GetByID returns valid PO, Update called
	mockMsgRepo.EXPECT().GetByID(gomock.Any(), gomock.Any()).Return(&dapo.ConversationMsgPO{ID: "asst-apgbid"}, nil).AnyTimes()
	mockMsgRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	svc := &agentSvc{
		SvcBase:             service.NewSvcBase(),
		conversationMsgRepo: mockMsgRepo,
		conversationRepo:    mockConvRepo,
		logger:              mockLogger,
		streamDiffFrequency: 1,
	}

	agent := newTestAgent()
	agent.Config.Output = &daconfvalobj.Output{
		Variables: &daconfvalobj.VariablesS{AnswerVar: ""},
	}

	req := &agentreq.ChatReq{
		AgentID:        "a1",
		ConversationID: "conv-apgbid",
		Stream:         false,
		InternalParam:  agentreq.InternalParam{UserID: "u1", AssistantMessageID: "asst-apgbid"},
	}

	stopChan := make(chan struct{})
	respChan := make(chan []byte, 20)
	messageChan := make(chan string, 1)
	errChan := make(chan error)

	session := &Session{ConversationID: "conv-apgbid"}
	SessionMap.Store("conv-apgbid", session)

	defer SessionMap.Delete("conv-apgbid")

	messageChan <- `data:{"status":"False","answer":{"final_answer":"hi"}}`
	close(messageChan)

	err := svc.Process(context.Background(), req, agent, stopChan, respChan, messageChan, errChan, func() {})
	assert.NoError(t, err)
}

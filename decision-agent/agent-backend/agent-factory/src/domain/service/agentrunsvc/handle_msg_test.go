package agentsvc

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/conversationmsgvo"
	agentreq "github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/agent/req"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/cmp/icmp/cmpmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/stretchr/testify/assert"
)

// ---- handleMessageAndTempArea ----

func TestHandleMessageAndTempArea_UpdateMsgError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := cmpmock.NewMockLogger(ctrl)
	mockConvMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)

	svc := &agentSvc{
		SvcBase:             service.NewSvcBase(),
		logger:              mockLogger,
		conversationMsgRepo: mockConvMsgRepo,
	}

	ctx := context.Background()
	req := &agentreq.ChatReq{
		AgentID:    "agent-1",
		AgentRunID: "run-1",
		InternalParam: agentreq.InternalParam{
			UserID:             "user-1",
			AssistantMessageID: "msg-1",
		},
	}
	messageVO := conversationmsgvo.Message{
		Content:     "hello",
		ContentType: "text",
	}

	mockLogger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()
	mockConvMsgRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(errors.New("db error"))

	err := svc.handleMessageAndTempArea(ctx, req, messageVO)

	assert.Error(t, err)
}

func TestHandleMessageAndTempArea_GetConversationError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := cmpmock.NewMockLogger(ctrl)
	mockConvMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
	mockConvRepo := idbaccessmock.NewMockIConversationRepo(ctrl)

	svc := &agentSvc{
		SvcBase:             service.NewSvcBase(),
		logger:              mockLogger,
		conversationMsgRepo: mockConvMsgRepo,
		conversationRepo:    mockConvRepo,
	}

	ctx := context.Background()
	req := &agentreq.ChatReq{
		AgentID:        "agent-1",
		AgentRunID:     "run-1",
		ConversationID: "conv-1",
		InternalParam: agentreq.InternalParam{
			UserID:             "user-1",
			AssistantMessageID: "msg-1",
		},
	}
	messageVO := conversationmsgvo.Message{
		Content:     "hello",
		ContentType: "text",
	}

	mockLogger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()
	mockConvMsgRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
	mockConvRepo.EXPECT().GetByID(gomock.Any(), "conv-1").Return(nil, errors.New("db error"))

	err := svc.handleMessageAndTempArea(ctx, req, messageVO)

	assert.Error(t, err)
}

func TestHandleMessageAndTempArea_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := cmpmock.NewMockLogger(ctrl)
	mockConvMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
	mockConvRepo := idbaccessmock.NewMockIConversationRepo(ctrl)

	svc := &agentSvc{
		SvcBase:             service.NewSvcBase(),
		logger:              mockLogger,
		conversationMsgRepo: mockConvMsgRepo,
		conversationRepo:    mockConvRepo,
	}

	ctx := context.Background()
	req := &agentreq.ChatReq{
		AgentID:        "agent-1",
		AgentRunID:     "run-1",
		ConversationID: "conv-1",
		InternalParam: agentreq.InternalParam{
			UserID:             "user-1",
			AssistantMessageID: "msg-1",
		},
	}
	messageVO := conversationmsgvo.Message{
		Content:     "hello",
		ContentType: "text",
	}

	convPO := &dapo.ConversationPO{ID: "conv-1"}

	mockLogger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()
	mockConvMsgRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
	mockConvRepo.EXPECT().GetByID(gomock.Any(), "conv-1").Return(convPO, nil)
	mockConvRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

	err := svc.handleMessageAndTempArea(ctx, req, messageVO)

	assert.NoError(t, err)
}

// ---- GetHistoryAndMsgIndex ----

func TestGetHistoryAndMsgIndex_EmptyConversationID_CreateError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := cmpmock.NewMockLogger(ctrl)
	mockConvRepo := idbaccessmock.NewMockIConversationRepo(ctrl)

	svc := &agentSvc{
		SvcBase:          service.NewSvcBase(),
		logger:           mockLogger,
		conversationRepo: mockConvRepo,
	}

	ctx := context.Background()
	req := &agentreq.ChatReq{
		AgentID:        "agent-1",
		ConversationID: "",
		Query:          "hello",
		InternalParam:  agentreq.InternalParam{UserID: "user-1"},
	}

	mockConvRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, errors.New("create error"))

	_, _, _, err := svc.GetHistoryAndMsgIndex(ctx, req, 0, nil)

	assert.Error(t, err)
}

func TestGetHistoryAndMsgIndex_EmptyConversationID_CreateSuccess(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := cmpmock.NewMockLogger(ctrl)
	mockConvRepo := idbaccessmock.NewMockIConversationRepo(ctrl)

	svc := &agentSvc{
		SvcBase:          service.NewSvcBase(),
		logger:           mockLogger,
		conversationRepo: mockConvRepo,
	}

	ctx := context.Background()
	req := &agentreq.ChatReq{
		AgentID:        "agent-1",
		ConversationID: "",
		Query:          "hello query",
		InternalParam:  agentreq.InternalParam{UserID: "user-1"},
	}

	createdConv := &dapo.ConversationPO{ID: "new-conv-id"}
	mockConvRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(createdConv, nil)

	convPO, contexts, msgIndex, err := svc.GetHistoryAndMsgIndex(ctx, req, 0, nil)

	assert.NoError(t, err)
	assert.NotNil(t, convPO)
	assert.Equal(t, "new-conv-id", convPO.ID)
	assert.Equal(t, "new-conv-id", req.ConversationID)
	assert.Nil(t, contexts)
	assert.Equal(t, 0, msgIndex)
}

func TestGetHistoryAndMsgIndex_ExistingConversation_GetByIDError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := cmpmock.NewMockLogger(ctrl)
	mockConvRepo := idbaccessmock.NewMockIConversationRepo(ctrl)

	svc := &agentSvc{
		SvcBase:          service.NewSvcBase(),
		logger:           mockLogger,
		conversationRepo: mockConvRepo,
	}

	ctx := context.Background()
	req := &agentreq.ChatReq{
		AgentID:        "agent-1",
		ConversationID: "existing-conv",
		InternalParam:  agentreq.InternalParam{UserID: "user-1"},
	}

	mockConvRepo.EXPECT().GetByID(gomock.Any(), "existing-conv").Return(nil, errors.New("db error"))

	_, _, _, err := svc.GetHistoryAndMsgIndex(ctx, req, 0, nil)

	assert.Error(t, err)
}

// ---- UpsertUserAndAssistantMsg: normal chat create user msg ----

func TestUpsertUserAndAssistantMsg_NormalChat_CreateUserMsgError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := cmpmock.NewMockLogger(ctrl)
	mockConvMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
	mockConvRepo := idbaccessmock.NewMockIConversationRepo(ctrl)

	svc := &agentSvc{
		SvcBase:             service.NewSvcBase(),
		logger:              mockLogger,
		conversationMsgRepo: mockConvMsgRepo,
		conversationRepo:    mockConvRepo,
	}

	ctx := context.Background()
	req := &agentreq.ChatReq{
		AgentID:        "agent-1",
		ConversationID: "conv-1",
		Query:          "hello",
		InternalParam:  agentreq.InternalParam{UserID: "user-1"},
	}
	convPO := &dapo.ConversationPO{ID: "conv-1"}

	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()
	mockConvMsgRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return("", errors.New("create user msg error"))

	_, _, _, err := svc.UpsertUserAndAssistantMsg(ctx, req, 0, convPO)

	assert.Error(t, err)
}

func TestUpsertUserAndAssistantMsg_NormalChat_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := cmpmock.NewMockLogger(ctrl)
	mockConvMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
	mockConvRepo := idbaccessmock.NewMockIConversationRepo(ctrl)

	svc := &agentSvc{
		SvcBase:             service.NewSvcBase(),
		logger:              mockLogger,
		conversationMsgRepo: mockConvMsgRepo,
		conversationRepo:    mockConvRepo,
	}

	ctx := context.Background()
	req := &agentreq.ChatReq{
		AgentID:        "agent-1",
		ConversationID: "conv-1",
		Query:          "hello",
		InternalParam:  agentreq.InternalParam{UserID: "user-1"},
	}
	convPO := &dapo.ConversationPO{ID: "conv-1"}

	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()
	mockConvMsgRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return("user-msg-id", nil).Times(1)
	mockConvRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
	mockConvMsgRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return("assistant-msg-id", nil).Times(1)

	userMsgID, assistantMsgID, _, err := svc.UpsertUserAndAssistantMsg(ctx, req, 0, convPO)

	assert.NoError(t, err)
	assert.Equal(t, "user-msg-id", userMsgID)
	assert.Equal(t, "assistant-msg-id", assistantMsgID)
}

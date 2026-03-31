package agentsvc

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/agentresperr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/comvalobj"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/conversationmsgvo"
	agentreq "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent/req"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent/req/chatopt"
	agentresp "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent/resp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp/cmpmock"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driver/iportdriver/iportdrivermock"
)

func TestAgentSvc_MsgResp2MsgPO_BasicSuccess(t *testing.T) {
	t.Parallel()

	svc := &agentSvc{SvcBase: service.NewSvcBase()}
	ctx := context.Background()
	msgResp := agentresp.ChatResp{
		ConversationID: "conv-1",
		Message:        conversationmsgvo.Message{Content: "hello", ContentType: "text"},
	}
	req := &agentreq.ChatReq{
		AgentID: "agent-1", AgentRunID: "run-1",
		InternalParam: agentreq.InternalParam{UserID: "user-1"},
	}
	po, _, err := svc.MsgResp2MsgPO(ctx, msgResp, req)
	assert.NoError(t, err)
	assert.Equal(t, "conv-1", po.ConversationID)
}

func TestAgentSvc_MsgResp2MsgPO_ContentMarshalError(t *testing.T) {
	t.Parallel()

	svc := &agentSvc{SvcBase: service.NewSvcBase()}
	ctx := context.Background()
	msgResp := agentresp.ChatResp{
		ConversationID: "conv-err-content",
		Message: conversationmsgvo.Message{
			Content: map[string]interface{}{"bad": func() {}},
		},
	}
	req := &agentreq.ChatReq{InternalParam: agentreq.InternalParam{UserID: "user-1"}}

	_, _, err := svc.MsgResp2MsgPO(ctx, msgResp, req)
	assert.Error(t, err)
}

func TestAgentSvc_MsgResp2MsgPO_ExtMarshalError(t *testing.T) {
	t.Parallel()

	svc := &agentSvc{SvcBase: service.NewSvcBase()}
	ctx := context.Background()
	msgResp := agentresp.ChatResp{
		ConversationID: "conv-err-ext",
		Message: conversationmsgvo.Message{
			Content: "ok",
			Ext: &conversationmsgvo.MessageExt{
				Error: &agentresperr.RespError{Type: agentresperr.RespErrorTypeAgentFactory, Error: func() {}},
			},
		},
	}
	req := &agentreq.ChatReq{InternalParam: agentreq.InternalParam{UserID: "user-1"}}

	_, _, err := svc.MsgResp2MsgPO(ctx, msgResp, req)
	assert.Error(t, err)
}

func TestAgentSvc_GetHistoryAndMsgIndex_NewConversation_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockConvRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
	svc := &agentSvc{SvcBase: service.NewSvcBase(), conversationRepo: mockConvRepo}
	mockConvRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&dapo.ConversationPO{ID: "new-conv"}, nil)

	ctx := context.Background()
	req := &agentreq.ChatReq{AgentID: "a1", ConversationID: "", Query: "hello"}
	convPO, _, _, err := svc.GetHistoryAndMsgIndex(ctx, req, 0, nil)
	assert.NoError(t, err)
	assert.Equal(t, "new-conv", convPO.ID)
}

func TestAgentSvc_GetHistoryAndMsgIndex_NewConversation_LongQuery(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockConvRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
	svc := &agentSvc{SvcBase: service.NewSvcBase(), conversationRepo: mockConvRepo}
	mockConvRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&dapo.ConversationPO{ID: "c1"}, nil)

	ctx := context.Background()
	req := &agentreq.ChatReq{ConversationID: "", Query: "这是一个超过五十个字符的非常长的查询字符串用于测试标题截取逻辑是否正确处理了多字节Unicode字符集的情况测试测试测试"}
	_, _, _, err := svc.GetHistoryAndMsgIndex(ctx, req, 0, nil)
	assert.NoError(t, err)
}

func TestAgentSvc_GetHistoryAndMsgIndex_NewConversation_CreateError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockConvRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
	svc := &agentSvc{SvcBase: service.NewSvcBase(), conversationRepo: mockConvRepo}
	mockConvRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, errors.New("db error"))

	ctx := context.Background()
	req := &agentreq.ChatReq{ConversationID: ""}
	_, _, _, err := svc.GetHistoryAndMsgIndex(ctx, req, 0, nil)
	assert.Error(t, err)
}

func TestAgentSvc_GetHistoryAndMsgIndex_ExistingConversation_GetError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockConvRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
	svc := &agentSvc{SvcBase: service.NewSvcBase(), conversationRepo: mockConvRepo}
	mockConvRepo.EXPECT().GetByID(gomock.Any(), "conv-1").Return(nil, errors.New("not found"))

	ctx := context.Background()
	req := &agentreq.ChatReq{ConversationID: "conv-1"}
	_, _, _, err := svc.GetHistoryAndMsgIndex(ctx, req, 0, nil)
	assert.Error(t, err)
}

func TestAgentSvc_GetHistoryAndMsgIndex_ExistingConversation_GetMaxIndexError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockConvRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
	mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
	svc := &agentSvc{SvcBase: service.NewSvcBase(), conversationRepo: mockConvRepo, conversationMsgRepo: mockMsgRepo}
	mockConvRepo.EXPECT().GetByID(gomock.Any(), "conv-1").Return(&dapo.ConversationPO{ID: "conv-1"}, nil)
	mockMsgRepo.EXPECT().GetMaxIndexByID(gomock.Any(), "conv-1").Return(0, errors.New("db error"))

	ctx := context.Background()
	req := &agentreq.ChatReq{ConversationID: "conv-1"}
	_, _, _, err := svc.GetHistoryAndMsgIndex(ctx, req, 0, nil)
	assert.Error(t, err)
}

func TestAgentSvc_GetHistoryAndMsgIndex_ExistingConversation_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockConvRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
	mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
	svc := &agentSvc{SvcBase: service.NewSvcBase(), conversationRepo: mockConvRepo, conversationMsgRepo: mockMsgRepo}
	mockConvRepo.EXPECT().GetByID(gomock.Any(), "conv-1").Return(&dapo.ConversationPO{ID: "conv-1"}, nil)
	mockMsgRepo.EXPECT().GetMaxIndexByID(gomock.Any(), "conv-1").Return(5, nil)

	ctx := context.Background()
	req := &agentreq.ChatReq{ConversationID: "conv-1"}
	convPO, _, idx, err := svc.GetHistoryAndMsgIndex(ctx, req, 0, nil)
	assert.NoError(t, err)
	assert.Equal(t, 5, idx)
	assert.Equal(t, "conv-1", convPO.ID)
}

func TestAgentSvc_GetHistoryAndMsgIndex_NeedHistory_GetHistoryError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConvRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
	mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
	mockConvSvc := iportdrivermock.NewMockIConversationSvc(ctrl)
	svc := &agentSvc{SvcBase: service.NewSvcBase(), conversationRepo: mockConvRepo, conversationMsgRepo: mockMsgRepo, conversationSvc: mockConvSvc}

	mockConvRepo.EXPECT().GetByID(gomock.Any(), "conv-need-history").Return(&dapo.ConversationPO{ID: "conv-need-history"}, nil)
	mockMsgRepo.EXPECT().GetMaxIndexByID(gomock.Any(), "conv-need-history").Return(2, nil)
	mockConvSvc.EXPECT().GetHistory(gomock.Any(), "conv-need-history", 5, "regen-user", "regen-asst").Return(nil, errors.New("history error"))

	req := &agentreq.ChatReq{
		ConversationID:           "conv-need-history",
		RegenerateUserMsgID:      "regen-user",
		RegenerateAssistantMsgID: "regen-asst",
		ChatOption:               chatopt.ChatOption{IsNeedHistory: true},
	}

	_, _, _, err := svc.GetHistoryAndMsgIndex(context.Background(), req, 5, nil)
	assert.Error(t, err)
}

func TestAgentSvc_GetHistoryAndMsgIndex_NeedHistory_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConvRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
	mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
	mockConvSvc := iportdrivermock.NewMockIConversationSvc(ctrl)
	svc := &agentSvc{SvcBase: service.NewSvcBase(), conversationRepo: mockConvRepo, conversationMsgRepo: mockMsgRepo, conversationSvc: mockConvSvc}

	mockConvRepo.EXPECT().GetByID(gomock.Any(), "conv-need-history-ok").Return(&dapo.ConversationPO{ID: "conv-need-history-ok"}, nil)
	mockMsgRepo.EXPECT().GetMaxIndexByID(gomock.Any(), "conv-need-history-ok").Return(7, nil)

	expectedHistory := []*comvalobj.LLMMessage{{Role: "user", Content: "hi"}}
	mockConvSvc.EXPECT().GetHistory(gomock.Any(), "conv-need-history-ok", 3, "", "").Return(expectedHistory, nil)

	req := &agentreq.ChatReq{
		ConversationID: "conv-need-history-ok",
		ChatOption:     chatopt.ChatOption{IsNeedHistory: true},
	}

	_, history, idx, err := svc.GetHistoryAndMsgIndex(context.Background(), req, 3, nil)
	assert.NoError(t, err)
	assert.Equal(t, 7, idx)
	assert.Equal(t, expectedHistory, history)
}

func TestAgentSvc_UpsertUserAndAssistantMsg_NormalChat_UserMsgCreateError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockConvRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
	mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
	svc := &agentSvc{SvcBase: service.NewSvcBase(), conversationRepo: mockConvRepo, conversationMsgRepo: mockMsgRepo}
	mockMsgRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return("", errors.New("create error"))

	ctx := context.Background()
	req := &agentreq.ChatReq{AgentID: "a1", ConversationID: "conv-1", Query: "hello", InternalParam: agentreq.InternalParam{UserID: "u1"}}
	_, _, _, err := svc.UpsertUserAndAssistantMsg(ctx, req, 0, &dapo.ConversationPO{ID: "conv-1"})
	assert.Error(t, err)
}

func TestAgentSvc_UpsertUserAndAssistantMsg_NormalChat_UpdateConvError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockConvRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
	mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
	svc := &agentSvc{SvcBase: service.NewSvcBase(), conversationRepo: mockConvRepo, conversationMsgRepo: mockMsgRepo}
	mockMsgRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return("user-msg-id", nil)
	mockConvRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(errors.New("update error"))

	ctx := context.Background()
	req := &agentreq.ChatReq{ConversationID: "conv-1", Query: "hello", InternalParam: agentreq.InternalParam{UserID: "u1"}}
	_, _, _, err := svc.UpsertUserAndAssistantMsg(ctx, req, 0, &dapo.ConversationPO{ID: "conv-1"})
	assert.Error(t, err)
}

func TestAgentSvc_UpsertUserAndAssistantMsg_NormalChat_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockConvRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
	mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
	svc := &agentSvc{SvcBase: service.NewSvcBase(), conversationRepo: mockConvRepo, conversationMsgRepo: mockMsgRepo}
	gomock.InOrder(
		mockMsgRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return("user-msg-id", nil),
		mockConvRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil),
		mockMsgRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return("asst-msg-id", nil),
	)

	ctx := context.Background()
	req := &agentreq.ChatReq{ConversationID: "conv-1", Query: "hello", InternalParam: agentreq.InternalParam{UserID: "u1"}}
	userMsgID, asstMsgID, _, err := svc.UpsertUserAndAssistantMsg(ctx, req, 0, &dapo.ConversationPO{ID: "conv-1"})
	assert.NoError(t, err)
	assert.Equal(t, "user-msg-id", userMsgID)
	assert.Equal(t, "asst-msg-id", asstMsgID)
}

func TestAgentSvc_UpsertUserAndAssistantMsg_RegenerateUserMsg_GetError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
	svc := &agentSvc{SvcBase: service.NewSvcBase(), conversationMsgRepo: mockMsgRepo}
	mockMsgRepo.EXPECT().GetByID(gomock.Any(), "user-msg-123").Return(nil, errors.New("not found"))

	ctx := context.Background()
	req := &agentreq.ChatReq{ConversationID: "conv-1", RegenerateUserMsgID: "user-msg-123", InternalParam: agentreq.InternalParam{UserID: "u1"}}
	_, _, _, err := svc.UpsertUserAndAssistantMsg(ctx, req, 0, &dapo.ConversationPO{ID: "conv-1"})
	assert.Error(t, err)
}

func TestAgentSvc_UpsertUserAndAssistantMsg_InterruptedMsg_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
	svc := &agentSvc{SvcBase: service.NewSvcBase(), conversationMsgRepo: mockMsgRepo}
	interruptedMsg := &dapo.ConversationMsgPO{ID: "interrupted-asst-msg", ReplyID: "user-msg-1", Index: 3}
	gomock.InOrder(
		mockMsgRepo.EXPECT().GetByID(gomock.Any(), "interrupted-asst-msg").Return(interruptedMsg, nil),
		mockMsgRepo.EXPECT().GetByID(gomock.Any(), "interrupted-asst-msg").Return(interruptedMsg, nil),
		mockMsgRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil),
	)

	ctx := context.Background()
	req := &agentreq.ChatReq{ConversationID: "conv-1", InterruptedAssistantMsgID: "interrupted-asst-msg", InternalParam: agentreq.InternalParam{UserID: "u1"}}
	userID, asstID, idx, err := svc.UpsertUserAndAssistantMsg(ctx, req, 0, &dapo.ConversationPO{ID: "conv-1"})
	assert.NoError(t, err)
	assert.Equal(t, "user-msg-1", userID)
	assert.Equal(t, "interrupted-asst-msg", asstID)
	assert.Equal(t, 3, idx)
}

func TestAgentSvc_HandleStopChan_MsgRepoUpdateError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
	svc := &agentSvc{SvcBase: service.NewSvcBase(), conversationMsgRepo: mockMsgRepo, logger: mockLogger}
	mockMsgRepo.EXPECT().GetByID(gomock.Any(), gomock.Any()).Return(&dapo.ConversationMsgPO{ID: "asst-1"}, nil)
	mockMsgRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(errors.New("update error"))

	session := &Session{ConversationID: "conv-1", TempMsgResp: agentresp.ChatResp{ConversationID: "conv-1"}}
	ctx := context.Background()
	req := &agentreq.ChatReq{AgentID: "a1", ConversationID: "conv-1", AgentRunID: "run-1", InternalParam: agentreq.InternalParam{UserID: "u1", AssistantMessageID: "asst-1"}}
	err := svc.HandleStopChan(ctx, req, session)
	assert.Error(t, err)
}

func TestAgentSvc_HandleStopChan_GetConversationError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockConvRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
	mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
	svc := &agentSvc{SvcBase: service.NewSvcBase(), conversationRepo: mockConvRepo, conversationMsgRepo: mockMsgRepo, logger: mockLogger}
	mockMsgRepo.EXPECT().GetByID(gomock.Any(), gomock.Any()).Return(&dapo.ConversationMsgPO{ID: "asst-1"}, nil)
	mockMsgRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
	mockConvRepo.EXPECT().GetByID(gomock.Any(), "conv-1").Return(nil, errors.New("conv not found"))

	session := &Session{ConversationID: "conv-1", TempMsgResp: agentresp.ChatResp{ConversationID: "conv-1"}}
	ctx := context.Background()
	req := &agentreq.ChatReq{AgentID: "a1", ConversationID: "conv-1", AgentRunID: "run-1", InternalParam: agentreq.InternalParam{UserID: "u1", AssistantMessageID: "asst-1"}}
	err := svc.HandleStopChan(ctx, req, session)
	assert.Error(t, err)
}

func TestAgentSvc_HandleStopChan_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockConvRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
	mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)
	mockLogger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()
	svc := &agentSvc{SvcBase: service.NewSvcBase(), conversationRepo: mockConvRepo, conversationMsgRepo: mockMsgRepo, logger: mockLogger}
	mockMsgRepo.EXPECT().GetByID(gomock.Any(), gomock.Any()).Return(&dapo.ConversationMsgPO{ID: "asst-1"}, nil)
	mockMsgRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
	mockConvRepo.EXPECT().GetByID(gomock.Any(), "conv-1").Return(&dapo.ConversationPO{ID: "conv-1"}, nil)
	mockConvRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

	session := &Session{ConversationID: "conv-1", TempMsgResp: agentresp.ChatResp{ConversationID: "conv-1"}}
	ctx := context.Background()
	req := &agentreq.ChatReq{AgentID: "a1", ConversationID: "conv-1", AgentRunID: "run-1", InternalParam: agentreq.InternalParam{UserID: "u1", AssistantMessageID: "asst-1"}}
	err := svc.HandleStopChan(ctx, req, session)
	assert.NoError(t, err)
}

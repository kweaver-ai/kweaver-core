package agentsvc

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/entity/conversationeo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service"
	agentreq "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent/req"
	conversationresp "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/conversation/conversationresp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp/cmpmock"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driver/iportdriver/iportdrivermock"
)

// UpsertUserAndAssistantMsg: RegenerateUserMsgID set → phase1 GetByID+Update ok, phase2 Detail fails
func TestUpsertMsg_RegenerateUserMsgID_DetailFails(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
	mockConvRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
	mockConvSvc := iportdrivermock.NewMockIConversationSvc(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)
	allowAnyLoggerCalls(mockLogger)

	// Phase 1: GetByID + Update user msg succeeds
	mockMsgRepo.EXPECT().GetByID(gomock.Any(), "user-regen-1").Return(&dapo.ConversationMsgPO{ID: "user-regen-1"}, nil)
	mockMsgRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
	// Phase 2: conversationSvc.Detail returns error
	mockConvSvc.EXPECT().Detail(gomock.Any(), "conv-regen-1").Return(conversationresp.ConversationDetail{}, errors.New("detail error"))

	svc := &agentSvc{
		SvcBase:             service.NewSvcBase(),
		conversationMsgRepo: mockMsgRepo,
		conversationRepo:    mockConvRepo,
		conversationSvc:     mockConvSvc,
		logger:              mockLogger,
	}

	req := &agentreq.ChatReq{
		AgentID:             "a1",
		ConversationID:      "conv-regen-1",
		RegenerateUserMsgID: "user-regen-1",
		InternalParam:       agentreq.InternalParam{UserID: "u1"},
	}
	convPO := &dapo.ConversationPO{ID: "conv-regen-1"}

	_, _, _, err := svc.UpsertUserAndAssistantMsg(context.Background(), req, 0, convPO)
	assert.Error(t, err)
}

// UpsertUserAndAssistantMsg: RegenerateUserMsgID set → GetByID fails
func TestUpsertMsg_RegenerateUserMsgID_GetByIDFails(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
	mockConvRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)
	allowAnyLoggerCalls(mockLogger)

	mockMsgRepo.EXPECT().GetByID(gomock.Any(), "user-regen-2").Return(nil, errors.New("not found"))

	svc := &agentSvc{
		SvcBase:             service.NewSvcBase(),
		conversationMsgRepo: mockMsgRepo,
		conversationRepo:    mockConvRepo,
		logger:              mockLogger,
	}

	req := &agentreq.ChatReq{
		AgentID:             "a1",
		ConversationID:      "conv-regen-2",
		RegenerateUserMsgID: "user-regen-2",
		InternalParam:       agentreq.InternalParam{UserID: "u1"},
	}
	convPO := &dapo.ConversationPO{ID: "conv-regen-2"}

	_, _, _, err := svc.UpsertUserAndAssistantMsg(context.Background(), req, 0, convPO)
	assert.Error(t, err)
}

// UpsertUserAndAssistantMsg: RegenerateUserMsgID set → Update user msg fails
func TestUpsertMsg_RegenerateUserMsgID_UpdateFails(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
	mockConvRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)
	allowAnyLoggerCalls(mockLogger)

	mockMsgRepo.EXPECT().GetByID(gomock.Any(), "user-regen-3").Return(&dapo.ConversationMsgPO{ID: "user-regen-3"}, nil)
	mockMsgRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(errors.New("update failed"))

	svc := &agentSvc{
		SvcBase:             service.NewSvcBase(),
		conversationMsgRepo: mockMsgRepo,
		conversationRepo:    mockConvRepo,
		logger:              mockLogger,
	}

	req := &agentreq.ChatReq{
		AgentID:             "a1",
		ConversationID:      "conv-regen-3",
		RegenerateUserMsgID: "user-regen-3",
		InternalParam:       agentreq.InternalParam{UserID: "u1"},
	}
	convPO := &dapo.ConversationPO{ID: "conv-regen-3"}

	_, _, _, err := svc.UpsertUserAndAssistantMsg(context.Background(), req, 0, convPO)
	assert.Error(t, err)
}

// UpsertUserAndAssistantMsg: RegenerateAssistantMsgID set → GetByID(×2) + Update
func TestUpsertMsg_RegenerateAssistantMsgID_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
	mockConvRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)
	allowAnyLoggerCalls(mockLogger)

	// Phase 1 GetByID (line 246): get assistant msg to find userMsgID
	// Phase 2 GetByID (line 293): get assistant msg again to update status
	mockMsgRepo.EXPECT().GetByID(gomock.Any(), "asst-existing-1").Return(&dapo.ConversationMsgPO{
		ID: "asst-existing-1", ReplyID: "user-existing-1",
	}, nil).Times(2)
	// Phase 2 Update: set status to processing
	mockMsgRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

	svc := &agentSvc{
		SvcBase:             service.NewSvcBase(),
		conversationMsgRepo: mockMsgRepo,
		conversationRepo:    mockConvRepo,
		logger:              mockLogger,
	}

	req := &agentreq.ChatReq{
		AgentID:                  "a1",
		ConversationID:           "conv-regen-asst-1",
		RegenerateAssistantMsgID: "asst-existing-1",
		InternalParam:            agentreq.InternalParam{UserID: "u1"},
	}
	convPO := &dapo.ConversationPO{ID: "conv-regen-asst-1"}

	userID, asstID, _, err := svc.UpsertUserAndAssistantMsg(context.Background(), req, 0, convPO)
	assert.NoError(t, err)
	assert.Equal(t, "user-existing-1", userID)
	assert.Equal(t, "asst-existing-1", asstID)
}

// UpsertUserAndAssistantMsg: InterruptedAssistantMsgID set → GetByID(×2) + Update
func TestUpsertMsg_InterruptedAssistantMsgID_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
	mockConvRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)
	allowAnyLoggerCalls(mockLogger)

	// Phase 1 GetByID (line 253): get assistant msg to find userMsgID
	// Phase 2 GetByID (line 313): get assistant msg again to update status
	mockMsgRepo.EXPECT().GetByID(gomock.Any(), "asst-interrupted-1").Return(&dapo.ConversationMsgPO{
		ID: "asst-interrupted-1", ReplyID: "user-interrupted-1",
	}, nil).Times(2)
	// Phase 2 Update: set status to processing
	mockMsgRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

	svc := &agentSvc{
		SvcBase:             service.NewSvcBase(),
		conversationMsgRepo: mockMsgRepo,
		conversationRepo:    mockConvRepo,
		logger:              mockLogger,
	}

	req := &agentreq.ChatReq{
		AgentID:                   "a1",
		ConversationID:            "conv-interrupted-1",
		InterruptedAssistantMsgID: "asst-interrupted-1",
		InternalParam:             agentreq.InternalParam{UserID: "u1"},
	}
	convPO := &dapo.ConversationPO{ID: "conv-interrupted-1"}

	userID, asstID, _, err := svc.UpsertUserAndAssistantMsg(context.Background(), req, 0, convPO)
	assert.NoError(t, err)
	assert.Equal(t, "user-interrupted-1", userID)
	assert.Equal(t, "asst-interrupted-1", asstID)
}

func TestUpsertMsg_RegenerateAssistantMsgID_SecondGetByIDFails(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
	mockConvRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)
	allowAnyLoggerCalls(mockLogger)

	mockMsgRepo.EXPECT().GetByID(gomock.Any(), "asst-existing-err").Return(&dapo.ConversationMsgPO{ID: "asst-existing-err", ReplyID: "user-1"}, nil)
	mockMsgRepo.EXPECT().GetByID(gomock.Any(), "asst-existing-err").Return(nil, errors.New("second get failed"))

	svc := &agentSvc{
		SvcBase:             service.NewSvcBase(),
		conversationMsgRepo: mockMsgRepo,
		conversationRepo:    mockConvRepo,
		logger:              mockLogger,
	}

	req := &agentreq.ChatReq{
		AgentID:                  "a1",
		ConversationID:           "conv-regen-asst-err",
		RegenerateAssistantMsgID: "asst-existing-err",
		InternalParam:            agentreq.InternalParam{UserID: "u1"},
	}
	_, _, _, err := svc.UpsertUserAndAssistantMsg(context.Background(), req, 0, &dapo.ConversationPO{ID: "conv-regen-asst-err"})
	assert.Error(t, err)
}

func TestUpsertMsg_InterruptedAssistantMsgID_UpdateFails(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
	mockConvRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)
	allowAnyLoggerCalls(mockLogger)

	mockMsgRepo.EXPECT().GetByID(gomock.Any(), "asst-interrupted-err").Return(&dapo.ConversationMsgPO{ID: "asst-interrupted-err", ReplyID: "user-1", Index: 2}, nil).Times(2)
	mockMsgRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(errors.New("update failed"))

	svc := &agentSvc{
		SvcBase:             service.NewSvcBase(),
		conversationMsgRepo: mockMsgRepo,
		conversationRepo:    mockConvRepo,
		logger:              mockLogger,
	}

	req := &agentreq.ChatReq{
		AgentID:                   "a1",
		ConversationID:            "conv-interrupted-err",
		InterruptedAssistantMsgID: "asst-interrupted-err",
		InternalParam:             agentreq.InternalParam{UserID: "u1"},
	}

	_, _, _, err := svc.UpsertUserAndAssistantMsg(context.Background(), req, 0, &dapo.ConversationPO{ID: "conv-interrupted-err"})
	assert.Error(t, err)
}

// UpsertUserAndAssistantMsg: InterruptedAssistantMsgID set → InterruptInfo in Ext should be cleared
func TestUpsertMsg_InterruptedAssistantMsgID_ClearsInterruptInfo(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
	mockConvRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)
	allowAnyLoggerCalls(mockLogger)

	// 构造带有 InterruptInfo 的 Ext JSON
	extWithInterrupt := `{"interrupt_info":{"handle":{"resume_data":"test"},"data":{"tool_name":"test_tool"}},"total_time":1.5,"agent_run_id":"run-1"}`

	mockMsgRepo.EXPECT().GetByID(gomock.Any(), "asst-interrupt-clear").Return(&dapo.ConversationMsgPO{
		ID: "asst-interrupt-clear", ReplyID: "user-interrupt-clear", Index: 2, Ext: &extWithInterrupt,
	}, nil).Times(2)

	// 捕获 Update 的参数，验证 InterruptInfo 已被清除
	var capturedPO *dapo.ConversationMsgPO

	mockMsgRepo.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, po *dapo.ConversationMsgPO) error {
		capturedPO = po
		return nil
	})

	svc := &agentSvc{
		SvcBase:             service.NewSvcBase(),
		conversationMsgRepo: mockMsgRepo,
		conversationRepo:    mockConvRepo,
		logger:              mockLogger,
	}

	req := &agentreq.ChatReq{
		AgentID:                   "a1",
		ConversationID:            "conv-interrupt-clear",
		InterruptedAssistantMsgID: "asst-interrupt-clear",
		InternalParam:             agentreq.InternalParam{UserID: "u1"},
	}

	userID, asstID, _, err := svc.UpsertUserAndAssistantMsg(context.Background(), req, 0, &dapo.ConversationPO{ID: "conv-interrupt-clear"})
	assert.NoError(t, err)
	assert.Equal(t, "user-interrupt-clear", userID)
	assert.Equal(t, "asst-interrupt-clear", asstID)

	// 验证 Ext 已更新且 InterruptInfo 被清除
	assert.NotNil(t, capturedPO)
	assert.NotNil(t, capturedPO.Ext)
	assert.NotContains(t, *capturedPO.Ext, "interrupt_info")
	// 验证其他字段保留
	assert.Contains(t, *capturedPO.Ext, "total_time")
	assert.Contains(t, *capturedPO.Ext, "agent_run_id")
}

func TestUpsertMsg_RegenerateUserMsgID_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
	mockConvRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
	mockConvSvc := iportdrivermock.NewMockIConversationSvc(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)
	allowAnyLoggerCalls(mockLogger)

	mockMsgRepo.EXPECT().GetByID(gomock.Any(), "user-regen-ok").Return(&dapo.ConversationMsgPO{ID: "user-regen-ok"}, nil)
	mockMsgRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
	mockConvSvc.EXPECT().Detail(gomock.Any(), "conv-regen-ok").Return(conversationresp.ConversationDetail{
		Conversation: conversationeo.Conversation{Messages: []*dapo.ConversationMsgPO{
			{ID: "user-regen-ok", Index: 1},
			{ID: "asst-after-user", Index: 2},
		}},
	}, nil)
	mockMsgRepo.EXPECT().GetByID(gomock.Any(), "asst-after-user").Return(&dapo.ConversationMsgPO{ID: "asst-after-user", Index: 2}, nil)
	mockMsgRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

	svc := &agentSvc{
		SvcBase:             service.NewSvcBase(),
		conversationMsgRepo: mockMsgRepo,
		conversationRepo:    mockConvRepo,
		conversationSvc:     mockConvSvc,
		logger:              mockLogger,
	}

	req := &agentreq.ChatReq{
		AgentID:             "a1",
		ConversationID:      "conv-regen-ok",
		RegenerateUserMsgID: "user-regen-ok",
		Query:               "new question",
		InternalParam:       agentreq.InternalParam{UserID: "u1"},
	}

	userID, asstID, idx, err := svc.UpsertUserAndAssistantMsg(context.Background(), req, 0, &dapo.ConversationPO{ID: "conv-regen-ok"})
	assert.NoError(t, err)
	assert.Equal(t, "user-regen-ok", userID)
	assert.Equal(t, "asst-after-user", asstID)
	assert.Equal(t, 2, idx)
}

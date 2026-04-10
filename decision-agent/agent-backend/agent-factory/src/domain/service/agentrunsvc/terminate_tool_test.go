package agentsvc

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/agentrespvo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/conversationmsgvo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/cmp/icmp/cmpmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/stretchr/testify/assert"
)

// ---- TerminateChat ----

func TestTerminateChat_NoStopChan_NoInterruptedMsg(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := cmpmock.NewMockLogger(ctrl)
	svc := &agentSvc{
		SvcBase: service.NewSvcBase(),
		logger:  mockLogger,
	}

	// Ensure no stopchan in map for this ID
	stopChanMap.Delete("conv-no-chan")

	ctx := context.Background()
	err := svc.TerminateChat(ctx, "conv-no-chan", "", "")

	assert.Error(t, err)
}

func TestTerminateChat_NoStopChan_WithInterruptedMsg_GetByIDError(t *testing.T) {
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

	// No stopchan in map, interruptedMsgID != "" → silent continue, then step 3 fails
	stopChanMap.Delete("conv-nostop-msg")

	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
	mockConvMsgRepo.EXPECT().GetByID(gomock.Any(), "msg-x").Return(nil, errors.New("db error"))

	ctx := context.Background()
	err := svc.TerminateChat(ctx, "conv-nostop-msg", "", "msg-x")

	assert.Error(t, err)
}

func TestTerminateChat_StopChanFound_NoInterruptedMsg(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := cmpmock.NewMockLogger(ctrl)
	svc := &agentSvc{
		SvcBase: service.NewSvcBase(),
		logger:  mockLogger,
	}

	ch := make(chan struct{}, 1)
	stopChanMap.Store("conv-ok", ch)

	mockLogger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()

	ctx := context.Background()
	err := svc.TerminateChat(ctx, "conv-ok", "", "")

	assert.NoError(t, err)
	// stopchan should be deleted from map
	_, ok := stopChanMap.Load("conv-ok")
	assert.False(t, ok)
}

func TestTerminateChat_WithInterruptedMsgID_GetMsgError(t *testing.T) {
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

	ch := make(chan struct{}, 1)
	stopChanMap.Store("conv-with-msg", ch)

	mockLogger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
	mockConvMsgRepo.EXPECT().GetByID(gomock.Any(), "msg-1").Return(nil, errors.New("db error"))

	ctx := context.Background()
	err := svc.TerminateChat(ctx, "conv-with-msg", "", "msg-1")

	assert.Error(t, err)
}

func TestTerminateChat_WithInterruptedMsgID_UpdateError(t *testing.T) {
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

	ch := make(chan struct{}, 1)
	stopChanMap.Store("conv-update-err", ch)

	msgPO := &dapo.ConversationMsgPO{ID: "msg-2"}

	mockLogger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
	mockConvMsgRepo.EXPECT().GetByID(gomock.Any(), "msg-2").Return(msgPO, nil)
	mockConvMsgRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(errors.New("update error"))

	ctx := context.Background()
	err := svc.TerminateChat(ctx, "conv-update-err", "", "msg-2")

	assert.Error(t, err)
}

func TestTerminateChat_WithInterruptedMsgID_Success(t *testing.T) {
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

	ch := make(chan struct{}, 1)
	stopChanMap.Store("conv-success", ch)

	msgPO := &dapo.ConversationMsgPO{ID: "msg-3"}

	mockLogger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
	mockConvMsgRepo.EXPECT().GetByID(gomock.Any(), "msg-3").Return(msgPO, nil)
	mockConvMsgRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

	ctx := context.Background()
	err := svc.TerminateChat(ctx, "conv-success", "", "msg-3")

	assert.NoError(t, err)
}

// ---- toolHandle ----

func TestToolHandle_NilAnswer(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := &agentSvc{SvcBase: service.NewSvcBase()}

	dto := &toolHandleDto{
		exploreAnswerList: []*agentrespvo.AnswerExplore{nil},
		i:                 0,
		skillsItem:        &conversationmsgvo.SkillsProcessItem{},
		skillsProcess:     []*conversationmsgvo.SkillsProcessItem{},
	}

	result, err := svc.toolHandle(context.Background(), dto)

	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestToolHandle_AnswerIsString(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := &agentSvc{SvcBase: service.NewSvcBase()}

	dto := &toolHandleDto{
		exploreAnswerList: []*agentrespvo.AnswerExplore{{Answer: "hello"}},
		i:                 0,
		skillsItem:        &conversationmsgvo.SkillsProcessItem{},
		skillsProcess:     []*conversationmsgvo.SkillsProcessItem{},
	}

	result, err := svc.toolHandle(context.Background(), dto)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "hello", result[0].Text)
}

func TestToolHandle_AnswerIsNotString(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := cmpmock.NewMockLogger(ctrl)
	svc := &agentSvc{
		SvcBase: service.NewSvcBase(),
		logger:  mockLogger,
	}

	dto := &toolHandleDto{
		exploreAnswerList: []*agentrespvo.AnswerExplore{{Answer: map[string]string{"key": "val"}}},
		i:                 0,
		skillsItem:        &conversationmsgvo.SkillsProcessItem{},
		skillsProcess:     []*conversationmsgvo.SkillsProcessItem{},
	}

	result, err := svc.toolHandle(context.Background(), dto)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Contains(t, result[0].Text, "key")
}

// ---- agentToolHandle ----

func TestAgentToolHandle_NilAnswer(t *testing.T) {
	t.Parallel()

	svc := &agentSvc{SvcBase: service.NewSvcBase()}

	dto := &agentToolHandleDto{
		exploreAnswerList: []*agentrespvo.AnswerExplore{nil},
		i:                 0,
		skillsItem:        &conversationmsgvo.SkillsProcessItem{},
		skillsProcess:     []*conversationmsgvo.SkillsProcessItem{},
	}

	result, err := svc.agentToolHandle(context.Background(), dto)

	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestAgentToolHandle_AnswerIsString(t *testing.T) {
	t.Parallel()

	svc := &agentSvc{SvcBase: service.NewSvcBase()}

	dto := &agentToolHandleDto{
		exploreAnswerList: []*agentrespvo.AnswerExplore{{Answer: "agent answer"}},
		i:                 0,
		skillsItem:        &conversationmsgvo.SkillsProcessItem{},
		skillsProcess:     []*conversationmsgvo.SkillsProcessItem{},
	}

	result, err := svc.agentToolHandle(context.Background(), dto)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "agent answer", result[0].Text)
}

func TestAgentToolHandle_AnswerIsNotString(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := cmpmock.NewMockLogger(ctrl)
	svc := &agentSvc{
		SvcBase: service.NewSvcBase(),
		logger:  mockLogger,
	}

	dto := &agentToolHandleDto{
		exploreAnswerList: []*agentrespvo.AnswerExplore{{Answer: map[string]int{"a": 1}}},
		i:                 0,
		skillsItem:        &conversationmsgvo.SkillsProcessItem{},
		skillsProcess:     []*conversationmsgvo.SkillsProcessItem{},
	}

	result, err := svc.agentToolHandle(context.Background(), dto)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
}

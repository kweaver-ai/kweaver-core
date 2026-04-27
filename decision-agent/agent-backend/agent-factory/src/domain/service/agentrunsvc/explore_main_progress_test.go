package agentsvc

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/agentrespvo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/conversationmsgvo"
	agentreq "github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/agent/req"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/cmp/icmp/cmpmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/stretchr/testify/assert"
)

// ---- mainHandle ----

func TestMainHandle_NilAnswer(t *testing.T) {
	t.Parallel()

	svc := &agentSvc{SvcBase: service.NewSvcBase()}

	thinking := ""
	dto := &mainHandleDto{
		exploreAnswerList: []*agentrespvo.AnswerExplore{nil},
		i:                 0,
		skillsItem:        &conversationmsgvo.SkillsProcessItem{},
		mainThinking:      &thinking,
		skillsProcess:     []*conversationmsgvo.SkillsProcessItem{},
	}

	result := svc.mainHandle(context.Background(), dto)
	assert.Empty(t, result)
}

func TestMainHandle_EmptyTextAndThinking(t *testing.T) {
	t.Parallel()

	svc := &agentSvc{SvcBase: service.NewSvcBase()}

	thinking := ""
	dto := &mainHandleDto{
		exploreAnswerList: []*agentrespvo.AnswerExplore{{Answer: "", Think: ""}},
		i:                 0,
		skillsItem:        &conversationmsgvo.SkillsProcessItem{},
		mainThinking:      &thinking,
		skillsProcess:     []*conversationmsgvo.SkillsProcessItem{},
	}

	result := svc.mainHandle(context.Background(), dto)
	// both text and thinking empty → not appended
	assert.Empty(t, result)
}

func TestMainHandle_WithText(t *testing.T) {
	t.Parallel()

	svc := &agentSvc{SvcBase: service.NewSvcBase()}

	thinking := ""
	dto := &mainHandleDto{
		exploreAnswerList: []*agentrespvo.AnswerExplore{{Answer: "hello", Think: "think1"}},
		i:                 0,
		skillsItem:        &conversationmsgvo.SkillsProcessItem{},
		mainThinking:      &thinking,
		skillsProcess:     []*conversationmsgvo.SkillsProcessItem{},
	}

	result := svc.mainHandle(context.Background(), dto)
	assert.Len(t, result, 1)
	assert.Equal(t, "hello", result[0].Text)
	assert.Equal(t, "think1", thinking)
}

// ---- handleExplore ----

func TestHandleExplore_EmptyList(t *testing.T) {
	t.Parallel()

	svc := &agentSvc{SvcBase: service.NewSvcBase()}

	dto := handleExploreDto{
		exploreAnswerList: []*agentrespvo.AnswerExplore{},
		nameToTypeMap:     map[string]string{},
	}

	mainThinking, skillsProcess, err := svc.handleExplore(context.Background(), dto)
	assert.NoError(t, err)
	assert.Empty(t, mainThinking)
	assert.Empty(t, skillsProcess)
}

func TestHandleExplore_MainSkill(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := &agentSvc{SvcBase: service.NewSvcBase()}

	dto := handleExploreDto{
		exploreAnswerList: []*agentrespvo.AnswerExplore{
			{AgentName: "main", Answer: "main answer", Think: "main thinking"},
		},
		nameToTypeMap: map[string]string{"main": "main"},
	}

	_, skillsProcess, err := svc.handleExplore(context.Background(), dto)
	assert.NoError(t, err)
	assert.Len(t, skillsProcess, 1)
	assert.Equal(t, "main answer", skillsProcess[0].Text)
}

func TestHandleExplore_ToolSkill(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := cmpmock.NewMockLogger(ctrl)
	svc := &agentSvc{
		SvcBase: service.NewSvcBase(),
		logger:  mockLogger,
	}

	dto := handleExploreDto{
		exploreAnswerList: []*agentrespvo.AnswerExplore{
			{AgentName: "tool_skill", Answer: "tool answer", Think: ""},
		},
		nameToTypeMap: map[string]string{"tool_skill": "tool"},
	}

	_, skillsProcess, err := svc.handleExplore(context.Background(), dto)
	assert.NoError(t, err)
	assert.Len(t, skillsProcess, 1)
	assert.Equal(t, "tool answer", skillsProcess[0].Text)
}

func TestHandleExplore_AgentSkill(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := cmpmock.NewMockLogger(ctrl)
	svc := &agentSvc{
		SvcBase: service.NewSvcBase(),
		logger:  mockLogger,
	}

	dto := handleExploreDto{
		exploreAnswerList: []*agentrespvo.AnswerExplore{
			{AgentName: "agent_skill", Answer: "agent answer", Think: ""},
		},
		nameToTypeMap: map[string]string{"agent_skill": "agent"},
	}

	_, skillsProcess, err := svc.handleExplore(context.Background(), dto)
	assert.NoError(t, err)
	assert.Len(t, skillsProcess, 1)
	assert.Equal(t, "agent answer", skillsProcess[0].Text)
}

// ---- handleProgress ----

func TestHandleProgress_EmptyProgresses_NoInterrupt(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := cmpmock.NewMockLogger(ctrl)
	svc := &agentSvc{
		SvcBase: service.NewSvcBase(),
		logger:  mockLogger,
	}

	mockLogger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()

	req := &agentreq.ChatReq{
		InternalParam: agentreq.InternalParam{AssistantMessageID: "msg-hp-1"},
	}

	result, err := svc.handleProgress(context.Background(), req, nil, 0)
	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestHandleProgress_WithCompletedAndProcessing(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := cmpmock.NewMockLogger(ctrl)
	svc := &agentSvc{
		SvcBase: service.NewSvcBase(),
		logger:  mockLogger,
	}

	mockLogger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()

	// Clean up any cached state
	progressSet.Delete("msg-hp-2")
	progressMap.Delete("msg-hp-2")

	req := &agentreq.ChatReq{
		InternalParam: agentreq.InternalParam{AssistantMessageID: "msg-hp-2"},
	}

	progresses := []*agentrespvo.Progress{
		{ID: "pg-1", Status: "completed"},
		{ID: "pg-2", Status: "processing"},
		{ID: "pg-3", Status: "failed"},
		{ID: "pg-4", Status: "skipped"},
	}

	result, err := svc.handleProgress(context.Background(), req, progresses, 0)
	assert.NoError(t, err)
	// completed + failed + skipped → 3, plus processing currentProgress → 4 total
	assert.Len(t, result, 4)
}

// ---- handleProgressOld ----

func TestHandleProgressOld_EmptyProgresses_NoInterrupt(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := cmpmock.NewMockLogger(ctrl)
	svc := &agentSvc{
		SvcBase: service.NewSvcBase(),
		logger:  mockLogger,
	}

	mockLogger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()

	progressSet.Delete("msg-hpo-1")
	progressMap.Delete("msg-hpo-1")

	req := &agentreq.ChatReq{
		InternalParam: agentreq.InternalParam{AssistantMessageID: "msg-hpo-1"},
	}

	result, err := svc.handleProgressOld(context.Background(), req, nil, 0)
	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestHandleProgressOld_WithCompletedAndProcessing(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := cmpmock.NewMockLogger(ctrl)
	svc := &agentSvc{
		SvcBase: service.NewSvcBase(),
		logger:  mockLogger,
	}

	mockLogger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()

	progressSet.Delete("msg-hpo-2")
	progressMap.Delete("msg-hpo-2")

	req := &agentreq.ChatReq{
		InternalParam: agentreq.InternalParam{AssistantMessageID: "msg-hpo-2"},
	}

	progresses := []*agentrespvo.Progress{
		{ID: "pg-old-1", Status: "completed"},
		{ID: "pg-old-2", Status: "processing"},
		{ID: "pg-old-3", Status: "failed"},
	}

	result, err := svc.handleProgressOld(context.Background(), req, progresses, 0)
	assert.NoError(t, err)
	// completed + failed → 2, plus processing → 3
	assert.Len(t, result, 3)
}

// ---- forResumeInterrupt ----

func TestForResumeInterrupt_NoInterruptedMsgID(t *testing.T) {
	t.Parallel()

	svc := &agentSvc{SvcBase: service.NewSvcBase()}

	req := &agentreq.ChatReq{
		InternalParam:             agentreq.InternalParam{AssistantMessageID: "msg-fri-1"},
		InterruptedAssistantMsgID: "",
	}

	result, err := svc.forResumeInterrupt(context.Background(), req)
	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestForResumeInterrupt_UnmarshalError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConvMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
	svc := &agentSvc{
		SvcBase:             service.NewSvcBase(),
		conversationMsgRepo: mockConvMsgRepo,
	}

	isInterruptPreProgressGetMap.Delete("msg-fri-5")

	invalidJSON := "{invalid-json"
	msgPO := &dapo.ConversationMsgPO{ID: "interrupted-msg-invalid", Content: &invalidJSON}
	mockConvMsgRepo.EXPECT().GetByID(gomock.Any(), "interrupted-msg-invalid").Return(msgPO, nil)

	req := &agentreq.ChatReq{
		InternalParam:             agentreq.InternalParam{AssistantMessageID: "msg-fri-5"},
		InterruptedAssistantMsgID: "interrupted-msg-invalid",
	}

	result, err := svc.forResumeInterrupt(context.Background(), req)
	assert.Error(t, err)
	assert.Empty(t, result)
}

func TestForResumeInterrupt_MiddleAnswerNilAndValidProgress(t *testing.T) {
	t.Parallel()

	t.Run("middle_answer is nil should warn and return empty", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockConvMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
		mockLogger := cmpmock.NewMockLogger(ctrl)
		mockLogger.EXPECT().Warnf(gomock.Any(), gomock.Any()).AnyTimes()

		svc := &agentSvc{
			SvcBase:             service.NewSvcBase(),
			conversationMsgRepo: mockConvMsgRepo,
			logger:              mockLogger,
		}

		isInterruptPreProgressGetMap.Delete("msg-fri-6")

		content := `{"middle_answer":null}`
		msgPO := &dapo.ConversationMsgPO{ID: "interrupted-msg-nil", Content: &content}
		mockConvMsgRepo.EXPECT().GetByID(gomock.Any(), "interrupted-msg-nil").Return(msgPO, nil)

		req := &agentreq.ChatReq{
			InternalParam:             agentreq.InternalParam{AssistantMessageID: "msg-fri-6"},
			InterruptedAssistantMsgID: "interrupted-msg-nil",
		}

		result, err := svc.forResumeInterrupt(context.Background(), req)
		assert.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("valid progress should append and mark fetched", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockConvMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
		svc := &agentSvc{
			SvcBase:             service.NewSvcBase(),
			conversationMsgRepo: mockConvMsgRepo,
		}

		isInterruptPreProgressGetMap.Delete("msg-fri-7")

		content := `{"middle_answer":{"progress":[{"id":"pg-1","status":"completed"}]}}`
		msgPO := &dapo.ConversationMsgPO{ID: "interrupted-msg-ok", Content: &content}
		mockConvMsgRepo.EXPECT().GetByID(gomock.Any(), "interrupted-msg-ok").Return(msgPO, nil)

		req := &agentreq.ChatReq{
			InternalParam:             agentreq.InternalParam{AssistantMessageID: "msg-fri-7"},
			InterruptedAssistantMsgID: "interrupted-msg-ok",
		}

		result, err := svc.forResumeInterrupt(context.Background(), req)
		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "pg-1", result[0].ID)

		_, ok := isInterruptPreProgressGetMap.Load("msg-fri-7")
		assert.True(t, ok)
	})
}

func TestForResumeInterrupt_AlreadyFetched(t *testing.T) {
	t.Parallel()

	svc := &agentSvc{SvcBase: service.NewSvcBase()}

	isInterruptPreProgressGetMap.Store("msg-fri-2", true)
	defer isInterruptPreProgressGetMap.Delete("msg-fri-2")

	req := &agentreq.ChatReq{
		InternalParam:             agentreq.InternalParam{AssistantMessageID: "msg-fri-2"},
		InterruptedAssistantMsgID: "interrupted-msg-x",
	}

	result, err := svc.forResumeInterrupt(context.Background(), req)
	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestForResumeInterrupt_GetByIDError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConvMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
	svc := &agentSvc{
		SvcBase:             service.NewSvcBase(),
		conversationMsgRepo: mockConvMsgRepo,
	}

	isInterruptPreProgressGetMap.Delete("msg-fri-3")
	mockConvMsgRepo.EXPECT().GetByID(gomock.Any(), "interrupted-msg-y").Return(nil, errors.New("db error"))

	req := &agentreq.ChatReq{
		InternalParam:             agentreq.InternalParam{AssistantMessageID: "msg-fri-3"},
		InterruptedAssistantMsgID: "interrupted-msg-y",
	}

	result, err := svc.forResumeInterrupt(context.Background(), req)
	assert.Error(t, err)
	assert.Empty(t, result)
}

func TestForResumeInterrupt_NilContent(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConvMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)
	svc := &agentSvc{
		SvcBase:             service.NewSvcBase(),
		conversationMsgRepo: mockConvMsgRepo,
		logger:              mockLogger,
	}

	isInterruptPreProgressGetMap.Delete("msg-fri-4")

	msgPO := &dapo.ConversationMsgPO{ID: "interrupted-msg-z", Content: nil}
	mockConvMsgRepo.EXPECT().GetByID(gomock.Any(), "interrupted-msg-z").Return(msgPO, nil)

	req := &agentreq.ChatReq{
		InternalParam:             agentreq.InternalParam{AssistantMessageID: "msg-fri-4"},
		InterruptedAssistantMsgID: "interrupted-msg-z",
	}

	result, err := svc.forResumeInterrupt(context.Background(), req)
	assert.NoError(t, err)
	assert.Empty(t, result)
}

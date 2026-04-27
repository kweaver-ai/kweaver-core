package agentsvc

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/conf"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/constant"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/comvalobj"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/daconfvalobj"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/httpaccess/sandboxplatformhttp/sandboxplatformdto"
	agentreq "github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/agent/req"
	agentresp "github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/agent/resp"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/cmp/icmp/cmpmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/idbaccess/idbaccessmock"
)

// ---------- GenerateAgentCallReq 额外分支 ----------

func TestAgentSvc_GenerateAgentCallReq_WithSelectedFiles(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := cmpmock.NewMockLogger(ctrl)
	mockLogger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
	svc := &agentSvc{SvcBase: service.NewSvcBase(), logger: mockLogger}
	ctx := context.Background()
	req := &agentreq.ChatReq{
		AgentID:        "agent-1",
		ConversationID: "conv-1",
		Query:          "hello",
		InternalParam:  agentreq.InternalParam{UserID: "u1"},
		SelectedFiles:  []agentreq.SelectedFile{{FileName: "file1.txt"}},
	}
	agent := newTestAgent()
	result, err := svc.GenerateAgentCallReq(ctx, req, nil, agent)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result)
}

func TestAgentSvc_GenerateAgentCallReq_RegenerateWithModelName(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := cmpmock.NewMockLogger(ctrl)
	mockLogger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
	svc := &agentSvc{SvcBase: service.NewSvcBase(), logger: mockLogger}
	ctx := context.Background()
	req := &agentreq.ChatReq{
		AgentID:                  "agent-1",
		Query:                    "hello",
		InternalParam:            agentreq.InternalParam{UserID: "u1"},
		RegenerateAssistantMsgID: "asst-123",
		ModelName:                "gpt-4",
	}
	agent := newTestAgent()
	result, err := svc.GenerateAgentCallReq(ctx, req, nil, agent)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestAgentSvc_GenerateAgentCallReq_RegenerateNoModelName(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := cmpmock.NewMockLogger(ctrl)
	mockLogger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
	svc := &agentSvc{SvcBase: service.NewSvcBase(), logger: mockLogger}
	ctx := context.Background()
	req := &agentreq.ChatReq{
		AgentID:                  "agent-1",
		Query:                    "hello",
		InternalParam:            agentreq.InternalParam{UserID: "u1"},
		RegenerateAssistantMsgID: "asst-123",
		ModelName:                "",
	}
	agent := newTestAgent()
	result, err := svc.GenerateAgentCallReq(ctx, req, nil, agent)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestAgentSvc_GenerateAgentCallReq_WithFileField(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := cmpmock.NewMockLogger(ctrl)
	mockLogger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
	svc := &agentSvc{SvcBase: service.NewSvcBase(), logger: mockLogger}
	ctx := context.Background()
	req := &agentreq.ChatReq{
		AgentID:   "agent-1",
		Query:     "hello",
		TempFiles: []valueobject.TempFile{{ID: "tmp1.pdf", Name: "tmp1.pdf"}},
	}
	agent := newTestAgent()
	agent.Config = daconfvalobj.Config{
		Input: &daconfvalobj.Input{
			Fields: daconfvalobj.Fields{
				{Name: "doc", Type: "file"},
				{Name: "custom_field", Type: "string"},
				{Name: "history", Type: "string"},
			},
		},
	}
	result, err := svc.GenerateAgentCallReq(ctx, req, nil, agent)
	assert.NoError(t, err)
	assert.Equal(t, req.TempFiles, result.Input["doc"])
}

func TestAgentSvc_GenerateAgentCallReq_DeepThinkingWithLLMs(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := cmpmock.NewMockLogger(ctrl)
	mockLogger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
	svc := &agentSvc{SvcBase: service.NewSvcBase(), logger: mockLogger}
	ctx := context.Background()
	req := &agentreq.ChatReq{
		AgentID:  "agent-1",
		Query:    "deep question",
		ChatMode: constant.DeepThinkingMode,
	}
	agent := newTestAgent()
	result, err := svc.GenerateAgentCallReq(ctx, req, nil, agent)
	assert.NoError(t, err)
	assert.Equal(t, constant.DeepThinkingMode, req.ChatMode)
	assert.NotNil(t, result)
}

func TestAgentSvc_GenerateAgentCallReq_DeepThinkingSwitchDefaultModel(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := cmpmock.NewMockLogger(ctrl)
	mockLogger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
	svc := &agentSvc{SvcBase: service.NewSvcBase(), logger: mockLogger}

	req := &agentreq.ChatReq{
		AgentID:       "agent-1",
		Query:         "deep question",
		ChatMode:      constant.DeepThinkingMode,
		InternalParam: agentreq.InternalParam{UserID: "u1"},
	}

	agent := newTestAgent()
	agent.Config.Llms = []*daconfvalobj.LlmItem{
		{
			IsDefault: true,
			LlmConfig: &daconfvalobj.LlmConfig{
				Name:        "default-llm",
				ModelType:   cdaenum.ModelTypeLlm,
				Temperature: 0.2,
				TopK:        1,
				MaxTokens:   1024,
			},
		},
		{
			IsDefault: false,
			LlmConfig: &daconfvalobj.LlmConfig{
				Name:        "reasoning-llm",
				ModelType:   cdaenum.ModelTypeRlm,
				Temperature: 0.2,
				TopK:        1,
				MaxTokens:   1024,
			},
		},
	}

	result, err := svc.GenerateAgentCallReq(context.Background(), req, nil, agent)
	require.NoError(t, err)
	require.NotNil(t, result)

	var defaultLlm, defaultRlm bool

	for _, llm := range result.Config.Llms {
		if llm.LlmConfig != nil && llm.LlmConfig.Name == "default-llm" {
			defaultLlm = llm.IsDefault
		}

		if llm.LlmConfig != nil && llm.LlmConfig.Name == "reasoning-llm" {
			defaultRlm = llm.IsDefault
		}
	}

	assert.False(t, defaultLlm)
	assert.True(t, defaultRlm)
}

func TestAgentSvc_GenerateAgentCallReq_RegenerateRaisesTemperatureAndTopK(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := cmpmock.NewMockLogger(ctrl)
	mockLogger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
	svc := &agentSvc{SvcBase: service.NewSvcBase(), logger: mockLogger}

	req := &agentreq.ChatReq{
		AgentID:                  "agent-1",
		Query:                    "regen",
		RegenerateAssistantMsgID: "asst-1",
		ModelName:                "target-llm",
		InternalParam:            agentreq.InternalParam{UserID: "u1"},
	}

	agent := newTestAgent()
	agent.Config.Llms = []*daconfvalobj.LlmItem{
		{
			IsDefault: true,
			LlmConfig: &daconfvalobj.LlmConfig{
				Name:        "target-llm",
				ModelType:   cdaenum.ModelTypeLlm,
				Temperature: 0.3,
				TopK:        1,
				MaxTokens:   1024,
			},
		},
	}

	result, err := svc.GenerateAgentCallReq(context.Background(), req, nil, agent)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Config.Llms, 1)

	assert.Equal(t, 0.8, result.Config.Llms[0].LlmConfig.Temperature)
	assert.Equal(t, 10, result.Config.Llms[0].LlmConfig.TopK)
}

func TestAgentSvc_GenerateAgentCallReq_InjectWorkspaceContextAndInputMapping(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := cmpmock.NewMockLogger(ctrl)
	mockLogger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
	svc := &agentSvc{SvcBase: service.NewSvcBase(), logger: mockLogger}

	req := &agentreq.ChatReq{
		AgentID:        "agent-1",
		ConversationID: "conv-ctx-1",
		Query:          "query",
		TempFiles:      []valueobject.TempFile{{ID: "tmp1", Name: "tmp1"}},
		CustomQuerys: map[string]interface{}{
			"city": "beijing",
		},
		SelectedFiles: []agentreq.SelectedFile{{FileName: "design.md"}},
		InternalParam: agentreq.InternalParam{UserID: "u1"},
	}

	contexts := []*comvalobj.LLMMessage{{Role: "user", Content: "old context"}}
	agent := newTestAgent()
	agent.Config.Input.Fields = daconfvalobj.Fields{
		&daconfvalobj.Field{Name: "doc", Type: cdaenum.InputFieldTypeFile},
		&daconfvalobj.Field{Name: "city", Type: cdaenum.InputFieldTypeString},
		&daconfvalobj.Field{Name: "header", Type: cdaenum.InputFieldTypeString},
	}

	result, err := svc.GenerateAgentCallReq(context.Background(), req, contexts, agent)
	require.NoError(t, err)
	require.NotNil(t, result)

	history, ok := result.Input["history"].([]*comvalobj.LLMMessage)
	require.True(t, ok)
	require.Len(t, history, 2)
	assert.Equal(t, "old context", history[0].Content)
	assert.True(t, strings.Contains(history[1].Content, "design.md"))

	assert.Equal(t, req.TempFiles, result.Input["doc"])
	assert.Equal(t, "beijing", result.Input["city"])
	_, exists := result.Input["header"]
	assert.False(t, exists)
}

// ---------- TerminateChat 测试 ----------

func TestAgentSvc_TerminateChat_StopChanNotFound(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := cmpmock.NewMockLogger(ctrl)
	mockLogger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()
	svc := &agentSvc{SvcBase: service.NewSvcBase(), logger: mockLogger}
	ctx := context.Background()
	err := svc.TerminateChat(ctx, "conv-nonexistent", "", "")
	assert.Error(t, err)
}

func TestAgentSvc_TerminateChat_StopChanNotFoundButHasInterrupted(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)
	mockLogger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()
	svc := &agentSvc{SvcBase: service.NewSvcBase(), logger: mockLogger, conversationMsgRepo: mockMsgRepo}

	msgPO := &dapo.ConversationMsgPO{ID: "asst-1"}
	mockMsgRepo.EXPECT().GetByID(gomock.Any(), "asst-1").Return(msgPO, nil)
	mockMsgRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

	ctx := context.Background()
	err := svc.TerminateChat(ctx, "conv-no-chan", "", "asst-1")
	assert.NoError(t, err)
}

func TestAgentSvc_TerminateChat_StopChanFound(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := cmpmock.NewMockLogger(ctrl)
	mockLogger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
	svc := &agentSvc{SvcBase: service.NewSvcBase(), logger: mockLogger}

	// Register a stopchan
	ch := make(chan struct{}, 1)
	stopChanMap.Store("conv-with-chan", ch)

	ctx := context.Background()
	err := svc.TerminateChat(ctx, "conv-with-chan", "", "")
	assert.NoError(t, err)
}

func TestAgentSvc_TerminateChat_InterruptedMsgGetError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)
	mockLogger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()
	svc := &agentSvc{SvcBase: service.NewSvcBase(), logger: mockLogger, conversationMsgRepo: mockMsgRepo}

	ch := make(chan struct{}, 1)
	stopChanMap.Store("conv-with-chan-2", ch)
	mockMsgRepo.EXPECT().GetByID(gomock.Any(), "asst-err").Return(nil, errors.New("not found"))

	ctx := context.Background()
	err := svc.TerminateChat(ctx, "conv-with-chan-2", "", "asst-err")
	assert.Error(t, err)
}

// ---------- waitForSessionReady 额外分支测试 ----------

func TestAgentSvc_WaitForSessionReady_SessionErrorState(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := cmpmock.NewMockLogger(ctrl)
	allowAnyLoggerCalls(mockLogger)

	getCallCount := 0
	mockSandbox := &mockGetSessionSandbox{
		getSessionFunc: func(ctx context.Context, sessionID string) (*sandboxplatformdto.GetSessionResp, error) {
			getCallCount++
			return &sandboxplatformdto.GetSessionResp{ID: sessionID, Status: "error"}, nil
		},
	}

	svc := &agentSvc{
		SvcBase:         service.NewSvcBase(),
		logger:          mockLogger,
		sandboxPlatform: mockSandbox,
		sandboxPlatformConf: &conf.SandboxPlatformConf{
			MaxRetries:    2,
			RetryInterval: "1ms",
		},
	}

	ctx := context.Background()
	_, err := svc.waitForSessionReady(ctx, "sess-err")
	assert.Error(t, err)
}

func TestAgentSvc_WaitForSessionReady_Timeout(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := cmpmock.NewMockLogger(ctrl)
	allowAnyLoggerCalls(mockLogger)

	mockSandbox := &mockGetSessionSandbox{
		getSessionFunc: func(ctx context.Context, sessionID string) (*sandboxplatformdto.GetSessionResp, error) {
			return &sandboxplatformdto.GetSessionResp{ID: sessionID, Status: "pending"}, nil
		},
	}

	svc := &agentSvc{
		SvcBase:         service.NewSvcBase(),
		logger:          mockLogger,
		sandboxPlatform: mockSandbox,
		sandboxPlatformConf: &conf.SandboxPlatformConf{
			MaxRetries:    2,
			RetryInterval: "1ms",
		},
	}

	ctx := context.Background()
	_, err := svc.waitForSessionReady(ctx, "sess-timeout")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout")
}

func TestAgentSvc_WaitForSessionReady_InvalidInterval(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := cmpmock.NewMockLogger(ctrl)
	allowAnyLoggerCalls(mockLogger)

	mockSandbox := &mockGetSessionSandbox{
		getSessionFunc: func(ctx context.Context, sessionID string) (*sandboxplatformdto.GetSessionResp, error) {
			return &sandboxplatformdto.GetSessionResp{ID: sessionID, Status: "running"}, nil
		},
	}

	svc := &agentSvc{
		SvcBase:         service.NewSvcBase(),
		logger:          mockLogger,
		sandboxPlatform: mockSandbox,
		sandboxPlatformConf: &conf.SandboxPlatformConf{
			MaxRetries:    2,
			RetryInterval: "invalid",
		},
	}

	ctx := context.Background()
	id, err := svc.waitForSessionReady(ctx, "sess-ok")
	assert.NoError(t, err)
	assert.Equal(t, "sess-ok", id)
}

func TestAgentSvc_WaitForSessionReady_GetSessionError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := cmpmock.NewMockLogger(ctrl)
	allowAnyLoggerCalls(mockLogger)

	mockSandbox := &mockGetSessionSandbox{
		getSessionFunc: func(ctx context.Context, sessionID string) (*sandboxplatformdto.GetSessionResp, error) {
			return nil, errors.New("network error")
		},
	}

	svc := &agentSvc{
		SvcBase:         service.NewSvcBase(),
		logger:          mockLogger,
		sandboxPlatform: mockSandbox,
		sandboxPlatformConf: &conf.SandboxPlatformConf{
			MaxRetries:    1,
			RetryInterval: "1ms",
		},
	}

	ctx := context.Background()
	_, err := svc.waitForSessionReady(ctx, "sess-get-err")
	assert.Error(t, err)
}

// ---------- createNewSession 额外分支 ----------

func TestAgentSvc_CreateNewSession_DefaultResourceValues(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := cmpmock.NewMockLogger(ctrl)
	allowAnyLoggerCalls(mockLogger)

	getCallCount := 0
	mockSandbox := &mockGetSessionSandbox{
		getSessionFunc: func(ctx context.Context, sessionID string) (*sandboxplatformdto.GetSessionResp, error) {
			getCallCount++
			return &sandboxplatformdto.GetSessionResp{ID: sessionID, Status: "running"}, nil
		},
		createSessionFunc: func(ctx context.Context, req sandboxplatformdto.CreateSessionReq) (*sandboxplatformdto.CreateSessionResp, error) {
			// Verify defaults applied
			assert.Equal(t, "1", req.CPU)
			assert.Equal(t, "512Mi", req.Memory)
			assert.Equal(t, "1Gi", req.Disk)
			assert.Equal(t, 300, req.Timeout)
			return &sandboxplatformdto.CreateSessionResp{ID: *req.ID}, nil
		},
	}

	svc := &agentSvc{
		SvcBase:         service.NewSvcBase(),
		logger:          mockLogger,
		sandboxPlatform: mockSandbox,
		sandboxPlatformConf: &conf.SandboxPlatformConf{
			MaxRetries:    1,
			RetryInterval: "1ms",
		},
	}

	ctx := context.Background()
	req := &agentreq.ChatReq{AgentID: "a1", InternalParam: agentreq.InternalParam{UserID: "u1"}}
	id, err := svc.createNewSession(ctx, "new-sess", req)
	assert.NoError(t, err)
	assert.Equal(t, "new-sess", id)
}

func TestAgentSvc_CreateNewSession_EmptyCreateRespID(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := cmpmock.NewMockLogger(ctrl)
	allowAnyLoggerCalls(mockLogger)

	mockSandbox := &mockGetSessionSandbox{
		getSessionFunc: func(ctx context.Context, sessionID string) (*sandboxplatformdto.GetSessionResp, error) {
			return &sandboxplatformdto.GetSessionResp{ID: sessionID, Status: "running"}, nil
		},
		createSessionFunc: func(ctx context.Context, req sandboxplatformdto.CreateSessionReq) (*sandboxplatformdto.CreateSessionResp, error) {
			return &sandboxplatformdto.CreateSessionResp{ID: ""}, nil // empty ID
		},
	}

	svc := &agentSvc{
		SvcBase:         service.NewSvcBase(),
		logger:          mockLogger,
		sandboxPlatform: mockSandbox,
		sandboxPlatformConf: &conf.SandboxPlatformConf{
			MaxRetries:    1,
			RetryInterval: "1ms",
		},
	}

	ctx := context.Background()
	req := &agentreq.ChatReq{AgentID: "a1"}
	id, err := svc.createNewSession(ctx, "sess-empty-resp", req)
	assert.NoError(t, err)
	assert.Equal(t, "sess-empty-resp", id)
}

func TestAgentSvc_CreateNewSession_CreateFailed(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := cmpmock.NewMockLogger(ctrl)
	allowAnyLoggerCalls(mockLogger)

	mockSandbox := &mockGetSessionSandbox{
		createSessionFunc: func(ctx context.Context, req sandboxplatformdto.CreateSessionReq) (*sandboxplatformdto.CreateSessionResp, error) {
			return nil, errors.New("create failed")
		},
	}

	svc := &agentSvc{
		SvcBase:         service.NewSvcBase(),
		logger:          mockLogger,
		sandboxPlatform: mockSandbox,
		sandboxPlatformConf: &conf.SandboxPlatformConf{
			MaxRetries:    1,
			RetryInterval: "1ms",
		},
	}

	ctx := context.Background()
	req := &agentreq.ChatReq{AgentID: "a1"}
	_, err := svc.createNewSession(ctx, "sess-fail", req)
	assert.Error(t, err)
}

// ---------- NewStreamingResponseLogger DEBUG 模式测试 ----------

func TestNewStreamingResponseLogger_DebugMode_Success(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	// IsDebugMode checks <SERVICE_NAME>_DEBUG_MODE env var at init time; we test via direct struct construction
	// Just verify LogChunk and Complete work on a real logger instance
	f, err := os.CreateTemp(tmpDir, "stream-*.log")
	if err != nil {
		t.Fatal(err)
	}

	f.Close()

	f2, err := os.Create(f.Name())
	if err != nil {
		t.Fatal(err)
	}

	l := &StreamingResponseLogger{
		file:           f2,
		conversationID: "conv-debug-test",
		logType:        ExecutorResponse,
	}
	l.LogChunk([]byte("test data"))
	assert.Equal(t, 1, l.chunksCount)
	l.Complete()
}

func TestNewStreamingResponseLogger_DebugMode_InvalidDir(t *testing.T) {
	// t.Parallel() - 移除：此测试使用 t.Setenv() 修改环境变量，不能与 t.Parallel() 同时使用
	t.Setenv("APP_DEBUG", "true")
	// Use a path that can't be created under /proc (on Linux) or a file as dir
	tmpFile, err := os.CreateTemp("", "testfile")
	if err != nil {
		t.Skip("cannot create temp file for test")
	}

	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Set log root to an existing file (not a dir) — MkdirAll will fail
	t.Setenv("AGENT_FACTORY_LOCAL_DEV_LOG_ROOT_DIR", tmpFile.Name())

	_, _ = NewStreamingResponseLogger("conv-fail", ProcessedResponse)
	// We just ensure it doesn't panic — error is acceptable
}

// ---------- UpsertUserAndAssistantMsg 额外分支 ----------

func TestAgentSvc_UpsertUserAndAssistantMsg_RegenerateAssistantMsg_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
	svc := &agentSvc{SvcBase: service.NewSvcBase(), conversationMsgRepo: mockMsgRepo}

	asstMsg := &dapo.ConversationMsgPO{ID: "asst-123", ReplyID: "user-1", Index: 2}
	gomock.InOrder(
		mockMsgRepo.EXPECT().GetByID(gomock.Any(), "asst-123").Return(asstMsg, nil),
		mockMsgRepo.EXPECT().GetByID(gomock.Any(), "asst-123").Return(asstMsg, nil),
		mockMsgRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil),
	)

	ctx := context.Background()
	req := &agentreq.ChatReq{
		ConversationID:           "conv-1",
		RegenerateAssistantMsgID: "asst-123",
		InternalParam:            agentreq.InternalParam{UserID: "u1"},
	}
	userID, asstID, idx, err := svc.UpsertUserAndAssistantMsg(ctx, req, 0, &dapo.ConversationPO{ID: "conv-1"})
	assert.NoError(t, err)
	assert.Equal(t, "user-1", userID)
	assert.Equal(t, "asst-123", asstID)
	assert.Equal(t, 2, idx)
}

func TestAgentSvc_UpsertUserAndAssistantMsg_InterruptedMsg_UpdateError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
	svc := &agentSvc{SvcBase: service.NewSvcBase(), conversationMsgRepo: mockMsgRepo}

	interruptedMsg := &dapo.ConversationMsgPO{ID: "interrupted", ReplyID: "user-1", Index: 3}
	gomock.InOrder(
		mockMsgRepo.EXPECT().GetByID(gomock.Any(), "interrupted").Return(interruptedMsg, nil),
		mockMsgRepo.EXPECT().GetByID(gomock.Any(), "interrupted").Return(interruptedMsg, nil),
		mockMsgRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(errors.New("update error")),
	)

	ctx := context.Background()
	req := &agentreq.ChatReq{
		ConversationID:            "conv-1",
		InterruptedAssistantMsgID: "interrupted",
		InternalParam:             agentreq.InternalParam{UserID: "u1"},
	}
	_, _, _, err := svc.UpsertUserAndAssistantMsg(ctx, req, 0, &dapo.ConversationPO{ID: "conv-1"})
	assert.Error(t, err)
}

// ---------- GetHistoryAndMsgIndex with sql not found ----------

func TestAgentSvc_GetHistoryAndMsgIndex_ExistingConversation_MaxIndexNotFound(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockConvRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
	mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
	svc := &agentSvc{SvcBase: service.NewSvcBase(), conversationRepo: mockConvRepo, conversationMsgRepo: mockMsgRepo}

	mockConvRepo.EXPECT().GetByID(gomock.Any(), "conv-1").Return(&dapo.ConversationPO{ID: "conv-1"}, nil)
	// Return "sql: no rows" style error — chelper.IsSqlNotFound should match
	mockMsgRepo.EXPECT().GetMaxIndexByID(gomock.Any(), "conv-1").Return(0, errors.New("record not found"))

	ctx := context.Background()
	req := &agentreq.ChatReq{ConversationID: "conv-1"}
	_, _, idx, err := svc.GetHistoryAndMsgIndex(ctx, req, 0, nil)
	// record not found is not sql.ErrNoRows so it will be treated as general error
	assert.Error(t, err)

	_ = idx
}

// ---------- HandleStopChan Success ensures conversation update ----------

func TestAgentSvc_HandleStopChan_UpdateConversationError(t *testing.T) {
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
	mockConvRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(errors.New("update error"))

	session := &Session{ConversationID: "conv-1", TempMsgResp: agentresp.ChatResp{ConversationID: "conv-1"}}
	ctx := context.Background()
	req := &agentreq.ChatReq{AgentID: "a1", ConversationID: "conv-1", AgentRunID: "run-1", InternalParam: agentreq.InternalParam{UserID: "u1", AssistantMessageID: "asst-1"}}
	err := svc.HandleStopChan(ctx, req, session)
	assert.Error(t, err)
}

// Use time package to avoid import error.
var (
	_ = time.Now
	_ = agentresp.ChatResp{}
)

package agentsvc

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/conf"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/constant"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/constant/otelconst"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service"
	agentreq "github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/agent/req"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/cmp/icmp/cmpmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driver/iportdriver/iportdrivermock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driver/iv3portdriver/v3portdrivermock"
)

func TestAgentSvc_Chat_GetAgentInfoError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSquare := v3portdrivermock.NewMockISquareSvc(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)
	allowAnyLoggerCalls(mockLogger)

	svc := &agentSvc{
		SvcBase:             service.NewSvcBase(),
		squareSvc:           mockSquare,
		logger:              mockLogger,
		sandboxPlatformConf: &conf.SandboxPlatformConf{},
	}

	mockSquare.EXPECT().GetAgentInfoByIDOrKey(gomock.Any(), gomock.Any()).Return(nil, errors.New("agent not found"))

	ctx := context.Background()
	req := &agentreq.ChatReq{
		AgentID: "a1", AgentVersion: "v1",
		InternalParam: agentreq.InternalParam{UserID: "u1"},
	}
	_, err := svc.Chat(ctx, req)
	assert.Error(t, err)
}

func TestAgentSvc_Chat_APIChat_NotPublished(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSquare := v3portdrivermock.NewMockISquareSvc(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)
	allowAnyLoggerCalls(mockLogger)

	svc := &agentSvc{
		SvcBase:             service.NewSvcBase(),
		squareSvc:           mockSquare,
		logger:              mockLogger,
		sandboxPlatformConf: &conf.SandboxPlatformConf{},
	}

	agentInfo := newTestAgent()
	agentInfo.PublishInfo.IsAPIAgent = 0
	mockSquare.EXPECT().GetAgentInfoByIDOrKey(gomock.Any(), gomock.Any()).Return(agentInfo, nil)

	ctx := context.Background()
	req := &agentreq.ChatReq{
		AgentID: "a1",
		InternalParam: agentreq.InternalParam{
			UserID:   "u1",
			CallType: constant.APIChat,
		},
	}
	_, err := svc.Chat(ctx, req)
	assert.Error(t, err)
}

func TestAgentSvc_Chat_GetHistoryError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSquare := v3portdrivermock.NewMockISquareSvc(ctrl)
	mockConvRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)
	allowAnyLoggerCalls(mockLogger)

	svc := &agentSvc{
		SvcBase:             service.NewSvcBase(),
		squareSvc:           mockSquare,
		logger:              mockLogger,
		conversationRepo:    mockConvRepo,
		sandboxPlatformConf: &conf.SandboxPlatformConf{},
	}

	agentInfo := newTestAgent()
	mockSquare.EXPECT().GetAgentInfoByIDOrKey(gomock.Any(), gomock.Any()).Return(agentInfo, nil)
	mockConvRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, errors.New("db error"))

	ctx := context.Background()
	req := &agentreq.ChatReq{
		AgentID:       "a1",
		InternalParam: agentreq.InternalParam{UserID: "u1"},
	}
	_, err := svc.Chat(ctx, req)
	assert.Error(t, err)
}

func TestAgentSvc_Chat_UpsertMsgError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSquare := v3portdrivermock.NewMockISquareSvc(ctrl)
	mockConvRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
	mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)
	allowAnyLoggerCalls(mockLogger)

	svc := &agentSvc{
		SvcBase:             service.NewSvcBase(),
		squareSvc:           mockSquare,
		logger:              mockLogger,
		conversationRepo:    mockConvRepo,
		conversationMsgRepo: mockMsgRepo,
		sandboxPlatformConf: &conf.SandboxPlatformConf{},
	}

	agentInfo := newTestAgent()
	mockSquare.EXPECT().GetAgentInfoByIDOrKey(gomock.Any(), gomock.Any()).Return(agentInfo, nil)
	mockConvRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&dapo.ConversationPO{ID: "conv-1"}, nil)
	mockMsgRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return("", errors.New("msg create error"))

	ctx := context.Background()
	req := &agentreq.ChatReq{
		AgentID:       "a1",
		InternalParam: agentreq.InternalParam{UserID: "u1"},
	}
	_, err := svc.Chat(ctx, req)
	assert.Error(t, err)
}

func TestAgentSvc_Chat_SessionSvcError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSquare := v3portdrivermock.NewMockISquareSvc(ctrl)
	mockConvRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
	mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
	mockSessionSvc := iportdrivermock.NewMockISessionSvc(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)
	allowAnyLoggerCalls(mockLogger)

	svc := &agentSvc{
		SvcBase:             service.NewSvcBase(),
		squareSvc:           mockSquare,
		logger:              mockLogger,
		conversationRepo:    mockConvRepo,
		conversationMsgRepo: mockMsgRepo,
		sessionSvc:          mockSessionSvc,
		sandboxPlatformConf: &conf.SandboxPlatformConf{},
	}

	agentInfo := newTestAgent()
	mockSquare.EXPECT().GetAgentInfoByIDOrKey(gomock.Any(), gomock.Any()).Return(agentInfo, nil)
	mockConvRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&dapo.ConversationPO{ID: "conv-s1"}, nil)
	gomock.InOrder(
		mockMsgRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return("user-msg-1", nil),
		mockConvRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil),
		mockMsgRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return("asst-msg-1", nil),
	)
	mockSessionSvc.EXPECT().HandleGetInfoOrCreate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(int64(0), 0, errors.New("session error"))

	ctx := context.Background()
	req := &agentreq.ChatReq{
		AgentID:       "a1",
		InternalParam: agentreq.InternalParam{UserID: "u1"},
	}
	_, err := svc.Chat(ctx, req)
	assert.Error(t, err)
}

func TestAgentSvc_Chat_BackfillsConversationIDOnInvokeAgentSpan(t *testing.T) {
	recorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	oldTP := otel.GetTracerProvider()
	otel.SetTracerProvider(tp)
	t.Cleanup(func() {
		_ = tp.Shutdown(context.Background())

		otel.SetTracerProvider(oldTP)
	})

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSquare := v3portdrivermock.NewMockISquareSvc(ctrl)
	mockConvRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
	mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
	mockSessionSvc := iportdrivermock.NewMockISessionSvc(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)
	allowAnyLoggerCalls(mockLogger)

	svc := &agentSvc{
		SvcBase:             service.NewSvcBase(),
		squareSvc:           mockSquare,
		logger:              mockLogger,
		conversationRepo:    mockConvRepo,
		conversationMsgRepo: mockMsgRepo,
		sessionSvc:          mockSessionSvc,
		sandboxPlatformConf: &conf.SandboxPlatformConf{},
	}

	agentInfo := newTestAgent()
	mockSquare.EXPECT().GetAgentInfoByIDOrKey(gomock.Any(), gomock.Any()).Return(agentInfo, nil)
	mockConvRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&dapo.ConversationPO{ID: "conv-backfill"}, nil)
	gomock.InOrder(
		mockMsgRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return("user-msg-backfill", nil),
		mockConvRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil),
		mockMsgRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return("asst-msg-backfill", nil),
	)
	mockSessionSvc.EXPECT().HandleGetInfoOrCreate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(int64(0), 0, errors.New("session error"))

	req := &agentreq.ChatReq{
		AgentID:       "a1",
		InternalParam: agentreq.InternalParam{UserID: "u1"},
	}
	_, err := svc.Chat(context.Background(), req)
	assert.Error(t, err)

	invokeAgentSpan := findSpanByName(recorder.Ended(), "invoke_agent")
	require.NotNil(t, invokeAgentSpan)
	assert.Equal(t, "conv-backfill", readAttribute(invokeAgentSpan.Attributes(), otelconst.AttrGenAIConversationID))
}

func findSpanByName(spans []sdktrace.ReadOnlySpan, name string) sdktrace.ReadOnlySpan {
	for _, span := range spans {
		if span.Name() == name {
			return span
		}
	}

	return nil
}

func readAttribute(attrs []attribute.KeyValue, key string) string {
	for _, attr := range attrs {
		if string(attr.Key) == key {
			return attr.Value.AsString()
		}
	}

	return ""
}

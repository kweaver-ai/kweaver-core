package agentsvc

import (
	"context"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/constant"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/comvalobj"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/daconfvalobj"
	agentreq "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent/req"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/square/squareresp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp/cmpmock"
	"github.com/stretchr/testify/assert"
)

// newTestAgent 创建带有初始化 Config.Input 的 agent，避免 nil pointer panic
func newTestAgent() *squareresp.AgentMarketAgentInfoResp {
	agent := squareresp.NewAgentMarketAgentInfoResp()
	agent.Config = daconfvalobj.Config{
		Input: &daconfvalobj.Input{
			Fields: daconfvalobj.Fields{},
		},
	}

	return agent
}

func TestAgentSvc_GenerateAgentCallReq_Basic(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := cmpmock.NewMockLogger(ctrl)
	mockLogger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()

	svc := &agentSvc{
		SvcBase: service.NewSvcBase(),
		logger:  mockLogger,
	}

	ctx := context.Background()
	req := &agentreq.ChatReq{
		InternalParam: agentreq.InternalParam{UserID: "user-001"},
		AgentID:       "agent-001",
		AgentRunID:    "run-001",
		Query:         "hello",
	}
	agent := newTestAgent()

	result, err := svc.GenerateAgentCallReq(ctx, req, nil, agent)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "agent-001", result.ID)
	assert.Equal(t, "user-001", result.UserID)
	assert.Equal(t, "hello", result.Input["query"])
	assert.Equal(t, constant.NormalMode, req.ChatMode)
}

func TestAgentSvc_GenerateAgentCallReq_DeepThinkingMode(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := cmpmock.NewMockLogger(ctrl)
	mockLogger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()

	svc := &agentSvc{
		SvcBase: service.NewSvcBase(),
		logger:  mockLogger,
	}

	ctx := context.Background()
	req := &agentreq.ChatReq{
		InternalParam: agentreq.InternalParam{UserID: "user-001"},
		AgentID:       "agent-001",
		Query:         "deep question",
		ChatMode:      constant.DeepThinkingMode,
	}
	agent := newTestAgent()

	result, err := svc.GenerateAgentCallReq(ctx, req, nil, agent)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, constant.DeepThinkingMode, req.ChatMode)
}

func TestAgentSvc_GenerateAgentCallReq_WithHistory(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := cmpmock.NewMockLogger(ctrl)
	mockLogger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()

	svc := &agentSvc{
		SvcBase: service.NewSvcBase(),
		logger:  mockLogger,
	}

	ctx := context.Background()
	history := []*comvalobj.LLMMessage{
		{Role: "user", Content: "prev question"},
		{Role: "assistant", Content: "prev answer"},
	}
	req := &agentreq.ChatReq{
		InternalParam: agentreq.InternalParam{UserID: "user-001"},
		AgentID:       "agent-001",
		Query:         "follow up",
		History:       history,
	}
	agent := newTestAgent()

	result, err := svc.GenerateAgentCallReq(ctx, req, nil, agent)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, history, result.Input["history"])
}

func TestAgentSvc_GenerateAgentCallReq_WithNilContexts(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := cmpmock.NewMockLogger(ctrl)
	mockLogger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()

	svc := &agentSvc{
		SvcBase: service.NewSvcBase(),
		logger:  mockLogger,
	}

	ctx := context.Background()
	req := &agentreq.ChatReq{
		InternalParam: agentreq.InternalParam{UserID: "user-001"},
		AgentID:       "agent-001",
		Query:         "test",
	}
	agent := newTestAgent()

	result, err := svc.GenerateAgentCallReq(ctx, req, nil, agent)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.Input["history"])
}

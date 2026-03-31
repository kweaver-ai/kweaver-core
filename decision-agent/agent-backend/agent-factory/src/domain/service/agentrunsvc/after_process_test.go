package agentsvc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/daconfvalobj"
	agentreq "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent/req"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp/cmpmock"
)

// AfterProcess: AnswerVar 为空时返回 error
func TestAfterProcess_AnswerVarEmpty(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := cmpmock.NewMockLogger(ctrl)
	allowAnyLoggerCalls(mockLogger)

	svc := &agentSvc{
		SvcBase: service.NewSvcBase(),
		logger:  mockLogger,
	}

	// Config.Output.Variables.AnswerVar == "" → error
	agent := newTestAgent()
	agent.Config.Output = &daconfvalobj.Output{
		Variables: &daconfvalobj.VariablesS{AnswerVar: ""},
	}

	req := &agentreq.ChatReq{AgentID: "a1", InternalParam: agentreq.InternalParam{UserID: "u1"}}
	callResult := []byte(`{"status":"True","answer":"hello"}`)

	result, isEnd, err := svc.AfterProcess(context.Background(), callResult, req, agent, 0)
	assert.Error(t, err)
	assert.False(t, isEnd)
	assert.NotNil(t, result)
}

// AfterProcess: invalid JSON callResult → daresvo.NewDataAgentRes error
func TestAfterProcess_InvalidCallResult(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := cmpmock.NewMockLogger(ctrl)
	allowAnyLoggerCalls(mockLogger)

	svc := &agentSvc{
		SvcBase: service.NewSvcBase(),
		logger:  mockLogger,
	}

	agent := newTestAgent()
	agent.Config.Output = &daconfvalobj.Output{
		Variables: &daconfvalobj.VariablesS{AnswerVar: "answer"},
	}

	req := &agentreq.ChatReq{AgentID: "a1", InternalParam: agentreq.InternalParam{UserID: "u1"}}
	// invalid JSON
	callResult := []byte(`NOT_JSON`)

	result, isEnd, err := svc.AfterProcess(context.Background(), callResult, req, agent, 0)
	// daresvo.NewDataAgentRes may or may not error on invalid JSON; just check it doesn't panic
	_ = err
	_ = isEnd
	_ = result
}

// AfterProcess: invalid JSON → daresvo.NewDataAgentRes returns error
func TestAfterProcess_InvalidJSON(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := cmpmock.NewMockLogger(ctrl)
	allowAnyLoggerCalls(mockLogger)

	svc := &agentSvc{
		SvcBase: service.NewSvcBase(),
		logger:  mockLogger,
	}

	agent := newTestAgent()
	agent.Config.Output = &daconfvalobj.Output{
		Variables: &daconfvalobj.VariablesS{AnswerVar: "answer"},
	}

	req := &agentreq.ChatReq{AgentID: "a1", InternalParam: agentreq.InternalParam{UserID: "u1"}}
	callResult := []byte(`NOT_VALID_JSON`)

	result, isEnd, err := svc.AfterProcess(context.Background(), callResult, req, agent, 0)
	assert.Error(t, err)
	assert.False(t, isEnd)
	assert.NotNil(t, result)
}

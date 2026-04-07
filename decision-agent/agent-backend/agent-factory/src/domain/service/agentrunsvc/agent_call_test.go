package agentsvc

import (
	"context"
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/agentexecutoraccess/agentexecutordto"
	"github.com/stretchr/testify/assert"
)

func TestAgentCall_Call_UnsupportedExecutorVersion(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	req := &agentexecutordto.AgentCallReq{
		ExecutorVersion: "v3", // unsupported version
	}

	call := &AgentCall{
		callCtx: ctx,
		req:     req,
	}

	dataChan, errChan, err := call.Call()

	assert.Nil(t, dataChan)
	assert.Nil(t, errChan)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "executor version v3 not supported")
}

func TestAgentCall_Cancel_CallsCancelFunc(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())

	req := &agentexecutordto.AgentCallReq{}

	call := &AgentCall{
		callCtx:    ctx,
		req:        req,
		cancelFunc: cancel,
	}

	// Cancel should not panic
	call.Cancel()
}

func TestAgentCall_Resume_PanicsWithoutAgentExecutorV2(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	req := &agentexecutordto.AgentCallReq{}

	call := &AgentCall{
		callCtx: ctx,
		req:     req,
		// agentExecutorV2 is nil
	}

	assert.Panics(t, func() {
		_, _, _ = call.Resume("agent-run-123", nil)
	})
}

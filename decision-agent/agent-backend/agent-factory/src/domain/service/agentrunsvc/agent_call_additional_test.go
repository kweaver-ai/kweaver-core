package agentsvc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	agentexecutoraccreq "github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/httpaccess/agentexecutoraccess/agentexecutoraccreq"
	agentexecutoraccres "github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/httpaccess/agentexecutoraccess/agentexecutoraccres"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/httpaccess/agentexecutoraccess/agentexecutordto"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/httpaccess/v2agentexecutoraccess/v2agentexecutordto"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/ctype"
)

// ---------- minimal hand-written mocks ----------

type mockV1Executor struct {
	callFn func(ctx context.Context, req *agentexecutordto.AgentCallReq) (chan string, chan error, error)
}

func (m *mockV1Executor) Call(ctx context.Context, req *agentexecutordto.AgentCallReq) (chan string, chan error, error) {
	if m.callFn != nil {
		return m.callFn(ctx, req)
	}

	ch := make(chan string)
	close(ch)

	return ch, make(chan error), nil
}

func (m *mockV1Executor) AgentCacheManage(ctx context.Context, req *agentexecutoraccreq.AgentCacheManageReq, visitorInfo *ctype.VisitorInfo) (agentexecutoraccres.AgentCacheManageResp, error) {
	return agentexecutoraccres.AgentCacheManageResp{}, nil
}

type mockV2Executor struct {
	callFn func(ctx context.Context, req *v2agentexecutordto.V2AgentCallReq) (chan string, chan error, error)
}

func (m *mockV2Executor) Call(ctx context.Context, req *v2agentexecutordto.V2AgentCallReq) (chan string, chan error, error) {
	if m.callFn != nil {
		return m.callFn(ctx, req)
	}

	ch := make(chan string)
	close(ch)

	return ch, make(chan error), nil
}

func (m *mockV2Executor) Resume(ctx context.Context, req *v2agentexecutordto.AgentResumeReq) (chan string, chan error, error) {
	ch := make(chan string)
	close(ch)

	return ch, make(chan error), nil
}

func (m *mockV2Executor) Terminate(ctx context.Context, req *v2agentexecutordto.AgentTerminateReq) error {
	return nil
}

// ---------- AgentCall.Call tests ----------

func TestAgentCall_Call_V1Executor(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	called := false
	v1 := &mockV1Executor{
		callFn: func(ctx context.Context, req *agentexecutordto.AgentCallReq) (chan string, chan error, error) {
			called = true
			ch := make(chan string)
			close(ch)
			return ch, make(chan error), nil
		},
	}

	req := &agentexecutordto.AgentCallReq{ExecutorVersion: "v1"}
	call := &AgentCall{callCtx: ctx, req: req, agentExecutorV1: v1}

	dataChan, errChan, err := call.Call()
	assert.NoError(t, err)
	assert.NotNil(t, dataChan)
	assert.NotNil(t, errChan)
	assert.True(t, called)
}

func TestAgentCall_Call_V2Executor(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	called := false
	v2 := &mockV2Executor{
		callFn: func(ctx context.Context, req *v2agentexecutordto.V2AgentCallReq) (chan string, chan error, error) {
			called = true
			ch := make(chan string)
			close(ch)
			return ch, make(chan error), nil
		},
	}

	req := &agentexecutordto.AgentCallReq{ExecutorVersion: "v2"}
	call := &AgentCall{callCtx: ctx, req: req, agentExecutorV2: v2}

	dataChan, errChan, err := call.Call()
	assert.NoError(t, err)
	assert.NotNil(t, dataChan)
	assert.NotNil(t, errChan)
	assert.True(t, called)
}

func TestAgentCall_Call_V2WithResumeInfo(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	v2 := &mockV2Executor{}
	resumeInfo := &v2agentexecutordto.AgentResumeInfo{Action: "confirm"}
	req := &agentexecutordto.AgentCallReq{
		ExecutorVersion:     "v2",
		ResumeInterruptInfo: resumeInfo,
	}
	call := &AgentCall{callCtx: ctx, req: req, agentExecutorV2: v2}

	dataChan, _, err := call.Call()
	assert.NoError(t, err)
	assert.NotNil(t, dataChan)
}

func TestAgentCall_Call_V1NilExecutor(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	req := &agentexecutordto.AgentCallReq{ExecutorVersion: "v1"}
	// agentExecutorV1 is nil
	call := &AgentCall{callCtx: ctx, req: req}

	_, _, err := call.Call()
	assert.Error(t, err)
}

func TestAgentCall_Call_V2NilExecutor(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	req := &agentexecutordto.AgentCallReq{ExecutorVersion: "v2"}
	// agentExecutorV2 is nil
	call := &AgentCall{callCtx: ctx, req: req}

	_, _, err := call.Call()
	assert.Error(t, err)
}

func TestAgentCall_Resume_WithV2Executor(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	v2 := &mockV2Executor{}
	req := &agentexecutordto.AgentCallReq{}
	call := &AgentCall{callCtx: ctx, req: req, agentExecutorV2: v2}

	dataChan, _, err := call.Resume("run-1", &v2agentexecutordto.AgentResumeInfo{Action: "confirm"})
	assert.NoError(t, err)
	assert.NotNil(t, dataChan)
}

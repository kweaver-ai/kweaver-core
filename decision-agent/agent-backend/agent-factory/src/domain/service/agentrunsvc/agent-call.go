package agentsvc

import (
	"context"
	"fmt"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/agentexecutoraccess/agentexecutordto"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/v2agentexecutoraccess"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/v2agentexecutoraccess/v2agentexecutordto"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/ihttpaccess/iagentexecutorhttp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/ihttpaccess/iv2agentexecutorhttp"
)

type AgentCall struct {
	callCtx         context.Context
	req             *agentexecutordto.AgentCallReq
	agentExecutorV1 iagentexecutorhttp.IAgentExecutor
	agentExecutorV2 iv2agentexecutorhttp.IV2AgentExecutor
	cancelFunc      context.CancelFunc
}

func (a *AgentCall) Call() (chan string, chan error, error) {
	if a.req.ExecutorVersion == "v2" && a.agentExecutorV2 != nil {
		v2Req := v2agentexecutoraccess.ConvertV1ToV2CallReq(a.req)

		// 如果有 resume 信息，添加到 _options 中（统一 Run 接口支持恢复执行）
		if a.req.ResumeInterruptInfo != nil {
			v2Req.AgentOptions.ResumeInfo = a.req.ResumeInterruptInfo
		}

		return a.agentExecutorV2.Call(a.callCtx, v2Req)
	}

	if a.req.ExecutorVersion == "v1" && a.agentExecutorV1 != nil {
		return a.agentExecutorV1.Call(a.callCtx, a.req)
	}

	return nil, nil, fmt.Errorf("executor version %s not supported", a.req.ExecutorVersion)
}

func (a *AgentCall) Resume(agentRunID string, resumeInfo *v2agentexecutordto.AgentResumeInfo) (messageChan chan string, errChan chan error, err error) {
	// 构造V2 Resume请求（直接使用，不需要转换）
	v2Req := &v2agentexecutordto.AgentResumeReq{
		AgentRunID: agentRunID,
		ResumeInfo: resumeInfo,
	}

	// 调用executor的Resume接口
	return a.agentExecutorV2.Resume(a.callCtx, v2Req)
}

func (a *AgentCall) Cancel() {
	a.cancelFunc()
}

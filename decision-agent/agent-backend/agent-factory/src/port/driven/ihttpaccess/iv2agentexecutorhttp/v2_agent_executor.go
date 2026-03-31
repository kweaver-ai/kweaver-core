package iv2agentexecutorhttp

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/v2agentexecutoraccess/v2agentexecutordto"
)

// IV2AgentExecutor v2 版本的 Agent Executor 接口
// 注意：ConversationSessionInit 只有 v1 接口，不在此接口中
type IV2AgentExecutor interface {
	// Call 调用 Agent 执行
	Call(ctx context.Context, req *v2agentexecutordto.V2AgentCallReq) (chan string, chan error, error)
	// Resume 恢复 Agent 执行（中断后恢复）
	Resume(ctx context.Context, req *v2agentexecutordto.AgentResumeReq) (chan string, chan error, error)
	// Terminate 终止 Agent 执行
	Terminate(ctx context.Context, req *v2agentexecutordto.AgentTerminateReq) error
}

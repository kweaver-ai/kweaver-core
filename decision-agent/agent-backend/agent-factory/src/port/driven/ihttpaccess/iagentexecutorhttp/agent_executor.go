package iagentexecutorhttp

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/agentexecutoraccess/agentexecutoraccreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/agentexecutoraccess/agentexecutoraccres"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/agentexecutoraccess/agentexecutordto"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/ctype"
)

type IAgentExecutor interface {
	Call(ctx context.Context, req *agentexecutordto.AgentCallReq) (chan string, chan error, error)
	// ConversationSessionInit(ctx context.Context, req *agentexecutoraccreq.ConversationSessionInitReq, visitorInfo *ctype.VisitorInfo) (resp agentexecutoraccres.ConversationSessionInitResp, err error)
	AgentCacheManage(ctx context.Context, req *agentexecutoraccreq.AgentCacheManageReq, visitorInfo *ctype.VisitorInfo) (resp agentexecutoraccres.AgentCacheManageResp, err error)
}

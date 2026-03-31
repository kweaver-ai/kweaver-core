package v2agentexecutoraccess

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/agentexecutoraccess/agentexecutordto"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/v2agentexecutoraccess/v2agentexecutordto"
)

// ConvertV1ToV2CallReq 将 v1 的 AgentCallReq 转换为 v2 的 V2AgentCallReq
func ConvertV1ToV2CallReq(v1Req *agentexecutordto.AgentCallReq) *v2agentexecutordto.V2AgentCallReq {
	v2Req := &v2agentexecutordto.V2AgentCallReq{
		AgentID:      v1Req.ID,
		AgentVersion: v1Req.AgentVersion,
		AgentConfig: v2agentexecutordto.Config{
			Config: v1Req.Config.Config,
		},
		AgentInput: v1Req.Input,
		UserID:     v1Req.UserID,
		Token:      v1Req.Token,
		CallType:   v1Req.CallType,
		AgentOptions: v2agentexecutordto.AgentOptions{
			ConversationID:        v1Req.Config.ConversationID,
			AgentRunID:            v1Req.Config.SessionID,
			IsNeedProgress:        v1Req.ChatOption.IsNeedProgress,
			EnableDependencyCache: v1Req.ChatOption.EnableDependencyCache,
		},
		VisitorType:       v1Req.VisitorType,
		XAccountID:        v1Req.XAccountID,
		XAccountType:      v1Req.XAccountType,
		XBusinessDomainID: v1Req.XBusinessDomainID,
	}

	return v2Req
}

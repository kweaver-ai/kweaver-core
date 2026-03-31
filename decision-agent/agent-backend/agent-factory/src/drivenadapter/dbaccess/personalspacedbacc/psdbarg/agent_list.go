package psdbarg

import "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/personal_space/personalspacereq"

type AgentListArg struct {
	ListReq                       *personalspacereq.AgentListReq
	CreatedBy                     string
	HasBuiltInAgentMgmtPermission bool
	AgentIDsByBizDomain           []string
}

func NewAgentListArg(listReq *personalspacereq.AgentListReq, createdBy string, hasBuiltInAgentMgmtPermission bool, agentIDsByBizDomain []string) *AgentListArg {
	return &AgentListArg{
		ListReq:                       listReq,
		CreatedBy:                     createdBy,
		HasBuiltInAgentMgmtPermission: hasBuiltInAgentMgmtPermission,
		AgentIDsByBizDomain:           agentIDsByBizDomain,
	}
}

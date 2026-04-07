package authzhttpreq

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdapmsenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
)

// ResourceFilterReq 资源过滤请求
type ResourceFilterReq struct {
	Method    string                `json:"method"`
	Accessor  *Accessor             `json:"accessor"`
	Resources []*Resource           `json:"resources"`
	Operation []cdapmsenum.Operator `json:"operation"`
}

func NewFilterCanUseAgentReq(uid string, agentIDs []string) *ResourceFilterReq {
	resources := []*Resource{}
	for _, agentID := range agentIDs {
		resources = append(resources, &Resource{
			ID:   agentID,
			Type: cdaenum.ResourceTypeDataAgent,
		})
	}

	return &ResourceFilterReq{
		Method: "GET",
		Accessor: &Accessor{
			ID:   uid,
			Type: cenum.PmsTargetObjTypeUser,
		},
		Resources: resources,
		Operation: []cdapmsenum.Operator{
			cdapmsenum.AgentUse,
		},
	}
}

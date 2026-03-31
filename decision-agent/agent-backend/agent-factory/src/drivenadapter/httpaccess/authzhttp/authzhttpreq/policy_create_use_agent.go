package authzhttpreq

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdapmsenum"
)

func NewGrantAgentUseReq(accessor *PolicyAccessor, agentID string, agentName string) *CreatePolicyReq {
	return &CreatePolicyReq{
		Accessor: accessor,
		Resource: &PolicyResource{
			ID:   agentID,
			Type: cdaenum.ResourceTypeDataAgent,
			Name: agentName,
		},
		Operation: &PolicyOperation{
			Allow: []PolicyOperationItem{
				{ID: cdapmsenum.AgentUse},
			},
			Deny: []PolicyOperationItem{},
		},
	}
}

func NewGrantAgentUseReqs(accessors []*PolicyAccessor, agentID string, agentName string) (reqs []*CreatePolicyReq, err error) {
	for _, accessor := range accessors {
		reqs = append(reqs, NewGrantAgentUseReq(accessor, agentID, agentName))
	}

	return
}

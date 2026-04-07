package authzhttpreq

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdapmsenum"
)

func NewDenyAgentUseReq(accessor *PolicyAccessor, agentID string, agentName string) *CreatePolicyReq {
	return &CreatePolicyReq{
		Accessor: accessor,
		Resource: &PolicyResource{
			ID:   agentID,
			Type: cdaenum.ResourceTypeDataAgent,
			Name: agentName,
		},
		Operation: &PolicyOperation{
			Allow: []PolicyOperationItem{},
			Deny: []PolicyOperationItem{
				{ID: cdapmsenum.AgentUse},
			},
		},
	}
}

func NewDenyAgentUseReqs(accessors []*PolicyAccessor, agentID string, agentName string) (reqs []*CreatePolicyReq, err error) {
	for _, accessor := range accessors {
		reqs = append(reqs, NewDenyAgentUseReq(accessor, agentID, agentName))
	}

	return
}

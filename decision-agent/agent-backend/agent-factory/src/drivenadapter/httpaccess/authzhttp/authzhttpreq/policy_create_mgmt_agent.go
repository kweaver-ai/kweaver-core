package authzhttpreq

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdapmsenum"
)

func NewGrantAgentMgmtReq(accessor *PolicyAccessor, agentID string, agentName string, operations []cdapmsenum.Operator) *CreatePolicyReq {
	allowOps := make([]PolicyOperationItem, 0)
	denyOps := make([]PolicyOperationItem, 0)

	for _, op := range operations {
		allowOps = append(allowOps, PolicyOperationItem{ID: op})
	}

	return &CreatePolicyReq{
		Accessor: accessor,
		Resource: &PolicyResource{
			ID:   agentID,
			Type: cdaenum.ResourceTypeDataAgent,
			Name: agentName,
		},
		Operation: &PolicyOperation{
			Allow: allowOps,
			Deny:  denyOps,
		},
	}
}

func NewGrantAgentMgmtReqs(accessors []*PolicyAccessor, agentID string, agentName string, operations []cdapmsenum.Operator) (reqs []*CreatePolicyReq, err error) {
	for _, accessor := range accessors {
		reqs = append(reqs, NewGrantAgentMgmtReq(accessor, agentID, agentName, operations))
	}

	return
}

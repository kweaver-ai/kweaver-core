package authzhttpreq

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdapmsenum"
)

func NewGrantAgentTplMgmtReq(accessor *PolicyAccessor, agentTplID string, agentTplName string, operations []cdapmsenum.Operator) *CreatePolicyReq {
	allowOps := make([]PolicyOperationItem, 0)
	denyOps := make([]PolicyOperationItem, 0)

	for _, op := range operations {
		allowOps = append(allowOps, PolicyOperationItem{ID: op})
	}

	return &CreatePolicyReq{
		Accessor: accessor,
		Resource: &PolicyResource{
			ID:   agentTplID,
			Type: cdaenum.ResourceTypeDataAgentTpl,
			Name: agentTplName,
		},
		Operation: &PolicyOperation{
			Allow: allowOps,
			Deny:  denyOps,
		},
	}
}

func NewGrantAgentTplMgmtReqs(accessors []*PolicyAccessor, agentTplID string, agentTplName string, operations []cdapmsenum.Operator) (reqs []*CreatePolicyReq, err error) {
	for _, accessor := range accessors {
		reqs = append(reqs, NewGrantAgentTplMgmtReq(accessor, agentTplID, agentTplName, operations))
	}

	return
}

package authzhttpreq

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdapmsenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
)

func NewSingleAppAccountCheckReq(appAccountID string, resourceID string, resourceType cdaenum.ResourceType, operation []cdapmsenum.Operator) *SingleCheckReq {
	return &SingleCheckReq{
		Method:    "GET",
		Accessor:  &Accessor{ID: appAccountID, Type: cenum.PmsTargetObjTypeAppAccount},
		Resource:  &Resource{ID: resourceID, Type: resourceType},
		Operation: operation,
	}
}

func NewSingleAppAccountAgentUseCheckReq(appAccountID string, agentID string) *SingleCheckReq {
	return NewSingleAppAccountCheckReq(appAccountID, agentID, cdaenum.ResourceTypeDataAgent, []cdapmsenum.Operator{cdapmsenum.AgentUse})
}

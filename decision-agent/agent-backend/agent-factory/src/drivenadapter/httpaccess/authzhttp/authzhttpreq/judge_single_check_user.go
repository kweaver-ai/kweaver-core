package authzhttpreq

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdapmsenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
)

func NewSingleUserCheckReq(userID string, resourceID string, resourceType cdaenum.ResourceType, operation []cdapmsenum.Operator) *SingleCheckReq {
	return &SingleCheckReq{
		Method:    "GET",
		Accessor:  &Accessor{ID: userID, Type: cenum.PmsTargetObjTypeUser},
		Resource:  &Resource{ID: resourceID, Type: resourceType},
		Operation: operation,
	}
}

func NewSingleUserAgentUseCheckReq(userID string, agentID string) *SingleCheckReq {
	return NewSingleUserCheckReq(userID, agentID, cdaenum.ResourceTypeDataAgent, []cdapmsenum.Operator{cdapmsenum.AgentUse})
}

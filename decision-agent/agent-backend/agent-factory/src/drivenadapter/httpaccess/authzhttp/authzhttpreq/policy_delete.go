package authzhttpreq

import "github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"

// PolicyDeleteParams 策略删除参数
type PolicyDeleteParams struct {
	Method    string                  `json:"method"`
	Resources []*PolicyDeleteResource `json:"resources"`
}

// PolicyDeleteResource 策略删除资源
type PolicyDeleteResource struct {
	ID   string               `json:"id"`
	Type cdaenum.ResourceType `json:"type"`
}

func NewPolicyAgentDeleteReq(agentID string) *PolicyDeleteParams {
	r := &PolicyDeleteResource{
		ID:   agentID,
		Type: cdaenum.ResourceTypeDataAgent,
	}

	return &PolicyDeleteParams{
		Method:    "DELETE",
		Resources: []*PolicyDeleteResource{r},
	}
}

package authzhttpreq

import (
	"fmt"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
)

// ListPolicyReq 查询策略列表请求
type ListPolicyReq struct {
	Limit        int                  `json:"limit"`
	Offset       int                  `json:"offset"`
	ResourceID   string               `json:"resource_id"`
	ResourceType cdaenum.ResourceType `json:"resource_type"`
}

func NewListPolicyReq(resourceID string, resourceType cdaenum.ResourceType) *ListPolicyReq {
	return &ListPolicyReq{
		Limit:        1000,
		Offset:       0,
		ResourceID:   resourceID,
		ResourceType: resourceType,
	}
}

func (req *ListPolicyReq) ToReqQuery() string {
	return fmt.Sprintf("limit=%d&offset=%d&resource_id=%s&resource_type=%s", req.Limit, req.Offset, req.ResourceID, req.ResourceType)
}

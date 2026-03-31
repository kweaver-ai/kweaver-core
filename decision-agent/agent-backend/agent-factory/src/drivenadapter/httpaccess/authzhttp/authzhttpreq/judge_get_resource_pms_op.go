package authzhttpreq

import "github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"

// ResourceOperationReq 获取资源操作请求
type ResourceOperationReq struct {
	Method    string      `json:"method"`
	Accessor  *Accessor   `json:"accessor"`
	Resources []*Resource `json:"resources"`
}

func NewResourceOperationReq(accessor *Accessor, resources []*Resource) *ResourceOperationReq {
	return &ResourceOperationReq{
		Method:    "GET",
		Accessor:  accessor,
		Resources: resources,
	}
}

func NewResourceOperationReqSingle(accessor *Accessor, resource *Resource) *ResourceOperationReq {
	return NewResourceOperationReq(accessor, []*Resource{resource})
}

func NewResourceOperationReqSingleByUid(uid string, resource *Resource) *ResourceOperationReq {
	accessor := &Accessor{ID: uid, Type: cenum.PmsTargetObjTypeUser}

	return NewResourceOperationReqSingle(accessor, resource)
}

package authzhttpreq

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdapmsenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
)

// ResourceListReq 资源列举请求
type ResourceListReq struct {
	Method    string                `json:"method"`
	Accessor  *Accessor             `json:"accessor"`
	Resource  *ResourceListInfo     `json:"resource"`
	Operation []cdapmsenum.Operator `json:"operation"`
}

// ResourceListInfo 资源列举信息
type ResourceListInfo struct {
	Type cdaenum.ResourceType `json:"type"`
}

func NewCanUseAgentListReqByAccessor(accessor *Accessor) *ResourceListReq {
	return &ResourceListReq{
		Method:   "GET",
		Accessor: accessor,
		Resource: &ResourceListInfo{
			Type: cdaenum.ResourceTypeDataAgent,
		},
		Operation: []cdapmsenum.Operator{
			cdapmsenum.AgentUse,
		},
	}
}

func NewCanUseAgentListReqByUid(uid string) *ResourceListReq {
	userAccessor := &Accessor{
		ID:   uid,
		Type: cenum.PmsTargetObjTypeUser,
	}

	return NewCanUseAgentListReqByAccessor(userAccessor)
}

//func NewCanUseAgentListReqByStar() *ResourceListReq {
//	starAccessor := &Accessor{
//		ID: "*",
//		Type: cenum.PmsTargetObjTypeAppAccount,
//	}
//
//	return NewCanUseAgentListReqByAccessor(starAccessor)
//}

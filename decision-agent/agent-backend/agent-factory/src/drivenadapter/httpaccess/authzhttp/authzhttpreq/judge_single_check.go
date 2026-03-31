package authzhttpreq

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdapmsenum"
)

// SingleCheckReq 单个决策请求
type SingleCheckReq struct {
	Method    string                `json:"method"`
	Accessor  *Accessor             `json:"accessor"`
	Resource  *Resource             `json:"resource"`
	Operation []cdapmsenum.Operator `json:"operation"`
}

func NewSingleCheckReq(accessor *Accessor, resource *Resource, operation []cdapmsenum.Operator) *SingleCheckReq {
	return &SingleCheckReq{
		Method:    "GET",
		Accessor:  accessor,
		Resource:  resource,
		Operation: operation,
	}
}

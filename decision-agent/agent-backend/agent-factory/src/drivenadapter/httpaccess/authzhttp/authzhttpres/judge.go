package authzhttpres

import "github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdapmsenum"

// SingleCheckResult 单个决策结果
type SingleCheckResult struct {
	Result bool `json:"result"`
}

// ResourceListItem 资源列表项
type ResourceListItem struct {
	ID string `json:"id"`
}

// ResourceOperationItem 资源操作项
type ResourceOperationItem struct {
	ID        string                `json:"id"`        // 资源ID
	Operation []cdapmsenum.Operator `json:"operation"` // 操作列表
}

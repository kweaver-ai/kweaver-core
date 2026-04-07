package authzhttpreq

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdapmsenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
)

// CreatePolicyReq 新建策略请求
type CreatePolicyReq struct {
	Accessor  *PolicyAccessor  `json:"accessor"`
	Resource  *PolicyResource  `json:"resource"`
	Operation *PolicyOperation `json:"operation"`
	Condition string           `json:"condition,omitempty"`
	ExpiresAt string           `json:"expires_at,omitempty"`
}

// PolicyAccessor 策略访问者
type PolicyAccessor struct {
	ID   string                 `json:"id"`
	Type cenum.PmsTargetObjType `json:"type"`
	// Name string `json:"name"`
}

// PolicyResource 策略资源
type PolicyResource struct {
	ID   string               `json:"id"`
	Type cdaenum.ResourceType `json:"type"`
	Name string               `json:"name"`
}

// PolicyOperation 策略操作
type PolicyOperation struct {
	Allow []PolicyOperationItem `json:"allow"`
	Deny  []PolicyOperationItem `json:"deny"`
}

// PolicyOperationItem 策略操作项
type PolicyOperationItem struct {
	ID cdapmsenum.Operator `json:"id"`
}

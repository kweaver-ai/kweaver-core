package authzhttpres

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdapmsenum"
)

// PolicyOperation 策略操作
type PolicyOperation struct {
	Allow []*PolicyOperationItem `json:"allow"`
	Deny  []*PolicyOperationItem `json:"deny"`
}

// PolicyOperationItem 策略操作项
type PolicyOperationItem struct {
	ID   cdapmsenum.Operator `json:"id"`
	Name string              `json:"name"`
}

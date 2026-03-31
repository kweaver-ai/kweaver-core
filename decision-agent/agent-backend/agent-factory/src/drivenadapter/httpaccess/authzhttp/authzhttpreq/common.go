package authzhttpreq

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
)

// Accessor 访问者
type Accessor struct {
	ID         string                 `json:"id"`
	Type       cenum.PmsTargetObjType `json:"type"`
	IP         string                 `json:"ip,omitempty"`
	ClientType string                 `json:"client_type,omitempty"`
}

// Resource 资源
type Resource struct {
	ID   string               `json:"id"`
	Type cdaenum.ResourceType `json:"type"`
	// IDPath string               `json:"id_path,omitempty"`
}

package bizdomainhttpreq

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
)

// QueryResourceAssociationsReq 关联关系查询请求
type QueryResourceAssociationsReq struct {
	BdID   string               `json:"bd_id"`            // 业务域ID
	ID     string               `json:"id,omitempty"`     // 资源ID
	Type   cdaenum.ResourceType `json:"type,omitempty"`   // 资源类型
	Limit  int                  `json:"limit,omitempty"`  // 限制数量
	Offset int                  `json:"offset,omitempty"` // 偏移量
}

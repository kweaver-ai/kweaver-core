package bizdomainhttpreq

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
)

// DisassociateResourceReq 资源取消关联请求
type DisassociateResourceReq struct {
	BdID string               `_query:"bd_id" validate:"required"` // 业务域ID
	ID   string               `_query:"id" validate:"required"`    // 资源ID
	Type cdaenum.ResourceType `_query:"type" validate:"required"`  // 资源类型
}

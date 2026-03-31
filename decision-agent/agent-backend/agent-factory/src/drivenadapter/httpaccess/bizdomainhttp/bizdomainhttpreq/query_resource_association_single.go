package bizdomainhttpreq

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
)

// QueryResourceAssociationSingleReq 查询单个资源关联关系请求
type QueryResourceAssociationSingleReq struct {
	BdID string               `json:"bd_id" validate:"required"` // 业务域ID
	ID   string               `json:"id" validate:"required"`    // 资源ID
	Type cdaenum.ResourceType `json:"type" validate:"required"`  // 资源类型
}

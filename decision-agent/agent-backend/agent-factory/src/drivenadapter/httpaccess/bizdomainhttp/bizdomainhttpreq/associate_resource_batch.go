package bizdomainhttpreq

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
)

// AssociateResourceBatchReq 批量资源关联请求
// 所有元素的 bd_id 必须一致
type AssociateResourceBatchReq []*AssociateResourceItem

// AssociateResourceItem 资源关联项
type AssociateResourceItem struct {
	BdID cenum.BizDomainID    `json:"bd_id" validate:"required"` // 业务域ID
	ID   string               `json:"id" validate:"required"`    // 资源ID
	Type cdaenum.ResourceType `json:"type" validate:"required"`  // 资源类型
}

func NewInitAllAgentToPublicBusinessDomainReq(agentIDs []string) (req AssociateResourceBatchReq) {
	req = make(AssociateResourceBatchReq, 0)
	for _, agentID := range agentIDs {
		req = append(req, &AssociateResourceItem{
			BdID: cenum.BizDomainPublic,
			ID:   agentID,
			Type: cdaenum.ResourceTypeDataAgent,
		})
	}

	return
}

func NewInitAllAgentTplToPublicBusinessDomainReq(agentTplIDs []string) (req AssociateResourceBatchReq) {
	req = make(AssociateResourceBatchReq, 0)
	for _, agentTplID := range agentTplIDs {
		req = append(req, &AssociateResourceItem{
			BdID: cenum.BizDomainPublic,
			ID:   agentTplID,
			Type: cdaenum.ResourceTypeDataAgentTpl,
		})
	}

	return
}

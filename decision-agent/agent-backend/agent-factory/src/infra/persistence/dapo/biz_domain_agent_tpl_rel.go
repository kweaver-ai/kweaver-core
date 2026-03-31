package dapo

// BizDomainAgentTplRelPo 业务域与agent模板关联表PO
type BizDomainAgentTplRelPo struct {
	ID          int64  `json:"id" db:"f_id"`
	BizDomainID string `json:"biz_domain_id" db:"f_biz_domain_id"`
	AgentTplID  int64  `json:"agent_tpl_id" db:"f_agent_tpl_id"`
	CreatedAt   int64  `json:"created_at" db:"f_created_at"`
}

func (p *BizDomainAgentTplRelPo) TableName() string {
	return "t_biz_domain_agent_tpl_rel"
}

package dapo

// BizDomainAgentRelPo 业务域与agent关联表PO
type BizDomainAgentRelPo struct {
	ID          int64  `json:"id" db:"f_id"`
	BizDomainID string `json:"biz_domain_id" db:"f_biz_domain_id"`
	AgentID     string `json:"agent_id" db:"f_agent_id"`
	CreatedAt   int64  `json:"created_at" db:"f_created_at"`
}

func (p *BizDomainAgentRelPo) TableName() string {
	return "t_biz_domain_agent_rel"
}

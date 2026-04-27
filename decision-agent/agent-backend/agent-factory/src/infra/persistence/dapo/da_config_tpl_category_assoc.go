package dapo

type PubTplCatAssocPo struct {
	ID             int64  `json:"id" db:"f_id"`
	PublishedTplID int64  `json:"published_tpl_id" db:"f_published_tpl_id"`
	CategoryID     string `json:"category_id" db:"f_category_id"`
}

func (p *PubTplCatAssocPo) TableName() string {
	return "t_data_agent_tpl_category_rel"
}

type DataAgentTplCategoryJoinPo struct {
	ID             int64  `json:"id" db:"f_id"`
	PublishedTplID int64  `json:"published_tpl_id" db:"f_published_tpl_id"`
	CategoryID     string `json:"category_id" db:"f_category_id"`
	CategoryName   string `json:"category_name" db:"f_category_name"`
}

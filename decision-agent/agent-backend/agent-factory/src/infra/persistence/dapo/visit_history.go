package dapo

type VisitHistoryPO struct {
	ID           string `db:"f_id"`
	AgentID      string `db:"f_agent_id"`
	AgentVersion string `db:"f_agent_version"`
	VisitCount   int    `db:"f_visit_count"`

	CustomSpaceID string `db:"f_custom_space_id"`

	CreateTime int64  `db:"f_create_time"`
	UpdateTime int64  `db:"f_update_time"`
	CreateBy   string `db:"f_create_by"`
	UpdateBy   string `db:"f_update_by"`
}

func (po *VisitHistoryPO) TableName() string {
	return "t_data_agent_visit_history"
}

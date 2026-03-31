package dapo

type ReleaseHistoryPO struct {
	ID           string `db:"f_id"`
	AgentID      string `db:"f_agent_id"`
	AgentConfig  string `db:"f_agent_config"`
	AgentVersion string `db:"f_agent_version"`
	AgentDesc    string `db:"f_agent_desc"`

	CreateTime int64  `db:"f_create_time"`
	UpdateTime int64  `db:"f_update_time"`
	CreateBy   string `db:"f_create_by"`
	UpdateBy   string `db:"f_update_by"`
}

func (po *ReleaseHistoryPO) TableName() string {
	return "t_data_agent_release_history"
}

package dapo

type CategoryPO struct {
	ID          string `db:"f_id"`
	Name        string `db:"f_name"`
	Description string `db:"f_description"`

	CreateTime int64  `db:"f_create_time"`
	UpdateTime int64  `db:"f_update_time"`
	CreateBy   string `db:"f_create_by"`
	UpdateBy   string `db:"f_update_by"`
}

func (po *CategoryPO) TableName() string {
	return "t_data_agent_release_category"
}

package dapo

type ReleaseCategoryRelPO struct {
	ID         string `json:"id" db:"f_id"`
	ReleaseID  string `json:"release_id" db:"f_release_id"`
	CategoryID string `json:"category_id" db:"f_category_id"`
}

func (po *ReleaseCategoryRelPO) TableName() string {
	return "t_data_agent_release_category_rel"
}

package dapo

// ProductPo 产品持久化对象
type ProductPo struct {
	ID      int64  `json:"id" db:"f_id"`           // 产品ID
	Name    string `json:"name" db:"f_name"`       // 产品名称
	Key     string `json:"key" db:"f_key"`         // 产品标识，唯一
	Profile string `json:"profile" db:"f_profile"` // 产品简介

	CreatedAt int64 `json:"created_at" db:"f_created_at"` // 创建时间（时间戳，单位：ms）
	UpdatedAt int64 `json:"updated_at" db:"f_updated_at"` // 更新时间（时间戳，单位：ms）

	CreatedBy string `json:"created_by" db:"f_created_by"` // 创建者ID
	UpdatedBy string `json:"updated_by" db:"f_updated_by"` // 更新者ID

	DeletedAt int64  `json:"deleted_at" db:"f_deleted_at"` // 删除时间（软删除）
	DeletedBy string `json:"deleted_by" db:"f_deleted_by"` // 删除者ID
}

func (p *ProductPo) TableName() string {
	return "t_product"
}

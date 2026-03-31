package model

import (
	"context"
	"database/sql"
)

//go:generate mockgen -source=category.go -destination=../../mocks/model_category.go -package=mocks

// CategoryDB 分类表对应的结构体
type CategoryDB struct {
	ID           int64  `json:"f_id" db:"f_id"`                       // id
	CategoryID   string `json:"f_category_id" db:"f_category_id"`     // 分类ID
	CategoryName string `json:"f_category_name" db:"f_category_name"` // 分类名称
	CreateUser   string `json:"f_create_user" db:"f_create_user"`     // 创建者
	CreateTime   int64  `json:"f_create_time" db:"f_create_time"`     // 创建时间
	UpdateUser   string `json:"f_update_user" db:"f_update_user"`     // 编辑者
	UpdateTime   int64  `json:"f_update_time" db:"f_update_time"`     // 编辑时间
}

// DBCategory 分类表数据库操作
type DBCategory interface {
	// Insert 插入分类
	Insert(ctx context.Context, tx *sql.Tx, category *CategoryDB) (categoryID string, err error)
	// UpdateByID 更新分类
	UpdateByID(ctx context.Context, tx *sql.Tx, category *CategoryDB) error
	// SelectList 查询分类列表
	SelectList(ctx context.Context, tx *sql.Tx) (categoryList []*CategoryDB, err error)
	// SelectListByCategoryIDOrName 根据分类ID或名称查询分类列表
	SelectListByCategoryIDOrName(ctx context.Context, tx *sql.Tx, categoryID string, categoryName string) (categoryList []*CategoryDB, err error)
	// SelectListByCategoryID 根据分类ID查询分类列表
	SelectListByCategoryID(ctx context.Context, tx *sql.Tx, categoryID string) (category *CategoryDB, err error)
	// DeleteByCategoryID 根据分类ID删除分类
	DeleteByCategoryID(ctx context.Context, tx *sql.Tx, categoryID string) error
}

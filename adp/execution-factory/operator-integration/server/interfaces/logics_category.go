package interfaces

import "context"

// BizCategory 业务分类
//
//go:generate mockgen -source=logics_category.go -destination=../mocks/category.go -package=mocks
type BizCategory string

func (c BizCategory) String() string {
	return string(c)
}

const (
	CategoryTypeOther  = BizCategory("other_category") // 其他分类
	CategoryTypeSystem = BizCategory("system")         // 系统内置分类
)

// CategoryInfo 分类信息
type CategoryInfo struct {
	CategoryType BizCategory `json:"category_type"`
	CategoryName string      `json:"name"` // (支持国际化)
}

// CreateCategoryReq 新增分类请求
type CreateCategoryReq struct {
	UserID       string      `header:"user_id"`
	CategoryType BizCategory `json:"category_type"`
	CategoryName string      `json:"name" validate:"required"`
}

// CreateCategoryResp 新增分类响应
type CreateCategoryResp struct {
	CategoryType BizCategory `json:"category_type"`
	CategoryName string      `json:"name"`
}

// UpdateCategoryReq 更新分类请求
type UpdateCategoryReq struct {
	UserID       string      `header:"user_id"`
	CategoryType BizCategory `uri:"category_type" validate:"required"`
	CategoryName string      `json:"name" validate:"required"`
}

// UpdateCategoryResp 更新分类响应
type UpdateCategoryResp struct {
	CategoryType BizCategory `json:"category_type"`
	CategoryName string      `json:"name"`
}

// DeleteCategoryReq 删除分类请求
type DeleteCategoryReq struct {
	UserID       string      `header:"user_id"`
	CategoryType BizCategory `uri:"category_type" validate:"required"`
}

// DeleteCategoryResp 删除分类响应
type DeleteCategoryResp struct {
	CategoryType BizCategory `json:"category_type"`
}

// CategoryManager 分类管理器
type CategoryManager interface {
	// 获取分类列表
	GetCategoryList(ctx context.Context) (categoryList []*CategoryInfo, err error)
	// 检查分类是否存在
	CheckCategory(category BizCategory) (isExist bool)
	// 获取分类名称
	GetCategoryName(ctx context.Context, category BizCategory) (categoryName string)
	// 更新分类
	UpdateCategory(ctx context.Context, req *UpdateCategoryReq) (resp *UpdateCategoryResp, err error)
	// 新增分类
	CreateCategory(ctx context.Context, req *CreateCategoryReq) (resp *CreateCategoryResp, err error)
	// 删除分类
	DeleteCategory(ctx context.Context, req *DeleteCategoryReq) (err error)
	// 批量新增分类
	BatchCreateCategory(ctx context.Context, req []*CreateCategoryReq) (resp []*CreateCategoryResp, err error)
}

package ormhelper

import "strings"

// SortOrder 排序方向枚举
type SortOrder string

const (
	SortOrderAsc  SortOrder = "ASC"
	SortOrderDesc SortOrder = "DESC"
)

// ToUpper 将排序方向转换为大写字符串
func (s SortOrder) ToUpper() SortOrder {
	return SortOrder(strings.ToUpper(string(s)))
}

// String 实现 Stringer 接口
func (s SortOrder) String() string {
	return string(s)
}

// IsValid 验证排序方向是否有效
func (s SortOrder) IsValid() bool {
	return s == SortOrderAsc || s == SortOrderDesc
}

// PaginationParams 分页参数
type PaginationParams struct {
	Page     int `json:"page" validate:"min=1"`              // 页码，从1开始
	PageSize int `json:"page_size" validate:"min=1,max=100"` // 每页数量
}

// SortField 排序字段
type SortField struct {
	Field string    `json:"field"` // 数据库字段名（调用方负责传入正确的字段名）
	Order SortOrder `json:"order"` // 排序方向
}

// SortParams 排序参数
type SortParams struct {
	Fields []SortField `json:"fields,omitempty"` // 支持多字段排序
}

// CursorParams 游标参数
type CursorParams struct {
	Field     string    `json:"field,omitempty"`     // 游标字段名（调用方负责传入正确的字段名）
	Value     any       `json:"value,omitempty"`     // 游标值
	Direction SortOrder `json:"direction,omitempty"` // 游标方向，默认 ASC
}

// QueryResult 通用查询结果
type QueryResult struct {
	Total      int64 `json:"total"`       // 总记录数
	Page       int   `json:"page"`        // 当前页码
	PageSize   int   `json:"page_size"`   // 每页数量
	TotalPages int   `json:"total_pages"` // 总页数
	HasNext    bool  `json:"has_next"`    // 是否有下一页
	HasPrev    bool  `json:"has_prev"`    // 是否有上一页
}

// CalculateQueryResult 计算查询结果的分页信息
func CalculateQueryResult(total int64, pagination *PaginationParams) *QueryResult {
	if pagination == nil || pagination.Page <= 0 || pagination.PageSize <= 0 {
		return &QueryResult{
			Total:      total,
			Page:       1,
			PageSize:   int(total),
			TotalPages: 1,
			HasNext:    false,
			HasPrev:    false,
		}
	}

	totalPages := int((total + int64(pagination.PageSize) - 1) / int64(pagination.PageSize))

	return &QueryResult{
		Total:      total,
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalPages: totalPages,
		HasNext:    pagination.Page < totalPages,
		HasPrev:    pagination.Page > 1,
	}
}

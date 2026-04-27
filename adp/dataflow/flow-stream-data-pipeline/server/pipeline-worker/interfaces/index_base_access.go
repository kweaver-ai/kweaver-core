package interfaces

import (
	"context"
)

const (
	CATEGORY_METRIC = "metric"
	CATEGORY_TRACE  = "trace"
)

// 索引库信息
type IndexBaseInfo struct {
	BaseType string `json:"base_type"`
	Name     string `json:"name"`
	DataType string `json:"data_type"`
	Category string `json:"category"`
}

//go:generate mockgen -source ../interfaces/index_base_access.go -destination ../interfaces/mock/mock_index_base_access.go
type IndexBaseAccess interface {
	GetIndexBasesByTypes(ctx context.Context, baseType []string) ([]*IndexBaseInfo, error)
}

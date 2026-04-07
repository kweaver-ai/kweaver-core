package interfaces

import "context"

const (
	CONTENT_TYPE_NAME = "Content-Type"
	CONTENT_TYPE_JSON = "application/json"
)

// 索引库信息
type IndexBase struct {
	BaseType string `json:"base_type"`
	Name     string `json:"name"`
	DataType string `json:"data_type"`
	Category string `json:"category"`
	// Indices  []string         `json:"indices"`
	// Mappings Mappings         `json:"mappings"`
	// Fields   []IndexBaseField `json:"-"`
}

// 索引库的字段信息
type Mappings struct {
	UserDefinedMappings []IndexBaseField `json:"user_defined_mappings"`
	MetaMappings        []IndexBaseField `json:"meta_mappings"`
	DynamicMappings     []IndexBaseField `json:"dynamic_mappings"`
}

type IndexBaseField struct {
	Field string `json:"field"`
	Type  string `json:"type"`
}

//go:generate mockgen -source ../interfaces/index_base_access.go -destination ../interfaces/mock/mock_index_base_access.go
type IndexBaseAccess interface {
	GetIndexBasesByTypes(ctx context.Context, baseTypes []string) ([]*IndexBase, error)
}

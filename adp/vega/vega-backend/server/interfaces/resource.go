// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

const (
	ResourceCategoryTable     string = "table"
	ResourceCategoryFile      string = "file"
	ResourceCategoryFileset   string = "fileset"
	ResourceCategoryAPI       string = "api"
	ResourceCategoryMetric    string = "metric"
	ResourceCategoryTopic     string = "topic"
	ResourceCategoryIndex     string = "index"
	ResourceCategoryLogicView string = "logicview"
	ResourceCategoryDataset   string = "dataset"
)

const (
	ResourceStatusActive     string = "active"
	ResourceStatusDisabled   string = "disabled"
	ResourceStatusDeprecated string = "deprecated"
	ResourceStatusStale      string = "stale"
)

var (
	RESOURCE_SORT = map[string]string{
		"name":        "f_name",
		"create_time": "f_create_time",
		"update_time": "f_update_time",
	}
)

// Resource represents a Data Resource entity.
type Resource struct {
	ID          string   `json:"id"`
	CatalogID   string   `json:"catalog_id"`
	Name        string   `json:"name"`
	Tags        []string `json:"tags"`
	Description string   `json:"description"`

	Category string `json:"category"` // 资源类别：table/file/fileset/...

	Status        string `json:"status"`         // 状态：active/stale/disabled
	StatusMessage string `json:"status_message"` // 状态消息

	// 新增字段：支持自动发现
	Database         string         `json:"database,omitempty"`          // 所属数据库（实例级 Catalog 时填充）
	SourceIdentifier string         `json:"source_identifier"`           // 源端标识（原始表名/路径）
	SourceMetadata   map[string]any `json:"source_metadata,omitempty"`   // 源端配置（JSON）
	SchemaDefinition []*Property    `json:"schema_definition,omitempty"` // Schema定义

	// 逻辑视图特有的字段
	LogicType       string                 `json:"logic_type,omitempty"`       // 逻辑类型: derived(衍生), composite(复合)
	LogicDefinition []*LogicDefinitionNode `json:"logic_definition,omitempty"` // 逻辑定义

	Creator    AccountInfo `json:"creator"`
	CreateTime int64       `json:"create_time"`
	Updater    AccountInfo `json:"updater"`
	UpdateTime int64       `json:"update_time"`

	Operations []string `json:"operations"`
}

type Property struct {
	Name         string            `json:"name"`
	Type         string            `json:"type"`
	OrigType     string            `json:"orig_type"`
	DisplayName  string            `json:"display_name"`
	OriginalName string            `json:"original_name"`
	Description  string            `json:"description"`
	Features     []PropertyFeature `json:"features"`
	Attributes   map[string]any    `json:"attributes"`
}

type PropertyFeature struct {
	FeatureName string         `json:"name"`
	DisplayName string         `json:"display_name"`
	FeatureType string         `json:"feature_type"` // 特性类型：keyword, fulltext, vector
	Description string         `json:"description"`
	RefProperty string         `json:"ref_property"`
	IsDefault   bool           `json:"is_default"`
	IsNative    bool           `json:"is_native"`
	Config      map[string]any `json:"config"`
}

// ResourcesQueryParams holds resource list query parameters.
type ResourcesQueryParams struct {
	PaginationQueryParams
	CatalogID string
	Category  string
	Status    string
	Database  string
}

// ResourceCreateRequest represents create resource request.
type ResourceRequest struct {
	ID          string   `json:"id,omitempty"`
	CatalogID   string   `json:"catalog_id"`
	Name        string   `json:"name"`
	Tags        []string `json:"tags"`
	Description string   `json:"description"`

	Category string `json:"category"`

	Status string `json:"status"`

	Database         string                 `json:"database,omitempty"`          // 所属数据库（实例级 Catalog 时填充）
	SourceIdentifier string                 `json:"source_identifier"`           // 源端标识（原始表名/路径）
	SourceMetadata   map[string]any         `json:"source_metadata,omitempty"`   // 源端配置（JSON）
	SchemaDefinition []*Property            `json:"schema_definition,omitempty"` // Schema定义
	LogicDefinition  []*LogicDefinitionNode `json:"logic_definition,omitempty"`  // 逻辑定义

	IfNameModify   bool      `json:"-"`
	OriginResource *Resource `json:"-"`
}

type ListResourcesQueryParams struct {
	PaginationQueryParams
	ID      string
	Keyword string
}

type ListResourceEntry struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Name string `json:"name"`
}

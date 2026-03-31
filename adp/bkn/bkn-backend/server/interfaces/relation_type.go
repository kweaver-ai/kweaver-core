// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

const (
	RELATION_TYPE_DIRECT    = "direct"
	RELATION_TYPE_DATA_VIEW = "data_view"
)

var (
	RELATION_TYPE_SORT = map[string]string{
		"name":        "f_name",
		"update_time": "f_update_time",
	}
)

type RelationTypeWithKeyField struct {
	RTID               string `json:"id" mapstructure:"id"`
	RTName             string `json:"name" mapstructure:"name"`
	SourceObjectTypeID string `json:"source_object_type_id" mapstructure:"source_object_type_id"`
	TargetObjectTypeID string `json:"target_object_type_id" mapstructure:"target_object_type_id"`

	Type         string `json:"type" mapstructure:"type"`
	MappingRules any    `json:"mapping_rules" mapstructure:"mapping_rules"` // 根据type来决定是不同的映射方式，direct对应的结构体是[]Mapping
}

// knowledge_network
type RelationType struct {
	RelationTypeWithKeyField `mapstructure:",squash"`
	CommonInfo               `mapstructure:",squash"`
	KNID                     string           `json:"kn_id" mapstructure:"kn_id"`
	Branch                   string           `json:"branch,omitempty" mapstructure:"branch"`
	SourceObjectType         SimpleObjectType `json:"source_object_type,omitempty" mapstructure:"source_object_type"` // 查看详情的时候给出名称
	TargetObjectType         SimpleObjectType `json:"target_object_type,omitempty" mapstructure:"target_object_type"` // 查看详情的时候给出名称

	Creator    AccountInfo `json:"creator" mapstructure:"creator"`
	CreateTime int64       `json:"create_time" mapstructure:"create_time"`
	Updater    AccountInfo `json:"updater" mapstructure:"updater"`
	UpdateTime int64       `json:"update_time" mapstructure:"update_time"`

	ModuleType string `json:"module_type" mapstructure:"module_type"`

	IfNameModify bool `json:"-"`

	// 向量
	Vector []float32 `json:"_vector,omitempty"`
	Score  *float64  `json:"_score,omitempty"` // opensearch检索的得分，在概念搜索时使用
}

// 非直接映射
type InDirectMapping struct {
	BackingDataSource  *ResourceInfo `json:"backing_data_source" mapstructure:"backing_data_source"`
	SourceMappingRules []Mapping     `json:"source_mapping_rules" mapstructure:"source_mapping_rules"`
	TargetMappingRules []Mapping     `json:"target_mapping_rules" mapstructure:"target_mapping_rules"`
}

// 直接映射的一个mapping
type Mapping struct {
	SourceProp SimpleProperty `json:"source_property" mapstructure:"source_property"`
	TargetProp SimpleProperty `json:"target_property" mapstructure:"target_property"`
}

// 对象类的分页查询
type RelationTypesQueryParams struct {
	PaginationQueryParameters
	NamePattern         string
	Tag                 string
	Branch              string
	KNID                string
	GroupID             string
	SourceObjectTypeIDs []string
	TargetObjectTypeIDs []string
	BoundObjectTypeIDs  []string
}

// 检索关系类列表
type RelationTypes struct {
	Entries     []*RelationType `json:"entries"`
	TotalCount  int64           `json:"total_count,omitempty"`
	SearchAfter []any           `json:"search_after,omitempty"`
	OverallMs   int64           `json:"overall_ms"`
}

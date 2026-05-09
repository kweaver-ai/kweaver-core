// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import "bkn-backend/interfaces/data_type"

const (
	MAX_PROPERTY_NUM = 1000
	// 逻辑属性类型
	LOGIC_PROPERTY_TYPE_METRIC   = "metric"
	LOGIC_PROPERTY_TYPE_OPERATOR = "operator"
)

var (
	OBJECT_TYPE_SORT = map[string]string{
		"name":        "f_name",
		"update_time": "f_update_time",
	}

	// 主键的属性的类型只能是: integer, unsigned integer, string, text
	ValidPrimaryKeyTypes = map[string]bool{
		data_type.DATATYPE_INTEGER:          true,
		data_type.DATATYPE_UNSIGNED_INTEGER: true,
		data_type.DATATYPE_STRING:           true,
		data_type.DATATYPE_TEXT:             true,
	}

	//  标题的属性的类型支持： integer, unsigned integer, float, decimal, string, text, date, timestamp, time, datetime, boolean
	ValidDisplayKeyTypes = map[string]bool{
		data_type.DATATYPE_INTEGER:          true,
		data_type.DATATYPE_UNSIGNED_INTEGER: true,
		data_type.DATATYPE_FLOAT:            true,
		data_type.DATATYPE_DECIMAL:          true,
		data_type.DATATYPE_STRING:           true,
		data_type.DATATYPE_TEXT:             true,
		data_type.DATATYPE_DATE:             true,
		data_type.DATATYPE_TIMESTAMP:        true,
		data_type.DATATYPE_TIME:             true,
		data_type.DATATYPE_DATETIME:         true,
		data_type.DATATYPE_BOOLEAN:          true,
	}

	// 逻辑资源类型需有效，当前支持 metric, operator
	ValidLogicSourceTypes = map[string]bool{
		LOGIC_PROPERTY_TYPE_METRIC:   true,
		LOGIC_PROPERTY_TYPE_OPERATOR: true,
	}

	// 有效的属性类型：integer, unsigned integer, float, decimal, string, text, date, timestamp, time, datetime, boolean, binary, json, vector, point, shape, ip。
	ValidDataPropertyTypes = map[string]bool{
		data_type.DATATYPE_INTEGER:          true,
		data_type.DATATYPE_UNSIGNED_INTEGER: true,
		data_type.DATATYPE_STRING:           true,
		data_type.DATATYPE_FLOAT:            true,
		data_type.DATATYPE_DECIMAL:          true,
		data_type.DATATYPE_TEXT:             true,
		data_type.DATATYPE_DATE:             true,
		data_type.DATATYPE_TIMESTAMP:        true,
		data_type.DATATYPE_TIME:             true,
		data_type.DATATYPE_DATETIME:         true,
		data_type.DATATYPE_BOOLEAN:          true,
		data_type.DATATYPE_BINARY:           true,
		data_type.DATATYPE_JSON:             true,
		data_type.DATATYPE_VECTOR:           true,
		data_type.DATATYPE_POINT:            true,
		data_type.DATATYPE_SHAPE:            true,
		data_type.DATATYPE_IP:               true,
	}
)

type ObjectTypeWithKeyField struct {
	OTID            string           `json:"id" mapstructure:"id"`
	OTName          string           `json:"name" mapstructure:"name"`
	DataSource      *ResourceInfo    `json:"data_source" mapstructure:"data_source"`
	DataProperties  []*DataProperty  `json:"data_properties,omitempty" mapstructure:"data_properties"`
	LogicProperties []*LogicProperty `json:"logic_properties,omitempty" mapstructure:"logic_properties"`
	PrimaryKeys     []string         `json:"primary_keys" mapstructure:"primary_keys"`
	DisplayKey      string           `json:"display_key" mapstructure:"display_key"`
	IncrementalKey  string           `json:"incremental_key" mapstructure:"incremental_key"`
	// ConditionOperations []string         `json:"condition_operations"`
}

// object type
type ObjectType struct {
	ObjectTypeWithKeyField `mapstructure:",squash"`
	CommonInfo             `mapstructure:",squash"`

	KNID          string          `json:"kn_id" mapstructure:"kn_id"`
	Branch        string          `json:"branch" mapstructure:"branch"`
	ConceptGroups []*ConceptGroup `json:"concept_groups,omitempty" mapstructure:"concept_groups"`

	Status *ObjectTypeStatus `json:"status,omitempty" mapstructure:"status"`

	Creator    AccountInfo `json:"creator" mapstructure:"creator"`
	CreateTime int64       `json:"create_time" mapstructure:"create_time"`
	Updater    AccountInfo `json:"updater" mapstructure:"updater"`
	UpdateTime int64       `json:"update_time" mapstructure:"update_time"`

	ModuleType string `json:"module_type" mapstructure:"module_type"`

	PropertyMap  map[string]string `json:"-"` // 以属性名为key，显示名为value的map
	IfNameModify bool              `json:"-"`

	// 向量
	Vector []float32 `json:"_vector,omitempty"`
	Score  *float64  `json:"_score,omitempty"` // opensearch检索的得分，在概念搜索时使用
}

type ObjectTypeStatus struct {
	IncrementalKey   string `json:"incremental_key" mapstructure:"incremental_key"`
	IncrementalValue string `json:"incremental_value" mapstructure:"incremental_value"`
	Index            string `json:"index" mapstructure:"index"`
	IndexAvailable   bool   `json:"index_available" mapstructure:"index_available"`
	DocCount         int64  `json:"doc_count" mapstructure:"doc_count"`
	StorageSize      int64  `json:"storage_size" mapstructure:"storage_size"`
	UpdateTime       int64  `json:"update_time" mapstructure:"update_time"`
}

type SimpleObjectType struct {
	OTID   string `json:"id" mapstructure:"id"`
	OTName string `json:"name" mapstructure:"name"`
	Branch string `json:"branch" mapstructure:"branch"`
	Icon   string `json:"icon" mapstructure:"icon"`
	Color  string `json:"color" mapstructure:"color"`
}

type DataProperty struct {
	Name        string `json:"name" mapstructure:"name"`
	DisplayName string `json:"display_name" mapstructure:"display_name"`
	Type        string `json:"type" mapstructure:"type"`
	Comment     string `json:"comment" mapstructure:"comment"`

	MappedField *Field `json:"mapped_field,omitempty" mapstructure:"mapped_field,omitempty"`

	IndexConfig *IndexConfig `json:"index_config,omitempty" mapstructure:"index_config,omitempty"`

	ConditionOperations []string `json:"condition_operations,omitempty"` // 字符串类型的字段支持的操作集
}

type LogicProperty struct {
	Name        string `json:"name" mapstructure:"name"`
	DisplayName string `json:"display_name" mapstructure:"display_name"`
	Type        string `json:"type" mapstructure:"type"`
	Comment     string `json:"comment" mapstructure:"comment"`
	// Index        bool          `json:"index" mapstructure:"index"`
	DataSource   *ResourceInfo `json:"data_source" mapstructure:"data_source"`
	Parameters   []Parameter   `json:"parameters" mapstructure:"parameters"`
	AnalysisDims []Field       `json:"analysis_dimensions,omitempty"`
}

type Field struct {
	Name        string  `json:"name" mapstructure:"name"`                                     // 技术名
	Type        string  `json:"type,omitempty" mapstructure:"type,omitempty"`                 // 字段类型
	DisplayName string  `json:"display_name,omitempty" mapstructure:"display_name,omitempty"` // 显示名
	Comment     *string `json:"comment,omitempty"`                                            // 从视图中获取到的字段的备注信息
}

type Parameter struct {
	Name        string  `json:"name" mapstructure:"name"`
	Type        string  `json:"type" mapstructure:"type"`                               // 参数类型
	Source      string  `json:"source,omitempty" mapstructure:"source,omitempty"`       // 来源类型
	Operation   string  `json:"operation,omitempty" mapstructure:"operation,omitempty"` // 指标属性的操作符。
	ValueFrom   string  `json:"value_from,omitempty" mapstructure:"value_from,omitempty"`
	Value       any     `json:"value,omitempty" mapstructure:"value,omitempty"`
	IfSystemGen *bool   `json:"if_system_generate,omitempty" mapstructure:"if_system_generate,omitempty"`
	Comment     *string `json:"comment,omitempty" mapstructure:"comment,omitempty"` // 参数备注，指标实时读取赋值
	Required    bool    `json:"required,omitempty" mapstructure:"required,omitempty"`
	Default     any     `json:"default,omitempty" mapstructure:"default,omitempty"`
}

type IndexConfig struct {
	KeywordConfig  KeywordConfig  `json:"keyword_config,omitempty" mapstructure:"keyword_config,omitempty"`
	FulltextConfig FulltextConfig `json:"fulltext_config,omitempty" mapstructure:"fulltext_config,omitempty"`
	VectorConfig   VectorConfig   `json:"vector_config,omitempty" mapstructure:"vector_config,omitempty"`
}

type KeywordConfig struct {
	Enabled        bool `json:"enabled" mapstructure:"enabled"`
	IgnoreAboveLen int  `json:"ignore_above_len" mapstructure:"ignore_above_len"`
}

type FulltextConfig struct {
	Enabled  bool   `json:"enabled" mapstructure:"enabled"`
	Analyzer string `json:"analyzer" mapstructure:"analyzer"`
}

type VectorConfig struct {
	Enabled bool   `json:"enabled" mapstructure:"enabled"`
	ModelID string `json:"model_id" mapstructure:"model_id"`

	//Model *SmallModel `json:"-"`
}

type SimpleProperty struct {
	Name        string `json:"name" mapstructure:"name"`
	DisplayName string `json:"display_name" mapstructure:"display_name"`
}

// 对象类的分页查询
type ObjectTypesQueryParams struct {
	PaginationQueryParameters
	NamePattern string
	Tag         string
	Branch      string
	KNID        string
	OTIDS       []string // 按对象类id过滤
}

// 检索的对象列表
type ObjectTypes struct {
	Entries     []*ObjectType `json:"entries"`
	TotalCount  int64         `json:"total_count,omitempty"`
	SearchAfter []any         `json:"search_after,omitempty"`
	OverallMs   int64         `json:"overall_ms"`
}

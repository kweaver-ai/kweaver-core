// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package condition

import (
	"encoding/json"
	"reflect"
)

// 字段范围
const (
	CUSTOM uint8 = iota
	ALL
)

const (
	DESENSITIZE_FIELD_SUFFIX = "_desensitize"

	AllField = "*"

	MetaField_ID = "__id"

	OS_MetaField_ID = "_id"

	ValueFrom_Const = "const"
	ValueFrom_Field = "field"
	ValueFrom_User  = "user"

	KNN_LIMIT_KEY_DEFAULT   = "k"
	KNN_LIMIT_VALUE_DEFAULT = 100
)

const (
	OperationAnd = "and"
	OperationOr  = "or"

	OperationEq          = "=="
	OperationNotEq       = "!="
	OperationGt          = ">"
	OperationGte         = ">="
	OperationLt          = "<"
	OperationLte         = "<="
	OperationIn          = "in"
	OperationNotIn       = "not_in"
	OperationLike        = "like"
	OperationNotLike     = "not_like"
	OperationContain     = "contain"
	OperationNotContain  = "not_contain"
	OperationRange       = "range"
	OperationOutRange    = "out_range"
	OperationExist       = "exist"
	OperationNotExist    = "not_exist"
	OperationEmpty       = "empty"
	OperationNotEmpty    = "not_empty"
	OperationRegex       = "regex"
	OperationMatch       = "match"
	OperationMatchPhrase = "match_phrase"
	OperationMultiMatch  = "multi_match"
	OperationKNN         = "knn"
	OperationKNNVector   = "knn_vector"
	OperationPrefix      = "prefix"
	OperationNotPrefix   = "not_prefix"
	OperationNull        = "null"
	OperationNotNull     = "not_null"
	OperationTrue        = "true"
	OperationFalse       = "false"
	OperationBefore      = "before"
	OperationCurrent     = "current"
	OperationBetween     = "between"
)

var (
	OperationMap = map[string]struct{}{
		"=":                  {}, // 兼容filter中定义的等于是 =
		OperationAnd:         {},
		OperationOr:          {},
		OperationEq:          {},
		OperationNotEq:       {},
		OperationGt:          {},
		OperationGte:         {},
		OperationLt:          {},
		OperationLte:         {},
		OperationIn:          {},
		OperationNotIn:       {},
		OperationLike:        {},
		OperationNotLike:     {},
		OperationContain:     {},
		OperationNotContain:  {},
		OperationRange:       {},
		OperationOutRange:    {},
		OperationExist:       {},
		OperationNotExist:    {},
		OperationEmpty:       {},
		OperationNotEmpty:    {},
		OperationRegex:       {},
		OperationMatch:       {},
		OperationMatchPhrase: {},
		OperationPrefix:      {},
		OperationNotPrefix:   {},
		OperationNull:        {},
		OperationNotNull:     {},
		OperationTrue:        {},
		OperationFalse:       {},
		OperationBefore:      {},
		OperationCurrent:     {},
		OperationBetween:     {},
		OperationKNN:         {},
		OperationMultiMatch:  {},
	}

	NotRequiredValueOperationMap = map[string]struct{}{
		OperationExist:    {},
		OperationNotExist: {},
		OperationEmpty:    {},
		OperationNotEmpty: {},
		OperationNull:     {},
		OperationNotNull:  {},
		OperationTrue:     {},
		OperationFalse:    {},
	}

	// match_type
	MatchTypeMap = map[string]bool{
		"best_fields":   true, // 哪个字段匹配最好就用那个字段的分数
		"most_fields":   true, // 多个字段的分数累加
		"cross_fields":  true, // 将多个字段视为一个大的组合字段来匹配
		"phrase":        true, // 要求查询词按完全相同的顺序出现在字段中
		"phrase_prefix": true, // 除最后一个词外按顺序匹配，最后一个词进行前缀匹配
		"bool_prefix":   true, // 将查询词作为前缀进行布尔查询,不要求词顺序紧密相邻
	}
)

type VectorResp struct {
	Object string    `json:"object"`
	Vector []float32 `json:"embedding"`
	Index  int       `json:"index"`
}

type Filter struct {
	Name      string `json:"name"`
	Operation string `json:"operation"`
	Value     any    `json:"value"`
}

type CondCfg struct {
	ObjectTypeID string     `json:"object_type_id,omitempty" mapstructure:"object_type_id"` // 行动条件需要标记是哪个行动类的
	Name         string     `json:"field,omitempty" mapstructure:"field"`
	Operation    string     `json:"operation,omitempty" mapstructure:"operation"`
	SubConds     []*CondCfg `json:"sub_conditions,omitempty" mapstructure:"sub_conditions"`
	ValueOptCfg  `mapstructure:",squash"`

	RemainCfg map[string]any `json:"-" mapstructure:",remain"`

	NameField *DataProperty `json:"-" mapstructure:"-"`
}

// MarshalJSON 自定义 JSON 序列化，将 RemainCfg 中的内容平铺到顶层
func (c *CondCfg) MarshalJSON() ([]byte, error) {
	// 创建一个临时结构体，用于序列化标准字段
	type Alias CondCfg
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(c),
	}

	// 先序列化标准字段
	data, err := json.Marshal(aux)
	if err != nil {
		return nil, err
	}

	// 解析为标准map
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	// 将 RemainCfg 中的内容平铺到顶层
	if c.RemainCfg != nil {
		for k, v := range c.RemainCfg {
			// 避免覆盖已存在的字段
			if _, exists := result[k]; !exists {
				result[k] = v
			}
		}
	}

	// 移除 nil 值
	for k, v := range result {
		if v == nil {
			delete(result, k)
		}
	}
	return json.Marshal(result)
}

type DataProperty struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Type        string `json:"type"`
	Comment     string `json:"comment"`

	MappedField Field `json:"mapped_field"`

	IndexConfig         *IndexConfig `json:"index_config,omitempty"`
	ConditionOperations []string     `json:"condition_operations,omitempty"` // 字符串类型的字段支持的操作集
}

type Field struct {
	Name        string `json:"name"` // 技术名
	Type        string `json:"type"`
	DisplayName string `json:"display_name"` // 显示名
}

type IndexConfig struct {
	KeywordConfig  KeywordConfig  `json:"keyword_config,omitempty"`
	FulltextConfig FulltextConfig `json:"fulltext_config,omitempty"`
	VectorConfig   VectorConfig   `json:"vector_config,omitempty"`
}

type KeywordConfig struct {
	Enabled        bool `json:"enabled"`
	IgnoreAboveLen int  `json:"ignore_above_len"`
}

type FulltextConfig struct {
	Enabled  bool   `json:"enabled"`
	Analyzer string `json:"analyzer"`
}

type VectorConfig struct {
	Enabled bool   `json:"enabled"`
	ModelID string `json:"model_id"`
}

// type ViewField struct {
// 	Name         string `json:"name"`
// 	Type         string `json:"type"`
// 	Comment      string `json:"comment"`
// 	DisplayName  string `json:"display_name"`
// 	OriginalName string `json:"original_name"`

// 	Path []string `json:"-"`
// }

type ValueOptCfg struct {
	ValueFrom string `json:"value_from,omitempty" mapstructure:"value_from"`
	Value     any    `json:"value,omitempty" mapstructure:"value"`
	RealValue any    `json:"real_value,omitempty" mapstructure:"real_value"`
}

// func (field *ViewField) InitFieldPath() {
// 	if len(field.Path) == 0 {
// 		field.Path = strings.Split(field.Name, ".")
// 	}
// }

func IsSlice(i any) bool {
	kind := reflect.ValueOf(i).Kind()
	return kind == reflect.Slice || kind == reflect.Array
}

func IsSameType(arr []any) bool {
	if len(arr) == 0 {
		return true
	}

	firstType := reflect.TypeOf(arr[0])
	for _, v := range arr {
		if reflect.TypeOf(v) != firstType {
			return false
		}
	}

	return true
}

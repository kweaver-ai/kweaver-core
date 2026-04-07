// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package condition

import (
	"reflect"
	"strings"
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

	// DefaultSmallModelEnabled 关闭的错误信息
	DEFAULT_SMALL_MODEL_ENABLED_FALSE_ERROR string = "DefaultSmallModelEnabled is false"
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
		"best_fields":   true,
		"most_fields":   true,
		"cross_fields":  true,
		"phrase":        true,
		"phrase_prefix": true,
		"bool_prefix":   true,
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
	Name        string     `json:"field,omitempty" mapstructure:"field"`
	Operation   string     `json:"operation,omitempty" mapstructure:"operation"`
	SubConds    []*CondCfg `json:"sub_conditions,omitempty" mapstructure:"sub_conditions"`
	ValueOptCfg `mapstructure:",squash"`

	RemainCfg map[string]any `mapstructure:",remain"`

	NameField *ViewField `json:"-" mapstructure:"-"`
}

type ViewField struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	Comment      string `json:"comment"`
	DisplayName  string `json:"display_name"`
	OriginalName string `json:"original_name"`

	Path []string `json:"-"`
}

type ValueOptCfg struct {
	ValueFrom string `json:"value_from,omitempty" mapstructure:"value_from"`
	Value     any    `json:"value,omitempty" mapstructure:"value"`
}

func (field *ViewField) InitFieldPath() {
	if len(field.Path) == 0 {
		field.Path = strings.Split(field.Name, ".")
	}
}

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

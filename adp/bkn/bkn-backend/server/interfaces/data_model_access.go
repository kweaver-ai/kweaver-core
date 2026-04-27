// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"bkn-backend/common/condition"
	"context"
)

const (
	VIEW_QueryType_DSL = "DSL"
	VIEW_QueryType_SQL = "SQL"
)

var (
	SQL_CONDITION_OPS = []string{
		condition.OperationEq,
		condition.OperationNotEq,
		condition.OperationGt,
		condition.OperationGte,
		condition.OperationLt,
		condition.OperationLte,
		condition.OperationIn,
		condition.OperationNotIn,
		condition.OperationLike,
		condition.OperationNotLike,
		condition.OperationRange,
		condition.OperationOutRange,
		condition.OperationExist,
		condition.OperationNotExist,
	}

	DSL_CONDITION_OPS = []string{
		condition.OperationEq,
		condition.OperationNotEq,
		condition.OperationGt,
		condition.OperationGte,
		condition.OperationLt,
		condition.OperationLte,
		condition.OperationIn,
		condition.OperationNotIn,
		condition.OperationLike,
		condition.OperationNotLike,
		condition.OperationRange,
		condition.OperationOutRange,
		condition.OperationExist,
		condition.OperationNotExist,
		condition.OperationRegex,
		condition.OperationMatch,
		condition.OperationMatchPhrase,
		condition.OperationKNN,
	}

	DSL_KEYWORD_OPS = []string{
		condition.OperationEq,
		condition.OperationNotEq,
		condition.OperationIn,
		condition.OperationNotIn,
		condition.OperationGt,
		condition.OperationGte,
		condition.OperationLt,
		condition.OperationLte,
		condition.OperationRange,
		condition.OperationOutRange,
		condition.OperationRegex,
		condition.OperationLike,
		condition.OperationNotLike,
	}

	DSL_KEYWORD_OPS_MAP = map[string]string{
		condition.OperationEq:       condition.OperationEq,
		condition.OperationNotEq:    condition.OperationNotEq,
		condition.OperationIn:       condition.OperationIn,
		condition.OperationNotIn:    condition.OperationNotIn,
		condition.OperationGt:       condition.OperationGt,
		condition.OperationGte:      condition.OperationGte,
		condition.OperationLt:       condition.OperationLt,
		condition.OperationLte:      condition.OperationLte,
		condition.OperationRange:    condition.OperationRange,
		condition.OperationOutRange: condition.OperationOutRange,
		condition.OperationRegex:    condition.OperationRegex,
		condition.OperationLike:     condition.OperationLike,
		condition.OperationNotLike:  condition.OperationNotLike,
	}

	DSL_TEXT_OPS = []string{
		condition.OperationMatch,
		condition.OperationMultiMatch,
	}

	DSL_TEXT_OPS_MAP = map[string]string{
		condition.OperationMatch:      condition.OperationMatch,
		condition.OperationMultiMatch: condition.OperationMultiMatch,
	}

	SQL_STRING_OPS = []string{
		condition.OperationEq,
		condition.OperationNotEq,
		condition.OperationLike,
		condition.OperationNotLike,
		condition.OperationIn,
		condition.OperationNotIn,
		condition.OperationGt,
		condition.OperationGte,
		condition.OperationLt,
		condition.OperationLte,
		condition.OperationRange,
		condition.OperationOutRange,
	}

	// 配置了对象索引的操作符集合
	INDEX_CONDITION_OPS = []string{
		condition.OperationEq,
		condition.OperationNotEq,
		condition.OperationGt,
		condition.OperationGte,
		condition.OperationLt,
		condition.OperationLte,
		condition.OperationIn,
		condition.OperationNotIn,
		condition.OperationLike,
		condition.OperationNotLike,
		condition.OperationRegex,
		condition.OperationRange,
		condition.OperationOutRange,
		condition.OperationExist,
		condition.OperationNotExist,
		condition.OperationRegex,
		condition.OperationMatch,
		condition.OperationMatchPhrase,
		condition.OperationKNN,
	}
)

// 指标模型结构体
type MetricModel struct {
	ModelID      string           `json:"id"`
	ModelName    string           `json:"name"`
	GroupID      string           `json:"group_id"`
	GroupName    string           `json:"group_name"`
	AnalysisDims []Field          `json:"analysis_dimensions,omitempty"`
	FieldsMap    map[string]Field `json:"fields_map"` // 字段集
}

//go:generate mockgen -source ../interfaces/data_model_access.go -destination ../interfaces/mock/mock_data_model_access.go
type DataModelAccess interface {
	GetMetricModelByID(ctx context.Context, id string) (*MetricModel, error)
}

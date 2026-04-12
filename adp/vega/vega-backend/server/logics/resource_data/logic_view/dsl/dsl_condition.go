// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package dsl

import (
	"context"
	"fmt"
	"strings"
	"time"

	"vega-backend/interfaces"
	"vega-backend/logics/filter_condition"
)

// ConvertFilterCondition converts a FilterCondition to OpenSearch DSL.
func (c *logicViewDSLGenerator) ConvertFilterCondition(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (map[string]any, error) {

	switch condition.GetOperation() {
	case filter_condition.OperationAnd:
		return c.ConvertFilterConditionAnd(ctx, condition, fieldsMap)

	case filter_condition.OperationOr:
		return c.ConvertFilterConditionOr(ctx, condition, fieldsMap)

	default:
		return c.ConvertFilterConditionWithOpr(ctx, condition, fieldsMap)
	}
}

// ConvertFilterConditionAnd converts an AndCond to OpenSearch DSL.
func (c *logicViewDSLGenerator) ConvertFilterConditionAnd(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (map[string]any, error) {

	condAnd, ok := condition.(*filter_condition.AndCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.AndCond")
	}

	must := make([]map[string]any, 0, len(condAnd.SubConds))
	for _, subCond := range condAnd.SubConds {
		convertedCond, err := c.ConvertFilterCondition(ctx, subCond, fieldsMap)
		if err != nil {
			return nil, err
		}
		must = append(must, convertedCond)
	}

	return map[string]any{
		"bool": map[string]any{
			"must": must,
		},
	}, nil
}

// ConvertFilterConditionOr converts an OrCond to OpenSearch DSL.
func (c *logicViewDSLGenerator) ConvertFilterConditionOr(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (map[string]any, error) {

	condOr, ok := condition.(*filter_condition.OrCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.OrCond")
	}

	should := make([]map[string]any, 0, len(condOr.SubConds))
	for _, subCond := range condOr.SubConds {
		convertedCond, err := c.ConvertFilterCondition(ctx, subCond, fieldsMap)
		if err != nil {
			return nil, err
		}
		should = append(should, convertedCond)
	}

	return map[string]any{
		"bool": map[string]any{
			"should":               should,
			"minimum_should_match": 1,
		},
	}, nil
}

// ConvertFilterConditionWithOpr converts a FilterCondition with operation to OpenSearch DSL.
func (c *logicViewDSLGenerator) ConvertFilterConditionWithOpr(ctx context.Context, condition interfaces.FilterCondition,
	fieldsMap map[string]*interfaces.Property) (map[string]any, error) {

	switch condition.GetOperation() {
	case filter_condition.OperationEqual, filter_condition.OperationEqual2:
		return c.ConvertFilterConditionEqual(ctx, condition, fieldsMap)
	case filter_condition.OperationNotEqual, filter_condition.OperationNotEqual2:
		return c.ConvertFilterConditionNotEqual(ctx, condition, fieldsMap)
	case filter_condition.OperationGt, filter_condition.OperationGt2:
		return c.ConvertFilterConditionGt(ctx, condition)
	case filter_condition.OperationGte, filter_condition.OperationGte2:
		return c.ConvertFilterConditionGte(ctx, condition)
	case filter_condition.OperationLt, filter_condition.OperationLt2:
		return c.ConvertFilterConditionLt(ctx, condition)
	case filter_condition.OperationLte, filter_condition.OperationLte2:
		return c.ConvertFilterConditionLte(ctx, condition)
	case filter_condition.OperationIn:
		return c.ConvertFilterConditionIn(ctx, condition, fieldsMap)
	case filter_condition.OperationNotIn:
		return c.ConvertFilterConditionNotIn(ctx, condition, fieldsMap)
	case filter_condition.OperationLike:
		return c.ConvertFilterConditionLike(ctx, condition, fieldsMap)
	case filter_condition.OperationNotLike:
		return c.ConvertFilterConditionNotLike(ctx, condition, fieldsMap)
	case filter_condition.OperationContain:
		return c.ConvertFilterConditionContain(ctx, condition)
	case filter_condition.OperationNotContain:
		return c.ConvertFilterConditionNotContain(ctx, condition)
	case filter_condition.OperationRange:
		return c.ConvertFilterConditionRange(ctx, condition)
	case filter_condition.OperationOutRange:
		return c.ConvertFilterConditionOutRange(ctx, condition)
	case filter_condition.OperationNull:
		return c.ConvertFilterConditionNull(ctx, condition)
	case filter_condition.OperationNotNull:
		return c.ConvertFilterConditionNotNull(ctx, condition)
	case filter_condition.OperationEmpty:
		return c.ConvertFilterConditionEmpty(ctx, condition)
	case filter_condition.OperationNotEmpty:
		return c.ConvertFilterConditionNotEmpty(ctx, condition)
	case filter_condition.OperationPrefix:
		return c.ConvertFilterConditionPrefix(ctx, condition)
	case filter_condition.OperationNotPrefix:
		return c.ConvertFilterConditionNotPrefix(ctx, condition)
	case filter_condition.OperationBetween:
		return c.ConvertFilterConditionBetween(ctx, condition)
	case filter_condition.OperationExist:
		return c.ConvertFilterConditionExist(ctx, condition)
	case filter_condition.OperationNotExist:
		return c.ConvertFilterConditionNotExist(ctx, condition)
	case filter_condition.OperationRegex:
		return c.ConvertFilterConditionRegex(ctx, condition)
	case filter_condition.OperationMatch:
		return c.ConvertFilterConditionMatch(ctx, condition)
	case filter_condition.OperationMatchPhrase:
		return c.ConvertFilterConditionMatchPhrase(ctx, condition)
	case filter_condition.OperationTrue:
		return c.ConvertFilterConditionTrue(ctx, condition)
	case filter_condition.OperationFalse:
		return c.ConvertFilterConditionFalse(ctx, condition)
	case filter_condition.OperationBefore:
		return c.ConvertFilterConditionBefore(ctx, condition)
	case filter_condition.OperationCurrent:
		return c.ConvertFilterConditionCurrent(ctx, condition)
	case filter_condition.OperationMultiMatch:
		return c.ConvertFilterConditionMultiMatch(ctx, condition)
	case filter_condition.OperationKnnVector:
		return c.ConvertFilterConditionKnnVector(ctx, condition, fieldsMap)
	default:
		return nil, fmt.Errorf("operation %s is not supported", condition.GetOperation())
	}
}

// ConvertFilterConditionMultiMatch converts a MultiMatchCond to OpenSearch DSL.
func (c *logicViewDSLGenerator) ConvertFilterConditionMultiMatch(ctx context.Context, condition interfaces.FilterCondition) (map[string]any, error) {

	cond, ok := condition.(*filter_condition.MultiMatchCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.MultiMatchCond")
	}

	value := cond.Cfg.ValueOptCfg.Value
	fields := make([]string, 0, len(cond.Fields))
	for _, field := range cond.Fields {
		fields = append(fields, field.Name)
	}

	multiMatchQuery := map[string]any{
		"query":  value,
		"fields": fields,
	}

	if cond.MatchType != "" {
		multiMatchQuery["type"] = cond.MatchType
	}

	return map[string]any{
		"multi_match": multiMatchQuery,
	}, nil
}

// ConvertFilterConditionEqual converts an EqualCond to OpenSearch DSL.
func (c *logicViewDSLGenerator) ConvertFilterConditionEqual(ctx context.Context, condition interfaces.FilterCondition, fieldsMap map[string]*interfaces.Property) (map[string]any, error) {

	cond, ok := condition.(*filter_condition.EqualCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.EqualCond")
	}

	fieldName := cond.Lfield.OriginalName
	if fieldName == "" {
		fieldName = cond.Lfield.Name
	}
	keyword, err := c.getKeywordSuffix(fieldName, fieldsMap)
	if err != nil {
		return nil, err
	}
	switch cond.Cfg.ValueFrom {
	case interfaces.ValueFrom_Const:
		return map[string]any{
			"term": map[string]any{
				fieldName + keyword: cond.Value,
			},
		}, nil
	case interfaces.ValueFrom_Field:
		return map[string]any{
			"script": map[string]any{
				"source": fmt.Sprintf("doc['%s'].value == doc['%s'].value", fieldName+keyword, cond.Rfield.OriginalName+keyword),
			},
		}, nil
	default:
		return nil, fmt.Errorf("value_from %s is not supported", cond.Cfg.ValueFrom)
	}
}

// ConvertFilterConditionNotEqual converts a NotEqualCond to OpenSearch DSL.
func (c *logicViewDSLGenerator) ConvertFilterConditionNotEqual(ctx context.Context, condition interfaces.FilterCondition, fieldsMap map[string]*interfaces.Property) (map[string]any, error) {

	cond, ok := condition.(*filter_condition.NotEqualCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.NotEqualCond")
	}

	fieldName := cond.Lfield.OriginalName
	if fieldName == "" {
		fieldName = cond.Lfield.Name
	}
	keyword, err := c.getKeywordSuffix(fieldName, fieldsMap)
	if err != nil {
		return nil, err
	}
	switch cond.Cfg.ValueFrom {
	case interfaces.ValueFrom_Const:
		return map[string]any{
			"bool": map[string]any{
				"must_not": map[string]any{
					"term": map[string]any{
						fieldName + keyword: cond.Value,
					},
				},
			},
		}, nil
	case interfaces.ValueFrom_Field:
		return map[string]any{
			"script": map[string]any{
				"source": fmt.Sprintf("doc['%s'].value != doc['%s'].value", fieldName+keyword, cond.Rfield.OriginalName+keyword),
			},
		}, nil
	default:
		return nil, fmt.Errorf("value_from %s is not supported", cond.Cfg.ValueFrom)
	}
}

// ConvertFilterConditionGt converts a GtCond to OpenSearch DSL.
func (c *logicViewDSLGenerator) ConvertFilterConditionGt(ctx context.Context, condition interfaces.FilterCondition) (map[string]any, error) {

	cond, ok := condition.(*filter_condition.GtCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.GtCond")
	}

	switch cond.Cfg.ValueFrom {
	case interfaces.ValueFrom_Const:
		return map[string]any{
			"range": map[string]any{
				cond.Lfield.OriginalName: map[string]any{
					"gt": cond.Value,
				},
			},
		}, nil
	case interfaces.ValueFrom_Field:
		return map[string]any{
			"script": map[string]any{
				"source": fmt.Sprintf("doc['%s'].value > doc['%s'].value", cond.Lfield.OriginalName, cond.Rfield.OriginalName),
			},
		}, nil
	default:
		return nil, fmt.Errorf("value_from %s is not supported", cond.Cfg.ValueFrom)
	}
}

// ConvertFilterConditionGte converts a GteCond to OpenSearch DSL.
func (c *logicViewDSLGenerator) ConvertFilterConditionGte(ctx context.Context, condition interfaces.FilterCondition) (map[string]any, error) {

	cond, ok := condition.(*filter_condition.GteCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.GteCond")
	}

	switch cond.Cfg.ValueFrom {
	case interfaces.ValueFrom_Const:
		return map[string]any{
			"range": map[string]any{
				cond.Lfield.OriginalName: map[string]any{
					"gte": cond.Value,
				},
			},
		}, nil
	case interfaces.ValueFrom_Field:
		return map[string]any{
			"script": map[string]any{
				"source": fmt.Sprintf("doc['%s'].value >= doc['%s'].value", cond.Lfield.OriginalName, cond.Rfield.OriginalName),
			},
		}, nil
	default:
		return nil, fmt.Errorf("value_from %s is not supported", cond.Cfg.ValueFrom)
	}
}

// ConvertFilterConditionLt converts a LtCond to OpenSearch DSL.
func (c *logicViewDSLGenerator) ConvertFilterConditionLt(ctx context.Context, condition interfaces.FilterCondition) (map[string]any, error) {

	cond, ok := condition.(*filter_condition.LtCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.LtCond")
	}

	switch cond.Cfg.ValueFrom {
	case interfaces.ValueFrom_Const:
		return map[string]any{
			"range": map[string]any{
				cond.Lfield.OriginalName: map[string]any{
					"lt": cond.Value,
				},
			},
		}, nil
	case interfaces.ValueFrom_Field:
		return map[string]any{
			"script": map[string]any{
				"source": fmt.Sprintf("doc['%s'].value < doc['%s'].value", cond.Lfield.OriginalName, cond.Rfield.OriginalName),
			},
		}, nil
	default:
		return nil, fmt.Errorf("value_from %s is not supported", cond.Cfg.ValueFrom)
	}
}

// ConvertFilterConditionLte converts a LteCond to OpenSearch DSL.
func (c *logicViewDSLGenerator) ConvertFilterConditionLte(ctx context.Context, condition interfaces.FilterCondition) (map[string]any, error) {

	cond, ok := condition.(*filter_condition.LteCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.LteCond")
	}

	switch cond.Cfg.ValueFrom {
	case interfaces.ValueFrom_Const:
		return map[string]any{
			"range": map[string]any{
				cond.Lfield.OriginalName: map[string]any{
					"lte": cond.Value,
				},
			},
		}, nil
	case interfaces.ValueFrom_Field:
		return map[string]any{
			"script": map[string]any{
				"source": fmt.Sprintf("doc['%s'].value <= doc['%s'].value", cond.Lfield.OriginalName, cond.Rfield.OriginalName),
			},
		}, nil
	default:
		return nil, fmt.Errorf("value_from %s is not supported", cond.Cfg.ValueFrom)
	}
}

// ConvertFilterConditionIn converts an InCond to OpenSearch DSL.
func (c *logicViewDSLGenerator) ConvertFilterConditionIn(ctx context.Context, condition interfaces.FilterCondition, fieldsMap map[string]*interfaces.Property) (map[string]any, error) {

	cond, ok := condition.(*filter_condition.InCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.InCond")
	}

	if cond.Cfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [in] only supports ValueFrom_Const, got %s", cond.Cfg.ValueFrom)
	}

	fieldName := cond.Lfield.OriginalName
	if fieldName == "" {
		fieldName = cond.Lfield.Name
	}
	keyword, err := c.getKeywordSuffix(fieldName, fieldsMap)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"terms": map[string]any{
			fieldName + keyword: cond.Value,
		},
	}, nil
}

// ConvertFilterConditionNotIn converts a NotInCond to OpenSearch DSL.
func (c *logicViewDSLGenerator) ConvertFilterConditionNotIn(ctx context.Context, condition interfaces.FilterCondition, fieldsMap map[string]*interfaces.Property) (map[string]any, error) {

	cond, ok := condition.(*filter_condition.NotInCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.NotInCond")
	}

	if cond.Cfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [not_in] only supports ValueFrom_Const, got %s", cond.Cfg.ValueFrom)
	}

	fieldName := cond.Lfield.OriginalName
	if fieldName == "" {
		fieldName = cond.Lfield.Name
	}
	keyword, err := c.getKeywordSuffix(fieldName, fieldsMap)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"bool": map[string]any{
			"must_not": map[string]any{
				"terms": map[string]any{
					fieldName + keyword: cond.Value,
				},
			},
		},
	}, nil
}

// ConvertFilterConditionLike converts a LikeCond to OpenSearch DSL.
func (c *logicViewDSLGenerator) ConvertFilterConditionLike(ctx context.Context, condition interfaces.FilterCondition, fieldsMap map[string]*interfaces.Property) (map[string]any, error) {

	cond, ok := condition.(*filter_condition.LikeCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.LikeCond")
	}

	if cond.Cfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [like] only supports ValueFrom_Const, got %s", cond.Cfg.ValueFrom)
	}

	fieldName := cond.Lfield.OriginalName
	keyword, err := c.getKeywordSuffix(fieldName, fieldsMap)
	if err != nil {
		return nil, err
	}

	vStr := c.replaceLikeWildcards(cond.Value)
	return map[string]any{
		"regexp": map[string]any{
			fieldName + keyword: vStr,
		},
	}, nil
}

// ConvertFilterConditionNotLike converts a NotLikeCond to OpenSearch DSL.
func (c *logicViewDSLGenerator) ConvertFilterConditionNotLike(ctx context.Context, condition interfaces.FilterCondition, fieldsMap map[string]*interfaces.Property) (map[string]any, error) {

	cond, ok := condition.(*filter_condition.NotLikeCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.NotLikeCond")
	}

	if cond.Cfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [not_like] only supports ValueFrom_Const, got %s", cond.Cfg.ValueFrom)
	}

	fieldName := cond.Lfield.OriginalName
	keyword, err := c.getKeywordSuffix(fieldName, fieldsMap)
	if err != nil {
		return nil, err
	}

	vStr := c.replaceLikeWildcards(cond.Value)
	return map[string]any{
		"bool": map[string]any{
			"must_not": map[string]any{
				"regexp": map[string]any{
					fieldName + keyword: vStr,
				},
			},
		},
	}, nil
}

// ConvertFilterConditionContain converts a ContainCond to OpenSearch DSL.
func (c *logicViewDSLGenerator) ConvertFilterConditionContain(ctx context.Context, condition interfaces.FilterCondition) (map[string]any, error) {

	cond, ok := condition.(*filter_condition.ContainCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.ContainCond")
	}

	if cond.Cfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [contain] only supports ValueFrom_Const, got %s", cond.Cfg.ValueFrom)
	}
	// 文本包含查询必须用 match /match_phrase
	values := cond.Value
	should := make([]map[string]any, len(values))
	for i, v := range values {
		should[i] = map[string]any{
			"term": map[string]any{
				cond.Lfield.OriginalName: v,
			},
		}
	}

	return map[string]any{
		"bool": map[string]any{
			"should":               should,
			"minimum_should_match": 1,
		},
	}, nil
}

// ConvertFilterConditionNotContain converts a NotContainCond to OpenSearch DSL.
func (c *logicViewDSLGenerator) ConvertFilterConditionNotContain(ctx context.Context, condition interfaces.FilterCondition) (map[string]any, error) {

	cond, ok := condition.(*filter_condition.NotContainCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.NotContainCond")
	}

	if cond.Cfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [not_contain] only supports ValueFrom_Const, got %s", cond.Cfg.ValueFrom)
	}

	values := cond.Value
	mustNot := make([]map[string]any, len(values))
	for i, v := range values {
		mustNot[i] = map[string]any{
			"term": map[string]any{
				cond.Lfield.OriginalName: v,
			},
		}
	}

	return map[string]any{
		"bool": map[string]any{
			"must_not": mustNot,
		},
	}, nil
}

// ConvertFilterConditionRange converts a RangeCond to OpenSearch DSL.
func (c *logicViewDSLGenerator) ConvertFilterConditionRange(ctx context.Context, condition interfaces.FilterCondition) (map[string]any, error) {

	cond, ok := condition.(*filter_condition.RangeCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.RangeCond")
	}

	if cond.Cfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [range] only supports ValueFrom_Const, got %s", cond.Cfg.ValueFrom)
	}

	values := cond.Value
	if len(values) != 2 {
		return nil, fmt.Errorf("range condition requires exactly 2 values")
	}

	return map[string]any{
		"range": map[string]any{
			cond.Lfield.OriginalName: map[string]any{
				"gte": values[0],
				"lte": values[1],
			},
		},
	}, nil
}

// ConvertFilterConditionOutRange converts an OutRangeCond to OpenSearch DSL.
func (c *logicViewDSLGenerator) ConvertFilterConditionOutRange(ctx context.Context, condition interfaces.FilterCondition) (map[string]any, error) {

	cond, ok := condition.(*filter_condition.OutRangeCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.OutRangeCond")
	}

	if cond.Cfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [out_range] only supports ValueFrom_Const, got %s", cond.Cfg.ValueFrom)
	}

	values := cond.Value
	if len(values) != 2 {
		return nil, fmt.Errorf("out_range condition requires exactly 2 values")
	}

	return map[string]any{
		"bool": map[string]any{
			"should": []map[string]any{
				{
					"range": map[string]any{
						cond.Lfield.OriginalName: map[string]any{
							"lt": values[0],
						},
					},
				},
				{
					"range": map[string]any{
						cond.Lfield.OriginalName: map[string]any{
							"gt": values[1],
						},
					},
				},
			},
			"minimum_should_match": 1,
		},
	}, nil
}

// ConvertFilterConditionNull converts a NullCond to OpenSearch DSL.
func (c *logicViewDSLGenerator) ConvertFilterConditionNull(ctx context.Context, condition interfaces.FilterCondition) (map[string]any, error) {

	cond, ok := condition.(*filter_condition.NullCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.NullCond")
	}

	return map[string]any{
		"bool": map[string]any{
			"must_not": map[string]any{
				"exists": map[string]any{
					"field": cond.Lfield.OriginalName,
				},
			},
		},
	}, nil
}

// ConvertFilterConditionMatch converts a MatchCond to OpenSearch DSL.
func (c *logicViewDSLGenerator) ConvertFilterConditionMatch(ctx context.Context, condition interfaces.FilterCondition) (map[string]any, error) {

	cond, ok := condition.(*filter_condition.MatchCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.MatchCond")
	}

	value := cond.Cfg.ValueOptCfg.Value

	// 如果是全部字段匹配
	if len(cond.Fields) > 1 {
		should := make([]map[string]any, 0, len(cond.Fields))
		for _, field := range cond.Fields {
			should = append(should, map[string]any{
				"match": map[string]any{
					field.Name: value,
				},
			})
		}
		return map[string]any{
			"bool": map[string]any{
				"should":               should,
				"minimum_should_match": 1,
			},
		}, nil
	} else if len(cond.Fields) == 1 {
		// 单个字段匹配
		field := cond.Fields[0]
		return map[string]any{
			"match": map[string]any{
				field.Name: value,
			},
		}, nil
	}

	return nil, fmt.Errorf("match condition has no fields")
}

// ConvertFilterConditionMatchPhrase converts a MatchPhraseCond to OpenSearch DSL.
func (c *logicViewDSLGenerator) ConvertFilterConditionMatchPhrase(ctx context.Context, condition interfaces.FilterCondition) (map[string]any, error) {

	cond, ok := condition.(*filter_condition.MatchPhraseCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.MatchPhraseCond")
	}

	value := cond.Cfg.ValueOptCfg.Value

	// 如果是全部字段匹配
	if len(cond.Fields) > 1 {
		should := make([]map[string]any, 0, len(cond.Fields))
		for _, field := range cond.Fields {
			should = append(should, map[string]any{
				"match_phrase": map[string]any{
					field.Name: value,
				},
			})
		}
		return map[string]any{
			"bool": map[string]any{
				"should":               should,
				"minimum_should_match": 1,
			},
		}, nil
	} else if len(cond.Fields) == 1 {
		// 单个字段匹配
		field := cond.Fields[0]
		return map[string]any{
			"match_phrase": map[string]any{
				field.Name: value,
			},
		}, nil
	}

	return nil, fmt.Errorf("match_phrase condition has no fields")
}

// ConvertFilterConditionNotNull converts a NotNullCond to OpenSearch DSL.
func (c *logicViewDSLGenerator) ConvertFilterConditionNotNull(ctx context.Context, condition interfaces.FilterCondition) (map[string]any, error) {

	cond, ok := condition.(*filter_condition.NotNullCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.NotNullCond")
	}

	return map[string]any{
		"exists": map[string]any{
			"field": cond.Lfield.OriginalName,
		},
	}, nil
}

// ConvertFilterConditionEmpty converts an EmptyCond to OpenSearch DSL.
func (c *logicViewDSLGenerator) ConvertFilterConditionEmpty(ctx context.Context, condition interfaces.FilterCondition) (map[string]any, error) {

	cond, ok := condition.(*filter_condition.EmptyCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.EmptyCond")
	}

	return map[string]any{
		"bool": map[string]any{
			"should": []map[string]any{
				{
					"term": map[string]any{
						cond.Lfield.OriginalName: "",
					},
				},
				{
					"bool": map[string]any{
						"must_not": map[string]any{
							"exists": map[string]any{
								"field": cond.Lfield.OriginalName,
							},
						},
					},
				},
			},
			"minimum_should_match": 1,
		},
	}, nil
}

// ConvertFilterConditionNotEmpty converts a NotEmptyCond to OpenSearch DSL.
func (c *logicViewDSLGenerator) ConvertFilterConditionNotEmpty(ctx context.Context, condition interfaces.FilterCondition) (map[string]any, error) {

	cond, ok := condition.(*filter_condition.NotEmptyCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.NotEmptyCond")
	}

	return map[string]any{
		"bool": map[string]any{
			"must": []map[string]any{
				{
					"exists": map[string]any{
						"field": cond.Lfield.OriginalName,
					},
				},
				{
					"bool": map[string]any{
						"must_not": map[string]any{
							"term": map[string]any{
								cond.Lfield.OriginalName: "",
							},
						},
					},
				},
			},
		},
	}, nil
}

// ConvertFilterConditionPrefix converts a PrefixCond to OpenSearch DSL.
func (c *logicViewDSLGenerator) ConvertFilterConditionPrefix(ctx context.Context, condition interfaces.FilterCondition) (map[string]any, error) {

	cond, ok := condition.(*filter_condition.PrefixCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.PrefixCond")
	}

	vStr := cond.Value
	return map[string]any{
		"prefix": map[string]any{
			cond.Lfield.OriginalName: vStr,
		},
	}, nil
}

// ConvertFilterConditionNotPrefix converts a NotPrefixCond to OpenSearch DSL.
func (c *logicViewDSLGenerator) ConvertFilterConditionNotPrefix(ctx context.Context, condition interfaces.FilterCondition) (map[string]any, error) {

	cond, ok := condition.(*filter_condition.NotPrefixCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.NotPrefixCond")
	}

	if cond.Cfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [not_prefix] only supports ValueFrom_Const, got %s", cond.Cfg.ValueFrom)
	}

	vStr := cond.Value
	return map[string]any{
		"bool": map[string]any{
			"must_not": map[string]any{
				"prefix": map[string]any{
					cond.Lfield.OriginalName: vStr,
				},
			},
		},
	}, nil
}

// ConvertFilterConditionBetween converts a BetweenCond to OpenSearch DSL.
func (c *logicViewDSLGenerator) ConvertFilterConditionBetween(ctx context.Context, condition interfaces.FilterCondition) (map[string]any, error) {

	cond, ok := condition.(*filter_condition.BetweenCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.BetweenCond")
	}

	if cond.Cfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [between] only supports ValueFrom_Const, got %s", cond.Cfg.ValueFrom)
	}

	values := cond.Value
	if len(values) != 2 {
		return nil, fmt.Errorf("between condition requires exactly 2 values")
	}

	return map[string]any{
		"range": map[string]any{
			cond.Lfield.OriginalName: map[string]any{
				"gte": values[0],
				"lte": values[1],
			},
		},
	}, nil
}

// ConvertFilterConditionExist converts an ExistCond to OpenSearch DSL.
func (c *logicViewDSLGenerator) ConvertFilterConditionExist(ctx context.Context, condition interfaces.FilterCondition) (map[string]any, error) {

	cond, ok := condition.(*filter_condition.ExistCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.ExistCond")
	}

	return map[string]any{
		"exists": map[string]any{
			"field": cond.Lfield.OriginalName,
		},
	}, nil
}

// ConvertFilterConditionNotExist converts a NotExistCond to OpenSearch DSL.
func (c *logicViewDSLGenerator) ConvertFilterConditionNotExist(ctx context.Context, condition interfaces.FilterCondition) (map[string]any, error) {

	cond, ok := condition.(*filter_condition.NotExistCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.NotExistCond")
	}

	return map[string]any{
		"bool": map[string]any{
			"must_not": map[string]any{
				"exists": map[string]any{
					"field": cond.Lfield.OriginalName,
				},
			},
		},
	}, nil
}

// ConvertFilterConditionRegex converts a RegexCond to OpenSearch DSL.
func (c *logicViewDSLGenerator) ConvertFilterConditionRegex(ctx context.Context, condition interfaces.FilterCondition) (map[string]any, error) {

	cond, ok := condition.(*filter_condition.RegexCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.RegexCond")
	}

	if cond.Cfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [regex] only supports ValueFrom_Const, got %s", cond.Cfg.ValueFrom)
	}

	return map[string]any{
		"regexp": map[string]any{
			cond.Lfield.OriginalName: cond.Value,
		},
	}, nil
}

// ConvertFilterConditionTrue converts a TrueCond to OpenSearch DSL.
func (c *logicViewDSLGenerator) ConvertFilterConditionTrue(ctx context.Context, condition interfaces.FilterCondition) (map[string]any, error) {

	cond, ok := condition.(*filter_condition.TrueCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.TrueCond")
	}

	return map[string]any{
		"term": map[string]any{
			cond.Lfield.OriginalName: true,
		},
	}, nil
}

// ConvertFilterConditionFalse converts a FalseCond to OpenSearch DSL.
func (c *logicViewDSLGenerator) ConvertFilterConditionFalse(ctx context.Context, condition interfaces.FilterCondition) (map[string]any, error) {

	cond, ok := condition.(*filter_condition.FalseCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.FalseCond")
	}

	return map[string]any{
		"term": map[string]any{
			cond.Lfield.OriginalName: false,
		},
	}, nil
}

// ConvertFilterConditionBefore converts a BeforeCond to OpenSearch DSL.
func (c *logicViewDSLGenerator) ConvertFilterConditionBefore(ctx context.Context, condition interfaces.FilterCondition) (map[string]any, error) {

	cond, ok := condition.(*filter_condition.BeforeCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.BeforeCond")
	}

	if cond.Cfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [before] only supports ValueFrom_Const, got %s", cond.Cfg.ValueFrom)
	}

	values := cond.Value
	if len(values) != 2 {
		return nil, fmt.Errorf("before condition requires exactly 2 values")
	}

	interval, ok := values[0].(float64)
	if !ok {
		return nil, fmt.Errorf("condition [before] interval value should be a number")
	}
	datetimeStr, ok := values[1].(string)
	if !ok {
		return nil, fmt.Errorf("condition [before] datetime value should be a string")
	}

	// Parse the datetime string
	datetime, err := time.Parse(time.RFC3339, datetimeStr)
	if err != nil {
		return nil, fmt.Errorf("condition [before] failed to parse datetime: %v", err)
	}

	// Subtract the interval hours from the datetime
	resultTime := datetime.Add(-time.Duration(interval) * time.Hour)

	return map[string]any{
		"range": map[string]any{
			cond.Lfield.OriginalName: map[string]any{
				"lt": resultTime.Format(time.RFC3339),
			},
		},
	}, nil
}

// ConvertFilterConditionCurrent converts a CurrentCond to OpenSearch DSL.
func (c *logicViewDSLGenerator) ConvertFilterConditionCurrent(ctx context.Context, condition interfaces.FilterCondition) (map[string]any, error) {

	cond, ok := condition.(*filter_condition.CurrentCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.CurrentCond")
	}

	if cond.Cfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [current] only supports ValueFrom_Const, got %s", cond.Cfg.ValueFrom)
	}
	// Get current time
	now := time.Now()
	// Calculate the start and end of the current period
	var startTime, endTime time.Time
	switch cond.Value {
	case filter_condition.CurrentYear:
		startTime = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
		endTime = startTime.AddDate(1, 0, 0)
	case filter_condition.CurrentMonth:
		startTime = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		endTime = startTime.AddDate(0, 1, 0)
	case filter_condition.CurrentWeek:
		// Get Monday of the current week
		weekday := now.Weekday()
		if weekday == time.Sunday {
			weekday = 7
		}
		startTime = time.Date(now.Year(), now.Month(), now.Day()-int(weekday)+1, 0, 0, 0, 0, now.Location())
		endTime = startTime.AddDate(0, 0, 7)
	case filter_condition.CurrentDay:
		startTime = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		endTime = startTime.AddDate(0, 0, 1)
	case filter_condition.CurrentHour:
		startTime = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, now.Location())
		endTime = startTime.Add(time.Hour)
	case filter_condition.CurrentMinute:
		startTime = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), 0, 0, now.Location())
		endTime = startTime.Add(time.Minute)
	}

	return map[string]any{
		"range": map[string]any{
			cond.Lfield.OriginalName: map[string]any{
				"gte": startTime.Format(time.RFC3339),
				"lt":  endTime.Format(time.RFC3339),
			},
		},
	}, nil
}

// ConvertFilterConditionKnnVector converts a KnnVectorCond to OpenSearch DSL.
func (c *logicViewDSLGenerator) ConvertFilterConditionKnnVector(ctx context.Context, condition interfaces.FilterCondition, fieldsMap map[string]*interfaces.Property) (map[string]any, error) {

	cond, ok := condition.(*filter_condition.KnnVectorCond)
	if !ok {
		return nil, fmt.Errorf("condition is not *filter_condition.KnnVectorCond")
	}

	value := cond.Cfg.ValueOptCfg.Value

	// 构建 knn 查询
	knnQuery := map[string]any{
		cond.FilterFieldName: map[string]any{
			"vector": value,
		},
	}

	// 添加 limit_key 和 limit_value
	if limitKey, ok := cond.Cfg.RemainCfg["limit_key"].(string); ok && limitKey != "" {
		if limitValue, ok := cond.Cfg.RemainCfg["limit_value"]; ok {
			knnQuery[cond.FilterFieldName].(map[string]any)[limitKey] = limitValue
		}
	} else {
		// 使用默认值
		knnQuery[cond.FilterFieldName].(map[string]any)["k"] = 10
	}

	// 添加子条件
	if len(cond.SubConds) > 0 {
		filterQueries := make([]map[string]any, 0, len(cond.SubConds))
		for _, subCond := range cond.SubConds {
			subQuery, err := c.ConvertFilterCondition(ctx, subCond, fieldsMap)
			if err != nil {
				return nil, err
			}
			filterQueries = append(filterQueries, subQuery)
		}

		return map[string]any{
			"knn": knnQuery,
			"filter": map[string]any{
				"bool": map[string]any{
					"must": filterQueries,
				},
			},
		}, nil
	}

	return map[string]any{
		"knn": knnQuery,
	}, nil
}

// replaceLikeWildcards，把 like 的通配符替换成正则表达式里的字符
func (c *logicViewDSLGenerator) replaceLikeWildcards(input string) string {
	if input == "" {
		return input
	}

	var result strings.Builder
	escaped := false
	runes := []rune(input)

	for i := 0; i < len(runes); i++ {
		r := runes[i]

		if escaped {
			// 转义字符后的字符
			switch r {
			case '%', '_', '\\':
				result.WriteRune(r)
			default:
				// 如果转义了非特殊字符，保留转义符和字符
				result.WriteRune('\\')
				result.WriteRune(r)
			}
			escaped = false
		} else if r == '\\' {
			// 遇到转义符，检查是否是最后一个字符
			if i == len(runes)-1 {
				// 转义符在末尾，直接输出
				result.WriteRune(r)
			} else {
				// 标记转义状态，但不立即输出转义符
				escaped = true
			}
		} else if r == '%' {
			result.WriteString(".*")
		} else if r == '_' {
			result.WriteString(".")
		} else {
			result.WriteRune(r)
		}
	}

	// 处理以转义符结尾的情况
	if escaped {
		result.WriteRune('\\')
	}

	return result.String()
}

// getKeywordSuffix text 类型在部分查询场景（如 eq/in）下，需使用 keyword 类型的子字段，返回关键字后缀，否则返回空字符串
func (c *logicViewDSLGenerator) getKeywordSuffix(fieldName string, fieldsMap map[string]*interfaces.Property) (string, error) {
	for _, prop := range fieldsMap {
		if prop.OriginalName == fieldName && prop.Type == interfaces.DataType_Text {
			for _, feature := range prop.Features {
				if feature.FeatureType == interfaces.PropertyFeatureType_Keyword {
					return "." + feature.FeatureName, nil
				}
			}
			return "", fmt.Errorf("text field %s has no keyword feature, cannot be used for comparison", fieldName)
		}
	}
	return "", nil
}

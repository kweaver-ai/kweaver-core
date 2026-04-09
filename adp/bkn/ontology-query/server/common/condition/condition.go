// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package condition

import (
	"context"
	"fmt"
	"strings"

	"github.com/kweaver-ai/kweaver-go-lib/logger"

	dtype "ontology-query/interfaces/data_type"
)

const MaxSubCondition = 100

// Filter field names in API are object-type data property names (not view column names).
const (
	errFmtUnknownObjectTypeProperty = "过滤字段「%s」不是对象类的数据属性名；请使用对象类中已定义的数据属性作为 field"
	errFmtBinaryPropertyNoFilter    = "过滤字段「%s」为 binary 类型，不支持过滤"
	errFmtMissingViewMappedField    = "属性「%s」未配置视图映射字段(mapped_field)，无法通过数据视图查询"
)

// sql的字符串转义
var Special = strings.NewReplacer(`\`, `\\\\`, `'`, `\'`, `%`, `\%`, `_`, `\_`)

//go:generate mockgen -source ../condition/condition.go -destination ../condition/mock/mock_condition.go
type Condition interface {
	Convert(ctx context.Context, vectorizer func(ctx context.Context, property *DataProperty, word string) ([]VectorResp, error)) (string, error)
	Convert2SQL(ctx context.Context) (string, error) // 把condition转成sql的where条件

	// RewriteCond(ctx context.Context, vectorizer func(ctx context.Context, property *DataProperty, word string) ([]VectorResp, error)) (*CondCfg, error)
}

// 将过滤条件拼接到 dsl 请求的 query 部分
func NewCondition(ctx context.Context, cfg *CondCfg, fieldScope uint8, fieldsMap map[string]*DataProperty) (cond Condition, err error) {
	if cfg == nil {
		return nil, nil
	}
	switch cfg.Operation {
	case OperationAnd:
		cond, err = newAndCond(ctx, cfg, fieldScope, fieldsMap)
	case OperationOr:
		cond, err = newOrCond(ctx, cfg, fieldScope, fieldsMap)
	default:
		cond, err = NewCondWithOpr(ctx, cfg, fieldScope, fieldsMap)
	}
	if err != nil {
		return nil, err
	}

	return cond, nil
}

func NewCondWithOpr(ctx context.Context, cfg *CondCfg, fieldScope uint8, fieldsMap map[string]*DataProperty) (cond Condition, err error) {
	// multi_match之外的才校验
	if cfg.Operation != OperationMultiMatch {
		// 判断除 * 之外的字段权限
		if cfg.Name != AllField {
			field, ok := fieldsMap[cfg.Name]
			if !ok {
				return nil, fmt.Errorf(errFmtUnknownObjectTypeProperty, cfg.Name)
			}

			// 如果字段类型是 binary 类型，则不支持过滤
			if field.Type == dtype.DATATYPE_BINARY {
				return nil, fmt.Errorf(errFmtBinaryPropertyNoFilter, cfg.Name)
			}

			cfg.NameField = field
		}
	}

	switch cfg.Operation {
	case OperationEq:
		cond, err = NewEqCond(ctx, cfg, fieldsMap)
	case OperationNotEq:
		cond, err = NewNotEqCond(ctx, cfg, fieldsMap)
	case OperationGt:
		cond, err = NewGtCond(ctx, cfg, fieldsMap)
	case OperationGte:
		cond, err = NewGteCond(ctx, cfg, fieldsMap)
	case OperationLt:
		cond, err = NewLtCond(ctx, cfg, fieldsMap)
	case OperationLte:
		cond, err = NewLteCond(ctx, cfg, fieldsMap)
	case OperationIn:
		cond, err = NewInCond(ctx, cfg, fieldsMap)
	case OperationNotIn:
		cond, err = NewNotInCond(ctx, cfg, fieldsMap)
	case OperationLike:
		cond, err = NewLikeCond(ctx, cfg, fieldsMap)
	case OperationNotLike:
		cond, err = NewNotLikeCond(ctx, cfg, fieldsMap)
	case OperationRange:
		cond, err = NewRangeCond(ctx, cfg, fieldsMap)
	case OperationOutRange:
		cond, err = NewOutRangeCond(ctx, cfg, fieldsMap)
	case OperationExist:
		cond, err = NewExistCond(cfg)
	case OperationNotExist:
		cond, err = NewNotExistCond(cfg)
	case OperationRegex:
		cond, err = NewRegexCond(ctx, cfg, fieldsMap)
	case OperationMatch:
		cond, err = NewMatchCond(ctx, cfg, fieldScope, fieldsMap)
	case OperationMatchPhrase:
		cond, err = NewMatchPhraseCond(ctx, cfg, fieldScope, fieldsMap)
	case OperationKNN:
		cond, err = NewKnnCond(ctx, cfg, fieldScope, fieldsMap)
	case OperationMultiMatch:
		cond, err = NewMultiMatchCond(ctx, cfg, fieldScope, fieldsMap)
	case OperationPrefix:
		cond, err = NewPrefixCond(ctx, cfg, fieldsMap)
	case OperationNotPrefix:
		cond, err = NewNotPrefixCond(ctx, cfg, fieldsMap)
	case OperationNull:
		cond, err = NewNullCond(ctx, cfg, fieldsMap)
	case OperationNotNull:
		cond, err = NewNotNullCond(ctx, cfg, fieldsMap)
	case OperationContain:
		cond, err = NewContainCond(ctx, cfg, fieldsMap)
	case OperationNotContain:
		cond, err = NewNotContainCond(ctx, cfg, fieldsMap)
	case OperationTrue:
		cond, err = NewTrueCond(ctx, cfg, fieldsMap)
	case OperationFalse:
		cond, err = NewFalseCond(ctx, cfg, fieldsMap)
	case OperationBefore:
		cond, err = NewBeforeCond(ctx, cfg, fieldsMap)
	case OperationCurrent:
		cond, err = NewCurrentCond(ctx, cfg, fieldsMap)
	case OperationBetween:
		cond, err = NewBetweenCond(ctx, cfg, fieldsMap)

	default:
		return nil, fmt.Errorf("not support condition's operation: %s", cfg.Operation)
	}
	if err != nil {
		return nil, err
	}

	return cond, nil
}

func getFilterFieldName(name string, fieldsMap map[string]*DataProperty, isFullTextQuery bool) string {
	// 全文检索允许字段为 "*"
	if name == AllField {
		return name
	}

	// 如果字段为 __id, 转化为 open search内置字段 _id
	if name == MetaField_ID {
		return OS_MetaField_ID
	}

	// 如果是脱敏字段，字段添加后缀 _desensitize
	desensitizeFieldName := name + DESENSITIZE_FIELD_SUFFIX

	fieldInfo, ok1 := fieldsMap[name]
	_, ok2 := fieldsMap[desensitizeFieldName]
	if ok1 && ok2 {
		// 脱敏字段
		name = desensitizeFieldName
	}

	// 全文检索情况下，text 类型的字段不需要添加 keyword 后缀
	// 精确查询情况下，text 类型的字段配了keyword索引，则给字段名加上后缀 .keyword
	if !isFullTextQuery && ok1 &&
		fieldInfo.Type == dtype.DATATYPE_TEXT &&
		fieldInfo.IndexConfig != nil && fieldInfo.IndexConfig.KeywordConfig.Enabled {
		name = wrapKeyWordFieldName(name)
	}

	return name
}

// 转换成 keyword
func wrapKeyWordFieldName(fields ...string) string {
	for _, field := range fields {
		if field == "" {
			logger.Warn("missing metric name")
			return ""
		}
	}

	return strings.Join(fields, ".") + "." + dtype.KEYWORD_SUFFIX
}

// 把本体对属性的过滤条件重写为视图的过滤条件
func RewriteCondition(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*DataProperty,
	vectorizer func(ctx context.Context, property *DataProperty, word string) ([]VectorResp, error)) (viewCfg *CondCfg, err error) {

	if cfg == nil {
		return nil, nil
	}
	switch cfg.Operation {
	case OperationAnd:
		viewCfg, err = rewriteAndCondition(ctx, cfg, fieldsMap, vectorizer)
	case OperationOr:
		viewCfg, err = rewriteOrCondition(ctx, cfg, fieldsMap, vectorizer)
	default:
		viewCfg, err = rewriteCondWithOpr(ctx, cfg, fieldsMap, vectorizer)
	}
	if err != nil {
		return nil, err
	}

	return viewCfg, nil
}

func rewriteCondWithOpr(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*DataProperty,
	vectorizer func(ctx context.Context, property *DataProperty, word string) ([]VectorResp, error)) (viewCfg *CondCfg, err error) {

	// multi_match之外的才校验
	if cfg.Operation != OperationMultiMatch {
		// 判断除 * 之外的字段权限
		if cfg.Name != AllField {
			field, ok := fieldsMap[cfg.Name]
			if !ok {
				return nil, fmt.Errorf(errFmtUnknownObjectTypeProperty, cfg.Name)
			}

			// 如果字段类型是 binary 类型，则不支持过滤
			if field.Type == dtype.DATATYPE_BINARY {
				return nil, fmt.Errorf(errFmtBinaryPropertyNoFilter, cfg.Name)
			}

			cfg.NameField = field
			if field.MappedField.Name == "" {
				return nil, fmt.Errorf(errFmtMissingViewMappedField, cfg.Name)
			}
		}
	}

	switch cfg.Operation {
	case OperationEq:
		viewCfg, err = rewriteEqCond(cfg)
	case OperationNotEq:
		viewCfg, err = rewriteNotEqCond(cfg)
	case OperationGt:
		viewCfg, err = rewriteGtCond(cfg)
	case OperationGte:
		viewCfg, err = rewriteGteCond(cfg)
	case OperationLt:
		viewCfg, err = rewriteLtCond(cfg)
	case OperationLte:
		viewCfg, err = rewriteLteCond(cfg)
	case OperationIn:
		viewCfg, err = rewriteInCond(cfg)
	case OperationNotIn:
		viewCfg, err = rewriteNotInCond(cfg)
	case OperationLike:
		viewCfg, err = rewriteLikeCond(cfg)
	case OperationNotLike:
		viewCfg, err = rewriteNotLikeCond(cfg)
	case OperationRange:
		viewCfg, err = rewriteRangeCond(cfg)
	case OperationOutRange:
		viewCfg, err = rewriteOutRangeCond(cfg)
	case OperationExist:
		viewCfg, err = rewriteExistCond(cfg)
	case OperationNotExist:
		viewCfg, err = rewriteNotExistCond(cfg)
	case OperationRegex:
		viewCfg, err = rewriteRegexCond(cfg)
	case OperationMatch:
		viewCfg, err = rewriteMatchCond(cfg)
	case OperationMatchPhrase:
		viewCfg, err = rewriteMatchPhraseCond(cfg)
	case OperationKNN:
		viewCfg, err = rewriteKnnCond(ctx, cfg, vectorizer)
	case OperationMultiMatch:
		viewCfg, err = rewriteMultiMatchCond(cfg, fieldsMap)
	case OperationPrefix:
		viewCfg, err = rewritePrefixCond(cfg)
	case OperationNotPrefix:
		viewCfg, err = rewriteNotPrefixCond(cfg)
	case OperationNull:
		viewCfg, err = rewriteNullCond(cfg)
	case OperationNotNull:
		viewCfg, err = rewriteNotNullCond(cfg)
	case OperationContain:
		viewCfg, err = rewriteContainCond(cfg)
	case OperationNotContain:
		viewCfg, err = rewriteNotContainCond(cfg)
	case OperationTrue:
		viewCfg, err = rewriteTrueCond(cfg)
	case OperationFalse:
		viewCfg, err = rewriteFalseCond(cfg)
	case OperationBefore:
		viewCfg, err = rewriteBeforeCond(cfg)
	case OperationCurrent:
		viewCfg, err = rewriteCurrentCond(cfg)
	case OperationBetween:
		viewCfg, err = rewriteBetweenCond(cfg)
	case OperationEmpty:
		viewCfg, err = rewriteEmptyCond(cfg)
	case OperationNotEmpty:
		viewCfg, err = rewriteNotEmptyCond(cfg)
	default:
		return nil, fmt.Errorf("not support condition's operation: %s", cfg.Operation)
	}
	if err != nil {
		return nil, err
	}

	return viewCfg, nil
}

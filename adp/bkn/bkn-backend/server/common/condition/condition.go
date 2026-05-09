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

	dtype "bkn-backend/interfaces/data_type"
)

const MaxSubCondition = 100

// sql的字符串转义
var Special = strings.NewReplacer(`\`, `\\\\`, `'`, `\'`, `%`, `\%`, `_`, `\_`)

//go:generate mockgen -source ../condition/condition.go -destination ../condition/mock/mock_condition.go
type Condition interface {
	Convert(ctx context.Context, vectorizer func(ctx context.Context, words []string) ([]*VectorResp, error)) (string, error)
	Convert2SQL(ctx context.Context) (string, error) // 把condition转成sql的where条件
}

// 将过滤条件拼接到 dsl 请求的 query 部分
func NewCondition(ctx context.Context, cfg *CondCfg, fieldScope uint8, fieldsMap map[string]*ViewField) (cond Condition, err error) {
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

func NewCondWithOpr(ctx context.Context, cfg *CondCfg, fieldScope uint8, fieldsMap map[string]*ViewField) (cond Condition, err error) {

	// multi_match之外的才校验
	if cfg.Operation != OperationMultiMatch {
		// 判断除 * 之外的字段权限
		if cfg.Field != AllField {
			field, ok := fieldsMap[cfg.Field]
			if !ok {
				return nil, fmt.Errorf("condition config field name '%s' must in view original fields", cfg.Field)
			}

			// 如果字段类型是 binary 类型，则不支持过滤
			if field.Type == dtype.DATATYPE_BINARY {
				return nil, fmt.Errorf("condition config field '%s' is binary type, do not support filtering", cfg.Field)
			}

			cfg.NameField = field
		}
	}

	switch cfg.Operation {
	case OperationEq:
		cond, err = NewEqCond(ctx, cfg, fieldsMap)
	case OperationNotEq:
		cond, err = NewNotEqCond(ctx, cfg, fieldsMap)
	case OperationIn:
		cond, err = NewInCond(ctx, cfg, fieldsMap)
	case OperationNotIn:
		cond, err = NewNotInCond(ctx, cfg, fieldsMap)
	case OperationLike:
		cond, err = NewLikeCond(ctx, cfg, fieldsMap)
	case OperationNotLike:
		cond, err = NewNotLikeCond(ctx, cfg, fieldsMap)
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

	default:
		return nil, fmt.Errorf("not support condition's operation: %s", cfg.Operation)
	}
	if err != nil {
		return nil, err
	}

	return cond, nil
}

func getFilterFieldName(name string, fieldsMap map[string]*ViewField, isFullTextQuery bool) string {
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
	// 精确查询情况下，text 类型的字段给字段名加上后缀 .keyword
	if !isFullTextQuery && fieldInfo.Type == dtype.DATATYPE_TEXT {
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

// ConvertCondCfgToFilterCondition converts CondCfg to dataset filter condition format
// Reference: ontology-query's RewriteCondition pattern
func ConvertCondCfgToFilterCondition(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*ViewField,
	vectorizer func(ctx context.Context, word string) ([]*VectorResp, error)) (map[string]any, error) {
	if cfg == nil {
		return nil, nil
	}

	switch cfg.Operation {
	case OperationEq:
		return convertEqCondToDatasetFilterCondition(cfg)
	case OperationNotEq:
		return convertNotEqCondToDatasetFilterCondition(cfg)
	case OperationIn:
		return convertInCondToDatasetFilterCondition(cfg)
	case OperationNotIn:
		return convertNotInCondToDatasetFilterCondition(cfg)
	case OperationLike:
		return convertLikeCondToDatasetFilterCondition(cfg)
	case OperationNotLike:
		return convertNotLikeCondToDatasetFilterCondition(cfg)
	case OperationMatch:
		return convertMatchCondToDatasetFilterCondition(cfg, fieldsMap)
	case OperationMatchPhrase:
		return convertMatchPhraseCondToDatasetFilterCondition(cfg, fieldsMap)
	case OperationRegex:
		return convertRegexCondToDatasetFilterCondition(cfg)
	case OperationKNN:
		return convertKnnCondToDatasetFilterCondition(ctx, cfg, fieldsMap, vectorizer)
	case OperationMultiMatch:
		return convertMultiMatchCondToDatasetFilterCondition(cfg, fieldsMap)
	case OperationAnd:
		return convertAndCondToDatasetFilterCondition(ctx, cfg, fieldsMap, vectorizer)
	case OperationOr:
		return convertOrCondToDatasetFilterCondition(ctx, cfg, fieldsMap, vectorizer)
	default:
		return nil, fmt.Errorf("not support condition's operation for dataset filter: %s", cfg.Operation)
	}
}

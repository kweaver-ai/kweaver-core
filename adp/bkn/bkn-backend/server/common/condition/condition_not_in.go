// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package condition

import (
	"context"
	"fmt"
	"strings"

	"github.com/bytedance/sonic"
)

type NotInCond struct {
	mCfg             *CondCfg
	mValue           []any
	mFilterFieldName string
}

func NewNotInCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*ViewField) (Condition, error) {
	if cfg.ValueOptCfg.ValueFrom != ValueFrom_Const {
		return nil, fmt.Errorf("condition [not_in] does not support value_from type '%s'", cfg.ValueFrom)
	}

	if !IsSlice(cfg.ValueOptCfg.Value) {
		return nil, fmt.Errorf("condition [not_in] right value should be an array")
	}

	if !IsSameType(cfg.ValueOptCfg.Value.([]any)) {
		return nil, fmt.Errorf("condition [not_in] right value should be an array composed of elements of same type")
	}

	return &NotInCond{
		mCfg:             cfg,
		mValue:           cfg.ValueOptCfg.Value.([]any),
		mFilterFieldName: getFilterFieldName(cfg.Field, fieldsMap, false),
	}, nil
}

func (cond *NotInCond) Convert(ctx context.Context, vectorizer func(ctx context.Context, words []string) ([]*VectorResp, error)) (string, error) {
	res, err := sonic.Marshal(cond.mValue)
	if err != nil {
		return "", fmt.Errorf("condition [not_in] json marshal right value failed, %s", err.Error())
	}

	dslStr := fmt.Sprintf(`
					{
						"bool": {
							"must_not": [
								{
									"terms": {
										"%s": %v
									}
								}
							]
						}
					}`, cond.mFilterFieldName, string(res))

	return dslStr, nil

}

func (cond *NotInCond) Convert2SQL(ctx context.Context) (string, error) {
	_, err := sonic.Marshal(cond.mValue)
	if err != nil {
		return "", fmt.Errorf("condition [in] json marshal right value failed, %s", err.Error())
	}

	valueList := make([]string, len(cond.mValue))
	for i, v := range cond.mValue {
		vStr, ok := v.(string)
		if ok {
			valueList[i] = fmt.Sprintf(`'%v'`, vStr)
		} else {
			valueList[i] = fmt.Sprintf(`%v`, vStr)
		}
	}
	value := strings.Join(valueList, ",")
	sqlStr := fmt.Sprintf(`"%s" NOT IN %s`, cond.mFilterFieldName, "("+value+")")
	return sqlStr, nil
}

// convertNotInCondToDatasetFilterCondition converts NotInCond to dataset filter condition format
func convertNotInCondToDatasetFilterCondition(cfg *CondCfg) (map[string]any, error) {
	return map[string]any{
		"field":      cfg.Field,
		"operation":  "not_in",
		"value":      cfg.Value, // 数组
		"value_from": "const",
	}, nil
}

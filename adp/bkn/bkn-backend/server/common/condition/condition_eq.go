// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package condition

import (
	"context"
	"fmt"

	"bkn-backend/common"
)

type EqCond struct {
	mCfg             *CondCfg
	mFilterFieldName string
}

func NewEqCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*FieldCfg) (Condition, error) {
	if cfg.ValueFrom != ValueFrom_Const {
		return nil, fmt.Errorf("condition [eq] does not support value_from type '%s'", cfg.ValueFrom)
	}

	if common.IsSlice(cfg.Value) {
		return nil, fmt.Errorf("condition [eq] only supports single value")
	}

	return &EqCond{
		mCfg:             cfg,
		mFilterFieldName: getFilterFieldName(cfg.Field, fieldsMap, false),
	}, nil

}

// 注：由于term是包含操作，而不是等值比较，如果field的值是数组的话，无法做到精确相等。
func (cond *EqCond) Convert(ctx context.Context, vectorizer func(ctx context.Context, words []string) ([]*VectorResp, error)) (string, error) {
	v := cond.mCfg.Value
	vStr, ok := v.(string)
	if ok {
		v = fmt.Sprintf("%q", vStr)
	}

	dslStr := fmt.Sprintf(`
					{
						"term": {
							"%s": {
								"value": %v
							}
						}
					}`, cond.mFilterFieldName, v)

	return dslStr, nil
}

func (cond *EqCond) Convert2SQL(ctx context.Context) (string, error) {
	v := cond.mCfg.Value
	vStr, ok := v.(string)
	if ok {
		v = fmt.Sprintf(`'%v'`, vStr)
	}
	sqlStr := fmt.Sprintf(`"%s" = %v`, cond.mFilterFieldName, v)

	return sqlStr, nil
}

// convertEqCondToDatasetFilterCondition converts EqCond to dataset filter condition format
func convertEqCondToDatasetFilterCondition(cfg *CondCfg) (map[string]any, error) {
	return map[string]any{
		"field":      cfg.Field,
		"operation":  "==",
		"value":      cfg.Value,
		"value_from": "const",
	}, nil
}

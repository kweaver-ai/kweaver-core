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

	"bkn-backend/common"
)

type InCond struct {
	mCfg             *CondCfg
	mValue           []any
	mFilterFieldName string
}

func NewInCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*ViewField) (Condition, error) {
	if cfg.ValueOptCfg.ValueFrom != ValueFrom_Const {
		return nil, fmt.Errorf("condition [in] does not support value_from type '%s'", cfg.ValueFrom)
	}

	if !common.IsSlice(cfg.ValueOptCfg.Value) {
		return nil, fmt.Errorf("condition [in] right value should be an array")
	}

	if !common.IsSameType(cfg.ValueOptCfg.Value.([]any)) {
		return nil, fmt.Errorf("condition [in] right value should be an array composed of elements of same type")
	}

	return &InCond{
		mCfg:             cfg,
		mValue:           cfg.ValueOptCfg.Value.([]any),
		mFilterFieldName: getFilterFieldName(cfg.Name, fieldsMap, false),
	}, nil
}

func (cond *InCond) Convert(ctx context.Context, vectorizer func(ctx context.Context, words []string) ([]*VectorResp, error)) (string, error) {
	res, err := sonic.Marshal(cond.mValue)
	if err != nil {
		return "", fmt.Errorf("condition [in] json marshal right value failed, %s", err.Error())
	}

	dslStr := fmt.Sprintf(`
					{
						"terms": {
							"%s": %v
						}
					}`, cond.mFilterFieldName, string(res))

	return dslStr, nil
}

func (cond *InCond) Convert2SQL(ctx context.Context) (string, error) {
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
			valueList[i] = fmt.Sprintf(`%v`, v)
		}
	}
	value := strings.Join(valueList, ",")
	sqlStr := fmt.Sprintf(`"%s" IN %s`, cond.mFilterFieldName, "("+value+")")

	return sqlStr, nil
}

// convertInCondToDatasetFilterCondition converts InCond to dataset filter condition format
func convertInCondToDatasetFilterCondition(cfg *CondCfg) (map[string]any, error) {
	return map[string]any{
		"field":      cfg.Name,
		"operation":  "in",
		"value":      cfg.Value, // 数组
		"value_from": "const",
	}, nil
}

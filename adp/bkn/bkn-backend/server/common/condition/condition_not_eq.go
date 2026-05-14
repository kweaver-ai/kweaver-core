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

type NotEqCond struct {
	mCfg             *CondCfg
	mFilterFieldName string
}

func NewNotEqCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*ViewField) (Condition, error) {
	if cfg.ValueFrom != ValueFrom_Const {
		return nil, fmt.Errorf("condition [not_eq] does not support value_from type '%s'", cfg.ValueFrom)
	}

	if common.IsSlice(cfg.Value) {
		return nil, fmt.Errorf("condition [not_eq] only supports single value")
	}

	return &NotEqCond{
		mCfg:             cfg,
		mFilterFieldName: getFilterFieldName(cfg.Field, fieldsMap, false),
	}, nil

}

/*
	{
	  "bool": {
	    "must_not": [
	      {
	        "term": {
	          "<field>": {
	            "value": <value>
	          }
	        }
	      }
	    ]
	  }
	}
*/
func (cond *NotEqCond) Convert(ctx context.Context, vectorizer func(ctx context.Context, words []string) ([]*VectorResp, error)) (string, error) {
	v := cond.mCfg.Value
	vStr, ok := v.(string)
	if ok {
		v = fmt.Sprintf("%q", vStr)
	}

	dslStr := fmt.Sprintf(`
					{
						"bool": {
							"must_not": [
								{
									"term": {
										"%s": {
											"value": %v
										}
									}
								}
							]
						}
					}`, cond.mFilterFieldName, v)

	return dslStr, nil
}

func (cond *NotEqCond) Convert2SQL(ctx context.Context) (string, error) {
	v := cond.mCfg.Value
	vStr, ok := v.(string)
	if ok {
		v = fmt.Sprintf(`'%v'`, vStr)
	}
	sqlStr := fmt.Sprintf(`"%s" <> %v`, cond.mFilterFieldName, v)

	return sqlStr, nil
}

// convertNotEqCondToDatasetFilterCondition converts NotEqCond to dataset filter condition format
func convertNotEqCondToDatasetFilterCondition(cfg *CondCfg) (map[string]any, error) {
	return map[string]any{
		"field":      cfg.Field,
		"operation":  "!=",
		"value":      cfg.Value,
		"value_from": "const",
	}, nil
}

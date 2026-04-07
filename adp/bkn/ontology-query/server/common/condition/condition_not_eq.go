// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package condition

import (
	"context"
	"fmt"

	"ontology-query/common"
)

type NotEqCond struct {
	mCfg             *CondCfg
	mFilterFieldName string
}

func NewNotEqCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*DataProperty) (Condition, error) {

	if common.IsSlice(cfg.ValueOptCfg.Value) {
		return nil, fmt.Errorf("condition [not_eq] only supports single value")
	}

	return &NotEqCond{
		mCfg:             cfg,
		mFilterFieldName: getFilterFieldName(cfg.Name, fieldsMap, false),
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
func (cond *NotEqCond) Convert(ctx context.Context, vectorizer func(ctx context.Context, property *DataProperty, word string) ([]VectorResp, error)) (string, error) {
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

func rewriteNotEqCond(cfg *CondCfg) (*CondCfg, error) {

	// 过滤条件中的属性字段换成映射的视图字段
	if cfg.NameField.Name == "" {
		return nil, fmt.Errorf("不等于过滤[!=]操作符使用的过滤字段[%s]在对象类的属性中不存在", cfg.Name)
	}
	return &CondCfg{
		Name:        cfg.NameField.MappedField.Name,
		Operation:   cfg.Operation,
		ValueOptCfg: cfg.ValueOptCfg,
	}, nil
}

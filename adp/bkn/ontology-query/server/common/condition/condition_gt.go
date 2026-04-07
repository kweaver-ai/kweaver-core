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

type GtCond struct {
	mCfg             *CondCfg
	mFilterFieldName string
}

func NewGtCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*DataProperty) (Condition, error) {

	if common.IsSlice(cfg.ValueOptCfg.Value) {
		return nil, fmt.Errorf("condition [eq] only supports single value")
	}

	return &GtCond{
		mCfg:             cfg,
		mFilterFieldName: getFilterFieldName(cfg.Name, fieldsMap, false),
	}, nil

}

// 注：由于term是包含操作，而不是等值比较，如果field的值是数组的话，无法做到精确相等。
func (cond *GtCond) Convert(ctx context.Context, vectorizer func(ctx context.Context, property *DataProperty, word string) ([]VectorResp, error)) (string, error) {
	v := cond.mCfg.Value
	vStr, ok := v.(string)
	if ok {
		v = fmt.Sprintf("%q", vStr)
	}
	dslStr := fmt.Sprintf(`
					{
						"range": {
							"%s": {
								"gt": %v
							}
						}
					}`, cond.mFilterFieldName, v)

	return dslStr, nil
}

func (cond *GtCond) Convert2SQL(ctx context.Context) (string, error) {
	v := cond.mCfg.Value
	vStr, ok := v.(string)
	if ok {
		v = fmt.Sprintf(`'%v'`, vStr)
	}
	sqlStr := fmt.Sprintf(`"%s" > %v`, cond.mFilterFieldName, v)

	return sqlStr, nil
}

func rewriteGtCond(cfg *CondCfg) (*CondCfg, error) {

	// 过滤条件中的属性字段换成映射的视图字段
	if cfg.NameField.Name == "" {
		return nil, fmt.Errorf("大于[>]操作符使用的过滤字段[%s]在对象类的属性中不存在", cfg.Name)
	}
	return &CondCfg{
		Name:        cfg.NameField.MappedField.Name,
		Operation:   cfg.Operation,
		ValueOptCfg: cfg.ValueOptCfg,
	}, nil
}

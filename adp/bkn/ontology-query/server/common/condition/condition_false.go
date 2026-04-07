// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package condition

import (
	"context"
	"fmt"

	dtype "ontology-query/interfaces/data_type"
)

type FalseCond struct {
	mCfg             *CondCfg
	mFilterFieldName string
}

func NewFalseCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*DataProperty) (Condition, error) {
	if cfg.NameField.Type != dtype.DATATYPE_BOOLEAN &&
		dtype.SimpleTypeMapping[cfg.NameField.Type] != dtype.SimpleBool {
		return nil, fmt.Errorf("condition [false] left field is not a boolean field: %s:%s", cfg.NameField.Name, cfg.NameField.Type)
	}

	return &FalseCond{
		mCfg:             cfg,
		mFilterFieldName: getFilterFieldName(cfg.Name, fieldsMap, false),
	}, nil
}

// term 查询逻辑等于 字段存在 + 相等
func (cond *FalseCond) Convert(ctx context.Context, vectorizer func(ctx context.Context, property *DataProperty, word string) ([]VectorResp, error)) (string, error) {
	dslStr := fmt.Sprintf(`
	{
		"term": {
			"%s": false
		}
	}`, cond.mFilterFieldName)

	return dslStr, nil
}

func (cond *FalseCond) Convert2SQL(ctx context.Context) (string, error) {
	sqlStr := fmt.Sprintf(`"%s" = false`, cond.mFilterFieldName)
	return sqlStr, nil
}

func rewriteFalseCond(cfg *CondCfg) (*CondCfg, error) {
	// 过滤条件中的属性字段换成映射的视图字段
	if cfg.NameField.Name == "" {
		return nil, fmt.Errorf("为假[false]操作符使用的过滤字段[%s]在对象类的属性中不存在", cfg.Name)
	}
	return &CondCfg{
		Name:        cfg.NameField.MappedField.Name,
		Operation:   cfg.Operation,
		ValueOptCfg: cfg.ValueOptCfg,
	}, nil
}

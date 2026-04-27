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

type TrueCond struct {
	mCfg             *CondCfg
	mFilterFieldName string
}

// bool 类型为真
func NewTrueCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*DataProperty) (Condition, error) {
	if cfg.NameField.Type != dtype.DATATYPE_BOOLEAN &&
		dtype.SimpleTypeMapping[cfg.NameField.Type] != dtype.SimpleBool {
		return nil, fmt.Errorf("condition [true] left field is not a boolean field: %s:%s", cfg.NameField.Name, cfg.NameField.Type)
	}

	return &TrueCond{
		mCfg:             cfg,
		mFilterFieldName: getFilterFieldName(cfg.Name, fieldsMap, false),
	}, nil
}

func (cond *TrueCond) Convert(ctx context.Context, vectorizer func(ctx context.Context, property *DataProperty, word string) ([]VectorResp, error)) (string, error) {
	dslStr := fmt.Sprintf(`
	{
		"term": {
			"%s": true
		}
	}`, cond.mFilterFieldName)

	return dslStr, nil
}

func (cond *TrueCond) Convert2SQL(ctx context.Context) (string, error) {
	sqlStr := fmt.Sprintf(`"%s" = true`, cond.mFilterFieldName)
	return sqlStr, nil
}

func rewriteTrueCond(cfg *CondCfg) (*CondCfg, error) {
	// 过滤条件中的属性字段换成映射的视图字段
	if cfg.NameField.Name == "" {
		return nil, fmt.Errorf("为真[true]操作符使用的过滤字段[%s]在对象类的属性中不存在", cfg.Name)
	}
	return &CondCfg{
		Name:        cfg.NameField.MappedField.Name,
		Operation:   cfg.Operation,
		ValueOptCfg: cfg.ValueOptCfg,
	}, nil
}

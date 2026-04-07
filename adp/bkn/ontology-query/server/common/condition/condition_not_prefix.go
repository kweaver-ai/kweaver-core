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

type NotPrefixCond struct {
	mCfg             *CondCfg
	mValue           string
	mFilterFieldName string
}

func NewNotPrefixCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*DataProperty) (Condition, error) {
	if !dtype.DataType_IsString(cfg.NameField.Type) &&
		dtype.SimpleTypeMapping[cfg.NameField.Type] != dtype.SimpleChar {
		return nil, fmt.Errorf("condition [not_prefix] left field is not a string field: %s:%s", cfg.NameField.Name, cfg.NameField.Type)
	}

	if cfg.ValueOptCfg.ValueFrom != ValueFrom_Const {
		return nil, fmt.Errorf("condition [not_prefix] does not support value_from type '%s'", cfg.ValueOptCfg.ValueFrom)
	}

	val, ok := cfg.ValueOptCfg.Value.(string)
	if !ok {
		return nil, fmt.Errorf("condition [not_prefix] right value is not a string value: %v", cfg.ValueOptCfg.Value)
	}

	return &NotPrefixCond{
		mCfg:             cfg,
		mValue:           val,
		mFilterFieldName: getFilterFieldName(cfg.Name, fieldsMap, false),
	}, nil
}

func (cond *NotPrefixCond) Convert(ctx context.Context, vectorizer func(ctx context.Context, property *DataProperty, word string) ([]VectorResp, error)) (string, error) {
	v := cond.mCfg.Value
	vStr, ok := v.(string)
	if ok {
		v = fmt.Sprintf("%q", vStr)
	}

	dslStr := fmt.Sprintf(`
	{
		"bool": {
			"must": {
				"exists": {
					"field": "%s"
				}
			},
			"must_not": {
				"prefix": {
					"%s": %v
				}
			}
		}
	}`, cond.mFilterFieldName, cond.mFilterFieldName, v)

	return dslStr, nil
}

func (cond *NotPrefixCond) Convert2SQL(ctx context.Context) (string, error) {
	v := cond.mCfg.Value
	vStr, ok := v.(string)
	if ok {
		v = Special.Replace(fmt.Sprintf("%v", vStr))
	}

	vStr = fmt.Sprintf("%v", v)
	sqlStr := fmt.Sprintf(`"%s" NOT LIKE '%s'`, cond.mFilterFieldName, vStr+"%")

	return sqlStr, nil
}

func rewriteNotPrefixCond(cfg *CondCfg) (*CondCfg, error) {
	// 过滤条件中的属性字段换成映射的视图字段
	if cfg.NameField.Name == "" {
		return nil, fmt.Errorf("开头不是[not_prefix]操作符使用的过滤字段[%s]在对象类的属性中不存在", cfg.Name)
	}
	return &CondCfg{
		Name:        cfg.NameField.MappedField.Name,
		Operation:   cfg.Operation,
		ValueOptCfg: cfg.ValueOptCfg,
	}, nil
}

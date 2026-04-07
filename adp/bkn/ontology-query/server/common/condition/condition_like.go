// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package condition

import (
	"context"
	"fmt"

	"ontology-query/common"
	dtype "ontology-query/interfaces/data_type"
)

type LikeCond struct {
	mCfg             *CondCfg
	mValue           string
	mFilterFieldName string
}

func NewLikeCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*DataProperty) (Condition, error) {
	_, ok := fieldsMap[cfg.Name]
	if !ok {
		return nil, fmt.Errorf("condition config field name '%s' must in view original fields", cfg.Name)
	}

	if !dtype.DataType_IsString(cfg.NameField.Type) &&
		dtype.SimpleTypeMapping[cfg.NameField.Type] != dtype.SimpleChar {
		return nil, fmt.Errorf("condition [like] left field is not a string field: %s:%s", cfg.NameField.Name, cfg.NameField.Type)
	}

	val, ok := cfg.ValueOptCfg.Value.(string)
	if !ok {
		return nil, fmt.Errorf("condition [like] right value is not a string value: %v", cfg.Value)
	}

	return &LikeCond{
		mCfg:             cfg,
		mValue:           val,
		mFilterFieldName: getFilterFieldName(cfg.Name, fieldsMap, false),
	}, nil
}

func (cond *LikeCond) Convert(ctx context.Context, vectorizer func(ctx context.Context, property *DataProperty, word string) ([]VectorResp, error)) (string, error) {
	// 替换 like 里的通配符
	v := common.ReplaceLikeWildcards(cond.mValue)
	v = fmt.Sprintf("%q", v)
	dslStr := fmt.Sprintf(`
					{
						"regexp": {
							"%s": %v
						}
					}`, cond.mFilterFieldName, v)

	return dslStr, nil
}

func (cond *LikeCond) Convert2SQL(ctx context.Context) (string, error) {
	v := cond.mCfg.Value
	vStr, ok := v.(string)
	if ok {
		v = Special.Replace(fmt.Sprintf("%v", vStr))
	}

	vStr = fmt.Sprintf("%v", v)
	sqlStr := fmt.Sprintf(`"%s" LIKE '%s'`, cond.mFilterFieldName, "%"+vStr+"%")

	return sqlStr, nil
}

func rewriteLikeCond(cfg *CondCfg) (*CondCfg, error) {

	// 过滤条件中的属性字段换成映射的视图字段
	if cfg.NameField.Name == "" {
		return nil, fmt.Errorf("相似过滤[like]操作符使用的过滤字段[%s]在对象类的属性中不存在", cfg.Name)
	}
	return &CondCfg{
		Name:      cfg.NameField.MappedField.Name,
		Operation: cfg.Operation,
		ValueOptCfg: ValueOptCfg{
			Value:     cfg.ValueOptCfg.Value,
			RealValue: cfg.ValueOptCfg.Value, // 把本体的like的 value 传到视图的 like 过滤的 real_value
		},
	}, nil
}

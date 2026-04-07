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

type NotLikeCond struct {
	mCfg             *CondCfg
	mValue           string
	mFilterFieldName string
}

func NewNotLikeCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*DataProperty) (Condition, error) {
	_, ok := fieldsMap[cfg.Name]
	if !ok {
		return nil, fmt.Errorf("condition config field name '%s' must in view original fields", cfg.Name)
	}

	if !dtype.DataType_IsString(cfg.NameField.Type) &&
		dtype.SimpleTypeMapping[cfg.NameField.Type] != dtype.SimpleChar {
		return nil, fmt.Errorf("condition [not_like] left field is not a string field: %s:%s", cfg.NameField.Name, cfg.NameField.Type)
	}

	val, ok := cfg.ValueOptCfg.Value.(string)
	if !ok {
		return nil, fmt.Errorf("condition [not_like] right value is not a string value: %v", cfg.Value)
	}

	return &NotLikeCond{
		mCfg:             cfg,
		mValue:           val,
		mFilterFieldName: getFilterFieldName(cfg.Name, fieldsMap, false),
	}, nil
}

func (cond *NotLikeCond) Convert(ctx context.Context, vectorizer func(ctx context.Context, property *DataProperty, word string) ([]VectorResp, error)) (string, error) {
	// 替换 like 里的通配符
	v := common.ReplaceLikeWildcards(cond.mValue)
	v = fmt.Sprintf("%q", v)

	dslStr := fmt.Sprintf(`
					{
						"bool": {
							"must_not": [
								{
									"regexp": {
										"%s": %v
									}
								}
							]
						}
					}`, cond.mFilterFieldName, v)

	return dslStr, nil
}

func (cond *NotLikeCond) Convert2SQL(ctx context.Context) (string, error) {
	v := cond.mCfg.Value
	vStr, ok := v.(string)
	if ok {
		v = Special.Replace(fmt.Sprintf("%v", vStr))
	}

	vStr = fmt.Sprintf("%v", v)
	sqlStr := fmt.Sprintf(`"%s" NOT LIKE '%s'`, cond.mFilterFieldName, "%"+vStr+"%")

	return sqlStr, nil
}

func rewriteNotLikeCond(cfg *CondCfg) (*CondCfg, error) {

	// 过滤条件中的属性字段换成映射的视图字段
	if cfg.NameField.Name == "" {
		return nil, fmt.Errorf("不相似过滤[not_like]操作符使用的过滤字段[%s]在对象类的属性中不存在", cfg.Name)
	}
	return &CondCfg{
		Name:      cfg.NameField.MappedField.Name,
		Operation: cfg.Operation,
		ValueOptCfg: ValueOptCfg{
			Value:     cfg.ValueOptCfg.Value,
			RealValue: cfg.ValueOptCfg.Value, // 把本体的 not_like 的 value 传到视图的 not_like 过滤的 real_value
		},
	}, nil
}

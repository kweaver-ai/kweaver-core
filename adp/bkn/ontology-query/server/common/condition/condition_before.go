// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package condition

import (
	"context"
	"fmt"
	"os"

	dtype "ontology-query/interfaces/data_type"
)

type BeforeCond struct {
	mCfg             *CondCfg
	mValue           any
	mUnit            string
	mFilterFieldName string
}

func NewBeforeCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*DataProperty) (Condition, error) {
	// 检查是否为日期/时间类型
	simpleType := dtype.SimpleTypeMapping[cfg.NameField.Type]
	if simpleType != dtype.SimpleDate && simpleType != dtype.SimpleDatetime && simpleType != dtype.SimpleTime {
		return nil, fmt.Errorf("condition [before] left field is not a date/time field: %s:%s", cfg.NameField.Name, cfg.NameField.Type)
	}

	if cfg.ValueOptCfg.ValueFrom != ValueFrom_Const {
		return nil, fmt.Errorf("condition [before] does not support value_from type '%s'", cfg.ValueOptCfg.ValueFrom)
	}

	unit, exist := cfg.RemainCfg["unit"].(string)
	if !exist {
		return nil, fmt.Errorf("condition [before] unit is not specified")
	}

	return &BeforeCond{
		mCfg:             cfg,
		mValue:           cfg.ValueOptCfg.Value,
		mUnit:            unit,
		mFilterFieldName: getFilterFieldName(cfg.Name, fieldsMap, false),
	}, nil
}

func (cond *BeforeCond) Convert(ctx context.Context, vectorizer func(ctx context.Context, property *DataProperty, word string) ([]VectorResp, error)) (string, error) {
	// before 操作符主要用于 SQL，OpenSearch DSL 暂不实现
	unitMap := map[string]string{
		"year":   "y",
		"month":  "M",
		"week":   "w",
		"day":    "d",
		"hour":   "h",
		"minute": "m",
		"second": "s",
	}

	unit, ok := unitMap[cond.mUnit]
	if !ok {
		unit = cond.mUnit // 如果已经缩写过则直接用
	}

	// 统一处理数值类型
	var val any = cond.mValue
	if f, ok := val.(float64); ok {
		val = int64(f)
	}

	return fmt.Sprintf(`{"range":{"%s":{"gte":"now-%v%s","lte":"now"}}}`,
		cond.mFilterFieldName, val, unit), nil
}

func (cond *BeforeCond) Convert2SQL(ctx context.Context) (string, error) {
	// 获取时区，默认为 UTC
	tz := os.Getenv("TZ")
	if tz == "" {
		tz = "UTC"
	}

	sqlStr := fmt.Sprintf(`"%s" >= DATE_add('%s', -%v, CURRENT_TIMESTAMP AT TIME ZONE 'UTC' AT TIME ZONE '%s') 
		AND "%s" <= CURRENT_TIMESTAMP AT TIME ZONE 'UTC' AT TIME ZONE '%s'`,
		cond.mFilterFieldName, cond.mUnit, cond.mValue, tz, cond.mFilterFieldName, tz)
	return sqlStr, nil
}

func rewriteBeforeCond(cfg *CondCfg) (*CondCfg, error) {
	// 过滤条件中的属性字段换成映射的视图字段
	if cfg.NameField == nil || cfg.NameField.Name == "" {
		return nil, fmt.Errorf("过滤条件 [before] 使用的字段「%s」在对象类数据属性中不存在", cfg.Name)
	}
	return &CondCfg{
		Name:        cfg.NameField.MappedField.Name,
		Operation:   cfg.Operation,
		ValueOptCfg: cfg.ValueOptCfg,
		RemainCfg:   cfg.RemainCfg,
	}, nil
}

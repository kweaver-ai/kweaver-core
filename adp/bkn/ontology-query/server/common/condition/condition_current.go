// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package condition

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	dtype "ontology-query/interfaces/data_type"
)

type CurrentCond struct {
	mCfg             *CondCfg
	mValue           string
	mFilterFieldName string
}

func NewCurrentCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*DataProperty) (Condition, error) {
	// 检查是否为日期/时间类型
	simpleType := dtype.SimpleTypeMapping[cfg.NameField.Type]
	if simpleType != dtype.SimpleDate && simpleType != dtype.SimpleDatetime && simpleType != dtype.SimpleTime {
		return nil, fmt.Errorf("condition [current] left field is not a date/time field: %s:%s", cfg.NameField.Name, cfg.NameField.Type)
	}

	if cfg.ValueOptCfg.ValueFrom != ValueFrom_Const {
		return nil, fmt.Errorf("condition [current] does not support value_from type '%s'", cfg.ValueOptCfg.ValueFrom)
	}

	val, ok := cfg.ValueOptCfg.Value.(string)
	if !ok {
		return nil, fmt.Errorf("condition [current] right value should be string")
	}

	// 验证 unit 值
	validUnits := map[string]bool{
		"year":   true,
		"month":  true,
		"week":   true,
		"day":    true,
		"hour":   true,
		"minute": true,
	}
	if !validUnits[val] {
		return nil, errors.New(`condition [current] right value should be one of ["year", "month", "week", "day", "hour", "minute"], actual is ` + val)
	}

	return &CurrentCond{
		mCfg:             cfg,
		mValue:           val,
		mFilterFieldName: getFilterFieldName(cfg.Name, fieldsMap, false),
	}, nil
}

func (cond *CurrentCond) Convert(ctx context.Context, vectorizer func(ctx context.Context, property *DataProperty, word string) ([]VectorResp, error)) (string, error) {
	// current 操作符主要用于 SQL，OpenSearch DSL 暂不实现
	return "", nil
}

func (cond *CurrentCond) Convert2SQL(ctx context.Context) (string, error) {
	// 获取时区，默认为 UTC
	tz := os.Getenv("TZ")
	if tz == "" {
		tz = "UTC"
	}

	// 加载时区
	location, err := time.LoadLocation(tz)
	if err != nil {
		location = time.UTC
	}

	now := time.Now().In(location)
	var start, end time.Time

	switch cond.mValue {
	case "year":
		start = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, location)
		end = start.AddDate(1, 0, 0)
	case "month":
		start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, location)
		end = start.AddDate(0, 1, 0)
	case "week":
		// 计算本周的周一
		weekday := now.Weekday()
		offset := int(time.Monday - weekday)
		if offset > 0 {
			offset -= 7 // 如果今天是周日，需要减去7天
		}
		start = time.Date(now.Year(), now.Month(), now.Day()+offset, 0, 0, 0, 0, location)
		end = start.AddDate(0, 0, 7)
	case "day":
		start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location)
		end = start.AddDate(0, 0, 1)
	case "hour":
		start = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, location)
		end = start.Add(time.Hour)
	case "minute":
		start = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), 0, 0, location)
		end = start.Add(time.Minute)
	default:
		return "", fmt.Errorf("unsupported unit: %s", cond.mValue)
	}

	sqlStr := fmt.Sprintf(`"%s" BETWEEN from_unixtime(%d) AND from_unixtime(%d)`,
		cond.mFilterFieldName, start.Unix(), end.Unix())
	return sqlStr, nil
}

func rewriteCurrentCond(cfg *CondCfg) (*CondCfg, error) {
	// 过滤条件中的属性字段换成映射的视图字段
	if cfg.NameField.Name == "" {
		return nil, fmt.Errorf("当前[current]操作符使用的过滤字段[%s]在对象类的属性中不存在", cfg.Name)
	}
	return &CondCfg{
		Name:        cfg.NameField.MappedField.Name,
		Operation:   cfg.Operation,
		ValueOptCfg: cfg.ValueOptCfg,
	}, nil
}

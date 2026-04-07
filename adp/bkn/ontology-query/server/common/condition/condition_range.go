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

type RangeCond struct {
	mCfg             *CondCfg
	mValue           []any
	mFilterFieldName string
}

func NewRangeCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*DataProperty) (Condition, error) {

	val, ok := cfg.ValueOptCfg.Value.([]any)
	if !ok {
		return nil, fmt.Errorf("condition [range] right value should be an array of length 2")
	}

	if len(val) != 2 {
		return nil, fmt.Errorf("condition [range] right value should be an array of length 2")
	}

	if !IsSameType(val) {
		return nil, fmt.Errorf("condition [range] right value should be of the same type")
	}

	return &RangeCond{
		mCfg:             cfg,
		mValue:           val,
		mFilterFieldName: getFilterFieldName(cfg.Name, fieldsMap, false),
	}, nil
}

// range 左闭右开区间
func (cond *RangeCond) Convert(ctx context.Context, vectorizer func(ctx context.Context, property *DataProperty, word string) ([]VectorResp, error)) (string, error) {
	gte := cond.mValue[0]
	lt := cond.mValue[1]

	if _, ok := gte.(string); ok {
		gte = fmt.Sprintf("%q", gte)
		lt = fmt.Sprintf("%q", lt)
	}

	var dslStr string
	if cond.mCfg.NameField.Type == dtype.DATATYPE_DATETIME {
		var format string
		switch gte.(type) {
		case string:
			format = "strict_date_optional_time"
		case float64:
			format = "epoch_millis"
			gte = int64(gte.(float64))
			lt = int64(lt.(float64))
		}

		dslStr = fmt.Sprintf(`
		{
			"range": {
				"%s": {
					"gte": %v,
					"lt": %v,
					"format": "%s"
				}
			}
		}`, cond.mFilterFieldName, gte, lt, format)

	} else {
		dslStr = fmt.Sprintf(`
			{
				"range": {
					"%s": {
						"gte": %v,
						"lt": %v
					}
				}
			}`, cond.mFilterFieldName, gte, lt)

	}

	return dslStr, nil
}

func (cond *RangeCond) Convert2SQL(ctx context.Context) (string, error) {
	// range表示左闭右开区间 [gte, lt)
	gte := cond.mValue[0]
	lt := cond.mValue[1]

	// 处理字符串类型的值，需要用单引号包裹
	gteStr, ok := gte.(string)
	if ok {
		gteStr = Special.Replace(fmt.Sprintf("%q", gteStr))
	} else {
		gteStr = fmt.Sprintf("%v", gte)
	}

	ltStr, ok := lt.(string)
	if ok {
		ltStr = Special.Replace(fmt.Sprintf("%q", ltStr))
	} else {
		ltStr = fmt.Sprintf("%v", lt)
	}

	// 构建SQL条件：字段名 >= 左边界 AND 字段名 < 右边界
	sqlStr := fmt.Sprintf("\"%s\" >= %s AND \"%s\" < %s", cond.mFilterFieldName, gteStr, cond.mFilterFieldName, ltStr)
	return sqlStr, nil
}

func rewriteRangeCond(cfg *CondCfg) (*CondCfg, error) {

	// 过滤条件中的属性字段换成映射的视图字段
	if cfg.NameField.Name == "" {
		return nil, fmt.Errorf("范围内过滤[range]操作符使用的过滤字段[%s]在对象类的属性中不存在", cfg.Name)
	}
	return &CondCfg{
		Name:        cfg.NameField.MappedField.Name,
		Operation:   cfg.Operation,
		ValueOptCfg: cfg.ValueOptCfg,
	}, nil
}

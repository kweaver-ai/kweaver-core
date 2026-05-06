// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package filter_condition

import (
	"context"
	"fmt"

	"vega-backend/interfaces"
)

type MatchCond struct {
	Cfg    *interfaces.FilterCondCfg
	Fields []*interfaces.Property
}

func (c *MatchCond) GetOperation() string { return OperationMatch }

func (c *MatchCond) SupportSubCond() bool       { return false }
func (c *MatchCond) NeedName() bool             { return false }
func (c *MatchCond) NeedValue() bool            { return true }
func (c *MatchCond) NeedConstValue() bool       { return true }
func (c *MatchCond) IsSingleValue() bool        { return true }
func (c *MatchCond) IsFixedLenArrayValue() bool { return false }
func (c *MatchCond) RequiredValueLen() int      { return -1 }

// match 条件, 判断字段是否匹配某个字符串
// 支持全部字段 *
func (c *MatchCond) New(ctx context.Context, cfg *interfaces.FilterCondCfg,
	fieldsMap map[string]*interfaces.Property) (interfaces.FilterCondition, error) {

	fields := make([]*interfaces.Property, 0)

	// 优先从 RemainCfg 中获取 fields 数组
	if cfgFields, ok := cfg.RemainCfg["fields"].([]any); ok {
		if len(cfgFields) == 1 && cfgFields[0].(string) == interfaces.AllField {
			for _, field := range fieldsMap {
				fields = append(fields, field)
			}
		} else {
			// 字段数组里的需要是个字符串数组
			for _, cfgField := range cfgFields {
				fieldName, ok := cfgField.(string)
				if !ok {
					return nil, fmt.Errorf("condition [match] 'fields' value should be a field name array, contain non string value[%v]", cfgField)
				}
				field, ok := fieldsMap[fieldName]
				if !ok {
					return nil, fmt.Errorf("condition [match] 'fields' exists any field not exists in resource [%s]", fieldName)
				}
				fields = append(fields, field)
			}
		}
	} else {
		// 兼容旧的单个 field 方式
		if cfg.Name == "" {
			return nil, fmt.Errorf("condition [match] left field is empty")
		}
		if cfg.Name == interfaces.AllField {
			for fieldName := range fieldsMap {
				fields = append(fields, fieldsMap[fieldName])
			}
		} else {
			field, ok := fieldsMap[cfg.Name]
			if !ok {
				return nil, fmt.Errorf("condition [match] left field '%s' not found", cfg.Name)
			}
			fields = append(fields, field)
		}
	}

	if cfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [match] does not support value_from type '%s'", cfg.ValueFrom)
	}

	return &MatchCond{
		Cfg:    cfg,
		Fields: fields,
	}, nil
}

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

type InCond struct {
	Cfg    *interfaces.FilterCondCfg
	Lfield *interfaces.Property
	Value  []any
}

func (c *InCond) GetOperation() string { return OperationIn }

func (c *InCond) SupportSubCond() bool       { return false }
func (c *InCond) NeedName() bool             { return true }
func (c *InCond) NeedValue() bool            { return true }
func (c *InCond) NeedConstValue() bool       { return true }
func (c *InCond) IsSingleValue() bool        { return false }
func (c *InCond) IsFixedLenArrayValue() bool { return false }
func (c *InCond) RequiredValueLen() int      { return -1 }

// in 条件, 判断字段是否在某个数组中
func (c *InCond) New(ctx context.Context, cfg *interfaces.FilterCondCfg,
	fieldsMap map[string]*interfaces.Property) (interfaces.FilterCondition, error) {

	if cfg.Name == "" {
		return nil, fmt.Errorf("condition [in] left field is empty")
	}
	field, ok := fieldsMap[cfg.Name]
	if !ok {
		// 如果字段未在Schema中定义，创建一个临时的Property对象
		field = &interfaces.Property{
			Name:         cfg.Name,
			OriginalName: cfg.Name,
		}
	}

	if cfg.ValueOptCfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [in] does not support value_from type '%s'", cfg.ValueFrom)
	}
	val, ok := cfg.ValueOptCfg.Value.([]any)
	if !ok {
		return nil, fmt.Errorf("condition [in] right value should be an array")
	}
	if len(val) == 0 {
		return nil, fmt.Errorf("condition [in] right value should be an array of length >= 1")
	}

	return &InCond{
		Cfg:    cfg,
		Lfield: field,
		Value:  val,
	}, nil
}

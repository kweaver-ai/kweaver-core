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

type NotInCond struct {
	Cfg    *interfaces.FilterCondCfg
	Lfield *interfaces.Property
	Value  []any
}

func (c *NotInCond) GetOperation() string { return OperationNotIn }

func (c *NotInCond) SupportSubCond() bool       { return false }
func (c *NotInCond) NeedName() bool             { return true }
func (c *NotInCond) NeedValue() bool            { return true }
func (c *NotInCond) NeedConstValue() bool       { return true }
func (c *NotInCond) IsSingleValue() bool        { return false }
func (c *NotInCond) IsFixedLenArrayValue() bool { return false }
func (c *NotInCond) RequiredValueLen() int      { return -1 }

// not_in 条件, 判断字段是否不在某个值数组中
func (c *NotInCond) New(ctx context.Context, cfg *interfaces.FilterCondCfg,
	fieldsMap map[string]*interfaces.Property) (interfaces.FilterCondition, error) {

	if cfg.Name == "" {
		return nil, fmt.Errorf("condition [not_in] left field is empty")
	}
	field, ok := fieldsMap[cfg.Name]
	if !ok {
		return nil, fmt.Errorf("condition [not_in] left field '%s' not found", cfg.Name)
	}

	if cfg.ValueOptCfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [not_in] does not support value_from type '%s'", cfg.ValueFrom)
	}
	val, ok := cfg.ValueOptCfg.Value.([]any)
	if !ok {
		return nil, fmt.Errorf("condition [not_in] right value should be an array")
	}
	if len(val) == 0 {
		return nil, fmt.Errorf("condition [not_in] right value should be an array of length >= 1")
	}

	return &NotInCond{
		Cfg:    cfg,
		Lfield: field,
		Value:  val,
	}, nil
}

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

type BetweenCond struct {
	Cfg    *interfaces.FilterCondCfg
	Lfield *interfaces.Property
	Value  []any
}

func (c *BetweenCond) GetOperation() string { return OperationBetween }

func (c *BetweenCond) SupportSubCond() bool       { return false }
func (c *BetweenCond) NeedName() bool             { return true }
func (c *BetweenCond) NeedValue() bool            { return true }
func (c *BetweenCond) NeedConstValue() bool       { return true }
func (c *BetweenCond) IsSingleValue() bool        { return false }
func (c *BetweenCond) IsFixedLenArrayValue() bool { return true }
func (c *BetweenCond) RequiredValueLen() int      { return 2 }

// between 条件，判断字段是否在某个区间内, 区间包含左右边界
func (c *BetweenCond) New(ctx context.Context, cfg *interfaces.FilterCondCfg,
	fieldsMap map[string]*interfaces.Property) (interfaces.FilterCondition, error) {

	if cfg.Name == "" {
		return nil, fmt.Errorf("condition [between] left field is empty")
	}
	field, ok := fieldsMap[cfg.Name]
	if !ok {
		return nil, fmt.Errorf("condition [between] left field '%s' not found", cfg.Name)
	}
	if !interfaces.DataType_IsDate(field.Type) && !interfaces.DataType_IsNumber(field.Type) {
		return nil, fmt.Errorf("condition [between] left field is not a date or number field: %s:%s", cfg.Name, field.Type)
	}

	if cfg.ValueOptCfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [between] does not support value_from type '%s'", cfg.ValueFrom)
	}
	val, ok := cfg.ValueOptCfg.Value.([]any)
	if !ok {
		return nil, fmt.Errorf("condition [between] right value should be an array")
	}
	if len(val) != 2 {
		return nil, fmt.Errorf("condition [between] right value should be an array of length 2")
	}

	return &BetweenCond{
		Cfg:    cfg,
		Lfield: field,
		Value:  val,
	}, nil
}

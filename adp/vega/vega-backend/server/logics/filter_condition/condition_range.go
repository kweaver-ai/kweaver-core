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

type RangeCond struct {
	Cfg    *interfaces.FilterCondCfg
	Lfield *interfaces.Property
	Value  []any
}

func (c *RangeCond) GetOperation() string { return OperationRange }

func (c *RangeCond) SupportSubCond() bool       { return false }
func (c *RangeCond) NeedName() bool             { return true }
func (c *RangeCond) NeedValue() bool            { return true }
func (c *RangeCond) NeedConstValue() bool       { return true }
func (c *RangeCond) IsSingleValue() bool        { return false }
func (c *RangeCond) IsFixedLenArrayValue() bool { return true }
func (c *RangeCond) RequiredValueLen() int      { return 2 }

// range 条件, 判断字段是否在某个范围内
func (c *RangeCond) New(ctx context.Context, cfg *interfaces.FilterCondCfg,
	fieldsMap map[string]*interfaces.Property) (interfaces.FilterCondition, error) {

	if cfg.Name == "" {
		return nil, fmt.Errorf("condition [range] left field is empty")
	}
	field, ok := fieldsMap[cfg.Name]
	if !ok {
		return nil, fmt.Errorf("condition [range] left field '%s' not found", cfg.Name)
	}
	if !interfaces.DataType_IsDate(field.Type) && !interfaces.DataType_IsNumber(field.Type) {
		return nil, fmt.Errorf("condition [range] left field is not a date/number field: %s:%s", cfg.Name, field.Type)
	}

	if cfg.ValueOptCfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [range] does not support value_from type '%s'", cfg.ValueFrom)
	}
	val, ok := cfg.ValueOptCfg.Value.([]any)
	if !ok {
		return nil, fmt.Errorf("condition [range] right value should be an array")
	}
	if len(val) != 2 {
		return nil, fmt.Errorf("condition [range] right value should be an array of length 2")
	}

	return &RangeCond{
		Cfg:    cfg,
		Lfield: field,
		Value:  val,
	}, nil
}

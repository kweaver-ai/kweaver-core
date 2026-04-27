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

type EqualCond struct {
	Cfg    *interfaces.FilterCondCfg
	Lfield *interfaces.Property
	Rfield *interfaces.Property
	Value  any
}

func (c *EqualCond) GetOperation() string { return OperationEqual }

func (c *EqualCond) SupportSubCond() bool       { return false }
func (c *EqualCond) NeedName() bool             { return true }
func (c *EqualCond) NeedValue() bool            { return true }
func (c *EqualCond) NeedConstValue() bool       { return false }
func (c *EqualCond) IsSingleValue() bool        { return true }
func (c *EqualCond) IsFixedLenArrayValue() bool { return false }
func (c *EqualCond) RequiredValueLen() int      { return -1 }

// eq 条件，判断字段是否等于右侧值
func (c *EqualCond) New(ctx context.Context, cfg *interfaces.FilterCondCfg,
	fieldsMap map[string]*interfaces.Property) (interfaces.FilterCondition, error) {

	if cfg.Name == "" {
		return nil, fmt.Errorf("condition [eq] left field is empty")
	}
	field, ok := fieldsMap[cfg.Name]
	if !ok {
		return nil, fmt.Errorf("condition [eq] field '%s' not found in schema", cfg.Name)
	}

	if IsSlice(cfg.ValueOptCfg.Value) {
		return nil, fmt.Errorf("condition [eq] only supports single value")
	}

	cond := &EqualCond{
		Cfg:    cfg,
		Lfield: field,
		Value:  cfg.ValueOptCfg.Value,
	}

	if cfg.ValueFrom == interfaces.ValueFrom_Field {
		valueStr, ok := cfg.Value.(string)
		if !ok {
			return nil, fmt.Errorf("condition [eq] right value should be a string field name")
		}
		rfield, ok := fieldsMap[valueStr]
		if !ok {
			return nil, fmt.Errorf("condition [eq] right field '%s' not found", valueStr)
		}
		cond.Rfield = rfield
	}

	return cond, nil
}

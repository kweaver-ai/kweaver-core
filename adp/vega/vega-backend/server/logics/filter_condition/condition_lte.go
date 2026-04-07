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

type LteCond struct {
	Cfg    *interfaces.FilterCondCfg
	Lfield *interfaces.Property
	Rfield *interfaces.Property
	Value  any
}

func (c *LteCond) GetOperation() string { return OperationLte }

func (c *LteCond) SupportSubCond() bool       { return false }
func (c *LteCond) NeedName() bool             { return true }
func (c *LteCond) NeedValue() bool            { return true }
func (c *LteCond) NeedConstValue() bool       { return false }
func (c *LteCond) IsSingleValue() bool        { return true }
func (c *LteCond) IsFixedLenArrayValue() bool { return false }
func (c *LteCond) RequiredValueLen() int      { return -1 }

// lte 条件, 判断字段是否小于等于某个值
func (c *LteCond) New(ctx context.Context, cfg *interfaces.FilterCondCfg,
	fieldsMap map[string]*interfaces.Property) (interfaces.FilterCondition, error) {

	if cfg.Name == "" {
		return nil, fmt.Errorf("condition [lte] left field is empty")
	}
	field, ok := fieldsMap[cfg.Name]
	if !ok {
		return nil, fmt.Errorf("condition [lte] left field '%s' not found", cfg.Name)
	}

	if IsSlice(cfg.ValueOptCfg.Value) {
		return nil, fmt.Errorf("condition [lte] only supports single value")
	}

	cond := &LteCond{
		Cfg:    cfg,
		Lfield: field,
		Value:  cfg.ValueOptCfg.Value,
	}

	if cfg.ValueFrom == interfaces.ValueFrom_Field {
		valueStr, ok := cfg.Value.(string)
		if !ok {
			return nil, fmt.Errorf("condition [lte] right value should be a string field name")
		}
		rfield, ok := fieldsMap[valueStr]
		if !ok {
			return nil, fmt.Errorf("condition [lte] right field '%s' not found", valueStr)
		}
		cond.Rfield = rfield
	}

	return cond, nil
}

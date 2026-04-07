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

type BeforeCond struct {
	Cfg    *interfaces.FilterCondCfg
	Lfield *interfaces.Property
	Value  []any
}

func (c *BeforeCond) GetOperation() string { return OperationBefore }

func (c *BeforeCond) SupportSubCond() bool       { return false }
func (c *BeforeCond) NeedName() bool             { return true }
func (c *BeforeCond) NeedValue() bool            { return true }
func (c *BeforeCond) NeedConstValue() bool       { return true }
func (c *BeforeCond) IsSingleValue() bool        { return false }
func (c *BeforeCond) IsFixedLenArrayValue() bool { return true }
func (c *BeforeCond) RequiredValueLen() int      { return 2 }

// before 条件，判断字段是否在某个时间之前
func (c *BeforeCond) New(ctx context.Context, cfg *interfaces.FilterCondCfg,
	fieldsMap map[string]*interfaces.Property) (interfaces.FilterCondition, error) {

	if cfg.Name == "" {
		return nil, fmt.Errorf("condition [before] left field is empty")
	}
	field, ok := fieldsMap[cfg.Name]
	if !ok {
		return nil, fmt.Errorf("condition [before] left field '%s' not found", cfg.Name)
	}
	if !interfaces.DataType_IsDate(field.Type) {
		return nil, fmt.Errorf("condition [before] left field is not a date field: %s:%s", cfg.Name, field.Type)
	}

	if cfg.ValueOptCfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [before] does not support value_from type '%s'", cfg.ValueFrom)
	}
	val, ok := cfg.ValueOptCfg.Value.([]any)
	if !ok {
		return nil, fmt.Errorf("condition [before] right value should be an array")
	}
	if len(val) != 2 {
		return nil, fmt.Errorf("condition [before] right value should be an array of length 2")
	}
	if _, ok := val[0].(int); ok {
		return nil, fmt.Errorf("condition [before]'s interval value should be an number")
	}
	if _, ok = val[1].(string); !ok {
		return nil, fmt.Errorf("condition [before]'s interval value should be a string")
	}

	return &BeforeCond{
		Cfg:    cfg,
		Lfield: field,
		Value:  val,
	}, nil
}

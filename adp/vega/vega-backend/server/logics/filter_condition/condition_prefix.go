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

type PrefixCond struct {
	Cfg    *interfaces.FilterCondCfg
	Lfield *interfaces.Property
	Value  string
}

func (c *PrefixCond) GetOperation() string { return OperationPrefix }

func (c *PrefixCond) SupportSubCond() bool       { return false }
func (c *PrefixCond) NeedName() bool             { return true }
func (c *PrefixCond) NeedValue() bool            { return true }
func (c *PrefixCond) NeedConstValue() bool       { return true }
func (c *PrefixCond) IsSingleValue() bool        { return true }
func (c *PrefixCond) IsFixedLenArrayValue() bool { return false }
func (c *PrefixCond) RequiredValueLen() int      { return -1 }

// prefix 条件, 判断字段是否以某个前缀开头
func (c *PrefixCond) New(ctx context.Context, cfg *interfaces.FilterCondCfg,
	fieldsMap map[string]*interfaces.Property) (interfaces.FilterCondition, error) {

	if cfg.Name == "" {
		return nil, fmt.Errorf("condition [prefix] left field is empty")
	}
	field, ok := fieldsMap[cfg.Name]
	if !ok {
		return nil, fmt.Errorf("condition [prefix] left field '%s' not found", cfg.Name)
	}
	if !interfaces.DataType_IsString(field.Type) {
		return nil, fmt.Errorf("condition [prefix] left field '%s' is not a string/text field", cfg.Name)
	}

	if cfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [prefix] does not support value_from type '%s'", cfg.ValueFrom)
	}
	val, ok := cfg.Value.(string)
	if !ok {
		return nil, fmt.Errorf("condition [prefix] right value is not a string value: %v", cfg.Value)
	}

	return &PrefixCond{
		Cfg:    cfg,
		Lfield: field,
		Value:  val,
	}, nil
}

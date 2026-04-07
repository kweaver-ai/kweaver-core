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

type NotPrefixCond struct {
	Cfg    *interfaces.FilterCondCfg
	Lfield *interfaces.Property
	Value  string
}

func (c *NotPrefixCond) GetOperation() string { return OperationNotPrefix }

func (c *NotPrefixCond) SupportSubCond() bool       { return false }
func (c *NotPrefixCond) NeedName() bool             { return true }
func (c *NotPrefixCond) NeedValue() bool            { return true }
func (c *NotPrefixCond) NeedConstValue() bool       { return true }
func (c *NotPrefixCond) IsSingleValue() bool        { return true }
func (c *NotPrefixCond) IsFixedLenArrayValue() bool { return false }
func (c *NotPrefixCond) RequiredValueLen() int      { return -1 }

// not_prefix 条件, 判断字段是否不以某个前缀开头
func (c *NotPrefixCond) New(ctx context.Context, cfg *interfaces.FilterCondCfg,
	fieldsMap map[string]*interfaces.Property) (interfaces.FilterCondition, error) {

	if cfg.Name == "" {
		return nil, fmt.Errorf("condition [not_prefix] left field is empty")
	}
	field, ok := fieldsMap[cfg.Name]
	if !ok {
		return nil, fmt.Errorf("condition [not_prefix] left field '%s' not found", cfg.Name)
	}
	if !interfaces.DataType_IsString(field.Type) {
		return nil, fmt.Errorf("condition [not_prefix] left field '%s' is not a string/text field", cfg.Name)
	}

	if cfg.ValueOptCfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [not_prefix] does not support value_from type '%s'", cfg.ValueFrom)
	}
	val, ok := cfg.ValueOptCfg.Value.(string)
	if !ok {
		return nil, fmt.Errorf("condition [not_prefix] right value is not a string value: %v", cfg.Value)
	}

	return &NotPrefixCond{
		Cfg:    cfg,
		Lfield: field,
		Value:  val,
	}, nil
}

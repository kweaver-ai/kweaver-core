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

type LikeCond struct {
	Cfg    *interfaces.FilterCondCfg
	Lfield *interfaces.Property
	Value  string
}

func (c *LikeCond) GetOperation() string { return OperationLike }

func (c *LikeCond) SupportSubCond() bool       { return false }
func (c *LikeCond) NeedName() bool             { return true }
func (c *LikeCond) NeedValue() bool            { return true }
func (c *LikeCond) NeedConstValue() bool       { return true }
func (c *LikeCond) IsSingleValue() bool        { return true }
func (c *LikeCond) IsFixedLenArrayValue() bool { return false }
func (c *LikeCond) RequiredValueLen() int      { return -1 }

// like 条件, 判断字段是否匹配某个字符串模式
func (c *LikeCond) New(ctx context.Context, cfg *interfaces.FilterCondCfg,
	fieldsMap map[string]*interfaces.Property) (interfaces.FilterCondition, error) {

	if cfg.Name == "" {
		return nil, fmt.Errorf("condition [like] left field is empty")
	}
	field, ok := fieldsMap[cfg.Name]
	if !ok {
		return nil, fmt.Errorf("condition [like] left field '%s' not found", cfg.Name)
	}
	if !interfaces.DataType_IsString(field.Type) {
		return nil, fmt.Errorf("condition [like] left field '%s' is not a string field", cfg.Name)
	}

	if cfg.ValueOptCfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [like] does not support value_from type '%s'", cfg.ValueFrom)
	}
	val, ok := cfg.ValueOptCfg.Value.(string)
	if !ok {
		return nil, fmt.Errorf("condition [like] right value is not a string value: %v", cfg.Value)
	}

	return &LikeCond{
		Cfg:    cfg,
		Lfield: field,
		Value:  val,
	}, nil
}

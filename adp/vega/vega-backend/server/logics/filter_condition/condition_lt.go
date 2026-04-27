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

type LtCond struct {
	Cfg    *interfaces.FilterCondCfg
	Lfield *interfaces.Property
	Rfield *interfaces.Property
	Value  any
}

func (c *LtCond) GetOperation() string { return OperationLt }

func (c *LtCond) SupportSubCond() bool       { return false }
func (c *LtCond) NeedName() bool             { return true }
func (c *LtCond) NeedValue() bool            { return true }
func (c *LtCond) NeedConstValue() bool       { return false }
func (c *LtCond) IsSingleValue() bool        { return true }
func (c *LtCond) IsFixedLenArrayValue() bool { return false }
func (c *LtCond) RequiredValueLen() int      { return -1 }

// lt 条件, 判断字段是否小于某个值
func (c *LtCond) New(ctx context.Context, cfg *interfaces.FilterCondCfg,
	fieldsMap map[string]*interfaces.Property) (interfaces.FilterCondition, error) {

	if cfg.Name == "" {
		return nil, fmt.Errorf("condition [lt] left field is empty")
	}
	field, ok := fieldsMap[cfg.Name]
	if !ok {
		// 如果字段未在Schema中定义，创建一个临时的Property对象
		field = &interfaces.Property{
			Name:         cfg.Name,
			OriginalName: cfg.Name,
		}
	}

	if IsSlice(cfg.ValueOptCfg.Value) {
		return nil, fmt.Errorf("condition [lt] only supports single value")
	}

	cond := &LtCond{
		Cfg:    cfg,
		Lfield: field,
		Value:  cfg.ValueOptCfg.Value,
	}

	if cfg.ValueFrom == interfaces.ValueFrom_Field {
		valueStr, ok := cfg.Value.(string)
		if !ok {
			return nil, fmt.Errorf("condition [lt] right value should be a string field name")
		}
		rfield, ok := fieldsMap[valueStr]
		if !ok {
			return nil, fmt.Errorf("condition [lt] right field '%s' not found", valueStr)
		}
		cond.Rfield = rfield
	}

	return cond, nil
}

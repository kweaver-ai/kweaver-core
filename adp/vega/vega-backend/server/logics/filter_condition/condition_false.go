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

type FalseCond struct {
	Cfg    *interfaces.FilterCondCfg
	Lfield *interfaces.Property
}

func (c *FalseCond) GetOperation() string { return OperationFalse }

func (c *FalseCond) SupportSubCond() bool       { return false }
func (c *FalseCond) NeedName() bool             { return true }
func (c *FalseCond) NeedValue() bool            { return false }
func (c *FalseCond) NeedConstValue() bool       { return false }
func (c *FalseCond) IsSingleValue() bool        { return false }
func (c *FalseCond) IsFixedLenArrayValue() bool { return false }
func (c *FalseCond) RequiredValueLen() int      { return -1 }

// false 条件，判断字段是否为 false
func (c *FalseCond) New(ctx context.Context, cfg *interfaces.FilterCondCfg,
	fieldsMap map[string]*interfaces.Property) (interfaces.FilterCondition, error) {

	if cfg.Name == "" {
		return nil, fmt.Errorf("condition [false] left field is empty")
	}
	field, ok := fieldsMap[cfg.Name]
	if !ok {
		return nil, fmt.Errorf("condition [false] left field '%s' not found", cfg.Name)
	}
	if field.Type != interfaces.DataType_Boolean {
		return nil, fmt.Errorf("condition [false] left field is not a boolean field: %s:%s", cfg.Name, field.Type)
	}

	return &FalseCond{
		Cfg:    cfg,
		Lfield: field,
	}, nil
}

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

type NullCond struct {
	Cfg    *interfaces.FilterCondCfg
	Lfield *interfaces.Property
}

func (c *NullCond) GetOperation() string { return OperationNull }

func (c *NullCond) SupportSubCond() bool       { return false }
func (c *NullCond) NeedName() bool             { return true }
func (c *NullCond) NeedValue() bool            { return false }
func (c *NullCond) NeedConstValue() bool       { return false }
func (c *NullCond) IsSingleValue() bool        { return false }
func (c *NullCond) IsFixedLenArrayValue() bool { return false }
func (c *NullCond) RequiredValueLen() int      { return -1 }

// null 条件, 判断字段是否为空
func (c *NullCond) New(ctx context.Context, cfg *interfaces.FilterCondCfg,
	fieldsMap map[string]*interfaces.Property) (interfaces.FilterCondition, error) {

	if cfg.Name == "" {
		return nil, fmt.Errorf("condition [null] left field is empty")
	}
	field, ok := fieldsMap[cfg.Name]
	if !ok {
		return nil, fmt.Errorf("condition [null] left field '%s' not found", cfg.Name)
	}

	return &NullCond{
		Cfg:    cfg,
		Lfield: field,
	}, nil

}

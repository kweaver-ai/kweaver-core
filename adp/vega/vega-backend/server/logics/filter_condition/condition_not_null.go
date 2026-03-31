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

type NotNullCond struct {
	Cfg    *interfaces.FilterCondCfg
	Lfield *interfaces.Property
}

func (c *NotNullCond) GetOperation() string { return OperationNotNull }

func (c *NotNullCond) SupportSubCond() bool       { return false }
func (c *NotNullCond) NeedName() bool             { return true }
func (c *NotNullCond) NeedValue() bool            { return false }
func (c *NotNullCond) NeedConstValue() bool       { return false }
func (c *NotNullCond) IsSingleValue() bool        { return false }
func (c *NotNullCond) IsFixedLenArrayValue() bool { return false }
func (c *NotNullCond) RequiredValueLen() int      { return -1 }

// not_null 条件, 判断字段是否不为空
func (c *NotNullCond) New(ctx context.Context, cfg *interfaces.FilterCondCfg,
	fieldsMap map[string]*interfaces.Property) (interfaces.FilterCondition, error) {

	if cfg.Name == "" {
		return nil, fmt.Errorf("condition [not_null] left field is empty")
	}
	field, ok := fieldsMap[cfg.Name]
	if !ok {
		return nil, fmt.Errorf("condition [not_null] left field '%s' not found", cfg.Name)
	}

	return &NotNullCond{
		Cfg:    cfg,
		Lfield: field,
	}, nil

}

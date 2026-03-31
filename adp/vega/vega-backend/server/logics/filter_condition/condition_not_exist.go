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

type NotExistCond struct {
	Cfg    *interfaces.FilterCondCfg
	Lfield *interfaces.Property
}

func (c *NotExistCond) GetOperation() string { return OperationNotExist }

func (c *NotExistCond) SupportSubCond() bool       { return false }
func (c *NotExistCond) NeedName() bool             { return true }
func (c *NotExistCond) NeedValue() bool            { return false }
func (c *NotExistCond) NeedConstValue() bool       { return false }
func (c *NotExistCond) IsSingleValue() bool        { return false }
func (c *NotExistCond) IsFixedLenArrayValue() bool { return false }
func (c *NotExistCond) RequiredValueLen() int      { return -1 }

func (c *NotExistCond) New(ctx context.Context, cfg *interfaces.FilterCondCfg,
	fieldsMap map[string]*interfaces.Property) (interfaces.FilterCondition, error) {

	if cfg.Name == "" {
		return nil, fmt.Errorf("condition [not_exist] left field is empty")
	}
	field, ok := fieldsMap[cfg.Name]
	if !ok {
		return nil, fmt.Errorf("condition [not_exist] left field '%s' not found", cfg.Name)
	}

	return &NotExistCond{
		Cfg:    cfg,
		Lfield: field,
	}, nil
}

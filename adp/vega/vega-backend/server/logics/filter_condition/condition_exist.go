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

type ExistCond struct {
	Cfg    *interfaces.FilterCondCfg
	Lfield *interfaces.Property
}

func (c *ExistCond) GetOperation() string { return OperationExist }

func (c *ExistCond) SupportSubCond() bool       { return false }
func (c *ExistCond) NeedName() bool             { return true }
func (c *ExistCond) NeedValue() bool            { return false }
func (c *ExistCond) NeedConstValue() bool       { return false }
func (c *ExistCond) IsSingleValue() bool        { return false }
func (c *ExistCond) IsFixedLenArrayValue() bool { return false }
func (c *ExistCond) RequiredValueLen() int      { return -1 }

// 存在 exist，判断字段是否存在
func (c *ExistCond) New(ctx context.Context, cfg *interfaces.FilterCondCfg,
	fieldsMap map[string]*interfaces.Property) (interfaces.FilterCondition, error) {

	if cfg.Name == "" {
		return nil, fmt.Errorf("condition [exist] left field is empty")
	}
	field, ok := fieldsMap[cfg.Name]
	if !ok {
		return nil, fmt.Errorf("condition [exist] left field '%s' not found", cfg.Name)
	}

	return &ExistCond{
		Cfg:    cfg,
		Lfield: field,
	}, nil
}

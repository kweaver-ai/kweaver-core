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

type TrueCond struct {
	Cfg    *interfaces.FilterCondCfg
	Lfield *interfaces.Property
}

func (c *TrueCond) GetOperation() string { return OperationTrue }

func (c *TrueCond) SupportSubCond() bool       { return false }
func (c *TrueCond) NeedName() bool             { return true }
func (c *TrueCond) NeedValue() bool            { return false }
func (c *TrueCond) NeedConstValue() bool       { return false }
func (c *TrueCond) IsSingleValue() bool        { return false }
func (c *TrueCond) IsFixedLenArrayValue() bool { return false }
func (c *TrueCond) RequiredValueLen() int      { return -1 }

// bool 类型为真
func (c *TrueCond) New(ctx context.Context, cfg *interfaces.FilterCondCfg,
	fieldsMap map[string]*interfaces.Property) (interfaces.FilterCondition, error) {
	if cfg.Name == "" {
		return nil, fmt.Errorf("condition [true] left field is empty")
	}
	field, ok := fieldsMap[cfg.Name]
	if !ok {
		return nil, fmt.Errorf("condition [true] left field '%s' not found", cfg.Name)
	}
	if field.Type != interfaces.DataType_Boolean {
		return nil, fmt.Errorf("condition [true] left field is not a boolean field: %s:%s", cfg.Name, field.Type)
	}

	return &TrueCond{
		Cfg:    cfg,
		Lfield: field,
	}, nil
}

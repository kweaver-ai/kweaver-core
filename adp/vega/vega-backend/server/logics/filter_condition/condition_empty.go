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

type EmptyCond struct {
	Cfg    *interfaces.FilterCondCfg
	Lfield *interfaces.Property
}

func (c *EmptyCond) GetOperation() string { return OperationEmpty }

func (c *EmptyCond) SupportSubCond() bool       { return false }
func (c *EmptyCond) NeedName() bool             { return true }
func (c *EmptyCond) NeedValue() bool            { return false }
func (c *EmptyCond) NeedConstValue() bool       { return false }
func (c *EmptyCond) IsSingleValue() bool        { return false }
func (c *EmptyCond) IsFixedLenArrayValue() bool { return false }
func (c *EmptyCond) RequiredValueLen() int      { return -1 }

// empty 条件，判断字段是否为空字符串
func (c *EmptyCond) New(ctx context.Context, cfg *interfaces.FilterCondCfg,
	fieldsMap map[string]*interfaces.Property) (interfaces.FilterCondition, error) {

	if cfg.Name == "" {
		return nil, fmt.Errorf("condition [empty] left field is empty")
	}
	field, ok := fieldsMap[cfg.Name]
	if !ok {
		return nil, fmt.Errorf("condition [empty] left field '%s' not found", cfg.Name)
	}
	// 只允许字符串类型
	if !interfaces.DataType_IsString(field.Type) {
		return nil, fmt.Errorf("condition [empty] left field %s is not of string type, but %s", cfg.Name, field.Type)
	}

	return &EmptyCond{
		Cfg:    cfg,
		Lfield: field,
	}, nil

}

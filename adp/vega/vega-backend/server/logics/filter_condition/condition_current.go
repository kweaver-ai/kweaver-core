// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package filter_condition

import (
	"context"
	"errors"
	"fmt"

	"vega-backend/interfaces"
)

var (
	CurrentYear   = "year"
	CurrentMonth  = "month"
	CurrentWeek   = "week"
	CurrentDay    = "day"
	CurrentHour   = "hour"
	CurrentMinute = "minute"

	CurrentFormatMap = map[string]bool{
		CurrentYear:   true,
		CurrentMonth:  true,
		CurrentWeek:   true,
		CurrentDay:    true,
		CurrentHour:   true,
		CurrentMinute: true,
	}
)

type CurrentCond struct {
	Cfg    *interfaces.FilterCondCfg
	Lfield *interfaces.Property
	Value  string
}

func (c *CurrentCond) GetOperation() string { return OperationCurrent }

func (c *CurrentCond) SupportSubCond() bool       { return false }
func (c *CurrentCond) NeedName() bool             { return true }
func (c *CurrentCond) NeedValue() bool            { return true }
func (c *CurrentCond) NeedConstValue() bool       { return true }
func (c *CurrentCond) IsSingleValue() bool        { return true }
func (c *CurrentCond) IsFixedLenArrayValue() bool { return false }
func (c *CurrentCond) RequiredValueLen() int      { return 1 }

// 当前时间 current，判断字段是否为当前时间，时间格式为 "%Y-%m-%d %H:%i"
func (c *CurrentCond) New(ctx context.Context, cfg *interfaces.FilterCondCfg,
	fieldsMap map[string]*interfaces.Property) (interfaces.FilterCondition, error) {

	if cfg.Name == "" {
		return nil, fmt.Errorf("condition [current] left field is empty")
	}
	field, ok := fieldsMap[cfg.Name]
	if !ok {
		return nil, fmt.Errorf("condition [current] left field '%s' not found", cfg.Name)
	}
	if !interfaces.DataType_IsDate(field.Type) {
		return nil, fmt.Errorf("condition [current] left field is not a date field: %s:%s", field.Name, field.Type)
	}

	if cfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [current] does not support value_from type '%s'", cfg.ValueFrom)
	}
	val, ok := cfg.Value.(string)
	if !ok {
		return nil, fmt.Errorf("condition [current] right value should be string")
	}
	if _, ok := CurrentFormatMap[val]; !ok {
		return nil, errors.New(`condition [current] right value should be 
		one of [` + CurrentYear + `, ` + CurrentMonth + `, ` + CurrentWeek + `, ` + CurrentDay + `, ` + CurrentHour + `, ` + CurrentMinute + `], actual is ` + val)
	}

	return &CurrentCond{
		Cfg:    cfg,
		Lfield: field,
		Value:  val,
	}, nil
}

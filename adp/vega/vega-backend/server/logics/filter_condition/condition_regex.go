// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package filter_condition

import (
	"context"
	"fmt"

	"github.com/dlclark/regexp2"

	"vega-backend/interfaces"
)

type RegexCond struct {
	Cfg    *interfaces.FilterCondCfg
	Lfield *interfaces.Property
	Value  string
	Regexp *regexp2.Regexp
}

func (c *RegexCond) GetOperation() string { return OperationRegex }

func (c *RegexCond) SupportSubCond() bool       { return false }
func (c *RegexCond) NeedName() bool             { return true }
func (c *RegexCond) NeedValue() bool            { return true }
func (c *RegexCond) NeedConstValue() bool       { return true }
func (c *RegexCond) IsSingleValue() bool        { return true }
func (c *RegexCond) IsFixedLenArrayValue() bool { return false }
func (c *RegexCond) RequiredValueLen() int      { return -1 }

// regex 条件, 判断字段是否匹配某个正则表达式
func (c *RegexCond) New(ctx context.Context, cfg *interfaces.FilterCondCfg,
	fieldsMap map[string]*interfaces.Property) (interfaces.FilterCondition, error) {

	if cfg.Name == "" {
		return nil, fmt.Errorf("condition [regex] left field is empty")
	}
	field, ok := fieldsMap[cfg.Name]
	if !ok {
		return nil, fmt.Errorf("condition [regex] left field '%s' not found", cfg.Name)
	}
	if !interfaces.DataType_IsString(field.Type) {
		return nil, fmt.Errorf("condition [regex] left field '%s' type must be string", cfg.Name)
	}

	if cfg.ValueOptCfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [regex] does not support value_from type '%s'", cfg.ValueFrom)
	}
	val, ok := cfg.ValueOptCfg.Value.(string)
	if !ok {
		return nil, fmt.Errorf("condition [regex] right value is not a string value: %v", cfg.Value)
	}
	regexp, err := regexp2.Compile(val, regexp2.RE2)
	if err != nil {
		return nil, fmt.Errorf("condition [regex] regular expression error: %s", err.Error())
	}

	return &RegexCond{
		Cfg:    cfg,
		Lfield: field,
		Value:  val,
		Regexp: regexp,
	}, nil
}

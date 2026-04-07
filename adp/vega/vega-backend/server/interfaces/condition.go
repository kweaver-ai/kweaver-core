// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import "context"

const (
	MaxSubCondition = 100

	AllField = "*"
)

const (
	ValueFrom_Const = "const"
	ValueFrom_Field = "field"
	ValueFrom_User  = "user"
)

type ValueOptCfg struct {
	ValueFrom string `json:"value_from,omitempty" mapstructure:"value_from"`
	Value     any    `json:"value,omitempty" mapstructure:"value"`
	RealValue any    `json:"real_value,omitempty" mapstructure:"real_value"`
}

type FilterCondCfg struct {
	Name        string           `json:"field,omitempty" mapstructure:"field"` // 传递name
	Operation   string           `json:"operation,omitempty" mapstructure:"operation"`
	SubConds    []*FilterCondCfg `json:"sub_conditions,omitempty" mapstructure:"sub_conditions"`
	ValueOptCfg `mapstructure:",squash"`

	RemainCfg map[string]any `mapstructure:",remain"`
}

//go:generate mockgen -source ../interfaces/condition.go -destination ../interfaces/mock/mock_condition.go
type FilterCondition interface {
	GetOperation() string

	SupportSubCond() bool
	NeedName() bool
	NeedValue() bool
	NeedConstValue() bool
	IsSingleValue() bool
	IsFixedLenArrayValue() bool
	RequiredValueLen() int

	New(ctx context.Context, cfg *FilterCondCfg, fieldsMap map[string]*Property) (cond FilterCondition, err error)
}

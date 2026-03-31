// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package condition

import (
	"context"
	"fmt"

	"github.com/dlclark/regexp2"

	dtype "ontology-query/interfaces/data_type"
)

type RegexCond struct {
	mCfg             *CondCfg
	mValue           string
	mRegexp          *regexp2.Regexp
	mFilterFieldName string
}

func NewRegexCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*DataProperty) (Condition, error) {
	_, ok := fieldsMap[cfg.Name]
	if !ok {
		return nil, fmt.Errorf("condition config field name '%s' must in view original fields", cfg.Name)
	}

	if !dtype.DataType_IsString(cfg.NameField.Type) {
		return nil, fmt.Errorf("condition [regex] left field is not a string field: %s:%s", cfg.NameField.Name, cfg.NameField.Type)
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
		mCfg:             cfg,
		mValue:           val,
		mRegexp:          regexp,
		mFilterFieldName: getFilterFieldName(cfg.Name, fieldsMap, false),
	}, nil
}

func (cond *RegexCond) Convert(ctx context.Context, vectorizer func(ctx context.Context, property *DataProperty, word string) ([]VectorResp, error)) (string, error) {
	dslStr := fmt.Sprintf(`
					{
						"regexp": {
							"%s": "%v"
						}
					}`, cond.mFilterFieldName, cond.mValue)

	return dslStr, nil
}

func (cond *RegexCond) Convert2SQL(ctx context.Context) (string, error) {
	return "", nil
}

func rewriteRegexCond(cfg *CondCfg) (*CondCfg, error) {

	// 过滤条件中的属性字段换成映射的视图字段
	if cfg.NameField.Name == "" {
		return nil, fmt.Errorf("正则过滤[regex]操作符使用的过滤字段[%s]在对象类的属性中不存在", cfg.Name)
	}
	return &CondCfg{
		Name:        cfg.NameField.MappedField.Name,
		Operation:   cfg.Operation,
		ValueOptCfg: cfg.ValueOptCfg,
	}, nil
}

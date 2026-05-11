// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package condition

import (
	"context"
	"fmt"

	"github.com/dlclark/regexp2"

	dtype "bkn-backend/interfaces/data_type"
)

type RegexCond struct {
	mCfg             *CondCfg
	mValue           string
	mRegexp          *regexp2.Regexp
	mFilterFieldName string
}

func NewRegexCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*FieldCfg) (Condition, error) {
	if !dtype.DataType_IsString(cfg.FieldCfg.Type) {
		return nil, fmt.Errorf("condition [regex] left field is not a string field: %s:%s", cfg.FieldCfg.Name, cfg.FieldCfg.Type)
	}

	if cfg.ValueFrom != ValueFrom_Const {
		return nil, fmt.Errorf("condition [regex] does not support value_from type '%s'", cfg.ValueFrom)
	}

	val, ok := cfg.Value.(string)
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
		mFilterFieldName: getFilterFieldName(cfg.Field, fieldsMap, false),
	}, nil
}

func (cond *RegexCond) Convert(ctx context.Context, vectorizer func(ctx context.Context, words []string) ([]*VectorResp, error)) (string, error) {
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

// convertRegexCondToDatasetFilterCondition converts RegexCond to dataset filter condition format
func convertRegexCondToDatasetFilterCondition(cfg *CondCfg) (map[string]any, error) {
	return map[string]any{
		"field":      cfg.Field,
		"operation":  "regex",
		"value":      cfg.Value,
		"value_from": "const",
	}, nil
}

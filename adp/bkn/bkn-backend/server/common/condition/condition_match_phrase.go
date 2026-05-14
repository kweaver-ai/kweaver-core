// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package condition

import (
	"context"
	"fmt"

	"github.com/bytedance/sonic"
)

type MatchPhraseCond struct {
	mCfg              *CondCfg
	mFilterFieldNames []string
}

func NewMatchPhraseCond(ctx context.Context, cfg *CondCfg, fieldScope uint8, fieldsMap map[string]*ViewField) (Condition, error) {
	if cfg.ValueFrom != ValueFrom_Const {
		return nil, fmt.Errorf("condition [match_phrase] does not support value_from type '%s'", cfg.ValueFrom)
	}

	name := getFilterFieldName(cfg.Field, fieldsMap, true)
	var fields []string
	// 如果指定*查询，并且视图的字段范围为部分字段，那么将查询的字段替换成视图的字段列表
	if name == AllField {
		// fields = make([]string, 0, len(fieldsMap))
		// for fieldName := range fieldsMap {
		// 	fields = append(fields, fieldName)
		// }
		fields = []string{
			"id",
			"name",
			"comment",
			"detail",
			"data_properties.name",
			"data_properties.display_name",
			"data_properties.comment",
			"logic_properties.name",
			"logic_properties.display_name",
			"logic_properties.comment",
		}
	} else {
		fields = append(fields, name)
	}

	return &MatchPhraseCond{
		mCfg:              cfg,
		mFilterFieldNames: fields,
	}, nil
}

func (cond *MatchPhraseCond) Convert(ctx context.Context, vectorizer func(ctx context.Context, words []string) ([]*VectorResp, error)) (string, error) {
	v := cond.mCfg.Value
	vStr, ok := v.(string)
	if ok {
		v = fmt.Sprintf("%q", vStr)
	}

	fields, err := sonic.Marshal(cond.mFilterFieldNames)
	if err != nil {
		return "", fmt.Errorf("condition [match_phrase] marshal fields error: %s", err.Error())
	}

	dslStr := fmt.Sprintf(`
					{
						"multi_match": {
							"query": %v,
							"type": "phrase",
							"fields": %v
						}
					}`, v, string(fields))

	return dslStr, nil
}

func (cond *MatchPhraseCond) Convert2SQL(ctx context.Context) (string, error) {
	return "", nil
}

// convertMatchPhraseCondToDatasetFilterCondition converts MatchPhraseCond to dataset filter condition format
func convertMatchPhraseCondToDatasetFilterCondition(cfg *CondCfg, fieldsMap map[string]*ViewField) (map[string]any, error) {
	name := getFilterFieldName(cfg.Field, fieldsMap, true)
	var fields []string
	// 如果指定*查询，并且视图的字段范围为部分字段，那么将查询的字段替换成视图的字段列表
	if name == AllField {
		// fields = make([]string, 0, len(fieldsMap))
		// for fieldName := range fieldsMap {
		// 	fields = append(fields, fieldName)
		// }
		fields = []string{
			"id",
			"name",
			"comment",
			"detail",
			"data_properties.name",
			"data_properties.display_name",
			"data_properties.comment",
			"logic_properties.name",
			"logic_properties.display_name",
			"logic_properties.comment",
		}
	} else {
		fields = append(fields, name)
	}

	return map[string]any{
		// "field":      "*",
		"fields":     fields,
		"operation":  "match_phrase",
		"value":      cfg.Value,
		"value_from": "const",
	}, nil
}

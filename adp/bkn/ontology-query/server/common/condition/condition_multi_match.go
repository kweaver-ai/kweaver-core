// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package condition

import (
	"context"
	"fmt"
	"ontology-query/common"
	dtype "ontology-query/interfaces/data_type"

	"github.com/bytedance/sonic"
)

type MultiMatchCond struct {
	mCfg              *CondCfg
	mFilterFieldNames []string
}

func NewMultiMatchCond(ctx context.Context, cfg *CondCfg, fieldScope uint8, fieldsMap map[string]*DataProperty) (Condition, error) {

	// 从cfg的 ReaminCfg 中获取 fields，这是属于 multi_match的fields字段，是个字符串数组，
	// 如果想要全部字段匹配，可不填或者填 ["*"], 不支持填字符串 *， 需要一个数组
	var fields []string
	cfgFields, exist := cfg.RemainCfg["fields"]
	if exist {
		// 存在 fields 时需要是一个数组
		if !common.IsSlice(cfgFields) {
			return nil, fmt.Errorf("multi_match：remain_cfg.fields 必须是数组（例如 [\"attr_a\"] 或 [\"*\"]）")
		}
		// 字段数组里的需要是个字符串数组
		for _, cfgField := range cfgFields.([]any) {
			field, ok := cfgField.(string)
			if !ok {
				return nil, fmt.Errorf("multi_match：fields 数组元素须全部为字符串，当前存在非字符串元素：%v", cfgField)
			}

			if field == AllField {
				expanded, err := expandIndexFieldNamesForMultiMatchStar(fieldsMap)
				if err != nil {
					return nil, err
				}
				fields = append(fields, expanded...)
				continue
			}

			fieldInfo := fieldsMap[field]
			if fieldInfo == nil {
				return nil, fmt.Errorf(errFmtUnknownObjectTypeProperty, field)
			}
			name := getFilterFieldName(field, fieldsMap, true)

			if fieldInfo.Type == dtype.DATATYPE_TEXT {
				fields = append(fields, name)
				continue
			}
			if fieldInfo.Type == dtype.DATATYPE_STRING &&
				fieldInfo.IndexConfig != nil && fieldInfo.IndexConfig.FulltextConfig.Enabled {
				fields = append(fields, name+"."+dtype.TEXT_SUFFIX)
				continue
			}
			return nil, fmt.Errorf("multi_match：属性「%s」须为 text 类型，或已启用全文索引的 string 类型后才能用于全文检索，请检查索引配置", field)
		}
	}

	// 校验match_type的有效性, match_type可以为空
	matchType, exist := cfg.RemainCfg["match_type"]
	if exist && matchType != "" {
		mtype, ok := matchType.(string)
		if !ok {
			return nil, fmt.Errorf("multi_match：match_type 须为字符串，当前类型无效：%v", matchType)
		}
		if !MatchTypeMap[mtype] {
			return nil, fmt.Errorf("multi_match：match_type「%s」无效，可选值为：best_fields、most_fields、cross_fields、phrase、phrase_prefix、bool_prefix", mtype)
		}
	}

	return &MultiMatchCond{
		mCfg:              cfg,
		mFilterFieldNames: fields,
	}, nil
}

func (cond *MultiMatchCond) Convert(ctx context.Context, vectorizer func(ctx context.Context, property *DataProperty, word string) ([]VectorResp, error)) (string, error) {
	v := cond.mCfg.Value
	vStr, ok := v.(string)
	if ok {
		v = fmt.Sprintf("%q", vStr)
	}

	fields, err := sonic.Marshal(cond.mFilterFieldNames)
	if err != nil {
		return "", fmt.Errorf("multi_match：序列化检索字段列表失败：%s", err.Error())
	}

	// 默认是 best_fields
	matchType := "best_fields"
	if mt, ok := cond.mCfg.RemainCfg["match_type"]; ok {
		if mtStr, ok := mt.(string); ok {
			matchType = mtStr
		} else {
			return "", fmt.Errorf("multi_match：match_type 须为字符串，当前为 %v", mt)
		}
	}

	dslStr := fmt.Sprintf(`
					{
						"multi_match": {
							"query": %v,
							"type": "%s"`, v, matchType)

	// 如果不指定 fields，则用 index.query.default_field 配置的字段查询，默认是*
	if len(cond.mFilterFieldNames) > 0 {
		dslStr = fmt.Sprintf(`%s,
							"fields": %v
						}
					}`, dslStr, string(fields))
	} else {
		dslStr = fmt.Sprintf(`%s
						}
					}`, dslStr)
	}

	return dslStr, nil
}

func (cond *MultiMatchCond) Convert2SQL(ctx context.Context) (string, error) {
	return "", nil
}

func rewriteMultiMatchCond(cfg *CondCfg, fieldsMap map[string]*DataProperty) (*CondCfg, error) {

	// 过滤条件中的属性字段换成映射的视图字段
	// 从cfg的 ReaminCfg 中获取 fields，这是属于 multi_match的fields字段，是个字符串数组，
	// 如果想要全部字段匹配，可不填或者填 ["*"], 不支持填字符串 *， 需要一个数组
	var fields []string
	cfgFields, exist := cfg.RemainCfg["fields"]
	if exist {
		// 存在 fields 时需要是一个数组
		if !common.IsSlice(cfgFields) {
			return nil, fmt.Errorf("multi_match：remain_cfg.fields 必须是数组（例如 [\"attr_a\"] 或 [\"*\"]）")
		}
		// 字段数组里的需要是个字符串数组
		for _, cfgField := range cfgFields.([]any) {
			field, ok := cfgField.(string)
			if !ok {
				return nil, fmt.Errorf("multi_match：fields 数组元素须全部为字符串，当前存在非字符串元素：%v", cfgField)
			}

			if field == AllField {
				expanded, err := expandViewFieldNamesForMultiMatchStar(fieldsMap)
				if err != nil {
					return nil, err
				}
				fields = append(fields, expanded...)
				continue
			}

			fieldInfo, ok1 := fieldsMap[field]
			if !ok1 || fieldInfo == nil {
				return nil, fmt.Errorf(errFmtUnknownObjectTypeProperty, field)
			}
			if fieldInfo.MappedField.Name == "" {
				return nil, fmt.Errorf(errFmtMissingViewMappedField, field)
			}

			if fieldInfo.Type == dtype.DATATYPE_TEXT {
				fields = append(fields, fieldInfo.MappedField.Name)
				continue
			}
			if fieldInfo.Type == dtype.DATATYPE_STRING &&
				fieldInfo.IndexConfig != nil && fieldInfo.IndexConfig.FulltextConfig.Enabled {
				fields = append(fields, fieldInfo.MappedField.Name+"."+dtype.TEXT_SUFFIX)
				continue
			}
			return nil, fmt.Errorf("multi_match：属性「%s」须为 text 类型，或已启用全文索引的 string 类型后才能用于数据视图全文检索", field)
		}
	}

	// 校验match_type的有效性, match_type可以为空
	matchType, exist := cfg.RemainCfg["match_type"]
	if exist && matchType != "" {
		mtype, ok := matchType.(string)
		if !ok {
			return nil, fmt.Errorf("multi_match：match_type 须为字符串，当前类型无效：%v", matchType)
		}
		if !MatchTypeMap[mtype] {
			return nil, fmt.Errorf("multi_match：match_type「%s」无效，可选值为：best_fields、most_fields、cross_fields、phrase、phrase_prefix、bool_prefix", mtype)
		}
	}

	return &CondCfg{
		RemainCfg: map[string]any{
			"fields":     fields,
			"match_type": matchType,
		},
		Operation:   cfg.Operation,
		ValueOptCfg: cfg.ValueOptCfg,
	}, nil
}

// expandIndexFieldNamesForMultiMatchStar resolves ["*"] into OpenSearch field names for text and
// fulltext-enabled string properties (uses .text subfield for the latter).
func expandIndexFieldNamesForMultiMatchStar(fieldsMap map[string]*DataProperty) ([]string, error) {
	var out []string
	for _, fieldInfo := range fieldsMap {
		if fieldInfo == nil {
			continue
		}
		if fieldInfo.Type == dtype.DATATYPE_TEXT {
			out = append(out, getFilterFieldName(fieldInfo.Name, fieldsMap, true))
			continue
		}
		if fieldInfo.Type == dtype.DATATYPE_STRING &&
			fieldInfo.IndexConfig != nil && fieldInfo.IndexConfig.FulltextConfig.Enabled {
			base := getFilterFieldName(fieldInfo.Name, fieldsMap, true)
			out = append(out, base+"."+dtype.TEXT_SUFFIX)
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("multi_match：当 fields 包含「*」时，对象类中至少需要存在一个 text 类型属性，或已启用全文索引的 string 类型属性，请检查模型与索引配置")
	}
	return out, nil
}

// expandViewFieldNamesForMultiMatchStar resolves ["*"] into view column names (MappedField.Name)
// for the same property kinds as the index path, skipping properties without mapped_field.
func expandViewFieldNamesForMultiMatchStar(fieldsMap map[string]*DataProperty) ([]string, error) {
	var out []string
	for _, fieldInfo := range fieldsMap {
		if fieldInfo == nil || fieldInfo.MappedField.Name == "" {
			continue
		}
		if fieldInfo.Type == dtype.DATATYPE_TEXT {
			out = append(out, fieldInfo.MappedField.Name)
			continue
		}
		if fieldInfo.Type == dtype.DATATYPE_STRING &&
			fieldInfo.IndexConfig != nil && fieldInfo.IndexConfig.FulltextConfig.Enabled {
			out = append(out, fieldInfo.MappedField.Name+"."+dtype.TEXT_SUFFIX)
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("multi_match：当 fields 包含「*」时，至少需要存在一个可全文检索且已配置视图映射列(mapped_field)的数据属性")
	}
	return out, nil
}

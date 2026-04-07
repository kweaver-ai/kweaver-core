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
			return nil, fmt.Errorf("condition [multi_match] 'fields' value should be an array")
		}
		// 字段数组里的需要是个字符串数组
		for _, cfgField := range cfgFields.([]any) {
			field, ok := cfgField.(string)
			if !ok {
				return nil, fmt.Errorf("condition [multi_match] 'fields' value should be a string array, contain non string value[%v]", cfgField)
			}

			// 字段数组里的每个元素都需要是字符串
			name := getFilterFieldName(field, fieldsMap, true)
			// 如果指定*查询，并且视图的字段范围为部分字段，那么将查询的字段替换成视图的字段列表
			if name == AllField {
				// * 只针对text字段和配了全文索引的属性做全文检索
				for _, fieldInfo := range fieldsMap {
					if fieldInfo.Type == dtype.DATATYPE_TEXT {
						// text字段直接拼
						fields = append(fields, name)
					} else {
						if fieldInfo.Type == dtype.DATATYPE_STRING &&
							fieldInfo.IndexConfig != nil && fieldInfo.IndexConfig.FulltextConfig.Enabled {
							// 配置了全文索引的属性,可以做match查询,否则报错,不能进行match查询
							// string 类型做了fulltext, 则match用 xxx.text 进行过滤
							fields = append(fields, name+"."+dtype.TEXT_SUFFIX)
						}
					}
				}
			} else {
				// 字段是否做了全文索引
				fieldInfo := fieldsMap[name]
				if fieldInfo.Type == dtype.DATATYPE_TEXT {
					// text字段直接拼
					fields = append(fields, name)
				} else {
					if fieldInfo.Type == dtype.DATATYPE_STRING &&
						fieldInfo.IndexConfig != nil && fieldInfo.IndexConfig.FulltextConfig.Enabled {
						// 配置了全文索引的属性,可以做match查询,否则报错,不能进行match查询
						// string 类型做了fulltext, 则match用 xxx.text 进行过滤
						fields = append(fields, name+"."+dtype.TEXT_SUFFIX)
					} else {
						return nil, fmt.Errorf(`the index of property [%s] is not configured for full-text search and cannot be used for [multi_match] filtering. Please check the index configuration of the object type and the current request`, name)
					}
				}
			}
		}
	}

	// 校验match_type的有效性, match_type可以为空
	matchType, exist := cfg.RemainCfg["match_type"]
	if exist && matchType != "" {
		mtype, ok := matchType.(string)
		if !ok {
			return nil, fmt.Errorf("condition [multi_match] 'match_type' value should be a string, actual is[%v]", matchType)
		}
		if !MatchTypeMap[mtype] {
			return nil, fmt.Errorf("condition [multi_match] 'match_type' value should be one of [%v], actual is[%v]", MatchTypeMap, mtype)
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
		return "", fmt.Errorf("condition [multi_match] marshal fields error: %s", err.Error())
	}

	// 默认是 best_fields
	matchType := "best_fields"
	if mt, exist := cond.mCfg.RemainCfg["match_type"]; exist {
		if mtStr, ok := mt.(string); exist && ok {
			matchType = mtStr
		} else {
			return "", fmt.Errorf("condition [multi_match] match_type[%v] should be a string", mt)
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
			return nil, fmt.Errorf("condition [multi_match] 'fields' value should be an array")
		}
		// 字段数组里的需要是个字符串数组
		for _, cfgField := range cfgFields.([]any) {
			field, ok := cfgField.(string)
			if !ok {
				return nil, fmt.Errorf("condition [multi_match] 'fields' value should be a string array, contain non string value[%v]", cfgField)
			}

			if field == AllField {
				fields = append(fields, AllField)
			} else {
				// 从属性集中取属性配置
				fieldInfo, ok1 := fieldsMap[field]
				if !ok1 {
					return nil, fmt.Errorf("全文匹配过滤[multi_match]操作符使用的过滤字段[%s]在对象类的属性中不存在", field)
				}
				// 从属性中获取映射的视图字段
				if fieldInfo.MappedField.Name == "" {
					return nil, fmt.Errorf("全文匹配过滤[multi_match]操作符使用的过滤字段[%s]映射的视图字段为空", field)
				}
				fields = append(fields, fieldInfo.MappedField.Name)
			}
		}
	}

	// 校验match_type的有效性, match_type可以为空
	matchType, exist := cfg.RemainCfg["match_type"]
	if exist && matchType != "" {
		mtype, ok := matchType.(string)
		if !ok {
			return nil, fmt.Errorf("condition [multi_match] 'match_type' value should be a string, actual is[%v]", matchType)
		}
		if !MatchTypeMap[mtype] {
			return nil, fmt.Errorf("condition [multi_match] 'match_type' value should be one of [%v], actual is[%v]", MatchTypeMap, mtype)
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

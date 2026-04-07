// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package condition

import (
	"context"
	"fmt"
	dtype "ontology-query/interfaces/data_type"

	"github.com/bytedance/sonic"
)

type MatchCond struct {
	mCfg              *CondCfg
	mFilterFieldNames []string
}

func NewMatchCond(ctx context.Context, cfg *CondCfg, fieldScope uint8, fieldsMap map[string]*DataProperty) (Condition, error) {

	name := getFilterFieldName(cfg.Name, fieldsMap, true)
	var fields []string
	// 如果指定*查询，并且视图的字段范围为部分字段，那么将查询的字段替换成视图的字段列表
	if name == AllField {
		// * 只针对text字段和配了全文索引的属性做全文检索
		for _, fieldInfo := range fieldsMap {
			fields = append(fields, fieldInfo.Name)
			// 如果是string类型且配置了全文索引，则字段名为 fieldInfo.Name+"."+dtype.TEXT_SUFFIX
			if fieldInfo.Type == dtype.DATATYPE_STRING &&
				fieldInfo.IndexConfig != nil && fieldInfo.IndexConfig.FulltextConfig.Enabled {
				fields = append(fields, fieldInfo.Name+"."+dtype.TEXT_SUFFIX)
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
				return nil, fmt.Errorf(`the index of property [%s] is not configured for full-text search and cannot be used for [match] filtering. Please check the index configuration of the object type and the current request`, name)
			}
		}
	}

	return &MatchCond{
		mCfg:              cfg,
		mFilterFieldNames: fields,
	}, nil
}

func (cond *MatchCond) Convert(ctx context.Context, vectorizer func(ctx context.Context, property *DataProperty, word string) ([]VectorResp, error)) (string, error) {
	v := cond.mCfg.Value
	vStr, ok := v.(string)
	if ok {
		v = fmt.Sprintf("%q", vStr)
	}

	fields, err := sonic.Marshal(cond.mFilterFieldNames)
	if err != nil {
		return "", fmt.Errorf("condition [match] marshal fields error: %s", err.Error())
	}

	dslStr := fmt.Sprintf(`
					{
						"multi_match": {
							"query": %v,
							"type": "best_fields",
							"fields": %v
						}
					}`, v, string(fields))

	return dslStr, nil
}

func (cond *MatchCond) Convert2SQL(ctx context.Context) (string, error) {
	return "", nil
}

func rewriteMatchCond(cfg *CondCfg) (*CondCfg, error) {

	// 过滤条件中的属性字段换成映射的视图字段
	fieldName := ""
	if cfg.Name == AllField {
		fieldName = AllField
	} else {
		if cfg.NameField.Name == "" {
			return nil, fmt.Errorf("全文匹配过滤[match]操作符使用的过滤字段[%s]在对象类的属性中不存在", cfg.Name)
		}
		fieldName = cfg.NameField.MappedField.Name
	}

	return &CondCfg{
		Name:        fieldName,
		Operation:   cfg.Operation,
		ValueOptCfg: cfg.ValueOptCfg,
	}, nil
}

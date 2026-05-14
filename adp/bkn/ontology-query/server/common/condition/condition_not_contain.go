// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package condition

import (
	"context"
	"fmt"
	"strings"
)

type NotContainCond struct {
	mCfg             *CondCfg
	IsSliceValue     bool
	mValue           any
	mSliceValue      []any
	mFilterFieldName string
}

// 不包含 not_contain，左侧属性值为数组，右侧值为单个值或数组，如果为数组，意味着数组内的值都应在属性值外
func NewNotContainCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*DataProperty) (Condition, error) {
	if cfg.ValueFrom != ValueFrom_Const {
		return nil, fmt.Errorf("condition [not_contain] does not support value_from type '%s'", cfg.ValueFrom)
	}

	notContainCond := &NotContainCond{
		mCfg:             cfg,
		mFilterFieldName: getFilterFieldName(cfg.Name, fieldsMap, false),
	}

	if IsSlice(cfg.Value) {
		val, ok := cfg.Value.([]any)
		if !ok {
			return nil, fmt.Errorf("condition [not_contain] right value is not a valid array")
		}
		if len(val) == 0 {
			return nil, fmt.Errorf("condition [not_contain] right value is an empty array")
		}

		notContainCond.IsSliceValue = true
		notContainCond.mSliceValue = val

	} else {
		notContainCond.IsSliceValue = false
		notContainCond.mValue = cfg.Value
	}

	return notContainCond, nil
}

/*
如果右侧为数组, 则生成如下dsl:

	{
	  "bool": {
	    "must_not": [
	      {
	        "term": {
	          "<field>": {
	            "value": <value1>
	          }
	        }
	      },
	      {
	        "term": {
	          "<field>": {
	            "value": <value2>
	          }
	        }
	      }
	    ]
	  }
	}

如果右侧为单个值, 则生成如下dsl:

	{
	  "bool": {
	    "must_not": {
	      "term": {
	        "%s": {
	          "value": <value>
	        }
	      }
	    }
	  }
	}
*/
func (cond *NotContainCond) Convert(ctx context.Context, vectorizer func(ctx context.Context, property *DataProperty, word string) ([]VectorResp, error)) (string, error) {
	var dslStr string
	if cond.IsSliceValue {
		subStrs := []string{}
		for _, val := range cond.mSliceValue {
			vStr, ok := val.(string)
			if ok {
				val = fmt.Sprintf("%q", vStr)
			}

			subStr := fmt.Sprintf(`
			{
				"term": {
					"%s": {
						"value": %v
					}
				}
			}`, cond.mFilterFieldName, val)

			subStrs = append(subStrs, subStr)

		}

		dslStr = fmt.Sprintf(`
		{
			"bool": {
				"must_not": [
					%s
				]
			}
		}
		`, strings.Join(subStrs, ","))

	} else {
		val := cond.mValue
		vStr, ok := val.(string)
		if ok {
			val = fmt.Sprintf("%q", vStr)
		}

		dslStr = fmt.Sprintf(`
		{
			"bool": {
				"must_not": {
					"term": {
						"%s": {
							"value": %v
						}
					}
				}
			}
		}`, cond.mFilterFieldName, val)
	}

	return dslStr, nil
}

func (cond *NotContainCond) Convert2SQL(ctx context.Context) (string, error) {
	// 使用json_array_contains函数配合NOT运算符实现not_contain操作
	// 左侧属性值为数组，右侧值为单个值或数组
	// 如果右侧为数组，意味着数组内的值都应在属性值外（即所有右侧值都不被包含）
	var sqlStr string

	if cond.IsSliceValue {
		// 右侧为数组，需要所有值都不在左侧数组中
		// 为每个值生成一个NOT json_array_contains条件，并用AND连接
		conditions := []string{}
		for _, val := range cond.mSliceValue {
			var condition string
			vStr, ok := val.(string)
			if ok {
				// 处理字符串值，转义单引号
				escapedVal := strings.ReplaceAll(vStr, "'", "''")
				condition = fmt.Sprintf(`NOT json_array_contains("%s", '%s')`, cond.mFilterFieldName, escapedVal)
			} else {
				// 处理非字符串值
				condition = fmt.Sprintf(`NOT json_array_contains("%s", %v)`, cond.mFilterFieldName, val)
			}
			conditions = append(conditions, condition)
		}

		// 使用AND连接所有条件，确保所有右侧值都不在左侧数组中
		sqlStr = strings.Join(conditions, " AND ")

	} else {
		// 右侧为单个值
		val := cond.mValue
		vStr, ok := val.(string)
		if ok {
			// 处理字符串值，转义单引号
			escapedVal := strings.ReplaceAll(vStr, "'", "''")
			sqlStr = fmt.Sprintf(`NOT json_array_contains("%s", '%s')`, cond.mFilterFieldName, escapedVal)
		} else {
			// 处理非字符串值
			sqlStr = fmt.Sprintf(`NOT json_array_contains("%s", %v)`, cond.mFilterFieldName, val)
		}
	}

	return sqlStr, nil
}

func rewriteNotContainCond(cfg *CondCfg) (*CondCfg, error) {
	// 过滤条件中的属性字段换成映射的视图字段
	if cfg.NameField.Name == "" {
		return nil, fmt.Errorf("不包含[not_contain]操作符使用的过滤字段[%s]在对象类的属性中不存在", cfg.Name)
	}
	return &CondCfg{
		Name:        cfg.NameField.MappedField.Name,
		Operation:   cfg.Operation,
		ValueOptCfg: cfg.ValueOptCfg,
	}, nil
}

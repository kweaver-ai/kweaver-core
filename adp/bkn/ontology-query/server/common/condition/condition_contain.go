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

type ContainCond struct {
	mCfg             *CondCfg
	IsSliceValue     bool
	mValue           any
	mSliceValue      []any
	mFilterFieldName string
}

// 包含 contain，左侧属性值为数组，右侧值为单个值或数组，如果为数组，意味着数组内的值都应在属性值内
func NewContainCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*DataProperty) (Condition, error) {
	if cfg.ValueOptCfg.ValueFrom != ValueFrom_Const {
		return nil, fmt.Errorf("condition [contain] does not support value_from type '%s'", cfg.ValueOptCfg.ValueFrom)
	}

	containCond := &ContainCond{
		mCfg:             cfg,
		mFilterFieldName: getFilterFieldName(cfg.Name, fieldsMap, false),
	}

	if IsSlice(cfg.ValueOptCfg.Value) {
		val, ok := cfg.ValueOptCfg.Value.([]any)
		if !ok {
			return nil, fmt.Errorf("condition [contain] right value is not a valid array")
		}
		if len(val) == 0 {
			return nil, fmt.Errorf("condition [contain] right value is an empty array")
		}

		containCond.IsSliceValue = true
		containCond.mSliceValue = val

	} else {
		containCond.IsSliceValue = false
		containCond.mValue = cfg.ValueOptCfg.Value
	}

	return containCond, nil
}

/*
如果右侧为数组，则生成如下dsl:

	{
	  "bool": {
	    "filter": [
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

如果右侧为单个值，则生成如下dsl:

	{
	  "term": {
	    "<field>": {
	      "value": <value>
	    }
	  }
	}
*/
func (cond *ContainCond) Convert(ctx context.Context, vectorizer func(ctx context.Context, property *DataProperty, word string) ([]VectorResp, error)) (string, error) {
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
				"filter": [
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
			"term": {
				"%s": {
					"value": %v
				}
			}
		}`, cond.mFilterFieldName, val)
	}

	return dslStr, nil
}

func (cond *ContainCond) Convert2SQL(ctx context.Context) (string, error) {
	// 使用json_array_contains函数实现contain操作
	// 左侧属性值为数组，右侧值为单个值或数组
	// 如果右侧为数组，意味着数组内的值都应在属性值内（即所有右侧值都需要被包含）
	var sqlStr string

	if cond.IsSliceValue {
		// 右侧为数组，需要所有值都在左侧数组中
		// 为每个值生成一个json_array_contains条件，并用AND连接
		conditions := []string{}
		for _, val := range cond.mSliceValue {
			var condition string
			vStr, ok := val.(string)
			if ok {
				// 处理字符串值，转义单引号
				escapedVal := strings.ReplaceAll(vStr, "'", "''")
				condition = fmt.Sprintf(`json_array_contains("%s", '%s')`, cond.mFilterFieldName, escapedVal)
			} else {
				// 处理非字符串值
				condition = fmt.Sprintf(`json_array_contains("%s", %v)`, cond.mFilterFieldName, val)
			}
			conditions = append(conditions, condition)
		}

		// 使用AND连接所有条件，确保所有右侧值都在左侧数组中
		sqlStr = strings.Join(conditions, " AND ")

	} else {
		// 右侧为单个值
		val := cond.mValue
		vStr, ok := val.(string)
		if ok {
			// 处理字符串值，转义单引号
			escapedVal := strings.ReplaceAll(vStr, "'", "''")
			sqlStr = fmt.Sprintf(`json_array_contains("%s", '%s')`, cond.mFilterFieldName, escapedVal)
		} else {
			// 处理非字符串值
			sqlStr = fmt.Sprintf(`json_array_contains("%s", %v)`, cond.mFilterFieldName, val)
		}
	}

	return sqlStr, nil
}

func rewriteContainCond(cfg *CondCfg) (*CondCfg, error) {
	// 过滤条件中的属性字段换成映射的视图字段
	if cfg.NameField.Name == "" {
		return nil, fmt.Errorf("包含[contain]操作符使用的过滤字段[%s]在对象类的属性中不存在", cfg.Name)
	}
	return &CondCfg{
		Name:        cfg.NameField.MappedField.Name,
		Operation:   cfg.Operation,
		ValueOptCfg: cfg.ValueOptCfg,
	}, nil
}

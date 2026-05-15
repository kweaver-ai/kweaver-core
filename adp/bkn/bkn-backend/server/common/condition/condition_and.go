// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package condition

import (
	"context"
	"fmt"
)

type AndCond struct {
	mCfg      *CondCfg
	mSubConds []Condition
}

func newAndCond(ctx context.Context, cfg *CondCfg, fieldScope uint8, fieldsMap map[string]*FieldCfg) (Condition, error) {
	subConds := []Condition{}

	if len(cfg.SubConds) == 0 {
		return nil, fmt.Errorf("sub condition size is 0")
	}

	if len(cfg.SubConds) > MaxSubCondition {
		return nil, fmt.Errorf("sub condition size limit %d but %d", MaxSubCondition, len(cfg.SubConds))
	}

	for _, subCond := range cfg.SubConds {
		cond, err := NewCondition(ctx, subCond, fieldScope, fieldsMap)
		if err != nil {
			return nil, err
		}

		if cond != nil {
			subConds = append(subConds, cond)
		}

	}

	return &AndCond{
		mCfg:      cfg,
		mSubConds: subConds,
	}, nil

}

func (cond *AndCond) Convert(ctx context.Context, vectorizer func(ctx context.Context, words []string) ([]*VectorResp, error)) (string, error) {
	res := `
	{
		"bool": {
			"must": [
				%s
			]
		}
	}
	`

	dslStr := ""
	validDSLs := []string{}
	for _, subCond := range cond.mSubConds {
		dsl, err := subCond.Convert(ctx, vectorizer)
		if err != nil {
			return "", err
		}

		// 过滤掉空字符串（被忽略的条件）
		if dsl != "" && dsl != "{}" {
			validDSLs = append(validDSLs, dsl)
		}
	}

	// 如果所有子条件都被过滤掉，返回空对象
	if len(validDSLs) == 0 {
		return "{}", nil
	}

	// 如果只有一个有效子条件，直接返回该子条件的 DSL，不需要包装在 bool.must 中
	if len(validDSLs) == 1 {
		return validDSLs[0], nil
	}

	// 多个有效子条件，用逗号连接
	for i, dsl := range validDSLs {
		if i != len(validDSLs)-1 {
			dslStr += dsl + ","
		} else {
			dslStr += dsl
		}
	}

	res = fmt.Sprintf(res, dslStr)
	return res, nil

}

func (cond *AndCond) Convert2SQL(ctx context.Context) (string, error) {
	sql := ""
	for i, subCond := range cond.mSubConds {
		where, err := subCond.Convert2SQL(ctx)
		if err != nil {
			return "", err
		}

		if i != len(cond.mSubConds)-1 {
			where += " AND "
		}

		sql += where

	}
	return sql, nil
}

// convertAndCondToDatasetFilterCondition converts AndCond to dataset filter condition format
// Reference: ontology-query's rewriteAndCondition pattern - recursively process sub-conditions
func convertAndCondToDatasetFilterCondition(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*FieldCfg,
	vectorizer func(ctx context.Context, word string) ([]*VectorResp, error)) (map[string]any, error) {
	if len(cfg.SubConds) == 0 {
		return nil, fmt.Errorf("sub condition size is 0")
	}

	if len(cfg.SubConds) > MaxSubCondition {
		return nil, fmt.Errorf("sub condition size limit %d but %d", MaxSubCondition, len(cfg.SubConds))
	}

	subConditions := make([]map[string]any, 0, len(cfg.SubConds))
	for _, subCond := range cfg.SubConds {
		subCondMap, err := ConvertCondCfgToFilterCondition(ctx, subCond, fieldsMap, vectorizer)
		if err != nil {
			return nil, err
		}
		if subCondMap != nil {
			subConditions = append(subConditions, subCondMap)
		}
	}

	// If all sub-conditions were filtered out, return nil
	if len(subConditions) == 0 {
		return nil, nil
	}

	// If only one sub-condition, return it directly without wrapping in "and"
	if len(subConditions) == 1 {
		return subConditions[0], nil
	}

	return map[string]any{
		"operation":      "and",
		"sub_conditions": subConditions,
	}, nil
}

// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package condition

import (
	"context"
	"fmt"
)

type OrCond struct {
	mCfg      *CondCfg
	mSubConds []Condition
}

func newOrCond(ctx context.Context, cfg *CondCfg, fieldScope uint8, fieldsMap map[string]*DataProperty) (Condition, error) {
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

		subConds = append(subConds, cond)
	}

	return &OrCond{
		mCfg:      cfg,
		mSubConds: subConds,
	}, nil

}

func (cond *OrCond) Convert(ctx context.Context, vectorizer func(ctx context.Context, property *DataProperty, word string) ([]VectorResp, error)) (string, error) {
	res := `
	{
		"bool": {
			"should": [
				%s
			]
		}
	}
	`

	dslStr := ""
	for i, subCond := range cond.mSubConds {
		dsl, err := subCond.Convert(ctx, vectorizer)
		if err != nil {
			return "", err
		}

		if i != len(cond.mSubConds)-1 {
			dsl += ","
		}

		dslStr += dsl

	}

	res = fmt.Sprintf(res, dslStr)
	return res, nil

}

func (cond *OrCond) Convert2SQL(ctx context.Context) (string, error) {
	sql := ""
	for i, subCond := range cond.mSubConds {
		where, err := subCond.Convert2SQL(ctx)
		if err != nil {
			return "", err
		}
		where = fmt.Sprintf("(%s)", where)
		if i != len(cond.mSubConds)-1 {
			where += " OR "
		}

		sql += where

	}
	sql = fmt.Sprintf("(%s)", sql)
	return sql, nil
}

func (cond *OrCond) RewriteCond(ctx context.Context,
	vectorizer func(ctx context.Context, property *DataProperty, word string) ([]VectorResp, error)) (*CondCfg, error) {

	return cond.mCfg, nil
}

func rewriteOrCondition(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*DataProperty,
	vectorizer func(ctx context.Context, property *DataProperty, word string) ([]VectorResp, error)) (*CondCfg, error) {

	subConds := []*CondCfg{}

	if len(cfg.SubConds) == 0 {
		return nil, fmt.Errorf("sub condition size is 0")
	}

	if len(cfg.SubConds) > MaxSubCondition {
		return nil, fmt.Errorf("sub condition size limit %d but %d", MaxSubCondition, len(cfg.SubConds))
	}

	for _, subCond := range cfg.SubConds {
		cond, err := RewriteCondition(ctx, subCond, fieldsMap, vectorizer)
		if err != nil {
			return nil, err
		}

		if cond != nil {
			subConds = append(subConds, cond)
		}
	}
	viewCondi := *cfg
	viewCondi.SubConds = subConds
	viewCondi.NameField = nil

	return &viewCondi, nil
}

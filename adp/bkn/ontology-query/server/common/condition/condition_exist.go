// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package condition

import (
	"context"
	"fmt"
)

type ExistCond struct {
	mCfg       *CondCfg
	mfieldName string
}

func NewExistCond(cfg *CondCfg) (Condition, error) {
	return &ExistCond{
		mCfg:       cfg,
		mfieldName: cfg.Name,
	}, nil
}

func (cond *ExistCond) Convert(ctx context.Context, vectorizer func(ctx context.Context, property *DataProperty, word string) ([]VectorResp, error)) (string, error) {
	dslStr := `
	{
		"exists": {
			"field": "%s"
		}
	}
	`

	return fmt.Sprintf(dslStr, cond.mfieldName), nil
}

// sql中没有字段存在的过滤条件,暂时用非空表达
func (cond *ExistCond) Convert2SQL(ctx context.Context) (string, error) {
	return fmt.Sprintf(`"%s" IS NOT NULL`, cond.mfieldName), nil
}

func rewriteExistCond(cfg *CondCfg) (*CondCfg, error) {

	// 过滤条件中的属性字段换成映射的视图字段
	if cfg.NameField.Name == "" {
		return nil, fmt.Errorf("存在[exist]操作符使用的过滤字段[%s]在对象类的属性中不存在", cfg.Name)
	}

	return &CondCfg{
		Name:        cfg.NameField.MappedField.Name,
		Operation:   cfg.Operation,
		ValueOptCfg: cfg.ValueOptCfg,
	}, nil
}

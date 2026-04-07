// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package filter_condition

import (
	"context"
	"fmt"

	"vega-backend/interfaces"
)

type KnnVectorCond struct {
	Cfg             *interfaces.FilterCondCfg
	FilterFieldName string
	SubConds        []interfaces.FilterCondition
}

func (c *KnnVectorCond) GetOperation() string { return OperationKnnVector }

func (c *KnnVectorCond) SupportSubCond() bool       { return true }
func (c *KnnVectorCond) NeedName() bool             { return true }
func (c *KnnVectorCond) NeedValue() bool            { return true }
func (c *KnnVectorCond) NeedConstValue() bool       { return true }
func (c *KnnVectorCond) IsSingleValue() bool        { return false }
func (c *KnnVectorCond) IsFixedLenArrayValue() bool { return false }
func (c *KnnVectorCond) RequiredValueLen() int      { return -1 }

// knn_vector 条件, 判断字段是否匹配某个向量
func (c *KnnVectorCond) New(ctx context.Context, cfg *interfaces.FilterCondCfg,
	fieldsMap map[string]*interfaces.Property) (interfaces.FilterCondition, error) {

	if cfg.Name == "" {
		return nil, fmt.Errorf("condition [knn_vector] left field is empty")
	}
	field, ok := fieldsMap[cfg.Name]
	if !ok {
		return nil, fmt.Errorf("condition [knn_vector] left field '%s' not found", cfg.Name)
	}
	if field.Type != interfaces.DataType_Vector {
		return nil, fmt.Errorf("condition [knn_vector] left field '%s' type must be vector", cfg.Name)
	}

	if cfg.ValueOptCfg.ValueFrom != interfaces.ValueFrom_Const {
		return nil, fmt.Errorf("condition [knn_vector] does not support value_from type '%s'", cfg.ValueFrom)
	}

	subConds := []interfaces.FilterCondition{}
	for _, subCond := range cfg.SubConds {
		cond, err := NewFilterCondition(ctx, subCond, fieldsMap)
		if err != nil {
			return nil, err
		}

		if cond != nil {
			subConds = append(subConds, cond)
		}
	}

	return &KnnVectorCond{
		Cfg:             cfg,
		FilterFieldName: cfg.Name,
		SubConds:        subConds,
	}, nil
}

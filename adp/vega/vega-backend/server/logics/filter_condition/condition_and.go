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

type AndCond struct {
	Cfg      *interfaces.FilterCondCfg
	SubConds []interfaces.FilterCondition
}

func (c *AndCond) GetOperation() string { return OperationAnd }

func (c *AndCond) SupportSubCond() bool       { return true }
func (c *AndCond) NeedName() bool             { return false }
func (c *AndCond) NeedValue() bool            { return false }
func (c *AndCond) NeedConstValue() bool       { return false }
func (c *AndCond) IsSingleValue() bool        { return false }
func (c *AndCond) IsFixedLenArrayValue() bool { return false }
func (c *AndCond) RequiredValueLen() int      { return -1 }

func (c *AndCond) New(ctx context.Context, cfg *interfaces.FilterCondCfg,
	fieldsMap map[string]*interfaces.Property) (cond interfaces.FilterCondition, err error) {

	subConds := []interfaces.FilterCondition{}

	if len(cfg.SubConds) == 0 {
		return nil, fmt.Errorf("sub condition size is 0")
	}

	if len(cfg.SubConds) > interfaces.MaxSubCondition {
		return nil, fmt.Errorf("sub condition size limit %d but %d", interfaces.MaxSubCondition, len(cfg.SubConds))
	}

	for _, subCond := range cfg.SubConds {
		cond, err = NewFilterCondition(ctx, subCond, fieldsMap)
		if err != nil {
			return nil, err
		}

		if cond != nil {
			subConds = append(subConds, cond)
		}

	}

	return &AndCond{
		Cfg:      cfg,
		SubConds: subConds,
	}, nil

}

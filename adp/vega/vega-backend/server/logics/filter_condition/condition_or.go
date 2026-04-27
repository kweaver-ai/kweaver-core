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

type OrCond struct {
	Cfg      *interfaces.FilterCondCfg
	SubConds []interfaces.FilterCondition
}

func (c *OrCond) GetOperation() string { return OperationOr }

func (c *OrCond) SupportSubCond() bool       { return true }
func (c *OrCond) NeedName() bool             { return false }
func (c *OrCond) NeedValue() bool            { return false }
func (c *OrCond) NeedConstValue() bool       { return false }
func (c *OrCond) IsSingleValue() bool        { return false }
func (c *OrCond) IsFixedLenArrayValue() bool { return false }
func (c *OrCond) RequiredValueLen() int      { return -1 }

func (c *OrCond) New(ctx context.Context, cfg *interfaces.FilterCondCfg,
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

		subConds = append(subConds, cond)
	}

	return &OrCond{
		Cfg:      cfg,
		SubConds: subConds,
	}, nil
}

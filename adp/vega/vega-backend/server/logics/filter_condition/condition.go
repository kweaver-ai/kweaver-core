// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package filter_condition

import (
	"context"
	"fmt"
	"reflect"

	"vega-backend/interfaces"
)

// 将过滤条件拼接到 dsl 请求的 query 部分
func NewFilterCondition(ctx context.Context, cfg *interfaces.FilterCondCfg,
	fieldsMap map[string]*interfaces.Property) (interfaces.FilterCondition, error) {

	if cfg == nil {
		return nil, nil
	}

	// 判断过滤器是否为空对象 {}
	if cfg.Name == "" && cfg.Operation == "" && len(cfg.SubConds) == 0 && cfg.ValueFrom == "" && cfg.Value == nil {
		return nil, nil
	}

	condFactory, exists := OperationMap[cfg.Operation]
	if !exists {
		return nil, fmt.Errorf("unsupported operation: %s", cfg.Operation)
	}

	cond, err := condFactory.New(ctx, cfg, fieldsMap)
	if err != nil {
		return nil, err
	}
	return cond, nil
}

func IsSlice(i any) bool {
	kind := reflect.ValueOf(i).Kind()
	return kind == reflect.Slice || kind == reflect.Array
}

func IsSameType(arr []any) bool {
	if len(arr) == 0 {
		return true
	}

	firstType := reflect.TypeOf(arr[0])
	for _, v := range arr {
		if reflect.TypeOf(v) != firstType {
			return false
		}
	}

	return true
}

// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package knsearch（可语义检索字段筛选）
// file: semantic_searchable_fields.go
package knsearch

import (
	"strings"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
)

// searchableField 可参与语义检索的字段信息
type searchableField struct {
	Name          string
	HasKnn        bool
	HasExactMatch bool
	HasMatch      bool
}

var semanticOps = map[interfaces.KnOperationType]struct{}{
	interfaces.KnOperationTypeEqual: {},
	interfaces.KnOperationTypeMatch: {},
	interfaces.KnOperationTypeKnn:   {},
}

// isTextField 判断属性是否为文本类型
func isTextField(prop *interfaces.KnSearchDataProperty) bool {
	if prop == nil {
		return false
	}
	t := strings.TrimSpace(strings.ToLower(prop.Type))
	if t == "" {
		return false
	}
	textTypes := []string{"text", "string", "varchar", "char"}
	for _, tt := range textTypes {
		if t == tt || strings.HasPrefix(t, tt+"[") {
			return true
		}
	}
	return false
}

// findSemanticSearchableFields 从对象类型中筛选可语义检索的字段
func findSemanticSearchableFields(objType *interfaces.KnSearchObjectType) []searchableField {
	if objType == nil || len(objType.DataProperties) == 0 {
		return nil
	}
	var out []searchableField
	for _, p := range objType.DataProperties {
		if p == nil {
			continue
		}
		name := strings.TrimSpace(p.Name)
		if name == "" {
			continue
		}
		if !isTextField(p) {
			continue
		}
		ops := p.ConditionOperations
		if len(ops) == 0 {
			continue
		}
		var hasExact, hasMatch, hasKnn bool
		for _, op := range ops {
			if _, ok := semanticOps[op]; !ok {
				continue
			}
			switch op {
			case interfaces.KnOperationTypeEqual:
				hasExact = true
			case interfaces.KnOperationTypeMatch:
				hasMatch = true
			case interfaces.KnOperationTypeKnn:
				hasKnn = true
			case interfaces.KnOperationTypeAnd, interfaces.KnOperationTypeOr,
				interfaces.KnOperationTypeNotEqual, interfaces.KnOperationTypeGreater, interfaces.KnOperationTypeLess,
				interfaces.KnOperationTypeGreaterOrEqual, interfaces.KnOperationTypeLessOrEqual,
				interfaces.KnOperationTypeIn, interfaces.KnOperationTypeNotIn,
				interfaces.KnOperationTypeLike, interfaces.KnOperationTypeNotLike,
				interfaces.KnOperationTypeRange, interfaces.KnOperationTypeOutRange,
				interfaces.KnOperationTypeExist, interfaces.KnOperationTypeNotExist,
				interfaces.KnOperationTypeRegex:
				// 非 semantic 检索相关操作类型，不设置 hasExact/hasMatch/hasKnn
			}
		}
		if !hasExact && !hasMatch && !hasKnn {
			continue
		}
		out = append(out, searchableField{
			Name:          name,
			HasKnn:        hasKnn,
			HasExactMatch: hasExact,
			HasMatch:      hasMatch,
		})
	}
	return out
}

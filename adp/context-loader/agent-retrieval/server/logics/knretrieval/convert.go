// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package knretrieval

import (
	"fmt"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
)

var knOperationTypeMap = map[string]interfaces.KnOperationType{
	"and":       interfaces.KnOperationTypeAnd,
	"or":        interfaces.KnOperationTypeOr,
	"==":        interfaces.KnOperationTypeEqual,
	"!=":        interfaces.KnOperationTypeNotEqual,
	">":         interfaces.KnOperationTypeGreater,
	"<":         interfaces.KnOperationTypeLess,
	">=":        interfaces.KnOperationTypeGreaterOrEqual,
	"<=":        interfaces.KnOperationTypeLessOrEqual,
	"in":        interfaces.KnOperationTypeIn,
	"not_in":    interfaces.KnOperationTypeNotIn,
	"like":      interfaces.KnOperationTypeLike,
	"not_like":  interfaces.KnOperationTypeNotLike,
	"range":     interfaces.KnOperationTypeRange,
	"out_range": interfaces.KnOperationTypeOutRange,
	"exist":     interfaces.KnOperationTypeExist,
	"not_exist": interfaces.KnOperationTypeNotExist,
	"regex":     interfaces.KnOperationTypeRegex,
	"match":     interfaces.KnOperationTypeMatch,
	"knn":       interfaces.KnOperationTypeKnn,
}

// ParseKnOperationType 将字符串解析为 KnOperationType。
// 如果输入字符串无效，则返回错误。
func ParseKnOperationType(s string) (interfaces.KnOperationType, error) {
	if op, exists := knOperationTypeMap[s]; exists {
		return op, nil
	}
	return "", fmt.Errorf("无效的 KnOperationType: %s", s)
}

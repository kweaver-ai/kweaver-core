// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package validator provides validation utilities and error type mappings.
package validator

import (
	myErr "github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/errors"
)

// TagToErrorType Validate tag 映射到错误分类
var TagToErrorType = map[string]string{
	// 必填类
	"required":        myErr.ErrExtCodeValidationRequired,
	"required_if":     myErr.ErrExtCodeValidationRequired,
	"required_unless": myErr.ErrExtCodeValidationRequired,
	"required_with":   myErr.ErrExtCodeValidationRequired,

	// 格式类
	"email":    myErr.ErrExtCodeValidationFormat,
	"url":      myErr.ErrExtCodeValidationFormat,
	"uuid":     myErr.ErrExtCodeValidationFormat,
	"datetime": myErr.ErrExtCodeValidationFormat,
	"numeric":  myErr.ErrExtCodeValidationFormat,
	"alpha":    myErr.ErrExtCodeValidationFormat,
	"alphanum": myErr.ErrExtCodeValidationFormat,
	"ip":       myErr.ErrExtCodeValidationFormat,
	"mac":      myErr.ErrExtCodeValidationFormat,

	// 范围类
	"min": myErr.ErrExtCodeValidationRange,
	"max": myErr.ErrExtCodeValidationRange,
	"len": myErr.ErrExtCodeValidationRange,
	"gte": myErr.ErrExtCodeValidationRange,
	"lte": myErr.ErrExtCodeValidationRange,
	"gt":  myErr.ErrExtCodeValidationRange,
	"lt":  myErr.ErrExtCodeValidationRange,

	// 枚举类
	"oneof": myErr.ErrExtCodeValidationEnum,
}

package validator

import (
	myErr "github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
)

// TagToErrorType Validate tag 映射到错误分类
var TagToErrorType = map[string]myErr.ErrorCode{
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

package persrecvalid

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

var (
	// 名称不能包含 \ / : * ? " < > | 特殊字符, 长度不能超过64个字符
	// 并且不能以空格开头和结尾，包括“只有单个空格的情况”（这个前端应该会做trim处理）
	NameRe = regexp.MustCompile(`^[^\s\\/:*?"<>|]([^\\/:*?"<>|]{0,62}[^\s\\/:*?"<>|])?$`)
	CodeRe = regexp.MustCompile(`^([\w@.\-]{1,64})$`)
)

const (
	CodeErrMsg = "只允许字母、数字、@、_、-、.，长度：1-64"
	NameErrMsg = `不能包含 \ / : * ? " < > | 特殊字符, 长度不能超过64个字符`
)

// CheckName name check
func CheckName(fieldLevel validator.FieldLevel) bool {
	name := fieldLevel.Field().String()
	if name == "" {
		return true // 允许空字符串，由其他校验器处理
	}

	return NameRe.MatchString(name)
}

func CheckCode(fieldLevel validator.FieldLevel) bool {
	code := fieldLevel.Field().String()
	if code == "" {
		return true // 允许空字符串，由其他校验器处理
	}

	return CodeRe.MatchString(code)
}

func GenNameErrMsg(target string) string {
	return target + NameErrMsg
}

func GenCodeErrMsg(target string) string {
	return target + CodeErrMsg
}

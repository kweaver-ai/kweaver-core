package customvalidator

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

var AgentAndTplNameRe = regexp.MustCompile(`^[\p{Han}a-zA-Z_][\p{Han}a-zA-Z0-9_]*$`)

const (
	AgentAndTplNameErrMsg = "仅支持中英文、数字及下划线，且不能以数字开头"
)

// CheckAgentAndTplName name check
func CheckAgentAndTplName(fieldLevel validator.FieldLevel) bool {
	name := fieldLevel.Field().String()
	if name == "" {
		return true // 允许空字符串，由其他校验器处理
	}

	return AgentAndTplNameRe.MatchString(name)
}

func GenAgentAndTplNameErrMsg(target string) string {
	return target + AgentAndTplNameErrMsg
}

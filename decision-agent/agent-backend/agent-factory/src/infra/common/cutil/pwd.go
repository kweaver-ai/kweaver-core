package cutil

import "regexp"

// 必须包含大小写字母、数字，只允许包含~!%#$@-_.特殊字符，长度为10-100
func CheckPassword(password string) bool {
	if len(password) < 10 || len(password) > 100 {
		return false
	}

	// 检查是否包含必需的字符类型
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	// 检查是否只包含允许的字符
	validChars := regexp.MustCompile(`^[A-Za-z0-9~!%#$@\-_.]+$`).MatchString(password)

	return hasUpper && hasLower && hasNumber && validChars
}

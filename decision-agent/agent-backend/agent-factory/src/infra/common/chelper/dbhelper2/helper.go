package dbhelper2

import (
	"fmt"
	"reflect"
	"strings"
)

func FillSQL(sql string, args ...interface{}) (filledSql string) {
	filledSql = sql

	for _, arg := range args {
		var argStr string
		switch v := arg.(type) {
		case string:
			argStr = fmt.Sprintf("'%s'", v)
		default:
			// 检查 nil 参数 - 不应该出现，如果出现则说明调用方有问题
			if arg == nil {
				panic("FillSQL: 参数不能为 nil，请检查调用代码")
			} else if reflect.TypeOf(arg).Kind() == reflect.String {
				// 检查是否为底层类型是 string 的自定义类型
				argStr = fmt.Sprintf("'%s'", arg)
			} else {
				argStr = fmt.Sprintf("%v", v)
			}
		}

		filledSql = strings.Replace(filledSql, "?", argStr, 1)
	}

	return
}

func GenInClauseGeneric[T any](args []T) (inClause string) {
	if len(args) == 0 {
		return ""
	}

	var sb strings.Builder

	for i, arg := range args {
		if i > 0 {
			sb.WriteString(",")
		}

		switch v := any(arg).(type) {
		case string:
			sb.WriteByte('\'')
			sb.WriteString(v)
			sb.WriteByte('\'')
		default:
			// 检查 nil 参数 - 不应该出现，如果出现则说明调用方有问题
			if any(arg) == nil {
				panic("GenInClauseGeneric: 参数不能为 nil，请检查调用代码")
			} else if reflect.TypeOf(arg).Kind() == reflect.String {
				// 检查是否为底层类型是 string 的自定义类型
				sb.WriteByte('\'')
				sb.WriteString(fmt.Sprintf("%v", arg))
				sb.WriteByte('\'')
			} else {
				sb.WriteString(fmt.Sprintf("%v", v))
			}
		}
	}

	return sb.String()
}

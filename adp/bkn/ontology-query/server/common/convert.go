// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package common

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
)

// string 转 []string
func StringToStringSlice(str string) []string {
	if str == "" {
		return []string{}
	}

	strSlice := []string{}
	strs := strings.Split(str, ",")
	for _, v := range strs {
		v = strings.Trim(v, " ")
		if v != "" {
			strSlice = append(strSlice, v)
		}
	}
	return strSlice
}

const (
	oneGiB = 1024 * 1024 * 1024 //1073741824.0 定义1GB的字节数
)

func BytesToGiB(bytes int64) float64 {
	return math.Round(float64(bytes)/oneGiB*100) / 100 // 四舍五入到小数点后两位
}

func GiBToBytes(gib int64) int64 {
	return gib * oneGiB
}

// AssertFloat64 尝试将 interface{} 转换为 float64，如果转换失败则返回错误。
func AnyToInt64(value any) (int64, error) {
	v := reflect.ValueOf(value)

	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int(), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return int64(v.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return int64(v.Float()), nil
	case reflect.String:
		return strconv.ParseInt(v.String(), 10, 64)
	default:
		return 0, fmt.Errorf("无法将类型 %T 转换为 int64", value)
	}
}

// 使用类型断言和标准库进行转换
func AnyToBool(value any) (bool, error) {
	// 检查是否是字符串类型
	if s, ok := value.(string); ok {
		// 使用 strconv.ParseBool 转换字符串
		// 它接受 "1", "t", "T", "TRUE", "true", "True" 为真
		// 接受 "0", "f", "F", "FALSE", "false", "False" 为假 [citation:3][citation:5][citation:8]
		return strconv.ParseBool(s)
	}
	// 检查是否是布尔类型本身，如果是则直接返回
	if b, ok := value.(bool); ok {
		return b, nil
	}
	return false, fmt.Errorf("unsupported type: %T", value)
}

func AnyToString(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", v)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%f", v)
	case bool:
		return strconv.FormatBool(v)
	case []byte:
		return string(v)
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", v)
	}
}

// ReplaceLikeWildcards，把 like 的通配符替换成正则表达式里的字符
func ReplaceLikeWildcards(input string) string {
	if input == "" {
		return input
	}

	var result strings.Builder
	escaped := false
	runes := []rune(input)

	for i := 0; i < len(runes); i++ {
		r := runes[i]

		if escaped {
			// 转义字符后的字符
			switch r {
			case '%', '_', '\\':
				result.WriteRune(r)
			default:
				// 如果转义了非特殊字符，保留转义符和字符
				result.WriteRune('\\')
				result.WriteRune(r)
			}
			escaped = false
		} else if r == '\\' {
			// 遇到转义符，检查是否是最后一个字符
			if i == len(runes)-1 {
				// 转义符在末尾，直接输出
				result.WriteRune(r)
			} else {
				// 标记转义状态，但不立即输出转义符
				escaped = true
			}
		} else if r == '%' {
			result.WriteString(".*")
		} else if r == '_' {
			result.WriteString(".")
		} else {
			result.WriteRune(r)
		}
	}

	// 处理以转义符结尾的情况
	if escaped {
		result.WriteRune('\\')
	}

	return result.String()
}

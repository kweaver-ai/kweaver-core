package utils

import "strings"

func StringRepeat(str string, n int, separator string) string {

	if n <= 0 {
		return ""
	}

	var builder strings.Builder
	for i := 0; i < n; i++ {
		if i > 0 {
			builder.WriteString(separator)
		}
		builder.WriteString(str)
	}

	return builder.String()
}

func StringExclude(arr []string, target string) []string {
	res := make([]string, 0)
	for _, v := range arr {
		if v != target {
			res = append(res, v)
		}
	}
	return res
}

// Slice 函数从字符串 str 中截取从 start 开始、长度为 length 的子串
func StringSlice(str string, start, length int) string {
	// 将字符串转换为 rune 切片以支持 Unicode
	runes := []rune(str)
	totalRunes := len(runes)

	// 处理 start 和 length 的边界情况
	if start < 0 {
		start = 0
	}
	if start >= totalRunes {
		return "" // 如果 start 超出字符串长度，返回空字符串
	}
	if length < 0 {
		length = 0
	}
	if start+length > totalRunes {
		length = totalRunes - start // 如果长度超出剩余字符数，调整长度
	}

	// 截取子串
	return string(runes[start : start+length])
}

func RemoveZeroWidthChars(s string) string {
	s = strings.ReplaceAll(s, "\u200B", "") // 零宽空格
	s = strings.ReplaceAll(s, "\u200C", "") // 零宽非连接符
	s = strings.ReplaceAll(s, "\u200D", "") // 零宽连接符
	s = strings.ReplaceAll(s, "\uFEFF", "") // 零宽不换行空格 (BOM)
	return s
}

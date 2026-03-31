// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package common

import (
	"math"
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

// 对字符串数组去重
func DuplicateSlice(strSlice []string) []string {
	keys := make(map[string]struct{})
	list := make([]string, 0, len(strSlice))

	for _, item := range strSlice {
		if _, ok := keys[item]; !ok {
			keys[item] = struct{}{}
			list = append(list, item)
		}
	}
	return list
}

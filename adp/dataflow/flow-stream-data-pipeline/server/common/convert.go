package common

import (
	"strings"
)

const (
	oneGiB = 1024 * 1024 * 1024 //1073741824.0 定义1GB的字节数
)

func StringToStringArr(str string) []string {
	if str != "" {
		return strings.Split(str, ",")
	} else {
		return []string{}
	}
}

func GiBToBytes(gib int64) int64 {
	return gib * oneGiB
}

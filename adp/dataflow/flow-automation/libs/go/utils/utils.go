package utils

import (
	"strings"
)

// IsIPv6 是否是ipv6类型地址
func IsIPv6(address string) bool {
	return strings.Count(address, ":") >= 2
}

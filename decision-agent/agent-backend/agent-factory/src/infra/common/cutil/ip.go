package cutil

import (
	"fmt"
	"strings"
)

// ParseHost 判定host是否为IPv6格式，如果是，返回 [host]
func ParseHost(host string) string {
	if strings.Contains(host, ":") {
		return fmt.Sprintf("[%s]", host)
	}

	return host
}

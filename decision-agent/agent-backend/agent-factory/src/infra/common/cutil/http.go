package cutil

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

var ipv6Reg = regexp.MustCompile("^[0-9a-fA-F:]*:[0-9a-fA-F:.]*$")

func IsHttpErr(r *http.Response) bool {
	return r.StatusCode < http.StatusOK || r.StatusCode >= http.StatusMultipleChoices
}

func SetTpTlsInsecureSkipVerify(tp *http.Transport) {
	if tp.TLSClientConfig == nil {
		tp.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	} else {
		tp.TLSClientConfig.InsecureSkipVerify = true
	}
}

func GetHTTPAccess(addr string, port int, protocol string) string {
	addr = DoubleStackHost(addr)
	address := fmt.Sprintf("%s:%d", addr, port)

	if port == 0 {
		address = addr
	}

	return fmt.Sprintf("%s://%s", protocol, address)
}

// DoubleStackHost 适配IPv6和IPv4, 对包含冒号不包含中括号的host添加中括号, 不对IP格式做校验
func DoubleStackHost(host string) string {
	host = strings.TrimSpace(host)
	if ipv6Reg.MatchString(host) {
		host = "[" + host + "]"
	}

	return host
}

package main

import (
	"fmt"
	"os"
	"strings"

	pkgerrors "github.com/pkg/errors"
)

// usageError 返回命令行工具的统一用法说明。
func usageError() error {
	return pkgerrors.New(`usage: go run ./cmd/openapi-docs <generate|compare|validate> [flags]`)
}

// optionalPath 将空串或 "-" 视为未配置路径，方便调用方跳过可选文件。
func optionalPath(path string) string {
	if strings.TrimSpace(path) == "" || path == "-" {
		return ""
	}

	return path
}

// exitWithError 将错误打印到标准错误输出，并以非零退出码结束进程。
func exitWithError(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package locale

import (
	"os"
	"path"
	"runtime"

	"github.com/kweaver-ai/kweaver-go-lib/i18n"
)

var (
	localeDir = "/locale"
)

func Register() {
	var abPath string

	// 优先使用包所在目录（保证 UT 与任意 cwd 下都能找到 locale）
	_, filename, _, ok := runtime.Caller(0)
	if ok {
		abPath = path.Dir(filename)
		if _, err := os.Stat(abPath); err == nil {
			i18n.RegisterI18n(abPath)
			return
		}
	}
	// 回退：使用 cwd + /locale（兼容旧行为）
	abPath, _ = os.Getwd()
	abPath += localeDir
	i18n.RegisterI18n(abPath)
}

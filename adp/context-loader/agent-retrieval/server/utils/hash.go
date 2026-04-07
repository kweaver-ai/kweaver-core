// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package utils package define util in program
// @File hash.go
// @Description hash util
package utils

import (
	"crypto/md5"
	"fmt"
	"io"

	jsoniter "github.com/json-iterator/go"
)

// ObjectMD5Hash 计算对象的MD5哈希值
func ObjectMD5Hash(data interface{}) (string, error) {
	b, err := jsoniter.Marshal(data)
	if err != nil {
		return "", err
	}
	h := md5.New() //nolint:gosec
	_, err = h.Write(b)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// MD5 Md5hash
func MD5(str string) string {
	h := md5.New() //nolint:gosec
	_, _ = io.WriteString(h, str)
	return fmt.Sprintf("%x", h.Sum(nil))
}

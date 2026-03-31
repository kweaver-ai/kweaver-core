// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package rest

import (
	"fmt"
	"reflect"

	"github.com/bytedance/sonic"
	"github.com/toon-format/toon-go"
)

const (
	// ContentTypeTOON TOON 响应 Content-Type
	ContentTypeTOON = "application/toon"
)

// ResponseFormat 响应格式：JSON 或 TOON
type ResponseFormat string

const (
	// FormatJSON 默认 JSON
	FormatJSON ResponseFormat = "json"
	// FormatTOON 压缩格式 TOON
	FormatTOON ResponseFormat = "toon"
)

// ParseResponseFormat 解析 response_format 参数，非法值返回错误
func ParseResponseFormat(s string) (ResponseFormat, error) {
	switch s {
	case "", "json":
		return FormatJSON, nil
	case "toon":
		return FormatTOON, nil
	default:
		return FormatJSON, fmt.Errorf("invalid response_format: %q (allowed: json, toon)", s)
	}
}

// MarshalResponse 按指定格式序列化 body，返回 Content-Type 与 body 字节；错误时不做静默降级
func MarshalResponse(format ResponseFormat, body interface{}) (contentType string, bodyBytes []byte, err error) {
	if body == nil {
		return ContentTypeJSON, nil, nil
	}
	switch format {
	case FormatJSON:
		bodyBytes, err = sonic.Marshal(body)
		if err != nil {
			return "", nil, err
		}
		return ContentTypeJSON, bodyBytes, nil
	case FormatTOON:
		bodyBytes, err = marshalTOON(body)
		if err != nil {
			return "", nil, err
		}
		return ContentTypeTOON, bodyBytes, nil
	default: // fallback to JSON for unknown values
		bodyBytes, err = sonic.Marshal(body)
		if err != nil {
			return "", nil, err
		}
		return ContentTypeJSON, bodyBytes, nil
	}
}

// marshalTOON 将 body 编码为 TOON。
// map/slice 等非 struct 类型直接交给 toon-go；
// struct 因为只有 json tag、没有 toon tag，需通过 JSON 中转为 map 再编码，确保字段名与 API 契约一致。
func marshalTOON(body interface{}) ([]byte, error) {
	v := reflect.ValueOf(body)
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return toon.Marshal(body, toon.WithLengthMarkers(true))
	}
	// struct → JSON → any → TOON（保留 json tag 的字段名）
	jsonBytes, err := sonic.Marshal(body)
	if err != nil {
		return nil, err
	}
	var m any
	if err := sonic.Unmarshal(jsonBytes, &m); err != nil {
		return nil, err
	}
	return toon.Marshal(m, toon.WithLengthMarkers(true))
}

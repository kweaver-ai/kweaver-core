// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package postgresql

import (
	"strings"

	"vega-backend/interfaces"
)

// TypeMapping 将 PostgreSQL udt_name / 常见 data_type 映射到 VEGA 类型。
var TypeMapping = map[string]string{
	// 整数
	"int2":        interfaces.DataType_Integer,
	"int4":        interfaces.DataType_Integer,
	"int8":        interfaces.DataType_Integer,
	"serial":      interfaces.DataType_Integer,
	"serial2":     interfaces.DataType_Integer,
	"serial4":     interfaces.DataType_Integer,
	"serial8":     interfaces.DataType_Integer,
	"smallserial": interfaces.DataType_Integer,
	"bigserial":   interfaces.DataType_Integer,
	"smallint":    interfaces.DataType_Integer,
	"integer":     interfaces.DataType_Integer,
	"bigint":      interfaces.DataType_Integer,

	// 浮点
	"float4":           interfaces.DataType_Float,
	"float8":           interfaces.DataType_Float,
	"double precision": interfaces.DataType_Float,
	"real":             interfaces.DataType_Float,
	"money":            interfaces.DataType_Float,

	// decimal
	"numeric": interfaces.DataType_Decimal,
	"decimal": interfaces.DataType_Decimal,

	// 字符串
	"varchar":           interfaces.DataType_String,
	"bpchar":            interfaces.DataType_String,
	"name":              interfaces.DataType_String,
	"uuid":              interfaces.DataType_String,
	"char":              interfaces.DataType_String,
	"character":         interfaces.DataType_String,
	"interval":          interfaces.DataType_String,
	"character varying": interfaces.DataType_String,
	"inet":              interfaces.DataType_Ip,
	"cidr":              interfaces.DataType_Ip,

	// 文本
	"text": interfaces.DataType_Text,

	// 日期时间
	"date":                        interfaces.DataType_Date,
	"time":                        interfaces.DataType_Time,
	"timetz":                      interfaces.DataType_Time,
	"time without time zone":      interfaces.DataType_Time,
	"time with time zone":         interfaces.DataType_Time,
	"timestamp":                   interfaces.DataType_Timestamp,
	"timestamptz":                 interfaces.DataType_Timestamp,
	"smalldatetime":               interfaces.DataType_Timestamp,
	"timestamp without time zone": interfaces.DataType_Timestamp,
	"timestamp with time zone":    interfaces.DataType_Timestamp,

	// 布尔
	"bool":    interfaces.DataType_Boolean,
	"boolean": interfaces.DataType_Boolean,

	// 二进制
	"bytea": interfaces.DataType_Binary,

	// JSON
	"json":  interfaces.DataType_Json,
	"jsonb": interfaces.DataType_Json,
}

// MapType 根据 information_schema 的 udt_name 或 data_type 映射到 VEGA 类型。
func MapType(udtName, dataType string) string {
	u := strings.ToLower(strings.TrimSpace(udtName))
	if u != "" {
		if vegaType, ok := TypeMapping[u]; ok {
			return vegaType
		}
		if strings.HasPrefix(u, "_") {
			return interfaces.DataType_Other
		}
	}
	d := strings.ToLower(strings.TrimSpace(dataType))
	if vegaType, ok := TypeMapping[d]; ok {
		return vegaType
	}
	return interfaces.DataType_Other
}

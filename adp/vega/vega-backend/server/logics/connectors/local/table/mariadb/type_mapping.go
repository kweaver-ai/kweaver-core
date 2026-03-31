// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package mariadb provides MariaDB database connector implementation.
package mariadb

import (
	"strings"

	"vega-backend/interfaces"
)

// TypeMapping maps MariaDB native types to VEGA types.
var TypeMapping = map[string]string{
	// Integer types
	"tinyint":   interfaces.DataType_Integer,
	"smallint":  interfaces.DataType_Integer,
	"mediumint": interfaces.DataType_Integer,
	"int":       interfaces.DataType_Integer,
	"integer":   interfaces.DataType_Integer,
	"bigint":    interfaces.DataType_Integer,
	"year":      interfaces.DataType_Integer,

	// Unsigned integer types
	"tinyint unsigned":   interfaces.DataType_UnsignedInteger,
	"smallint unsigned":  interfaces.DataType_UnsignedInteger,
	"mediumint unsigned": interfaces.DataType_UnsignedInteger,
	"int unsigned":       interfaces.DataType_UnsignedInteger,
	"integer unsigned":   interfaces.DataType_UnsignedInteger,
	"bigint unsigned":    interfaces.DataType_UnsignedInteger,

	// Float types
	"float":            interfaces.DataType_Float,
	"double":           interfaces.DataType_Float,
	"real":             interfaces.DataType_Float,
	"double precision": interfaces.DataType_Float,

	// Decimal types
	"decimal": interfaces.DataType_Decimal,
	"numeric": interfaces.DataType_Decimal,
	"fixed":   interfaces.DataType_Decimal,
	"dec":     interfaces.DataType_Decimal,

	// String types
	"char":    interfaces.DataType_String,
	"varchar": interfaces.DataType_String,

	// Text types
	"tinytext":   interfaces.DataType_Text,
	"text":       interfaces.DataType_Text,
	"mediumtext": interfaces.DataType_Text,
	"longtext":   interfaces.DataType_Text,

	// Date/Time types
	"date":      interfaces.DataType_Date,
	"datetime":  interfaces.DataType_Datetime,
	"timestamp": interfaces.DataType_Timestamp,
	"time":      interfaces.DataType_Time,

	// Boolean
	"boolean": interfaces.DataType_Boolean,
	"bool":    interfaces.DataType_Boolean,
	"bit":     interfaces.DataType_Boolean,

	// Binary types
	"binary":     interfaces.DataType_Binary,
	"varbinary":  interfaces.DataType_Binary,
	"tinyblob":   interfaces.DataType_Binary,
	"blob":       interfaces.DataType_Binary,
	"mediumblob": interfaces.DataType_Binary,
	"longblob":   interfaces.DataType_Binary,

	// JSON
	"json": interfaces.DataType_Json,
}

// MapType returns VEGA type for MariaDB native type.
// nativeType 应使用 COLUMN_TYPE（如 "int unsigned"）以正确识别 unsigned 类型。
func MapType(nativeType string) string {
	t := strings.ToLower(strings.TrimSpace(nativeType))

	// 检查是否包含括号，如果包含则去除括号及其内容
	if idx := strings.Index(t, "("); idx != -1 {
		t = strings.TrimSpace(t[:idx])
	}

	if vegaType, ok := TypeMapping[t]; ok {
		return vegaType
	}
	return interfaces.DataType_Other // default
}

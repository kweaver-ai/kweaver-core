// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

const (
	// 整数型
	DataType_Integer         = "integer"
	DataType_UnsignedInteger = "unsigned integer"

	// 浮点型
	DataType_Float = "float"

	// 任意精度数
	DataType_Decimal = "decimal"

	// 字符串型
	DataType_String = "string"
	DataType_Text   = "text"

	// 时间型
	DataType_Date      = "date"
	DataType_Time      = "time"
	DataType_Datetime  = "datetime"
	DataType_Timestamp = "timestamp"

	// ip类型
	DataType_Ip = "ip"

	// 布尔型
	DataType_Boolean = "boolean"

	// 二进制数据类型
	DataType_Binary = "binary"

	// json类型
	DataType_Json = "json"

	// 空间类型
	DataType_Point = "point"
	DataType_Shape = "shape"

	// 向量类型
	DataType_Vector = "vector"

	// 其他类型
	DataType_Other = "other"
)

var (
	STRING_TYPES = map[string]struct{}{
		DataType_String: {},
		DataType_Text:   {},
	}

	NUMBER_TYPES = map[string]struct{}{
		DataType_Integer:         {},
		DataType_UnsignedInteger: {},
		DataType_Float:           {},
		DataType_Decimal:         {},
	}

	DATE_TYPES = map[string]struct{}{
		DataType_Date:      {},
		DataType_Time:      {},
		DataType_Datetime:  {},
		DataType_Timestamp: {},
	}
)

func DataType_IsString(t string) bool {
	_, ok := STRING_TYPES[t]
	return ok
}

func DataType_IsNumber(t string) bool {
	_, ok := NUMBER_TYPES[t]
	return ok
}

func DataType_IsDate(t string) bool {
	_, ok := DATE_TYPES[t]
	return ok
}

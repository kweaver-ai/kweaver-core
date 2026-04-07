// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package oracle provides Oracle database connector implementation.
package oracle

// TypeMapping maps Oracle native types to VEGA types.
var TypeMapping = map[string]string{
	// Integer types
	"number":         "decimal",
	"integer":        "integer",
	"int":            "integer",
	"smallint":       "integer",
	"binary_integer": "integer",
	"pls_integer":    "integer",

	// Float types
	"float":            "float",
	"binary_float":     "float",
	"binary_double":    "float",
	"real":             "float",
	"double precision": "float",

	// Decimal types
	"dec":     "decimal",
	"decimal": "decimal",
	"numeric": "decimal",

	// String types
	"char":      "string",
	"nchar":     "string",
	"varchar":   "string",
	"varchar2":  "string",
	"nvarchar2": "string",

	// Text types
	"clob":  "text",
	"nclob": "text",
	"long":  "text",

	// Date/Time types
	"date":                           "datetime",
	"timestamp":                      "datetime",
	"timestamp with time zone":       "datetime",
	"timestamp with local time zone": "datetime",
	"timestamp with tz":              "datetime",
	"timestamp with local tz":        "datetime",

	// Boolean
	"boolean": "boolean",

	// Binary types
	"raw":      "binary",
	"long raw": "binary",
	"blob":     "binary",

	// RowID
	"rowid":  "string",
	"urowid": "string",

	// XML
	"xmltype": "json",
}

// MapType returns VEGA type for Oracle native type.
func MapType(nativeType string) string {
	if vegaType, ok := TypeMapping[nativeType]; ok {
		return vegaType
	}
	return "unsupported" // default
}

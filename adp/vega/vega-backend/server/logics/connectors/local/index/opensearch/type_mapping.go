// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package opensearch provides OpenSearch/ElasticSearch connector implementation.
package opensearch

// TypeMapping maps OpenSearch native types to VEGA types.
var TypeMapping = map[string]string{
	// String types
	"text":    "text",
	"keyword": "string",
	"string":  "string", // Legacy type in older versions

	// Numeric types
	"byte":          "integer",
	"short":         "integer",
	"integer":       "integer",
	"long":          "integer",
	"unsigned_long": "unsigned_integer",

	// Float types
	"float":        "float",
	"half_float":   "float",
	"scaled_float": "float",
	"double":       "float",

	// Decimal types
	"double_precision": "decimal",

	// Boolean
	"boolean": "boolean",

	// Date/Time types
	"date":       "datetime",
	"date_nanos": "datetime",

	// Binary
	"binary": "binary",

	// Range types
	"integer_range": "string",
	"float_range":   "string",
	"long_range":    "string",
	"double_range":  "string",
	"date_range":    "string",
	"ip_range":      "string",

	// Object types
	"object": "json",
	"nested": "json",

	// Geo types
	"geo_point": "string",
	"geo_shape": "string",

	// IP type
	"ip": "string",

	// Completion type
	"completion": "string",

	// Token count
	"token_count": "integer",

	// Percolator
	"percolator": "string",

	// Join type
	"join": "string",

	// Rank feature
	"rank_feature":  "float",
	"rank_features": "float",

	// Dense vector
	"dense_vector": "string",

	// Sparse vector
	"sparse_vector": "string",

	// Search as you type
	"search_as_you_type": "text",

	// Alias field
	"alias": "string",

	// Flattened
	"flattened": "json",

	// Shape
	"shape": "string",

	// Version
	"version": "string",

	// Murmur3
	"murmur3": "string",

	// Aggregate metric
	"aggregate_metric_double": "float",
}

// MapType returns VEGA type for OpenSearch native type.
func MapType(nativeType string) string {
	if vegaType, ok := TypeMapping[nativeType]; ok {
		return vegaType
	}
	return "unsupported" // default
}

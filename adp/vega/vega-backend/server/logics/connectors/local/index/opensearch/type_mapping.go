// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package opensearch provides OpenSearch/ElasticSearch connector implementation.
package opensearch

import (
	"vega-backend/interfaces"
)

// TypeMapping maps OpenSearch native types to VEGA types.
var TypeMapping = map[string]string{
	// String types
	"text":    interfaces.DataType_Text,
	"keyword": interfaces.DataType_String,

	// Numeric types
	"byte":          interfaces.DataType_Integer,
	"short":         interfaces.DataType_Integer,
	"integer":       interfaces.DataType_Integer,
	"long":          interfaces.DataType_Integer,
	"unsigned_long": interfaces.DataType_UnsignedInteger,

	// Float types
	"float":        interfaces.DataType_Float,
	"half_float":   interfaces.DataType_Float,
	"scaled_float": interfaces.DataType_Float,
	"double":       interfaces.DataType_Float,

	// Decimal types
	"double_precision": interfaces.DataType_Decimal,

	// Boolean
	"boolean": interfaces.DataType_Boolean,

	// Date/Time types
	"date":       interfaces.DataType_Datetime,
	"date_nanos": interfaces.DataType_Datetime,

	// Binary
	"binary": interfaces.DataType_Binary,

	// Range types
	"integer_range": interfaces.DataType_String,
	"float_range":   interfaces.DataType_String,
	"long_range":    interfaces.DataType_String,
	"double_range":  interfaces.DataType_String,
	"date_range":    interfaces.DataType_String,
	"ip_range":      interfaces.DataType_String,

	// Object types
	"object": interfaces.DataType_Json,
	"nested": interfaces.DataType_Json,

	// Geo types
	"geo_point": interfaces.DataType_String,
	"geo_shape": interfaces.DataType_String,

	// IP type
	"ip": interfaces.DataType_String,

	// Completion type
	"completion": interfaces.DataType_String,

	// Token count
	"token_count": interfaces.DataType_Integer,

	// Percolator
	"percolator": interfaces.DataType_String,

	// Join type
	"join": interfaces.DataType_String,

	// Rank feature
	"rank_feature":  interfaces.DataType_Float,
	"rank_features": interfaces.DataType_Float,

	// Dense vector
	"dense_vector": interfaces.DataType_String,

	// Sparse vector
	"sparse_vector": interfaces.DataType_String,

	// Search as you type
	"search_as_you_type": interfaces.DataType_Text,

	// Alias field
	"alias": interfaces.DataType_String,

	// Flattened
	"flattened": interfaces.DataType_Json,

	// Shape
	"shape": interfaces.DataType_String,

	// Version
	"version": interfaces.DataType_String,

	// Murmur3
	"murmur3": interfaces.DataType_String,

	// Aggregate metric
	"aggregate_metric_double": interfaces.DataType_Float,
}

// MapType returns VEGA type for OpenSearch native type.
func MapType(nativeType string) string {
	if vegaType, ok := TypeMapping[nativeType]; ok {
		return vegaType
	}
	return interfaces.DataType_Other // default
}

// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package errors Query 模块错误码
package errors

// Query 错误码
const (
	// 400 Bad Request
	VegaBackend_Query_InvalidParameter                      = "VegaBackend.Query.InvalidParameter"
	VegaBackend_Query_InvalidParameter_QueryIDRequired      = "VegaBackend.Query.InvalidParameter.QueryIDRequired"
	VegaBackend_Query_InvalidParameter_LimitExceeded        = "VegaBackend.Query.InvalidParameter.LimitExceeded"
	VegaBackend_Query_InvalidParameter_JoinTableNotInTables = "VegaBackend.Query.InvalidParameter.JoinTableNotInTables"
	VegaBackend_Query_InvalidParameter_QueryTimeout         = "VegaBackend.Query.InvalidParameter.QueryTimeout"

	// 404 Not Found
	VegaBackend_Query_CatalogNotFound  = "VegaBackend.Query.CatalogNotFound"
	VegaBackend_Query_ResourceNotFound = "VegaBackend.Query.ResourceNotFound"

	// 410 Gone - Session 过期
	VegaBackend_Query_SessionExpired = "VegaBackend.Query.SessionExpired"

	// 429 Too Many Requests - 并发限流
	VegaBackend_Query_ConcurrencyLimitExceeded = "VegaBackend.Query.ConcurrencyLimitExceeded"

	// 501 Not Implemented
	VegaBackend_Query_MultiCatalogNotSupported = "VegaBackend.Query.MultiCatalogNotSupported"

	// 500 Internal Server Error
	VegaBackend_Query_ExecuteFailed = "VegaBackend.Query.ExecuteFailed"
)

var QueryErrCodeList = []string{
	VegaBackend_Query_InvalidParameter,
	VegaBackend_Query_InvalidParameter_QueryIDRequired,
	VegaBackend_Query_InvalidParameter_LimitExceeded,
	VegaBackend_Query_InvalidParameter_JoinTableNotInTables,
	VegaBackend_Query_InvalidParameter_QueryTimeout,
	VegaBackend_Query_CatalogNotFound,
	VegaBackend_Query_ResourceNotFound,
	VegaBackend_Query_SessionExpired,
	VegaBackend_Query_ConcurrencyLimitExceeded,
	VegaBackend_Query_MultiCatalogNotSupported,
	VegaBackend_Query_ExecuteFailed,
}

// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package errors Extensions（Issue #382 方案 B）错误码
package errors

const (
	VegaBackend_Extensions_InvalidFormat              = "VegaBackend.Extensions.InvalidFormat"
	VegaBackend_Extensions_QuotaExceeded              = "VegaBackend.Extensions.QuotaExceeded"
	VegaBackend_Extensions_ReservedKey                = "VegaBackend.Extensions.ReservedKey"
	VegaBackend_Extensions_MismatchedQueryPairs       = "VegaBackend.Extensions.MismatchedQueryPairs"
	VegaBackend_Extensions_TooManyFilterPairs         = "VegaBackend.Extensions.TooManyFilterPairs"
	VegaBackend_Extensions_PropertyQuotaExceeded      = "VegaBackend.Extensions.PropertyQuotaExceeded"
)

// ExtensionsErrCodeList 须在 init 中 rest.Register，否则返回该码时进程会 fatal（missing errorCode）。
var ExtensionsErrCodeList = []string{
	VegaBackend_Extensions_InvalidFormat,
	VegaBackend_Extensions_QuotaExceeded,
	VegaBackend_Extensions_PropertyQuotaExceeded,
	VegaBackend_Extensions_ReservedKey,
	VegaBackend_Extensions_MismatchedQueryPairs,
	VegaBackend_Extensions_TooManyFilterPairs,
}

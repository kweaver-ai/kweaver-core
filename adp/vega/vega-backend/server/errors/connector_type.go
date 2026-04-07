// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package errors ConnectorType 模块错误码
package errors

// ConnectorType 错误码
const (
	// 400 Bad Request
	VegaBackend_ConnectorType_InvalidParameter          = "VegaBackend.ConnectorType.InvalidParameter"
	VegaBackend_ConnectorType_InvalidParameter_Type     = "VegaBackend.ConnectorType.InvalidParameter.Type"
	VegaBackend_ConnectorType_InvalidParameter_Name     = "VegaBackend.ConnectorType.InvalidParameter.Name"
	VegaBackend_ConnectorType_InvalidParameter_Mode     = "VegaBackend.ConnectorType.InvalidParameter.Mode"
	VegaBackend_ConnectorType_InvalidParameter_Category = "VegaBackend.ConnectorType.InvalidParameter.Category"
	VegaBackend_ConnectorType_InvalidParameter_Endpoint = "VegaBackend.ConnectorType.InvalidParameter.Endpoint"
	VegaBackend_ConnectorType_BadRequest                = "VegaBackend.ConnectorType.BadRequest"

	// 404 Not Found / 409 Conflict
	VegaBackend_ConnectorType_NotFound   = "VegaBackend.ConnectorType.NotFound"
	VegaBackend_ConnectorType_TypeExists = "VegaBackend.ConnectorType.TypeExists"
	VegaBackend_ConnectorType_NameExists = "VegaBackend.ConnectorType.NameExists"

	// 500 Internal Server Error
	VegaBackend_ConnectorType_InternalError                = "VegaBackend.ConnectorType.InternalError"
	VegaBackend_ConnectorType_InternalError_RegisterFailed = "VegaBackend.ConnectorType.InternalError.RegisterFailed"
	VegaBackend_ConnectorType_InternalError_GetFailed      = "VegaBackend.ConnectorType.InternalError.GetFailed"
	VegaBackend_ConnectorType_InternalError_UpdateFailed   = "VegaBackend.ConnectorType.InternalError.UpdateFailed"
	VegaBackend_ConnectorType_InternalError_DeleteFailed   = "VegaBackend.ConnectorType.InternalError.DeleteFailed"
)

var ConnectorTypeErrCodeList = []string{
	VegaBackend_ConnectorType_InvalidParameter,
	VegaBackend_ConnectorType_InvalidParameter_Type,
	VegaBackend_ConnectorType_InvalidParameter_Name,
	VegaBackend_ConnectorType_InvalidParameter_Mode,
	VegaBackend_ConnectorType_InvalidParameter_Category,
	VegaBackend_ConnectorType_InvalidParameter_Endpoint,
	VegaBackend_ConnectorType_BadRequest,
	VegaBackend_ConnectorType_NotFound,
	VegaBackend_ConnectorType_NameExists,
	VegaBackend_ConnectorType_TypeExists,
	VegaBackend_ConnectorType_InternalError,
	VegaBackend_ConnectorType_InternalError_RegisterFailed,
	VegaBackend_ConnectorType_InternalError_GetFailed,
	VegaBackend_ConnectorType_InternalError_UpdateFailed,
	VegaBackend_ConnectorType_InternalError_DeleteFailed,
}

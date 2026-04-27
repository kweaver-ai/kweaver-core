// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package errors Resource 模块错误码
package errors

// Resource 错误码
const (
	// 400 Bad Request
	VegaBackend_Resource_InvalidParameter           = "VegaBackend.Resource.InvalidParameter"
	VegaBackend_Resource_InvalidParameter_Type      = "VegaBackend.Resource.InvalidParameter.Type"
	VegaBackend_Resource_InvalidParameter_Name      = "VegaBackend.Resource.InvalidParameter.Name"
	VegaBackend_Resource_InvalidParameter_CatalogID = "VegaBackend.Resource.InvalidParameter.CatalogID"
	VegaBackend_Resource_LengthExceeded_Name        = "VegaBackend.Resource.LengthExceeded.Name"
	VegaBackend_Resource_LengthExceeded_Description = "VegaBackend.Resource.LengthExceeded.Description"
	VegaBackend_InvalidParameter_Aggregation        = "VegaBackend.InvalidParameter.Aggregation"
	VegaBackend_InvalidParameter_GroupBy            = "VegaBackend.InvalidParameter.GroupBy"
	VegaBackend_InvalidParameter_OrderBy            = "VegaBackend.InvalidParameter.OrderBy"
	VegaBackend_InvalidParameter_Having             = "VegaBackend.InvalidParameter.Having"

	// 403 Forbidden
	VegaBackend_Resource_NotFound        = "VegaBackend.Resource.NotFound"
	VegaBackend_Resource_NameExists      = "VegaBackend.Resource.NameExists"
	VegaBackend_Resource_IDExists        = "VegaBackend.Resource.IDExists"
	VegaBackend_Resource_CatalogNotFound = "VegaBackend.Resource.CatalogNotFound"
	VegaBackend_Resource_IsDisabled      = "VegaBackend.Resource.IsDisabled"
	VegaBackend_Resource_AlreadyEnabled  = "VegaBackend.Resource.AlreadyEnabled"
	VegaBackend_Resource_AlreadyDisabled = "VegaBackend.Resource.AlreadyDisabled"

	// 500 Internal Server Error
	VegaBackend_Resource_InternalError                       = "VegaBackend.Resource.InternalError"
	VegaBackend_Resource_InternalError_CreateFailed          = "VegaBackend.Resource.InternalError.CreateFailed"
	VegaBackend_Resource_InternalError_GetFailed             = "VegaBackend.Resource.InternalError.GetFailed"
	VegaBackend_Resource_InternalError_UpdateFailed          = "VegaBackend.Resource.InternalError.UpdateFailed"
	VegaBackend_Resource_InternalError_DeleteFailed          = "VegaBackend.Resource.InternalError.DeleteFailed"
	VegaBackend_Resource_InternalError_SyncFailed            = "VegaBackend.Resource.InternalError.SyncFailed"
	VegaBackend_Resource_InternalError_InvalidCategory       = "VegaBackend.Resource.InternalError.InvalidCategory"
	VegaBackend_Resource_InternalError_GetAccountNamesFailed = "VegaBackend.Resource.InternalError.GetAccountNamesFailed"
)

var ResourceErrCodeList = []string{
	VegaBackend_Resource_InvalidParameter,
	VegaBackend_Resource_InvalidParameter_Type,
	VegaBackend_Resource_InvalidParameter_Name,
	VegaBackend_Resource_InvalidParameter_CatalogID,
	VegaBackend_Resource_LengthExceeded_Name,
	VegaBackend_Resource_LengthExceeded_Description,
	VegaBackend_InvalidParameter_Aggregation,
	VegaBackend_InvalidParameter_GroupBy,
	VegaBackend_InvalidParameter_OrderBy,
	VegaBackend_InvalidParameter_Having,
	VegaBackend_Resource_NotFound,
	VegaBackend_Resource_NameExists,
	VegaBackend_Resource_IDExists,
	VegaBackend_Resource_CatalogNotFound,
	VegaBackend_Resource_IsDisabled,
	VegaBackend_Resource_AlreadyEnabled,
	VegaBackend_Resource_AlreadyDisabled,
	VegaBackend_Resource_InternalError,
	VegaBackend_Resource_InternalError_CreateFailed,
	VegaBackend_Resource_InternalError_GetFailed,
	VegaBackend_Resource_InternalError_UpdateFailed,
	VegaBackend_Resource_InternalError_DeleteFailed,
	VegaBackend_Resource_InternalError_SyncFailed,
	VegaBackend_Resource_InternalError_InvalidCategory,
	VegaBackend_Resource_InternalError_GetAccountNamesFailed,
}

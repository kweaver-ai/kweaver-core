// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package errors Catalog 模块错误码
package errors

// Catalog 错误码
const (
	// 400 Bad Request
	VegaBackend_Catalog_InvalidParameter                 = "VegaBackend.Catalog.InvalidParameter"
	VegaBackend_Catalog_InvalidParameter_Type            = "VegaBackend.Catalog.InvalidParameter.Type"
	VegaBackend_Catalog_InvalidParameter_Name            = "VegaBackend.Catalog.InvalidParameter.Name"
	VegaBackend_Catalog_InvalidParameter_ID              = "VegaBackend.Catalog.InvalidParameter.ID"
	VegaBackend_Catalog_InvalidParameter_ConnectorType   = "VegaBackend.Catalog.InvalidParameter.ConnectorType"
	VegaBackend_Catalog_InvalidParameter_ConnectorConfig = "VegaBackend.Catalog.InvalidParameter.ConnectorConfig"
	VegaBackend_Catalog_LengthExceeded_Name              = "VegaBackend.Catalog.LengthExceeded.Name"
	VegaBackend_Catalog_LengthExceeded_Description       = "VegaBackend.Catalog.LengthExceeded.Description"

	// 403 Forbidden
	VegaBackend_Catalog_NotFound   = "VegaBackend.Catalog.NotFound"
	VegaBackend_Catalog_NameExists = "VegaBackend.Catalog.NameExists"
	VegaBackend_Catalog_IDExists   = "VegaBackend.Catalog.IDExists"
	VegaBackend_Catalog_IDMismatch = "VegaBackend.Catalog.IDMismatch"
	VegaBackend_Catalog_HasAssets  = "VegaBackend.Catalog.HasAssets"
	VegaBackend_Catalog_IsDisabled = "VegaBackend.Catalog.IsDisabled"

	// 500 Internal Server Error
	VegaBackend_Catalog_InternalError                       = "VegaBackend.Catalog.InternalError"
	VegaBackend_Catalog_InternalError_CreateFailed          = "VegaBackend.Catalog.InternalError.CreateFailed"
	VegaBackend_Catalog_InternalError_CreateResourcesFailed = "VegaBackend.Catalog.InternalError.CreateResourcesFailed"
	VegaBackend_Catalog_InternalError_GetAccountNamesFailed = "VegaBackend.Catalog.InternalError.GetAccountNamesFailed"
	VegaBackend_Catalog_InternalError_GetFailed             = "VegaBackend.Catalog.InternalError.GetFailed"
	VegaBackend_Catalog_InternalError_UpdateFailed          = "VegaBackend.Catalog.InternalError.UpdateFailed"
	VegaBackend_Catalog_InternalError_DeleteFailed          = "VegaBackend.Catalog.InternalError.DeleteFailed"
	VegaBackend_Catalog_InternalError_TestConnectionFailed  = "VegaBackend.Catalog.InternalError.TestConnectionFailed"
	VegaBackend_Catalog_InternalError_EncryptFailed         = "VegaBackend.Catalog.InternalError.EncryptFailed"

	VegaBackend_Catalog_InvalidParameter_SensitiveFieldNotEncrypted = "VegaBackend.Catalog.InvalidParameter.SensitiveFieldNotEncrypted"
)

var CatalogErrCodeList = []string{
	VegaBackend_Catalog_InvalidParameter,
	VegaBackend_Catalog_InvalidParameter_Type,
	VegaBackend_Catalog_InvalidParameter_Name,
	VegaBackend_Catalog_InvalidParameter_ID,
	VegaBackend_Catalog_InvalidParameter_ConnectorType,
	VegaBackend_Catalog_InvalidParameter_ConnectorConfig,
	VegaBackend_Catalog_LengthExceeded_Name,
	VegaBackend_Catalog_LengthExceeded_Description,
	VegaBackend_Catalog_NotFound,
	VegaBackend_Catalog_NameExists,
	VegaBackend_Catalog_IDExists,
	VegaBackend_Catalog_IDMismatch,
	VegaBackend_Catalog_HasAssets,
	VegaBackend_Catalog_IsDisabled,
	VegaBackend_Catalog_InternalError,
	VegaBackend_Catalog_InternalError_CreateFailed,
	VegaBackend_Catalog_InternalError_CreateResourcesFailed,
	VegaBackend_Catalog_InternalError_GetFailed,
	VegaBackend_Catalog_InternalError_UpdateFailed,
	VegaBackend_Catalog_InternalError_DeleteFailed,
	VegaBackend_Catalog_InternalError_TestConnectionFailed,
	VegaBackend_Catalog_InternalError_EncryptFailed,
	VegaBackend_Catalog_InvalidParameter_SensitiveFieldNotEncrypted,
}

// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package errors

// DiscoverSchedule 相关错误码
const (
	// 404 Not Found
	VegaBackend_DiscoverSchedule_NotFound = "VegaBackend.DiscoverSchedule.NotFound"

	// 400 Bad Request
	VegaBackend_DiscoverSchedule_InvalidCronExpr   = "VegaBackend.DiscoverSchedule.InvalidCronExpr"
	VegaBackend_DiscoverSchedule_InvalidStrategies = "VegaBackend.DiscoverSchedule.InvalidStrategies"

	// 409 Conflict
	VegaBackend_DiscoverSchedule_IdMismatch             = "VegaBackend.DiscoverSchedule.IdMismatch"
	VegaBackend_DiscoverSchedule_CatalogMismatch        = "VegaBackend.DiscoverSchedule.CatalogMismatch"
	VegaBackend_DiscoverSchedule_EnabledFieldNotAllowed = "VegaBackend.DiscoverSchedule.EnabledFieldNotAllowed"

	// 500 Internal Server Error
	VegaBackend_DiscoverSchedule_InternalError_GetFailed    = "VegaBackend.DiscoverSchedule.InternalError.GetFailed"
	VegaBackend_DiscoverSchedule_InternalError_CreateFailed = "VegaBackend.DiscoverSchedule.InternalError.CreateFailed"
	VegaBackend_DiscoverSchedule_InternalError_UpdateFailed = "VegaBackend.DiscoverSchedule.InternalError.UpdateFailed"
	VegaBackend_DiscoverSchedule_InternalError_DeleteFailed = "VegaBackend.DiscoverSchedule.InternalError.DeleteFailed"
)

var (
	DiscoverScheduleErrCodeList = []string{
		VegaBackend_DiscoverSchedule_NotFound,
		VegaBackend_DiscoverSchedule_InvalidCronExpr,
		VegaBackend_DiscoverSchedule_InvalidStrategies,
		VegaBackend_DiscoverSchedule_IdMismatch,
		VegaBackend_DiscoverSchedule_CatalogMismatch,
		VegaBackend_DiscoverSchedule_EnabledFieldNotAllowed,
		VegaBackend_DiscoverSchedule_InternalError_GetFailed,
		VegaBackend_DiscoverSchedule_InternalError_CreateFailed,
		VegaBackend_DiscoverSchedule_InternalError_UpdateFailed,
		VegaBackend_DiscoverSchedule_InternalError_DeleteFailed,
	}
)

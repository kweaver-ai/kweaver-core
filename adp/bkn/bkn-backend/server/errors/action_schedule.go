// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package errors

const (
	// 400 Bad Request
	BknBackend_ActionSchedule_InvalidParameter      = "BknBackend.ActionSchedule.InvalidParameter"
	BknBackend_ActionSchedule_InvalidCronExpression = "BknBackend.ActionSchedule.InvalidCronExpression"
	BknBackend_ActionSchedule_InvalidStatus         = "BknBackend.ActionSchedule.InvalidStatus"
	BknBackend_ActionSchedule_ActionTypeNotFound    = "BknBackend.ActionSchedule.ActionTypeNotFound"

	// 404 Not Found
	BknBackend_ActionSchedule_NotFound = "BknBackend.ActionSchedule.NotFound"

	// 500 Internal Server Error
	BknBackend_ActionSchedule_InternalError       = "BknBackend.ActionSchedule.InternalError"
	BknBackend_ActionSchedule_CreateFailed        = "BknBackend.ActionSchedule.CreateFailed"
	BknBackend_ActionSchedule_UpdateFailed        = "BknBackend.ActionSchedule.UpdateFailed"
	BknBackend_ActionSchedule_DeleteFailed        = "BknBackend.ActionSchedule.DeleteFailed"
	BknBackend_ActionSchedule_GetFailed           = "BknBackend.ActionSchedule.GetFailed"
	BknBackend_ActionSchedule_GetActionTypeFailed = "BknBackend.ActionSchedule.GetActionTypeFailed"
)

var (
	actionScheduleErrCodeList = []string{
		BknBackend_ActionSchedule_InvalidParameter,
		BknBackend_ActionSchedule_InvalidCronExpression,
		BknBackend_ActionSchedule_InvalidStatus,
		BknBackend_ActionSchedule_ActionTypeNotFound,
		BknBackend_ActionSchedule_NotFound,
		BknBackend_ActionSchedule_InternalError,
		BknBackend_ActionSchedule_CreateFailed,
		BknBackend_ActionSchedule_UpdateFailed,
		BknBackend_ActionSchedule_DeleteFailed,
		BknBackend_ActionSchedule_GetFailed,
		BknBackend_ActionSchedule_GetActionTypeFailed,
	}
)

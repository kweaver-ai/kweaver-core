// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package errors 服务错误码
package errors

// Task 相关错误码
const (
	// 404 Not Found
	VegaBackend_Task_NotFound         = "VegaBackend.Task.NotFound"
	VegaBackend_BuildTask_NotFound    = "VegaBackend.BuildTask.NotFound"
	VegaBackend_DiscoverTask_NotFound = "VegaBackend.DiscoverTask.NotFound"

	// 400 Bad Request
	VegaBackend_BuildTask_Exist                       = "VegaBackend.BuildTask.Exist"
	VegaBackend_BuildTask_Running                     = "VegaBackend.BuildTask.Running"
	VegaBackend_BuildTask_InvalidStatus               = "VegaBackend.BuildTask.InvalidStatus"
	VegaBackend_BuildTask_InvalidExecuteType          = "VegaBackend.BuildTask.InvalidExecuteType"
	VegaBackend_BuildTask_InvalidParameter_ResourceID = "VegaBackend.BuildTask.InvalidParameter.ResourceID"
	VegaBackend_BuildTask_InvalidParameter_Mode       = "VegaBackend.BuildTask.InvalidParameter.Mode"
	VegaBackend_DiscoverTask_InvalidStatus            = "VegaBackend.DiscoverTask.InvalidStatus"

	// 409 Conflict
	VegaBackend_BuildTask_InvalidStateTransition = "VegaBackend.BuildTask.InvalidStateTransition"
	VegaBackend_BuildTask_HasRunningExecution    = "VegaBackend.BuildTask.HasRunningExecution"
	VegaBackend_DiscoverTask_HasRunningExecution = "VegaBackend.DiscoverTask.HasRunningExecution"

	// 500 Internal Server Error
	VegaBackend_BuildTask_InternalError_CreateFailed    = "VegaBackend.BuildTask.InternalError.CreateFailed"
	VegaBackend_BuildTask_InternalError_GetFailed       = "VegaBackend.BuildTask.InternalError.GetFailed"
	VegaBackend_BuildTask_InternalError_UpdateFailed    = "VegaBackend.BuildTask.InternalError.UpdateFailed"
	VegaBackend_BuildTask_InternalError_DeleteFailed    = "VegaBackend.BuildTask.InternalError.DeleteFailed"
	VegaBackend_DiscoverTask_InternalError_GetFailed    = "VegaBackend.DiscoverTask.InternalError.GetFailed"
	VegaBackend_DiscoverTask_InternalError_DeleteFailed = "VegaBackend.DiscoverTask.InternalError.DeleteFailed"
)

var (
	TaskErrCodeList = []string{
		VegaBackend_Task_NotFound,
		VegaBackend_BuildTask_NotFound,
		VegaBackend_DiscoverTask_NotFound,
		VegaBackend_BuildTask_Exist,
		VegaBackend_BuildTask_Running,
		VegaBackend_BuildTask_InvalidStatus,
		VegaBackend_BuildTask_InvalidExecuteType,
		VegaBackend_BuildTask_InvalidParameter_ResourceID,
		VegaBackend_BuildTask_InvalidParameter_Mode,
		VegaBackend_DiscoverTask_InvalidStatus,
		VegaBackend_BuildTask_InvalidStateTransition,
		VegaBackend_BuildTask_HasRunningExecution,
		VegaBackend_DiscoverTask_HasRunningExecution,
		VegaBackend_BuildTask_InternalError_CreateFailed,
		VegaBackend_BuildTask_InternalError_GetFailed,
		VegaBackend_BuildTask_InternalError_UpdateFailed,
		VegaBackend_BuildTask_InternalError_DeleteFailed,
		VegaBackend_DiscoverTask_InternalError_GetFailed,
		VegaBackend_DiscoverTask_InternalError_DeleteFailed,
	}
)

// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package errors 服务错误码
package errors

// Task 相关错误码
const (
	// 404 Not Found
	VegaBackend_Task_NotFound = "VegaBackend.Task.NotFound"

	// 400 Bad Request
	VegaBackend_BuildTask_Exist = "VegaBackend.BuildTask.Exist"

	// 500 Internal Server Error
	VegaBackend_BuildTask_InternalError_CreateFailed = "VegaBackend.BuildTask.InternalError.CreateFailed"
	VegaBackend_BuildTask_InternalError_GetFailed    = "VegaBackend.BuildTask.InternalError.GetFailed"
)

var (
	TaskErrCodeList = []string{
		VegaBackend_Task_NotFound,
		VegaBackend_BuildTask_Exist,
		VegaBackend_BuildTask_InternalError_CreateFailed,
		VegaBackend_BuildTask_InternalError_GetFailed,
	}
)

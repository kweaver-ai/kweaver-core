// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package errors

const (
	// 400
	BknBackend_Job_InvalidParameter                  = "BknBackend.Job.InvalidParameter"
	BknBackend_Job_InvalidParameter_JobType          = "BknBackend.Job.InvalidParameter.JobType"
	BknBackend_Job_InvalidParameter_JobState         = "BknBackend.Job.InvalidParameter.JobState"
	BknBackend_Job_InvalidParameter_JobConceptConfig = "BknBackend.Job.InvalidParameter.JobConceptConfig"
	BknBackend_Job_InvalidParameter_TaskState        = "BknBackend.Job.InvalidParameter.TaskState"
	BknBackend_Job_InvalidParameter_ConceptType      = "BknBackend.Job.InvalidParameter.ConceptType"
	BknBackend_Job_NullParameter_Name                = "BknBackend.Job.NullParameter.Name"
	BknBackend_Job_LengthExceeded_Name               = "BknBackend.Job.LengthExceeded.Name"
	BknBackend_Job_NoneConceptType                   = "BknBackend.Job.NoneConceptType"
	BknBackend_Job_InvalidObjectType                 = "BknBackend.Job.InvalidObjectType"

	// 403
	BknBackend_Job_CreateConflict = "BknBackend.Job.CreateConflict"
	BknBackend_Job_JobRunning     = "BknBackend.Job.JobRunning"

	// 404
	BknBackend_Job_JobNotFound = "BknBackend.Job.JobNotFound"

	// 500
	BknBackend_Job_InternalError                         = "BknBackend.Job.InternalError"
	BknBackend_Job_InternalError_BeginTransactionFailed  = "BknBackend.Job.InternalError.BeginTransactionFailed"
	BknBackend_Job_InternalError_CommitTransactionFailed = "BknBackend.Job.InternalError.CommitTransactionFailed"
	BknBackend_Job_InternalError_MissingTransaction      = "BknBackend.Job.InternalError.MissingTransaction"
)

var (
	JobErrCodeList = []string{
		BknBackend_Job_InvalidParameter,
		BknBackend_Job_InvalidParameter_JobType,
		BknBackend_Job_InvalidParameter_JobState,
		BknBackend_Job_InvalidParameter_JobConceptConfig,
		BknBackend_Job_InvalidParameter_TaskState,
		BknBackend_Job_InvalidParameter_ConceptType,
		BknBackend_Job_NullParameter_Name,
		BknBackend_Job_LengthExceeded_Name,
		BknBackend_Job_NoneConceptType,
		BknBackend_Job_InvalidObjectType,

		BknBackend_Job_CreateConflict,
		BknBackend_Job_JobRunning,

		BknBackend_Job_JobNotFound,

		BknBackend_Job_InternalError,
		BknBackend_Job_InternalError_BeginTransactionFailed,
		BknBackend_Job_InternalError_CommitTransactionFailed,
		BknBackend_Job_InternalError_MissingTransaction,
	}
)

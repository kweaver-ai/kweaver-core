// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package errors

// Action Execution 错误码
const (
	// 400
	OntologyQuery_ActionExecution_InvalidParameter = "OntologyQuery.ActionExecution.InvalidParameter"

	// 404
	OntologyQuery_ActionExecution_ActionTypeNotFound = "OntologyQuery.ActionExecution.ActionTypeNotFound"
	OntologyQuery_ActionExecution_ExecutionNotFound  = "OntologyQuery.ActionExecution.ExecutionNotFound"

	// 409
	OntologyQuery_ActionExecution_DuplicateExecution = "OntologyQuery.ActionExecution.DuplicateExecution"

	// 500
	OntologyQuery_ActionExecution_GetActionTypeFailed   = "OntologyQuery.ActionExecution.GetActionTypeFailed"
	OntologyQuery_ActionExecution_CreateExecutionFailed = "OntologyQuery.ActionExecution.CreateExecutionFailed"
	OntologyQuery_ActionExecution_ExecuteToolFailed     = "OntologyQuery.ActionExecution.ExecuteToolFailed"
	OntologyQuery_ActionExecution_ExecuteMCPFailed      = "OntologyQuery.ActionExecution.ExecuteMCPFailed"
	OntologyQuery_ActionExecution_QueryExecutionsFailed = "OntologyQuery.ActionExecution.QueryExecutionsFailed"
	OntologyQuery_ActionExecution_CancelExecutionFailed = "OntologyQuery.ActionExecution.CancelExecutionFailed"
)

var (
	actionExecutionErrCodeList = []string{
		// 400
		OntologyQuery_ActionExecution_InvalidParameter,

		// 404
		OntologyQuery_ActionExecution_ActionTypeNotFound,
		OntologyQuery_ActionExecution_ExecutionNotFound,

		// 409
		OntologyQuery_ActionExecution_DuplicateExecution,

		// 500
		OntologyQuery_ActionExecution_GetActionTypeFailed,
		OntologyQuery_ActionExecution_CreateExecutionFailed,
		OntologyQuery_ActionExecution_ExecuteToolFailed,
		OntologyQuery_ActionExecution_ExecuteMCPFailed,
		OntologyQuery_ActionExecution_QueryExecutionsFailed,
		OntologyQuery_ActionExecution_CancelExecutionFailed,
	}
)

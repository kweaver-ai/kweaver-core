// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package errors

// 指标模型
const (
	// 400
	OntologyQuery_ActionType_InvalidParameter                    = "OntologyQuery.ActionType.InvalidParameter"
	OntologyQuery_ActionType_InvalidParameter_DynamicParams      = "OntologyQuery.ActionType.InvalidParameter.DynamicParams"
	OntologyQuery_ActionType_InvalidParameter_IgnoringStoreCache = "OntologyQuery.ActionType.InvalidParameter.IgnoringStoreCache"
	OntologyQuery_ActionType_InvalidParameter_IncludeTypeInfo    = "OntologyQuery.ActionType.InvalidParameter.IncludeTypeInfo"

	//404
	OntologyQuery_ActionType_ActionTypeNotFound = "OntologyQuery.ActionType.ActionTypeNotFound"

	// 500
	OntologyQuery_ActionType_InternalError                          = "OntologyQuery.ActionType.InternalError"
	OntologyQuery_ActionType_InternalError_GetActionTypesByIDFailed = "OntologyQuery.ActionType.InternalError.GetActionTypesByIDFailed"
)

var (
	actionTypeErrCodeList = []string{
		// 400
		OntologyQuery_ActionType_InvalidParameter,
		OntologyQuery_ActionType_InvalidParameter_DynamicParams,
		OntologyQuery_ActionType_InvalidParameter_IgnoringStoreCache,
		OntologyQuery_ActionType_InvalidParameter_IncludeTypeInfo,

		// 404
		OntologyQuery_ActionType_ActionTypeNotFound,

		// 500
		OntologyQuery_ActionType_InternalError,
		OntologyQuery_ActionType_InternalError_GetActionTypesByIDFailed,
	}
)

// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package errors

// 指标模型
const (
	// 400
	OntologyQuery_ObjectType_InvalidParameter                      = "OntologyQuery.ObjectType.InvalidParameter"
	OntologyQuery_ObjectType_InvalidParameter_DynamicParams        = "OntologyQuery.ObjectType.InvalidParameter.DynamicParams"
	OntologyQuery_ObjectType_InvalidParameter_IgnoringStoreCache   = "OntologyQuery.ObjectType.InvalidParameter.IgnoringStoreCache"
	OntologyQuery_ObjectType_InvalidParameter_IncludeTypeInfo      = "OntologyQuery.ObjectType.InvalidParameter.IncludeTypeInfo"
	OntologyQuery_ObjectType_InvalidParameter_SmallModel           = "OntologyQuery.ObjectType.InvalidParameter.SmallModel"
	OntologyQuery_ObjectType_UnsupportLogicPropertyParameterSource = "OntologyQuery.ObjectType.UnsupportLogicPropertyParameterSource"

	//404
	OntologyQuery_ObjectType_ObjectTypeNotFound = "OntologyQuery.ObjectType.ObjectTypeNotFound"
	OntologyQuery_ObjectType_SmallModelNotFound = "OntologyQuery.ObjectType.SmallModelNotFound"

	// 500
	OntologyQuery_ObjectType_InternalError_ExecuteOperatorFailed        = "OntologyQuery.ObjectType.InternalError.ExecuteOperatorFailed"
	OntologyQuery_ObjectType_InternalError_GetAgentOperatorByIDFailed   = "OntologyQuery.ObjectType.InternalError.GetAgentOperatorByIDFailed"
	OntologyQuery_ObjectType_InternalError_GetMetricDataByIDFailed      = "OntologyQuery.ObjectType.InternalError.GetMetricDataByIDFailed"
	OntologyQuery_ObjectType_InternalError_GetViewDataByIDFailed        = "OntologyQuery.ObjectType.InternalError.GetViewDataByIDFailed"
	OntologyQuery_ObjectType_InternalError_GetObjectTypesByIDFailed     = "OntologyQuery.ObjectType.InternalError.GetObjectTypesByIDFailed"
	OntologyQuery_ObjectType_InternalError_GetSmallModelByIDFailed      = "OntologyQuery.ObjectType.InternalError.GetSmallModelByIDFailed"
	OntologyQuery_ObjectType_InternalError_ProcessLogicPropertiesFailed = "OntologyQuery.ObjectType.InternalError.ProcessLogicPropertiesFailed"
)

var (
	objectTypeErrCodeList = []string{
		// 400
		OntologyQuery_ObjectType_InvalidParameter,
		OntologyQuery_ObjectType_InvalidParameter_DynamicParams,
		OntologyQuery_ObjectType_InvalidParameter_IgnoringStoreCache,
		OntologyQuery_ObjectType_InvalidParameter_IncludeTypeInfo,
		OntologyQuery_ObjectType_InvalidParameter_SmallModel,
		OntologyQuery_ObjectType_UnsupportLogicPropertyParameterSource,
		// OntologyQuery_ObjectType_UnsupportLogicPropertyType,

		// 404
		OntologyQuery_ObjectType_ObjectTypeNotFound,
		OntologyQuery_ObjectType_SmallModelNotFound,

		// 500
		OntologyQuery_ObjectType_InternalError_ExecuteOperatorFailed,
		OntologyQuery_ObjectType_InternalError_GetAgentOperatorByIDFailed,
		OntologyQuery_ObjectType_InternalError_GetMetricDataByIDFailed,
		OntologyQuery_ObjectType_InternalError_GetViewDataByIDFailed,
		OntologyQuery_ObjectType_InternalError_GetObjectTypesByIDFailed,
		OntologyQuery_ObjectType_InternalError_GetSmallModelByIDFailed,
		OntologyQuery_ObjectType_InternalError_ProcessLogicPropertiesFailed,
	}
)

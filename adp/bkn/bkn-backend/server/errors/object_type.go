// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package errors

// 业务知识网络错误码
const (
	// 400
	BknBackend_ObjectType_Duplicated_IDInFile               = "BknBackend.ObjectType.Duplicated.IDInFile"
	BknBackend_ObjectType_Duplicated_Name                   = "BknBackend.ObjectType.Duplicated.Name"
	BknBackend_ObjectType_InvalidParameter                  = "BknBackend.ObjectType.InvalidParameter"
	BknBackend_ObjectType_InvalidParameter_ConceptCondition = "BknBackend.ObjectType.InvalidParameter.ConceptCondition"
	BknBackend_ObjectType_InvalidParameter_PropertyName     = "BknBackend.ObjectType.InvalidParameter.PropertyName"
	BknBackend_ObjectType_InvalidParameter_SmallModel       = "BknBackend.ObjectType.InvalidParameter.SmallModel"
	BknBackend_ObjectType_LengthExceeded_Name               = "BknBackend.ObjectType.LengthExceeded.Name"
	BknBackend_ObjectType_NullParameter_Name                = "BknBackend.ObjectType.NullParameter.Name"
	BknBackend_ObjectType_NullParameter_PrimaryKeys         = "BknBackend.ObjectType.NullParameter.PrimaryKeys"
	BknBackend_ObjectType_NullParameter_DisplayKey          = "BknBackend.ObjectType.NullParameter.DisplayKey"
	BknBackend_ObjectType_NullParameter_PropertyName        = "BknBackend.ObjectType.NullParameter.PropertyName"
	BknBackend_ObjectType_ObjectTypeIDExisted               = "BknBackend.ObjectType.ObjectTypeIDExisted"
	BknBackend_ObjectType_ObjectTypeNameExisted             = "BknBackend.ObjectType.ObjectTypeNameExisted"
	BknBackend_ObjectType_ObjectTypeBoundByActionType       = "BknBackend.ObjectType.ObjectTypeBoundByActionType"
	BknBackend_ObjectType_ObjectTypeBoundByRelationType     = "BknBackend.ObjectType.ObjectTypeBoundByRelationType"

	// 404
	BknBackend_ObjectType_ObjectTypeNotFound = "BknBackend.ObjectType.ObjectTypeNotFound"
	BknBackend_ObjectType_SmallModelNotFound = "BknBackend.ObjectType.SmallModelNotFound"

	// 500
	BknBackend_ObjectType_InternalError                                  = "BknBackend.ObjectType.InternalError"
	BknBackend_ObjectType_InternalError_MissingTransaction               = "BknBackend.ObjectType.InternalError.MissingTransaction"
	BknBackend_ObjectType_InternalError_BeginTransactionFailed           = "BknBackend.ObjectType.InternalError.BeginTransactionFailed"
	BknBackend_ObjectType_InternalError_CheckObjectTypeIfExistFailed     = "BknBackend.ObjectType.InternalError.CheckObjectTypeIfExistFailed"
	BknBackend_ObjectType_InternalError_CreateConceptGroupRelationFailed = "BknBackend.ObjectType.InternalError.CreateConceptGroupRelationFailed"
	BknBackend_ObjectType_InternalError_GetDataViewByIDFailed            = "BknBackend.ObjectType.InternalError.GetDataViewByIDFailed"
	BknBackend_ObjectType_InternalError_GetMetricModelByIDFailed         = "BknBackend.ObjectType.InternalError.GetMetricModelByIDFailed"
	BknBackend_ObjectType_InternalError_GetObjectTypeByIDFailed          = "BknBackend.ObjectType.InternalError.GetObjectTypeByIDFailed"
	BknBackend_ObjectType_InternalError_GetObjectTypesByIDsFailed        = "BknBackend.ObjectType.InternalError.GetObjectTypesByIDsFailed"
	BknBackend_ObjectType_InternalError_GetSmallModelByIDFailed          = "BknBackend.ObjectType.InternalError.GetSmallModelByIDFailed"
	BknBackend_ObjectType_InternalError_InsertOpenSearchDataFailed       = "BknBackend.ObjectType.InternalError.InsertOpenSearchDataFailed"
)

var (
	ObjectTypeErrCodeList = []string{
		// 400
		BknBackend_ObjectType_Duplicated_IDInFile,
		BknBackend_ObjectType_Duplicated_Name,
		BknBackend_ObjectType_InvalidParameter,
		BknBackend_ObjectType_InvalidParameter_ConceptCondition,
		BknBackend_ObjectType_InvalidParameter_PropertyName,
		BknBackend_ObjectType_InvalidParameter_SmallModel,
		BknBackend_ObjectType_LengthExceeded_Name,
		BknBackend_ObjectType_NullParameter_Name,
		BknBackend_ObjectType_NullParameter_PrimaryKeys,
		BknBackend_ObjectType_NullParameter_DisplayKey,
		BknBackend_ObjectType_NullParameter_PropertyName,
		BknBackend_ObjectType_ObjectTypeIDExisted,
		BknBackend_ObjectType_ObjectTypeNameExisted,
		BknBackend_ObjectType_ObjectTypeBoundByActionType,
		BknBackend_ObjectType_ObjectTypeBoundByRelationType,

		// 404
		BknBackend_ObjectType_ObjectTypeNotFound,
		BknBackend_ObjectType_SmallModelNotFound,

		// 500
		BknBackend_ObjectType_InternalError,
		BknBackend_ObjectType_InternalError_MissingTransaction,
		BknBackend_ObjectType_InternalError_BeginTransactionFailed,
		BknBackend_ObjectType_InternalError_CheckObjectTypeIfExistFailed,
		BknBackend_ObjectType_InternalError_CreateConceptGroupRelationFailed,
		BknBackend_ObjectType_InternalError_GetDataViewByIDFailed,
		BknBackend_ObjectType_InternalError_GetMetricModelByIDFailed,
		BknBackend_ObjectType_InternalError_GetObjectTypeByIDFailed,
		BknBackend_ObjectType_InternalError_GetObjectTypesByIDsFailed,
		BknBackend_ObjectType_InternalError_GetSmallModelByIDFailed,
		BknBackend_ObjectType_InternalError_InsertOpenSearchDataFailed,
	}
)

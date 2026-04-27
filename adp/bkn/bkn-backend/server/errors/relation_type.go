// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package errors

// 业务知识网络错误码
const (
	// 400
	BknBackend_RelationType_Duplicated_IDInFile               = "BknBackend.RelationType.Duplicated.IDInFile"
	BknBackend_RelationType_InvalidParameter                  = "BknBackend.RelationType.InvalidParameter"
	BknBackend_RelationType_InvalidParameter_ConceptCondition = "BknBackend.RelationType.InvalidParameter.ConceptCondition"
	BknBackend_RelationType_RelationTypeIDExisted             = "BknBackend.RelationType.RelationTypeIDExisted"
	BknBackend_RelationType_LengthExceeded_Name               = "BknBackend.RelationType.LengthExceeded.Name"
	BknBackend_RelationType_NullParameter_Name                = "BknBackend.RelationType.NullParameter.Name"

	// 500
	BknBackend_RelationType_InternalError                                = "BknBackend.RelationType.InternalError"
	BknBackend_RelationType_InternalError_MissingTransaction             = "BknBackend.RelationType.InternalError.MissingTransaction"
	BknBackend_RelationType_InternalError_BeginTransactionFailed         = "BknBackend.RelationType.InternalError.BeginTransactionFailed"
	BknBackend_RelationType_InternalError_CheckRelationTypeIfExistFailed = "BknBackend.RelationType.InternalError.CheckRelationTypeIfExistFailed"
	BknBackend_RelationType_InternalError_GetDataViewByIDFailed          = "BknBackend.RelationType.InternalError.GetDataViewByIDFailed"
	BknBackend_RelationType_InternalError_GetRelationTypesByIDsFailed    = "BknBackend.RelationType.InternalError.GetRelationTypesByIDsFailed"
	BknBackend_RelationType_InternalError_InsertOpenSearchDataFailed     = "BknBackend.RelationType.InternalError.InsertOpenSearchDataFailed"
	BknBackend_RelationType_RelationTypeNotFound                         = "BknBackend.RelationType.RelationTypeNotFound"
)

var (
	RelationTypeErrCodeList = []string{
		// 400
		BknBackend_RelationType_Duplicated_IDInFile,
		BknBackend_RelationType_InvalidParameter,
		BknBackend_RelationType_InvalidParameter_ConceptCondition,
		BknBackend_RelationType_RelationTypeIDExisted,
		BknBackend_RelationType_LengthExceeded_Name,
		BknBackend_RelationType_NullParameter_Name,

		// 500
		BknBackend_RelationType_InternalError,
		BknBackend_RelationType_InternalError_MissingTransaction,
		BknBackend_RelationType_InternalError_CheckRelationTypeIfExistFailed,
		BknBackend_RelationType_InternalError_BeginTransactionFailed,
		BknBackend_RelationType_InternalError_GetDataViewByIDFailed,
		BknBackend_RelationType_InternalError_GetRelationTypesByIDsFailed,
		BknBackend_RelationType_InternalError_InsertOpenSearchDataFailed,
		BknBackend_RelationType_RelationTypeNotFound,
	}
)

// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package errors

// 业务知识网络错误码
const (
	// 400
	BknBackend_ActionType_Duplicated_IDInFile               = "BknBackend.ActionType.Duplicated.IDInFile"
	BknBackend_ActionType_Duplicated_Name                   = "BknBackend.ActionType.Duplicated.Name"
	BknBackend_ActionType_InvalidParameter                  = "BknBackend.ActionType.InvalidParameter"
	BknBackend_ActionType_InvalidParameter_ConceptCondition = "BknBackend.ActionType.InvalidParameter.ConceptCondition"
	BknBackend_ActionType_ActionTypeIDExisted               = "BknBackend.ActionType.ActionTypeIDExisted"
	BknBackend_ActionType_ActionTypeNameExisted             = "BknBackend.ActionType.ActionTypeNameExisted"
	BknBackend_ActionType_LengthExceeded_Name               = "BknBackend.ActionType.LengthExceeded.Name"
	BknBackend_ActionType_NullParameter_Name                = "BknBackend.ActionType.NullParameter.Name"

	// 500
	BknBackend_ActionType_InternalError                              = "BknBackend.ActionType.InternalError"
	BknBackend_ActionType_InternalError_MissingTransaction           = "BknBackend.ActionType.InternalError.MissingTransaction"
	BknBackend_ActionType_InternalError_BeginTransactionFailed       = "BknBackend.ActionType.InternalError.BeginTransactionFailed"
	BknBackend_ActionType_InternalError_CheckActionTypeIfExistFailed = "BknBackend.ActionType.InternalError.CheckActionTypeIfExistFailed"
	BknBackend_ActionType_InternalError_GetActionTypesByIDsFailed    = "BknBackend.ActionType.InternalError.GetActionTypesByIDsFailed"
	BknBackend_ActionType_InternalError_InsertOpenSearchDataFailed   = "BknBackend.ActionType.InternalError.InsertOpenSearchDataFailed"
	BknBackend_ActionType_ActionTypeNotFound                         = "BknBackend.ActionType.ActionTypeNotFound"
)

var (
	ActionTypeErrCodeList = []string{
		// 400
		BknBackend_ActionType_Duplicated_IDInFile,
		BknBackend_ActionType_Duplicated_Name,
		BknBackend_ActionType_InvalidParameter,
		BknBackend_ActionType_InvalidParameter_ConceptCondition,
		BknBackend_ActionType_ActionTypeIDExisted,
		BknBackend_ActionType_ActionTypeNameExisted,
		BknBackend_ActionType_LengthExceeded_Name,
		BknBackend_ActionType_NullParameter_Name,

		// 500
		BknBackend_ActionType_InternalError,
		BknBackend_ActionType_InternalError_MissingTransaction,
		BknBackend_ActionType_InternalError_CheckActionTypeIfExistFailed,
		BknBackend_ActionType_InternalError_BeginTransactionFailed,
		BknBackend_ActionType_InternalError_GetActionTypesByIDsFailed,
		BknBackend_ActionType_InternalError_InsertOpenSearchDataFailed,
		BknBackend_ActionType_ActionTypeNotFound,
	}
)

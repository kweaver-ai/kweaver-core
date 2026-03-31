// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package errors

// 风险类错误码
const (
	// 400
	BknBackend_RiskType_Duplicated_IDInFile       = "BknBackend.RiskType.Duplicated.IDInFile"
	BknBackend_RiskType_Duplicated_Name           = "BknBackend.RiskType.Duplicated.Name"
	BknBackend_RiskType_InvalidParameter          = "BknBackend.RiskType.InvalidParameter"
	BknBackend_RiskType_RiskTypeIDExisted         = "BknBackend.RiskType.RiskTypeIDExisted"
	BknBackend_RiskType_RiskTypeNameExisted       = "BknBackend.RiskType.RiskTypeNameExisted"
	BknBackend_RiskType_RiskFunctionToolNotFound  = "BknBackend.RiskType.RiskFunctionToolNotFound"
	BknBackend_RiskType_InvalidMaxAcceptableLevel = "BknBackend.RiskType.InvalidMaxAcceptableLevel"
	BknBackend_RiskType_NullParameter_Name        = "BknBackend.RiskType.NullParameter.Name"
	BknBackend_RiskType_LengthExceeded_Name       = "BknBackend.RiskType.LengthExceeded.Name"

	// 500
	BknBackend_RiskType_InternalError                            = "BknBackend.RiskType.InternalError"
	BknBackend_RiskType_InternalError_CheckRiskTypeIfExistFailed = "BknBackend.RiskType.InternalError.CheckRiskTypeIfExistFailed"
	BknBackend_RiskType_InternalError_GetRiskTypesByIDsFailed    = "BknBackend.RiskType.InternalError.GetRiskTypesByIDsFailed"
	BknBackend_RiskType_RiskTypeNotFound                         = "BknBackend.RiskType.RiskTypeNotFound"
)

var (
	RiskTypeErrCodeList = []string{
		BknBackend_RiskType_Duplicated_IDInFile,
		BknBackend_RiskType_Duplicated_Name,
		BknBackend_RiskType_InvalidParameter,
		BknBackend_RiskType_RiskTypeIDExisted,
		BknBackend_RiskType_RiskTypeNameExisted,
		BknBackend_RiskType_RiskFunctionToolNotFound,
		BknBackend_RiskType_InvalidMaxAcceptableLevel,
		BknBackend_RiskType_NullParameter_Name,
		BknBackend_RiskType_LengthExceeded_Name,
		BknBackend_RiskType_InternalError,
		BknBackend_RiskType_InternalError_CheckRiskTypeIfExistFailed,
		BknBackend_RiskType_InternalError_GetRiskTypesByIDsFailed,
		BknBackend_RiskType_RiskTypeNotFound,
	}
)

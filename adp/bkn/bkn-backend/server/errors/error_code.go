// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package errors 服务错误码
package errors

import (
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	"bkn-backend/locale"
)

// 公共错误码, 服务内所有模块均可使用
const (
	// 400
	BknBackend_CountExceeded_TagTotal          = "BknBackend.CountExceeded.TagTotal"
	BknBackend_InvalidParameter_Condition      = "BknBackend.InvalidParameter.Condition"
	BknBackend_InvalidParameter_DataTagName    = "BknBackend.InvalidParameter.DataTagName"
	BknBackend_InvalidParameter_Direction      = "BknBackend.InvalidParameter.Direction"
	BknBackend_InvalidParameter_ID             = "BknBackend.InvalidParameter.ID"
	BknBackend_InvalidParameter_ImportMode     = "BknBackend.InvalidParameter.ImportMode"
	BknBackend_InvalidParameter_Mode           = "BknBackend.InvalidParameter.Mode"
	BknBackend_InvalidParameter_Limit          = "BknBackend.InvalidParameter.Limit"
	BknBackend_InvalidParameter_ModuleType     = "BknBackend.InvalidParameter.ModuleType"
	BknBackend_InvalidParameter_Offset         = "BknBackend.InvalidParameter.Offset"
	BknBackend_InvalidParameter_OverrideMethod = "BknBackend.InvalidParameter.OverrideMethod"
	BknBackend_InvalidParameter_RequestBody    = "BknBackend.InvalidParameter.RequestBody"
	BknBackend_InvalidParameter_Sort           = "BknBackend.InvalidParameter.Sort"
	BknBackend_InvalidParameter_ConditionValue = "BknBackend.InvalidParameter.ConditionValue"

	BknBackend_NullParameter_ConditionName      = "BknBackend.NullParameter.ConditionName"
	BknBackend_NullParameter_ConditionOperation = "BknBackend.NullParameter.ConditionOperation"
	BknBackend_UnsupportConditionOperation      = "BknBackend.UnsupportConditionOperation"
	BknBackend_CountExceeded_Conditions         = "BknBackend.CountExceeded.Conditions"

	// 406
	BknBackend_InvalidRequestHeader_ContentType = "BknBackend.InvalidRequestHeader.ContentType"

	// Permission
	BknBackend_InternalError_CheckPermissionFailed = "BknBackend.InternalError.CheckPermissionFailed"
	BknBackend_InternalError_CreateResourcesFailed = "BknBackend.InternalError.CreateResourcesFailed"
	BknBackend_InternalError_DeleteResourcesFailed = "BknBackend.InternalError.DeleteResourcesFailed"
	BknBackend_InternalError_FilterResourcesFailed = "BknBackend.InternalError.FilterResourcesFailed"
	BknBackend_InternalError_UpdateResourceFailed  = "BknBackend.InternalError.UpdateResourceFailed"
	BknBackend_InternalError_MQPublishMsgFailed    = "BknBackend.InternalError.MQPublishMsgFailed"

	// 500
	BknBackend_InternalError_MarshalDataFailed   = "BknBackend.InternalError.MarshalDataFailed"
	BknBackend_InternalError_UnMarshalDataFailed = "BknBackend.InternalError.UnMarshalDataFailed"
)

var (
	errCodeList = []string{
		// ---公共错误码---
		// 400
		BknBackend_CountExceeded_TagTotal,
		BknBackend_InvalidParameter_Direction,
		BknBackend_InvalidParameter_Condition,
		BknBackend_InvalidParameter_DataTagName,
		BknBackend_InvalidParameter_ID,
		BknBackend_InvalidParameter_ImportMode,
		BknBackend_InvalidParameter_Mode,
		BknBackend_InvalidParameter_Limit,
		BknBackend_InvalidParameter_ModuleType,
		BknBackend_InvalidParameter_Offset,
		BknBackend_InvalidParameter_OverrideMethod,
		BknBackend_InvalidParameter_RequestBody,
		BknBackend_InvalidParameter_Sort,

		BknBackend_NullParameter_ConditionName,
		BknBackend_NullParameter_ConditionOperation,
		BknBackend_UnsupportConditionOperation,
		BknBackend_CountExceeded_Conditions,
		BknBackend_InvalidParameter_ConditionValue,

		// permission
		BknBackend_InternalError_CheckPermissionFailed,
		BknBackend_InternalError_CreateResourcesFailed,
		BknBackend_InternalError_DeleteResourcesFailed,
		BknBackend_InternalError_FilterResourcesFailed,
		BknBackend_InternalError_UpdateResourceFailed,
		BknBackend_InternalError_MQPublishMsgFailed,

		// 406
		BknBackend_InvalidRequestHeader_ContentType,

		// 500
		BknBackend_InternalError_MarshalDataFailed,
		BknBackend_InternalError_UnMarshalDataFailed,
	}
)

func init() {
	locale.Register()
	rest.Register(errCodeList)
	rest.Register(KNErrCodeList)
	rest.Register(ObjectTypeErrCodeList)
	rest.Register(RelationTypeErrCodeList)
	rest.Register(ActionTypeErrCodeList)
	rest.Register(actionScheduleErrCodeList)
	rest.Register(JobErrCodeList)
	rest.Register(ConceptGroupErrCodeList)
	rest.Register(RiskTypeErrCodeList)
}

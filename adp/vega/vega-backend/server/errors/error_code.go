// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package errors 服务错误码
package errors

import (
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	"vega-backend/locale"
)

// 公共错误码
const (
	// 400 Bad Request
	VegaBackend_InvalidParameter_Name                     = "VegaBackend.InvalidParameter.Name"
	VegaBackend_InvalidParameter_Description              = "VegaBackend.InvalidParameter.Description"
	VegaBackend_InvalidParameter_Tag                      = "VegaBackend.InvalidParameter.Tag"
	VegaBackend_InvalidParameter_RequestBody              = "VegaBackend.InvalidParameter.RequestBody"
	VegaBackend_InvalidParameter_ResourceType             = "VegaBackend.InvalidParameter.ResourceType"
	VegaBackend_InvalidParameter_StreamSize               = "VegaBackend.InvalidParameter.StreamSize"
	VegaBackend_InvalidParameter_Limit                    = "VegaBackend.InvalidParameter.Limit"
	VegaBackend_InvalidParameter_Offset                   = "VegaBackend.InvalidParameter.Offset"
	VegaBackend_InvalidParameter_Sort                     = "VegaBackend.InvalidParameter.Sort"
	VegaBackend_InvalidParameter_Direction                = "VegaBackend.InvalidParameter.Direction"
	VegaBackend_InvalidParameter_ID                       = "VegaBackend.InvalidParameter.ID"
	VegaBackend_InvalidParameter_OverrideMethod           = "VegaBackend.InvalidParameter.OverrideMethod"
	VegaBackend_InvalidParameter_Format                   = "VegaBackend.InvalidParameter.Format"
	VegaBackend_InvalidParameter_FilterCondition          = "VegaBackend.InvalidParameter.FilterCondition"
	VegaBackend_InvalidParameter_FilterConditionValue     = "VegaBackend.InvalidParameter.FilterConditionValue"
	VegaBackend_InvalidParameter_FilterConditionValueFrom = "VegaBackend.InvalidParameter.FilterConditionValueFrom"
	VegaBackend_NullParameter_FilterConditionName         = "VegaBackend.NullParameter.FilterConditionName"
	VegaBackend_NullParameter_FilterConditionValue        = "VegaBackend.NullParameter.FilterConditionValue"
	VegaBackend_NullParameter_FilterConditionOperation    = "VegaBackend.NullParameter.FilterConditionOperation"
	VegaBackend_CountExceeded_FilterConditionSubConds     = "VegaBackend.CountExceeded.FilterConditionSubConds"
	VegaBackend_UnsupportFilterConditionOperation         = "VegaBackend.UnsupportFilterConditionOperation"

	// 406 Not Acceptable
	VegaBackend_InvalidRequestHeader_ContentType = "VegaBackend.InvalidRequestHeader.ContentType"

	// 500 Internal Server Error
	VegaBackend_InternalError_BeginTransactionFailed  = "VegaBackend.InternalError.BeginTransactionFailed"
	VegaBackend_InternalError_CommitTransactionFailed = "VegaBackend.InternalError.CommitTransactionFailed"
	VegaBackend_InternalError_MarshalDataFailed       = "VegaBackend.InternalError.MarshalDataFailed"
	VegaBackend_InternalError_UnMarshalDataFailed     = "VegaBackend.InternalError.UnMarshalDataFailed"
	VegaBackend_InternalError_CheckPermissionFailed   = "VegaBackend.InternalError.CheckPermissionFailed"
	VegaBackend_InternalError_CreateResourcesFailed   = "VegaBackend.InternalError.CreateResourcesFailed"
	VegaBackend_InternalError_DeleteResourcesFailed   = "VegaBackend.InternalError.DeleteResourcesFailed"
	VegaBackend_InternalError_FilterResourcesFailed   = "VegaBackend.InternalError.FilterResourcesFailed"
	VegaBackend_InternalError_UpdateResourceFailed    = "VegaBackend.InternalError.UpdateResourceFailed"
)

var (
	commonErrCodeList = []string{
		VegaBackend_InvalidParameter_Name,
		VegaBackend_InvalidParameter_Description,
		VegaBackend_InvalidParameter_Tag,

		VegaBackend_InvalidParameter_RequestBody,
		VegaBackend_InvalidParameter_ResourceType,
		VegaBackend_InvalidParameter_StreamSize,
		VegaBackend_InvalidParameter_Limit,
		VegaBackend_InvalidParameter_Offset,
		VegaBackend_InvalidParameter_Sort,
		VegaBackend_InvalidParameter_Direction,
		VegaBackend_InvalidParameter_ID,
		VegaBackend_InvalidParameter_OverrideMethod,
		VegaBackend_InvalidParameter_Format,
		VegaBackend_InvalidParameter_FilterCondition,
		VegaBackend_InvalidParameter_FilterConditionValue,
		VegaBackend_InvalidParameter_FilterConditionValueFrom,
		VegaBackend_NullParameter_FilterConditionName,
		VegaBackend_NullParameter_FilterConditionValue,
		VegaBackend_NullParameter_FilterConditionOperation,
		VegaBackend_CountExceeded_FilterConditionSubConds,
		VegaBackend_UnsupportFilterConditionOperation,

		VegaBackend_InvalidRequestHeader_ContentType,

		VegaBackend_InternalError_BeginTransactionFailed,
		VegaBackend_InternalError_CommitTransactionFailed,
		VegaBackend_InternalError_MarshalDataFailed,
		VegaBackend_InternalError_UnMarshalDataFailed,
		VegaBackend_InternalError_CheckPermissionFailed,
		VegaBackend_InternalError_CreateResourcesFailed,
		VegaBackend_InternalError_DeleteResourcesFailed,
		VegaBackend_InternalError_FilterResourcesFailed,
		VegaBackend_InternalError_UpdateResourceFailed,
	}
)

func init() {
	locale.Register()
	rest.Register(commonErrCodeList)
	rest.Register(CatalogErrCodeList)
	rest.Register(ResourceErrCodeList)
	rest.Register(ConnectorTypeErrCodeList)
	rest.Register(TaskErrCodeList)
	rest.Register(QueryErrCodeList)
	rest.Register(LogicViewErrCodeList)
	rest.Register(DiscoverScheduleErrCodeList)
}

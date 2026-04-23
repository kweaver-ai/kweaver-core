// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package errors 服务错误码
package errors

import (
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	"ontology-query/locale"
)

// 公共错误码, 服务内所有模块均可使用
const (
	// 400
	OntologyQuery_InvalidParameter_Condition      = "OntologyQuery.InvalidParameter.Condition"
	OntologyQuery_InvalidParameter_OverrideMethod = "OntologyQuery.InvalidParameter.OverrideMethod"
	OntologyQuery_NullParameter_OverrideMethod    = "OntologyQuery.NullParameter.OverrideMethod"

	// 406
	OntologyQuery_InvalidRequestHeader_ContentType = "OntologyQuery.InvalidRequestHeader.ContentType"

	// Permission
	// OntologyQuery_InternalError_CheckPermissionFailed = "OntologyQuery.InternalError.CheckPermissionFailed"
	// OntologyQuery_InternalError_CreateResourcesFailed = "OntologyQuery.InternalError.CreateResourcesFailed"
	// OntologyQuery_InternalError_DeleteResourcesFailed = "OntologyQuery.InternalError.DeleteResourcesFailed"
	// OntologyQuery_InternalError_FilterResourcesFailed = "OntologyQuery.InternalError.FilterResourcesFailed"
	// OntologyQuery_InternalError_UpdateResourceFailed  = "OntologyQuery.InternalError.UpdateResourceFailed"
	// OntologyQuery_InternalError_MQPublishMsgFailed    = "OntologyQuery.InternalError.MQPublishMsgFailed"

	// 500
	OntologyQuery_InternalError                                = "OntologyQuery.InternalError"
	OntologyQuery_InternalError_MarshalDataFailed              = "OntologyQuery.InternalError.MarshalDataFailed"
	OntologyQuery_InternalError_UnMarshalDataFailed            = "OntologyQuery.InternalError.UnMarshalDataFailed"
	OntologyQuery_InternalError_SearchDataFromOpensearchFailed = "OntologyQuery.InternalError.SearchDataFromOpensearchFailed"
)

var (
	errCodeList = []string{
		// ---公共错误码---
		// 400
		OntologyQuery_InvalidParameter_Condition,
		OntologyQuery_InvalidParameter_OverrideMethod,
		OntologyQuery_NullParameter_OverrideMethod,

		// permission
		// OntologyQuery_InternalError_CheckPermissionFailed,
		// OntologyQuery_InternalError_CreateResourcesFailed,
		// OntologyQuery_InternalError_DeleteResourcesFailed,
		// OntologyQuery_InternalError_FilterResourcesFailed,
		// OntologyQuery_InternalError_UpdateResourceFailed,
		// OntologyQuery_InternalError_MQPublishMsgFailed,

		// 406
		OntologyQuery_InvalidRequestHeader_ContentType,

		// 500
		OntologyQuery_InternalError,
		OntologyQuery_InternalError_MarshalDataFailed,
		OntologyQuery_InternalError_UnMarshalDataFailed,
		OntologyQuery_InternalError_SearchDataFromOpensearchFailed,
	}
)

func init() {
	locale.Register()
	rest.Register(errCodeList)
	rest.Register(objectTypeErrCodeList)
	rest.Register(knowledgeNetworkErrCodeList)
	rest.Register(actionExecutionErrCodeList)
	rest.Register(actionTypeErrCodeList)
	rest.Register(riskTypeErrCodeList)
	rest.Register(metricErrCodeList)
}

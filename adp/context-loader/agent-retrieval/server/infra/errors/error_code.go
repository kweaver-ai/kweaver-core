// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package errors 定义错误码
// @file errors_code.go
// @description: 定义错误码
package errors

// common拓展错误码定义
const (
	ErrExtCommonOperationForbidden      = "CommonOperationForbidden"      // 没有操作权限
	ErrExtCommonAddForbidden            = "CommonAddForbidden"            // 没有新建权限
	ErrExtCommonEditForbidden           = "CommonEditForbidden"           // 没有编辑权限
	ErrExtCommonDeleteForbidden         = "CommonDeleteForbidden"         // 没有删除权限
	ErrExtCommonPublishForbidden        = "CommonPublishForbidden"        // 没有发布权限
	ErrExtCommonUnpublishForbidden      = "CommonUnpublishForbidden"      // 没有下架权限
	ErrExtCommonPermissionForbidden     = "CommonPermissionForbidden"     // 没有权限管理权限
	ErrExtCommonPublicAccessForbidden   = "CommonPublicAccessForbidden"   // 没有公共访问权限
	ErrExtCommonUseForbidden            = "CommonUseForbidden"            // 没有使用权限
	ErrExtCommonViewForbidden           = "CommonViewForbidden"           // 没有查看权限
	ErrExtCommonUserNotFound            = "CommonUserNotFound"            // 用户不存在
	ErrExtCommonAnonymousUserNotAllowed = "CommonAnonymousUserNotAllowed" // 匿名用户不允许访问
	ErrExtCommonExternalServerError     = "CommonExternalServerError"     // 外部服务异常
)

// MCP拓展错误码定义
const (
	ErrExtMCPInstanceAlreadyExists = "MCPInstanceAlreadyExists" // MCP实例已存在
	ErrExtMCPInstanceNotFound      = "MCPInstanceNotFound"      // MCP实例不存在
)

// 业务知识网络行动召回拓展错误码定义
const (
	ErrExtKnActionRecallUnsupportedType     = "KnActionRecallUnsupportedType"     // 不支持的行动源类型
	ErrExtKnActionRecallNoActionsFound      = "KnActionRecallNoActionsFound"      // 未找到可用行动
	ErrExtKnActionRecallSchemaConvertFailed = "KnActionRecallSchemaConvertFailed" // Schema转换失败
	ErrExtKnActionRecallToolNotFound        = "KnActionRecallToolNotFound"        // 工具不存在
)

// 通用错误码定义
const (
	ErrExtCommonNameInvalid = "CommonNameInvalid" // 仅支持输入中文、字母、数字、下划线或空格
)

// 验证器错误码定义
const (
	ErrExtCodeValidationRequired = "ValidationRequired" // 必填项
	ErrExtCodeValidationFormat   = "ValidationFormat"   // 格式错误
	ErrExtCodeValidationRange    = "ValidationRange"    // 范围错误
	ErrExtCodeValidationEnum     = "ValidationEnum"     // 枚举错误
)

const (
	CommonSolution = "Common"
	NoneSolution   = "None"
)

const (
	NoneErrorLink = "None"
)

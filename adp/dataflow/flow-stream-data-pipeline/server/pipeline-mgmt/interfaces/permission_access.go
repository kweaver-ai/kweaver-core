// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
)

type contextKey string // 自定义专属的key类型

const (
	ADMIN_ACCOUNT_ID   = "266c6a42-6131-4d62-8f39-853e7093701c"
	ADMIN_ACCOUNT_TYPE = "user"

	HTTP_HEADER_ACCOUNT_ID   = "x-account-id"
	HTTP_HEADER_ACCOUNT_TYPE = "x-account-type"

	ACCOUNT_INFO_KEY contextKey = "x-account-info" // 避免直接使用string

	// 访问者类型
	ACCESSOR_TYPE_USER = "user"
	ACCESSOR_TYPE_APP  = "app"

	// 创建时无资源id，用 * 表示
	RESOURCE_ID_ALL = "*"

	// 资源类型
	RESOURCE_TYPE_PIPELINE = "stream_data_pipeline"

	// 资源操作类型
	OPERATION_TYPE_VIEW_DETAIL = "view_detail"
	OPERATION_TYPE_CREATE      = "create"
	OPERATION_TYPE_MODIFY      = "modify"
	OPERATION_TYPE_DELETE      = "delete"
	OPERATION_TYPE_DATA_QUERY  = "data_query"
	OPERATION_TYPE_AUTHORIZE   = "authorize"

	// 更新资源名称的topic
	AUTHORIZATION_RESOURCE_NAME_MODIFY = "authorization.resource.name.modify"
)

var (
	COMMON_OPERATIONS = []string{
		OPERATION_TYPE_VIEW_DETAIL,
		OPERATION_TYPE_CREATE,
		OPERATION_TYPE_MODIFY,
		OPERATION_TYPE_DELETE,
		OPERATION_TYPE_AUTHORIZE,
	}
)

// 检查权限
type PermissionCheck struct {
	Accessor   PermissionAccessor `json:"accessor"`
	Resource   PermissionResource `json:"resource"`
	Operations []string           `json:"operation"`
	Method     string             `json:"method"`
}

// 检查权限结果
type PermissionCheckResult struct {
	Result bool `json:"result"`
}

// 访问者信息
type PermissionAccessor struct {
	Type string `json:"type,omitempty"` // 分 user: 实名， app: 应用账户
	ID   string `json:"id,omitempty"`   // 用户ID
}

// 资源信息
type PermissionResource struct {
	Type string `json:"type,omitempty"` // 资源类型
	ID   string `json:"id,omitempty"`   // 资源ID
	Name string `json:"name,omitempty"` // 资源名称
}

// 过滤/删除
type PermissionResourcesFilter struct {
	Accessor       PermissionAccessor   `json:"accessor,omitempty"`
	Resources      []PermissionResource `json:"resources,omitempty"`
	Operations     []string             `json:"operation,omitempty"`
	AllowOperation bool                 `json:"allow_operation"`
	Method         string               `json:"method,omitempty"`
}

// 设置权限
type PermissionPolicy struct {
	Accessor   PermissionAccessor  `json:"accessor"`
	Resource   PermissionResource  `json:"resource"`
	Operations PermissionPolicyOps `json:"operation"`
	Condition  string              `json:"condition"`
	ExpiresAt  string              `json:"expires_at,omitempty"`
}

type PermissionPolicyOps struct {
	Allow []PermissionOperation `json:"allow"`
	Deny  []PermissionOperation `json:"deny"`
}

type PermissionOperation struct {
	Operation string `json:"id"`
}

type PermissionResourceOps struct {
	ResourceID string   `json:"id"`
	Operations []string `json:"operation,omitempty"`
}

//go:generate mockgen -source ../interfaces/permission_access.go -destination ../interfaces/mock/mock_permission_access.go
type PermissionAccess interface {
	CheckPermission(ctx context.Context, check PermissionCheck) (bool, error)
	FilterResources(ctx context.Context, filter PermissionResourcesFilter) (map[string]PermissionResourceOps, error)
	GetResourcesOperations(ctx context.Context, filter PermissionResourcesFilter) (map[string]PermissionResourceOps, error)

	CreateResources(ctx context.Context, policies []PermissionPolicy) error
	DeleteResources(ctx context.Context, resources []PermissionResource) error
}

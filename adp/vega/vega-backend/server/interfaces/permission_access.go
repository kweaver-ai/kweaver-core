// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
)

const (
	ACCESSOR_TYPE_USER = "user"
	ACCESSOR_TYPE_APP  = "app"

	RESOURCE_ID_ALL = "*"

	RESOURCE_TYPE_CATALOG        = "catalog"
	RESOURCE_TYPE_CONNECTOR_TYPE = "connector_type"
	RESOURCE_TYPE_RESOURCE       = "resource"

	OPERATION_TYPE_VIEW_DETAIL = "view_detail"
	OPERATION_TYPE_CREATE      = "create"
	OPERATION_TYPE_MODIFY      = "modify"
	OPERATION_TYPE_DELETE      = "delete"
	OPERATION_TYPE_AUTHORIZE   = "authorize"
	OPERATION_TYPE_TASK_MANAGE = "task_manage"

	AUTHORIZATION_RESOURCE_NAME_MODIFY = "authorization.resource.name.modify"
)

var (
	COMMON_OPERATIONS = []string{
		OPERATION_TYPE_VIEW_DETAIL,
		OPERATION_TYPE_CREATE,
		OPERATION_TYPE_MODIFY,
		OPERATION_TYPE_DELETE,
		OPERATION_TYPE_AUTHORIZE,
		OPERATION_TYPE_TASK_MANAGE,
	}
)

type PermissionCheck struct {
	Accessor   PermissionAccessor `json:"accessor"`
	Resource   PermissionResource `json:"resource"`
	Operations []string           `json:"operation"`
	Method     string             `json:"method"`
}

type PermissionCheckResult struct {
	Result bool `json:"result"`
}

type PermissionAccessor struct {
	Type string `json:"type,omitempty"`
	ID   string `json:"id,omitempty"`
}

type PermissionResource struct {
	Type string `json:"type,omitempty"`
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type PermissionResourcesFilter struct {
	Accessor       PermissionAccessor   `json:"accessor,omitempty"`
	Resources      []PermissionResource `json:"resources,omitempty"`
	Operations     []string             `json:"operation,omitempty"`
	AllowOperation bool                 `json:"allow_operation"`
	Method         string               `json:"method,omitempty"`
}

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
	Operations []string `json:"allow_operation,omitempty"`
}

//go:generate mockgen -source ../interfaces/permission_access.go -destination ../interfaces/mock/mock_permission_access.go
type PermissionAccess interface {
	CheckPermission(ctx context.Context, check PermissionCheck) (bool, error)
	CreateResources(ctx context.Context, policies []PermissionPolicy) error
	DeleteResources(ctx context.Context, resources []PermissionResource) error
	FilterResources(ctx context.Context, filter PermissionResourcesFilter) ([]PermissionResourceOps, error)
}

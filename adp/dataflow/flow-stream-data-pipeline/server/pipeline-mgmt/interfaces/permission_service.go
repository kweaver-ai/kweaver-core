// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import "context"

type AccountInfo struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Name string `json:"name"`
}

//go:generate mockgen -source ../interfaces/permission_service.go -destination ../interfaces/mock/mock_permission_service.go
type PermissionService interface {
	CheckPermission(ctx context.Context, resource PermissionResource, ops []string) error
	FilterResources(ctx context.Context, resourceType string, ids []string,
		ops []string, allowOperation bool, fullOps []string) (map[string]PermissionResourceOps, error)
	GetResourcesOperations(ctx context.Context, resourceType string, ids []string, fullOps []string) (map[string]PermissionResourceOps, error)

	CreateResources(ctx context.Context, resources []PermissionResource, ops []string) error
	DeleteResources(ctx context.Context, resourceType string, ids []string) error
	UpdateResource(ctx context.Context, resource PermissionResource) error
}

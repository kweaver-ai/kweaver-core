// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import "context"

//go:generate mockgen -source ../interfaces/permission_service.go -destination ../interfaces/mock/mock_permission_service.go
type PermissionService interface {
	CheckPermission(ctx context.Context, resource Resource, ops []string) error
	CreateResources(ctx context.Context, resources []Resource, ops []string) error
	DeleteResources(ctx context.Context, resourceType string, ids []string) error
	FilterResources(ctx context.Context, resourceType string, ids []string,
		ops []string, allowOperation bool, fullOps []string) (map[string]ResourceOps, error)
	UpdateResource(ctx context.Context, resource Resource) error
}

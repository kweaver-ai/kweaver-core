// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package permission

import (
	"context"

	"flow-stream-data-pipeline/common"
	"flow-stream-data-pipeline/pipeline-mgmt/interfaces"
)

// NoopPermissionService 空权限服务（认证禁用时跳过所有权限检查）
type NoopPermissionService struct {
	appSetting *common.AppSetting
}

func NewNoopPermissionService(appSetting *common.AppSetting) interfaces.PermissionService {
	return &NoopPermissionService{appSetting: appSetting}
}

func (n *NoopPermissionService) CheckPermission(ctx context.Context, resource interfaces.PermissionResource, ops []string) error {
	return nil // 始终通过，不检查权限
}

func (n *NoopPermissionService) CreateResources(ctx context.Context, resources []interfaces.PermissionResource, ops []string) error {
	return nil // 静默跳过
}

func (n *NoopPermissionService) DeleteResources(ctx context.Context, resourceType string, ids []string) error {
	return nil // 静默跳过
}

func (n *NoopPermissionService) FilterResources(ctx context.Context, resourceType string, ids []string,
	ops []string, allowOperation bool, fullOps []string) (map[string]interfaces.PermissionResourceOps, error) {
	// 返回所有资源，不做过滤
	result := make(map[string]interfaces.PermissionResourceOps)
	for _, id := range ids {
		result[id] = interfaces.PermissionResourceOps{
			ResourceID: id,
			Operations: fullOps,
		}
	}
	return result, nil
}

func (n *NoopPermissionService) UpdateResource(ctx context.Context, resource interfaces.PermissionResource) error {
	return nil // 静默跳过
}

func (n *NoopPermissionService) GetResourcesOperations(ctx context.Context, resourceType string,
	ids []string, fullOps []string) (map[string]interfaces.PermissionResourceOps, error) {
	// 返回空操作
	result := make(map[string]interfaces.PermissionResourceOps)
	for _, id := range ids {
		result[id] = interfaces.PermissionResourceOps{
			ResourceID: id,
			Operations: fullOps,
		}
	}
	return result, nil
}

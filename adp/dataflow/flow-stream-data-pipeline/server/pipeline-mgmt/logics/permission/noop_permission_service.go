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

func (n *NoopPermissionService) CheckPermission(ctx context.Context, resource interfaces.Resource, ops []string) error {
	return nil // 始终通过，不检查权限
}

func (n *NoopPermissionService) CreateResources(ctx context.Context, resources []interfaces.Resource, ops []string) error {
	return nil // 静默跳过
}

func (n *NoopPermissionService) DeleteResources(ctx context.Context, resourceType string, ids []string) error {
	return nil // 静默跳过
}

func (n *NoopPermissionService) FilterResources(ctx context.Context, resourceType string, ids []string,
	ops []string, allowOperation bool, fullOps []string) (map[string]interfaces.ResourceOps, error) {
	// 返回所有资源，不做过滤
	result := make(map[string]interfaces.ResourceOps)
	for _, id := range ids {
		result[id] = interfaces.ResourceOps{
			ResourceID: id,
			Operations: fullOps,
		}
	}
	return result, nil
}

func (n *NoopPermissionService) UpdateResource(ctx context.Context, resource interfaces.Resource) error {
	return nil // 静默跳过
}

func (n *NoopPermissionService) GetResourceOps(ctx context.Context, resourceType string, ids []string) (map[string]interfaces.ResourceOps, error) {
	// 返回空操作
	result := make(map[string]interfaces.ResourceOps)
	for _, id := range ids {
		result[id] = interfaces.ResourceOps{
			ResourceID: id,
			Operations: []string{},
		}
	}
	return result, nil
}

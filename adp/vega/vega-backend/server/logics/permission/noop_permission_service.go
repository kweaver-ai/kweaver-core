package permission

import (
	"context"

	"vega-backend/common"
	"vega-backend/interfaces"
)

type NoopPermissionService struct {
	appSetting *common.AppSetting
}

func NewNoopPermissionService(appSetting *common.AppSetting) interfaces.PermissionService {
	return &NoopPermissionService{appSetting: appSetting}
}

func (n *NoopPermissionService) CheckPermission(ctx context.Context, resource interfaces.PermissionResource, ops []string) error {
	return nil
}

func (n *NoopPermissionService) CreateResources(ctx context.Context, resources []interfaces.PermissionResource, ops []string) error {
	return nil
}

func (n *NoopPermissionService) DeleteResources(ctx context.Context, resourceType string, ids []string) error {
	return nil
}

func (n *NoopPermissionService) FilterResources(ctx context.Context, resourceType string, ids []string,
	ops []string, allowOperation bool) (map[string]interfaces.PermissionResourceOps, error) {
	result := make(map[string]interfaces.PermissionResourceOps)
	for _, id := range ids {
		result[id] = interfaces.PermissionResourceOps{
			ResourceID: id,
			Operations: ops,
		}
	}
	return result, nil
}

func (n *NoopPermissionService) UpdateResource(ctx context.Context, resource interfaces.PermissionResource) error {
	return nil
}

// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package resource provides Resource management business logic.
package resource

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/rs/xid"
	"go.opentelemetry.io/otel/codes"

	"vega-backend/common"
	taskAccess "vega-backend/drivenadapters/build_task"
	resourceAccess "vega-backend/drivenadapters/resource"
	verrors "vega-backend/errors"
	"vega-backend/interfaces"
	"vega-backend/logics/catalog"
	dataset "vega-backend/logics/dataset"
	"vega-backend/logics/permission"
	"vega-backend/logics/user_mgmt"
)

var (
	rServiceOnce sync.Once
	rService     interfaces.ResourceService
)

type resourceService struct {
	appSetting *common.AppSetting
	cs         interfaces.CatalogService
	ds         interfaces.DatasetService
	ps         interfaces.PermissionService
	ra         interfaces.ResourceAccess
	ums        interfaces.UserMgmtService
	bta        interfaces.BuildTaskAccess
}

// NewResourceService creates a new ResourceService.
func NewResourceService(appSetting *common.AppSetting) interfaces.ResourceService {
	rServiceOnce.Do(func() {
		rService = &resourceService{
			appSetting: appSetting,
			cs:         catalog.NewCatalogService(appSetting),
			ds:         dataset.NewDatasetService(appSetting),
			ps:         permission.NewPermissionService(appSetting),
			ra:         resourceAccess.NewResourceAccess(appSetting),
			ums:        user_mgmt.NewUserMgmtService(appSetting),
			bta:        taskAccess.NewBuildTaskAccess(appSetting),
		}
	})
	return rService
}

// Create creates a new Resource.
func (rs *resourceService) Create(ctx context.Context, req *interfaces.ResourceRequest) (*interfaces.Resource, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Create resource")
	defer span.End()

	// 判断userid是否有创建数据资源的权限（策略决策）
	err := rs.ps.CheckPermission(ctx, interfaces.PermissionResource{
		Type: interfaces.RESOURCE_TYPE_RESOURCE,
		ID:   interfaces.RESOURCE_ID_ALL,
	}, []string{interfaces.OPERATION_TYPE_CREATE})
	if err != nil {
		return nil, err
	}

	// Get account info from context
	accountInfo := interfaces.AccountInfo{}
	if v := ctx.Value(interfaces.ACCOUNT_INFO_KEY); v != nil {
		accountInfo = v.(interfaces.AccountInfo)
	}

	// 检查catalog是否存在
	exists, err := rs.cs.CheckExistByID(ctx, req.CatalogID)
	if err != nil {
		span.SetStatus(codes.Error, "Check catalog exist failed")
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Catalog_InternalError_GetFailed).
			WithErrorDetails(err.Error())
	}
	if !exists {
		span.SetStatus(codes.Error, "Catalog not found")
		return nil, rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Catalog_NotFound)
	}

	now := time.Now().UnixMilli()
	id := req.ID
	if id == "" {
		id = xid.New().String()
	}

	var logicType string
	switch req.Category {
	case interfaces.ResourceCategoryLogicView:
		logicType, err = rs.validateLogicDefinition(ctx, req)
		if err != nil {
			return nil, err
		}
		viewFields, err := rs.parseLogicDefinition(ctx, req.LogicDefinition)
		if err != nil {
			return nil, err
		}
		req.SchemaDefinition = viewFields
		req.SourceIdentifier = fmt.Sprintf("%s.%s", req.CatalogID, id)
	}

	resource := &interfaces.Resource{
		ID:               id,
		CatalogID:        req.CatalogID,
		Name:             req.Name,
		Tags:             req.Tags,
		Description:      req.Description,
		Category:         req.Category,
		Status:           req.Status,
		Database:         req.Database,
		SourceIdentifier: req.SourceIdentifier,
		SourceMetadata:   req.SourceMetadata,
		SchemaDefinition: req.SchemaDefinition,
		LogicType:        logicType,
		LogicDefinition:  req.LogicDefinition,
		Creator:          accountInfo,
		CreateTime:       now,
		Updater:          accountInfo,
		UpdateTime:       now,
	}

	err = rs.ra.Create(ctx, resource)
	if err != nil {
		logger.Errorf("Create resource failed: %v", err)
		o11y.Error(ctx, fmt.Sprintf("Create resource failed: %v", err))
		span.SetStatus(codes.Error, "Create resource failed")
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError_CreateFailed).
			WithErrorDetails(err.Error())
	}

	switch resource.Category {
	case interfaces.ResourceCategoryDataset:
		// create dataset
		if err := rs.ds.Create(ctx, resource); err != nil {
			logger.Errorf("Create dataset failed: %v", err)
			// 数据集创建失败不影响资源创建，只记录错误
		}
	}

	// 注册资源
	err = rs.ps.CreateResources(ctx, []interfaces.PermissionResource{{
		ID:   resource.ID,
		Type: interfaces.RESOURCE_TYPE_RESOURCE,
		Name: resource.Name,
	}}, interfaces.COMMON_OPERATIONS)
	if err != nil {
		logger.Errorf("CreateResources error: %s", err.Error())
		span.SetStatus(codes.Error, "创建资源失败")
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			verrors.VegaBackend_Catalog_InternalError_CreateResourcesFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return resource, nil
}

// Get retrieves a Resource by ID.
func (rs *resourceService) GetByID(ctx context.Context, id string) (*interfaces.Resource, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Get resource")
	defer span.End()

	resource, err := rs.ra.GetByID(ctx, id)
	if err != nil {
		span.SetStatus(codes.Error, "Get resource failed")
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError_GetFailed).
			WithErrorDetails(err.Error())
	}
	if resource == nil {
		span.SetStatus(codes.Error, "Resource not found")
		return nil, rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Resource_NotFound)
	}

	// 根据权限过滤有查看权限的对象，过滤后的数组的总长度就是总数，无需再请求总数
	matchResoucesMap, err := rs.ps.FilterResources(ctx, interfaces.RESOURCE_TYPE_RESOURCE, []string{resource.ID},
		[]string{interfaces.OPERATION_TYPE_VIEW_DETAIL}, true, interfaces.COMMON_OPERATIONS)
	if err != nil {
		span.SetStatus(codes.Error, "Filter resources error")
		return nil, err
	}

	if resrc, exist := matchResoucesMap[resource.ID]; exist {
		resource.Operations = resrc.Operations // 用户当前有权限的操作
	} else {
		return nil, rest.NewHTTPError(ctx, http.StatusForbidden, rest.PublicError_Forbidden).
			WithErrorDetails(fmt.Sprintf("Access denied: insufficient permissions for[%v]", interfaces.OPERATION_TYPE_VIEW_DETAIL))
	}

	accountInfos := []*interfaces.AccountInfo{&resource.Creator, &resource.Updater}
	err = rs.ums.GetAccountNames(ctx, accountInfos)
	if err != nil {
		span.SetStatus(codes.Error, "GetAccountNames error")

		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			verrors.VegaBackend_Resource_InternalError_GetAccountNamesFailed).WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return resource, nil
}

// GetByIDs retrieves Resources by IDs.
func (rs *resourceService) GetByIDs(ctx context.Context, ids []string) ([]*interfaces.Resource, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Get resources by IDs")
	defer span.End()

	if len(ids) == 0 {
		span.SetStatus(codes.Ok, "")
		return []*interfaces.Resource{}, nil
	}

	resources, err := rs.ra.GetByIDs(ctx, ids)
	if err != nil {
		span.SetStatus(codes.Error, "Get resources failed")
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError_GetFailed).
			WithErrorDetails(err.Error())
	}

	// 根据权限过滤有查看权限的对象，过滤后的数组的总长度就是总数，无需再请求总数
	matchResoucesMap, err := rs.ps.FilterResources(ctx, interfaces.RESOURCE_TYPE_RESOURCE, ids,
		[]string{interfaces.OPERATION_TYPE_VIEW_DETAIL}, true, interfaces.COMMON_OPERATIONS)
	if err != nil {
		span.SetStatus(codes.Error, "Filter resources error")
		return nil, err
	}

	accountInfos := make([]*interfaces.AccountInfo, 0)
	for _, resource := range resources {
		if resrc, exist := matchResoucesMap[resource.ID]; exist {
			resource.Operations = resrc.Operations // 用户当前有权限的操作
		} else {
			return nil, rest.NewHTTPError(ctx, http.StatusForbidden, rest.PublicError_Forbidden).
				WithErrorDetails(fmt.Sprintf("Access denied: insufficient permissions for[%v]", interfaces.OPERATION_TYPE_VIEW_DETAIL))
		}
		accountInfos = append(accountInfos, &resource.Creator, &resource.Updater)
	}

	err = rs.ums.GetAccountNames(ctx, accountInfos)
	if err != nil {
		span.SetStatus(codes.Error, "GetAccountNames error")

		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			verrors.VegaBackend_Resource_InternalError_GetAccountNamesFailed).WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return resources, nil
}

// GetByCatalogID retrieves all Resources under a Catalog.
func (rs *resourceService) GetByCatalogID(ctx context.Context, catalogID string) ([]*interfaces.Resource, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Get resources by catalog ID")
	defer span.End()

	resources, err := rs.ra.GetByCatalogID(ctx, catalogID)
	if err != nil {
		span.SetStatus(codes.Error, "Get resources failed")
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError_GetFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return resources, nil
}

// GetByName retrieves a Resource by catalog and name.
func (rs *resourceService) GetByName(ctx context.Context, catalogID string, name string) (*interfaces.Resource, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Get resource by name")
	defer span.End()

	resource, err := rs.ra.GetByName(ctx, catalogID, name)
	if err != nil {
		span.SetStatus(codes.Error, "Get resource failed")
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError_GetFailed).
			WithErrorDetails(err.Error())
	}
	if resource == nil {
		span.SetStatus(codes.Error, "Resource not found")
		return nil, rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Resource_NotFound)
	}

	span.SetStatus(codes.Ok, "")
	return resource, nil
}

// List lists Resources with filters.
func (rs *resourceService) List(ctx context.Context, params interfaces.ResourcesQueryParams) ([]*interfaces.Resource, int64, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "List resources")
	defer span.End()

	// 查询所有资源的ID
	ids, err := rs.ra.ListIDs(ctx, params)
	if err != nil {
		span.SetStatus(codes.Error, "List resource IDs failed")
		return []*interfaces.Resource{}, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError_GetFailed).
			WithErrorDetails(err.Error())
	}

	if len(ids) == 0 {
		span.SetStatus(codes.Ok, "")
		return []*interfaces.Resource{}, 0, nil
	}

	// 根据权限过滤有查看权限的ID数组
	// 分批处理，每批1万个ids, fix权限接口报错prepared statement contains too many placeholders
	batchSize := 10000
	// 所有有权限的resource及其操作权限
	matchResourceOpsMap := make(map[string]interfaces.PermissionResourceOps)

	for i := 0; i < len(ids); i += batchSize {
		end := i + batchSize
		if end > len(ids) {
			end = len(ids)
		}
		batchIDs := ids[i:end]

		var batchMatchResources map[string]interfaces.PermissionResourceOps
		// 校验权限管理的操作权限
		batchMatchResources, err = rs.ps.FilterResources(ctx, interfaces.RESOURCE_TYPE_RESOURCE,
			batchIDs, []string{interfaces.OPERATION_TYPE_VIEW_DETAIL}, true, interfaces.COMMON_OPERATIONS)
		if err != nil {
			span.SetStatus(codes.Error, "Filter resources error")
			return []*interfaces.Resource{}, 0, err
		}

		// 合并结果
		for _, resourceOps := range batchMatchResources {
			matchResourceOpsMap[resourceOps.ResourceID] = resourceOps
		}
	}

	// 提取有权限的资源ID，保持与ids的顺序一致
	authorizedIDs := make([]string, 0, len(matchResourceOpsMap))
	for _, id := range ids {
		if _, exist := matchResourceOpsMap[id]; exist {
			authorizedIDs = append(authorizedIDs, id)
		}
	}
	total := int64(len(authorizedIDs))

	// 如果没有有权限的资源，直接返回空结果
	if total == 0 {
		span.SetStatus(codes.Ok, "")
		return []*interfaces.Resource{}, total, nil
	}

	// 根据有权限的ID数组查询完整资源，并应用分页
	// limit = -1,则返回所有
	if params.Limit != -1 {
		// 分页处理authorizedIDs
		// 检查起始位置是否越界
		if params.Offset < 0 || params.Offset >= len(authorizedIDs) {
			span.SetStatus(codes.Ok, "")
			return []*interfaces.Resource{}, total, nil
		}
		// 计算结束位置
		end := params.Offset + params.Limit
		if end > len(authorizedIDs) {
			end = len(authorizedIDs)
		}
		// 只查询当前页的资源ID
		authorizedIDs = authorizedIDs[params.Offset:end]
	}

	// 根据有权限的ID数组查询完整资源
	// 分批处理，每批10000个ids, 避免prepared statement contains too many placeholders错误
	resources := make([]*interfaces.Resource, 0, len(authorizedIDs))
	queryBatchSize := 10000
	for i := 0; i < len(authorizedIDs); i += queryBatchSize {
		end := i + queryBatchSize
		if end > len(authorizedIDs) {
			end = len(authorizedIDs)
		}
		batchIDs := authorizedIDs[i:end]

		batchResources, err := rs.ra.GetByIDsBasic(ctx, batchIDs)
		if err != nil {
			span.SetStatus(codes.Error, "Get resources by IDs failed")
			return []*interfaces.Resource{}, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError_GetFailed).
				WithErrorDetails(err.Error())
		}

		resources = append(resources, batchResources...)
	}

	// 设置资源操作权限
	for _, c := range resources {
		if resrc, exist := matchResourceOpsMap[c.ID]; exist {
			c.Operations = resrc.Operations // 用户当前有权限的操作
		}
	}

	accountInfos := make([]*interfaces.AccountInfo, 0, len(resources)*2)
	for _, c := range resources {
		accountInfos = append(accountInfos, &c.Creator, &c.Updater)
	}

	err = rs.ums.GetAccountNames(ctx, accountInfos)
	if err != nil {
		span.SetStatus(codes.Error, "GetAccountNames error")

		return []*interfaces.Resource{}, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError_GetFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return resources, total, nil
}

// Update updates a Resource.
func (rs *resourceService) Update(ctx context.Context, id string, req *interfaces.ResourceRequest) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Update resource")
	defer span.End()

	resource := req.OriginResource
	if resource == nil {
		span.SetStatus(codes.Error, "Resource not found")
		return rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Resource_NotFound)
	}

	// 判断userid是否有修改权限
	err := rs.ps.CheckPermission(ctx, interfaces.PermissionResource{
		Type: interfaces.RESOURCE_TYPE_RESOURCE,
		ID:   resource.ID,
	}, []string{interfaces.OPERATION_TYPE_MODIFY})
	if err != nil {
		return err
	}

	switch req.OriginResource.Category {
	case interfaces.ResourceCategoryLogicView:
		logicType, err := rs.validateLogicDefinition(ctx, req)
		if err != nil {
			return err
		}
		viewFields, err := rs.parseLogicDefinition(ctx, req.LogicDefinition)
		if err != nil {
			return err
		}
		resource.SchemaDefinition = viewFields
		resource.LogicType = logicType
		resource.LogicDefinition = req.LogicDefinition
	}

	// 检查catalog是否存在
	exists, err := rs.cs.CheckExistByID(ctx, req.CatalogID)
	if err != nil {
		return err
	}
	if !exists {
		return rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Catalog_NotFound)
	}

	// Apply updates
	resource.Name = req.Name
	resource.Tags = req.Tags
	resource.Description = req.Description

	// Get account info
	accountInfo := interfaces.AccountInfo{}
	if v := ctx.Value(interfaces.ACCOUNT_INFO_KEY); v != nil {
		accountInfo = v.(interfaces.AccountInfo)
	}

	now := time.Now().UnixMilli()
	resource.Updater = accountInfo
	resource.UpdateTime = now

	if err := rs.ra.Update(ctx, resource); err != nil {
		span.SetStatus(codes.Error, "Update resource failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError_UpdateFailed).
			WithErrorDetails(err.Error())
	}

	// 请求更新资源名称的接口，更新资源的名称
	if req.IfNameModify {
		err = rs.ps.UpdateResource(ctx, interfaces.PermissionResource{
			ID:   resource.ID,
			Type: interfaces.RESOURCE_TYPE_RESOURCE,
			Name: resource.Name,
		})
		if err != nil {
			return err
		}
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// UpdateStatus updates a Resource's status.
func (rs *resourceService) UpdateStatus(ctx context.Context, id string, status string, statusMessage string) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Update resource status")
	defer span.End()

	if err := rs.ra.UpdateStatus(ctx, id, status, statusMessage); err != nil {
		span.SetStatus(codes.Error, "Update resource status failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError_UpdateFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// DeleteByIDs deletes Resources by IDs.
func (rs *resourceService) DeleteByIDs(ctx context.Context, ids []string) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Delete resources")
	defer span.End()

	if len(ids) == 0 {
		span.SetStatus(codes.Ok, "")
		return nil
	}

	// 判断userid是否有删除权限
	matchResoucesMap, err := rs.ps.FilterResources(ctx, interfaces.RESOURCE_TYPE_RESOURCE, ids,
		[]string{interfaces.OPERATION_TYPE_DELETE}, true, interfaces.COMMON_OPERATIONS)
	if err != nil {
		span.SetStatus(codes.Error, "Filter resources error")
		return err
	}

	// 检查是否有删除权限
	if len(matchResoucesMap) != len(ids) {
		// 请求的资源id可以重复，未去重，资源过滤出来的资源id是去重过的，所以单纯判断数量不准确
		for _, id := range ids {
			if _, exist := matchResoucesMap[id]; !exist {
				return rest.NewHTTPError(ctx, http.StatusForbidden, rest.PublicError_Forbidden).
					WithErrorDetails("Access denied: insufficient permissions for resource's delete operation.")
			}
		}
	}

	// 先获取要删除的资源信息，以便对不同的资源进行不同的处理
	resources, err := rs.ra.GetByIDs(ctx, ids)
	if err != nil {
		span.SetStatus(codes.Error, "Get resources failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError_GetFailed).
			WithErrorDetails(err.Error())
	}

	for _, resource := range resources {
		switch resource.Category {
		case interfaces.ResourceCategoryTable:
			// Check if dataset has build tasks
			buildTask, err := rs.bta.GetByResourceID(ctx, resource.ID)
			if err != nil {
				span.SetStatus(codes.Error, "Get build task failed")
				return rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError_GetFailed).
					WithErrorDetails(err.Error())
			} else if buildTask != nil {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_BuildTask_Exist).
					WithErrorDetails("Cannot delete dataset, please delete build task first")
			}
			if resource.LocalIndexName != "" {
				// 删除本地索引
				err := rs.ds.Delete(ctx, resource.LocalIndexName)
				if err != nil {
					logger.Errorf("Delete local index failed: %v", err)
					// 索引删除失败不影响资源删除，只记录错误
				}
			}
		case interfaces.ResourceCategoryDataset:
			// Delete dataset
			if err := rs.ds.Delete(ctx, resource.ID); err != nil {
				logger.Errorf("Delete dataset failed: %v", err)
				// 数据集删除失败不影响资源删除，只记录错误
			}
		}
	}

	if err := rs.ra.DeleteByIDs(ctx, ids); err != nil {
		span.SetStatus(codes.Error, "Delete resources failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError_DeleteFailed).
			WithErrorDetails(err.Error())
	}

	//  清除资源策略
	err = rs.ps.DeleteResources(ctx, interfaces.RESOURCE_TYPE_RESOURCE, ids)
	if err != nil {
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// CheckExistByID checks if a resource exists by ID.
func (rs *resourceService) CheckExistByID(ctx context.Context, id string) (bool, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Check resource exist by ID")
	defer span.End()

	resource, err := rs.ra.GetByID(ctx, id)
	if err != nil {
		span.SetStatus(codes.Error, "GetByID failed")
		return false, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError_GetFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return resource != nil, nil
}

// CheckExistByName checks if a Resource exists by name.
func (rs *resourceService) CheckExistByName(ctx context.Context, catalogID string, name string) (bool, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Check resource exist by name")
	defer span.End()

	resource, err := rs.ra.GetByName(ctx, catalogID, name)
	if err != nil {
		span.SetStatus(codes.Error, "GetByName failed")
		return false, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError_GetFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return resource != nil, nil
}

// UpdateResource updates a Resource directly.
func (rs *resourceService) UpdateResource(ctx context.Context, resource *interfaces.Resource) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Update resource")
	defer span.End()

	if err := rs.ra.Update(ctx, resource); err != nil {
		span.SetStatus(codes.Error, "Update resource failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError_UpdateFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// ListResourceSrcs lists Resource Sources with filters.
func (rs *resourceService) ListResourceSrcs(ctx context.Context, params interfaces.ListResourcesQueryParams) ([]*interfaces.ListResourceEntry, int64, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "ListResourceSrcs")
	defer span.End()

	// 先查询所有资源源的ID
	ids, err := rs.ra.ListResourceSrcsIDs(ctx, params)
	if err != nil {
		span.SetStatus(codes.Error, "ListResourceSrcsIDs failed")
		return []*interfaces.ListResourceEntry{}, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError_GetFailed).
			WithErrorDetails(err.Error())
	}

	if len(ids) == 0 {
		return []*interfaces.ListResourceEntry{}, 0, nil
	}

	// 使用分批处理的方式过滤权限，每批处理1万个ID
	batchSize := 10000
	// 所有有权限的resource及其操作权限
	matchResourceOpsMap := make(map[string]interfaces.PermissionResourceOps)

	for i := 0; i < len(ids); i += batchSize {
		end := i + batchSize
		if end > len(ids) {
			end = len(ids)
		}
		batchIDs := ids[i:end]

		var batchMatchResources map[string]interfaces.PermissionResourceOps
		// 校验权限管理的操作权限
		batchMatchResources, err = rs.ps.FilterResources(ctx, interfaces.RESOURCE_TYPE_RESOURCE,
			batchIDs, []string{interfaces.OPERATION_TYPE_VIEW_DETAIL}, false, interfaces.COMMON_OPERATIONS)
		if err != nil {
			return []*interfaces.ListResourceEntry{}, 0, err
		}

		// 合并结果
		for _, resourceOps := range batchMatchResources {
			matchResourceOpsMap[resourceOps.ResourceID] = resourceOps
		}
	}

	// 提取有权限的资源ID，保持与ids的顺序一致
	authorizedIDs := make([]string, 0, len(matchResourceOpsMap))
	for _, id := range ids {
		if _, exist := matchResourceOpsMap[id]; exist {
			authorizedIDs = append(authorizedIDs, id)
		}
	}
	total := int64(len(authorizedIDs))

	// 如果没有有权限的资源，直接返回空结果
	if total == 0 {
		span.SetStatus(codes.Ok, "")
		return []*interfaces.ListResourceEntry{}, total, nil
	}

	// 根据有权限的ID数组应用分页
	if params.Limit != -1 {
		// 分页处理authorizedIDs
		// 检查起始位置是否越界
		if params.Offset < 0 || params.Offset >= len(authorizedIDs) {
			span.SetStatus(codes.Ok, "")
			return []*interfaces.ListResourceEntry{}, total, nil
		}
		// 计算结束位置
		end := params.Offset + params.Limit
		if end > len(authorizedIDs) {
			end = len(authorizedIDs)
		}
		// 只查询当前页的资源ID
		authorizedIDs = authorizedIDs[params.Offset:end]
	}

	// 根据有权限的ID数组查询完整资源
	// 分批处理，每批10000个ids, 避免prepared statement contains too many placeholders错误
	entries := make([]*interfaces.ListResourceEntry, 0, len(authorizedIDs))
	queryBatchSize := 10000
	for i := 0; i < len(authorizedIDs); i += queryBatchSize {
		end := i + queryBatchSize
		if end > len(authorizedIDs) {
			end = len(authorizedIDs)
		}
		batchIDs := authorizedIDs[i:end]

		batchEntries, err := rs.ra.ListResourceSrcsByIDs(ctx, batchIDs)
		if err != nil {
			span.SetStatus(codes.Error, "ListResourceSrcsByIDs failed")
			return []*interfaces.ListResourceEntry{}, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError_GetFailed).
				WithErrorDetails(err.Error())
		}

		entries = append(entries, batchEntries...)
	}

	// 根据有权限的ID过滤结果
	results := make([]*interfaces.ListResourceEntry, 0)
	for _, entry := range entries {
		if _, exist := matchResourceOpsMap[entry.ID]; exist {
			results = append(results, entry)
		}
	}

	span.SetStatus(codes.Ok, "")
	return results, total, nil
}

// CheckExistByCategories checks if Resources exists by catalog ID and categories.
func (rs *resourceService) CheckExistByCategories(ctx context.Context, catalogID string, categories []string) (bool, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "CheckExistByCategories")
	defer span.End()

	return rs.ra.CheckExistByCategories(ctx, catalogID, categories)
}


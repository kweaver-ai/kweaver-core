// Package auth provides authorization service.
package auth

import (
	"context"
	"net/http"

	infraerrors "github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
)

const (
	// 权限查询相关常量
	MaxInQuerySize   = 1000 // IN查询最大参数数量，避免数据库限制
	InQueryBatchSize = 200  // 分批IN查询的批次大小
)

// CheckCreatePermission 检查新增权限
func (s *authServiceImpl) CheckCreatePermission(ctx context.Context, accessor *interfaces.AuthAccessor, resourceType interfaces.AuthResourceType) error {
	authorized, err := s.OperationCheckAll(ctx, accessor, interfaces.ResourceIDAll, resourceType, interfaces.AuthOperationTypeCreate)
	if err != nil {
		return err
	}
	if !authorized {
		return infraerrors.NewHTTPError(ctx, http.StatusForbidden, infraerrors.ErrExtCommonAddForbidden, nil)
	}
	return nil
}

// CheckModifyPermission 检查编辑权限
func (s *authServiceImpl) CheckModifyPermission(ctx context.Context, accessor *interfaces.AuthAccessor, resourceID string, resourceType interfaces.AuthResourceType) error {
	authorized, err := s.OperationCheckAll(ctx, accessor, resourceID, resourceType, interfaces.AuthOperationTypeModify)
	if err != nil {
		return err
	}
	if !authorized {
		return infraerrors.NewHTTPError(ctx, http.StatusForbidden, infraerrors.ErrExtCommonEditForbidden, nil)
	}
	return nil
}

// CheckViewPermission 检查查看权限
func (s *authServiceImpl) CheckViewPermission(ctx context.Context, accessor *interfaces.AuthAccessor, resourceID string, resourceType interfaces.AuthResourceType) error {
	authorized, err := s.OperationCheckAll(ctx, accessor, resourceID, resourceType, interfaces.AuthOperationTypeView)
	if err != nil {
		return err
	}
	if !authorized {
		return infraerrors.NewHTTPError(ctx, http.StatusForbidden, infraerrors.ErrExtCommonViewForbidden, nil)
	}
	return nil
}

// CheckDeletePermission 检查删除权限
func (s *authServiceImpl) CheckDeletePermission(ctx context.Context, accessor *interfaces.AuthAccessor, resourceID string, resourceType interfaces.AuthResourceType) error {
	authorized, err := s.OperationCheckAll(ctx, accessor, resourceID, resourceType, interfaces.AuthOperationTypeDelete)
	if err != nil {
		return err
	}
	if !authorized {
		return infraerrors.NewHTTPError(ctx, http.StatusForbidden, infraerrors.ErrExtCommonDeleteForbidden, nil)
	}
	return nil
}

// CheckPublishPermission 检查发布权限
func (s *authServiceImpl) CheckPublishPermission(ctx context.Context, accessor *interfaces.AuthAccessor, resourceID string, resourceType interfaces.AuthResourceType) error {
	authorized, err := s.OperationCheckAll(ctx, accessor, resourceID, resourceType, interfaces.AuthOperationTypePublish)
	if err != nil {
		return err
	}
	if !authorized {
		return infraerrors.NewHTTPError(ctx, http.StatusForbidden, infraerrors.ErrExtCommonPublishForbidden, nil)
	}
	return nil
}

// CheckUnpublishPermission 检查下架权限
func (s *authServiceImpl) CheckUnpublishPermission(ctx context.Context, accessor *interfaces.AuthAccessor, resourceID string, resourceType interfaces.AuthResourceType) error {
	authorized, err := s.OperationCheckAll(ctx, accessor, resourceID, resourceType, interfaces.AuthOperationTypeUnpublish)
	if err != nil {
		return err
	}
	if !authorized {
		return infraerrors.NewHTTPError(ctx, http.StatusForbidden, infraerrors.ErrExtCommonUnpublishForbidden, nil)
	}
	return nil
}

// CheckAuthorizePermission 检查权限管理权限
func (s *authServiceImpl) CheckAuthorizePermission(ctx context.Context, accessor *interfaces.AuthAccessor, resourceID string, resourceType interfaces.AuthResourceType) error {
	authorized, err := s.OperationCheckAll(ctx, accessor, resourceID, resourceType, interfaces.AuthOperationTypeAuthorize)
	if err != nil {
		return err
	}
	if !authorized {
		return infraerrors.NewHTTPError(ctx, http.StatusForbidden, infraerrors.ErrExtCommonPermissionForbidden, nil)
	}
	return nil
}

// CheckPublicAccessPermission 检查公共访问权限
func (s *authServiceImpl) CheckPublicAccessPermission(ctx context.Context, accessor *interfaces.AuthAccessor, resourceID string, resourceType interfaces.AuthResourceType) error {
	authorized, err := s.OperationCheckAll(ctx, accessor, resourceID, resourceType, interfaces.AuthOperationTypePublicAccess)
	if err != nil {
		return err
	}
	if !authorized {
		return infraerrors.NewHTTPError(ctx, http.StatusForbidden, infraerrors.ErrExtCommonPublicAccessForbidden, nil)
	}
	return nil
}

// CheckExecutePermission 检查使用权限
func (s *authServiceImpl) CheckExecutePermission(ctx context.Context, accessor *interfaces.AuthAccessor, resourceID string, resourceType interfaces.AuthResourceType) error {
	authorized, err := s.OperationCheckAll(ctx, accessor, resourceID, resourceType, interfaces.AuthOperationTypeExecute)
	if err != nil {
		return err
	}
	if !authorized {
		return infraerrors.NewHTTPError(ctx, http.StatusForbidden, infraerrors.ErrExtCommonUseForbidden, nil)
	}
	return nil
}

// MultiCheckOperationPermission 多操作权限检查
func (s *authServiceImpl) MultiCheckOperationPermission(ctx context.Context, accessor *interfaces.AuthAccessor, resourceID string,
	resourceType interfaces.AuthResourceType, operations ...interfaces.AuthOperationType) error {
	authorized, err := s.OperationCheckAll(ctx, accessor, resourceID, resourceType, operations...)
	if err != nil {
		return err
	}
	if !authorized {
		return infraerrors.NewHTTPError(ctx, http.StatusForbidden, infraerrors.ErrExtCommonOperationForbidden, nil)
	}
	return nil
}

// OperationCheckAll 检查操作权限
func (s *authServiceImpl) OperationCheckAll(
	ctx context.Context,
	accessor *interfaces.AuthAccessor,
	resourceID string,
	resourceType interfaces.AuthResourceType,
	operations ...interfaces.AuthOperationType,
) (bool, error) {
	req := &interfaces.AuthOperationCheckRequest{
		Accessor: accessor,
		Resource: &interfaces.AuthResource{
			ID:   resourceID,
			Type: string(resourceType),
		},
		Operation: operations,
		Method:    interfaces.AuthMethodGet,
	}
	resp, err := s.authorization.OperationCheck(ctx, req)
	if err != nil {
		err := infraerrors.NewHTTPError(ctx, http.StatusForbidden, infraerrors.ErrExtCommonOperationForbidden, err.Error())
		return false, err
	}
	return resp.Result, nil
}

// OperationCheckAny 检查操作权限
func (s *authServiceImpl) OperationCheckAny(
	ctx context.Context,
	accessor *interfaces.AuthAccessor,
	resourceID string,
	resourceType interfaces.AuthResourceType,
	operations ...interfaces.AuthOperationType,
) (bool, error) {
	for _, operation := range operations {
		authorized, err := s.OperationCheckAll(ctx, accessor, resourceID, resourceType, operation)
		if err != nil {
			return false, err
		}
		if authorized {
			return true, nil
		}
	}
	return false, nil
}

// ResourceFilterIDs 资源过滤
func (s *authServiceImpl) ResourceFilterIDs(
	ctx context.Context,
	accessor *interfaces.AuthAccessor,
	resourceIDS []string,
	resourceType interfaces.AuthResourceType,
	operations ...interfaces.AuthOperationType,
) ([]string, error) {
	req := &interfaces.AuthResourceFilterRequest{
		Accessor:   accessor,
		Resources:  []*interfaces.AuthResource{},
		Operations: operations,
		Method:     interfaces.AuthMethodGet,
	}

	for _, resourceID := range resourceIDS {
		req.Resources = append(req.Resources, &interfaces.AuthResource{
			ID:   resourceID,
			Type: string(resourceType),
		})
	}

	resp, err := s.authorization.ResourceFilter(ctx, req)
	if err != nil {
		return nil, err
	}
	resourceIDs := make([]string, 0, len(resp))
	for _, resource := range resp {
		resourceIDs = append(resourceIDs, resource.ID)
	}
	return resourceIDs, nil
}

// ResourceListIDs 获取资源列表
func (s *authServiceImpl) ResourceListIDs(
	ctx context.Context,
	accessor *interfaces.AuthAccessor,
	resourceType interfaces.AuthResourceType,
	operations ...interfaces.AuthOperationType,
) ([]string, error) {
	req := &interfaces.ResourceListRequest{
		Accessor: accessor,
		Resource: &interfaces.AuthResource{
			Type: string(resourceType),
		},
		Method:    interfaces.AuthMethodGet,
		Operation: operations,
	}

	resp, err := s.authorization.ResourceList(ctx, req)
	if err != nil {
		return nil, err
	}
	resourceIDs := make([]string, 0, len(resp))
	for _, resource := range resp {
		resourceIDs = append(resourceIDs, resource.ID)
	}
	return resourceIDs, nil
}

// SelectListWithAuth 查询列表并进行权限检查（独立泛型函数） -- 全量过滤
func SelectListWithAuth[T any, PT interfaces.PtrBizIdentifiable[T]](ctx context.Context,
	page int,
	pageSize int,
	all bool,
	queryOption interfaces.QueryOption[T, PT],
	resourceListFunc interfaces.ResourceListFunc,
) (resp *interfaces.QueryResponse[T], err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	// 1. 执行查询获取所有数据
	allData, err := queryOption()
	if err != nil {
		return nil, err
	}

	if allData == nil {
		return &interfaces.QueryResponse[T]{
			Data: []*T{},
			CommonPageResult: interfaces.CommonPageResult{
				TotalCount: 0,
				Page:       page,
				PageSize:   pageSize,
				TotalPage:  0,
				HasNext:    false,
				HasPrev:    false,
			},
		}, nil
	}

	// 2. 获取用户有权限的资源ID列表
	authorizedIDs, err := resourceListFunc()
	if err != nil {
		return nil, err
	}

	// 3. 权限过滤
	var filteredData []PT
	if len(authorizedIDs) > 0 {
		authMap := make(map[string]bool, len(authorizedIDs))
		for _, id := range authorizedIDs {
			authMap[id] = true
		}

		if authMap[interfaces.ResourceIDAll] {
			filteredData = allData
		} else {
			for _, item := range allData {
				if item != nil && authMap[item.GetBizID()] {
					filteredData = append(filteredData, item)
				}
			}
		}
	}

	// 4. 分页处理
	totalCount := len(filteredData)

	var pageData []*T
	if all {
		pageData = make([]*T, totalCount)
		for i, item := range filteredData[:totalCount] {
			pageData[i] = (*T)(item)
		}
		return &interfaces.QueryResponse[T]{
			Data: pageData,
			CommonPageResult: interfaces.CommonPageResult{
				TotalCount: totalCount,
				Page:       1,
				PageSize:   totalCount,
				TotalPage:  1,
				HasNext:    false,
				HasPrev:    false,
			},
		}, nil
	}

	// 设置默认分页参数
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	// 计算分页
	totalPages := (totalCount + pageSize - 1) / pageSize
	hasNext := page < totalPages
	hasPrev := page > 1

	// 计算数据切片范围
	startIndex := (page - 1) * pageSize
	endIndex := startIndex + pageSize

	if startIndex < totalCount {
		if endIndex > totalCount {
			endIndex = totalCount
		}
		sliceData := filteredData[startIndex:endIndex]
		pageData = make([]*T, len(sliceData))
		for i, item := range sliceData {
			pageData[i] = (*T)(item)
		}
	}

	resp = &interfaces.QueryResponse[T]{
		Data: pageData,
		CommonPageResult: interfaces.CommonPageResult{
			TotalCount: totalCount,
			Page:       page,
			PageSize:   pageSize,
			TotalPage:  totalPages,
			HasNext:    hasNext,
			HasPrev:    hasPrev,
		},
	}
	return
}

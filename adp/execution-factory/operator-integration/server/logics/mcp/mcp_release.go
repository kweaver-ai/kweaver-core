package mcp

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common/ormhelper"
	infraerrors "github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/auth"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
)

const (
	defaultPageSize = 1000
	defaultPage     = 1
)

// QueryRelease 查询MCP Server Release列表
func (s *mcpServiceImpl) QueryRelease(ctx context.Context, req *interfaces.MCPServerReleaseListRequest) (result *interfaces.MCPServerReleaseListResponse, err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	filter := make(map[string]interface{})
	filter["all"] = req.All
	if req.Name != "" {
		filter["name"] = req.Name
	}
	if req.Source != "" {
		filter["source"] = req.Source
	}
	if req.Category != "" {
		filter["category"] = req.Category
	}
	if req.CreateUser != "" {
		filter["create_user"] = req.CreateUser
	}
	if req.ReleaseUser != "" {
		filter["release_user"] = req.ReleaseUser
	}
	if req.Mode != "" {
		filter["mode"] = req.Mode
	}
	// 排序字段
	sortField := "f_update_time"
	if req.SortBy != "" {
		sortField = sortFieldMap[req.SortBy]
		if sortField == "" {
			return nil, fmt.Errorf("invalid sort field: %s", req.SortBy)
		}
	}
	sort := &ormhelper.SortParams{
		Fields: []ormhelper.SortField{
			{Field: sortField, Order: ormhelper.SortOrder(req.SortOrder)},
		},
	}
	// 查询MCP Server Release列表
	queryTotal := func(newCtx context.Context) (int64, error) {
		var total int64
		total, err = s.DBMCPServerRelease.CountByWhereClause(newCtx, nil, filter)
		if err != nil {
			return 0, err
		}
		return total, nil
	}
	queryBatch := func(newCtx context.Context, pageSize, offset int, cursorValue *model.MCPServerReleaseDB) ([]*model.MCPServerReleaseDB, error) {
		var releaseList []*model.MCPServerReleaseDB
		var cursor *ormhelper.CursorParams
		if cursorValue != nil {
			cursor = &ormhelper.CursorParams{
				Field:     sortField,
				Direction: ormhelper.SortOrder(req.SortOrder),
			}
			switch sortField {
			case "f_update_time":
				cursor.Value = cursorValue.UpdateTime
			case "f_create_time":
				cursor.Value = cursorValue.CreateTime
			case "f_name":
				cursor.Value = cursorValue.Name
			}
			// 如果使用游标，offset不需要
			offset = 0
		}
		filter["limit"] = pageSize
		filter["offset"] = offset
		releaseList, err = s.DBMCPServerRelease.SelectListPage(newCtx, nil, filter, sort, cursor)
		if err != nil {
			return nil, err
		}
		return releaseList, nil
	}

	businessDomainIds := strings.Split(req.BusinessDomainID, ",")
	resourceToBdMap, err := s.BusinessDomainService.BatchResourceList(ctx, businessDomainIds, interfaces.AuthResourceTypeMCP)
	if err != nil {
		return
	}

	queryBuilder := auth.NewQueryBuilder[model.MCPServerReleaseDB]().
		SetPage(req.Page, req.PageSize).SetAll(req.All).
		SetQueryFunctions(queryTotal, queryBatch).
		SetFilteredQueryFunctions(func(newCtx context.Context, ids []string) (int64, error) {
			filter["in"] = ids
			return queryTotal(newCtx)
		}, func(newCtx context.Context, pageSize, offset int, ids []string, cursorValue *model.MCPServerReleaseDB) ([]*model.MCPServerReleaseDB, error) {
			filter["in"] = ids
			return queryBatch(newCtx, pageSize, offset, cursorValue)
		}).
		SetAuthFilter(func(newCtx context.Context) ([]string, error) {
			var accessor *interfaces.AuthAccessor
			accessor, err = s.AuthService.GetAccessor(newCtx, req.UserID)
			if err != nil {
				return nil, err
			}
			return s.AuthService.ResourceListIDs(newCtx, accessor, interfaces.AuthResourceTypeMCP, interfaces.AuthOperationTypePublicAccess)
		}).
		SetBusinessDomainFilter(func(newCtx context.Context) ([]string, error) {
			resourceIDs := make([]string, 0, len(resourceToBdMap))
			for resourceID := range resourceToBdMap {
				resourceIDs = append(resourceIDs, resourceID)
			}
			return resourceIDs, nil
		})
	resp, err := queryBuilder.Execute(ctx)
	if err != nil {
		return
	}
	configList := resp.Data

	userIDs := []string{}
	data := make([]*interfaces.MCPServerConfigInfo, 0, len(configList))
	for _, config := range configList {
		userIDs = append(userIDs, config.CreateUser, config.UpdateUser, config.ReleaseUser)
		data = append(data, s.releaseModelToResponse(config))
	}

	// 渲染用户名称
	userMap, err := s.UserMgnt.GetUsersName(ctx, userIDs)
	if err != nil {
		return
	}
	for _, config := range data {
		config.BusinessDomainID = utils.GetValueOrDefault(resourceToBdMap, config.MCPID, req.BusinessDomainID)
		config.CreateUser = utils.GetValueOrDefault(userMap, config.CreateUser, interfaces.UnknownUser)
		config.UpdateUser = utils.GetValueOrDefault(userMap, config.UpdateUser, interfaces.UnknownUser)
		config.ReleaseUser = utils.GetValueOrDefault(userMap, config.ReleaseUser, interfaces.UnknownUser)
	}

	queryResult := &ormhelper.QueryResult{
		Total:      int64(resp.TotalCount),
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: resp.TotalPage,
		HasNext:    resp.HasNext,
		HasPrev:    resp.HasPrev,
	}
	result = &interfaces.MCPServerReleaseListResponse{
		MCPServerListResponse: interfaces.MCPServerListResponse{
			QueryResult: queryResult,
			Data:        data,
		},
	}
	return
}

// GetReleaseDetail 获取MCP Server Release详情
func (s *mcpServiceImpl) GetReleaseDetail(ctx context.Context, req *interfaces.MCPServerReleaseDetailRequest) (resp *interfaces.MCPServerReleaseDetailResponse, err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	// 检查查看权限
	accessor, err := s.AuthService.GetAccessor(ctx, req.UserID)
	if err != nil {
		return
	}
	err = s.AuthService.CheckPublicAccessPermission(ctx, accessor, req.ID, interfaces.AuthResourceTypeMCP)
	if err != nil {
		return
	}

	release, err := s.DBMCPServerRelease.SelectByMCPID(ctx, nil, req.ID)
	if err != nil {
		err = fmt.Errorf("select mcp server release failed: %w", err)
		s.logger.WithContext(ctx).Error(err)
		err = infraerrors.NewHTTPError(ctx, http.StatusInternalServerError, infraerrors.ErrExtMCPParseFailed, err.Error())
		return
	}

	if release == nil {
		err = infraerrors.NewHTTPError(ctx, http.StatusNotFound, infraerrors.ErrExtMCPNotFound, "mcp not found")
		return
	}

	releaseInfo := s.releaseModelToResponse(release)

	// 渲染用户名称
	userIDs := []string{release.CreateUser, release.UpdateUser, release.ReleaseUser}
	userMap, err := s.UserMgnt.GetUsersName(ctx, userIDs)
	if err != nil {
		return
	}
	releaseInfo.CreateUser = utils.GetValueOrDefault(userMap, release.CreateUser, interfaces.UnknownUser)
	releaseInfo.UpdateUser = utils.GetValueOrDefault(userMap, release.UpdateUser, interfaces.UnknownUser)
	releaseInfo.ReleaseUser = utils.GetValueOrDefault(userMap, release.ReleaseUser, interfaces.UnknownUser)

	// 生成MCP Server连接信息
	connectionInfo := s.generateExternalConnectionInfo(release.MCPID, interfaces.MCPMode(release.Mode), interfaces.MCPCreationType(release.CreationType))

	resp = &interfaces.MCPServerReleaseDetailResponse{
		MCPServerDetailResponse: interfaces.MCPServerDetailResponse{
			BaseInfo:       releaseInfo,
			ConnectionInfo: connectionInfo,
		},
	}
	return
}

// QueryReleaseBatch 查询MCP Server Release列表
func (s *mcpServiceImpl) QueryReleaseBatch(ctx context.Context, req *interfaces.MCPServerReleaseBatchRequest) (mapData []map[string]any, err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	mcpIDs := strings.Split(req.MCPIDs, ",")
	fields := strings.Split(req.Fields, ",")

	columns := []string{}
	for _, field := range fields {
		if slices.Contains(interfaces.MCPFields, field) {
			columns = append(columns, "f_"+field)
		} else {
			err = infraerrors.DefaultHTTPError(ctx, http.StatusBadRequest, fmt.Sprintf("invalid field: %s", field))
			return
		}
	}

	mcpIDColumn := "f_mcp_id"
	if !slices.Contains(columns, mcpIDColumn) {
		columns = append(columns, mcpIDColumn)
	}

	// 查询MCP Server配置列表
	accessor, err := s.AuthService.GetAccessor(ctx, req.UserID)
	if err != nil {
		return
	}
	resp, err := auth.SelectListWithAuth(
		ctx, defaultPage, defaultPageSize, false,
		func() ([]*model.MCPServerReleaseDB, error) {
			return s.DBMCPServerRelease.SelectByMCPIDs(ctx, nil, mcpIDs, columns)
		},
		func() ([]string, error) {
			return s.AuthService.ResourceListIDs(ctx, accessor, interfaces.AuthResourceTypeMCP, interfaces.AuthOperationTypePublicAccess)
		},
	)
	if err != nil {
		return
	}

	releaseList := resp.Data
	userIDs := []string{}
	data := make([]*interfaces.MCPServerConfigInfo, 0, len(releaseList))
	for _, config := range releaseList {
		userIDs = append(userIDs, config.CreateUser, config.UpdateUser, config.ReleaseUser)
		releaseInfo := s.releaseModelToResponse(config)
		releaseInfo.Status = ""
		data = append(data, releaseInfo)
	}

	// 渲染用户名称
	userMap, err := s.UserMgnt.GetUsersName(ctx, userIDs)
	if err != nil {
		return
	}
	for _, config := range data {
		config.CreateUser = utils.GetValueOrDefault(userMap, config.CreateUser, interfaces.UnknownUser)
		config.UpdateUser = utils.GetValueOrDefault(userMap, config.UpdateUser, interfaces.UnknownUser)
		config.ReleaseUser = utils.GetValueOrDefault(userMap, config.ReleaseUser, interfaces.UnknownUser)
	}

	fields = append(fields, "mcp_id")
	mapData = make([]map[string]any, 0, len(data))
	for _, config := range data {
		mapData = append(mapData, config.ToMapByFields(fields))
	}
	return mapData, nil
}

func (s *mcpServiceImpl) releaseModelToResponse(config *model.MCPServerReleaseDB) *interfaces.MCPServerConfigInfo {
	return &interfaces.MCPServerConfigInfo{
		MCPCoreConfigInfo: interfaces.MCPCoreConfigInfo{
			Mode:    interfaces.MCPMode(config.Mode),
			Command: config.Command,
			Args:    utils.JSONToObject[[]string](config.Args),
			URL:     config.URL,
			Headers: utils.JSONToObject[map[string]string](config.Headers),
			Env:     utils.JSONToObject[map[string]string](config.Env),
		},
		MCPID:        config.MCPID,
		Name:         config.Name,
		Description:  config.Description,
		Status:       string(interfaces.BizStatusPublished),
		Source:       config.Source,
		IsInternal:   config.IsInternal,
		Category:     config.Category,
		CreateUser:   config.CreateUser,
		CreateTime:   config.CreateTime,
		UpdateUser:   config.UpdateUser,
		UpdateTime:   config.UpdateTime,
		ReleaseTime:  config.ReleaseTime,
		ReleaseUser:  config.ReleaseUser,
		Version:      config.Version,
		CreationType: interfaces.MCPCreationType(config.CreationType),
	}
}

func (s *mcpServiceImpl) publishMCP(ctx context.Context, tx *sql.Tx, mcpConfigDB *model.MCPServerConfigDB, userID string) (release *model.MCPServerReleaseDB, err error) {
	release, err = s.DBMCPServerRelease.SelectByMCPID(ctx, tx, mcpConfigDB.MCPID)
	if err != nil {
		s.logger.WithContext(ctx).Errorf("failed to check existing release: %v", err)
		err = infraerrors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	version := mcpConfigDB.Version

	// 如果存在，则更新版本
	if release != nil {
		s.configModelToReleaseModel(mcpConfigDB, release)
		release.Version = version
		release.ReleaseUser = userID // 设置发布用户

		err = s.DBMCPServerRelease.UpdateByMCPID(ctx, tx, release)
		if err != nil {
			err = fmt.Errorf("failed to update existing release: %w", err)
			s.logger.WithContext(ctx).Errorf("failed to update existing release: %v", err)
			err = infraerrors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
			return
		}
		return release, nil
	}

	// 不存在，则创建新版本
	release = &model.MCPServerReleaseDB{}
	s.configModelToReleaseModel(mcpConfigDB, release)
	release.Version = version
	release.CreationType = mcpConfigDB.CreationType
	release.CreateUser = mcpConfigDB.CreateUser
	release.CreateTime = mcpConfigDB.CreateTime
	release.UpdateUser = mcpConfigDB.UpdateUser
	release.UpdateTime = mcpConfigDB.UpdateTime
	release.ReleaseUser = userID

	err = s.DBMCPServerRelease.Insert(ctx, tx, release)
	if err != nil {
		err = fmt.Errorf("failed to create new release: %w", err)
		s.logger.WithContext(ctx).Errorf("failed to create new release: %v", err)
		err = infraerrors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	return release, nil
}

func (s *mcpServiceImpl) unpublishMCP(ctx context.Context, tx *sql.Tx, mcpConfigDB *model.MCPServerConfigDB) (err error) {
	err = s.DBMCPServerRelease.DeleteByMCPID(ctx, tx, mcpConfigDB.MCPID)
	if err != nil {
		err = fmt.Errorf("failed to delete release: %w", err)
		s.logger.WithContext(ctx).Errorf("failed to delete release: %v", err)
		err = infraerrors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
	}
	return
}

func (s *mcpServiceImpl) configModelToReleaseModel(config *model.MCPServerConfigDB, release *model.MCPServerReleaseDB) {
	release.MCPID = config.MCPID
	release.CreateUser = config.CreateUser
	release.CreateTime = config.CreateTime
	release.UpdateUser = config.UpdateUser
	release.UpdateTime = config.UpdateTime
	release.Name = config.Name
	release.Description = config.Description
	release.Mode = config.Mode
	release.URL = config.URL
	release.Headers = config.Headers
	release.Command = config.Command
	release.Env = config.Env
	release.Args = config.Args
	release.Category = config.Category
	release.Source = config.Source
	release.IsInternal = config.IsInternal

	release.ReleaseDesc = ""
}

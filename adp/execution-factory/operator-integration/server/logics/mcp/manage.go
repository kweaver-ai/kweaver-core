package mcp

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	icommon "github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common"
	infracommon "github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common/ormhelper"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	infraerrors "github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/telemetry"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/auth"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/common"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/metric"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
)

// 排序字段与数据库字段映射
var sortFieldMap = map[string]string{
	"create_time": "f_create_time",
	"update_time": "f_update_time",
	"name":        "f_name",
}

const (
	mcpToolMaxCount = 30
)

// ParseSSE 解析SSE MCPServer
func (s *mcpServiceImpl) ParseSSE(ctx context.Context, req *interfaces.MCPParseSSERequest) (resp *interfaces.MCPParseSSEResponse, err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	mcpCoreInfo := interfaces.MCPCoreConfigInfo{
		Mode:    req.Mode,
		URL:     req.URL,
		Headers: req.Headers,
	}

	listToolsReq := ListToolsRequest{
		MCPCoreInfo: &mcpCoreInfo,
	}

	toolsResponse, err := s.listTools(ctx, &listToolsReq)
	if err != nil {
		s.logger.WithContext(ctx).Errorf("list tools error: %v", err)
		return
	}
	resp = &interfaces.MCPParseSSEResponse{
		Tools:          toolsResponse.Tools,
		ServerInitInfo: toolsResponse.ServerInitInfo,
	}
	return
}

// AddMCPServer 添加MCP Server
func (s *mcpServiceImpl) AddMCPServer(ctx context.Context, req *interfaces.MCPServerAddRequest) (resp *interfaces.MCPServerAddResponse, err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	telemetry.SetSpanAttributes(ctx, map[string]interface{}{
		"user_id": req.UserID,
	})
	// 检查是否有新建权限
	accessor, err := s.AuthService.GetAccessor(ctx, req.UserID)
	if err != nil {
		return
	}
	err = s.AuthService.CheckCreatePermission(ctx, accessor, interfaces.AuthResourceTypeMCP)
	if err != nil {
		return
	}

	tx, err := s.DBTx.GetTx(ctx)
	if err != nil {
		s.logger.WithContext(ctx).Errorf("get tx failed, err: %v", err)
		err = infraerrors.DefaultHTTPError(ctx, http.StatusInternalServerError, fmt.Sprintf("get tx failed, err: %v", err))
		return
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()

	// 默认不是内置, 内置工具调用内置接口
	req.IsInternal = false
	mcpserverConfig := s.registerReqToModel(req)

	MCPID, err := s.addMCPConfig(ctx, tx, mcpserverConfig)
	if err != nil {
		return
	}

	// 添加mcp工具配置信息
	if req.CreationType == interfaces.MCPCreationTypeToolImported {
		var mcpTools []*model.MCPToolDB
		mcpTools, err = s.syncMCPTools(ctx, tx, req.UserID, MCPID, mcpserverConfig.Version, req.ToolConfigs)
		if err != nil {
			return nil, err
		}

		// 创建mcp Server实例
		err = s.createMCPServerInstance(ctx, mcpserverConfig, mcpTools)
		if err != nil {
			return nil, err
		}
	}

	// 触发新建策略，创建人默认拥有对当前资源的所有操作权限
	err = s.AuthService.CreateOwnerPolicy(ctx, accessor, &interfaces.AuthResource{
		ID:   MCPID,
		Type: string(interfaces.AuthResourceTypeMCP),
		Name: mcpserverConfig.Name,
	})
	if err != nil {
		return
	}
	// 记录审计日志
	go func() {
		accountAuthContext, ok := icommon.GetAccountAuthContextFromCtx(ctx)
		if !ok {
			s.logger.WithContext(ctx).Errorf("get account auth context from ctx error")
			return
		}
		s.AuditLog.Logger(ctx, &metric.AuditLogBuilderParams{
			TokenInfo: accountAuthContext.TokenInfo,
			Accessor:  accessor,
			Operation: metric.AuditLogOperationCreate,
			Object: &metric.AuditLogObject{
				Type: metric.AuditLogObjectMCP,
				Name: mcpserverConfig.Name,
				ID:   MCPID,
			},
		})
	}()

	resp = &interfaces.MCPServerAddResponse{
		MCPID:  MCPID,
		Status: string(interfaces.BizStatusUnpublish),
	}
	return
}

func (s *mcpServiceImpl) addMCPConfig(ctx context.Context, tx *sql.Tx, mcpConfig *model.MCPServerConfigDB) (string, error) {
	// 参数校验
	err := s.Validator.ValidatorMCPName(ctx, mcpConfig.Name)
	if err != nil {
		return "", err
	}
	err = s.Validator.ValidatorMCPDesc(ctx, mcpConfig.Description)
	if err != nil {
		return "", err
	}

	// 校验分类
	if !s.CategoryManager.CheckCategory(interfaces.BizCategory(mcpConfig.Category)) {
		return "", infraerrors.DefaultHTTPError(ctx, http.StatusBadRequest, "invalid category")
	}

	// 根据名称进行校验，名称不能重复
	err = s.checkDuplicateName(ctx, mcpConfig.Name, "")
	if err != nil {
		return "", err
	}

	MCPID, err := s.DBMCPServerConfig.Insert(ctx, tx, mcpConfig)
	if err != nil {
		s.logger.WithContext(ctx).Errorf("addMCPConfig Insert failed, err: %v", err)
		err = infraerrors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return "", err
	}

	// 关联业务域
	businessDomainId, _ := icommon.GetBusinessDomainFromCtx(ctx)
	err = s.BusinessDomainService.AssociateResource(ctx, businessDomainId, MCPID, interfaces.AuthResourceTypeMCP)
	if err != nil {
		return "", err
	}

	return MCPID, nil
}

// syncMCPTools 添加MCP工具配置信息
func (s *mcpServiceImpl) syncMCPTools(ctx context.Context, tx *sql.Tx, userID, mcpID string, mcpVersion int, toolConfigs []*interfaces.MCPToolConfigInfo) (mcpTools []*model.MCPToolDB, err error) {
	// todo: 校验工具数量不能超过30个
	if len(toolConfigs) > mcpToolMaxCount {
		return nil, infraerrors.NewHTTPError(ctx, http.StatusBadRequest, infraerrors.ErrExtMCPToolMaxCount, fmt.Sprintf("mcp tool count must be less than %d", mcpToolMaxCount), mcpToolMaxCount)
	}

	toolNames := make(map[string]bool)
	mcpTools = make([]*model.MCPToolDB, len(toolConfigs))
	for i, toolConfig := range toolConfigs {
		if toolConfig.ToolName != "" {
			// 校验工具名称是否重复
			if toolNames[toolConfig.ToolName] {
				return nil, infraerrors.NewHTTPError(ctx, http.StatusBadRequest, infraerrors.ErrExtMCPToolNameDuplicate, fmt.Sprintf("mcp tool name %s is duplicate", toolConfig.ToolName), toolConfig.ToolName)
			}
			toolNames[toolConfig.ToolName] = true
			// 校验工具名称是否合法
			err = s.Validator.ValidatorToolName(ctx, toolConfig.ToolName)
			if err != nil {
				return nil, err
			}
		}
		// 校验工具描述是否合法
		if toolConfig.ToolDescription != "" {
			err = s.Validator.ValidatorToolDesc(ctx, toolConfig.ToolDescription)
			if err != nil {
				return nil, err
			}
		}
		mcpTools[i] = &model.MCPToolDB{
			MCPID:       mcpID,
			MCPVersion:  mcpVersion,
			BoxID:       toolConfig.BoxID,
			BoxName:     toolConfig.BoxName,
			ToolID:      toolConfig.ToolID,
			Name:        toolConfig.ToolName,
			Description: toolConfig.ToolDescription,
			UseRule:     toolConfig.UseRule,
			CreateUser:  userID,
			UpdateUser:  userID,
		}
	}

	// 先根据mcpID和mcpVersion进行数据删除
	err = s.DBMCPTool.DeleteByMCPIDAndVersion(ctx, tx, mcpID, mcpVersion)
	if err != nil {
		s.logger.WithContext(ctx).Errorf("syncMCPTools DeleteByMCPIDAndVersion failed, err: %v", err)
		err = infraerrors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return nil, err
	}

	if len(mcpTools) > 0 {
		// 批量数据插入
		err = s.DBMCPTool.BatchInsert(ctx, tx, mcpTools)
		if err != nil {
			s.logger.WithContext(ctx).Errorf("syncMCPTools BatchInsert failed, err: %v", err)
			err = infraerrors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
			return nil, err
		}
	}
	return mcpTools, nil
}

// DeleteMCPServer 删除MCP Server
func (s *mcpServiceImpl) DeleteMCPServer(ctx context.Context, req *interfaces.MCPServerDeleteRequest) (err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	telemetry.SetSpanAttributes(ctx, map[string]interface{}{
		"mcp_id":  req.MCPID,
		"user_id": req.UserID,
	})
	// 检查删除权限
	accessor, err := s.AuthService.GetAccessor(ctx, req.UserID)
	if err != nil {
		return
	}
	err = s.AuthService.CheckDeletePermission(ctx, accessor, req.MCPID, interfaces.AuthResourceTypeMCP)
	if err != nil {
		return
	}

	tx, err := s.DBTx.GetTx(ctx)
	if err != nil {
		s.logger.WithContext(ctx).Errorf("get tx failed, err: %v", err)
		err = infraerrors.DefaultHTTPError(ctx, http.StatusInternalServerError, fmt.Sprintf("get tx failed, err: %v", err))
		return
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()

	// 删除MCP Server配置
	configDB, err := s.removeMCPConfig(ctx, tx, req.MCPID)
	if err != nil {
		return
	}

	// 删除MCP工具配置信息
	if configDB.CreationType == interfaces.MCPCreationTypeToolImported.String() {
		err = s.removeMCPTools(ctx, tx, req.MCPID, configDB.Version)
		if err != nil {
			return err
		}

		// 删除mcp Server实例
		err = s.MCPInstanceService.DeleteAllMCPInstances(ctx, req.MCPID)
		if err != nil {
			return err
		}
	}

	// 取消关联业务域
	businessDomainId, _ := icommon.GetBusinessDomainFromCtx(ctx)
	err = s.BusinessDomainService.DisassociateResource(ctx, businessDomainId, req.MCPID, interfaces.AuthResourceTypeMCP)
	if err != nil {
		return
	}

	// 触发权限策略删除
	err = s.AuthService.DeletePolicy(ctx, []string{req.MCPID}, interfaces.AuthResourceTypeMCP)
	if err != nil {
		return
	}
	// 记录审计日志
	go func() {
		accountAuthContext, ok := icommon.GetAccountAuthContextFromCtx(ctx)
		if !ok {
			s.logger.WithContext(ctx).Errorf("get account auth context from ctx error")
			return
		}
		s.AuditLog.Logger(ctx, &metric.AuditLogBuilderParams{
			TokenInfo: accountAuthContext.TokenInfo,
			Accessor:  accessor,
			Operation: metric.AuditLogOperationDelete,
			Object: &metric.AuditLogObject{
				Type: metric.AuditLogObjectMCP,
				ID:   req.MCPID,
				Name: configDB.Name,
			},
		})
	}()
	return nil
}

func (s *mcpServiceImpl) removeMCPConfig(ctx context.Context, tx *sql.Tx, mcpID string) (config *model.MCPServerConfigDB, err error) {
	// 检查MCP Server配置是否存在
	config, err = s.DBMCPServerConfig.SelectByID(ctx, tx, mcpID)
	if err != nil {
		s.logger.WithContext(ctx).Errorf("removeMCPConfig SelectByID failed, err: %v", err)
		err = infraerrors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	if config == nil {
		err = infraerrors.NewHTTPError(ctx, http.StatusNotFound, infraerrors.ErrExtMCPNotFound, "mcp not found")
		return
	}
	if config.Status != string(interfaces.BizStatusUnpublish) && config.Status != string(interfaces.BizStatusOffline) {
		err = infraerrors.NewHTTPError(ctx, http.StatusBadRequest, infraerrors.ErrExtMCPUnSupportDelete,
			fmt.Sprintf("current mcp status %s, can not be deleted", config.Status))
		return
	}
	// 删除MCP Server配置
	err = s.DBMCPServerConfig.DeleteByID(ctx, tx, mcpID)
	if err != nil {
		s.logger.WithContext(ctx).Errorf("delete mcp config failed, err: %v", err)
		err = infraerrors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	// 删除MCP Server发布历史
	err = s.DBMCPServerReleaseHistory.DeleteByMCPID(ctx, tx, mcpID)
	if err != nil {
		s.logger.WithContext(ctx).Errorf("delete mcp release history failed, err: %v", err)
		err = infraerrors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
	}
	return
}

func (s *mcpServiceImpl) removeMCPTools(ctx context.Context, tx *sql.Tx, mcpID string, mcpVersion int) (err error) {
	err = s.DBMCPTool.DeleteByMCPIDAndVersion(ctx, tx, mcpID, mcpVersion)
	if err != nil {
		s.logger.WithContext(ctx).Errorf("removeMCPTools DeleteByMCPIDAndVersion failed, err: %v", err)
		err = infraerrors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
	}
	return
}

// QueryPage 分页查询MCP Server列表
func (s *mcpServiceImpl) QueryPage(ctx context.Context, req *interfaces.MCPServerListRequest) (result *interfaces.MCPServerListResponse, err error) {
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
	if req.Status != "" {
		filter["status"] = req.Status
	}
	if req.CreateUser != "" {
		filter["create_user"] = req.CreateUser
	}
	if req.Mode != "" {
		filter["mode"] = req.Mode
	}

	// 排序字段
	sortField := "f_update_time"
	if req.SortBy != "" {
		sortField = sortFieldMap[req.SortBy]
		if sortField == "" {
			err = fmt.Errorf("invalid sort field: %s", req.SortBy)
			return
		}
	}
	// 查询MCP Server配置列表
	sort := &ormhelper.SortParams{
		Fields: []ormhelper.SortField{
			{Field: sortField, Order: ormhelper.SortOrder(req.SortOrder)},
		},
	}
	queryTotalFunc := func(newCtx context.Context) (int64, error) {
		var total int64
		total, err = s.DBMCPServerConfig.CountByWhereClause(newCtx, nil, filter)
		if err != nil {
			return 0, err
		}
		return total, nil
	}
	queryBatchFunc := func(newCtx context.Context, pageSize, offset int, cursorValue *model.MCPServerConfigDB) (
		[]*model.MCPServerConfigDB, error) {
		var configList []*model.MCPServerConfigDB
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
		configList, err = s.DBMCPServerConfig.SelectListPage(newCtx, nil, filter, sort, cursor)
		if err != nil {
			return nil, err
		}
		return configList, nil
	}

	businessDomainIds := strings.Split(req.BusinessDomainID, ",")
	resourceToBdMap, err := s.BusinessDomainService.BatchResourceList(ctx, businessDomainIds, interfaces.AuthResourceTypeMCP)
	if err != nil {
		return
	}

	queryBuilder := auth.NewQueryBuilder[model.MCPServerConfigDB]().
		SetPage(req.Page, req.PageSize).SetAll(req.All).
		SetQueryFunctions(queryTotalFunc, queryBatchFunc).
		SetFilteredQueryFunctions(func(newCtx context.Context, ids []string) (int64, error) {
			filter["in"] = ids
			return queryTotalFunc(newCtx)
		}, func(newCtx context.Context, pageSize, offset int, ids []string, cursorValue *model.MCPServerConfigDB) ([]*model.MCPServerConfigDB, error) {
			filter["in"] = ids
			return queryBatchFunc(newCtx, pageSize, offset, cursorValue)
		}).
		SetAuthFilter(func(newCtx context.Context) ([]string, error) {
			var accessor *interfaces.AuthAccessor
			accessor, err = s.AuthService.GetAccessor(newCtx, req.UserID)
			if err != nil {
				return nil, err
			}
			return s.AuthService.ResourceListIDs(newCtx, accessor, interfaces.AuthResourceTypeMCP, interfaces.AuthOperationTypeView)
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
		userIDs = append(userIDs, config.CreateUser, config.UpdateUser)
		data = append(data, s.modelToResponse(config))
	}

	// 获取工具配置信息
	toolConfigMap, err := s.getMCPToolConfigs(ctx, configList)
	if err != nil {
		return nil, err
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
		config.ToolConfigs = toolConfigMap[s.genToolConfigMapKey(config.MCPID, config.Version)]
	}

	queryResult := &ormhelper.QueryResult{
		Total:      int64(resp.TotalCount),
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: resp.TotalPage,
		HasNext:    resp.HasNext,
		HasPrev:    resp.HasPrev,
	}
	result = &interfaces.MCPServerListResponse{
		QueryResult: queryResult,
		Data:        data,
	}
	return
}

// GetDetail 获取MCP Server详情
func (s *mcpServiceImpl) GetDetail(ctx context.Context, req *interfaces.MCPServerDetailRequest) (resp *interfaces.MCPServerDetailResponse, err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	// 检查查看权限
	accessor, err := s.AuthService.GetAccessor(ctx, req.UserID)
	if err != nil {
		return
	}
	err = s.AuthService.CheckViewPermission(ctx, accessor, req.ID, interfaces.AuthResourceTypeMCP)
	if err != nil {
		return
	}

	mcpConfigDB, err := s.DBMCPServerConfig.SelectByID(ctx, nil, req.ID)
	if err != nil {
		s.logger.WithContext(ctx).Errorf("select mcp config by id failed, err: %v", err)
		err = infraerrors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	if mcpConfigDB == nil {
		err = infraerrors.NewHTTPError(ctx, http.StatusNotFound, infraerrors.ErrExtMCPNotFound, "mcp not found")
		return
	}

	mcpConfig := s.modelToResponse(mcpConfigDB)

	// 渲染用户名称
	userIDs := []string{mcpConfigDB.CreateUser, mcpConfigDB.UpdateUser}
	userMap, err := s.UserMgnt.GetUsersName(ctx, userIDs)
	if err != nil {
		return
	}
	mcpConfig.CreateUser = utils.GetValueOrDefault(userMap, mcpConfigDB.CreateUser, interfaces.UnknownUser)
	mcpConfig.UpdateUser = utils.GetValueOrDefault(userMap, mcpConfigDB.UpdateUser, interfaces.UnknownUser)

	// 组装响应结果
	response := &interfaces.MCPServerDetailResponse{
		BaseInfo: mcpConfig,
	}

	// 当前状态为发布状态时，生成MCP Server连接信息
	if mcpConfigDB.Status == string(interfaces.BizStatusPublished) {
		response.ConnectionInfo = s.generateExternalConnectionInfo(mcpConfigDB.MCPID, mcpConfig.Mode, mcpConfig.CreationType)
	}

	// 组装MCP工具配置信息
	if mcpConfig.CreationType == interfaces.MCPCreationTypeToolImported {
		toolConfigs, err := s.getMCPToolConfig(ctx, mcpConfig.MCPID, mcpConfig.Version)
		if err != nil {
			return nil, err
		}
		response.BaseInfo.ToolConfigs = toolConfigs
	}
	return response, nil
}

// UpdateMCPServer 更新MCP Server
func (s *mcpServiceImpl) UpdateMCPServer(ctx context.Context, req *interfaces.MCPServerUpdateRequest) (resp *interfaces.MCPServerUpdateResponse, err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	telemetry.SetSpanAttributes(ctx, map[string]interface{}{
		"mcp_id":  req.MCPID,
		"user_id": req.UserID,
	})
	// 检查编辑权限
	accessor, err := s.AuthService.GetAccessor(ctx, req.UserID)
	if err != nil {
		return
	}
	err = s.AuthService.CheckModifyPermission(ctx, accessor, req.MCPID, interfaces.AuthResourceTypeMCP)
	if err != nil {
		return
	}

	tx, err := s.DBTx.GetTx(ctx)
	if err != nil {
		s.logger.WithContext(ctx).Errorf("get db tx failed, err: %v", err)
		err = infraerrors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()

	// 自定义MCP Server更新
	newMCPConfig := s.mcpUpdateReqToModel(req)
	config, oldVersion, currentVersion, err := s.updateMCPConfig(ctx, tx, newMCPConfig)
	if err != nil {
		return
	}

	// 同步MCP工具配置信息
	if config.CreationType == interfaces.MCPCreationTypeToolImported.String() {
		mcpTools, err := s.syncMCPTools(ctx, tx, req.UserID, req.MCPID, config.Version, req.ToolConfigs)
		if err != nil {
			return nil, err
		}

		// 更新mcp Server实例
		err = s.refreshMCPServerInstance(ctx, oldVersion, currentVersion, config, mcpTools)
		if err != nil {
			return nil, err
		}
	}

	// 记录审计日志
	go func() {
		accountAuthContext, ok := infracommon.GetAccountAuthContextFromCtx(ctx)
		if !ok {
			s.logger.WithContext(ctx).Errorf("get account auth context from ctx failed")
			return
		}
		s.AuditLog.Logger(ctx, &metric.AuditLogBuilderParams{
			TokenInfo: accountAuthContext.TokenInfo,
			Accessor:  accessor,
			Operation: metric.AuditLogOperationEdit,
			Object: &metric.AuditLogObject{
				Type: metric.AuditLogObjectMCP,
				ID:   config.MCPID,
				Name: config.Name,
			},
		})
	}()
	resp = &interfaces.MCPServerUpdateResponse{
		MCPID:  config.MCPID,
		Status: interfaces.BizStatus(config.Status),
	}
	return
}

// updateMCPConfig 更新MCP Server配置
func (s *mcpServiceImpl) updateMCPConfig(ctx context.Context, tx *sql.Tx, newMCPConfig *model.MCPServerConfigDB) (config *model.MCPServerConfigDB, oldVersion, currentVersion int, err error) {
	// 参数校验
	err = s.Validator.ValidatorMCPName(ctx, newMCPConfig.Name)
	if err != nil {
		return nil, 0, 0, err
	}
	err = s.Validator.ValidatorMCPDesc(ctx, newMCPConfig.Description)
	if err != nil {
		return nil, 0, 0, err
	}
	// 校验分类
	if !s.CategoryManager.CheckCategory(interfaces.BizCategory(newMCPConfig.Category)) {
		return nil, 0, 0, infraerrors.DefaultHTTPError(ctx, http.StatusBadRequest, "invalid category")
	}

	// 根据ID获取MCP Server配置
	config, err = s.DBMCPServerConfig.SelectByID(ctx, nil, newMCPConfig.MCPID)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("select mcp server config failed: %w", err)
	}

	if config == nil {
		// 配置不存在
		return nil, 0, 0, infraerrors.NewHTTPError(ctx, http.StatusNotFound, infraerrors.ErrExtMCPNotFound, "mcp not found")
	}

	// 新增字段，兼容旧版本
	if config.CreationType == "" {
		config.CreationType = interfaces.MCPCreationTypeCustom.String()
	}

	// 校验状态转换是否合法
	targetState, err := common.GetEditStatusTrans(ctx, interfaces.BizStatus(config.Status))
	if err != nil {
		return nil, 0, 0, err
	}

	// 名字是否有变化
	isNameChange := config.Name != newMCPConfig.Name
	if isNameChange {
		// 当名称有变化时，校验名称是否重复
		err = s.checkDuplicateName(ctx, newMCPConfig.Name, config.MCPID)
		if err != nil {
			return nil, 0, 0, err
		}
	}

	// 更新版本号
	oldVersion = config.Version
	currentVersion, err = s.updateMCPConfigVersion(ctx, tx, config)
	if err != nil {
		return nil, 0, 0, err
	}

	// 更新MCP Server配置，定义哪些字段可以更新
	config.Name = newMCPConfig.Name
	config.Description = newMCPConfig.Description
	config.Source = newMCPConfig.Source
	config.Category = newMCPConfig.Category
	config.UpdateUser = newMCPConfig.UpdateUser
	config.UpdateTime = time.Now().UnixNano()

	config.Mode = newMCPConfig.Mode
	config.Command = newMCPConfig.Command
	config.Args = newMCPConfig.Args
	config.URL = newMCPConfig.URL
	config.Headers = newMCPConfig.Headers
	config.Env = newMCPConfig.Env
	config.Status = string(targetState)

	if config.CreationType == interfaces.MCPCreationTypeToolImported.String() {
		config.URL = s.generateInternalMCPURL(config.MCPID, config.Version, interfaces.MCPMode(config.Mode))
	}

	// 状态更新
	err = s.DBMCPServerConfig.UpdateByID(ctx, tx, config)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("update mcp server config failed: %w", err)
	}

	// 如果名称有变化，触发权限资源变更通知
	if isNameChange {
		authResource := &interfaces.AuthResource{
			ID:   config.MCPID,
			Name: config.Name,
			Type: string(interfaces.AuthResourceTypeMCP),
		}
		err = s.AuthService.NotifyResourceChange(ctx, authResource)
		if err != nil {
			return nil, 0, 0, err
		}
	}
	return config, oldVersion, currentVersion, nil
}

func (s *mcpServiceImpl) UpdateMCPStatus(ctx context.Context, req *interfaces.UpdateMCPStatusRequest) (resp *interfaces.UpdateMCPStatusResponse, err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	telemetry.SetSpanAttributes(ctx, map[string]interface{}{
		"mcp_id":  req.MCPID,
		"user_id": req.UserID,
	})
	// 检查发布或者下架权限
	accessor, err := s.AuthService.GetAccessor(ctx, req.UserID)
	if err != nil {
		return
	}
	var operation metric.AuditLogOperationType
	if req.Status == interfaces.BizStatusPublished {
		operation = metric.AuditLogOperationPublish
		err = s.AuthService.CheckPublishPermission(ctx, accessor, req.MCPID, interfaces.AuthResourceTypeMCP)
	} else if req.Status == interfaces.BizStatusOffline {
		operation = metric.AuditLogOperationUnpublish
		err = s.AuthService.CheckUnpublishPermission(ctx, accessor, req.MCPID, interfaces.AuthResourceTypeMCP)
	}
	if err != nil {
		return
	}

	var tx *sql.Tx
	tx, err = s.DBTx.GetTx(ctx)
	if err != nil {
		s.logger.WithContext(ctx).Errorf("get db tx failed, err: %v", err)
		err = infraerrors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()
	mcpConfigDB, resp, err := s.modifyMCPStatus(ctx, tx, req)
	if err != nil {
		return
	}
	if operation == "" {
		return
	}
	// 记录审计日志
	go func() {
		accountAuthContext, ok := infracommon.GetAccountAuthContextFromCtx(ctx)
		if !ok {
			s.logger.WithContext(ctx).Errorf("get account auth context from ctx error")
			return
		}
		s.AuditLog.Logger(ctx, &metric.AuditLogBuilderParams{
			TokenInfo: accountAuthContext.TokenInfo,
			Accessor:  accessor,
			Operation: operation,
			Object: &metric.AuditLogObject{
				Type: metric.AuditLogObjectMCP,
				ID:   req.MCPID,
				Name: mcpConfigDB.Name,
			},
		})
	}()
	return
}

func (s *mcpServiceImpl) modifyMCPStatus(ctx context.Context, tx *sql.Tx, req *interfaces.UpdateMCPStatusRequest) (mcpConfigDB *model.MCPServerConfigDB,
	resp *interfaces.UpdateMCPStatusResponse, err error) {
	// 检查MCP配置信息是否存在
	mcpConfigDB, err = s.DBMCPServerConfig.SelectByID(ctx, tx, req.MCPID)
	if err != nil {
		s.logger.WithContext(ctx).Errorf("select mcp server config failed, err: %v", err)
		err = infraerrors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	if mcpConfigDB == nil {
		err = infraerrors.NewHTTPError(ctx, http.StatusNotFound, infraerrors.ErrExtMCPNotFound, "mcp not found")
		return
	}

	// 校验状态转换是否合法
	if !common.CheckStatusTransition(interfaces.BizStatus(mcpConfigDB.Status), req.Status) {
		err = infraerrors.NewHTTPError(ctx, http.StatusBadRequest, infraerrors.ErrExtMCPStatusInvalid,
			fmt.Sprintf("current mcp status %s, can not be transition to %s", mcpConfigDB.Status, req.Status))
		return
	}

	mcpConfigDB.UpdateUser = req.UserID
	mcpConfigDB.UpdateTime = time.Now().UnixNano()

	switch req.Status {
	case interfaces.BizStatusPublished:
		// 检查是否重名
		err = s.checkDuplicateName(ctx, mcpConfigDB.Name, mcpConfigDB.MCPID)
		if err != nil {
			return
		}
		mcpConfigDB.Status = string(req.Status)
		// 发布MCP
		var mcpReleaseDB *model.MCPServerReleaseDB
		mcpReleaseDB, err = s.publishMCP(ctx, tx, mcpConfigDB, req.UserID)
		if err != nil {
			return
		}
		// 新增发布历史记录
		err = s.addMCPHistory(ctx, tx, mcpReleaseDB, req.UserID)
		if err != nil {
			return
		}
	case interfaces.BizStatusOffline:
		// 下架操作
		err = s.unpublishMCP(ctx, tx, mcpConfigDB)
		if err != nil {
			return
		}
	case interfaces.BizStatusUnpublish, interfaces.BizStatusEditing:
		// 编辑中或者未发布状态，更新版本号
		_, err = s.updateMCPConfigVersion(ctx, tx, mcpConfigDB)
		if err != nil {
			return
		}
	}

	// 更新MCP配置表状态
	err = s.DBMCPServerConfig.UpdateStatus(ctx, tx, req.MCPID, string(req.Status), req.UserID, mcpConfigDB.Version)
	if err != nil {
		s.logger.WithContext(ctx).Errorf("update mcp server config status failed, err: %v", err)
		err = infraerrors.DefaultHTTPError(ctx, http.StatusInternalServerError, fmt.Sprintf("update mcp server config status failed, err: %v", err))
		return
	}

	resp = &interfaces.UpdateMCPStatusResponse{
		MCPID:  req.MCPID,
		Status: req.Status,
	}
	return
}

// DebugTool 调试工具
func (s *mcpServiceImpl) DebugTool(ctx context.Context, req *interfaces.MCPToolDebugRequest) (resp *interfaces.MCPToolDebugResponse, err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	telemetry.SetSpanAttributes(ctx, map[string]interface{}{
		"mcp_id":  req.MCPID,
		"user_id": req.UserID,
	})
	// 校验使用权限
	accessor, err := s.AuthService.GetAccessor(ctx, req.UserID)
	if err != nil {
		return
	}
	err = s.AuthService.CheckExecutePermission(ctx, accessor, req.MCPID, interfaces.AuthResourceTypeMCP)
	if err != nil {
		return
	}

	// 1. 获取MCP Server配置
	mcpConfigDB, err := s.DBMCPServerConfig.SelectByID(ctx, nil, req.MCPID)
	if err != nil {
		s.logger.WithContext(ctx).Errorf("select mcp server config failed, err: %v", err)
		err = infraerrors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	if mcpConfigDB == nil {
		err = infraerrors.NewHTTPError(ctx, http.StatusNotFound, infraerrors.ErrExtMCPNotFound, "mcp not found")
		return
	}

	mcpConfig := s.modelToResponse(mcpConfigDB)

	// 2. 调用工具
	callToolReq := &CallToolRequest{
		ListToolsRequest: &ListToolsRequest{
			CreationType: mcpConfig.CreationType,
			MCPID:        mcpConfig.MCPID,
			Version:      mcpConfig.Version,
			MCPCoreInfo: &interfaces.MCPCoreConfigInfo{
				Mode:    mcpConfig.Mode,
				URL:     mcpConfig.URL,
				Headers: mcpConfig.Headers,
			},
		},
		ToolName: req.ToolName,
		Params:   req.Parameters,
	}
	callToolResp, err := s.callTool(ctx, callToolReq)
	if err != nil {
		return
	}

	// 记录审计日志
	go func() {
		accountAuthContext, ok := infracommon.GetAccountAuthContextFromCtx(ctx)
		if !ok {
			s.logger.WithContext(ctx).Errorf("get account auth context from ctx error")
			return
		}

		s.AuditLog.Logger(ctx, &metric.AuditLogBuilderParams{
			TokenInfo: accountAuthContext.TokenInfo,
			Accessor:  accessor,
			Operation: metric.AuditLogOperationExecute,
			Object: &metric.AuditLogObject{
				Type: metric.AuditLogObjectMCP,
				ID:   req.MCPID,
				Name: mcpConfig.Name,
			},
		})
	}()
	resp = &interfaces.MCPToolDebugResponse{
		Content: callToolResp.Content,
		IsError: callToolResp.IsError,
	}
	return
}

// registerReqToModel 将注册请求转换为模型
func (s *mcpServiceImpl) registerReqToModel(req *interfaces.MCPServerAddRequest) (config *model.MCPServerConfigDB) {
	config = &model.MCPServerConfigDB{
		MCPID:        uuid.New().String(),
		Version:      1,
		Name:         req.Name,
		Description:  req.Description,
		CreationType: req.CreationType.String(),
		Mode:         string(req.Mode),
		URL:          req.URL,
		Headers:      utils.ObjectToJSON(req.Headers),
		Command:      req.Command,
		Env:          utils.ObjectToJSON(req.Env),
		Args:         utils.ObjectToJSON(req.Args),
		Status:       string(interfaces.BizStatusUnpublish),
		Category:     req.Category,
		Source:       req.Source,
		IsInternal:   req.IsInternal,
		CreateUser:   req.UserID,
		UpdateUser:   req.UserID,
	}

	if req.CreationType == interfaces.MCPCreationTypeToolImported {
		config.Mode = interfaces.MCPModeStream.String()
		config.URL = s.generateInternalMCPURL(config.MCPID, config.Version, interfaces.MCPMode(config.Mode))
	}
	return config
}

// mcpUpdateReqToModel 将更新请求转换为模型
func (s *mcpServiceImpl) mcpUpdateReqToModel(req *interfaces.MCPServerUpdateRequest) *model.MCPServerConfigDB {
	return &model.MCPServerConfigDB{
		MCPID:        req.MCPID,
		CreationType: req.CreationType.String(),
		Name:         req.Name,
		Description:  req.Description,
		Source:       req.Source,
		IsInternal:   false,
		Category:     req.Category,
		UpdateUser:   req.UserID,
		UpdateTime:   time.Now().UnixNano(),
		Mode:         string(req.Mode),
		Command:      req.Command,
		Args:         utils.ObjectToJSON(req.Args),
		URL:          req.URL,
		Headers:      utils.ObjectToJSON(req.Headers),
		Env:          utils.ObjectToJSON(req.Env),
	}
}

// modelToResponse 将模型转换为响应
func (s *mcpServiceImpl) modelToResponse(config *model.MCPServerConfigDB) *interfaces.MCPServerConfigInfo {
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
		Version:      config.Version,
		CreationType: interfaces.MCPCreationType(config.CreationType),
		Name:         config.Name,
		Description:  config.Description,
		Status:       config.Status,
		Source:       config.Source,
		IsInternal:   config.IsInternal,
		Category:     config.Category,
		CreateUser:   config.CreateUser,
		CreateTime:   config.CreateTime,
		UpdateUser:   config.UpdateUser,
		UpdateTime:   config.UpdateTime,
	}
}

// checkDuplicateName 检查是否重名
func (s *mcpServiceImpl) checkDuplicateName(ctx context.Context, name, mcpID string) (err error) {
	// 根据名称进行校验，名称不能重复
	configDB, err := s.DBMCPServerConfig.SelectByName(ctx, nil, name, []string{interfaces.BizStatusPublished.String()})
	if err != nil {
		s.logger.WithContext(ctx).Errorf("checkDuplicateName count by name failed, name: %s, err: %v", name, err)
		err = infraerrors.DefaultHTTPError(ctx, http.StatusInternalServerError, fmt.Sprintf("check duplicate name, err: %s", err.Error()))
		return
	}
	if configDB == nil || (mcpID != "" && configDB.MCPID == mcpID) {
		return
	}
	err = infraerrors.NewHTTPError(ctx, http.StatusBadRequest, infraerrors.ErrExtMCPExists,
		fmt.Sprintf("mcp server name %s already exists, please use another name", name),
		name)
	return
}

// updateMCPConfigVersion 更新MCP配置表版本号
func (s *mcpServiceImpl) updateMCPConfigVersion(ctx context.Context, tx *sql.Tx, mcpConfigDB *model.MCPServerConfigDB) (version int, err error) {
	if mcpConfigDB.Status == string(interfaces.BizStatusPublished) || mcpConfigDB.Status == string(interfaces.BizStatusOffline) {
		// 为了向下兼容，版本号从发布历史中获取+1
		releaseHistorys, err := s.DBMCPServerReleaseHistory.SelectByMCPID(ctx, tx, mcpConfigDB.MCPID)
		if err != nil {
			return 0, err
		}
		if len(releaseHistorys) > 0 {
			mcpConfigDB.Version = releaseHistorys[0].Version + 1
		}
	}
	version = mcpConfigDB.Version
	return version, nil
}

// createMCPServerInstance 创建MCP Server实例
func (s *mcpServiceImpl) createMCPServerInstance(ctx context.Context, mcpConfigDB *model.MCPServerConfigDB, tools []*model.MCPToolDB) (err error) {
	// 创建mcp Server实例
	req := &interfaces.MCPInstanceCreateRequest{
		MCPID:        mcpConfigDB.MCPID,
		Version:      mcpConfigDB.Version,
		Name:         mcpConfigDB.Name,
		Instructions: mcpConfigDB.Description,
	}
	toolConfigs, err := s.getMCPToolDeployConfigs(ctx, tools)
	if err != nil {
		s.logger.WithContext(ctx).Warnf("createMCPServerInstance getMCPToolDeployConfigs failed, err: %v", err)
		return err
	}
	req.ToolConfigs = toolConfigs
	_, err = s.MCPInstanceService.CreateMCPInstance(ctx, req)
	if err != nil {
		return err
	}
	// todo: 选择更新mcp url
	return nil
}

func (s *mcpServiceImpl) UpgradeMCPInstance(ctx context.Context, mcpID string) (err error) {
	// 获取MCP配置信息
	mcpConfigDB, err := s.DBMCPServerConfig.SelectByID(ctx, nil, mcpID)
	if err != nil {
		return err
	}
	if mcpConfigDB == nil {
		return infraerrors.DefaultHTTPError(ctx, http.StatusNotFound, fmt.Sprintf("mcp server %s not found", mcpID))
	}
	// 获取工具信息
	tools, err := s.DBMCPTool.SelectListByMCPIDAndVersion(ctx, nil, mcpID, mcpConfigDB.Version)
	if err != nil {
		return err
	}
	toolConfigs, err := s.getMCPToolDeployConfigs(ctx, tools)
	if err != nil {
		s.logger.WithContext(ctx).Warnf("createMCPServerInstance getMCPToolDeployConfigs failed, err: %v", err)
		return err
	}
	req := &interfaces.MCPInstanceCreateRequest{
		MCPID:        mcpConfigDB.MCPID,
		Version:      mcpConfigDB.Version,
		Name:         mcpConfigDB.Name,
		Instructions: mcpConfigDB.Description,
		ToolConfigs:  toolConfigs,
	}
	_, err = s.MCPInstanceService.UpgradeMCPInstance(ctx, req)
	if err != nil {
		return err
	}
	return
}

// refreshMCPServerInstance 刷新MCP Server实例
func (s *mcpServiceImpl) refreshMCPServerInstance(ctx context.Context, oldVersion, currentVersion int, mcpConfigDB *model.MCPServerConfigDB, tools []*model.MCPToolDB) (err error) {
	if currentVersion > oldVersion {
		// 创建新的mcp Server实例
		return s.createMCPServerInstance(ctx, mcpConfigDB, tools)
	}
	// 更新mcp Server实例
	return s.updateMCPServerInstance(ctx, mcpConfigDB, tools)
}

func (s *mcpServiceImpl) updateMCPServerInstance(ctx context.Context, mcpConfigDB *model.MCPServerConfigDB, tools []*model.MCPToolDB) (err error) {
	// 更新mcp Server实例
	req := &interfaces.MCPInstanceUpdateRequest{
		MCPServerName: mcpConfigDB.Name,
		Instructions:  mcpConfigDB.Description,
	}
	toolConfigs, err := s.getMCPToolDeployConfigs(ctx, tools)
	if err != nil {
		return err
	}
	req.ToolConfigs = toolConfigs
	_, err = s.MCPInstanceService.UpdateMCPInstance(ctx, mcpConfigDB.MCPID, mcpConfigDB.Version, req)
	if err != nil {
		return err
	}
	return nil
}

// getMCPToolDeployConfigs 获取MCP工具部署配置
func (s *mcpServiceImpl) getMCPToolDeployConfigs(ctx context.Context, tools []*model.MCPToolDB) ([]*interfaces.MCPToolConfig, error) {
	toolConfigs := make([]*interfaces.MCPToolConfig, len(tools))
	for i, tool := range tools {
		mcpToolConfig, err := s.generateMCPToolConfig(ctx, tool)
		if err != nil {
			return nil, err
		}
		toolConfigs[i] = mcpToolConfig
	}
	return toolConfigs, nil
}

func (s *mcpServiceImpl) generateMCPToolConfig(ctx context.Context, tool *model.MCPToolDB) (*interfaces.MCPToolConfig, error) {
	toolConfig := &interfaces.MCPToolConfig{
		ToolID:      tool.MCPToolID,
		Description: tool.Description,
	}
	// 获取工具箱下工具信息
	toolInfo, err := s.ToolService.GetBoxTool(ctx, &interfaces.GetToolReq{
		BoxID:  tool.BoxID,
		ToolID: tool.ToolID,
	})
	if err != nil {
		return nil, err
	}

	if tool.Name != "" {
		toolConfig.Name = tool.Name
	} else {
		toolConfig.Name = toolInfo.Name
	}

	if tool.Description != "" {
		toolConfig.Description = tool.Description
	} else {
		toolConfig.Description = toolInfo.Description
	}
	if tool.UseRule != "" {
		toolConfig.Description += "\n use rule:" + tool.UseRule
	}

	// 将元数据信息转换为json schema
	toolConfig.InputSchema, err = s.convertInputSchema(ctx, toolInfo)
	if err != nil {
		return nil, err
	}
	return toolConfig, nil
}

func (s *mcpServiceImpl) convertInputSchema(ctx context.Context, toolInfo *interfaces.ToolInfo) (json.RawMessage, error) {
	if toolInfo.MetadataType != interfaces.MetadataTypeAPI && toolInfo.MetadataType != interfaces.MetadataTypeFunc {
		s.logger.WithContext(ctx).Warnf("unsupported metadata type: %s", toolInfo.MetadataType)
		err := errors.DefaultHTTPError(ctx, http.StatusBadRequest, fmt.Sprintf("unsupported metadata type: %s", toolInfo.MetadataType))
		return nil, err
	}
	if toolInfo.Metadata == nil {
		s.logger.WithContext(ctx).Warnf("tool metadata is nil")
		return nil, fmt.Errorf("tool metadata is nil")
	}
	metadata := toolInfo.Metadata
	if metadata.APISpec == nil {
		s.logger.WithContext(ctx).Warnf("tool apispec is nil")
		return nil, fmt.Errorf("tool apispec is nil")
	}
	converter := NewSimpleConverter()
	result := converter.ConvertFromBytes(utils.ObjectToByte(metadata.APISpec))
	if !result.Success {
		s.logger.WithContext(ctx).Warnf("convert metadata failed: %s", result.Error)
		return nil, fmt.Errorf("convert metadata failed: %s", result.Error)
	}
	return json.RawMessage(utils.ObjectToJSON(result.Data)), nil
}

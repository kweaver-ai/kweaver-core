package mcp

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	icommon "github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/metric"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
)

// Import 导入MCP
func (s *mcpServiceImpl) Import(ctx context.Context, tx *sql.Tx, mode interfaces.ImportType, data *interfaces.ComponentImpexConfigModel, userID string) (err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	if data == nil || data.MCP == nil || len(data.MCP.Configs) == 0 {
		err = errors.NewHTTPError(ctx, http.StatusBadRequest, errors.ErrExtCommonImportDataEmpty, "mcp configs is empty")
		return
	}
	// 导入预检查
	waitUpdataMCPList, err := s.importPreCheck(ctx, mode, data.MCP.Configs)
	if err != nil {
		return
	}
	accessor, err := s.AuthService.GetAccessor(ctx, userID)
	if err != nil {
		return
	}
	createMap, updateMap, depToolBoxMap, err := s.batchImportMcpMetadata(ctx, tx, data.MCP.Configs, waitUpdataMCPList, accessor)
	if err != nil {
		return
	}
	// 导入依赖
	if depToolBoxMap != nil && data.Toolbox != nil && len(data.Toolbox.Configs) > 0 {
		toolboxImportData := &interfaces.ComponentImpexConfigModel{
			Toolbox: &interfaces.ToolBoxImpexConfig{
				Configs: []*interfaces.ToolBoxImpexItem{},
			},
			Operator: data.Operator,
		}
		for _, tbItem := range data.Toolbox.Configs {
			if _, ok := depToolBoxMap[tbItem.BoxID]; ok {
				toolboxImportData.Toolbox.Configs = append(toolboxImportData.Toolbox.Configs, tbItem)
			}
		}
		err = s.ToolService.Import(ctx, tx, mode, toolboxImportData, userID)
		if err != nil {
			return
		}
	}

	// 导入后置操作：配置权限，添加审计日志
	s.importPostProcess(ctx, createMap, updateMap, accessor)
	return
}

// 导入后置操作：配置权限，添加审计日志
func (s *mcpServiceImpl) importPostProcess(ctx context.Context, createMCPMap, updateMCPMap map[string]*model.MCPServerConfigDB, accessor *interfaces.AuthAccessor) {
	// 触发新建策略，创建人默认拥有对当前资源的所有操作权限
	for _, mcpDB := range createMCPMap {
		err := s.AuthService.CreateOwnerPolicy(ctx, accessor, &interfaces.AuthResource{
			ID:   mcpDB.MCPID,
			Type: string(interfaces.AuthResourceTypeMCP),
			Name: mcpDB.Name,
		})
		if err != nil {
			s.logger.WithContext(ctx).Errorf("[importPostProcess] CreateOwnerPolicy err:%v", err)
		}
		// 记录设计日志及后续通知
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
					ID:   mcpDB.MCPID,
					Name: mcpDB.Name,
				},
			})
		}()
	}
	// 更新
	for _, mcpDB := range updateMCPMap {
		// 通知资源变更
		authResource := &interfaces.AuthResource{
			ID:   mcpDB.MCPID,
			Name: mcpDB.Name,
			Type: string(interfaces.AuthResourceTypeMCP),
		}
		err := s.AuthService.NotifyResourceChange(ctx, authResource)
		if err != nil {
			s.logger.WithContext(ctx).Errorf("[importPostProcess] CreateOwnerPolicy err:%v", err)
		}
		// 记录设计日志及后续通知
		go func() {
			accountAuthContext, ok := icommon.GetAccountAuthContextFromCtx(ctx)
			if !ok {
				s.logger.WithContext(ctx).Errorf("get account auth context from ctx error")
				return
			}
			s.AuditLog.Logger(ctx, &metric.AuditLogBuilderParams{
				TokenInfo: accountAuthContext.TokenInfo,
				Accessor:  accessor,
				Operation: metric.AuditLogOperationEdit,
				Object: &metric.AuditLogObject{
					Type: metric.AuditLogObjectMCP,
					ID:   mcpDB.MCPID,
					Name: mcpDB.Name,
				},
			})
		}()
	}
}

func (s *mcpServiceImpl) batchImportMcpMetadata(ctx context.Context, tx *sql.Tx, items []*interfaces.MCPServersImpexItem, waitUpdataMCPList []*model.MCPServerConfigDB,
	accessor *interfaces.AuthAccessor) (createMCPMap, updateMCPMap map[string]*model.MCPServerConfigDB,
	depToolBoxMap map[string]bool, err error) {
	// 收集需要新增的mcp
	createMCPMap = map[string]*model.MCPServerConfigDB{}
	// 收集需要更新的mcp
	updateMCPMap = map[string]*model.MCPServerConfigDB{}
	for _, mcpDB := range waitUpdataMCPList {
		// 检查MCP编辑权限
		err = s.AuthService.CheckModifyPermission(ctx, accessor, mcpDB.MCPID, interfaces.AuthResourceTypeMCP)
		if err != nil {
			return
		}
		// 内置MCP不允许编辑
		if mcpDB.IsInternal {
			err = errors.NewHTTPError(ctx, http.StatusForbidden, errors.ErrExtCommonInternalComponentNotAllowed,
				fmt.Sprintf("internal toolbox %v not allowed to update", mcpDB.MCPID), mcpDB.Name)
			return
		}
		updateMCPMap[mcpDB.MCPID] = mcpDB
	}
	// 检查是否存在依赖工具
	depToolBoxMap = map[string]bool{}
	for _, item := range items {
		var mcpTools []*model.MCPToolDB
		if mcpDB, ok := updateMCPMap[item.MCPID]; ok {
			mcpTools, err = s.importByUpsert(ctx, tx, mcpDB, item, accessor.ID)
		} else {
			createMCPMap[item.MCPID], mcpTools, err = s.importByCreate(ctx, tx, item, accessor.ID)
		}
		if err != nil {
			return
		}
		// 记录依赖工具
		for _, tool := range mcpTools {
			depToolBoxMap[tool.BoxID] = true
		}
	}
	return
}

func (s *mcpServiceImpl) importPreCheck(ctx context.Context, mode interfaces.ImportType, items []*interfaces.MCPServersImpexItem) (mcpList []*model.MCPServerConfigDB, err error) {
	// 收集mcpID、name 检查
	mcpIDs := []string{}
	for _, item := range items {
		mcpIDs = append(mcpIDs, item.MCPID)
		// 内置MCP不允许导入
		if item.IsInternal {
			err = errors.NewHTTPError(ctx, http.StatusForbidden, errors.ErrExtCommonInternalComponentNotAllowed,
				fmt.Sprintf("internal mcp %v not allowed to import", item.MCPID), item.Name)
			return
		}
		err = s.checkDuplicateName(ctx, item.Name, item.MCPID)
		if err != nil {
			return
		}
	}
	// 检查ID资源是否冲突
	mcpList, err = s.DBMCPServerConfig.SelectByMCPIDs(ctx, mcpIDs)
	if err != nil {
		s.logger.WithContext(ctx).Errorf("select mcp server config by ids error, err: %v", err)
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err)
		return
	}
	// 创建模式：如果MCP已存在，则返回冲突错误
	if mode == interfaces.ImportTypeCreate && len(mcpList) > 0 {
		err = errors.NewHTTPError(ctx, http.StatusConflict, errors.ErrExtCommonResourceIDConflict, "mcp id already exists")
	}
	return
}

func (s *mcpServiceImpl) importCheck(ctx context.Context, item *interfaces.MCPServersImpexItem) (err error) {
	err = s.Validator.ValidatorStruct(ctx, item)
	if err != nil {
		return
	}
	err = s.Validator.ValidatorMCPName(ctx, item.Name)
	if err != nil {
		return
	}
	err = s.Validator.ValidatorMCPDesc(ctx, item.Description)
	if err != nil {
		return
	}
	// 校验工具数量不能超过30个
	if len(item.MCPTools) > mcpToolMaxCount {
		err = errors.NewHTTPError(ctx, http.StatusBadRequest, errors.ErrExtMCPToolMaxCount, fmt.Sprintf("mcp tool count must be less than %d", mcpToolMaxCount), mcpToolMaxCount)
		return
	}
	item.IsInternal = false
	// 检查分类
	categoryName := s.CategoryManager.GetCategoryName(ctx, interfaces.BizCategory(item.Category))
	if categoryName == "" {
		item.Category = interfaces.CategoryTypeOther.String()
	}
	toolNames := make(map[string]bool)
	for _, toolConfig := range item.MCPTools {
		// 校验工具名称是否重复
		if toolNames[toolConfig.Name] {
			return errors.NewHTTPError(ctx, http.StatusBadRequest, errors.ErrExtMCPToolNameDuplicate, fmt.Sprintf("mcp tool name %s is duplicate", toolConfig.Name), toolConfig.Name)
		}
		err = s.Validator.ValidatorToolName(ctx, toolConfig.Name)
		if err != nil {
			return
		}
		err = s.Validator.ValidatorToolDesc(ctx, toolConfig.Description)
		if err != nil {
			return
		}
		toolNames[toolConfig.Name] = true
	}
	return nil
}

// importByCreate 新建
func (s *mcpServiceImpl) importByCreate(ctx context.Context, tx *sql.Tx, mcpConfigItem *interfaces.MCPServersImpexItem, userID string) (
	mcpConfigDB *model.MCPServerConfigDB, mcpTools []*model.MCPToolDB, err error) {
	// 校验导入的MCP Server配置信息
	err = s.importCheck(ctx, mcpConfigItem)
	if err != nil {
		return
	}
	// 导入参数检查
	mcpConfigDB = &model.MCPServerConfigDB{
		MCPID:        mcpConfigItem.MCPID,
		Name:         mcpConfigItem.Name,
		Description:  mcpConfigItem.Description,
		CreationType: mcpConfigItem.CreationType.String(),
		Mode:         mcpConfigItem.Mode.String(),
		URL:          mcpConfigItem.URL,
		Headers:      utils.ObjectToJSON(mcpConfigItem.Headers),
		Command:      mcpConfigItem.Command,
		Env:          utils.ObjectToJSON(mcpConfigItem.Env),
		Args:         utils.ObjectToJSON(mcpConfigItem.Args),
		Source:       mcpConfigItem.Source,
		Category:     mcpConfigItem.Category,
		IsInternal:   false,
		Version:      1,
		Status:       mcpConfigItem.Status.String(),
		CreateTime:   time.Now().UnixNano(),
		CreateUser:   userID,
		UpdateTime:   time.Now().UnixNano(),
		UpdateUser:   userID,
	}
	// 手动创建的MCP Server，需要更新依赖工具
	if mcpConfigDB.CreationType == interfaces.MCPCreationTypeToolImported.String() {
		mcpConfigDB.URL = s.generateInternalMCPURL(mcpConfigDB.MCPID, mcpConfigDB.Version, interfaces.MCPMode(mcpConfigDB.Mode))
		toolConfigs := []*interfaces.MCPToolConfigInfo{}
		for _, toolReq := range mcpConfigItem.MCPTools {
			toolConfig := &interfaces.MCPToolConfigInfo{
				BoxID:           toolReq.BoxID,
				ToolID:          toolReq.ToolID,
				BoxName:         toolReq.BoxName,
				ToolName:        toolReq.Name,
				ToolDescription: toolReq.Description,
				UseRule:         toolReq.UseRule,
			}
			toolConfigs = append(toolConfigs, toolConfig)
		}
		mcpTools, err = s.syncMCPTools(ctx, tx, userID, mcpConfigDB.MCPID, mcpConfigDB.Version, toolConfigs)
		if err != nil {
			return
		}
	}
	_, err = s.addMCPConfig(ctx, tx, mcpConfigDB)
	if err != nil {
		return
	}
	// 发布MCP
	if mcpConfigDB.Status == interfaces.BizStatusPublished.String() {
		_, err = s.publishMCP(ctx, tx, mcpConfigDB, userID)
		if err != nil {
			return
		}
	}
	return
}

// importByUpsert 更新或创建
func (s *mcpServiceImpl) importByUpsert(ctx context.Context, tx *sql.Tx, mcpConfigDB *model.MCPServerConfigDB, mcpConfigItem *interfaces.MCPServersImpexItem, userID string) (
	mcpTools []*model.MCPToolDB, err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	// 校验参数
	err = s.importCheck(ctx, mcpConfigItem)
	if err != nil {
		return
	}
	// 更新版本号
	currentVersion, err := s.updateMCPConfigVersion(ctx, tx, mcpConfigDB)
	if err != nil {
		return
	}
	// 更新MCP Server配置，定义哪些字段可以更新
	mcpConfigDB.Name = mcpConfigItem.Name
	mcpConfigDB.Description = mcpConfigItem.Description
	mcpConfigDB.Source = mcpConfigItem.Source
	mcpConfigDB.Category = mcpConfigItem.Category
	mcpConfigDB.UpdateUser = userID
	mcpConfigDB.UpdateTime = time.Now().UnixNano()
	mcpConfigDB.Version = currentVersion

	mcpConfigDB.Mode = mcpConfigItem.Mode.String()
	mcpConfigDB.Command = mcpConfigItem.Command
	mcpConfigDB.Args = utils.ObjectToJSON(mcpConfigItem.Args)
	mcpConfigDB.URL = mcpConfigItem.URL
	mcpConfigDB.Headers = utils.ObjectToJSON(mcpConfigItem.Headers)
	mcpConfigDB.Env = utils.ObjectToJSON(mcpConfigItem.Env)
	mcpConfigDB.Status = mcpConfigItem.Status.String()
	// 手动创建的MCP Server，需要更新依赖工具
	if mcpConfigDB.CreationType == interfaces.MCPCreationTypeToolImported.String() {
		mcpConfigDB.URL = s.generateInternalMCPURL(mcpConfigDB.MCPID, mcpConfigDB.Version, interfaces.MCPMode(mcpConfigDB.Mode))
		toolConfigs := []*interfaces.MCPToolConfigInfo{}
		for _, toolReq := range mcpConfigItem.MCPTools {
			toolConfig := &interfaces.MCPToolConfigInfo{
				BoxID:           toolReq.BoxID,
				ToolID:          toolReq.ToolID,
				BoxName:         toolReq.BoxName,
				ToolName:        toolReq.Name,
				ToolDescription: toolReq.Description,
				UseRule:         toolReq.UseRule,
			}
			toolConfigs = append(toolConfigs, toolConfig)
		}
		mcpTools, err = s.syncMCPTools(ctx, tx, userID, mcpConfigDB.MCPID, mcpConfigDB.Version, toolConfigs)
		if err != nil {
			return
		}
	}
	// 状态更新
	err = s.DBMCPServerConfig.UpdateByID(ctx, tx, mcpConfigDB)
	if err != nil {
		s.logger.WithContext(ctx).Errorf("update mcp config err: %s", err.Error())
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err)
		return
	}
	// 发布MCP
	if mcpConfigDB.Status == interfaces.BizStatusPublished.String() {
		_, err = s.publishMCP(ctx, tx, mcpConfigDB, userID)
		if err != nil {
			return
		}
	}
	return
}

// 导出预检查
func (s *mcpServiceImpl) exportPreCheck(ctx context.Context, req *interfaces.ExportReq) (mcpConfigDBs []*model.MCPServerConfigDB, err error) {
	// 批量鉴权
	var accessor *interfaces.AuthAccessor
	accessor, err = s.AuthService.GetAccessor(ctx, req.UserID)
	if err != nil {
		return
	}
	// 检查查看权限
	checkMCPIDs, err := s.AuthService.ResourceFilterIDs(ctx, accessor, req.IDs,
		interfaces.AuthResourceTypeMCP, interfaces.AuthOperationTypeView)
	if err != nil {
		return
	}
	if len(checkMCPIDs) != len(req.IDs) {
		clist := utils.FindMissingElements(req.IDs, checkMCPIDs)
		err = errors.NewHTTPError(ctx, http.StatusForbidden, errors.ErrExtCommonOperationForbidden,
			fmt.Sprintf("mcp server config %v not access", clist))
		return
	}
	// 检查数据是否存在
	mcpConfigDBs, err = s.DBMCPServerConfig.SelectByMCPIDs(ctx, req.IDs)
	if err != nil {
		s.logger.WithContext(ctx).Errorf("select mcp server config by ids failed, err: %v", err)
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err)
		return
	}
	if len(mcpConfigDBs) != len(req.IDs) {
		checkMCPIDs := []string{}
		for _, configDB := range mcpConfigDBs {
			checkMCPIDs = append(checkMCPIDs, configDB.MCPID)
		}
		clist := utils.FindMissingElements(req.IDs, checkMCPIDs)
		err = errors.NewHTTPError(ctx, http.StatusNotFound, errors.ErrExtMCPNotFound,
			fmt.Sprintf("mcp server config %v not found", clist))
		return
	}
	return
}

// Export 导出MCP
func (s *mcpServiceImpl) Export(ctx context.Context, req *interfaces.ExportReq) (data *interfaces.ComponentImpexConfigModel, err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	// 导出预检查
	mcpConfigDBs, err := s.exportPreCheck(ctx, req)
	if err != nil {
		return
	}
	// 收集所有的MCP信息
	data = &interfaces.ComponentImpexConfigModel{
		MCP: &interfaces.MCPImpexConfig{},
	}
	configs, depToolBoxIDs, err := s.batchGetExportMetadata(ctx, mcpConfigDBs)
	if err != nil {
		return
	}
	data.MCP = configs
	// 导出依赖
	depToolBoxIDs = utils.UniqueStrings(depToolBoxIDs)
	if len(depToolBoxIDs) == 0 {
		return
	}
	toolboxImpexConfig, err := s.ToolService.Export(ctx, &interfaces.ExportReq{
		UserID: req.UserID,
		IDs:    depToolBoxIDs,
	})
	if err != nil {
		return
	}
	if toolboxImpexConfig.Toolbox != nil {
		data.Toolbox = toolboxImpexConfig.Toolbox
	}

	if toolboxImpexConfig.Operator != nil {
		data.Operator = toolboxImpexConfig.Operator
	}
	return
}

// 收集导出元数据及其依赖
func (s *mcpServiceImpl) batchGetExportMetadata(ctx context.Context, mcpConfigDBs []*model.MCPServerConfigDB) (config *interfaces.MCPImpexConfig,
	depToolBoxIDs []string, err error) {
	config = &interfaces.MCPImpexConfig{}
	depToolBoxIDs = []string{}
	// 依赖的工具
	for _, configDB := range mcpConfigDBs {
		// 内置MCP不允许导出
		if configDB.IsInternal {
			err = errors.NewHTTPError(ctx, http.StatusForbidden, errors.ErrExtCommonInternalComponentNotAllowed,
				fmt.Sprintf("internal mcp %v not allowed to export", configDB.MCPID), configDB.Name)
			return
		}
		var mcpConfig *interfaces.MCPServersImpexItem
		mcpConfig, err = s.assembleMCPServersImpexModel(ctx, configDB)
		if err != nil {
			return
		}
		// 如果时工具导入收集
		if configDB.CreationType == interfaces.MCPCreationTypeToolImported.String() {
			// 获取依赖的工具
			mcpConfig.MCPTools = []*interfaces.MCPToolItem{}
			var mcpToolDBs []*model.MCPToolDB
			mcpToolDBs, err = s.DBMCPTool.SelectListByMCPIDAndVersion(ctx, nil, configDB.MCPID, configDB.Version)
			if err != nil {
				s.logger.WithContext(ctx).Errorf("select mcp server tool by ids failed, err: %v", err)
				err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err)
				return
			}
			for _, mcpToolDB := range mcpToolDBs {
				depToolBoxIDs = append(depToolBoxIDs, mcpToolDB.BoxID)
				mcpConfig.MCPTools = append(mcpConfig.MCPTools, &interfaces.MCPToolItem{
					MCPToolID:   mcpToolDB.MCPToolID,
					MCPID:       mcpToolDB.MCPID,
					MCPVersion:  mcpToolDB.MCPVersion,
					BoxID:       mcpToolDB.BoxID,
					BoxName:     mcpToolDB.BoxName,
					ToolID:      mcpToolDB.ToolID,
					Name:        mcpToolDB.Name,
					Description: mcpToolDB.Description,
					UseRule:     mcpToolDB.UseRule,
				})
			}
		}
		config.Configs = append(config.Configs, mcpConfig)
	}
	return
}

// 组装MCPServersImpexModel
func (s *mcpServiceImpl) assembleMCPServersImpexModel(ctx context.Context, configDB *model.MCPServerConfigDB) (mcpConfig *interfaces.MCPServersImpexItem, err error) {
	// 收集MCP信息
	args := []string{}
	if configDB.Args != "" {
		err = utils.StringToObject(configDB.Args, &args)
		if err != nil {
			s.logger.WithContext(ctx).Errorf("unmarshal mcp server config args failed, err: %v", err)
			err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err)
			return
		}
	}
	headers := map[string]string{}
	if configDB.Headers != "" {
		err = utils.StringToObject(configDB.Headers, &headers)
		if err != nil {
			s.logger.WithContext(ctx).Errorf("unmarshal mcp server config headers failed, err: %v", err)
			err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err)
			return
		}
	}
	env := map[string]string{}
	if configDB.Env != "" {
		err = utils.StringToObject(configDB.Env, &env)
		if err != nil {
			s.logger.WithContext(ctx).Errorf("unmarshal mcp server config env failed, err: %v", err)
			err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err)
			return
		}
	}
	mcpConfig = &interfaces.MCPServersImpexItem{
		MCPCoreConfigInfo: interfaces.MCPCoreConfigInfo{
			Mode:    interfaces.MCPMode(configDB.Mode),
			Command: configDB.Command,
			Args:    args,
			URL:     configDB.URL,
			Headers: headers,
			Env:     env,
		},
		MCPID:        configDB.MCPID,
		Version:      configDB.Version,
		CreationType: interfaces.MCPCreationType(configDB.CreationType),
		Name:         configDB.Name,
		Description:  configDB.Description,
		Status:       interfaces.BizStatus(configDB.Status),
		Source:       configDB.Source,
		IsInternal:   configDB.IsInternal,
		Category:     configDB.Category,
		CreateUser:   configDB.CreateUser,
		CreateTime:   configDB.CreateTime,
		UpdateUser:   configDB.UpdateUser,
		UpdateTime:   configDB.UpdateTime,
	}
	return
}

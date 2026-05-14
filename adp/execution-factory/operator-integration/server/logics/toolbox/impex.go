package toolbox

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/creasty/defaults"
	"github.com/google/uuid"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common"
	icommon "github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/metadata"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/metric"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
)

// Import 导入
func (s *ToolServiceImpl) Import(ctx context.Context, tx *sql.Tx, mode interfaces.ImportType, data *interfaces.ComponentImpexConfigModel, userID string) (err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	if data == nil || data.Toolbox == nil || len(data.Toolbox.Configs) == 0 {
		err = errors.NewHTTPError(ctx, http.StatusBadRequest, errors.ErrExtCommonImportDataEmpty, "toolbox configs is empty")
		return
	}
	// 导入预检查
	waitUpdataBoxList, err := s.importPreCheck(ctx, mode, data.Toolbox.Configs)
	if err != nil {
		return
	}
	var accessor *interfaces.AuthAccessor
	if icommon.IsPublicAPIFromCtx(ctx) {
		accessor, err = s.AuthService.GetAccessor(ctx, userID)
		if err != nil {
			s.Logger.WithContext(ctx).Warnf("[Import] GetAccessor err:%v", err)
			return
		}
	}
	// 导入工具箱、工具信息
	createMap, updateMap, err := s.batchImportToolBoxMetadata(ctx, tx, data.Toolbox.Configs, waitUpdataBoxList, accessor, userID)
	if err != nil {
		s.Logger.WithContext(ctx).Warnf("[Import] batchImportToolBoxMetadata err:%v", err)
		return
	}
	// 导入依赖
	if data.Operator != nil && len(data.Operator.Configs) > 0 {
		err = s.OperatorMgnt.Import(ctx, tx, mode, data.Operator, userID)
		if err != nil {
			s.Logger.WithContext(ctx).Warnf("[Import] OperatorMgnt.Import err:%v", err)
			return
		}
	}
	// 导入后置处理
	err = s.importPostProcess(ctx, createMap, updateMap, accessor)
	if err != nil {
		s.Logger.WithContext(ctx).Warnf("[Import] importPostProcess err:%v", err)
	}
	return
}

// 后置操作：添加权限配置，及审计日志记录
func (s *ToolServiceImpl) importPostProcess(ctx context.Context, createBoxMap, updateBoxMap map[string]*model.ToolboxDB, accessor *interfaces.AuthAccessor) (err error) {
	businessDomainID, _ := icommon.GetBusinessDomainFromCtx(ctx)
	for _, boxDB := range createBoxMap {
		// 关联业务域
		err = s.BusinessDomainService.AssociateResource(ctx, businessDomainID, boxDB.BoxID, interfaces.AuthResourceTypeToolBox)
		if err != nil {
			s.Logger.WithContext(ctx).Errorf("[importPostProcess] AssociateResource err:%v", err)
			return
		}

		// 触发新建策略，创建人默认拥有对当前资源的所有操作权限（内部调用不创建）
		if accessor != nil {
			err := s.AuthService.CreateOwnerPolicy(ctx, accessor, &interfaces.AuthResource{
				ID:   boxDB.BoxID,
				Type: interfaces.AuthResourceTypeToolBox.String(),
				Name: boxDB.Name,
			})
			if err != nil {
				s.Logger.WithContext(ctx).Errorf("[importPostProcess] CreateOwnerPolicy err:%v", err)
			}
		}
		// 记录设计日志及后续通知（内部调用不记录）
		if accessor != nil {
			go func() {
				accountAuthContext, ok := common.GetAccountAuthContextFromCtx(ctx)
				if !ok {
					s.Logger.WithContext(ctx).Warnf("[importPostProcess] GetAccountAuthContextFromCtx err :%v", err)
					return
				}
				s.AuditLog.Logger(ctx, &metric.AuditLogBuilderParams{
					TokenInfo: accountAuthContext.TokenInfo,
					Accessor:  accessor,
					Operation: metric.AuditLogOperationCreate,
					Object: &metric.AuditLogObject{
						Type: metric.AuditLogObjectTool,
						ID:   boxDB.BoxID,
						Name: boxDB.Name,
					},
				})
			}()
		}
		// 内置组件：创建全员授权策略（public_access + execute）
		if boxDB.IsInternal {
			policyErr := s.AuthService.CreateIntCompPolicyForAllUsers(ctx, &interfaces.AuthResource{
				ID:   boxDB.BoxID,
				Type: interfaces.AuthResourceTypeToolBox.String(),
				Name: boxDB.Name,
			})
			if policyErr != nil {
				s.Logger.WithContext(ctx).Errorf("[importPostProcess] CreateIntCompPolicyForAllUsers err:%v", policyErr)
			}
		}
	}
	// 更新工具箱
	for _, boxDB := range updateBoxMap {
		// 通知资源变更
		authResource := &interfaces.AuthResource{
			ID:   boxDB.BoxID,
			Name: boxDB.Name,
			Type: interfaces.AuthResourceTypeToolBox.String(),
		}
		err := s.AuthService.NotifyResourceChange(ctx, authResource)
		if err != nil {
			s.Logger.WithContext(ctx).Errorf("[importPostProcess] NotifyResourceChange err:%v", err)
		}
		// 内置组件：创建全员授权策略（public_access + execute）
		if boxDB.IsInternal {
			policyErr := s.AuthService.CreateIntCompPolicyForAllUsers(ctx, &interfaces.AuthResource{
				ID:   boxDB.BoxID,
				Type: interfaces.AuthResourceTypeToolBox.String(),
				Name: boxDB.Name,
			})
			if policyErr != nil {
				s.Logger.WithContext(ctx).Errorf("[importPostProcess] CreateIntCompPolicyForAllUsers err:%v", policyErr)
			}
		}
		// 记录设计日志及后续通知（内部调用不记录）
		if accessor != nil {
			go func() {
				accountAuthContext, ok := common.GetAccountAuthContextFromCtx(ctx)
				if !ok {
					s.Logger.WithContext(ctx).Warnf("[importPostProcess] GetAccountAuthContextFromCtx err :%v", err)
					return
				}
				s.AuditLog.Logger(ctx, &metric.AuditLogBuilderParams{
					TokenInfo: accountAuthContext.TokenInfo,
					Accessor:  accessor,
					Operation: metric.AuditLogOperationEdit,
					Object: &metric.AuditLogObject{
						Type: metric.AuditLogObjectTool,
						ID:   boxDB.BoxID,
						Name: boxDB.Name,
					},
				})
			}()
		}
	}
	return nil
}

// 导入预备检查
func (s *ToolServiceImpl) importPreCheck(ctx context.Context, mode interfaces.ImportType, items []*interfaces.ToolBoxImpexItem) (boxList []*model.ToolboxDB, err error) {
	// 收集工具箱ID，及名字
	boxIDs := []string{}
	for _, item := range items {
		boxIDs = append(boxIDs, item.BoxID)
		if icommon.IsPublicAPIFromCtx(ctx) && item.IsInternal {
			err = errors.NewHTTPError(ctx, http.StatusForbidden, errors.ErrExtCommonInternalComponentNotAllowed,
				fmt.Sprintf("internal toolbox %v not allowed to import", item.BoxID), item.BoxName)
			return
		}
		// 工具箱重名校验
		err = s.checkBoxDuplicateName(ctx, item.BoxName, item.BoxID)
		if err != nil {
			return
		}
	}
	// 检查ID资源是否冲突
	boxIDs = utils.UniqueStrings(boxIDs)
	boxList, err = s.ToolBoxDB.SelectListByBoxIDs(ctx, boxIDs)
	if err != nil {
		s.Logger.WithContext(ctx).Errorf("select toolbox by ids failed, err: %v", err)
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	// 创建模式：如果工具箱已存在，则返回冲突错误
	if mode == interfaces.ImportTypeCreate && len(boxList) > 0 {
		err = errors.NewHTTPError(ctx, http.StatusConflict, errors.ErrExtCommonResourceIDConflict, "toolbox id already exists")
	}
	return
}

// 批量导入工具箱及工具元数据
func (s *ToolServiceImpl) batchImportToolBoxMetadata(ctx context.Context, tx *sql.Tx, items []*interfaces.ToolBoxImpexItem, waitUpdataBoxList []*model.ToolboxDB,
	accessor *interfaces.AuthAccessor, userID string) (createBoxMap, updateBoxMap map[string]*model.ToolboxDB, err error) {
	// 收集需要新增的ToolBox
	createBoxMap = map[string]*model.ToolboxDB{}
	// 收集需要更新的工具ToolBox
	updateBoxMap = map[string]*model.ToolboxDB{}
	// 获取用户ID（内部调用使用传入 userID）
	uid := userID
	if accessor != nil {
		uid = accessor.ID
	}
	// 检查是否有更新权限，并收集需要更新的工具箱
	for _, boxDB := range waitUpdataBoxList {
		// 检查工具箱编辑权限（内部调用不鉴权）
		if icommon.IsPublicAPIFromCtx(ctx) {
			err = s.AuthService.CheckModifyPermission(ctx, accessor, boxDB.BoxID, interfaces.AuthResourceTypeToolBox)
			if err != nil {
				return
			}
			// 内置工具箱不能编辑
			if boxDB.IsInternal {
				err = errors.NewHTTPError(ctx, http.StatusForbidden, errors.ErrExtCommonInternalComponentNotAllowed,
					fmt.Sprintf("internal toolbox %v not allowed to update", boxDB.BoxID), boxDB.Name)
				return
			}
		}
		updateBoxMap[boxDB.BoxID] = boxDB
	}
	// 遍历导入项，根据是否存在工具箱ID判断是新增还是更新
	for _, item := range items {
		if boxDB, ok := updateBoxMap[item.BoxID]; ok {
			err = s.importByUpsert(ctx, tx, boxDB, item, uid)
			if err != nil {
				return
			}
		} else {
			boxDB, err = s.importByCreate(ctx, tx, item, uid)
			if err != nil {
				return
			}
			createBoxMap[boxDB.BoxID] = boxDB
		}
	}
	return
}

// importByCreate 导入工具箱
func (s *ToolServiceImpl) importByCreate(ctx context.Context, tx *sql.Tx, item *interfaces.ToolBoxImpexItem, userID string) (boxDB *model.ToolboxDB, err error) {
	// 校验导入的工具箱信息
	toolDBs, metadataDBs, err := s.importCheck(ctx, item, userID)
	if err != nil {
		return
	}
	// 添加工具箱
	boxDB = &model.ToolboxDB{
		BoxID:        item.BoxID,
		Name:         item.BoxName,
		Description:  item.BoxDesc,
		Source:       item.Source,
		ServerURL:    item.BoxSvcURL,
		Category:     item.CategoryType,
		Status:       item.Status.String(),
		IsInternal:   item.IsInternal,
		CreateTime:   time.Now().UnixNano(),
		CreateUser:   userID,
		UpdateUser:   userID,
		UpdateTime:   time.Now().UnixNano(),
		MetadataType: string(item.MetadataType),
	}
	if item.Status == interfaces.BizStatusPublished {
		boxDB.ReleaseUser = userID
		boxDB.ReleaseTime = time.Now().UnixNano()
	}
	_, err = s.ToolBoxDB.InsertToolBox(ctx, tx, boxDB)
	if err != nil {
		s.Logger.WithContext(ctx).Errorf("insert toolbox failed, err: %v", err)
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	// 处理元数据
	metadataMap := map[string]interfaces.IMetadataDB{}
	for _, metadataDB := range metadataDBs {
		version := metadataDB.GetVersion()
		metadataDB.SetVersion(uuid.New().String())
		metadataMap[version] = metadataDB
	}
	newMetadataDBs := []interfaces.IMetadataDB{}
	toolIDs := []string{}
	for _, toolDB := range toolDBs {
		if metadataDB, ok := metadataMap[toolDB.SourceID]; ok {
			toolDB.SourceID = metadataDB.GetVersion()
			newMetadataDBs = append(newMetadataDBs, metadataDB)
		}
		toolIDs = append(toolIDs, toolDB.ToolID)
	}
	// 检查工具是否重复
	duplicateTools, err := s.ToolDB.SelectToolBoxByToolIDs(ctx, toolIDs)
	if err != nil {
		s.Logger.WithContext(ctx).Errorf("select tool by source ids failed, err: %v", err)
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	if len(duplicateTools) > 0 {
		err = errors.NewHTTPError(ctx, http.StatusConflict, errors.ErrExtCommonResourceIDConflict, fmt.Sprintf("tool resource conflict, tool ids: %v", toolIDs))
		return
	}
	// 添加元数据
	if len(newMetadataDBs) > 0 {
		_, err = s.MetadataService.BatchRegisterMetadata(ctx, tx, newMetadataDBs)
		if err != nil {
			s.Logger.WithContext(ctx).Errorf("insert metadata failed, err: %v", err)
			err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
			return
		}
	}
	// 添加工具
	if len(toolDBs) > 0 {
		_, err = s.ToolDB.InsertTools(ctx, tx, toolDBs)
		if err != nil {
			s.Logger.WithContext(ctx).Errorf("insert tool failed, err: %v", err)
			err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
			return
		}
	}
	return
}

// importByUpsert 更新或创建
func (s *ToolServiceImpl) importByUpsert(ctx context.Context, tx *sql.Tx, toolBoxDB *model.ToolboxDB, item *interfaces.ToolBoxImpexItem, userID string) (err error) {
	// 校验导入的工具箱信息
	toolDBs, metadataDBs, err := s.importCheck(ctx, item, userID)
	if err != nil {
		return
	}
	// 检查工具箱元数据是否一致
	if toolBoxDB.MetadataType != string(item.MetadataType) {
		err = errors.NewHTTPError(ctx, http.StatusBadRequest, errors.ErrExtCommonMetadataTypeConflict,
			fmt.Sprintf("toolbox %s metadata type conflict, expect %v, got %v", toolBoxDB.BoxID, toolBoxDB.MetadataType, item.MetadataType))
		return
	}
	toolBoxDB.Name = item.BoxName
	toolBoxDB.Description = item.BoxDesc
	toolBoxDB.ServerURL = item.BoxSvcURL
	toolBoxDB.Category = item.CategoryType
	toolBoxDB.IsInternal = item.IsInternal
	toolBoxDB.UpdateTime = time.Now().UnixNano()
	toolBoxDB.UpdateUser = userID
	toolBoxDB.Status = item.Status.String()
	if item.Status == interfaces.BizStatusPublished {
		toolBoxDB.ReleaseUser = userID
		toolBoxDB.ReleaseTime = time.Now().UnixNano()
	}
	err = s.ToolBoxDB.UpdateToolBox(ctx, tx, toolBoxDB)
	if err != nil {
		s.Logger.WithContext(ctx).Errorf("update toolbox failed, err: %v", err)
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	// 获取工具箱内的工具
	tools, err := s.ToolDB.SelectToolByBoxID(ctx, toolBoxDB.BoxID)
	if err != nil {
		s.Logger.WithContext(ctx).Errorf("select tools failed, err: %v", err)
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	// 删除工具箱中的工具
	err = s.deleteTools(ctx, tx, toolBoxDB.BoxID, tools)
	if err != nil {
		return
	}
	// 添加元数据
	if len(metadataDBs) > 0 {
		_, err = s.MetadataService.BatchRegisterMetadata(ctx, tx, metadataDBs)
		if err != nil {
			s.Logger.WithContext(ctx).Errorf("insert metadata failed, err: %v", err)
			err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
			return
		}
	}
	// 添加工具
	if len(toolDBs) > 0 {
		_, err = s.ToolDB.InsertTools(ctx, tx, toolDBs)
		if err != nil {
			s.Logger.WithContext(ctx).Errorf("insert tool failed, err: %v", err)
			err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
			return
		}
	}
	return
}

// importCheck 校验导入的工具箱信息
func (s *ToolServiceImpl) importCheck(ctx context.Context, item *interfaces.ToolBoxImpexItem, userID string) (toolDBs []*model.ToolDB,
	metadataList []interfaces.IMetadataDB, err error) {
	// 注入默认值并校验
	err = defaults.Set(item)
	if err != nil {
		s.Logger.WithContext(ctx).Errorf("set default value failed, err: %v", err)
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	err = s.Validator.ValidatorStruct(ctx, item)
	if err != nil {
		return
	}
	// 校验工具箱信息
	err = s.Validator.ValidatorToolBoxName(ctx, item.BoxName)
	if err != nil {
		return
	}
	// 检查desc
	err = s.Validator.ValidatorToolBoxDesc(ctx, item.BoxDesc)
	if err != nil {
		return
	}
	// 检查分类是否存在
	if !s.CategoryManager.CheckCategory(interfaces.BizCategory(item.CategoryType)) {
		// 设置为默认分类
		item.CategoryType = interfaces.CategoryTypeOther.String()
	}
	// 检查是否为内置
	toolDBs = []*model.ToolDB{}
	toolNames := make(map[string]bool)
	for _, toolImpexItem := range item.Tools {
		if _, ok := toolNames[toolImpexItem.Name]; ok {
			err = errors.NewHTTPError(ctx, http.StatusBadRequest, errors.ErrExtToolNameDuplicate,
				fmt.Sprintf("tool name %v duplicate", toolImpexItem.Name), toolImpexItem.Name)
			return
		}
		// 校验工具信息
		err = s.Validator.ValidatorToolName(ctx, toolImpexItem.Name)
		if err != nil {
			return
		}
		if toolImpexItem.Description == "" {
			toolImpexItem.Description = toolImpexItem.Name
		}
		err = s.Validator.ValidatorToolDesc(ctx, toolImpexItem.Description)
		if err != nil {
			return
		}
		toolNames[toolImpexItem.Name] = true
		toolDBs = append(toolDBs, &model.ToolDB{
			ToolID:      toolImpexItem.ToolID,
			BoxID:       item.BoxID,
			Name:        toolImpexItem.Name,
			Description: toolImpexItem.Description,
			SourceID:    toolImpexItem.SourceID,
			SourceType:  toolImpexItem.SourceType,
			Status:      toolImpexItem.Status.String(),
			UseRule:     toolImpexItem.UseRule,
			Parameters:  utils.ObjectToJSON(toolImpexItem.GlobalParameters),
			CreateUser:  userID,
			CreateTime:  time.Now().UnixNano(),
			UpdateUser:  userID,
			UpdateTime:  time.Now().UnixNano(),
			ExtendInfo:  utils.ObjectToJSON(toolImpexItem.ExtendInfo),
		})
		switch toolImpexItem.SourceType {
		case model.SourceTypeOpenAPI:
			if toolImpexItem.Metadata == nil {
				err = errors.DefaultHTTPError(ctx, http.StatusBadRequest, "tool metadata is nil")
				return
			}
			err = s.Validator.ValidatorStruct(ctx, toolImpexItem.Metadata)
			if err != nil {
				return
			}
			if toolImpexItem.MetadataType != "" && toolImpexItem.MetadataType != item.MetadataType {
				err = errors.NewHTTPError(ctx, http.StatusBadRequest, errors.ErrExtToolTypeMismatch,
					fmt.Sprintf("tool type %v mismatch", toolImpexItem.MetadataType))
				return
			}
			metadataDB := &model.APIMetadataDB{
				Version:     toolImpexItem.Metadata.Version,
				CreateUser:  userID,
				CreateTime:  time.Now().UnixNano(),
				UpdateUser:  userID,
				UpdateTime:  time.Now().UnixNano(),
				Summary:     toolImpexItem.Metadata.Summary,
				Description: toolImpexItem.Metadata.Description,
				Path:        toolImpexItem.Metadata.Path,
				ServerURL:   toolImpexItem.Metadata.ServerURL,
				Method:      toolImpexItem.Metadata.Method,
				APISpec:     utils.ObjectToJSON(toolImpexItem.Metadata.APISpec),
			}
			metadataList = append(metadataList, metadataDB)
		case model.SourceTypeFunction:
			if toolImpexItem.Metadata == nil {
				err = errors.DefaultHTTPError(ctx, http.StatusBadRequest, "tool metadata is nil")
				return
			}
			err = s.Validator.ValidatorStruct(ctx, toolImpexItem.Metadata)
			if err != nil {
				return
			}
			if toolImpexItem.MetadataType != "" && toolImpexItem.MetadataType != item.MetadataType {
				err = errors.NewHTTPError(ctx, http.StatusBadRequest, errors.ErrExtToolTypeMismatch,
					fmt.Sprintf("tool type %v mismatch", toolImpexItem.MetadataType))
				return
			}
			metadataDB := &model.FunctionMetadataDB{
				Version:      toolImpexItem.Metadata.Version,
				CreateUser:   userID,
				CreateTime:   time.Now().UnixNano(),
				UpdateUser:   userID,
				UpdateTime:   time.Now().UnixNano(),
				Summary:      toolImpexItem.Metadata.Summary,
				Description:  toolImpexItem.Metadata.Description,
				Path:         toolImpexItem.Metadata.Path,
				ServerURL:    toolImpexItem.Metadata.ServerURL,
				Method:       toolImpexItem.Metadata.Method,
				APISpec:      utils.ObjectToJSON(toolImpexItem.Metadata.APISpec),
				ScriptType:   string(toolImpexItem.ScriptType),
				Dependencies: utils.ObjectToJSON(toolImpexItem.Dependencies),
				Code:         toolImpexItem.Code,
			}
			metadataList = append(metadataList, metadataDB)
		case model.SourceTypeOperator:
		}
	}
	return
}

// 导出预检查
func (s *ToolServiceImpl) exportPreCheck(ctx context.Context, req *interfaces.ExportReq) (boxDBs []*model.ToolboxDB, err error) {
	// 批量鉴权
	var accessor *interfaces.AuthAccessor
	accessor, err = s.AuthService.GetAccessor(ctx, req.UserID)
	if err != nil {
		return
	}
	// 检查查看权限权限
	checkBoxIDs, err := s.AuthService.ResourceFilterIDs(ctx, accessor, req.IDs,
		interfaces.AuthResourceTypeToolBox, interfaces.AuthOperationTypeView)
	if err != nil {
		return
	}
	if len(checkBoxIDs) != len(req.IDs) {
		clist := utils.FindMissingElements(req.IDs, checkBoxIDs)
		err = errors.NewHTTPError(ctx, http.StatusForbidden, errors.ErrExtCommonOperationForbidden,
			fmt.Sprintf("toolbox %v not access", clist))
		return
	}
	// 检查数据是否存在
	boxDBs, err = s.ToolBoxDB.SelectListByBoxIDs(ctx, req.IDs)
	if err != nil {
		s.Logger.WithContext(ctx).Errorf("select toolbox list err: %s", err.Error())
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	if len(boxDBs) != len(req.IDs) {
		checkBoxes := []string{}
		for _, v := range boxDBs {
			checkBoxes = append(checkBoxes, v.BoxID)
		}
		clist := utils.FindMissingElements(req.IDs, checkBoxes)
		err = errors.NewHTTPError(ctx, http.StatusNotFound, errors.ErrExtToolNotFound,
			fmt.Sprintf("toolbox %v not found", clist))
		return
	}
	return
}

// Export 导出
func (s *ToolServiceImpl) Export(ctx context.Context, req *interfaces.ExportReq) (data *interfaces.ComponentImpexConfigModel, err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)

	boxDBs, err := s.exportPreCheck(ctx, req)
	if err != nil {
		return
	}
	// 批量获取工具箱内工具信息
	toolBoxConfig, depOperatorIDs, err := s.batchGetToolBoxInfo(ctx, boxDBs)
	if err != nil {
		return
	}
	data = &interfaces.ComponentImpexConfigModel{
		Toolbox: toolBoxConfig,
	}
	// 批量获取算子依赖信息
	depOperatorIDs = utils.UniqueStrings(depOperatorIDs)
	if len(depOperatorIDs) == 0 {
		return
	}
	operatorImpexConfig, err := s.OperatorMgnt.Export(ctx, &interfaces.ExportReq{
		UserID: req.UserID,
		IDs:    depOperatorIDs,
	})
	if err != nil {
		return
	}
	data.Operator = operatorImpexConfig.Operator
	return
}

// 批量获取工具箱内工具信息
func (s *ToolServiceImpl) batchGetToolBoxInfo(ctx context.Context, boxDBs []*model.ToolboxDB) (toolBoxInfo *interfaces.ToolBoxImpexConfig,
	depOperatorIDs []string, err error) {
	toolsMap := map[string][]*interfaces.ToolImpexItem{} // 工具箱下工具的导出信息
	toolBoxInfo = &interfaces.ToolBoxImpexConfig{
		Configs: []*interfaces.ToolBoxImpexItem{},
	}
	// 组装工具箱信息
	boxIDs := []string{}
	for _, boxDB := range boxDBs {
		if boxDB.IsInternal {
			err = errors.NewHTTPError(ctx, http.StatusForbidden, errors.ErrExtCommonInternalComponentNotAllowed,
				fmt.Sprintf("internal toolbox %v not allowed to export", boxDB.BoxID), boxDB.Name)
			return
		}
		toolBoxInfo.Configs = append(toolBoxInfo.Configs, &interfaces.ToolBoxImpexItem{
			BoxID:        boxDB.BoxID,
			BoxName:      boxDB.Name,
			BoxDesc:      boxDB.Description,
			BoxSvcURL:    boxDB.ServerURL,
			Status:       interfaces.BizStatus(boxDB.Status),
			CategoryType: boxDB.Category,
			CategoryName: s.CategoryManager.GetCategoryName(ctx, interfaces.BizCategory(boxDB.Category)),
			IsInternal:   boxDB.IsInternal,
			Source:       boxDB.Source,
			Tools:        []*interfaces.ToolImpexItem{},
			CreateTime:   boxDB.CreateTime,
			UpdateTime:   boxDB.UpdateTime,
			CreateUser:   boxDB.CreateUser,
			UpdateUser:   boxDB.UpdateUser,
			MetadataType: interfaces.MetadataType(boxDB.MetadataType),
		})
		// 收集工具箱ID并初始化工具映射
		boxIDs = append(boxIDs, boxDB.BoxID)
		toolsMap[boxDB.BoxID] = []*interfaces.ToolImpexItem{}
	}
	// 获取工具箱内的全部工具
	tools, err := s.ToolDB.SelectToolBoxByIDs(ctx, boxIDs)
	if err != nil {
		s.Logger.WithContext(ctx).Errorf("select toolbox by ids:%v, err:%v", boxIDs, err)
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	// 组装工具信息并收集带查询元数据信息
	sourceMap := map[model.SourceType][]string{} // 元数据ID映射
	for _, toolDB := range tools {
		var toolInfo *interfaces.ToolInfo
		toolInfo, err = s.toolDBToToolInfo(ctx, toolDB)
		if err != nil {
			return
		}
		toolImpexItem := &interfaces.ToolImpexItem{
			ToolInfo:   *toolInfo,
			SourceID:   toolDB.SourceID,
			SourceType: toolDB.SourceType,
		}
		switch toolDB.SourceType {
		case model.SourceTypeOpenAPI:
			sourceMap[model.SourceTypeOpenAPI] = append(sourceMap[model.SourceTypeOpenAPI], toolDB.SourceID)
		case model.SourceTypeFunction:
			sourceMap[model.SourceTypeFunction] = append(sourceMap[model.SourceTypeFunction], toolDB.SourceID)
		case model.SourceTypeOperator:
			sourceMap[model.SourceTypeOperator] = append(sourceMap[model.SourceTypeOperator], toolDB.SourceID)
			depOperatorIDs = append(depOperatorIDs, toolDB.SourceID)
		}
		toolsMap[toolDB.BoxID] = append(toolsMap[toolDB.BoxID], toolImpexItem)
	}

	// 批量获取元数据
	sourceIDToMetadataMap, err := s.MetadataService.BatchGetMetadataBySourceIDs(ctx, sourceMap)
	if err != nil {
		return
	}
	// 组装工具元数据信息
	for _, toolBox := range toolBoxInfo.Configs {
		// 获取工具箱内的工具
		for _, toolInfo := range toolsMap[toolBox.BoxID] {
			metadataDB, ok := sourceIDToMetadataMap[toolInfo.SourceID]
			if !ok {
				continue
			}
			toolInfo.MetadataType = interfaces.MetadataType(metadataDB.GetType())
			if toolInfo.SourceType != model.SourceTypeOperator {
				// 算子工具不直接导出元数据, 而是通过算子依赖导出
				toolInfo.Metadata = metadata.MetadataDBToStruct(metadataDB)
				dependencies := []interfaces.DependencyInfo{}
				if metadataDB.GetDependencies() != "" {
					dependencies = utils.JSONToObject[[]interfaces.DependencyInfo](metadataDB.GetDependencies())
				}
				toolInfo.FunctionContent = interfaces.FunctionContent{
					ScriptType:      interfaces.ScriptType(metadataDB.GetScriptType()),
					Code:            metadataDB.GetCode(),
					Dependencies:    dependencies,
					DependenciesURL: metadataDB.GetDependenciesURL(),
				}
			}
			toolBox.Tools = append(toolBox.Tools, toolInfo)
		}
	}
	return
}

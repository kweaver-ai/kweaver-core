package operator

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/creasty/defaults"
	"github.com/google/uuid"
	icommon "github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/metadata"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/metric"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
)

// Export 导出算子
func (m *operatorManager) Export(ctx context.Context, req *interfaces.ExportReq) (data *interfaces.ComponentImpexConfigModel, err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	// 导出预检查
	operatorList, err := m.exportPreCheck(ctx, req)
	if err != nil {
		return
	}
	data = &interfaces.ComponentImpexConfigModel{
		Operator: &interfaces.OperatorImpexConfig{},
	}
	// 导出依赖及追加依赖算子
	allOperatorDBs, compositeConfigs, err := m.getCompositeOperatorDependencies(ctx, operatorList, req.UserID)
	if err != nil {
		return
	}
	data.Operator.CompositeConfigs = compositeConfigs

	// 批量获取算子元数据
	items, err := m.batchGetOperatorInfo(ctx, allOperatorDBs)
	if err != nil {
		return
	}
	data.Operator.Configs = items
	return
}

// Import 导入算子
func (m *operatorManager) Import(ctx context.Context, tx *sql.Tx, mode interfaces.ImportType, data *interfaces.OperatorImpexConfig, userID string) (err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	if data == nil || len(data.Configs) == 0 {
		err = errors.NewHTTPError(ctx, http.StatusBadRequest, errors.ErrExtCommonImportDataEmpty, "operator configs is empty")
		return
	}
	// 导入预备检查
	operatorList, err := m.importPreCheck(ctx, mode, data.Configs)
	if err != nil {
		return
	}
	var accessor *interfaces.AuthAccessor
	if icommon.IsPublicAPIFromCtx(ctx) {
		accessor, err = m.AuthService.GetAccessor(ctx, userID)
		if err != nil {
			return
		}
		}
	// 导入算子元数据
	createMap, updateMap, err := m.batchImportOperatorMetadata(ctx, tx, data.Configs, operatorList, accessor, userID)
	if err != nil {
		return
	}

	err = m.importPostProcess(ctx, createMap, updateMap, accessor)
	return
}

// 后置操作：添加权限配置，及审计日志记录
func (m *operatorManager) importPostProcess(ctx context.Context, createMap, updateMap map[string]*model.OperatorRegisterDB, accessor *interfaces.AuthAccessor) (err error) {
	businessDomainID, _ := icommon.GetBusinessDomainFromCtx(ctx)
	for _, operatorDB := range createMap {
		// 关联业务域
		err = m.BusinessDomainService.AssociateResource(ctx, businessDomainID, operatorDB.OperatorID, interfaces.AuthResourceTypeOperator)
		if err != nil {
			return
		}

		// 触发新建策略，创建人默认拥有对当前资源的所有操作权限（内部调用不创建）
		if accessor != nil {
			err := m.AuthService.CreateOwnerPolicy(ctx, accessor, &interfaces.AuthResource{
				ID:   operatorDB.OperatorID,
				Type: string(interfaces.AuthResourceTypeOperator),
				Name: operatorDB.Name,
				})
			if err != nil {
				m.Logger.WithContext(ctx).Warnf("[importPostProcess] CreateOwnerPolicy err :%v", err)
				}
			}
		// 记录设计日志及后续通知（内部调用不记录）
		if accessor != nil {
			go func() {
				accountAuthContext, ok := icommon.GetAccountAuthContextFromCtx(ctx)
				if !ok {
					m.Logger.WithContext(ctx).Warnf("[importPostProcess] GetAccountAuthContextFromCtx err :%v", err)
					return
					}
				m.AuditLog.Logger(ctx, &metric.AuditLogBuilderParams{
					TokenInfo: accountAuthContext.TokenInfo,
					Accessor:  accessor,
					Operation: metric.AuditLogOperationCreate,
					Object: &metric.AuditLogObject{
						Type: metric.AuditLogObjectOperator,
						ID:   operatorDB.OperatorID,
						Name: operatorDB.Name,
						},
					})
				}()
			}
	}
	// 更新算子
	for _, operatorDB := range updateMap {
		// 通知资源变更
		authResource := &interfaces.AuthResource{
			ID:   operatorDB.OperatorID,
			Name: operatorDB.Name,
			Type: string(interfaces.AuthResourceTypeOperator),
		}
		err := m.AuthService.NotifyResourceChange(ctx, authResource)
		if err != nil {
			m.Logger.WithContext(ctx).Warnf("[importPostProcess] NotifyResourceChange err :%v", err)
		}
		// 记录设计日志及后续通知（内部调用不记录）
		if accessor != nil {
			go func() {
				accountAuthContext, ok := icommon.GetAccountAuthContextFromCtx(ctx)
				if !ok {
					m.Logger.WithContext(ctx).Warnf("[importPostProcess] GetAccountAuthContextFromCtx err :%v", err)
					return
					}
				m.AuditLog.Logger(ctx, &metric.AuditLogBuilderParams{
					TokenInfo: accountAuthContext.TokenInfo,
					Accessor:  accessor,
					Operation: metric.AuditLogOperationEdit,
					Object: &metric.AuditLogObject{
						Type: metric.AuditLogObjectOperator,
						ID:   operatorDB.OperatorID,
						Name: operatorDB.Name,
						},
					})
				}()
			}
	}
	return nil
}

// 导入预备检查
func (m *operatorManager) importPreCheck(ctx context.Context, mode interfaces.ImportType, items []*interfaces.OperatorImpexItem) (operatorList []*model.OperatorRegisterDB, err error) {
	// 获取待导入算子ID列表、name列表
	operatorIDs := make([]string, 0)
	for _, operatorItem := range items {
		operatorIDs = append(operatorIDs, operatorItem.OperatorID)
		// 内置算子不允许导入
		if icommon.IsPublicAPIFromCtx(ctx) && operatorItem.IsInternal {
			err = errors.NewHTTPError(ctx, http.StatusForbidden, errors.ErrExtCommonInternalComponentNotAllowed,
				fmt.Sprintf("internal operator %v not allowed to import", operatorItem.OperatorID), operatorItem.OperatorName)
			return
		}
		// 算子重名校验
		err = m.checkDuplicateName(ctx, operatorItem.OperatorName, operatorItem.OperatorID)
		if err != nil {
			return
		}
	}
	operatorIDs = utils.UniqueStrings(operatorIDs)
	// 检查ID资源是否冲突
	operatorList, err = m.DBOperatorManager.SelectByOperatorIDs(ctx, operatorIDs)
	if err != nil {
		m.Logger.WithContext(ctx).Errorf("select operator list err: %v", err.Error())
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	// 创建模式：如果算子已存在，则返回冲突错误
	if mode == interfaces.ImportTypeCreate && len(operatorList) > 0 {
		err = errors.NewHTTPError(ctx, http.StatusConflict, errors.ErrExtCommonResourceIDConflict, "operator id already exists")
	}
	return
}

// 批量导入算子元数据
func (m *operatorManager) batchImportOperatorMetadata(ctx context.Context, tx *sql.Tx, items []*interfaces.OperatorImpexItem, needUpdateOperatorList []*model.OperatorRegisterDB,
	accessor *interfaces.AuthAccessor, userID string) (createMap, updateMap map[string]*model.OperatorRegisterDB, err error) {
	// 需要新增的算子列表
	createMap = map[string]*model.OperatorRegisterDB{}
	// 需要更新的算子列表
	updateMap = map[string]*model.OperatorRegisterDB{}
	for _, operatorDB := range needUpdateOperatorList {
		// 检查算子编辑权限（内部调用不鉴权）
		if icommon.IsPublicAPIFromCtx(ctx) {
			err = m.AuthService.CheckModifyPermission(ctx, accessor, operatorDB.OperatorID, interfaces.AuthResourceTypeOperator)
			if err != nil {
				return
			}
			// 内置算子不允许更新
			if operatorDB.IsInternal {
				err = errors.NewHTTPError(ctx, http.StatusForbidden, errors.ErrExtCommonInternalComponentNotAllowed,
					fmt.Sprintf("internal operator %v not allowed to update", operatorDB.OperatorID), operatorDB.Name)
				return
			}
		}
		updateMap[operatorDB.OperatorID] = operatorDB
	}
	for _, operatorItem := range items {
		// 参数预备检查
		var newOperatorDB *model.OperatorRegisterDB
		var newMetadataDB interfaces.IMetadataDB
		uid := userID
		if accessor != nil {
			uid = accessor.ID
		}
		newOperatorDB, newMetadataDB, err = m.importCheck(ctx, operatorItem, uid)
		if err != nil {
			return
		}
		operatorDB, ok := updateMap[newOperatorDB.OperatorID]
		if ok { // 更新算子
			err = m.updateOperatorConfig(ctx, tx, operatorDB, newOperatorDB, newMetadataDB)
			if err != nil {
				return
			}
			updateMap[operatorDB.OperatorID] = operatorDB
			if operatorDB.Status == interfaces.BizStatusPublished.String() {
				err = m.publishRelease(ctx, tx, operatorDB, operatorDB.UpdateUser)
			}
		} else { // 新增算子
			err = m.addOperatorConfig(ctx, tx, newOperatorDB, newMetadataDB) // 新增算子
			if err != nil {
				return
			}
			createMap[newOperatorDB.OperatorID] = newOperatorDB
			// 发布算子
			if newOperatorDB.Status == interfaces.BizStatusPublished.String() {
				err = m.publishRelease(ctx, tx, newOperatorDB, newOperatorDB.CreateUser)
			}
		}
		if err != nil {
			return
		}
	}
	return
}

// 添加算子配置
func (m *operatorManager) addOperatorConfig(ctx context.Context, tx *sql.Tx, operatorDB *model.OperatorRegisterDB, metadataDB interfaces.IMetadataDB) (err error) {
	// 检查该版本元数据是否存在，如果存在报错冲突
	exists, _, err := m.MetadataService.CheckMetadataExists(ctx, interfaces.MetadataType(metadataDB.GetType()), metadataDB.GetVersion())
	if err != nil {
		m.Logger.WithContext(ctx).Errorf("check metadata version exists failed, err: %v", err)
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	if exists {
		err = errors.NewHTTPError(ctx, http.StatusConflict, errors.ErrExtCommonResourceIDConflict,
			fmt.Sprintf("metadata version %s already exists", metadataDB.GetVersion()))
		return
	}
	version, err := m.MetadataService.RegisterMetadata(ctx, tx, metadataDB)
	if err != nil {
		m.Logger.WithContext(ctx).Errorf("register metadata failed, err: %v", err)
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	operatorDB.MetadataVersion = version
	_, err = m.DBOperatorManager.InsertOperator(ctx, tx, operatorDB)
	if err != nil {
		m.Logger.WithContext(ctx).Errorf("insert operator failed, err: %v", err)
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	return
}

// 更新（升级）算子配置
func (m *operatorManager) updateOperatorConfig(ctx context.Context, tx *sql.Tx, operatorDB,
	newOperatorDB *model.OperatorRegisterDB, newMetadataDB interfaces.IMetadataDB) (err error) {
	// 检查元数据类型是否一致
	if operatorDB.MetadataType != newOperatorDB.MetadataType {
		err = errors.NewHTTPError(ctx, http.StatusBadRequest, errors.ErrExtCommonMetadataTypeConflict,
			fmt.Sprintf("operator %s metadata type conflict, expect %v, got %v", operatorDB.OperatorID, operatorDB.MetadataType, newOperatorDB.MetadataType))
		return
	}
	operatorDB.Category = newOperatorDB.Category
	operatorDB.Name = newOperatorDB.Name
	operatorDB.OperatorType = newOperatorDB.OperatorType
	operatorDB.ExecutionMode = newOperatorDB.ExecutionMode
	operatorDB.Category = newOperatorDB.Category
	operatorDB.Source = newOperatorDB.Source
	operatorDB.ExecuteControl = newOperatorDB.ExecuteControl
	operatorDB.ExtendInfo = newOperatorDB.ExtendInfo
	operatorDB.UpdateUser = newOperatorDB.CreateUser
	operatorDB.UpdateTime = time.Now().UnixNano()
	operatorDB.MetadataType = newOperatorDB.MetadataType
	switch interfaces.BizStatus(operatorDB.Status) {
	case interfaces.BizStatusPublished, interfaces.BizStatusOffline:
		newMetadataDB.SetVersion(uuid.New().String())
		operatorDB.MetadataVersion, err = m.MetadataService.RegisterMetadata(ctx, tx, newMetadataDB)
	case interfaces.BizStatusUnpublish, interfaces.BizStatusEditing:
		// 检查元数据是否存在
		var metadataDB interfaces.IMetadataDB
		var has bool
		has, metadataDB, err = m.MetadataService.CheckMetadataExists(ctx, interfaces.MetadataType(newOperatorDB.MetadataType), operatorDB.MetadataVersion)
		if err != nil {
			m.Logger.WithContext(ctx).Errorf("select api metadata failed, err: %v", err)
			err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
			return
		}
		if has {
			metadataDB.SetSummary(newMetadataDB.GetSummary())
			metadataDB.SetDescription(newMetadataDB.GetDescription())
			metadataDB.SetScriptType(newMetadataDB.GetScriptType())
			metadataDB.SetServerURL(newMetadataDB.GetServerURL())
			metadataDB.SetPath(newMetadataDB.GetPath())
			metadataDB.SetMethod(newMetadataDB.GetMethod())
			metadataDB.SetCode(newMetadataDB.GetCode())
			metadataDB.SetScriptType(newMetadataDB.GetScriptType())
			metadataDB.SetDependencies(newMetadataDB.GetDependencies())
			metadataDB.SetDependenciesURL(newMetadataDB.GetDependenciesURL())
			metadataDB.SetUpdateInfo(newOperatorDB.CreateUser)
			err = m.MetadataService.UpdateMetadata(ctx, tx, metadataDB)
		} else {
			operatorDB.MetadataVersion, err = m.MetadataService.RegisterMetadata(ctx, tx, newMetadataDB)
		}
	}
	if err != nil {
		m.Logger.WithContext(ctx).Errorf("upsert api metadata failed, err: %v", err)
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	operatorDB.Status = newOperatorDB.Status
	err = m.DBOperatorManager.UpdateByOperatorID(ctx, tx, operatorDB)
	if err != nil {
		m.Logger.WithContext(ctx).Errorf("update operator failed, err: %v", err)
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	return
}

func (m *operatorManager) importCheck(ctx context.Context, item *interfaces.OperatorImpexItem, userID string) (operatorDB *model.OperatorRegisterDB,
	metadataDB interfaces.IMetadataDB, err error) {
	// 校验算子信息
	err = m.Validator.ValidateOperatorName(ctx, item.OperatorName)
	if err != nil {
		return
	}
	item.IsInternal = false
	if item.OperatorInfo == nil {
		item.OperatorInfo = &interfaces.OperatorInfo{}
		err = defaults.Set(item.OperatorInfo)
		if err != nil {
			err = errors.DefaultHTTPError(ctx, http.StatusBadRequest, err.Error())
			return
		}
	}
	if item.OperatorExecuteControl == nil {
		item.OperatorExecuteControl = &interfaces.OperatorExecuteControl{}
		err = defaults.Set(item.OperatorExecuteControl)
		if err != nil {
			err = errors.DefaultHTTPError(ctx, http.StatusBadRequest, err.Error())
			return
		}
	}
	// 如果是数据源算子，只能够是同步算子
	isDataSource, err := checkIsDataSource(ctx, item.OperatorInfo.ExecutionMode, item.OperatorInfo.IsDataSource)
	if err != nil {
		return
	}
	// 检查分类是否存在,不存在设置为默认分类
	if !m.CategoryManager.CheckCategory(item.OperatorInfo.Category) {
		item.OperatorInfo.Category = interfaces.CategoryTypeOther
	}
	// 检查元数据
	if item.Metadata == nil {
		err = errors.DefaultHTTPError(ctx, http.StatusBadRequest, fmt.Sprintf("operator %v metadata is nil", item.OperatorName))
		return
	}
	switch item.MetadataType {
	case interfaces.MetadataTypeAPI:
		metadata := &interfaces.MetadataInfo{
			APISpec: &interfaces.APISpec{},
		}
		err = utils.AnyToObject(item.Metadata, metadata)
		if err != nil {
			err = errors.DefaultHTTPError(ctx, http.StatusBadRequest, err.Error())
			return
		}
		err = m.Validator.ValidatorStruct(ctx, metadata)
		if err != nil {
			return
		}
		metadataDB = &model.APIMetadataDB{
			Version:     metadata.Version,
			CreateUser:  userID,
			CreateTime:  time.Now().UnixNano(),
			UpdateUser:  userID,
			UpdateTime:  time.Now().UnixNano(),
			Summary:     metadata.Summary,
			Description: metadata.Description,
			Path:        metadata.Path,
			ServerURL:   metadata.ServerURL,
			Method:      metadata.Method,
			APISpec:     utils.ObjectToJSON(metadata.APISpec),
		}
	case interfaces.MetadataTypeFunc:
		err = m.Validator.ValidatorStruct(ctx, item.Metadata)
		if err != nil {
			return
		}
		metadataDB = &model.FunctionMetadataDB{
			Version:         item.Metadata.Version,
			CreateUser:      userID,
			CreateTime:      time.Now().UnixNano(),
			UpdateUser:      userID,
			UpdateTime:      time.Now().UnixNano(),
			Summary:         item.Metadata.Summary,
			Description:     item.Metadata.Description,
			Path:            item.Metadata.Path,
			ServerURL:       item.Metadata.ServerURL,
			Method:          item.Metadata.Method,
			APISpec:         utils.ObjectToJSON(item.Metadata.APISpec),
			ScriptType:      string(item.Metadata.FunctionContent.ScriptType),
			Dependencies:    utils.ObjectToJSON(item.Metadata.FunctionContent.Dependencies),
			DependenciesURL: item.Metadata.FunctionContent.DependenciesURL,
			Code:            item.Metadata.FunctionContent.Code,
		}
	default:
		err = errors.DefaultHTTPError(ctx, http.StatusBadRequest, fmt.Sprintf("operator %v metadata type %v is not supported", item.OperatorName, item.MetadataType))
		return
	}
	// 如果算子描述为空，默认使用算子名称
	if metadataDB.GetDescription() == "" {
		metadataDB.SetDescription(metadataDB.GetSummary())
	}
	err = m.Validator.ValidateOperatorDesc(ctx, metadataDB.GetDescription())
	if err != nil {
		return
	}
	operatorDB = &model.OperatorRegisterDB{
		OperatorID:      item.OperatorID,
		Name:            item.OperatorName,
		MetadataType:    string(item.MetadataType),
		MetadataVersion: metadataDB.GetVersion(),
		Status:          item.Status.String(),
		OperatorType:    string(item.OperatorInfo.Type),
		ExecutionMode:   string(item.OperatorInfo.ExecutionMode),
		Category:        string(item.OperatorInfo.Category),
		Source:          item.OperatorInfo.Source,
		ExecuteControl:  utils.ObjectToJSON(item.OperatorExecuteControl),
		ExtendInfo:      utils.ObjectToJSON(item.ExtendInfo),
		CreateUser:      userID,
		CreateTime:      time.Now().UnixNano(),
		UpdateUser:      userID,
		UpdateTime:      time.Now().UnixNano(),
		IsDataSource:    isDataSource,
		IsInternal:      false,
	}
	metadataDB.SetCreateInfo(userID)
	metadataDB.SetUpdateInfo(userID)
	return
}

// 导出预检查
func (m *operatorManager) exportPreCheck(ctx context.Context, req *interfaces.ExportReq) (operatorList []*model.OperatorRegisterDB, err error) {
	// 批量鉴权
	var accessor *interfaces.AuthAccessor
	accessor, err = m.AuthService.GetAccessor(ctx, req.UserID)
	if err != nil {
		return
	}
	// 检查查看权限
	checkOperatorIDs, err := m.AuthService.ResourceFilterIDs(ctx, accessor, req.IDs,
		interfaces.AuthResourceTypeOperator, interfaces.AuthOperationTypeView)
	if err != nil {
		return
	}
	if len(checkOperatorIDs) != len(req.IDs) {
		clist := utils.FindMissingElements(req.IDs, checkOperatorIDs)
		err = errors.NewHTTPError(ctx, http.StatusForbidden, errors.ErrExtCommonOperationForbidden,
			fmt.Sprintf("operator %v not access", clist))
		return
	}
	// 检查算子是否存在
	operatorList, err = m.DBOperatorManager.SelectByOperatorIDs(ctx, req.IDs)
	if err != nil {
		m.Logger.WithContext(ctx).Errorf("select operator list err: %s", err.Error())
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	if len(operatorList) != len(req.IDs) {
		checkOperatorIDs := []string{}
		for _, v := range operatorList {
			checkOperatorIDs = append(checkOperatorIDs, v.OperatorID)
		}
		clist := utils.FindMissingElements(req.IDs, checkOperatorIDs)
		err = errors.NewHTTPError(ctx, http.StatusNotFound, errors.ErrExtOperatorNotFound,
			fmt.Sprintf("operator %v not found", clist))
		return
	}
	return
}

// 拉取组合算子依赖并进行去重
func (m *operatorManager) getCompositeOperatorDependencies(ctx context.Context, operatorDBs []*model.OperatorRegisterDB, userID string) (allOperatorDBs []*model.OperatorRegisterDB,
	compositeConfigs []any, err error) {
	dagIDs := []string{}
	// 收集算子ID
	allOperatorDBs = append(allOperatorDBs, operatorDBs...)
	operatorMap := map[string]bool{}
	for _, operatorDB := range operatorDBs {
		operatorMap[operatorDB.OperatorID] = true
		if operatorDB.OperatorType == string(interfaces.OperatorTypeComposite) && operatorDB.ExtendInfo != "" {
			extendInfo := map[string]interface{}{}
			err = utils.StringToObject(operatorDB.ExtendInfo, &extendInfo)
			if err != nil {
				m.Logger.WithContext(ctx).Errorf("string to object err: %s", err.Error())
				err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
				return
			}
			if extendInfo["dag_id"].(string) != "" {
				dagIDs = append(dagIDs, extendInfo["dag_id"].(string))
			}
		}
	}
	if len(dagIDs) == 0 {
		return
	}
	// 依赖算子ID
	dependencyOperatorIDs := []string{}
	// 请求组合算子依赖
	var resp *interfaces.FlowAutomationExportResp
	resp, err = m.FlowAutomation.Export(ctx, dagIDs)
	if err != nil {
		m.Logger.WithContext(ctx).Errorf("export dag err: %s", err.Error())
		return
	}
	compositeConfigs = append(compositeConfigs, resp.Configs...)
	dependencyOperatorIDs = append(dependencyOperatorIDs, resp.OperatorIDs...)
	// 去重
	dependencyOperatorIDs = utils.UniqueStrings(dependencyOperatorIDs)
	for _, operatorID := range dependencyOperatorIDs {
		if operatorMap[operatorID] {
			dependencyOperatorIDs = utils.RemoveStringFromSlice(dependencyOperatorIDs, operatorID)
			continue
		}
	}
	if len(dependencyOperatorIDs) == 0 {
		return
	}
	operatorList, err := m.exportPreCheck(ctx, &interfaces.ExportReq{
		UserID: userID,
		IDs:    dependencyOperatorIDs,
	})
	if err != nil {
		return
	}
	// 去重
	for _, operatorDB := range operatorList {
		if !operatorMap[operatorDB.OperatorID] {
			allOperatorDBs = append(allOperatorDBs, operatorDB)
			continue
		}
	}
	return
}

// batchGetOperatorInfo 批量获取算子信息
func (m *operatorManager) batchGetOperatorInfo(ctx context.Context, operatorDBs []*model.OperatorRegisterDB) (items []*interfaces.OperatorImpexItem, err error) {
	items = []*interfaces.OperatorImpexItem{}
	// 收集组合算子流程ID
	sourceMap := map[model.SourceType][]string{}
	for _, v := range operatorDBs {
		if v.IsInternal {
			err = errors.NewHTTPError(ctx, http.StatusForbidden, errors.ErrExtCommonInternalComponentNotAllowed,
				fmt.Sprintf("operator %v not allowed to export", v.OperatorID), v.Name)
			return
		}
		extendInfo := map[string]interface{}{}
		err = utils.StringToObject(v.ExtendInfo, &extendInfo)
		if err != nil {
			m.Logger.WithContext(ctx).Errorf("string to object err: %s", err.Error())
			err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
			return
		}
		executeControl := &interfaces.OperatorExecuteControl{}
		err = utils.StringToObject(v.ExecuteControl, &executeControl)
		if err != nil {
			m.Logger.WithContext(ctx).Errorf("string to object err: %s", err.Error())
			err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, &interfaces.OperatorImpexItem{
			OperatorID:   v.OperatorID,
			OperatorName: v.Name,
			Version:      v.MetadataVersion,
			Status:       interfaces.BizStatus(v.Status),
			MetadataType: interfaces.MetadataType(v.MetadataType),
			ExtendInfo:   extendInfo,
			OperatorInfo: &interfaces.OperatorInfo{
				Type:          interfaces.OperatorType(v.OperatorType),
				ExecutionMode: interfaces.ExecutionMode(v.ExecutionMode),
				Category:      interfaces.BizCategory(v.Category),
				Source:        v.Source,
				IsDataSource:  &v.IsDataSource,
			},
			OperatorExecuteControl: executeControl,
			CreateUser:             v.CreateUser,
			CreateTime:             v.CreateTime,
			UpdateUser:             v.UpdateUser,
			UpdateTime:             v.UpdateTime,
			IsInternal:             v.IsInternal,
		})
		switch v.MetadataType {
		case string(interfaces.MetadataTypeAPI):
			sourceMap[model.SourceTypeOpenAPI] = append(sourceMap[model.SourceTypeOpenAPI], v.MetadataVersion)
		case string(interfaces.MetadataTypeFunc):
			sourceMap[model.SourceTypeFunction] = append(sourceMap[model.SourceTypeFunction], v.MetadataVersion)
		}
	}
	// 收集metadata信息
	sourceIDToMetadataMap, err := m.MetadataService.BatchGetMetadataBySourceIDs(ctx, sourceMap)
	if err != nil {
		m.Logger.WithContext(ctx).Errorf("batch get metadata err: %s", err.Error())
		return
	}
	// 填充metadata信息
	for _, item := range items {
		item.Metadata = metadata.MetadataDBToStruct(sourceIDToMetadataMap[item.Version])
	}
	return
}

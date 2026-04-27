package operator

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common"
	icommon "github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/metric"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
)

// RegisterInternalOperator 注册内置算子
func (m *operatorManager) RegisterInternalOperator(ctx context.Context, req *interfaces.RegisterInternalOperatorReq) (resp *interfaces.OperatorRegisterResp, err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	if !req.IsPublic && req.UserID == "" {
		req.UserID = interfaces.SystemUser
	}
	// 校验传入参数
	err = m.Validator.ValidateOperatorName(ctx, req.Name)
	if err != nil {
		return
	}
	isDataSource, err := checkIsDataSource(ctx, req.ExecutionMode, req.IsDataSource)
	if err != nil {
		m.Logger.WithContext(ctx).Warnf("check is data source failed, err: %v", err.Error())
		return
	}
	// 解析元数据
	var metadataList []interfaces.IMetadataDB
	switch req.MetadataType {
	case interfaces.MetadataTypeAPI:
		metadataList, err = m.MetadataService.ParseMetadata(ctx, req.MetadataType, req.OpenAPIInput)
		if err != nil {
			m.Logger.WithContext(ctx).Errorf("parse metadata failed, err: %v", err)
			return
		}
	case interfaces.MetadataTypeFunc:
		var metadatas []interfaces.IMetadataDB
		for _, funcInput := range req.Functions {
			metadatas, err = m.MetadataService.ParseMetadata(ctx, req.MetadataType, funcInput)
			if err != nil {
				m.Logger.WithContext(ctx).Errorf("parse metadata failed, err: %v", err)
				return
			}
			metadataList = append(metadataList, metadatas...)
		}
	}
	// 内置算子仅能注册一个算子
	if len(metadataList) != 1 {
		err = errors.NewHTTPError(ctx, http.StatusBadRequest, errors.ErrExtOperatorRegisterFailed, "internal operator can only register one operator")
		return
	}
	metadataDB := metadataList[0]
	err = m.validateOperator(ctx, metadataDB)
	if err != nil {
		return
	}
	// metdataParam := &interfaces.APIMetadataEdit{
	// 	Summary:     item.Summary,
	// 	Path:        item.Path,
	// 	Method:      item.Method,
	// 	Description: item.Description,
	// 	APISpec:     item.APISpec,
	// 	ServerURL:   item.ServerURL,
	// }
	checkConfig := &interfaces.IntCompConfig{
		ComponentType: interfaces.ComponentTypeOperator,
		ComponentID:   req.OperatorID,
		ConfigVersion: req.ConfigVersion,
		ConfigSource:  req.ConfigSource,
		ProtectedFlag: req.ProtectedFlag,
	}
	// 检查算子是否已经存在
	has, operatorDB, err := m.DBOperatorManager.SelectByOperatorID(ctx, nil, req.OperatorID)
	if err != nil {
		m.Logger.WithContext(ctx).Errorf("select operator by operatorID failed, err: %v", err.Error())
		err = errors.NewHTTPError(ctx, http.StatusBadRequest, errors.ErrExtOperatorRegisterFailed,
			fmt.Sprintf("select operator by operatorID failed, err: %v", err.Error()))
		return
	}
	tx, err := m.DBTx.GetTx(ctx)
	if err != nil {
		m.Logger.WithContext(ctx).Errorf("get tx failed, err: %v", err.Error())
		err = errors.NewHTTPError(ctx, http.StatusInternalServerError, errors.ErrExtOperatorRegisterFailed,
			fmt.Sprintf("get tx failed, err: %v", err.Error()))
		return
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()
	if !has {
		operatorDB = &model.OperatorRegisterDB{
			OperatorID:     req.OperatorID,
			Name:           req.Name,
			MetadataType:   string(req.MetadataType),
			Status:         string(interfaces.BizStatusPublished),
			OperatorType:   string(req.OperatorType),
			ExecutionMode:  string(req.ExecutionMode),
			Category:       interfaces.CategoryTypeSystem.String(),
			Source:         req.Source,
			IsInternal:     true,
			ExecuteControl: utils.ObjectToJSON(req.OperatorExecuteControl),
			ExtendInfo:     utils.ObjectToJSON(req.ExtendInfo),
			CreateUser:     req.UserID,
			UpdateUser:     req.UserID,
			CreateTime:     time.Now().UnixNano(),
			UpdateTime:     time.Now().UnixNano(),
			IsDataSource:   isDataSource,
		}
		resp, err = m.createInternalOperator(ctx, tx, operatorDB, metadataDB, checkConfig, req.UserID, req.BusinessDomainID)
	} else {
		// 检查来源是否一致
		if !operatorDB.IsInternal || operatorDB.Source != req.Source {
			// 算子已经存在
			err = errors.NewHTTPError(ctx, http.StatusBadRequest, errors.ErrExtOperatorExists,
				"operator not internal or source not match")
			return
		}
		operatorDB.MetadataType = string(req.MetadataType)
		operatorDB.OperatorType = string(req.OperatorType)
		operatorDB.ExecutionMode = string(req.ExecutionMode)
		operatorDB.Source = req.Source
		operatorDB.ExecuteControl = utils.ObjectToJSON(req.OperatorExecuteControl)
		operatorDB.ExtendInfo = utils.ObjectToJSON(req.ExtendInfo)
		operatorDB.UpdateUser = req.UserID
		operatorDB.UpdateTime = time.Now().UnixNano()
		operatorDB.IsDataSource = isDataSource
		resp, err = m.upgradeInternalOperator(ctx, tx, operatorDB, metadataDB, checkConfig, req.Name, req.UserID)
	}
	if err != nil {
		return
	}
	// 发布内置算子
	err = m.publishRelease(ctx, tx, operatorDB, req.UserID)
	return
}

// 创建内置算子
func (m *operatorManager) createInternalOperator(ctx context.Context, tx *sql.Tx, operatorDB *model.OperatorRegisterDB,
	metadataDB interfaces.IMetadataDB, config *interfaces.IntCompConfig, userID, businessDomainId string) (resp *interfaces.OperatorRegisterResp, err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	var operatorID string
	if common.IsPublicAPIFromCtx(ctx) { // 检查是否有新建权限
		var accessor *interfaces.AuthAccessor
		accessor, err = m.AuthService.GetAccessor(ctx, userID)
		if err != nil {
			return
		}
		err = m.AuthService.CheckCreatePermission(ctx, accessor, interfaces.AuthResourceTypeOperator)
		if err != nil {
			return
		}
		defer func() {
			if err == nil {
				// 默认添加所有者权限
				err = m.AuthService.CreateOwnerPolicy(ctx, accessor, &interfaces.AuthResource{
					ID:   operatorID,
					Type: interfaces.AuthResourceTypeOperator.String(),
					Name: operatorDB.Name,
				})
				// 记录审计日志
				go func() {
					accountAuthContext, ok := icommon.GetAccountAuthContextFromCtx(ctx)
					if !ok {
						m.Logger.WithContext(ctx).Warnf("[createInternalOperator] GetAccountAuthContextFromCtx err :%v", err)
						return
					}
					m.AuditLog.Logger(ctx, &metric.AuditLogBuilderParams{
						TokenInfo: accountAuthContext.TokenInfo,
						Accessor:  accessor,
						Operation: metric.AuditLogOperationCreate,
						Object: &metric.AuditLogObject{
							Type: metric.AuditLogObjectOperator,
							Name: operatorDB.Name,
							ID:   operatorID,
						},
					})
				}()
			}
		}()
	}
	// 检查算子是否存在
	err = m.checkDuplicateName(ctx, operatorDB.Name, operatorDB.OperatorID)
	if err != nil {
		return
	}
	// 添加内置算子元数据
	metadataDB.SetCreateInfo(userID)
	metadataDB.SetUpdateInfo(userID)
	version, err := m.MetadataService.RegisterMetadata(ctx, tx, metadataDB)
	if err != nil {
		m.Logger.WithContext(ctx).Errorf("insert api metadata failed, err: %v", err.Error())
		err = errors.NewHTTPError(ctx, http.StatusInternalServerError, errors.ErrExtOperatorRegisterFailed,
			fmt.Sprintf("insert api metadata failed, err: %v", err.Error()))
		return
	}
	// 添加内置算子
	operatorDB.MetadataVersion = version
	operatorID, err = m.DBOperatorManager.InsertOperator(ctx, tx, operatorDB)
	if err != nil {
		m.Logger.WithContext(ctx).Errorf("insert operator failed, err: %v", err.Error())
		err = errors.NewHTTPError(ctx, http.StatusInternalServerError, errors.ErrExtOperatorRegisterFailed,
			fmt.Sprintf("insert operator failed, err: %v", err.Error()))
		return
	}
	// 添加配置
	err = m.IntCompConfigSvc.UpdateConfig(ctx, tx, config)
	if err != nil {
		return
	}

	// 关联业务域
	err = m.BusinessDomainService.AssociateResource(ctx, businessDomainId, operatorID, interfaces.AuthResourceTypeOperator)
	if err != nil {
		return
	}

	// 创建内置组件权限策略
	err = m.AuthService.CreateIntCompPolicyForAllUsers(ctx, &interfaces.AuthResource{
		ID:   operatorID,
		Type: interfaces.AuthResourceTypeOperator.String(),
		Name: operatorDB.Name,
	})
	if err != nil {
		return
	}

	resp = &interfaces.OperatorRegisterResp{
		OperatorID: operatorID,
		Version:    version,
		Status:     interfaces.ResultStatusSuccess,
	}
	return
}

// 升级内置算子
func (m *operatorManager) upgradeInternalOperator(ctx context.Context, tx *sql.Tx, operatorDB *model.OperatorRegisterDB,
	metadataDB interfaces.IMetadataDB, config *interfaces.IntCompConfig, name, userID string) (resp *interfaces.OperatorRegisterResp, err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	var action interfaces.IntCompConfigAction
	// 检查是否有编辑权限
	if common.IsPublicAPIFromCtx(ctx) {
		var accessor *interfaces.AuthAccessor
		accessor, err = m.AuthService.GetAccessor(ctx, userID)
		if err != nil {
			return
		}
		err = m.AuthService.CheckModifyPermission(ctx, accessor, operatorDB.OperatorID, interfaces.AuthResourceTypeOperator)
		if err != nil {
			return
		}
		defer func() {
			if err != nil || action == interfaces.IntCompConfigActionTypeSkip {
				return
			}
			// 记录审计日志
			go func() {
				accountAuthContext, ok := icommon.GetAccountAuthContextFromCtx(ctx)
				if !ok {
					m.Logger.WithContext(ctx).Warnf("[upgradeInternalOperator] GetAccountAuthContextFromCtx err :%v", err)
					return
				}
				m.AuditLog.Logger(ctx, &metric.AuditLogBuilderParams{
					TokenInfo: accountAuthContext.TokenInfo,
					Accessor:  accessor,
					Operation: metric.AuditLogOperationEdit,
					Object: &metric.AuditLogObject{
						Type: metric.AuditLogObjectOperator,
						Name: operatorDB.Name,
						ID:   operatorDB.OperatorID,
					},
				})
			}()
		}()
	}

	action, err = m.IntCompConfigSvc.CompareConfig(ctx, config)
	if err != nil {
		return
	}
	resp = &interfaces.OperatorRegisterResp{
		OperatorID: operatorDB.OperatorID,
		Version:    operatorDB.MetadataVersion,
		Status:     interfaces.ResultStatusSuccess,
	}
	if action == interfaces.IntCompConfigActionTypeSkip { // 无变化，无需更新
		return
	}
	var isNameChange bool
	defer func() {
		if err == nil && isNameChange {
			// 修改名称时，更新名称
			err = m.AuthService.NotifyResourceChange(ctx, &interfaces.AuthResource{
				ID:   operatorDB.OperatorID,
				Type: interfaces.AuthResourceTypeOperator.String(),
				Name: operatorDB.Name,
			})
		}
	}()
	if operatorDB.Name != name {
		// 更新配置
		err = m.checkDuplicateName(ctx, name, operatorDB.OperatorID)
		if err != nil {
			return
		}
		operatorDB.Name = name
		isNameChange = true
	}
	// 查询元数据
	has, currentMetadataDB, err := m.MetadataService.CheckMetadataExists(ctx, interfaces.MetadataType(operatorDB.MetadataType), operatorDB.MetadataVersion)
	if err != nil {
		m.Logger.WithContext(ctx).Errorf("select api metadata by version failed, err: %v", err.Error())
		err = errors.NewHTTPError(ctx, http.StatusInternalServerError, errors.ErrExtOperatorRegisterFailed,
			fmt.Sprintf("select api metadata by version failed, err: %v", err.Error()))
		return
	}
	if !has {
		currentMetadataDB = metadataDB
		currentMetadataDB.SetCreateInfo(userID)
	}
	currentMetadataDB.SetUpdateInfo(userID)
	currentMetadataDB.SetSummary(metadataDB.GetSummary())
	currentMetadataDB.SetDescription(metadataDB.GetDescription())
	currentMetadataDB.SetMethod(metadataDB.GetMethod())
	currentMetadataDB.SetPath(metadataDB.GetPath())
	currentMetadataDB.SetScriptType(metadataDB.GetScriptType())
	currentMetadataDB.SetServerURL(metadataDB.GetServerURL())
	currentMetadataDB.SetAPISpec(metadataDB.GetAPISpec())
	currentMetadataDB.SetCode(metadataDB.GetCode())
	currentMetadataDB.SetScriptType(metadataDB.GetScriptType())
	currentMetadataDB.SetDependencies(metadataDB.GetDependencies())
	currentMetadataDB.SetDependenciesURL(metadataDB.GetDependenciesURL())
	currentMetadataDB.SetVersion(uuid.New().String())
	// 升级元数据，并更新版本号
	version, err := m.MetadataService.RegisterMetadata(ctx, tx, currentMetadataDB)
	if err != nil {
		m.Logger.WithContext(ctx).Errorf("register metadata failed, err: %v", err.Error())
		return
	}
	resp.Version = version
	// 更新算子
	operatorDB.MetadataVersion = version
	operatorDB.UpdateUser = userID
	operatorDB.UpdateTime = time.Now().UnixNano()
	err = m.DBOperatorManager.UpdateByOperatorID(ctx, tx, operatorDB)
	if err != nil {
		m.Logger.WithContext(ctx).Errorf("update operator failed, err: %v", err.Error())
		err = errors.NewHTTPError(ctx, http.StatusInternalServerError, errors.ErrExtOperatorRegisterFailed,
			fmt.Sprintf("update operator failed, err: %v", err.Error()))
		return
	}
	// 更新配置
	err = m.IntCompConfigSvc.UpdateConfig(ctx, tx, config)
	return
}

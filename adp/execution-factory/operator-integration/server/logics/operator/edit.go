package operator

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	icommon "github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common"
	infraerrors "github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/telemetry"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/common"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/metric"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
)

// EditOperator 编辑算子（仅支持编辑当前版本）
func (m *operatorManager) EditOperator(ctx context.Context, req *interfaces.OperatorEditReq) (resp *interfaces.OperatorEditResp, err error) {
	// 记录可观测性
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	telemetry.SetSpanAttributes(ctx, map[string]interface{}{
		"operator_id": req.OperatorID,
		"user_id":     req.UserID,
	})
	// 校验数据的合法性
	operator, metadataDB, accessor, needUpdateMetadata, err := m.preCheckEdit(ctx, req, false)
	if err != nil {
		m.Logger.WithContext(ctx).Warnf("pre check edit failed, err: %v", err)
		return
	}
	var isDataSource bool
	if req.OperatorInfoEdit != nil {
		isDataSource, err = checkIsDataSource(ctx, req.OperatorInfoEdit.ExecutionMode, req.OperatorInfoEdit.IsDataSource)
		if err != nil {
			m.Logger.WithContext(ctx).Warnf("check is data source failed, err: %v", err)
			return
		}
	}
	resp, err = m.editOperator(ctx, req, operator, metadataDB, needUpdateMetadata, false, isDataSource)
	if err != nil {
		m.Logger.WithContext(ctx).Errorf("edit operator failed, err: %v", err)
		return
	}
	// 异步记录审计日志
	go func() {
		accountAuthCtx, ok := icommon.GetAccountAuthContextFromCtx(ctx)
		if !ok {
			m.Logger.WithContext(ctx).Errorf("account auth context not found")
			return
		}
		m.AuditLog.Logger(ctx, &metric.AuditLogBuilderParams{
			TokenInfo: accountAuthCtx.TokenInfo,
			Accessor:  accessor,
			Operation: metric.AuditLogOperationEdit,
			Object: &metric.AuditLogObject{
				Type: metric.AuditLogObjectOperator,
				ID:   operator.OperatorID,
				Name: operator.Name,
			},
		})
	}()
	return resp, nil
}

// editOperator
func (m *operatorManager) editOperator(ctx context.Context, req *interfaces.OperatorEditReq, operator *model.OperatorRegisterDB,
	metadataDB interfaces.IMetadataDB, needUpdateMetadata, directPublish, isDataSource bool) (resp *interfaces.OperatorEditResp, err error) {
	// 判断名字是否变更
	var nameChanged bool
	if req.Name != "" && req.Name != operator.Name {
		// TODO: 检查名字是否重名
		nameChanged = true
	}
	tx, err := m.DBTx.GetTx(ctx)
	if err != nil {
		m.Logger.WithContext(ctx).Warnf("get tx failed, OperatorID: %s, Version: %s, err: %v", operator.OperatorID, operator.MetadataVersion, err)
		err = infraerrors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()
	switch interfaces.BizStatus(operator.Status) {
	case interfaces.BizStatusUnpublish, interfaces.BizStatusEditing:
		if directPublish {
			operator.Status = string(interfaces.BizStatusPublished)
		}
		err = m.modifyOperatorInfo(ctx, tx, req, operator, metadataDB, needUpdateMetadata, isDataSource)
	case interfaces.BizStatusPublished:
		operator.Status = string(interfaces.BizStatusEditing)
		if directPublish {
			operator.Status = string(interfaces.BizStatusPublished)
		}
		err = m.upgradeOperatorInfo(ctx, tx, req, operator, metadataDB, needUpdateMetadata, isDataSource)
	case interfaces.BizStatusOffline:
		operator.Status = string(interfaces.BizStatusUnpublish)
		if directPublish {
			operator.Status = string(interfaces.BizStatusPublished)
		}
		err = m.upgradeOperatorInfo(ctx, tx, req, operator, metadataDB, needUpdateMetadata, isDataSource)
	default: // 无效状态
		err = infraerrors.NewHTTPError(ctx, http.StatusBadRequest, infraerrors.ErrExtOperatorUnSupportEdit, "invalid operator status")
	}
	if err != nil {
		return
	}
	if operator.Status == interfaces.BizStatusPublished.String() {
		err = m.publishRelease(ctx, tx, operator, req.UserID)
		if err != nil {
			return
		}
	}
	if nameChanged {
		// 名字变更，通知所有订阅者
		err = m.AuthService.NotifyResourceChange(ctx, &interfaces.AuthResource{
			Type: interfaces.AuthResourceTypeOperator.String(),
			ID:   operator.OperatorID,
			Name: operator.Name,
		})
		if err != nil {
			return
		}
	}
	// 检查名字是否变更，如果变更需要检查是否重名
	resp = &interfaces.OperatorEditResp{
		Status:     interfaces.BizStatus(operator.Status),
		OperatorID: operator.OperatorID,
		Version:    operator.MetadataVersion,
	}
	return
}

// UpdateOperatorStatus 更新算子状态
func (m *operatorManager) UpdateOperatorStatus(ctx context.Context, req *interfaces.OperatorStatusUpdateReq, userID string) (err error) {
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	// 获取事务
	tx, err := m.DBTx.GetTx(ctx)
	if err != nil {
		m.Logger.WithContext(ctx).Warnf("get tx failed, err: %v", err)
		err = infraerrors.DefaultHTTPError(ctx, http.StatusInternalServerError, "get tx failed")
		return
	}
	defer func() {
		if err != nil {
			e := tx.Rollback()
			if e != nil {
				m.Logger.Errorf("rollback failed, err: %v", e)
			}
		} else {
			e := tx.Commit()
			if e != nil {
				m.Logger.Errorf("commit failed, err: %v", e)
			}
		}
	}()
	// 更新算子状态
	for _, item := range req.StatusItems {
		err = m.updateSinglOperatorStatus(ctx, tx, item, userID)
		if err != nil {
			return
		}
	}
	return
}

// updateSinglOperatorStatus 更新单个算子状态
func (m *operatorManager) updateSinglOperatorStatus(ctx context.Context, tx *sql.Tx, itemReq *interfaces.OperatorStatusItem, userID string) (err error) {
	var has bool
	var operator *model.OperatorRegisterDB
	// 获取算子
	has, operator, err = m.DBOperatorManager.SelectByOperatorID(ctx, tx, itemReq.OperatorID)
	if err != nil {
		m.Logger.WithContext(ctx).Warnf("select operator failed, OperatorID: %s, err: %v", itemReq.OperatorID, err)
		err = infraerrors.DefaultHTTPError(ctx, http.StatusInternalServerError, "select operator failed")
		return err
	}
	if !has {
		// 算子不存在
		err = infraerrors.DefaultHTTPError(ctx, http.StatusNotFound, "operator not found")
		return err
	}
	// 验证并执行状态转换
	if !common.CheckStatusTransition(interfaces.BizStatus(operator.Status), itemReq.Status) {
		err = infraerrors.NewHTTPError(ctx, http.StatusBadRequest, infraerrors.ErrExtOperatorStatusInvalid,
			fmt.Sprintf("invalid status transition from %s to %s", operator.Status, itemReq.Status.String()))
		return
	}
	operator.Status = itemReq.Status.String()
	accessor, err := m.AuthService.GetAccessor(ctx, userID)
	if err != nil {
		return
	}
	// 根据状态处理变更操作
	var operation metric.AuditLogOperationType
	switch interfaces.BizStatus(operator.Status) {
	case interfaces.BizStatusPublished:
		operation = metric.AuditLogOperationPublish
		// 检查发布权限
		err = m.AuthService.CheckPublishPermission(ctx, accessor, operator.OperatorID, interfaces.AuthResourceTypeOperator)
		if err != nil {
			return
		}
		// 检查是否重名
		err = m.checkDuplicateName(ctx, operator.Name, operator.OperatorID)
		if err != nil {
			return
		}
		// 更新配置
		err = m.DBOperatorManager.UpdateOperatorStatus(ctx, tx, operator, userID)
		if err != nil {
			m.Logger.WithContext(ctx).Errorf("update operator status failed, err: %v")
			return infraerrors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		}
		err = m.publishRelease(ctx, tx, operator, userID)
	case interfaces.BizStatusUnpublish, interfaces.BizStatusEditing:
		// 检查编辑权限
		err = m.AuthService.CheckModifyPermission(ctx, accessor, operator.OperatorID, interfaces.AuthResourceTypeOperator)
		if err != nil {
			return
		}
		// 仅更新状态
		err = m.DBOperatorManager.UpdateOperatorStatus(ctx, tx, operator, userID)
		if err != nil {
			m.Logger.WithContext(ctx).Errorf("update operator status failed, err: %v")
			err = infraerrors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		}
	case interfaces.BizStatusOffline:
		operation = metric.AuditLogOperationUnpublish
		// 检查下架权限
		err = m.AuthService.CheckUnpublishPermission(ctx, accessor, operator.OperatorID, interfaces.AuthResourceTypeOperator)
		if err != nil {
			return
		}
		// 更新配置
		err = m.DBOperatorManager.UpdateOperatorStatus(ctx, tx, operator, userID)
		if err != nil {
			m.Logger.WithContext(ctx).Errorf("update operator status failed, err: %v")
			return infraerrors.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		}
		// 下架
		err = m.unpublishRelease(ctx, tx, operator, userID)
	default:
		err = infraerrors.NewHTTPError(ctx, http.StatusBadRequest, infraerrors.ErrExtOperatorStatusInvalid, "invalid operator status")
	}
	if err != nil {
		return
	}
	if operation == "" {
		return
	}
	// 异步记录审计日志
	go func() {
		accountAuthContext, ok := icommon.GetAccountAuthContextFromCtx(ctx)
		if !ok {
			m.Logger.WithContext(ctx).Errorf("get account auth context failed")
			return
		}
		m.AuditLog.Logger(ctx, &metric.AuditLogBuilderParams{
			TokenInfo: accountAuthContext.TokenInfo,
			Accessor:  accessor,
			Operation: operation,
			Object: &metric.AuditLogObject{
				Type: metric.AuditLogObjectOperator,
				ID:   operator.OperatorID,
				Name: operator.Name,
			},
		})
	}()
	return
}

// checkDuplicateName 检查是否重名
func (m *operatorManager) checkDuplicateName(ctx context.Context, name, operatorID string) (err error) {
	has, operatorDB, err := m.DBOperatorManager.SelectByNameAndStatus(ctx, nil, name, interfaces.BizStatusPublished.String())
	if err != nil {
		m.Logger.WithContext(ctx).Warnf("select operator by name failed, err: %v", err)
		err = infraerrors.DefaultHTTPError(ctx, http.StatusInternalServerError, "select operator by name failed")
		return
	}
	if !has || (operatorID != "" && operatorDB.OperatorID == operatorID) {
		return
	}
	err = infraerrors.NewHTTPError(ctx, http.StatusConflict, infraerrors.ErrExtOperatorExistsSameName,
		"operator name already exists, please use a different name", name)
	return
}

// 编辑前置检查:校验编辑请求的合法性: 检查数据是否存在、是否合法、是否有权限修改，并返回查询信息
func (m *operatorManager) preCheckEdit(ctx context.Context, req *interfaces.OperatorEditReq, directPublish bool) (operatorDB *model.OperatorRegisterDB,
	metadataDB interfaces.IMetadataDB, accessor *interfaces.AuthAccessor, needUpdateMetadata bool, err error) {
	// 获取算子
	var has bool
	has, operatorDB, err = m.DBOperatorManager.SelectByOperatorID(ctx, nil, req.OperatorID)
	if err != nil {
		m.Logger.WithContext(ctx).Errorf("select operator failed, OperatorID: %s, err: %v", req.OperatorID, err)
		err = infraerrors.DefaultHTTPError(ctx, http.StatusInternalServerError, "select operator failed")
		return
	}
	if !has {
		// 算子不存在
		err = infraerrors.DefaultHTTPError(ctx, http.StatusNotFound, "operator not found")
		return
	}
	// 检查参数合法性
	if req.Name != "" {
		err = m.Validator.ValidateOperatorName(ctx, req.Name)
		if err != nil {
			return
		}
	}
	if req.Description != "" {
		err = m.Validator.ValidateOperatorDesc(ctx, req.Description)
		if err != nil {
			return
		}
	}
	// TODO：理论上系统算子需要增加校验，系统算子发布后不允许编辑(例如，只有系统管理员可以编辑系统算子)
	accessor, err = m.AuthService.GetAccessor(ctx, req.UserID)
	if err != nil {
		return
	}
	if directPublish {
		err = m.AuthService.MultiCheckOperationPermission(ctx, accessor, req.OperatorID, interfaces.AuthResourceTypeOperator,
			interfaces.AuthOperationTypeModify, interfaces.AuthOperationTypePublish)
	} else {
		// 检查是否有编辑权限
		err = m.AuthService.CheckModifyPermission(ctx, accessor, req.OperatorID, interfaces.AuthResourceTypeOperator)
	}
	if err != nil {
		return
	}
	// 根据version获取元数据
	exists, metadataDB, err := m.MetadataService.CheckMetadataExists(ctx, interfaces.MetadataType(operatorDB.MetadataType), operatorDB.MetadataVersion)
	if err != nil {
		m.Logger.WithContext(ctx).Warnf("select api metadata failed, OperatorID: %s, Version: %s, err: %v", operatorDB.OperatorID, operatorDB.MetadataVersion, err)
		return
	}
	if !exists {
		// 如元数据不存在
		err = infraerrors.NewHTTPError(ctx, http.StatusNotFound, infraerrors.ErrExtMetadataNotFound, map[string]any{
			"operator_id":      req.OperatorID,
			"metadata_type":    req.MetadataType,
			"metadata_version": operatorDB.MetadataVersion,
			"error":            "metadata not found",
		})
		return
	}
	var updateMetadataDB interfaces.IMetadataDB
	updateMetadataDB, err = m.getUpdateMetadataDB(ctx, req, operatorDB, metadataDB)
	if err != nil { // 不需要更新元数据
		return
	}
	var desc string
	if updateMetadataDB != nil {
		if req.MetadataType == interfaces.MetadataTypeFunc {
			// 更新的函数内容是否有变化
			if updateMetadataDB.GetScriptType() != metadataDB.GetScriptType() {
				needUpdateMetadata = true
				metadataDB.SetScriptType(updateMetadataDB.GetScriptType())
			}
			if updateMetadataDB.GetCode() != metadataDB.GetCode() {
				needUpdateMetadata = true
				metadataDB.SetCode(updateMetadataDB.GetCode())
			}
			if len(updateMetadataDB.GetDependencies()) > 0 || len(updateMetadataDB.GetDependencies()) != len(metadataDB.GetDependencies()) {
				needUpdateMetadata = true
				metadataDB.SetDependencies(updateMetadataDB.GetDependencies())
			}
			if updateMetadataDB.GetDependenciesURL() != metadataDB.GetDependenciesURL() {
				needUpdateMetadata = true
				metadataDB.SetDependenciesURL(updateMetadataDB.GetDependenciesURL())
			}
		}
		if metadataDB.GetServerURL() != updateMetadataDB.GetServerURL() {
			metadataDB.SetServerURL(updateMetadataDB.GetServerURL())
			needUpdateMetadata = true
		}
		if metadataDB.GetSummary() != updateMetadataDB.GetSummary() {
			err = m.Validator.ValidateOperatorName(ctx, updateMetadataDB.GetSummary())
			if err != nil {
				return
			}
			metadataDB.SetSummary(updateMetadataDB.GetSummary())
			needUpdateMetadata = true
		}
		if updateMetadataDB.GetAPISpec() != "" {
			metadataDB.SetAPISpec(updateMetadataDB.GetAPISpec())
			needUpdateMetadata = true
		}
		desc = updateMetadataDB.GetDescription()
	}
	if req.Description != "" {
		desc = req.Description
	}
	if metadataDB.GetDescription() != desc {
		err = m.Validator.ValidateOperatorDesc(ctx, desc)
		if err != nil {
			return
		}
		metadataDB.SetDescription(desc)
		needUpdateMetadata = true
	}
	return
}

// 获取待更新的元数据
func (m *operatorManager) getUpdateMetadataDB(ctx context.Context, req *interfaces.OperatorEditReq, operatorDB *model.OperatorRegisterDB,
	metadataDB interfaces.IMetadataDB) (updateMetadataDB interfaces.IMetadataDB, err error) {
	// 解析传入数据
	switch req.MetadataType {
	case interfaces.MetadataTypeAPI:
		if req.OpenAPIInput == nil || req.Data == nil {
			return
		}
		var updateMetadataDBs []interfaces.IMetadataDB
		updateMetadataDBs, err = m.MetadataService.ParseMetadata(ctx, req.MetadataType, req.OpenAPIInput)
		if err != nil {
			return
		}
		switch interfaces.OperatorType(operatorDB.OperatorType) {
		case interfaces.OperatorTypeBase:
			for _, md := range updateMetadataDBs {
				// 如果是基础算子，根据path和method匹配元数据
				if metadataDB.GetPath() == md.GetPath() && metadataDB.GetMethod() == md.GetMethod() {
					updateMetadataDB = md
					break
				}
			}
			// 检查是否有更新
			if updateMetadataDB == nil {
				// 交互设计要求返回指定错误信息：https://confluence.aishu.cn/pages/viewpage.action?pageId=280780968
				err = infraerrors.NewHTTPError(ctx, http.StatusNotFound, infraerrors.ErrExtCommonNoMatchedMethodPath,
					"no matched method path found or metadata data not exist").WithDescription(infraerrors.ErrExtToolNotExistInFile)
				return
			}
		case interfaces.OperatorTypeComposite:
			// 如果是复合算子，只更新第一个元数据
			updateMetadataDB = updateMetadataDBs[0]
		}
	case interfaces.MetadataTypeFunc:
		if req.FunctionInputEdit == nil {
			return
		}
		funcInput := &interfaces.FunctionInput{
			Name:            req.Name,
			Description:     req.Description,
			Inputs:          req.FunctionInputEdit.Inputs,
			Outputs:         req.FunctionInputEdit.Outputs,
			ScriptType:      req.FunctionInputEdit.ScriptType,
			Code:            req.FunctionInputEdit.Code,
			Dependencies:    req.FunctionInputEdit.Dependencies,
			DependenciesURL: req.FunctionInputEdit.DependenciesURL,
		}
		var updateMetadataDBs []interfaces.IMetadataDB
		updateMetadataDBs, err = m.MetadataService.ParseMetadata(ctx, req.MetadataType, funcInput)
		if err != nil {
			return
		}
		updateMetadataDB = updateMetadataDBs[0]
	default:
		err = infraerrors.DefaultHTTPError(ctx, http.StatusBadRequest, "unsupported metadata type")
		return
	}
	return
}

// modifyOperatorInfo 修改算子注册配置
func (m *operatorManager) modifyOperatorInfo(ctx context.Context, tx *sql.Tx, req *interfaces.OperatorEditReq, operator *model.OperatorRegisterDB,
	metdataDB interfaces.IMetadataDB, needUpdateMetadata, isDataSource bool) (err error) {
	err = m.modifyOperator(ctx, tx, req, operator, isDataSource)
	if err != nil {
		return
	}
	if !needUpdateMetadata {
		return
	}
	// 更新算子元数据
	metdataDB.SetUpdateInfo(req.UserID)
	err = m.MetadataService.UpdateMetadata(ctx, tx, metdataDB)
	if err != nil {
		m.Logger.WithContext(ctx).Warnf("modify api metadata failed, OperatorID: %s, Version: %s, err: %v", operator.OperatorID, operator.MetadataVersion, err)
	}
	return
}

// modifyOperator 编辑算子
func (m *operatorManager) modifyOperator(ctx context.Context, tx *sql.Tx, req *interfaces.OperatorEditReq,
	operator *model.OperatorRegisterDB, isDataSource bool) (err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	// 更新参数
	operator.UpdateUser = req.UserID
	if req.OperatorInfoEdit != nil {
		operator.OperatorType = string(req.OperatorInfoEdit.Type)
		operator.ExecutionMode = string(req.OperatorInfoEdit.ExecutionMode)
		operator.Category = string(req.OperatorInfoEdit.Category)
		operator.Source = req.OperatorInfoEdit.Source
		operator.IsDataSource = isDataSource
	}
	if req.OperatorExecuteControl != nil {
		operator.ExecuteControl = utils.ObjectToJSON(req.OperatorExecuteControl)
	}
	if req.ExtendInfo != nil {
		operator.ExtendInfo = utils.ObjectToJSON(req.ExtendInfo)
	}
	// 如果name发生变化，则根据operatorID更新name
	if req.Name != "" && req.Name != operator.Name { // 检查是否重名
		err = m.checkDuplicateName(ctx, req.Name, operator.OperatorID)
		if err != nil {
			// 交互设计要求返回指定错误信息：https://confluence.aishu.cn/pages/viewpage.action?pageId=280780968
			httErr := &infraerrors.HTTPError{}
			if errors.As(err, &httErr) && httErr.HTTPCode == http.StatusConflict {
				err = httErr.WithDescription(infraerrors.ErrExtCommonNameExists)
			}
			return
		}
		operator.Name = req.Name
	}
	// 更新算子信息
	err = m.DBOperatorManager.UpdateByOperatorID(ctx, tx, operator)
	if err != nil {
		m.Logger.WithContext(ctx).Warnf("update operator failed, OperatorID: %s, Version: %s, err: %v", operator.OperatorID, operator.MetadataVersion, err)
		err = infraerrors.DefaultHTTPError(ctx, http.StatusInternalServerError, "update operator failed, err")
	}
	return
}

// upgradeOperatorInfo 升级算子信息
/*
	已发布版本元数据出现变更，因此需要生成一条新的元数据记录
	1. 元数据表中生成一条新的记录
	2. 更改注册表配置： 包含version，以及本次变更的信息
	3. 如果 direct_publish 为 true， 则直接发布, 需要向release/release_history中添加一条记录
*/

func (m *operatorManager) upgradeOperatorInfo(ctx context.Context, tx *sql.Tx, req *interfaces.OperatorEditReq, operator *model.OperatorRegisterDB,
	metadataDB interfaces.IMetadataDB, needUpdateMetadata, isDataSource bool) (err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	// 升级元数据
	if needUpdateMetadata {
		metadataDB.SetVersion(uuid.New().String())
		metadataDB.SetUpdateInfo(req.UserID)
		_, err = m.MetadataService.RegisterMetadata(ctx, tx, metadataDB)
		if err != nil {
			m.Logger.WithContext(ctx).Errorf("register metadata failed, err: %v", err)
			err = infraerrors.DefaultHTTPError(ctx, http.StatusInternalServerError, "register metadata failed")
			return
		}
	}
	// 3. 组装算子注册信息， 新增到算子注册表
	operator.MetadataVersion = metadataDB.GetVersion()
	err = m.modifyOperator(ctx, tx, req, operator, isDataSource)
	return
}

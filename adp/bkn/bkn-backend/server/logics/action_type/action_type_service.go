// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package action_type

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	bknsdk "github.com/kweaver-ai/bkn-specification/sdk/golang/bkn"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/rs/xid"
	"go.opentelemetry.io/otel/codes"

	"bkn-backend/common"
	cond "bkn-backend/common/condition"
	berrors "bkn-backend/errors"
	"bkn-backend/infra/otel/otellog"
	"bkn-backend/infra/otel/oteltrace"
	"bkn-backend/interfaces"
	"bkn-backend/logics"
	"bkn-backend/logics/batchindex"
	"bkn-backend/logics/object_type"
	"bkn-backend/logics/permission"
	"bkn-backend/logics/user_mgmt"
)

var (
	atServiceOnce sync.Once
	atService     interfaces.ActionTypeService
)

type actionTypeService struct {
	appSetting *common.AppSetting
	db         *sql.DB
	ata        interfaces.ActionTypeAccess
	aoa        interfaces.AgentOperatorAccess
	cga        interfaces.ConceptGroupAccess
	mfa        interfaces.ModelFactoryAccess
	ots        interfaces.ObjectTypeService
	ps         interfaces.PermissionService
	ums        interfaces.UserMgmtService
	vba        interfaces.VegaBackendAccess
}

func NewActionTypeService(appSetting *common.AppSetting) interfaces.ActionTypeService {
	atServiceOnce.Do(func() {
		atService = &actionTypeService{
			appSetting: appSetting,
			db:         logics.DB,
			ata:        logics.ATA,
			aoa:        logics.AOA,
			cga:        logics.CGA,
			mfa:        logics.MFA,
			ots:        object_type.NewObjectTypeService(appSetting),
			ps:         permission.NewPermissionService(appSetting),
			ums:        user_mgmt.NewUserMgmtService(appSetting),
			vba:        logics.VBA,
		}
	})
	return atService
}

func (ats *actionTypeService) CheckActionTypeExistByID(ctx context.Context, knID string, branch string, atID string) (string, bool, error) {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "CheckActionTypeExistByID")
	defer span.End()

	atName, exist, err := ats.ata.CheckActionTypeExistByID(ctx, knID, branch, atID)
	if err != nil {
		logger.Errorf("CheckActionTypeExistByID error: %s", err.Error())
		// 记录处理的 sql 字符串
		otellog.LogError(ctx, fmt.Sprintf("按ID[%v]获取行动类失败", atID), err)
		span.SetStatus(codes.Error, fmt.Sprintf("按ID[%v]获取行动类失败", atID))
		return "", exist, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_ActionType_InternalError_CheckActionTypeIfExistFailed).WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return atName, exist, nil
}

func (ats *actionTypeService) CheckActionTypeExistByName(ctx context.Context, knID string, branch string, atName string) (string, bool, error) {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "CheckActionTypeExistByName")
	defer span.End()

	actionTypeID, exist, err := ats.ata.CheckActionTypeExistByName(ctx, knID, branch, atName)
	if err != nil {
		logger.Errorf("CheckActionTypeExistByName error: %s", err.Error())
		// 记录处理的 sql 字符串
		otellog.LogError(ctx, fmt.Sprintf("按名称[%s]获取行动类失败", atName), err)
		span.SetStatus(codes.Error, fmt.Sprintf("按名称[%s]获取行动类失败", atName))
		return actionTypeID, exist, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_ActionType_InternalError_CheckActionTypeIfExistFailed).WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return actionTypeID, exist, nil
}

// validateActionSourceStrict checks tool-box / MCP tool existence when strict_mode applies (via agent-operator-integration internal APIs).
func (ats *actionTypeService) validateActionSourceStrict(ctx context.Context, at *interfaces.ActionType) error {
	if at == nil {
		return nil
	}
	src := at.ActionSource
	switch src.Type {
	case interfaces.ACTION_SOURCE_TYPE_TOOL:
		if src.BoxID == "" || src.ToolID == "" {
			return nil
		}
		if err := ats.aoa.GetToolByID(ctx, src.BoxID, src.ToolID); err != nil {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ActionType_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("Action type [%s] tool binding is missing or invalid: box_id=%s tool_id=%s (%v)",
					at.ATName, src.BoxID, src.ToolID, err))
		}
	case interfaces.ACTION_SOURCE_TYPE_MCP:
		if src.McpID == "" || src.ToolName == "" {
			return nil
		}
		if err := ats.aoa.GetMcpToolByName(ctx, src.McpID, src.ToolName); err != nil {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ActionType_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("Action type [%s] MCP tool binding is missing or invalid: mcp_id=%s tool_name=%s (%v)",
					at.ATName, src.McpID, src.ToolName, err))
		}
	}
	return nil
}

func (ats *actionTypeService) CreateActionTypes(ctx context.Context, tx *sql.Tx, actionTypes []*interfaces.ActionType, mode string, strictMode bool) ([]string, error) {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "CreateActionTypes")
	defer span.End()

	// 判断userid是否有修改业务知识网络的权限
	err := ats.ps.CheckPermission(ctx, interfaces.PermissionResource{
		Type: interfaces.RESOURCE_TYPE_KN,
		ID:   actionTypes[0].KNID,
	}, []string{interfaces.OPERATION_TYPE_MODIFY})
	if err != nil {
		return []string{}, err
	}

	// 0. 开始事务
	if tx == nil {
		tx, err = ats.db.Begin()
		if err != nil {
			logger.Errorf("Begin transaction error: %s", err.Error())
			span.SetStatus(codes.Error, "事务开启失败")
			otellog.LogError(ctx, "Begin transaction error", err)

			return []string{}, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				berrors.BknBackend_ActionType_InternalError_BeginTransactionFailed).
				WithErrorDetails(err.Error())
		}
		// 0.1 异常时
		defer func() {
			switch err {
			case nil:
				// 提交事务
				err = tx.Commit()
				if err != nil {
					logger.Errorf("CreateActionType Transaction Commit Failed:%v", err)
					span.SetStatus(codes.Error, "提交事务失败")
					otellog.LogError(ctx, "CreateActionType Transaction Commit Failed", err)
					return
				}
				logger.Infof("CreateActionType Transaction Commit Success")
				otellog.LogDebug(ctx, "CreateActionType Transaction Commit Success")
			default:
				rollbackErr := tx.Rollback()
				if rollbackErr != nil {
					logger.Errorf("CreateActionType Transaction Rollback Error:%v", rollbackErr)
					span.SetStatus(codes.Error, "事务回滚失败")
					otellog.LogError(ctx, "CreateActionType Transaction Rollback Error", err)
				}
			}
		}()
	}

	currentTime := time.Now().UnixMilli()
	for _, actionType := range actionTypes {
		// 若提交的模型id为空，生成分布式ID
		if actionType.ATID == "" {
			actionType.ATID = xid.New().String()
		}

		accountInfo := interfaces.AccountInfo{}
		if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
			accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
		}

		actionType.Creator = accountInfo
		actionType.Updater = accountInfo

		actionType.CreateTime = currentTime
		actionType.UpdateTime = currentTime

		// When strictMode is true, validate bound object type and affect object types exist
		if strictMode {
			if actionType.ObjectTypeID != "" {
				_, err = ats.ots.GetObjectTypeByID(ctx, tx, actionType.KNID, actionType.Branch, actionType.ObjectTypeID)
				if err != nil {
					return []string{}, err
				}
			}
			if actionType.Affect != nil && actionType.Affect.ObjectTypeID != "" {
				_, err = ats.ots.GetObjectTypeByID(ctx, tx, actionType.KNID, actionType.Branch, actionType.Affect.ObjectTypeID)
				if err != nil {
					return []string{}, err
				}
			}
			err = ats.validateActionSourceStrict(ctx, actionType)
			if err != nil {
				return []string{}, err
			}
		}

		bknAction := logics.ToBKNActionType(actionType)
		actionType.BKNRawContent = bknsdk.SerializeActionType(bknAction)
	}

	createActionTypes, updateActionTypes, err := ats.handleActionTypeImportMode(ctx, mode, actionTypes)
	if err != nil {
		return []string{}, err
	}

	// 创建
	atIDs := []string{}
	for _, actionType := range createActionTypes {
		atIDs = append(atIDs, actionType.ATID)
		err = ats.ata.CreateActionType(ctx, tx, actionType)
		if err != nil {
			logger.Errorf("CreateActionType error: %s", err.Error())
			span.SetStatus(codes.Error, "创建行动类失败")
			return []string{}, rest.NewHTTPError(ctx, http.StatusInternalServerError, berrors.BknBackend_ActionType_InternalError).
				WithErrorDetails(err.Error())
		}
	}

	// 更新
	for _, actionType := range updateActionTypes {
		// 提交的已存在，需要更新
		err = ats.UpdateActionType(ctx, tx, actionType, strictMode)
		if err != nil {
			return []string{}, err
		}
	}

	insetActionTypes := createActionTypes
	insetActionTypes = append(insetActionTypes, updateActionTypes...)
	err = ats.InsertDatasetData(ctx, insetActionTypes)
	if err != nil {
		logger.Errorf("InsertDatasetData error: %s", err.Error())
		span.SetStatus(codes.Error, "行动类索引写入失败")
		return []string{}, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_ActionType_InternalError_InsertOpenSearchDataFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return atIDs, nil
}

// ValidateActionTypes checks dependency existence only; does not write to the database.
func (ats *actionTypeService) ValidateActionTypes(ctx context.Context, knID string, branch string,
	actionTypes []*interfaces.ActionType, strictMode bool, batch *interfaces.BatchIDIndex, mode string) error {

	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "ValidateActionTypes")
	defer span.End()

	if len(actionTypes) == 0 {
		span.SetStatus(codes.Ok, "")
		return nil
	}

	err := ats.ps.CheckPermission(ctx, interfaces.PermissionResource{
		Type: interfaces.RESOURCE_TYPE_KN,
		ID:   knID,
	}, []string{interfaces.OPERATION_TYPE_MODIFY})
	if err != nil {
		return err
	}
	_, _, err = ats.handleActionTypeImportMode(ctx, mode, actionTypes)
	if err != nil {
		return err
	}

	for _, actionType := range actionTypes {
		actionType.KNID = knID
		actionType.Branch = branch
		if strictMode {
			if actionType.ObjectTypeID != "" {
				if batch == nil || !batchindex.HasObjectTypeID(actionType.ObjectTypeID, batch) {
					_, err = ats.ots.GetObjectTypeByID(ctx, nil, knID, branch, actionType.ObjectTypeID)
					if err != nil {
						return err
					}
				}
			}
			if actionType.Affect != nil && actionType.Affect.ObjectTypeID != "" {
				if batch == nil || !batchindex.HasObjectTypeID(actionType.Affect.ObjectTypeID, batch) {
					_, err = ats.ots.GetObjectTypeByID(ctx, nil, knID, branch, actionType.Affect.ObjectTypeID)
					if err != nil {
						return err
					}
				}
			}
			err = ats.validateActionSourceStrict(ctx, actionType)
			if err != nil {
				return err
			}
		}
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

func (ats *actionTypeService) ListActionTypes(ctx context.Context, query interfaces.ActionTypesQueryParams) ([]*interfaces.ActionType, int, error) {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "ListActionTypes")
	defer span.End()

	// 判断userid是否有查看业务知识网络的权限
	err := ats.ps.CheckPermission(ctx, interfaces.PermissionResource{
		Type: interfaces.RESOURCE_TYPE_KN,
		ID:   query.KNID,
	}, []string{interfaces.OPERATION_TYPE_VIEW_DETAIL})
	if err != nil {
		return []*interfaces.ActionType{}, 0, err
	}

	//获取行动类列表
	actionTypes, err := ats.ata.ListActionTypes(ctx, query)
	if err != nil {
		logger.Errorf("ListActionTypes error: %s", err.Error())
		span.SetStatus(codes.Error, "List action types error")
		return []*interfaces.ActionType{}, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_ActionType_InternalError).WithErrorDetails(err.Error())
	}
	if len(actionTypes) == 0 {
		span.SetStatus(codes.Ok, "")
		return actionTypes, 0, nil
	}

	// 获取绑定对象类的名称拿到
	for _, actionType := range actionTypes {
		objectTypeMap, err := ats.ots.GetObjectTypesMapByIDs(ctx, query.KNID,
			query.Branch, []string{actionType.ObjectTypeID}, false)
		if err != nil {
			return []*interfaces.ActionType{}, 0, err
		}

		if objectTypeMap[actionType.ObjectTypeID] != nil {
			actionType.ObjectType = interfaces.SimpleObjectType{
				OTID:   objectTypeMap[actionType.ObjectTypeID].OTID,
				OTName: objectTypeMap[actionType.ObjectTypeID].OTName,
				Icon:   objectTypeMap[actionType.ObjectTypeID].Icon,
				Color:  objectTypeMap[actionType.ObjectTypeID].Color,
			}
		}
	}
	total := len(actionTypes)

	// limit = -1,则返回所有
	if query.Limit != -1 {
		// 分页
		// 检查起始位置是否越界
		if query.Offset < 0 || query.Offset >= len(actionTypes) {
			span.SetStatus(codes.Ok, "")
			return []*interfaces.ActionType{}, total, nil
		}
		// 计算结束位置
		end := query.Offset + query.Limit
		if end > len(actionTypes) {
			end = len(actionTypes)
		}

		actionTypes = actionTypes[query.Offset:end]
	}

	accountInfos := make([]*interfaces.AccountInfo, 0, len(actionTypes)*2)
	for _, at := range actionTypes {
		accountInfos = append(accountInfos, &at.Creator, &at.Updater)
	}

	err = ats.ums.GetAccountNames(ctx, accountInfos)
	if err != nil {
		span.SetStatus(codes.Error, "GetAccountNames error")
		return []*interfaces.ActionType{}, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_ActionType_InternalError).WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return actionTypes, total, nil
}

func (ats *actionTypeService) GetActionTypesByIDs(ctx context.Context, knID string, branch string, atIDs []string) ([]*interfaces.ActionType, error) {
	// 获取行动类
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "GetActionTypesByIDs")
	defer span.End()

	// 判断userid是否有查看业务知识网络的权限
	err := ats.ps.CheckPermission(ctx, interfaces.PermissionResource{
		Type: interfaces.RESOURCE_TYPE_KN,
		ID:   knID,
	}, []string{interfaces.OPERATION_TYPE_VIEW_DETAIL})
	if err != nil {
		return []*interfaces.ActionType{}, err
	}

	// id去重后再查
	atIDs = common.DuplicateSlice(atIDs)

	// 获取模型基本信息
	actionTypes, err := ats.ata.GetActionTypesByIDs(ctx, knID, branch, atIDs)
	if err != nil {
		logger.Errorf("GetActionTypesByATIDs error: %s", err.Error())
		span.SetStatus(codes.Error, fmt.Sprintf("Get action type[%v] error: %v", atIDs, err))
		return []*interfaces.ActionType{}, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_ActionType_InternalError_GetActionTypesByIDsFailed).
			WithErrorDetails(err.Error())
	}

	if len(actionTypes) != len(atIDs) {
		errStr := fmt.Sprintf("Exists any action types not found, expect action types nums is [%d], actual action types num is [%d]", len(atIDs), len(actionTypes))
		logger.Errorf(errStr)
		span.SetStatus(codes.Error, errStr)
		return []*interfaces.ActionType{}, rest.NewHTTPError(ctx, http.StatusNotFound,
			berrors.BknBackend_ActionType_ActionTypeNotFound).WithErrorDetails(errStr)
	}

	// todo:翻译绑定的对象类、影响对象类、和对应的api文档
	// 获取绑定对象类和影响对象类的名称拿到
	for _, actionType := range actionTypes {
		affectObjectTypeID := ""
		if actionType.Affect != nil && actionType.Affect.ObjectTypeID != "" {
			affectObjectTypeID = actionType.Affect.ObjectTypeID
		}

		objectTypeMap, err := ats.ots.GetObjectTypesMapByIDs(ctx, knID, branch,
			[]string{actionType.ObjectTypeID, affectObjectTypeID}, false)
		if err != nil {
			return []*interfaces.ActionType{}, err
		}

		if objectTypeMap[actionType.ObjectTypeID] != nil {
			actionType.ObjectType = interfaces.SimpleObjectType{
				OTID:   objectTypeMap[actionType.ObjectTypeID].OTID,
				OTName: objectTypeMap[actionType.ObjectTypeID].OTName,
				Icon:   objectTypeMap[actionType.ObjectTypeID].Icon,
				Color:  objectTypeMap[actionType.ObjectTypeID].Color,
			}
		}

		if objectTypeMap[affectObjectTypeID] != nil {
			actionType.Affect.ObjectType = interfaces.SimpleObjectType{
				OTID:   objectTypeMap[affectObjectTypeID].OTID,
				OTName: objectTypeMap[affectObjectTypeID].OTName,
				Icon:   objectTypeMap[affectObjectTypeID].Icon,
				Color:  objectTypeMap[affectObjectTypeID].Color,
			}
		}
	}

	span.SetStatus(codes.Ok, "")
	return actionTypes, nil
}

// 更新行动类
func (ats *actionTypeService) UpdateActionType(ctx context.Context, tx *sql.Tx, actionType *interfaces.ActionType, strictMode bool) error {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "UpdateActionType")
	defer span.End()

	// 判断userid是否有修改业务知识网络的权限
	err := ats.ps.CheckPermission(ctx, interfaces.PermissionResource{
		Type: interfaces.RESOURCE_TYPE_KN,
		ID:   actionType.KNID,
	}, []string{interfaces.OPERATION_TYPE_MODIFY})
	if err != nil {
		return err
	}

	if strictMode {
		if actionType.ObjectTypeID != "" {
			_, err = ats.ots.GetObjectTypeByID(ctx, tx, actionType.KNID, actionType.Branch, actionType.ObjectTypeID)
			if err != nil {
				return err
			}
		}
		if actionType.Affect != nil && actionType.Affect.ObjectTypeID != "" {
			_, err = ats.ots.GetObjectTypeByID(ctx, tx, actionType.KNID, actionType.Branch, actionType.Affect.ObjectTypeID)
			if err != nil {
				return err
			}
		}
		err = ats.validateActionSourceStrict(ctx, actionType)
		if err != nil {
			return err
		}
	}

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}
	actionType.Updater = accountInfo

	currentTime := time.Now().UnixMilli() // 行动类的update_time是int类型
	actionType.UpdateTime = currentTime

	bknAction := logics.ToBKNActionType(actionType)
	actionType.BKNRawContent = bknsdk.SerializeActionType(bknAction)

	if tx == nil {
		// 0. 开始事务
		tx, err = ats.db.Begin()
		if err != nil {
			logger.Errorf("Begin transaction error: %s", err.Error())
			span.SetStatus(codes.Error, "事务开启失败")
			otellog.LogError(ctx, "Begin transaction error", err)
			return rest.NewHTTPError(ctx, http.StatusInternalServerError, berrors.BknBackend_ActionType_InternalError_BeginTransactionFailed).
				WithErrorDetails(err.Error())
		}
		// 0.1 异常时
		defer func() {
			switch err {
			case nil:
				// 提交事务
				err = tx.Commit()
				if err != nil {
					logger.Errorf("UpdateActionType Transaction Commit Failed:%v", err)
					span.SetStatus(codes.Error, "提交事务失败")
					otellog.LogError(ctx, "UpdateActionType Transaction Commit Failed", err)
				}
				logger.Infof("UpdateActionType Transaction Commit Success:%v", actionType.ATName)
				otellog.LogDebug(ctx, fmt.Sprintf("UpdateActionType Transaction Commit Success: %s", actionType.ATName))
			default:
				rollbackErr := tx.Rollback()
				if rollbackErr != nil {
					logger.Errorf("UpdateActionType Transaction Rollback Error:%v", rollbackErr)
					span.SetStatus(codes.Error, "事务回滚失败")
					otellog.LogError(ctx, "UpdateActionType Transaction Rollback Error", rollbackErr)
				}
			}
		}()
	}

	// 更新模型信息
	err = ats.ata.UpdateActionType(ctx, tx, actionType)
	if err != nil {
		logger.Errorf("UpdateActionType error: %s", err.Error())
		span.SetStatus(codes.Error, "修改行动类失败")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_ActionType_InternalError).
			WithErrorDetails(err.Error())
	}

	err = ats.InsertDatasetData(ctx, []*interfaces.ActionType{actionType})
	if err != nil {
		logger.Errorf("InsertDatasetData error: %s", err.Error())
		span.SetStatus(codes.Error, "行动类索引写入失败")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_ActionType_InternalError_InsertOpenSearchDataFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

func (ats *actionTypeService) DeleteActionTypesByIDs(ctx context.Context, tx *sql.Tx, knID string, branch string, atIDs []string) error {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "DeleteActionTypesByIDs")
	defer span.End()

	// 判断userid是否有修改业务知识网络的权限
	err := ats.ps.CheckPermission(ctx, interfaces.PermissionResource{
		Type: interfaces.RESOURCE_TYPE_KN,
		ID:   knID,
	}, []string{interfaces.OPERATION_TYPE_MODIFY})
	if err != nil {
		return err
	}

	if tx == nil {
		// 0. 开始事务
		tx, err = ats.db.Begin()
		if err != nil {
			logger.Errorf("Begin transaction error: %s", err.Error())
			span.SetStatus(codes.Error, "事务开启失败")
			otellog.LogError(ctx, "Begin transaction error", err)

			return rest.NewHTTPError(ctx, http.StatusInternalServerError,
				berrors.BknBackend_ActionType_InternalError_BeginTransactionFailed).
				WithErrorDetails(err.Error())
		}
		// 0.1 异常时
		defer func() {
			switch err {
			case nil:
				// 提交事务
				err = tx.Commit()
				if err != nil {
					logger.Errorf("DeleteActionTypes Transaction Commit Failed:%v", err)
					span.SetStatus(codes.Error, "提交事务失败")
					otellog.LogError(ctx, "DeleteActionTypes Transaction Commit Failed", err)
				}
				logger.Infof("DeleteActionTypes Transaction Commit Success: kn_id:%s,ot_ids:%v", knID, atIDs)
				otellog.LogDebug(ctx, fmt.Sprintf("DeleteActionTypes Transaction Commit Success: kn_id:%s,ot_ids:%v", knID, atIDs))
			default:
				rollbackErr := tx.Rollback()
				if rollbackErr != nil {
					logger.Errorf("DeleteActionTypes Transaction Rollback Error:%v", rollbackErr)
					span.SetStatus(codes.Error, "事务回滚失败")
					otellog.LogError(ctx, "DeleteActionTypes Transaction Rollback Error", rollbackErr)
				}
			}
		}()
	}

	// 删除行动类
	rowsAffect, err := ats.ata.DeleteActionTypesByIDs(ctx, tx, knID, branch, atIDs)
	if err != nil {
		logger.Errorf("DeleteActionTypes error: %s", err.Error())
		span.SetStatus(codes.Error, "删除行动类失败")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_ActionType_InternalError).WithErrorDetails(err.Error())
	}

	logger.Infof("DeleteActionTypes: Rows affected is %v, request delete ATIDs is %v!", rowsAffect, len(atIDs))
	if rowsAffect != int64(len(atIDs)) {
		logger.Warnf("Delete action types number %v not equal requerst action types number %v!", rowsAffect, len(atIDs))
		otellog.LogWarn(ctx, fmt.Sprintf("Delete action types number %v not equal requerst action types number %v!", rowsAffect, len(atIDs)))
	}

	for _, atID := range atIDs {
		docid := interfaces.GenerateConceptDocuemtnID(knID, interfaces.MODULE_TYPE_ACTION_TYPE, atID, branch)
		err = ats.vba.DeleteDatasetDocumentByID(ctx, interfaces.BKN_DATASET_ID, docid)
		if err != nil {
			logger.Errorf("DeleteDatasetDocumentByID error: %s", err.Error())
			span.SetStatus(codes.Error, "删除行动类概念索引失败")
			return err
		}
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 内部接口，不校验权限， tx必须传
func (ats *actionTypeService) DeleteActionTypesByKnID(ctx context.Context, tx *sql.Tx, knID string, branch string) error {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "DeleteActionTypesByKnID")
	defer span.End()

	if tx == nil {
		otellog.LogError(ctx, "missing transaction", nil)
		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_ActionType_InternalError_MissingTransaction).
			WithErrorDetails("missing transaction")
	}

	// 删除行动类
	rowsAffect, err := ats.ata.DeleteActionTypesByKnID(ctx, tx, knID, branch)
	if err != nil {
		logger.Errorf("DeleteActionTypes error: %s", err.Error())
		span.SetStatus(codes.Error, "删除行动类失败")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_ActionType_InternalError).WithErrorDetails(err.Error())
	}

	logger.Infof("DeleteActionTypesByKnID success, the kn_id is [%s], branch is [%s], rowsAffect is [%d]",
		knID, branch, rowsAffect)
	span.SetStatus(codes.Ok, "")
	return nil
}

func (ats *actionTypeService) handleActionTypeImportMode(ctx context.Context, mode string,
	actionTypes []*interfaces.ActionType) ([]*interfaces.ActionType, []*interfaces.ActionType, error) {

	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "handleActionTypeImportMode")
	defer span.End()

	creates := []*interfaces.ActionType{}
	updates := []*interfaces.ActionType{}

	// 3. 校验 若模型的id不为空，则用请求体的id与现有模型ID的重复性
	for _, actionType := range actionTypes {
		creates = append(creates, actionType)
		idExist := false
		_, idExist, err := ats.CheckActionTypeExistByID(ctx, actionType.KNID, actionType.Branch, actionType.ATID)
		if err != nil {
			return creates, updates, err
		}

		// 校验 请求体与现有模型名称的重复性
		existID, nameExist, err := ats.CheckActionTypeExistByName(ctx, actionType.KNID, actionType.Branch, actionType.ATName)
		if err != nil {
			return creates, updates, err
		}

		// 根据mode来区别，若是ignore，就从结果集中忽略，若是overwrite，就调用update，若是normal就报错。
		if idExist || nameExist {
			switch mode {
			case interfaces.ImportMode_Normal:
				if idExist {
					errDetails := fmt.Sprintf("The action type with id [%s] already exists!", actionType.ATID)
					logger.Error(errDetails)
					span.SetStatus(codes.Error, errDetails)
					return creates, updates, rest.NewHTTPError(ctx, http.StatusBadRequest,
						berrors.BknBackend_ActionType_ActionTypeIDExisted).
						WithErrorDetails(errDetails)
				}

				if nameExist {
					errDetails := fmt.Sprintf("action type name '%s' already exists", actionType.ATName)
					logger.Error(errDetails)
					span.SetStatus(codes.Error, errDetails)
					return creates, updates, rest.NewHTTPError(ctx, http.StatusForbidden,
						berrors.BknBackend_ActionType_ActionTypeNameExisted).
						WithDescription(map[string]any{"name": actionType.ATName}).
						WithErrorDetails(errDetails)
				}

			case interfaces.ImportMode_Ignore:
				// 存在重复的就跳过
				// 从create数组中删除
				creates = creates[:len(creates)-1]
			case interfaces.ImportMode_Overwrite:
				if idExist && nameExist {
					// 如果 id 和名称都存在，但是存在的名称对应的行动类 id 和当前行动类 id 不一样，则报错
					if existID != actionType.ATID {
						errDetails := fmt.Sprintf("ActionType ID '%s' and name '%s' already exist, but the exist action type id is '%s'",
							actionType.ATID, actionType.ATName, existID)
						logger.Error(errDetails)
						span.SetStatus(codes.Error, errDetails)
						return creates, updates, rest.NewHTTPError(ctx, http.StatusForbidden,
							berrors.BknBackend_ActionType_ActionTypeNameExisted).
							WithErrorDetails(errDetails)
					} else {
						// 如果 id 和名称、度量名称都存在，存在的名称对应的模型 id 和当前模型 id 一样，则覆盖更新
						// 从create数组中删除, 放到更新数组中
						creates = creates[:len(creates)-1]
						updates = append(updates, actionType)
					}
				}

				// id 已存在，且名称不存在，覆盖更新
				if idExist && !nameExist {
					// 从create数组中删除, 放到更新数组中
					creates = creates[:len(creates)-1]
					updates = append(updates, actionType)
				}

				// 如果 id 不存在，name 存在，报错
				if !idExist && nameExist {
					errDetails := fmt.Sprintf("ActionType ID '%s' does not exist, but name '%s' already exists",
						actionType.ATID, actionType.ATName)
					logger.Error(errDetails)
					span.SetStatus(codes.Error, errDetails)
					return creates, updates, rest.NewHTTPError(ctx, http.StatusForbidden,
						berrors.BknBackend_ActionType_ActionTypeNameExisted).
						WithErrorDetails(errDetails)
				}

				// 如果 id 不存在，name不存在，度量名称不存在，不需要做什么，创建
				// if !idExist && !nameExist {}
			}
		}
	}

	span.SetStatus(codes.Ok, "")
	return creates, updates, nil
}

func (ats *actionTypeService) InsertDatasetData(ctx context.Context, actionTypes []*interfaces.ActionType) error {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "行动类索引写入")
	defer span.End()

	if len(actionTypes) == 0 {
		return nil
	}

	if ats.appSetting.ServerSetting.DefaultSmallModelEnabled {
		words := []string{}
		for _, actionType := range actionTypes {
			arr := []string{actionType.ATName}
			arr = append(arr, actionType.Tags...)
			arr = append(arr, actionType.Comment, actionType.BKNRawContent)
			word := strings.Join(arr, "\n")
			words = append(words, word)
		}

		dftModel, err := ats.mfa.GetDefaultModel(ctx)
		if err != nil {
			logger.Errorf("GetDefaultModel error: %s", err.Error())
			span.SetStatus(codes.Error, "获取默认模型失败")
			return err
		}
		vectors, err := ats.mfa.GetVector(ctx, dftModel, words)
		if err != nil {
			logger.Errorf("GetVector error: %s", err.Error())
			span.SetStatus(codes.Error, "获取行动类向量失败")
			return err
		}

		if len(vectors) != len(actionTypes) {
			logger.Errorf("GetVector error: expect vectors num is [%d], actual vectors num is [%d]", len(actionTypes), len(vectors))
			span.SetStatus(codes.Error, "获取行动类向量失败")
			return fmt.Errorf("GetVector error: expect vectors num is [%d], actual vectors num is [%d]", len(actionTypes), len(vectors))
		}

		for i, actionType := range actionTypes {
			actionType.Vector = vectors[i].Vector
		}
	}

	documents := []map[string]any{}
	for _, actionType := range actionTypes {
		docid := interfaces.GenerateConceptDocuemtnID(actionType.KNID, interfaces.MODULE_TYPE_ACTION_TYPE,
			actionType.ATID, actionType.Branch)
		actionType.ModuleType = interfaces.MODULE_TYPE_ACTION_TYPE

		// Convert to map for dataset
		docBytes, err := sonic.Marshal(actionType)
		if err != nil {
			logger.Errorf("Failed to marshal ActionType: %s", err.Error())
			span.SetStatus(codes.Error, "序列化行动类失败")
			return err
		}

		var doc map[string]any
		if err := sonic.Unmarshal(docBytes, &doc); err != nil {
			logger.Errorf("Failed to unmarshal ActionType: %s", err.Error())
			span.SetStatus(codes.Error, "反序列化行动类失败")
			return err
		}

		// Serialize parameters to JSON string
		if params, exists := doc["parameters"]; exists {
			paramsBytes, err := sonic.Marshal(params)
			if err != nil {
				logger.Errorf("Failed to marshal action_type parameters: %s", err.Error())
				span.SetStatus(codes.Error, "序列化行动类参数失败")
				return err
			}
			doc["parameters"] = string(paramsBytes)
		}

		// Serialize condition to JSON string
		if cond, exists := doc["condition"]; exists && cond != nil {
			condBytes, err := sonic.Marshal(cond)
			if err != nil {
				logger.Errorf("Failed to marshal action_type condition: %s", err.Error())
				span.SetStatus(codes.Error, "序列化行动类条件失败")
				return err
			}
			doc["condition"] = string(condBytes)
		}

		// Set document ID
		doc["_id"] = docid
		documents = append(documents, doc)
	}

	err := ats.vba.WriteDatasetDocuments(ctx, interfaces.BKN_DATASET_ID, documents)
	if err != nil {
		logger.Errorf("WriteDatasetDocuments error: %s", err.Error())
		span.SetStatus(codes.Error, "行动类概念索引写入失败")
		return err
	}

	return nil
}

func (ats *actionTypeService) SearchActionTypes(ctx context.Context, query *interfaces.ConceptsQuery) (interfaces.ActionTypes, error) {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "SearchActionTypes")
	defer span.End()

	response := interfaces.ActionTypes{}
	var err error

	// 判断userid是否有查看业务知识网络的权限
	err = ats.ps.CheckPermission(ctx, interfaces.PermissionResource{
		Type: interfaces.RESOURCE_TYPE_KN,
		ID:   query.KNID,
	}, []string{interfaces.OPERATION_TYPE_VIEW_DETAIL})
	if err != nil {
		return response, err
	}

	// 转换条件为 dataset filter condition
	var filterCondition map[string]any
	if query.ActualCondition != nil {
		filterCondition, err = cond.ConvertCondCfgToFilterCondition(ctx, query.ActualCondition,
			interfaces.CONCPET_QUERY_FIELD,
			func(ctx context.Context, word string) ([]*cond.VectorResp, error) {
				if !ats.appSetting.ServerSetting.DefaultSmallModelEnabled {
					err = errors.New(cond.DEFAULT_SMALL_MODEL_ENABLED_FALSE_ERROR)
					span.SetStatus(codes.Error, err.Error())
					return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
						berrors.BknBackend_ActionType_InternalError).
						WithErrorDetails(err.Error())
				}
				dftModel, err := ats.mfa.GetDefaultModel(ctx)
				if err != nil {
					logger.Errorf("GetDefaultModel error: %s", err.Error())
					span.SetStatus(codes.Error, "获取默认模型失败")
					return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
						berrors.BknBackend_ActionType_InternalError).
						WithErrorDetails(err.Error())
				}
				result, err := ats.mfa.GetVector(ctx, dftModel, []string{word})
				if err != nil {
					logger.Errorf("GetVector error: %s", err.Error())
					span.SetStatus(codes.Error, "获取业务知识网络向量失败")
					return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
						berrors.BknBackend_ActionType_InternalError).
						WithErrorDetails(err.Error())
				}
				return result, nil
			})
		if err != nil {
			return response, rest.NewHTTPError(ctx, http.StatusBadRequest,
				berrors.BknBackend_ActionType_InvalidParameter_ConceptCondition).
				WithErrorDetails(fmt.Sprintf("failed to convert condition to filter condition, %s", err.Error()))
		}
	}

	// 1. 获取组下的关系类
	atIDMap := map[string]bool{} // 分组下的对象类id
	atIDs := []string{}          // 不同组下的对象类可以重叠，所以需要对对象类id的数组去重
	if len(query.ConceptGroups) > 0 {
		// 校验分组是否都存在，按分组id获取分组
		cgCnt, err := ats.cga.GetConceptGroupsTotal(ctx, interfaces.ConceptGroupsQueryParams{
			KNID:   query.KNID,
			Branch: query.Branch,
			CGIDs:  query.ConceptGroups,
		})
		if err != nil {
			logger.Errorf("GetConceptGroupsTotal in knowledge network[%s] error: %s", query.KNID, err.Error())
			span.SetStatus(codes.Error, fmt.Sprintf("GetConceptGroupsTotal in knowledge network[%s], error: %v", query.KNID, err))

			return response, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				berrors.BknBackend_ActionType_InternalError).WithErrorDetails(err.Error())
		}
		if cgCnt == 0 {
			errStr := fmt.Sprintf("all concept group not found, expect concept group nums is [%d], actual concept group num is [%d]",
				cgCnt, len(query.ConceptGroups))
			logger.Errorf(errStr)

			return response, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				berrors.BknBackend_ActionType_InternalError).
				WithErrorDetails(errStr)
		}

		// 在当前业务知识网络下查找属于请求的分组范围内的行动类ID
		atIDArr, err := ats.cga.GetActionTypeIDsFromConceptGroupRelation(ctx, interfaces.ConceptGroupRelationsQueryParams{
			KNID:        query.KNID,
			Branch:      query.Branch,
			ConceptType: interfaces.MODULE_TYPE_OBJECT_TYPE, // 概念与分组关系中的概念类型
			CGIDs:       query.ConceptGroups,
		})
		if err != nil {
			errStr := fmt.Sprintf("GetActionTypeIDsFromConceptGroupRelation failed, kn_id:[%s],branch:[%s],cg_ids:[%v], error: %v",
				query.KNID, query.Branch, query.ConceptGroups, err)
			logger.Errorf(errStr)
			span.SetStatus(codes.Error, errStr)
			span.End()

			return response, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				berrors.BknBackend_ActionType_InternalError).WithErrorDetails(errStr)
		}
		// 概念分组下没有行动类,返回空
		if len(atIDArr) == 0 {
			return response, nil
		}

		for _, atID := range atIDArr {
			if !atIDMap[atID] {
				atIDMap[atID] = true
				atIDs = append(atIDs, atID)
			}
		}
	}

	// 根据NeedTotal参数决定是否查询total
	if query.NeedTotal {
		if len(atIDMap) == 0 {
			// 查询总数
			params := &interfaces.ResourceDataQueryParams{
				FilterCondition: filterCondition,
				Offset:          0,
				Limit:           1, // 查询1条数据，获取total
				NeedTotal:       true,
			}
			datasetResp, err := ats.vba.QueryResourceData(ctx, interfaces.BKN_DATASET_ID, params)
			if err != nil {
				logger.Errorf("QueryResourceData error: %s", err.Error())
				span.SetStatus(codes.Error, "业务知识网络行动类检索查询总数失败")
				return response, rest.NewHTTPError(ctx, http.StatusInternalServerError,
					berrors.BknBackend_ActionType_InternalError).
					WithErrorDetails(err.Error())
			}
			response.TotalCount = datasetResp.TotalCount
		} else {
			// 指定了分组，需要查询分组内且符合条件的总数
			total, err := ats.GetTotalWithLargeATIDs(ctx, filterCondition, atIDs)
			if err != nil {
				return response, err
			}
			response.TotalCount = total
		}
	}

	// 4. 迭代查询直到获取足够数量或没有更多数据
	// 保留 search_after 的循环查询结构，当前使用 offset 分页作为临时方案
	actionTypes := []*interfaces.ActionType{}
	var totalFilteredCount int64 = 0
	offset := 0
	limit := query.Limit
	if limit == 0 {
		limit = interfaces.SearchAfter_Limit
	}

	for {
		// 调用 dataset 查询
		params := &interfaces.ResourceDataQueryParams{
			FilterCondition: filterCondition,
			Offset:          offset,
			Limit:           limit,
			NeedTotal:       false,
			Sort:            query.Sort,
		}
		datasetResp, err := ats.vba.QueryResourceData(ctx, interfaces.BKN_DATASET_ID, params)
		if err != nil {
			logger.Errorf("QueryResourceData error: %s", err.Error())
			span.SetStatus(codes.Error, "业务知识网络行动类检索查询失败")
			return response, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				berrors.BknBackend_ActionType_InternalError).
				WithErrorDetails(err.Error())
		}

		// 如果没有数据了，跳出循环
		if len(datasetResp.Entries) == 0 {
			break
		}

		// 5. 处理查询结果
		for _, entry := range datasetResp.Entries {
			// Deserialize condition from JSON string
			if condStr, exists := entry["condition"]; exists {
				if condStrStr, ok := condStr.(string); ok && condStrStr != "" {
					var condCfg interfaces.ActionCondCfg
					if err := sonic.Unmarshal([]byte(condStrStr), &condCfg); err != nil {
						logger.Errorf("Failed to unmarshal action_type condition: %s", err.Error())
						return response, rest.NewHTTPError(ctx, http.StatusBadRequest,
							berrors.BknBackend_InternalError_UnMarshalDataFailed).
							WithErrorDetails(fmt.Sprintf("failed to Unmarshal condition, %s", err.Error()))
					}
					entry["condition"] = &condCfg
				} else if condStr == nil {
					entry["condition"] = nil
				}
			}

			// Deserialize parameters from JSON string
			if paramsStr, exists := entry["parameters"]; exists {
				if paramsStrStr, ok := paramsStr.(string); ok && paramsStrStr != "" {
					var params []interfaces.Parameter
					if err := sonic.Unmarshal([]byte(paramsStrStr), &params); err != nil {
						logger.Errorf("Failed to unmarshal action_type parameters: %s", err.Error())
						return response, rest.NewHTTPError(ctx, http.StatusBadRequest,
							berrors.BknBackend_InternalError_UnMarshalDataFailed).
							WithErrorDetails(fmt.Sprintf("failed to Unmarshal parameters, %s", err.Error()))
					}
					entry["parameters"] = params
				}
			}

			// 转成 action type 的 struct
			jsonByte, err := json.Marshal(entry)
			if err != nil {
				return response, rest.NewHTTPError(ctx, http.StatusBadRequest,
					berrors.BknBackend_InternalError_MarshalDataFailed).
					WithErrorDetails(fmt.Sprintf("failed to Marshal dataset entry, %s", err.Error()))
			}
			var actionType interfaces.ActionType
			err = json.Unmarshal(jsonByte, &actionType)
			if err != nil {
				return response, rest.NewHTTPError(ctx, http.StatusBadRequest,
					berrors.BknBackend_InternalError_UnMarshalDataFailed).
					WithErrorDetails(fmt.Sprintf("failed to Unmarshal dataset entry to Action Type, %s", err.Error()))
			}

			// 如果没有指定分组，或者行动类属于分组，则添加
			if len(atIDMap) == 0 || atIDMap[actionType.ATID] {
				// 提取 _score（如果有）
				if scoreVal, ok := entry["_score"]; ok {
					if scoreFloat, ok := scoreVal.(float64); ok {
						score := float64(scoreFloat)
						actionType.Score = &score
					}
				}
				actionType.Vector = nil
				actionTypes = append(actionTypes, &actionType)
				totalFilteredCount++

				// 如果已经收集到足够的数量，跳出循环
				if len(actionTypes) >= query.Limit && query.Limit > 0 {
					break
				}
			}
		}

		query.SearchAfter = datasetResp.SearchAfter

		// 如果已经收集到足够的数量或者没有更多数据了，跳出循环
		if (query.Limit > 0 && len(actionTypes) >= query.Limit) || len(datasetResp.Entries) < limit {
			break
		}

		// 更新 offset 用于下一次查询（当前使用 offset 分页作为临时方案）
		offset += limit
	}

	response.Entries = actionTypes
	response.SearchAfter = query.SearchAfter
	return response, nil
}

func (ats *actionTypeService) GetTotal(ctx context.Context, filterCondition map[string]any) (total int64, err error) {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "GetTotal")
	defer span.End()

	// 添加 module_type 过滤条件
	if filterCondition == nil {
		filterCondition = map[string]any{
			"field":      "module_type",
			"operation":  "==",
			"value":      interfaces.MODULE_TYPE_ACTION_TYPE,
			"value_from": "const",
		}
	} else {
		filterCondition = map[string]any{
			"operation": "and",
			"sub_conditions": []map[string]any{
				filterCondition,
				{
					"field":      "module_type",
					"operation":  "==",
					"value":      interfaces.MODULE_TYPE_ACTION_TYPE,
					"value_from": "const",
				},
			},
		}
	}

	params := &interfaces.ResourceDataQueryParams{
		FilterCondition: filterCondition,
		Offset:          0,
		Limit:           1, // 查询1条数据，获取total
		NeedTotal:       true,
	}
	datasetResp, err := ats.vba.QueryResourceData(ctx, interfaces.BKN_DATASET_ID, params)
	if err != nil {
		span.SetStatus(codes.Error, "Search total documents count failed")
		return total, rest.NewHTTPError(ctx, http.StatusInternalServerError, berrors.BknBackend_ActionType_InternalError).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	if datasetResp == nil {
		return 0, nil
	}
	return datasetResp.TotalCount, nil
}

// 内部调用，不加权限校验
func (ats *actionTypeService) GetActionTypeIDsByKnID(ctx context.Context, knID string, branch string) ([]string, error) {
	// 获取行动类
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "GetActionTypeIDsByKnID")
	defer span.End()

	// 获取模型基本信息
	atIDs, err := ats.ata.GetActionTypeIDsByKnID(ctx, knID, branch)
	if err != nil {
		logger.Errorf("GetActionTypeIDsByKnID error: %s", err.Error())
		span.SetStatus(codes.Error, fmt.Sprintf("Get action type[%v] error: %v", atIDs, err))
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_ActionType_InternalError_GetActionTypesByIDsFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return atIDs, nil
}

// 分批查询
func (ats *actionTypeService) GetTotalWithLargeATIDs(ctx context.Context,
	filterCondition map[string]any,
	atIDs []string) (int64, error) {

	total := int64(0)
	for i := 0; i < len(atIDs); i += interfaces.GET_TOTAL_CONCEPTID_BATCH_SIZE {
		end := i + interfaces.GET_TOTAL_CONCEPTID_BATCH_SIZE
		if end > len(atIDs) {
			end = len(atIDs)
		}

		batchIDs := atIDs[i:end]
		batchTotal, err := ats.GetTotalWithATIDs(ctx, filterCondition, batchIDs)
		if err != nil {
			return 0, err
		}

		total += batchTotal
	}

	return total, nil
}

// 查询指定行动类ID列表的行动类总数
func (ats *actionTypeService) GetTotalWithATIDs(ctx context.Context,
	filterCondition map[string]any,
	atIDs []string) (int64, error) {

	// 构建包含 ATID 过滤的 filter condition
	atIDCondition := map[string]any{
		"field":      "id",
		"operation":  "in",
		"value":      atIDs,
		"value_from": "const",
	}

	var combinedCondition map[string]any
	if filterCondition == nil {
		combinedCondition = atIDCondition
	} else {
		combinedCondition = map[string]any{
			"operation": "and",
			"sub_conditions": []map[string]any{
				filterCondition,
				atIDCondition,
			},
		}
	}

	// 执行计数查询
	total, err := ats.GetTotal(ctx, combinedCondition)
	if err != nil {
		return total, err
	}

	return total, nil
}

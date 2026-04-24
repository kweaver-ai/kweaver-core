// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package object_type

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
	"bkn-backend/logics/permission"
	"bkn-backend/logics/user_mgmt"
)

var (
	otServiceOnce sync.Once
	otService     interfaces.ObjectTypeService
)

type objectTypeService struct {
	appSetting *common.AppSetting
	db         *sql.DB
	aoa        interfaces.AgentOperatorAccess
	cga        interfaces.ConceptGroupAccess
	dda        interfaces.DataModelAccess
	dva        interfaces.DataViewAccess
	mfa        interfaces.ModelFactoryAccess
	ota        interfaces.ObjectTypeAccess
	ps         interfaces.PermissionService
	ums        interfaces.UserMgmtService
	vba        interfaces.VegaBackendAccess
}

func NewObjectTypeService(appSetting *common.AppSetting) interfaces.ObjectTypeService {
	otServiceOnce.Do(func() {
		otService = &objectTypeService{
			appSetting: appSetting,
			db:         logics.DB,
			aoa:        logics.AOA,
			cga:        logics.CGA,
			dda:        logics.DDA,
			dva:        logics.DVA,
			mfa:        logics.MFA,
			ota:        logics.OTA,
			ps:         permission.NewPermissionService(appSetting),
			ums:        user_mgmt.NewUserMgmtService(appSetting),
			vba:        logics.VBA,
		}
	})
	return otService
}

// validateObjectTypeStrictExternalDeps checks backing data view or vega resource, vector embedding models, and logic property metric/operator references.
func (ots *objectTypeService) validateObjectTypeStrictExternalDeps(ctx context.Context, objectType *interfaces.ObjectType) error {
	if objectType.DataSource != nil && objectType.DataSource.ID != "" {
		dsType := objectType.DataSource.Type
		if dsType == "" {
			dsType = interfaces.DATA_SOURCE_TYPE_DATA_VIEW
		}
		switch dsType {
		case interfaces.DATA_SOURCE_TYPE_RESOURCE:
			res, err := ots.vba.GetResourceByID(ctx, objectType.DataSource.ID)
			if err != nil {
				return rest.NewHTTPError(ctx, http.StatusBadRequest,
					berrors.BknBackend_ObjectType_InvalidParameter).
					WithErrorDetails(fmt.Sprintf("对象类[%s]的资源[%s]获取失败: %s", objectType.OTName, objectType.DataSource.ID, err.Error()))
			}
			if res == nil {
				return rest.NewHTTPError(ctx, http.StatusBadRequest,
					berrors.BknBackend_ObjectType_InvalidParameter).
					WithErrorDetails(fmt.Sprintf("对象类[%s]的资源[%s]不存在", objectType.OTName, objectType.DataSource.ID))
			}
		default:
			dataView, err := ots.dva.GetDataViewByID(ctx, objectType.DataSource.ID)
			if err != nil {
				return rest.NewHTTPError(ctx, http.StatusBadRequest,
					berrors.BknBackend_ObjectType_InvalidParameter).
					WithErrorDetails(fmt.Sprintf("对象类[%s]的数据视图[%s]获取失败: %s", objectType.OTName, objectType.DataSource.ID, err.Error()))
			}
			if dataView == nil {
				return rest.NewHTTPError(ctx, http.StatusBadRequest,
					berrors.BknBackend_ObjectType_InvalidParameter).
					WithErrorDetails(fmt.Sprintf("对象类[%s]的数据视图[%s]不存在", objectType.OTName, objectType.DataSource.ID))
			}
		}
	}
	if objectType.DataProperties != nil {
		for _, prop := range objectType.DataProperties {
			if prop.IndexConfig != nil && prop.IndexConfig.VectorConfig.Enabled && prop.IndexConfig.VectorConfig.ModelID != "" {
				model, err := ots.mfa.GetModelByID(ctx, prop.IndexConfig.VectorConfig.ModelID)
				if err != nil {
					return rest.NewHTTPError(ctx, http.StatusBadRequest,
						berrors.BknBackend_ObjectType_InvalidParameter).
						WithErrorDetails(fmt.Sprintf("对象类[%s]属性[%s]的小模型[%s]获取失败: %s",
							objectType.OTName, prop.Name, prop.IndexConfig.VectorConfig.ModelID, err.Error()))
				}
				if model == nil {
					return rest.NewHTTPError(ctx, http.StatusBadRequest,
						berrors.BknBackend_ObjectType_InvalidParameter).
						WithErrorDetails(fmt.Sprintf("对象类[%s]属性[%s]的小模型[%s]不存在",
							objectType.OTName, prop.Name, prop.IndexConfig.VectorConfig.ModelID))
				}
				if model.ModelType != interfaces.SMALL_MODEL_TYPE_EMBEDDING {
					return rest.NewHTTPError(ctx, http.StatusBadRequest,
						berrors.BknBackend_ObjectType_InvalidParameter_SmallModel).
						WithErrorDetails(fmt.Sprintf("model type %s is not %s model", model.ModelType, interfaces.SMALL_MODEL_TYPE_EMBEDDING))
				}
				if model.EmbeddingDim == 0 || model.BatchSize == 0 || model.MaxTokens == 0 {
					return rest.NewHTTPError(ctx, http.StatusBadRequest,
						berrors.BknBackend_ObjectType_InvalidParameter_SmallModel).
						WithErrorDetails(fmt.Sprintf("model %s has invalid embedding dim, batch size or max tokens", model.ModelID))
				}
			}
		}
	}
	// Schema for logic properties (type, data_source) is validated in driveradapters.ValidateObjectType.
	for _, lp := range objectType.LogicProperties {
		switch lp.Type {
		case interfaces.LOGIC_PROPERTY_TYPE_METRIC:
			model, err := ots.dda.GetMetricModelByID(ctx, lp.DataSource.ID)
			if err != nil {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ObjectType_InvalidParameter).
					WithErrorDetails(fmt.Sprintf("对象类[%s]逻辑属性[%s]的指标模型[%s]获取失败: %s",
						objectType.OTName, lp.Name, lp.DataSource.ID, err.Error()))
			}
			if model == nil {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ObjectType_InvalidParameter).
					WithErrorDetails(fmt.Sprintf("对象类[%s]逻辑属性[%s]的指标模型[%s]不存在",
						objectType.OTName, lp.Name, lp.DataSource.ID))
			}
		case interfaces.LOGIC_PROPERTY_TYPE_OPERATOR:
			op, err := ots.aoa.GetAgentOperatorByID(ctx, lp.DataSource.ID)
			if err != nil {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ObjectType_InvalidParameter).
					WithErrorDetails(fmt.Sprintf("对象类[%s]逻辑属性[%s]的算子[%s]获取失败: %s",
						objectType.OTName, lp.Name, lp.DataSource.ID, err.Error()))
			}
			if op.OperatorId == "" {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ObjectType_InvalidParameter).
					WithErrorDetails(fmt.Sprintf("对象类[%s]逻辑属性[%s]的算子[%s]不存在",
						objectType.OTName, lp.Name, lp.DataSource.ID))
			}
		}
	}
	return nil
}

func (ots *objectTypeService) CheckObjectTypeExistByID(ctx context.Context,
	knID string, branch string, otID string) (string, bool, error) {

	ctx, span := oteltrace.StartNamedInternalSpan(ctx, fmt.Sprintf("校验对象类[%s]的存在性", otID))
	defer span.End()

	otName, exist, err := ots.ota.CheckObjectTypeExistByID(ctx, knID, branch, otID)
	if err != nil {
		logger.Errorf("CheckObjectTypeExistByID error: %s", err.Error())
		// 记录处理的 sql 字符串
		otellog.LogError(ctx, fmt.Sprintf("在业务知识网络[%s]下按ID[%s]获取对象类失败", knID, otID), err)
		span.SetStatus(codes.Error, fmt.Sprintf("在业务知识网络[%s]下按ID[%s]获取对象类失败", knID, otID))
		return "", exist, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_ObjectType_InternalError_CheckObjectTypeIfExistFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return otName, exist, nil
}

func (ots *objectTypeService) CheckObjectTypeExistByName(ctx context.Context,
	knID string, branch string, otName string) (string, bool, error) {

	ctx, span := oteltrace.StartNamedInternalSpan(ctx, fmt.Sprintf("校验对象类[%s]的存在性", otName))
	defer span.End()

	otID, exist, err := ots.ota.CheckObjectTypeExistByName(ctx, knID, branch, otName)
	if err != nil {
		logger.Errorf("CheckObjectTypeExistByName error: %s", err.Error())
		// 记录处理的 sql 字符串
		otellog.LogError(ctx, fmt.Sprintf("在业务知识网络[%s]下按名称[%s]获取对象类失败", knID, otName), err)
		span.SetStatus(codes.Error, fmt.Sprintf("在业务知识网络[%s]下按名称[%s]获取对象类失败", knID, otName))
		return otID, exist, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_ObjectType_InternalError_CheckObjectTypeIfExistFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return otID, exist, nil
}

func (ots *objectTypeService) CreateObjectTypes(ctx context.Context, tx *sql.Tx,
	objectTypes []*interfaces.ObjectType, mode string, needCreateConceptGroupRelation bool, strictMode bool) ([]string, error) {

	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "Create object type")
	defer span.End()

	// 判断userid是否有修改业务知识网络的权限
	err := ots.ps.CheckPermission(ctx, interfaces.PermissionResource{
		Type: interfaces.RESOURCE_TYPE_KN,
		ID:   objectTypes[0].KNID,
	}, []string{interfaces.OPERATION_TYPE_MODIFY})
	if err != nil {
		return []string{}, err
	}

	currentTime := time.Now().UnixMilli()
	for _, objectType := range objectTypes {
		// 若提交的模型id为空，生成分布式ID
		if objectType.OTID == "" {
			objectType.OTID = xid.New().String()
		}

		accountInfo := interfaces.AccountInfo{}
		if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
			accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
		}
		objectType.Creator = accountInfo
		objectType.Updater = accountInfo

		objectType.CreateTime = currentTime
		objectType.UpdateTime = currentTime

		if strictMode {
			if err := ots.validateObjectTypeStrictExternalDeps(ctx, objectType); err != nil {
				return []string{}, err
			}
		}

		bknObj := logics.ToBKNObjectType(objectType)
		objectType.BKNRawContent = bknsdk.SerializeObjectType(bknObj)
	}

	// 0. 开始事务
	if tx == nil {
		tx, err = ots.db.Begin()
		if err != nil {
			logger.Errorf("Begin transaction error: %s", err.Error())
			span.SetStatus(codes.Error, "事务开启失败")
			otellog.LogError(ctx, "Begin transaction error", err)
			return []string{}, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				berrors.BknBackend_ObjectType_InternalError_BeginTransactionFailed).
				WithErrorDetails(err.Error())
		}
		// 0.1 异常时
		defer func() {
			switch err {
			case nil:
				// 提交事务
				err = tx.Commit()
				if err != nil {
					logger.Errorf("CreateObjectType Transaction Commit Failed:%v", err)
					span.SetStatus(codes.Error, "提交事务失败")
					otellog.LogError(ctx, "CreateObjectType Transaction Commit Failed", err)
					return
				}
				logger.Infof("CreateObjectType Transaction Commit Success")
				otellog.LogDebug(ctx, "CreateObjectType Transaction Commit Success")
			default:
				rollbackErr := tx.Rollback()
				if rollbackErr != nil {
					logger.Errorf("CreateObjectType Transaction Rollback Error:%v", rollbackErr)
					span.SetStatus(codes.Error, "事务回滚失败")
					otellog.LogError(ctx, "CreateObjectType Transaction Rollback Error", err)
				}
			}
		}()
	}

	createObjectTypes, updateObjectTypes, err := ots.handleObjectTypeImportMode(ctx, mode, objectTypes)
	if err != nil {
		return []string{}, err
	}

	// 创建
	otIDs := []string{}
	for _, objectType := range createObjectTypes {
		otIDs = append(otIDs, objectType.OTID)
		err = ots.ota.CreateObjectType(ctx, tx, objectType)
		if err != nil {
			logger.Errorf("CreateObjectType error: %s", err.Error())
			span.SetStatus(codes.Error, "创建对象类失败")

			return []string{}, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				berrors.BknBackend_ObjectType_InternalError).
				WithErrorDetails(err.Error())
		}

		err = ots.ota.CreateObjectTypeStatus(ctx, tx, objectType)
		if err != nil {
			logger.Errorf("CreateObjectTypeStatus error: %s", err.Error())
			span.SetStatus(codes.Error, "创建对象类状态失败")

			return []string{}, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				berrors.BknBackend_ObjectType_InternalError).
				WithErrorDetails(err.Error())
		}

		// 按需建立对象类到各个组的关系
		if needCreateConceptGroupRelation {
			// 建立对象类到各个组的关系，已经存在的关系就不需要建立，需要先获取一下对象类与组的关系
			if len(objectType.ConceptGroups) > 0 {
				err = ots.handleGroupRelations(ctx, tx, objectType, currentTime, strictMode)
				if err != nil {
					span.SetStatus(codes.Error, "处理对象类与分组的关系失败")
					return []string{}, err
				}
			}
		}
	}

	// 更新
	for _, objectType := range updateObjectTypes {
		err = ots.UpdateObjectType(ctx, tx, objectType, strictMode)
		if err != nil {
			return []string{}, err
		}
	}

	insetObjectTypes := createObjectTypes
	insetObjectTypes = append(insetObjectTypes, updateObjectTypes...)
	err = ots.InsertDatasetData(ctx, insetObjectTypes)
	if err != nil {
		logger.Errorf("InsertDatasetData error: %s", err.Error())
		span.SetStatus(codes.Error, "对象类索引写入失败")

		return []string{}, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_ObjectType_InternalError_InsertOpenSearchDataFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return otIDs, nil
}

// ValidateObjectTypes checks dependency existence only; does not write to the database.
func (ots *objectTypeService) ValidateObjectTypes(ctx context.Context, knID string, branch string,
	objectTypes []*interfaces.ObjectType, strictMode bool, batch *interfaces.BatchIDIndex, mode string) error {

	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "ValidateObjectTypes")
	defer span.End()

	if len(objectTypes) == 0 {
		return nil
	}

	err := ots.ps.CheckPermission(ctx, interfaces.PermissionResource{
		Type: interfaces.RESOURCE_TYPE_KN,
		ID:   knID,
	}, []string{interfaces.OPERATION_TYPE_MODIFY})
	if err != nil {
		return err
	}

	_, _, err = ots.handleObjectTypeImportMode(ctx, mode, objectTypes)
	if err != nil {
		return err
	}

	for _, objectType := range objectTypes {
		objectType.KNID = knID
		objectType.Branch = branch
		if strictMode {
			if err := ots.validateObjectTypeStrictExternalDeps(ctx, objectType); err != nil {
				return err
			}

			// 校验概念分组存在性；batch 中含糊的同批分组 ID 视为将创建，跳过查库
			if len(objectType.ConceptGroups) > 0 {
				cgIDs := []string{}
				for _, cg := range objectType.ConceptGroups {
					cgIDs = append(cgIDs, cg.CGID)
				}
				cgIDs = common.DuplicateSlice(cgIDs)

				var needDBLookup []string
				for _, id := range cgIDs {
					if batch != nil && batchindex.HasConceptGroupID(id, batch) {
						continue
					}
					needDBLookup = append(needDBLookup, id)
				}
				if len(needDBLookup) == 0 {
					continue
				}

				tx, err := ots.db.Begin()
				if err != nil {
					return rest.NewHTTPError(ctx, http.StatusInternalServerError,
						berrors.BknBackend_ObjectType_InternalError_BeginTransactionFailed).
						WithErrorDetails(err.Error())
				}
				defer func() { _ = tx.Rollback() }()

				conceptGroups, err := ots.cga.GetConceptGroupsByIDs(ctx, tx, knID, branch, needDBLookup)
				if err != nil {
					return rest.NewHTTPError(ctx, http.StatusInternalServerError,
						berrors.BknBackend_ObjectType_InternalError).
						WithErrorDetails(fmt.Sprintf("GetConceptGroupsByIDs failed: %s", err.Error()))
				}
				if len(conceptGroups) != len(needDBLookup) {
					return rest.NewHTTPError(ctx, http.StatusBadRequest,
						berrors.BknBackend_ObjectType_InvalidParameter).
						WithErrorDetails(fmt.Sprintf("Exists any concept group not found, expect [%d], actual [%d]", len(needDBLookup), len(conceptGroups)))
				}
			}
		}
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

func (ots *objectTypeService) ListObjectTypes(ctx context.Context, tx *sql.Tx,
	query interfaces.ObjectTypesQueryParams) ([]*interfaces.ObjectType, int, error) {

	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "查询对象类列表")
	defer span.End()

	// 判断userid是否有查看业务知识网络的权限
	err := ots.ps.CheckPermission(ctx, interfaces.PermissionResource{
		Type: interfaces.RESOURCE_TYPE_KN,
		ID:   query.KNID,
	}, []string{interfaces.OPERATION_TYPE_VIEW_DETAIL})
	if err != nil {
		return []*interfaces.ObjectType{}, 0, err
	}

	// 0. 开始事务
	if tx == nil {
		tx, err = ots.db.Begin()
		if err != nil {
			logger.Errorf("Begin transaction error: %s", err.Error())
			span.SetStatus(codes.Error, "事务开启失败")
			otellog.LogError(ctx, "Begin transaction error", err)
			return []*interfaces.ObjectType{}, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				berrors.BknBackend_ObjectType_InternalError_BeginTransactionFailed).
				WithErrorDetails(err.Error())
		}
		// 0.1 异常时
		defer func() {
			switch err {
			case nil:
				// 提交事务
				err = tx.Commit()
				if err != nil {
					logger.Errorf("ListObjectTypes Transaction Commit Failed:%v", err)
					span.SetStatus(codes.Error, "提交事务失败")
					otellog.LogError(ctx, "ListObjectTypes Transaction Commit Failed", err)
					return
				}
				logger.Infof("ListObjectTypes Transaction Commit Success")
				otellog.LogDebug(ctx, "ListObjectTypes Transaction Commit Success")
			default:
				rollbackErr := tx.Rollback()
				if rollbackErr != nil {
					logger.Errorf("ListObjectTypes Transaction Rollback Error:%v", rollbackErr)
					span.SetStatus(codes.Error, "事务回滚失败")
					otellog.LogError(ctx, "ListObjectTypes Transaction Rollback Error", err)
				}
			}
		}()
	}

	//获取对象类列表
	objectTypes, err := ots.ota.ListObjectTypes(ctx, tx, query)
	if err != nil {
		logger.Errorf("ListObjectTypes error: %s", err.Error())
		span.SetStatus(codes.Error, "List object types error")

		return []*interfaces.ObjectType{}, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_ObjectType_InternalError).WithErrorDetails(err.Error())
	}
	if len(objectTypes) == 0 {
		span.SetStatus(codes.Ok, "")
		return objectTypes, 0, nil
	}

	total := len(objectTypes)
	// limit = -1,则返回所有
	if query.Limit != -1 {

		// 分页
		// 检查起始位置是否越界
		if query.Offset < 0 || query.Offset >= len(objectTypes) {
			span.SetStatus(codes.Ok, "")
			return []*interfaces.ObjectType{}, total, nil
		}
		// 计算结束位置
		end := query.Offset + query.Limit
		if end > len(objectTypes) {
			end = len(objectTypes)
		}
		objectTypes = objectTypes[query.Offset:end]
	}

	accountInfos := make([]*interfaces.AccountInfo, 0, len(objectTypes)*2)
	for _, objectType := range objectTypes {
		accountInfos = append(accountInfos, &objectType.Creator, &objectType.Updater)
	}

	err = ots.ums.GetAccountNames(ctx, accountInfos)
	if err != nil {
		span.SetStatus(codes.Error, "GetAccountNames error")

		return []*interfaces.ObjectType{}, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_ObjectType_InternalError).WithErrorDetails(err.Error())
	}

	// 获取对象类所属的分组 -- 注掉，不显示分组信息
	// otGroups, err := ots.cga.GetConceptGroupsByOTIDs(ctx, tx, interfaces.ConceptGroupRelationsQueryParams{
	// 	KNID:   query.KNID,
	// 	Branch: query.Branch,
	// 	OTIDs:  otIDs,
	// })
	// if err != nil {
	// 	span.SetStatus(codes.Error, "GetConceptGroupsByOTIDs error")

	// 	return []*interfaces.ObjectType{}, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError,
	// 		berrors.BknBackend_ObjectType_InternalError).WithErrorDetails(err.Error())
	// }

	// 注掉，不显示视图信息和属性映射字段的显示名
	// for _, objectType := range objectTypes {
	// 	// 获取视图字段的显示名
	// 	if objectType.DataSource != nil && objectType.DataSource.ID != "" {
	// 		dataView, err := ots.dva.GetDataViewByID(ctx, objectType.DataSource.ID)
	// 		if err != nil {
	// 			return []*interfaces.ObjectType{}, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError,
	// 				berrors.BknBackend_ObjectType_InternalError_GetDataViewByIDFailed).
	// 				WithErrorDetails(err.Error())
	// 		}
	// 		if dataView == nil {
	// 			otellog.LogWarn(ctx, fmt.Sprintf("Object type [%s]'s Data view %s not found", objectType.OTID, objectType.DataSource.ID))
	// 		} else {
	// 			objectType.DataSource.Name = dataView.ViewName
	// 			// 翻译数据属性映射的字段显示名
	// 			for j, prop := range objectType.DataProperties {
	// 				// 不为空时，才翻译字段显示名。为空则不翻译
	// 				if prop.MappedField != nil {
	// 					if field, exists := dataView.FieldsMap[prop.MappedField.Name]; exists {
	// 						objectType.DataProperties[j].MappedField.DisplayName = field.DisplayName
	// 						objectType.DataProperties[j].MappedField.Type = field.Type
	// 					}
	// 				}
	// 				// 字符串类型的属性支持的操作符返回
	// 				objectType.DataProperties[j].ConditionOperations = ots.processConditionOperations(objectType, prop, dataView)
	// 			}
	// 		}
	// 	}

	// 	// 给对象类加上分组信息
	// 	objectType.ConceptGroups = otGroups[objectType.OTID]
	// }

	span.SetStatus(codes.Ok, "")
	return objectTypes, total, nil
}

func (ots *objectTypeService) GetObjectTypesByIDs(ctx context.Context, tx *sql.Tx,
	knID string, branch string, otIDs []string) ([]*interfaces.ObjectType, error) {
	// 获取对象类
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, fmt.Sprintf("查询对象类[%s]信息", otIDs))
	defer span.End()

	// 判断userid是否有查看业务知识网络的权限
	err := ots.ps.CheckPermission(ctx, interfaces.PermissionResource{
		Type: interfaces.RESOURCE_TYPE_KN,
		ID:   knID,
	}, []string{interfaces.OPERATION_TYPE_VIEW_DETAIL})
	if err != nil {
		return []*interfaces.ObjectType{}, err
	}

	// 0. 开始事务
	if tx == nil {
		tx, err = ots.db.Begin()
		if err != nil {
			logger.Errorf("Begin transaction error: %s", err.Error())
			span.SetStatus(codes.Error, "事务开启失败")
			otellog.LogError(ctx, "Begin transaction error", err)
			return []*interfaces.ObjectType{}, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				berrors.BknBackend_ObjectType_InternalError_BeginTransactionFailed).
				WithErrorDetails(err.Error())
		}
		// 0.1 异常时
		defer func() {
			switch err {
			case nil:
				// 提交事务
				err = tx.Commit()
				if err != nil {
					logger.Errorf("GetObjectTypes Transaction Commit Failed:%v", err)
					span.SetStatus(codes.Error, "提交事务失败")
					otellog.LogError(ctx, "GetObjectTypes Transaction Commit Failed", err)
					return
				}
				logger.Infof("GetObjectTypes Transaction Commit Success")
				otellog.LogDebug(ctx, "GetObjectTypes Transaction Commit Success")
			default:
				rollbackErr := tx.Rollback()
				if rollbackErr != nil {
					logger.Errorf("GetObjectTypes Transaction Rollback Error:%v", rollbackErr)
					span.SetStatus(codes.Error, "事务回滚失败")
					otellog.LogError(ctx, "GetObjectTypes Transaction Rollback Error", err)
				}
			}
		}()
	}

	// id去重后再查
	otIDs = common.DuplicateSlice(otIDs)

	// 获取对象类基本信息
	objectTypes, err := ots.ota.GetObjectTypesByIDs(ctx, tx, knID, branch, otIDs)
	if err != nil {
		logger.Errorf("GetObjectTypesByObjectTypeIDs error: %s", err.Error())
		span.SetStatus(codes.Error, fmt.Sprintf("Get object types[%s] error: %v", otIDs, err))

		return []*interfaces.ObjectType{}, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_ObjectType_InternalError_GetObjectTypesByIDsFailed).WithErrorDetails(err.Error())
	}

	if len(objectTypes) != len(otIDs) {
		errStr := fmt.Sprintf("Exists any object types not found, expect object types nums is [%d], actual object types num is [%d]", len(otIDs), len(objectTypes))
		logger.Errorf(errStr)
		span.SetStatus(codes.Error, errStr)

		return []*interfaces.ObjectType{}, rest.NewHTTPError(ctx, http.StatusNotFound,
			berrors.BknBackend_ObjectType_ObjectTypeNotFound).WithErrorDetails(errStr)
	}

	// 获取对象类所属的分组
	otGroups, err := ots.cga.GetConceptGroupsByOTIDs(ctx, tx, interfaces.ConceptGroupRelationsQueryParams{
		KNID:   knID,
		Branch: branch,
		OTIDs:  otIDs,
	})
	if err != nil {
		span.SetStatus(codes.Error, "GetConceptGroupsByOTIDs error")

		return []*interfaces.ObjectType{}, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_ObjectType_InternalError).WithErrorDetails(err.Error())
	}

	// 数据视图不为空时，需要把id转成名称
	// 请求视图
	for _, objectType := range objectTypes {
		// 处理数据源和操作符
		err = ots.processObjectTypeDetails(ctx, objectType)
		if err != nil {
			return []*interfaces.ObjectType{}, err
		}
		// 给对象类加上分组信息
		objectType.ConceptGroups = otGroups[objectType.OTID]
	}

	accountInfos := make([]*interfaces.AccountInfo, 0, len(objectTypes)*2)
	for _, objectType := range objectTypes {
		accountInfos = append(accountInfos, &objectType.Creator, &objectType.Updater)
	}

	err = ots.ums.GetAccountNames(ctx, accountInfos)
	if err != nil {
		span.SetStatus(codes.Error, "GetAccountNames error")

		return []*interfaces.ObjectType{}, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_ObjectType_InternalError).WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return objectTypes, nil
}

// hasDataPropertyIndexAffectingChanges 检测单个数据属性的关键字段是否发生变化
// 影响索引的字段包括：Name, Type, IndexConfig, MappedField.Name, MappedField.Type
func hasDataPropertyIndexAffectingChanges(oldProp, newProp *interfaces.DataProperty) bool {
	if oldProp == nil || newProp == nil {
		return oldProp != newProp
	}

	// 比较属性名称
	if oldProp.Name != newProp.Name {
		return true
	}

	// 比较属性类型
	if oldProp.Type != newProp.Type {
		return true
	}

	// 比较索引配置
	if !compareIndexConfig(oldProp.IndexConfig, newProp.IndexConfig) {
		return true // 如果配置不同，返回 true（有变化）
	}

	// 比较映射字段名称和类型
	if !compareMappedField(oldProp.MappedField, newProp.MappedField) {
		return true
	}

	return false
}

// compareIndexConfig 比较两个索引配置是否相同
func compareIndexConfig(oldConfig, newConfig *interfaces.IndexConfig) bool {
	if oldConfig == nil && newConfig == nil {
		return true // 都为空 = 状态相同（都没有配置）
	}
	if oldConfig == nil || newConfig == nil {
		return false // 一个为空一个不为空 = 状态不同
	}

	// 使用 JSON 序列化比较，确保准确性
	oldBytes, err := sonic.Marshal(oldConfig)
	if err != nil {
		return false
	}
	newBytes, err := sonic.Marshal(newConfig)
	if err != nil {
		return false
	}

	return string(oldBytes) == string(newBytes)
}

// compareMappedField 比较两个映射字段是否相同（只比较 Name 和 Type）
func compareMappedField(oldField, newField *interfaces.Field) bool {
	if oldField == nil && newField == nil {
		return true
	}
	if oldField == nil || newField == nil {
		return false
	}

	// 比较字段名称
	if oldField.Name != newField.Name {
		return false
	}

	// 比较字段类型
	if oldField.Type != newField.Type {
		return false
	}

	return true
}

// hasAnyDataPropertyIndexAffectingChanges 检测数据属性列表中是否有影响索引的变化
func hasAnyDataPropertyIndexAffectingChanges(oldProps, newProps []*interfaces.DataProperty) bool {
	// 将旧属性列表转换为以 Name 为 key 的 map
	oldPropMap := make(map[string]*interfaces.DataProperty)
	for _, prop := range oldProps {
		if prop != nil {
			oldPropMap[prop.Name] = prop
		}
	}

	// 遍历新属性列表，查找对应的旧属性进行比较
	for _, newProp := range newProps {
		if newProp == nil {
			continue
		}

		oldProp, exists := oldPropMap[newProp.Name]
		if !exists {
			// 新增属性可能影响索引
			return true
		}

		// 比较属性是否有影响索引的变化
		if hasDataPropertyIndexAffectingChanges(oldProp, newProp) {
			return true
		}

		// 从 map 中删除已比较的属性
		delete(oldPropMap, newProp.Name)
	}

	// 如果旧属性列表中有新列表不存在的属性，也可能影响索引（删除属性）
	if len(oldPropMap) > 0 {
		return true
	}

	return false
}

// 更新对象类
func (ots *objectTypeService) UpdateObjectType(ctx context.Context, tx *sql.Tx, objectType *interfaces.ObjectType, strictMode bool) error {

	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "Update object type")
	defer span.End()

	// 判断userid是否有修改业务知识网络的权限
	err := ots.ps.CheckPermission(ctx, interfaces.PermissionResource{
		Type: interfaces.RESOURCE_TYPE_KN,
		ID:   objectType.KNID,
	}, []string{interfaces.OPERATION_TYPE_MODIFY})
	if err != nil {
		return err
	}

	if strictMode {
		if err := ots.validateObjectTypeStrictExternalDeps(ctx, objectType); err != nil {
			return err
		}
	}

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}
	objectType.Updater = accountInfo

	currentTime := time.Now().UnixMilli() // 对象类的update_time是int类型
	objectType.UpdateTime = currentTime

	bknObj := logics.ToBKNObjectType(objectType)
	objectType.BKNRawContent = bknsdk.SerializeObjectType(bknObj)

	if tx == nil {
		// 0. 开始事务
		tx, err = ots.db.Begin()
		if err != nil {
			logger.Errorf("Begin transaction error: %s", err.Error())
			span.SetStatus(codes.Error, "事务开启失败")
			otellog.LogError(ctx, "Begin transaction error", err)

			return rest.NewHTTPError(ctx, http.StatusInternalServerError,
				berrors.BknBackend_ObjectType_InternalError_BeginTransactionFailed).
				WithErrorDetails(err.Error())
		}
		// 0.1 异常时
		defer func() {
			switch err {
			case nil:
				// 提交事务
				err = tx.Commit()
				if err != nil {
					logger.Errorf("UpdateObjectType Transaction Commit Failed:%v", err)
					span.SetStatus(codes.Error, "提交事务失败")
					otellog.LogError(ctx, "UpdateObjectType Transaction Commit Failed", err)
				}
				logger.Infof("UpdateObjectType Transaction Commit Success:%v", objectType.OTName)
				otellog.LogDebug(ctx, fmt.Sprintf("UpdateObjectType Transaction Commit Success: %s", objectType.OTName))
			default:
				rollbackErr := tx.Rollback()
				if rollbackErr != nil {
					logger.Errorf("UpdateObjectType Transaction Rollback Error:%v", rollbackErr)
					span.SetStatus(codes.Error, "事务回滚失败")
					otellog.LogError(ctx, "UpdateObjectType Transaction Rollback Error", rollbackErr)
				}
			}
		}()
	}

	// 获取旧的对象类数据，用于比较数据属性变化
	oldObjectType, err := ots.ota.GetObjectTypeByID(ctx, tx, objectType.KNID, objectType.Branch, objectType.OTID)
	if err != nil {
		logger.Errorf("GetObjectTypeByID error: %s", err.Error())
		span.SetStatus(codes.Error, "获取旧对象类数据失败")
		otellog.LogError(ctx, "GetObjectTypeByID error", err)

		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_ObjectType_InternalError_GetObjectTypeByIDFailed).
			WithErrorDetails(err.Error())
	}

	// 检测数据属性是否有影响索引的变化
	if oldObjectType != nil && hasAnyDataPropertyIndexAffectingChanges(oldObjectType.DataProperties, objectType.DataProperties) {
		// 更新索引状态为不可用
		otStatus := *oldObjectType.Status
		otStatus.IndexAvailable = false
		otStatus.UpdateTime = currentTime
		err = ots.ota.UpdateObjectTypeStatus(ctx, tx, objectType.KNID, objectType.Branch, objectType.OTID, otStatus)
		if err != nil {
			logger.Errorf("UpdateObjectTypeStatus error: %s", err.Error())
			span.SetStatus(codes.Error, "更新对象类索引状态失败")
			otellog.LogError(ctx, "UpdateObjectTypeStatus error", err)

			return rest.NewHTTPError(ctx, http.StatusInternalServerError,
				berrors.BknBackend_ObjectType_InternalError).
				WithErrorDetails(fmt.Sprintf("更新对象类索引状态失败: %s", err.Error()))
		}
		logger.Infof("数据属性变化影响索引，已将对象类[%s]的索引状态设置为不可用", objectType.OTID)
		otellog.LogInfo(ctx, fmt.Sprintf("数据属性变化影响索引，已将对象类[%s]的索引状态设置为不可用", objectType.OTID))
	}

	// 更新模型信息
	err = ots.ota.UpdateObjectType(ctx, tx, objectType)
	if err != nil {
		logger.Errorf("UpdateObjectType error: %s", err.Error())
		span.SetStatus(codes.Error, "修改对象类失败")

		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_ObjectType_InternalError).
			WithErrorDetails(err.Error())
	}

	// 4. 同步分组关系（全量替换）
	if err := ots.syncObjectGroups(ctx, tx, *objectType, currentTime, strictMode); err != nil {
		return err
	}

	err = ots.InsertDatasetData(ctx, []*interfaces.ObjectType{objectType})
	if err != nil {
		logger.Errorf("InsertDatasetData error: %s", err.Error())
		span.SetStatus(codes.Error, "对象类索引写入失败")

		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_ObjectType_InternalError_InsertOpenSearchDataFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 更新对象类数据属性
func (ots *objectTypeService) UpdateDataProperties(ctx context.Context,
	objectType *interfaces.ObjectType, dataProperties []*interfaces.DataProperty, strictMode bool) error {

	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "Update object type")
	defer span.End()

	// 判断userid是否有修改业务知识网络的权限
	err := ots.ps.CheckPermission(ctx, interfaces.PermissionResource{
		Type: interfaces.RESOURCE_TYPE_KN,
		ID:   objectType.KNID,
	}, []string{interfaces.OPERATION_TYPE_MODIFY})
	if err != nil {
		return err
	}

	// When strictMode is true, validate embedding small model for any submitted property with vector index enabled.
	if strictMode {
		for _, prop := range dataProperties {
			if prop.IndexConfig != nil && prop.IndexConfig.VectorConfig.Enabled {
				model, err := ots.mfa.GetModelByID(ctx, prop.IndexConfig.VectorConfig.ModelID)
				if err != nil {
					return rest.NewHTTPError(ctx, http.StatusInternalServerError,
						berrors.BknBackend_ObjectType_InternalError_GetSmallModelByIDFailed).
						WithErrorDetails(err.Error())
				}
				if model == nil {
					return rest.NewHTTPError(ctx, http.StatusNotFound,
						berrors.BknBackend_ObjectType_SmallModelNotFound).
						WithErrorDetails(fmt.Sprintf("small model %s not found", prop.IndexConfig.VectorConfig.ModelID))
				}
				if model.ModelType != interfaces.SMALL_MODEL_TYPE_EMBEDDING {
					return rest.NewHTTPError(ctx, http.StatusBadRequest,
						berrors.BknBackend_ObjectType_InvalidParameter_SmallModel).
						WithErrorDetails(fmt.Sprintf("small model type %s is not %s model", model.ModelType, interfaces.SMALL_MODEL_TYPE_EMBEDDING))
				}
				if model.EmbeddingDim == 0 || model.BatchSize == 0 || model.MaxTokens == 0 {
					return rest.NewHTTPError(ctx, http.StatusBadRequest,
						berrors.BknBackend_ObjectType_InvalidParameter_SmallModel).
						WithErrorDetails(fmt.Sprintf("small model %s has invalid embedding dim, batch size or max tokens", model.ModelID))
				}
			}
		}
	}

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}
	objectType.Updater = accountInfo
	currentTime := time.Now().UnixMilli() // 对象类的update_time是int类型
	objectType.UpdateTime = currentTime

	// 深拷贝旧的数据属性，用于后续比较
	oldDataPropertiesBytes, err := sonic.Marshal(objectType.DataProperties)
	if err != nil {
		logger.Errorf("Failed to marshal old DataProperties, err: %v", err.Error())
		span.SetStatus(codes.Error, "序列化旧数据属性失败")
		otellog.LogError(ctx, "Failed to marshal old DataProperties, err", err)

		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_ObjectType_InternalError).
			WithErrorDetails(fmt.Sprintf("序列化旧数据属性失败: %s", err.Error()))
	}

	var oldDataProperties []*interfaces.DataProperty
	err = sonic.Unmarshal(oldDataPropertiesBytes, &oldDataProperties)
	if err != nil {
		logger.Errorf("Failed to unmarshal old DataProperties, err: %v", err.Error())
		span.SetStatus(codes.Error, "反序列化旧数据属性失败")
		otellog.LogError(ctx, "Failed to unmarshal old DataProperties, err", err)

		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_ObjectType_InternalError).
			WithErrorDetails(fmt.Sprintf("反序列化旧数据属性失败: %s", err.Error()))
	}

	propMap := map[string]int{}
	for idx, prop := range objectType.DataProperties {
		propMap[prop.Name] = idx
	}
	for _, prop := range dataProperties {
		if idx, ok := propMap[prop.Name]; ok {
			objectType.DataProperties[idx] = prop // 更新已存在的数据属性
		} else {
			objectType.DataProperties = append(objectType.DataProperties, prop) // 添加新的数据属性
		}
	}

	bknObj := logics.ToBKNObjectType(objectType)
	objectType.BKNRawContent = bknsdk.SerializeObjectType(bknObj)

	// 0. 开始事务
	var tx *sql.Tx
	tx, err = ots.db.Begin()
	if err != nil {
		logger.Errorf("Begin transaction error: %s", err.Error())
		span.SetStatus(codes.Error, "事务开启失败")
		otellog.LogError(ctx, "Begin transaction error", err)

		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_ObjectType_InternalError_BeginTransactionFailed).
			WithErrorDetails(err.Error())
	}
	// 0.1 异常时
	defer func() {
		switch err {
		case nil:
			// 提交事务
			err = tx.Commit()
			if err != nil {
				logger.Errorf("UpdateObjectType Transaction Commit Failed:%v", err)
				span.SetStatus(codes.Error, "提交事务失败")
				otellog.LogError(ctx, "UpdateObjectType Transaction Commit Failed", err)
			}
			logger.Infof("UpdateObjectType Transaction Commit Success:%v", objectType.OTName)
			otellog.LogDebug(ctx, fmt.Sprintf("UpdateObjectType Transaction Commit Success: %s", objectType.OTName))
		default:
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				logger.Errorf("UpdateObjectType Transaction Rollback Error:%v", rollbackErr)
				span.SetStatus(codes.Error, "事务回滚失败")
				otellog.LogError(ctx, "UpdateObjectType Transaction Rollback Error", rollbackErr)
			}
		}
	}()

	// 检测数据属性是否有影响索引的变化
	if hasAnyDataPropertyIndexAffectingChanges(oldDataProperties, objectType.DataProperties) {
		// 更新索引状态为不可用
		if objectType.Status != nil {
			otStatus := *objectType.Status
			otStatus.IndexAvailable = false
			otStatus.UpdateTime = currentTime
			// UpdateDataProperties 方法没有 tx 参数，需要在内部管理事务
			// 但为了保持一致性，我们使用 db.Exec 直接执行
			err = ots.ota.UpdateObjectTypeStatus(ctx, tx, objectType.KNID, objectType.Branch, objectType.OTID, otStatus)
			if err != nil {
				logger.Errorf("UpdateObjectTypeStatus error: %s", err.Error())
				span.SetStatus(codes.Error, "更新对象类索引状态失败")
				otellog.LogError(ctx, "UpdateObjectTypeStatus error", err)

				return rest.NewHTTPError(ctx, http.StatusInternalServerError,
					berrors.BknBackend_ObjectType_InternalError).
					WithErrorDetails(fmt.Sprintf("更新对象类索引状态失败: %s", err.Error()))
			}
			logger.Infof("数据属性变化影响索引，已将对象类[%s]的索引状态设置为不可用", objectType.OTID)
			otellog.LogInfo(ctx, fmt.Sprintf("数据属性变化影响索引，已将对象类[%s]的索引状态设置为不可用", objectType.OTID))
		}
	}

	// 更新模型信息
	err = ots.ota.UpdateDataProperties(ctx, tx, objectType)
	if err != nil {
		logger.Errorf("UpdateObjectType error: %s", err.Error())
		span.SetStatus(codes.Error, "修改对象类失败")

		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_ObjectType_InternalError).
			WithErrorDetails(err.Error())
	}

	err = ots.InsertDatasetData(ctx, []*interfaces.ObjectType{objectType})
	if err != nil {
		logger.Errorf("InsertDatasetData error: %s", err.Error())
		span.SetStatus(codes.Error, "对象类索引写入失败")

		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_ObjectType_InternalError_InsertOpenSearchDataFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

func (ots *objectTypeService) DeleteObjectTypesByIDs(ctx context.Context, tx *sql.Tx, knID string, branch string, otIDs []string) error {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "Delete object types")
	defer span.End()

	// 判断userid是否有修改业务知识网络的权限
	err := ots.ps.CheckPermission(ctx, interfaces.PermissionResource{
		Type: interfaces.RESOURCE_TYPE_KN,
		ID:   knID,
	}, []string{interfaces.OPERATION_TYPE_MODIFY})
	if err != nil {
		return err
	}

	if tx == nil {
		// 0. 开始事务
		tx, err = ots.db.Begin()
		if err != nil {
			logger.Errorf("Begin transaction error: %s", err.Error())
			span.SetStatus(codes.Error, "事务开启失败")
			otellog.LogError(ctx, "Begin transaction error", err)

			return rest.NewHTTPError(ctx, http.StatusInternalServerError,
				berrors.BknBackend_ObjectType_InternalError_BeginTransactionFailed).
				WithErrorDetails(err.Error())
		}
		// 0.1 异常时
		defer func() {
			switch err {
			case nil:
				// 提交事务
				err = tx.Commit()
				if err != nil {
					logger.Errorf("DeleteObjectTypes Transaction Commit Failed:%v", err)
					span.SetStatus(codes.Error, "提交事务失败")
					otellog.LogError(ctx, "DeleteObjectTypes Transaction Commit Failed", err)
				}
				logger.Infof("DeleteObjectTypes Transaction Commit Success: kn_id:%s,ot_ids:%v", knID, otIDs)
				otellog.LogDebug(ctx, fmt.Sprintf("DeleteObjectTypes Transaction Commit Success: kn_id:%s,ot_ids:%v", knID, otIDs))
			default:
				rollbackErr := tx.Rollback()
				if rollbackErr != nil {
					logger.Errorf("DeleteObjectTypes Transaction Rollback Error:%v", rollbackErr)
					span.SetStatus(codes.Error, "事务回滚失败")
					otellog.LogError(ctx, "DeleteObjectTypes Transaction Rollback Error", rollbackErr)
				}
			}
		}()
	}

	// 删除对象类
	rowsAffect, err := ots.ota.DeleteObjectTypesByIDs(ctx, tx, knID, branch, otIDs)
	if err != nil {
		logger.Errorf("DeleteObjectTypes error: %s", err.Error())
		span.SetStatus(codes.Error, "删除对象类失败")

		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_ObjectType_InternalError).WithErrorDetails(err.Error())
	}

	logger.Infof("DeleteObjectTypes: Rows affected is %v, request delete ObjectTypeIDs is %v!", rowsAffect, len(otIDs))
	if rowsAffect != int64(len(otIDs)) {
		logger.Warnf("Delete object types number %v not equal requerst object types number %v!", rowsAffect, len(otIDs))
		otellog.LogWarn(ctx, fmt.Sprintf("Delete object types number %v not equal requerst object types number %v!", rowsAffect, len(otIDs)))
	}

	rowsAffect, err = ots.ota.DeleteObjectTypeStatusByIDs(ctx, tx, knID, branch, otIDs)
	if err != nil {
		logger.Errorf("DeleteObjectTypeStatusByIDs error: %s", err.Error())
		span.SetStatus(codes.Error, "删除对象类状态失败")

		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_ObjectType_InternalError).WithErrorDetails(err.Error())
	}

	// 记录info日志，删除的条数
	logger.Infof("DeleteObjectTypeStatusByIDs success, the kn_id is [%s], branch is [%s], ot_ids is [%v], rowsAffect is [%d]",
		knID, branch, otIDs, rowsAffect)

	for _, otID := range otIDs {
		docid := interfaces.GenerateConceptDocuemtnID(knID, interfaces.MODULE_TYPE_OBJECT_TYPE, otID, branch)
		err = ots.vba.DeleteDatasetDocumentByID(ctx, interfaces.BKN_DATASET_ID, docid)
		if err != nil {
			logger.Errorf("DeleteDatasetDocumentByID error: %s", err.Error())
			span.SetStatus(codes.Error, "删除对象类概念索引失败")
			return err
		}
	}

	// 从概念与分组的关系表中删除该对象所建立的关系
	// 删除对象类与分组的绑定关系
	rowsAffect, err = ots.cga.DeleteObjectTypesFromGroup(ctx, tx, interfaces.ConceptGroupRelationsQueryParams{
		KNID:        knID,
		Branch:      branch,
		ConceptType: interfaces.MODULE_TYPE_OBJECT_TYPE,
		OTIDs:       otIDs,
	})
	if err != nil {
		errStr := fmt.Sprintf("DeleteObjectTypesFromGroup failed, the kn_id is [%s], branch is [%s], ot_ids is [%v], error is [%s]",
			knID, "branch", otIDs, err.Error())
		logger.Errorf(errStr)

		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_ObjectType_InternalError).
			WithErrorDetails(errStr)
	}
	// 记录info日志，删除的条数
	logger.Infof("DeleteObjectTypesFromGroup success, the kn_id is [%s], branch is [%s], ot_ids is [%v], rowsAffect is [%d]",
		knID, branch, otIDs, rowsAffect)

	span.SetStatus(codes.Ok, "")
	return nil
}

// 内部方法，删除对象类与状态，不检查权限，tx必须传入
func (ots *objectTypeService) DeleteObjectTypesByKnID(ctx context.Context, tx *sql.Tx, knID string, branch string) error {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "Delete object types")
	defer span.End()

	if tx == nil {
		logger.Errorf("missing transaction")
		otellog.LogError(ctx, "missing transaction", nil)
		span.SetStatus(codes.Error, "缺少事务")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_ObjectType_InternalError_BeginTransactionFailed).
			WithErrorDetails("missing transaction")
	}

	// 删除对象类
	rowsAffect, err := ots.ota.DeleteObjectTypesByKnID(ctx, tx, knID, branch)
	if err != nil {
		logger.Errorf("DeleteObjectTypes error: %s", err.Error())
		span.SetStatus(codes.Error, "删除对象类失败")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_ObjectType_InternalError).WithErrorDetails(err.Error())
	}

	logger.Infof("DeleteObjectTypes: Rows affected is %v!", rowsAffect)
	rowsAffect, err = ots.ota.DeleteObjectTypeStatusByKnID(ctx, tx, knID, branch)
	if err != nil {
		logger.Errorf("DeleteObjectTypeStatusByIDs error: %s", err.Error())
		span.SetStatus(codes.Error, "删除对象类状态失败")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_ObjectType_InternalError).WithErrorDetails(err.Error())
	}

	// 记录info日志，删除的条数
	logger.Infof("DeleteObjectTypesByKnID success, the kn_id is [%s], branch is [%s], rowsAffect is [%d]",
		knID, branch, rowsAffect)
	span.SetStatus(codes.Ok, "")
	return nil
}

func (ots *objectTypeService) handleObjectTypeImportMode(ctx context.Context, mode string,
	objectTypes []*interfaces.ObjectType) ([]*interfaces.ObjectType, []*interfaces.ObjectType, error) {

	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "object type import mode logic")
	defer span.End()

	creates := []*interfaces.ObjectType{}
	updates := []*interfaces.ObjectType{}

	// 3. 校验 若模型的id不为空，则用请求体的id与现有模型ID的重复性
	for _, objectType := range objectTypes {
		creates = append(creates, objectType)
		idExist := false
		_, idExist, err := ots.CheckObjectTypeExistByID(ctx, objectType.KNID, objectType.Branch, objectType.OTID)
		if err != nil {
			return creates, updates, err
		}

		// 校验 请求体与现有模型名称的重复性
		existID, nameExist, err := ots.CheckObjectTypeExistByName(ctx, objectType.KNID, objectType.Branch, objectType.OTName)
		if err != nil {
			return creates, updates, err
		}

		// 根据mode来区别，若是ignore，就从结果集中忽略，若是overwrite，就调用update，若是normal就报错。
		if idExist || nameExist {
			switch mode {
			case interfaces.ImportMode_Normal:
				if idExist {
					errDetails := fmt.Sprintf("The object type with id [%s] already exists!", objectType.OTID)
					logger.Error(errDetails)
					span.SetStatus(codes.Error, errDetails)
					return creates, updates, rest.NewHTTPError(ctx, http.StatusBadRequest,
						berrors.BknBackend_ObjectType_ObjectTypeIDExisted).
						WithErrorDetails(errDetails)
				}

				if nameExist {
					errDetails := fmt.Sprintf("object type name '%s' already exists", objectType.OTName)
					logger.Error(errDetails)
					span.SetStatus(codes.Error, errDetails)
					return creates, updates, rest.NewHTTPError(ctx, http.StatusForbidden,
						berrors.BknBackend_ObjectType_ObjectTypeNameExisted).
						WithDescription(map[string]any{"name": objectType.OTName}).
						WithErrorDetails(errDetails)
				}

			case interfaces.ImportMode_Ignore:
				// 存在重复的就跳过
				// 从create数组中删除
				creates = creates[:len(creates)-1]
			case interfaces.ImportMode_Overwrite:
				if idExist && nameExist {
					// 如果 id 和名称都存在，但是存在的名称对应的视图 id 和当前视图 id 不一样，则报错
					if existID != objectType.OTID {
						errDetails := fmt.Sprintf("ObjectType ID '%s' and name '%s' already exist, but the exist object type id is '%s'",
							objectType.OTID, objectType.OTName, existID)
						logger.Error(errDetails)
						span.SetStatus(codes.Error, errDetails)
						return creates, updates, rest.NewHTTPError(ctx, http.StatusForbidden,
							berrors.BknBackend_ObjectType_ObjectTypeNameExisted).
							WithErrorDetails(errDetails)
					} else {
						// 如果 id 和名称、度量名称都存在，存在的名称对应的模型 id 和当前模型 id 一样，则覆盖更新
						// 从create数组中删除, 放到更新数组中
						creates = creates[:len(creates)-1]
						updates = append(updates, objectType)
					}
				}

				// id 已存在，且名称不存在，覆盖更新
				if idExist && !nameExist {
					// 从create数组中删除, 放到更新数组中
					creates = creates[:len(creates)-1]
					updates = append(updates, objectType)
				}

				// 如果 id 不存在，name 存在，报错
				if !idExist && nameExist {
					errDetails := fmt.Sprintf("ObjectType ID '%s' does not exist, but name '%s' already exists",
						objectType.OTID, objectType.OTName)
					logger.Error(errDetails)
					span.SetStatus(codes.Error, errDetails)
					return creates, updates, rest.NewHTTPError(ctx, http.StatusForbidden,
						berrors.BknBackend_ObjectType_ObjectTypeNameExisted).
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

// 内部使用，无需校验权限
func (ots *objectTypeService) GetObjectTypesMapByIDs(ctx context.Context, knID string,
	branch string, otIDs []string, needPropMap bool) (map[string]*interfaces.ObjectType, error) {
	// 获取对象类
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, fmt.Sprintf("查询对象类[%v]信息", otIDs))
	defer span.End()

	// 判断userid是否有修改业务知识网络的权限
	err := ots.ps.CheckPermission(ctx, interfaces.PermissionResource{
		Type: interfaces.RESOURCE_TYPE_KN,
		ID:   knID,
	}, []string{interfaces.OPERATION_TYPE_VIEW_DETAIL})
	if err != nil {
		return map[string]*interfaces.ObjectType{}, err
	}

	// id去重后再查
	otIDs = common.DuplicateSlice(otIDs)

	// 获取模型基本信息
	objectTypeArr, err := ots.ota.GetObjectTypesByIDs(ctx, nil, knID, branch, otIDs)
	if err != nil {
		logger.Errorf("GetObjectTypesByObjectTypeIDs error: %s", err.Error())
		span.SetStatus(codes.Error, fmt.Sprintf("Get object type[%v] error: %v", otIDs, err))
		return map[string]*interfaces.ObjectType{}, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_ObjectType_InternalError_GetObjectTypesByIDsFailed).
			WithErrorDetails(err.Error())
	}

	objectTypeMap := map[string]*interfaces.ObjectType{}
	for _, object := range objectTypeArr {
		if needPropMap {
			propMap := map[string]string{}
			for _, prop := range object.DataProperties {
				propMap[prop.Name] = prop.DisplayName
			}
			object.PropertyMap = propMap
		}
		objectTypeMap[object.OTID] = object
	}

	span.SetStatus(codes.Ok, "")
	return objectTypeMap, nil
}

func (ots *objectTypeService) InsertDatasetData(ctx context.Context, objectTypes []*interfaces.ObjectType) error {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "对象类索引写入")
	defer span.End()

	if len(objectTypes) == 0 {
		return nil
	}

	if ots.appSetting.ServerSetting.DefaultSmallModelEnabled {
		words := []string{}
		for _, objectType := range objectTypes {
			arr := []string{objectType.OTName}
			arr = append(arr, objectType.Tags...)
			arr = append(arr, objectType.Comment, objectType.BKNRawContent)
			word := strings.Join(arr, "\n")
			words = append(words, word)
		}

		dftModel, err := ots.mfa.GetDefaultModel(ctx)
		if err != nil {
			logger.Errorf("GetDefaultModel error: %s", err.Error())
			span.SetStatus(codes.Error, "获取默认模型失败")
			return err
		}
		vectors, err := ots.mfa.GetVector(ctx, dftModel, words)
		if err != nil {
			logger.Errorf("GetVector error: %s", err.Error())
			span.SetStatus(codes.Error, "获取业务知识网络向量失败")
			return err
		}

		if len(vectors) != len(objectTypes) {
			logger.Errorf("GetVector error: expect vectors num is [%d], actual vectors num is [%d]", len(objectTypes), len(vectors))
			span.SetStatus(codes.Error, "获取业务知识网络向量失败")
			return fmt.Errorf("GetVector error: expect vectors num is [%d], actual vectors num is [%d]", len(objectTypes), len(vectors))
		}

		for i, objectType := range objectTypes {
			objectType.Vector = vectors[i].Vector
		}
	}

	documents := []map[string]any{}
	for _, objectType := range objectTypes {
		docid := interfaces.GenerateConceptDocuemtnID(objectType.KNID, interfaces.MODULE_TYPE_OBJECT_TYPE,
			objectType.OTID, objectType.Branch)
		objectType.ModuleType = interfaces.MODULE_TYPE_OBJECT_TYPE

		// Convert to map for dataset
		docBytes, err := sonic.Marshal(objectType)
		if err != nil {
			logger.Errorf("Failed to marshal ObjectType: %s", err.Error())
			span.SetStatus(codes.Error, "序列化对象类失败")
			return err
		}

		var doc map[string]any
		if err := sonic.Unmarshal(docBytes, &doc); err != nil {
			logger.Errorf("Failed to unmarshal ObjectType: %s", err.Error())
			span.SetStatus(codes.Error, "反序列化对象类失败")
			return err
		}

		// Serialize logic_properties[].parameters to JSON string
		if logicProps, ok := doc["logic_properties"].([]any); ok {
			for _, lp := range logicProps {
				if lpMap, ok := lp.(map[string]any); ok {
					if params, exists := lpMap["parameters"]; exists {
						paramsBytes, err := sonic.Marshal(params)
						if err != nil {
							logger.Errorf("Failed to marshal logic_properties parameters: %s", err.Error())
							span.SetStatus(codes.Error, "序列化逻辑属性参数失败")
							return err
						}
						lpMap["parameters"] = string(paramsBytes)
					}
				}
			}
		}

		// Set document ID
		doc["_id"] = docid
		documents = append(documents, doc)
	}

	err := ots.vba.WriteDatasetDocuments(ctx, interfaces.BKN_DATASET_ID, documents)
	if err != nil {
		logger.Errorf("WriteDatasetDocuments error: %s", err.Error())
		span.SetStatus(codes.Error, "对象类概念索引写入失败")
		return err
	}

	return nil
}

// type vectorFunc func(ctx context.Context, words []string) ([]cond.VectorResp, error)

func (ots *objectTypeService) SearchObjectTypes(ctx context.Context,
	query *interfaces.ConceptsQuery) (interfaces.ObjectTypes, error) {

	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "业务知识网络对象类检索")
	defer span.End()

	response := interfaces.ObjectTypes{}
	var err error

	// 判断userid是否有查看业务知识网络的权限
	err = ots.ps.CheckPermission(ctx, interfaces.PermissionResource{
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
				if !ots.appSetting.ServerSetting.DefaultSmallModelEnabled {
					err = errors.New(cond.DEFAULT_SMALL_MODEL_ENABLED_FALSE_ERROR)
					span.SetStatus(codes.Error, err.Error())
					return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
						berrors.BknBackend_ObjectType_InternalError).
						WithErrorDetails(err.Error())
				}
				dftModel, err := ots.mfa.GetDefaultModel(ctx)
				if err != nil {
					logger.Errorf("GetDefaultModel error: %s", err.Error())
					span.SetStatus(codes.Error, "获取默认模型失败")
					return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
						berrors.BknBackend_ObjectType_InternalError).
						WithErrorDetails(err.Error())
				}
				result, err := ots.mfa.GetVector(ctx, dftModel, []string{word})
				if err != nil {
					logger.Errorf("GetVector error: %s", err.Error())
					span.SetStatus(codes.Error, "获取业务知识网络向量失败")
					return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
						berrors.BknBackend_ObjectType_InternalError).
						WithErrorDetails(err.Error())
				}
				return result, nil
			})
		if err != nil {
			return response, rest.NewHTTPError(ctx, http.StatusBadRequest,
				berrors.BknBackend_ObjectType_InvalidParameter_ConceptCondition).
				WithErrorDetails(fmt.Sprintf("failed to convert condition to filter condition, %s", err.Error()))
		}
	}

	// 1. 获取组下的对象类
	otIDMap := map[string]bool{} // 分组下的对象类id
	otIDs := []string{}          // 不同组下的对象类可以重叠，所以需要对对象类id的数组去重
	if len(query.ConceptGroups) > 0 {

		// 校验分组是否都存在，按分组id获取分组
		cgCnt, err := ots.cga.GetConceptGroupsTotal(ctx, interfaces.ConceptGroupsQueryParams{
			KNID:   query.KNID,
			Branch: query.Branch,
			CGIDs:  query.ConceptGroups,
		})
		if err != nil {
			logger.Errorf("GetConceptGroupsTotal in knowledge network[%s] error: %s", query.KNID, err.Error())
			span.SetStatus(codes.Error, fmt.Sprintf("GetConceptGroupsTotal in knowledge network[%s], error: %v", query.KNID, err))

			return response, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				berrors.BknBackend_ObjectType_InternalError).WithErrorDetails(err.Error())
		}
		if cgCnt == 0 {
			errStr := fmt.Sprintf("all concept group not found, expect concept group nums is [%d], actual concept group num is [%d]",
				cgCnt, len(query.ConceptGroups))
			logger.Errorf(errStr)

			return response, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				berrors.BknBackend_ObjectType_InternalError).
				WithErrorDetails(errStr)
		}

		// 在当前业务知识网络下查找属于请求的分组范围内的对象类ID
		otIDArr, err := ots.cga.GetConceptIDsByConceptGroupIDs(ctx, query.KNID,
			query.Branch, query.ConceptGroups, interfaces.MODULE_TYPE_OBJECT_TYPE)
		if err != nil {
			errStr := fmt.Sprintf("GetConceptIDsByConceptGroupIDs failed, kn_id:[%s],branch:[%s],cg_ids:[%v], error: %v",
				query.KNID, query.Branch, query.ConceptGroups, err)
			logger.Errorf(errStr)
			span.SetStatus(codes.Error, errStr)
			span.End()

			return response, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				berrors.BknBackend_ObjectType_InternalError).WithErrorDetails(err.Error())
		}

		// 概念分组下没有对象类,返回空
		if len(otIDArr) == 0 {
			return response, nil
		}

		for _, otID := range otIDArr {
			if !otIDMap[otID] {
				otIDMap[otID] = true
				otIDs = append(otIDs, otID)
			}
		}
	}

	// 根据NeedTotal参数决定是否查询total
	if query.NeedTotal {
		if len(otIDMap) == 0 {
			// 查询总数
			params := &interfaces.ResourceDataQueryParams{
				FilterCondition: filterCondition,
				Offset:          0,
				Limit:           1, // 查询1条数据，获取total
				NeedTotal:       true,
			}
			datasetResp, err := ots.vba.QueryResourceData(ctx, interfaces.BKN_DATASET_ID, params)
			if err != nil {
				logger.Errorf("QueryDatasetData error: %s", err.Error())
				span.SetStatus(codes.Error, "业务知识网络对象类检索查询总数失败")
				return response, rest.NewHTTPError(ctx, http.StatusInternalServerError,
					berrors.BknBackend_ObjectType_InternalError).
					WithErrorDetails(err.Error())
			}
			response.TotalCount = datasetResp.TotalCount
		} else {
			// 指定了分组，需要查询分组内且符合条件的总数
			total, err := ots.GetTotalWithLargeOTIDs(ctx, filterCondition, otIDs)
			if err != nil {
				return response, err
			}
			response.TotalCount = total
		}
	}

	// 4. 迭代查询直到获取足够数量或没有更多数据
	// 保留 search_after 的循环查询结构，当前使用 offset 分页作为临时方案
	objectTypes := []*interfaces.ObjectType{}
	var totalFilteredCount int64 = 0
	limit := query.Limit
	if limit == 0 {
		limit = interfaces.SearchAfter_Limit
	}

	for {
		// 调用 dataset 查询
		params := &interfaces.ResourceDataQueryParams{
			FilterCondition: filterCondition,
			Limit:           limit,
			NeedTotal:       false,
			SearchAfter:     query.SearchAfter,
			Sort:            query.Sort,
		}
		datasetResp, err := ots.vba.QueryResourceData(ctx, interfaces.BKN_DATASET_ID, params)
		if err != nil {
			logger.Errorf("QueryResourceData error: %s", err.Error())
			span.SetStatus(codes.Error, "业务知识网络对象类检索查询失败")
			return response, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				berrors.BknBackend_ObjectType_InternalError).
				WithErrorDetails(err.Error())
		}

		// 如果没有数据了，跳出循环
		if len(datasetResp.Entries) == 0 {
			break
		}

		// 5. 处理查询结果
		for _, entry := range datasetResp.Entries {
			// Deserialize logic_properties[].parameters from JSON string
			if logicProps, exists := entry["logic_properties"]; exists {
				if logicPropsArr, ok := logicProps.([]any); ok {
					for _, lp := range logicPropsArr {
						if lpMap, ok := lp.(map[string]any); ok {
							if paramsStr, exists := lpMap["parameters"]; exists {
								if paramsStrStr, ok := paramsStr.(string); ok && paramsStrStr != "" {
									var params []interfaces.Parameter
									if err := sonic.Unmarshal([]byte(paramsStrStr), &params); err != nil {
										logger.Errorf("Failed to unmarshal object_type logic_properties parameters: %s", err.Error())
										return response, rest.NewHTTPError(ctx, http.StatusBadRequest,
											berrors.BknBackend_InternalError_UnMarshalDataFailed).
											WithErrorDetails(fmt.Sprintf("failed to Unmarshal logic_properties parameters, %s", err.Error()))
									}
									lpMap["parameters"] = params
								}
							}
						}
					}
				}
			}

			// 转成 object type 的 struct
			jsonByte, err := json.Marshal(entry)
			if err != nil {
				return response, rest.NewHTTPError(ctx, http.StatusBadRequest,
					berrors.BknBackend_InternalError_MarshalDataFailed).
					WithErrorDetails(fmt.Sprintf("failed to Marshal dataset entry, %s", err.Error()))
			}
			var objectType interfaces.ObjectType
			err = json.Unmarshal(jsonByte, &objectType)
			if err != nil {
				return response, rest.NewHTTPError(ctx, http.StatusBadRequest,
					berrors.BknBackend_InternalError_UnMarshalDataFailed).
					WithErrorDetails(fmt.Sprintf("failed to Unmarshal dataset entry to Object Type, %s", err.Error()))
			}

			// 如果没有指定分组，或者对象类属于分组，则添加
			if len(otIDMap) == 0 || otIDMap[objectType.OTID] {
				// 处理数据源和操作符
				err = ots.processObjectTypeDetails(ctx, &objectType)
				if err != nil {
					return response, err
				}
				// 提取 _score（如果有）
				if scoreVal, ok := entry["_score"]; ok {
					if scoreFloat, ok := scoreVal.(float64); ok {
						score := float64(scoreFloat)
						objectType.Score = &score
					}
				}
				objectType.Vector = nil

				objectTypes = append(objectTypes, &objectType)
				totalFilteredCount++

				// 如果已经收集到足够的数量，跳出循环
				if len(objectTypes) >= query.Limit && query.Limit > 0 {
					break
				}
			}
		}

		// 用结果的search after 继续查询
		query.SearchAfter = datasetResp.SearchAfter

		// 如果已经收集到足够的数量或者没有更多数据了，跳出循环
		if (query.Limit > 0 && len(objectTypes) >= query.Limit) || len(datasetResp.Entries) < limit {
			break
		}
	}

	response.Entries = objectTypes
	response.SearchAfter = query.SearchAfter
	return response, nil
}

// 提取出来的处理对象类型详情的函数
func (ots *objectTypeService) processObjectTypeDetails(ctx context.Context, objectType *interfaces.ObjectType) error {

	// 查视图或 vega Resource 组装 ops. 不需要组装,因为保存的时候会保存进去
	if objectType.DataSource != nil && objectType.DataSource.ID != "" {
		dsType := objectType.DataSource.Type
		if dsType == "" {
			dsType = interfaces.DATA_SOURCE_TYPE_DATA_VIEW
		}
		switch dsType {
		case interfaces.DATA_SOURCE_TYPE_RESOURCE:
			res, err := ots.vba.GetResourceByID(ctx, objectType.DataSource.ID)
			if err != nil || res == nil {
				otellog.LogWarn(ctx, fmt.Sprintf("Object type [%s]'s vega Resource %s not found, error: %v",
					objectType.OTID, objectType.DataSource.ID, err))
			} else {
				objectType.DataSource.Name = res.Name
				fieldsMap := logics.VegaResourceSchemaToFieldsMap(res)
				dslView := &interfaces.DataView{QueryType: interfaces.VIEW_QueryType_DSL}
				for j, prop := range objectType.DataProperties {
					if prop.MappedField != nil {
						if field, exists := fieldsMap[prop.MappedField.Name]; exists {
							objectType.DataProperties[j].MappedField.DisplayName = field.DisplayName
							objectType.DataProperties[j].MappedField.Type = field.Type
						}
					}
					objectType.DataProperties[j].ConditionOperations = ots.processConditionOperations(objectType, prop, dslView)
				}
			}
		default:
			dataView, err := ots.dva.GetDataViewByID(ctx, objectType.DataSource.ID)
			if err != nil || dataView == nil {
				otellog.LogWarn(ctx, fmt.Sprintf("Object type [%s]'s Data view %s not found, error: %v",
					objectType.OTID, objectType.DataSource.ID, err))
			} else {
				objectType.DataSource.Name = dataView.ViewName
				// 视图不为空，则把支持的操作符返回
				for j, prop := range objectType.DataProperties {
					// 不为空时，才翻译字段显示名。为空则不翻译
					if prop.MappedField != nil {
						if field, exists := dataView.FieldsMap[prop.MappedField.Name]; exists {
							objectType.DataProperties[j].MappedField.DisplayName = field.DisplayName
							objectType.DataProperties[j].MappedField.Type = field.Type
						}
					}
					// 字符串类型的属性支持的操作符返回
					objectType.DataProperties[j].ConditionOperations = ots.processConditionOperations(objectType, prop, dataView)
				}
			}
		}

		// 逻辑属性，资源id转名称
		for j, logicProp := range objectType.LogicProperties {
			if logicProp.DataSource != nil {
				switch logicProp.DataSource.Type {
				case interfaces.LOGIC_PROPERTY_TYPE_METRIC:
					if logicProp.DataSource.ID != "" {
						// 获取指标模型名称
						model, err := ots.dda.GetMetricModelByID(ctx, logicProp.DataSource.ID)
						if err != nil || model == nil {
							// 依赖不存在或者请求报错，不报错，跳过
							otellog.LogWarn(ctx, fmt.Sprintf("Object type [%s]'s logic property [%s] metric model [%s] not found, error: %v",
								objectType.OTID, logicProp.Name, objectType.DataSource.ID, err))
						} else {
							// 依赖存在时才做相关操作
							objectType.LogicProperties[j].DataSource.Name = model.ModelName

							// 逻辑属性-指标，返回指标模型的分析维度
							objectType.LogicProperties[j].AnalysisDims = model.AnalysisDims

							// 对参数填充comment
							processMetricPropertyParamComment(ctx, logicProp, model, objectType, j)
						}
					}
				case interfaces.LOGIC_PROPERTY_TYPE_OPERATOR:
					//todo: 算子的名称,前端翻译
				}
				// todo: 处理动态参数,动态参数统一放在一个新字段上,供统一召回的大模型使用(检索那边也需要处理一下)
			}
		}
	}
	return nil
}

// 处理指标属性的参数的comment
func processMetricPropertyParamComment(ctx context.Context, logicProp *interfaces.LogicProperty, model *interfaces.MetricModel,
	objectType *interfaces.ObjectType, j int) {

	// 对参数填充comment
	for k, param := range logicProp.Parameters {
		// 存在则给，否则不给，不报错，记录warn日志
		if model != nil && model.FieldsMap != nil {
			if field, exist := model.FieldsMap[param.Name]; exist {
				objectType.LogicProperties[j].Parameters[k].Comment = field.Comment
				continue
			} else {
				// 字段不存在，记录warn日志
				otellog.LogWarn(ctx, fmt.Sprintf("Object type [%s]'s logic property [%s]'s parameter[%s] not found in metric model[%s]",
					objectType.OTID, logicProp.Name, param.Name, objectType.DataSource.ID))
			}
		}

		// 处理特殊参数或记录warn日志
		switch param.Name {
		case "instant":
			comment := "是否是即时查询。可选，默认为 false。当 instant = true 时，表示即时查询；当 instant = false 时，表示范围查询。"
			objectType.LogicProperties[j].Parameters[k].Comment = &comment
		case "start":
			comment := "指标查询的开始时间。 start=<unix_timestamp>，单位到毫秒。 例如: 1646360670123"
			objectType.LogicProperties[j].Parameters[k].Comment = &comment
		case "end":
			comment := "指标查询的结束时间。end=<unix_timestamp>，单位到毫秒。例如: 1646471470123"
			objectType.LogicProperties[j].Parameters[k].Comment = &comment
		case "step":
			comment := "范围查询的步长。当 instant 为 false 时, 必须。step=<time_durations>，用一个数字，后面跟时间单位来定义。"
			objectType.LogicProperties[j].Parameters[k].Comment = &comment
		}
	}
}

func (ots *objectTypeService) GetTotal(ctx context.Context, filterCondition map[string]any) (total int64, err error) {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "logic layer: search object type total ")
	defer span.End()

	params := &interfaces.ResourceDataQueryParams{
		FilterCondition: filterCondition,
		Offset:          0,
		Limit:           1, // 查询1条数据，获取total
		NeedTotal:       true,
	}
	datasetResp, err := ots.vba.QueryResourceData(ctx, interfaces.BKN_DATASET_ID, params)
	if err != nil {
		span.SetStatus(codes.Error, "Search total documents count failed")
		return total, rest.NewHTTPError(ctx, http.StatusInternalServerError, berrors.BknBackend_ObjectType_InternalError).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	if datasetResp == nil {
		return 0, nil
	}
	return datasetResp.TotalCount, nil
}

// 内部调用，不加权限校验
func (ots *objectTypeService) GetObjectTypeIDsByKnID(ctx context.Context,
	knID string, branch string) ([]string, error) {
	// 获取对象类
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, fmt.Sprintf("按kn_id[%s]获取对象类IDs", knID))
	defer span.End()

	// 获取对象类基本信息
	otIDs, err := ots.ota.GetObjectTypeIDsByKnID(ctx, knID, branch)
	if err != nil {
		logger.Errorf("GetObjectTypeIDsByKnID error: %s", err.Error())
		span.SetStatus(codes.Error, fmt.Sprintf("Get object type ids by kn_id[%s] error: %v", knID, err))

		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_ObjectType_InternalError_GetObjectTypesByIDsFailed).WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return otIDs, nil
}

func (ots *objectTypeService) GetAllObjectTypesByKnID(ctx context.Context,
	knID string, branch string) (map[string]*interfaces.ObjectType, error) {
	// 获取对象类
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, fmt.Sprintf("按kn_id[%s]获取对象类基本信息", knID))
	defer span.End()

	// 获取对象类基本信息
	objectTypes, err := ots.ota.GetAllObjectTypesByKnID(ctx, knID, branch)
	if err != nil {
		logger.Errorf("GetAllObjectTypesByKnID error: %s", err.Error())
		span.SetStatus(codes.Error, fmt.Sprintf("Get all object type by kn_id[%s] error: %v", knID, err))

		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_ObjectType_InternalError_GetObjectTypesByIDsFailed).WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return objectTypes, nil
}

// 内部接口，不检查权限
func (ots *objectTypeService) GetObjectTypeByID(ctx context.Context, tx *sql.Tx,
	knID string, branch string, otID string) (*interfaces.ObjectType, error) {
	// 获取对象类
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, fmt.Sprintf("查询对象类[%s]信息", otID))
	defer span.End()

	var err error
	// 0. 开始事务
	if tx == nil {
		tx, err = ots.db.Begin()
		if err != nil {
			logger.Errorf("Begin transaction error: %s", err.Error())
			span.SetStatus(codes.Error, "事务开启失败")
			otellog.LogError(ctx, "Begin transaction error", err)
			return &interfaces.ObjectType{}, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				berrors.BknBackend_ObjectType_InternalError_BeginTransactionFailed).
				WithErrorDetails(err.Error())
		}
		// 0.1 异常时
		defer func() {
			switch err {
			case nil:
				// 提交事务
				err = tx.Commit()
				if err != nil {
					logger.Errorf("GetObjectTypeByID Transaction Commit Failed:%v", err)
					span.SetStatus(codes.Error, "提交事务失败")
					otellog.LogError(ctx, "GetObjectTypeByID Transaction Commit Failed", err)
					return
				}
				logger.Infof("GetObjectTypeByID Transaction Commit Success")
				otellog.LogDebug(ctx, "GetObjectTypeByID Transaction Commit Success")
			default:
				rollbackErr := tx.Rollback()
				if rollbackErr != nil {
					logger.Errorf("GetObjectTypeByID Transaction Rollback Error:%v", rollbackErr)
					span.SetStatus(codes.Error, "事务回滚失败")
					otellog.LogError(ctx, "GetObjectTypeByID Transaction Rollback Error", err)
				}
			}
		}()
	}

	// 获取对象类基本信息
	objectType, err := ots.ota.GetObjectTypeByID(ctx, tx, knID, branch, otID)
	if err != nil {
		logger.Errorf("GetObjectTypeByID error: %s", err.Error())
		span.SetStatus(codes.Error, fmt.Sprintf("Get object type by id[%s] error: %v", otID, err))

		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_ObjectType_InternalError_GetObjectTypeByIDFailed).WithErrorDetails(err.Error())
	}
	if objectType == nil {
		return nil, rest.NewHTTPError(ctx, http.StatusNotFound, berrors.BknBackend_ObjectType_ObjectTypeNotFound).
			WithErrorDetails(fmt.Sprintf("对象类[id:%s]不存在: %v", otID, err))
	}

	span.SetStatus(codes.Ok, "")
	return objectType, nil
}

// 处理字符串类型的操作符
func (ots *objectTypeService) processConditionOperations(objectType *interfaces.ObjectType, prop *interfaces.DataProperty,
	dataView *interfaces.DataView) []string {

	ops := []string{}
	if objectType.Status != nil && !objectType.Status.IndexAvailable {
		// 索引不可用时,按视图的字段来做,varchar是opensearch没有的,是数据库字段.keyword和text是opensearch独有的,所以按字段类型来分
		switch prop.Type {
		case "keyword":
			ops = interfaces.DSL_KEYWORD_OPS
		case "varchar", "string":
			// string的原始类型可以是keyword或者varchar,所以按视图类型来区别一下
			if dataView.QueryType == interfaces.VIEW_QueryType_DSL {
				ops = interfaces.DSL_KEYWORD_OPS
			} else {
				ops = interfaces.SQL_STRING_OPS
			}
		case "text":
			if dataView.QueryType == interfaces.VIEW_QueryType_DSL {
				ops = interfaces.DSL_TEXT_OPS // dsl的text有match
				ops = append(ops, interfaces.DSL_KEYWORD_OPS...)
			} else {
				ops = interfaces.SQL_STRING_OPS
			}
		case "vector":
			// 小模型打开了才能支持knn操作
			if ots.appSetting.ServerSetting.DefaultSmallModelEnabled {
				ops = append(ops, cond.OperationKNN)
			}
		}
	} else {
		opMap := make(map[string]string)
		// 先看本类型，text 类型支持 match,其余的字符串类型可支持 == != in not_in
		switch prop.Type {
		case "keyword", "varchar", "string":
			// Copy map content instead of assigning reference to avoid concurrent map access
			for k, v := range interfaces.DSL_KEYWORD_OPS_MAP {
				opMap[k] = v
			}
		case "text":
			// Copy map content instead of assigning reference to avoid concurrent map access
			for k, v := range interfaces.DSL_KEYWORD_OPS_MAP {
				opMap[k] = v
			}
			for k, v := range interfaces.DSL_TEXT_OPS_MAP {
				opMap[k] = v
			}
		case "vector":
			opMap[cond.OperationKNN] = cond.OperationKNN
		}

		// 配置了keyword索引
		if prop.IndexConfig != nil && prop.IndexConfig.KeywordConfig.Enabled {
			// 把 keyword 支持的操作符添加
			for k, v := range interfaces.DSL_KEYWORD_OPS_MAP {
				opMap[k] = v
			}
		}
		// 配置了full text索引,则可以做  match 的操作
		if prop.IndexConfig != nil && prop.IndexConfig.FulltextConfig.Enabled {
			opMap[cond.OperationMatch] = cond.OperationMatch
			opMap[cond.OperationMultiMatch] = cond.OperationMultiMatch
		}
		// 配置了 vector 索引, 且向量化小模型是打开的,则可以做 knn 的操作
		if prop.IndexConfig != nil && prop.IndexConfig.VectorConfig.Enabled &&
			ots.appSetting.ServerSetting.DefaultSmallModelEnabled {

			opMap[cond.OperationKNN] = cond.OperationKNN
		}

		for k := range opMap {
			ops = append(ops, k)
		}
	}
	return ops
}

// 处理对象类与组的关系，并保存
func (ots *objectTypeService) handleGroupRelations(ctx context.Context, tx *sql.Tx,
	objectType *interfaces.ObjectType, currentTime int64, strictMode bool) error {

	var err error
	cgIDs := []string{}
	for _, cg := range objectType.ConceptGroups {
		cgIDs = append(cgIDs, cg.CGID)
	}
	// id去重后再查
	cgIDs = common.DuplicateSlice(cgIDs)

	// When strictMode is true, validate all concept groups exist
	if strictMode {
		conceptGroups, err := ots.cga.GetConceptGroupsByIDs(ctx, tx, objectType.KNID, objectType.Branch, cgIDs)
		if err != nil {
			errStr := fmt.Sprintf("GetConceptGroupsByIDs failed, the kn_id: [%s], branch: [%s], cg_ids: [%v], error: %s",
				objectType.KNID, objectType.Branch, cgIDs, err.Error())
			logger.Errorf(errStr)

			return rest.NewHTTPError(ctx, http.StatusInternalServerError,
				berrors.BknBackend_ObjectType_InternalError).
				WithErrorDetails(errStr)
		}
		if len(conceptGroups) != len(cgIDs) {
			errStr := fmt.Sprintf("Exists any concept group not found, expect concept group nums is [%d], actual concept group num is [%d]",
				len(cgIDs), len(conceptGroups))
			logger.Errorf(errStr)

			return rest.NewHTTPError(ctx, http.StatusInternalServerError,
				berrors.BknBackend_ObjectType_InternalError).
				WithErrorDetails(errStr)
		}
	}

	// 创建
	for _, cg := range objectType.ConceptGroups {
		cgRelationID := xid.New().String()
		err = ots.cga.CreateConceptGroupRelation(ctx, tx, &interfaces.ConceptGroupRelation{
			ID:          cgRelationID,
			KNID:        objectType.KNID,
			Branch:      objectType.Branch,
			CGID:        cg.CGID,
			ConceptType: interfaces.MODULE_TYPE_OBJECT_TYPE,
			ConceptID:   objectType.OTID,
			CreateTime:  currentTime,
		})
		if err != nil {
			errStr := fmt.Sprintf("CreateConceptGroupRelation failed, the concept group is [%s], knowledge network is [%s], branch is [%s], object type is [%s]",
				cg.CGID, objectType.KNID, objectType.Branch, objectType.OTID)
			logger.Errorf(errStr)

			return rest.NewHTTPError(ctx, http.StatusInternalServerError,
				berrors.BknBackend_ObjectType_InternalError_CreateConceptGroupRelationFailed).
				WithErrorDetails(err.Error())
		}
	}
	return nil
}

// syncObjectGroups 同步分组关系（更新时使用，全量替换）
func (ots *objectTypeService) syncObjectGroups(ctx context.Context, tx *sql.Tx,
	objectType interfaces.ObjectType, currentTime int64, strictMode bool) error {

	cgIDs := []string{}
	for _, cg := range objectType.ConceptGroups {
		cgIDs = append(cgIDs, cg.CGID)
	}
	// id去重后再查
	cgIDs = common.DuplicateSlice(cgIDs)

	// When strictMode is true and cgIDs not empty, validate all concept groups exist
	if strictMode && len(cgIDs) > 0 {
		conceptGroups, err := ots.cga.GetConceptGroupsByIDs(ctx, tx, objectType.KNID, objectType.Branch, cgIDs)
		if err != nil {
			errStr := fmt.Sprintf("GetConceptGroupsByIDs failed, the kn_id: [%s], branch: [%s], cg_ids: [%v], error: %s",
				objectType.KNID, objectType.Branch, cgIDs, err.Error())
			logger.Errorf(errStr)

			return rest.NewHTTPError(ctx, http.StatusInternalServerError,
				berrors.BknBackend_ObjectType_InternalError).
				WithErrorDetails(errStr)
		}
		if len(conceptGroups) != len(cgIDs) {
			errStr := fmt.Sprintf("Exists any concept group not found, expect concept group nums is [%d], actual concept group num is [%d]",
				len(cgIDs), len(conceptGroups))
			logger.Errorf(errStr)

			return rest.NewHTTPError(ctx, http.StatusBadRequest,
				berrors.BknBackend_ObjectType_InvalidParameter).
				WithErrorDetails(errStr)
		}
	}

	// 1. 获取对象类现有的分组关系
	existingRelation, err := ots.cga.GetConceptGroupsByOTIDs(ctx, tx, interfaces.ConceptGroupRelationsQueryParams{
		KNID:   objectType.KNID,
		Branch: objectType.Branch,
		OTIDs:  []string{objectType.OTID},
	})
	if err != nil {
		logger.Errorf(err.Error())
		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_ObjectType_InternalError).WithErrorDetails(err.Error())
	}

	// 2. 计算需要添加和删除的分组
	existingGroupIDs := make(map[string]bool)
	if len(existingRelation) == 1 {
		// 对象类已建立的关系
		for _, rel := range existingRelation[objectType.OTID] {
			existingGroupIDs[rel.CGID] = true
		}
	}

	newGroupIDs := make(map[string]bool)
	for _, ref := range objectType.ConceptGroups {
		newGroupIDs[ref.CGID] = true
	}

	// 计算差异
	groupsToAdd := make([]string, 0)
	groupsToRemove := make([]string, 0)

	for groupID := range newGroupIDs {
		if !existingGroupIDs[groupID] {
			groupsToAdd = append(groupsToAdd, groupID)
		}
	}

	for groupID := range existingGroupIDs {
		if !newGroupIDs[groupID] {
			groupsToRemove = append(groupsToRemove, groupID)
		}
	}

	// 3. 执行添加操作
	if len(groupsToAdd) > 0 {
		// 构建新增关系记录
		for _, cgID := range groupsToAdd {
			cgRelationID := xid.New().String()
			err = ots.cga.CreateConceptGroupRelation(ctx, tx, &interfaces.ConceptGroupRelation{
				ID:          cgRelationID,
				KNID:        objectType.KNID,
				Branch:      objectType.Branch,
				CGID:        cgID,
				ConceptType: interfaces.MODULE_TYPE_OBJECT_TYPE,
				ConceptID:   objectType.OTID,
				CreateTime:  currentTime,
			})
			if err != nil {
				errStr := fmt.Sprintf("CreateConceptGroupRelation failed, the concept group is [%s], knowledge network is [%s], branch is [%s], object type is [%s], error is [%s]",
					cgID, objectType.KNID, objectType.Branch, objectType.OTID, err.Error())
				logger.Errorf(errStr)

				return rest.NewHTTPError(ctx, http.StatusInternalServerError,
					berrors.BknBackend_ObjectType_InternalError_CreateConceptGroupRelationFailed).
					WithErrorDetails(errStr)
			}
		}
	}

	// 4. 执行删除操作
	if len(groupsToRemove) > 0 {
		// 删除对象类与分组的绑定关系
		rowsAffect, err := ots.cga.DeleteObjectTypesFromGroup(ctx, tx, interfaces.ConceptGroupRelationsQueryParams{
			KNID:        objectType.KNID,
			Branch:      objectType.Branch,
			CGIDs:       groupsToRemove,
			ConceptType: interfaces.MODULE_TYPE_OBJECT_TYPE,
			OTIDs:       []string{objectType.OTID},
		})
		if err != nil {
			errStr := fmt.Sprintf("DeleteObjectTypesFromGroup failed, the concept group is [%v], kn_id is [%s], branch is [%s], object type is [%s], error is [%s]",
				groupsToRemove, objectType.KNID, objectType.Branch, objectType.OTID, err.Error())
			logger.Errorf(errStr)

			return rest.NewHTTPError(ctx, http.StatusInternalServerError,
				berrors.BknBackend_ObjectType_InternalError).
				WithErrorDetails(errStr)
		}
		// 记录ingo日志，删除的条数
		logger.Infof("DeleteObjectTypesFromGroup success, the concept group is [%v], kn_id is [%s], branch is [%s], object type is [%s], rowsAffect is [%d]",
			groupsToRemove, objectType.KNID, objectType.Branch, objectType.OTID, rowsAffect)
	}

	return nil
}

// 分批查询
func (ots *objectTypeService) GetTotalWithLargeOTIDs(ctx context.Context,
	filterCondition map[string]any,
	otIDs []string) (int64, error) {

	total := int64(0)
	for i := 0; i < len(otIDs); i += interfaces.GET_TOTAL_CONCEPTID_BATCH_SIZE {
		end := i + interfaces.GET_TOTAL_CONCEPTID_BATCH_SIZE
		if end > len(otIDs) {
			end = len(otIDs)
		}

		batchIDs := otIDs[i:end]
		batchTotal, err := ots.GetTotalWithOTIDs(ctx, filterCondition, batchIDs)
		if err != nil {
			return 0, err
		}

		total += batchTotal
	}

	return total, nil
}

// 查询指定对象类ID列表的对象类总数
func (ots *objectTypeService) GetTotalWithOTIDs(ctx context.Context,
	filterCondition map[string]any,
	otIDs []string) (int64, error) {

	// 构建包含 OTID 过滤的 filter condition
	otIDCondition := map[string]any{
		"field":      "id",
		"operation":  "in",
		"value":      otIDs,
		"value_from": "const",
	}

	var combinedCondition map[string]any
	if filterCondition == nil {
		combinedCondition = otIDCondition
	} else {
		combinedCondition = map[string]any{
			"operation": "and",
			"sub_conditions": []map[string]any{
				filterCondition,
				otIDCondition,
			},
		}
	}

	// 执行计数查询
	total, err := ots.GetTotal(ctx, combinedCondition)
	if err != nil {
		return total, err
	}

	return total, nil
}

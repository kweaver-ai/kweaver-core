// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package relation_type

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
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	bknsdk "github.com/kweaver-ai/bkn-specification/sdk/golang/bkn"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/rs/xid"
	"go.opentelemetry.io/otel/codes"

	"bkn-backend/common"
	cond "bkn-backend/common/condition"
	berrors "bkn-backend/errors"
	"bkn-backend/interfaces"
	"bkn-backend/logics"
	"bkn-backend/logics/object_type"
	"bkn-backend/logics/permission"
	"bkn-backend/logics/user_mgmt"
)

var (
	rtServiceOnce sync.Once
	rtService     interfaces.RelationTypeService
)

type relationTypeService struct {
	appSetting *common.AppSetting
	db         *sql.DB
	cga        interfaces.ConceptGroupAccess
	dva        interfaces.DataViewAccess
	mfa        interfaces.ModelFactoryAccess
	ots        interfaces.ObjectTypeService
	ps         interfaces.PermissionService
	rta        interfaces.RelationTypeAccess
	ums        interfaces.UserMgmtService
	vba        interfaces.VegaBackendAccess
}

func NewRelationTypeService(appSetting *common.AppSetting) interfaces.RelationTypeService {
	rtServiceOnce.Do(func() {
		rtService = &relationTypeService{
			appSetting: appSetting,
			db:         logics.DB,
			cga:        logics.CGA,
			dva:        logics.DVA,
			mfa:        logics.MFA,
			ots:        object_type.NewObjectTypeService(appSetting),
			ps:         permission.NewPermissionService(appSetting),
			rta:        logics.RTA,
			ums:        user_mgmt.NewUserMgmtService(appSetting),
			vba:        logics.VBA,
		}
	})
	return rtService
}

func (rts *relationTypeService) CheckRelationTypeExistByID(ctx context.Context, knID string, branch string, rtID string) (string, bool, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("校验关系类[%s]的存在性", rtID))
	defer span.End()

	rtName, exist, err := rts.rta.CheckRelationTypeExistByID(ctx, knID, branch, rtID)
	if err != nil {
		logger.Errorf("CheckRelationTypeExistByID error: %s", err.Error())
		// 记录处理的 sql 字符串
		o11y.Error(ctx, fmt.Sprintf("按ID[%s]获取关系类失败: %v", rtID, err))
		span.SetStatus(codes.Error, fmt.Sprintf("按ID[%s]获取关系类失败", rtID))
		return "", exist, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_RelationType_InternalError_CheckRelationTypeIfExistFailed).WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return rtName, exist, nil
}

func (rts *relationTypeService) CreateRelationTypes(ctx context.Context, tx *sql.Tx,
	relationTypes []*interfaces.RelationType, mode string, validateDependency bool) ([]string, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "Create relation type")
	defer span.End()

	// 判断userid是否有修改业务知识网络的权限
	err := rts.ps.CheckPermission(ctx, interfaces.Resource{
		Type: interfaces.RESOURCE_TYPE_KN,
		ID:   relationTypes[0].KNID,
	}, []string{interfaces.OPERATION_TYPE_MODIFY})
	if err != nil {
		return []string{}, err
	}

	// 0. 开始事务
	if tx == nil {
		tx, err = rts.db.Begin()
		if err != nil {
			logger.Errorf("Begin transaction error: %s", err.Error())
			span.SetStatus(codes.Error, "事务开启失败")
			o11y.Error(ctx, fmt.Sprintf("Begin transaction error: %s", err.Error()))
			return []string{}, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				berrors.BknBackend_RelationType_InternalError_BeginTransactionFailed).
				WithErrorDetails(err.Error())
		}
		// 0.1 异常时
		defer func() {
			switch err {
			case nil:
				// 提交事务
				err = tx.Commit()
				if err != nil {
					logger.Errorf("CreateRelationType Transaction Commit Failed:%v", err)
					span.SetStatus(codes.Error, "提交事务失败")
					o11y.Error(ctx, fmt.Sprintf("CreateRelationType Transaction Commit Failed: %s", err.Error()))
					return
				}
				logger.Infof("CreateRelationType Transaction Commit Success")
				o11y.Debug(ctx, "CreateRelationType Transaction Commit Success")
			default:
				rollbackErr := tx.Rollback()
				if rollbackErr != nil {
					logger.Errorf("CreateRelationType Transaction Rollback Error:%v", rollbackErr)
					span.SetStatus(codes.Error, "事务回滚失败")
					o11y.Error(ctx, fmt.Sprintf("CreateRelationType Transaction Rollback Error: %s", err.Error()))
				}
			}
		}()
	}

	currentTime := time.Now().UnixMilli()
	for _, relationType := range relationTypes {
		// 若提交的模型id为空，生成分布式ID
		if relationType.RTID == "" {
			relationType.RTID = xid.New().String()
		}

		accountInfo := interfaces.AccountInfo{}
		if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
			accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
		}
		relationType.Creator = accountInfo
		relationType.Updater = accountInfo

		relationType.CreateTime = currentTime
		relationType.UpdateTime = currentTime

		// 校验起点对象类、终点对象类非空时，需校验存在性
		err = rts.validateDependency(ctx, tx, relationType, validateDependency)
		if err != nil {
			return []string{}, err
		}

		bknRel := logics.ToBKNRelationType(relationType)
		relationType.BKNRawContent = bknsdk.SerializeRelationType(bknRel)
	}

	createRelationTypes, updateRelationTypes, err := rts.handleRelationTypeImportMode(ctx, mode, relationTypes)
	if err != nil {
		return []string{}, err
	}

	// 1. 创建模型
	rtIDs := []string{}
	for _, relationType := range createRelationTypes {
		rtIDs = append(rtIDs, relationType.RTID)
		err = rts.rta.CreateRelationType(ctx, tx, relationType)
		if err != nil {
			logger.Errorf("CreateRelationType error: %s", err.Error())
			span.SetStatus(codes.Error, "创建关系类失败")
			return []string{}, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				berrors.BknBackend_RelationType_InternalError).
				WithErrorDetails(err.Error())
		}
	}

	// 更新
	for _, relationType := range updateRelationTypes {
		err = rts.UpdateRelationType(ctx, tx, relationType)
		if err != nil {
			return []string{}, err
		}
	}

	insetRelationTypes := createRelationTypes
	insetRelationTypes = append(insetRelationTypes, updateRelationTypes...)
	err = rts.InsertDatasetData(ctx, insetRelationTypes)
	if err != nil {
		logger.Errorf("InsertDatasetData error: %s", err.Error())
		span.SetStatus(codes.Error, "关系类索引写入失败")
		return []string{}, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_RelationType_InternalError_InsertOpenSearchDataFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return rtIDs, nil
}

func (rts *relationTypeService) ListRelationTypes(ctx context.Context,
	query interfaces.RelationTypesQueryParams) ([]*interfaces.RelationType, int, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "查询关系类列表")
	defer span.End()

	// 判断userid是否有查看业务知识网络的权限
	err := rts.ps.CheckPermission(ctx, interfaces.Resource{
		Type: interfaces.RESOURCE_TYPE_KN,
		ID:   query.KNID,
	}, []string{interfaces.OPERATION_TYPE_VIEW_DETAIL})
	if err != nil {
		return []*interfaces.RelationType{}, 0, err
	}

	//获取关系类列表
	relationTypes, err := rts.rta.ListRelationTypes(ctx, query)
	if err != nil {
		logger.Errorf("ListRelationTypes error: %s", err.Error())
		span.SetStatus(codes.Error, "List relation types error")

		return []*interfaces.RelationType{}, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_RelationType_InternalError).WithErrorDetails(err.Error())
	}
	if len(relationTypes) == 0 {
		span.SetStatus(codes.Ok, "")
		return relationTypes, 0, nil
	}

	// 把起点终点对象类的名称拿到
	for _, relationType := range relationTypes {
		// 起点终点对象类的名称拿到
		objectTypeMap, err := rts.ots.GetObjectTypesMapByIDs(ctx, query.KNID, query.Branch,
			[]string{relationType.SourceObjectTypeID, relationType.TargetObjectTypeID}, true)
		if err != nil {
			return []*interfaces.RelationType{}, 0, err
		}

		sourceObj := objectTypeMap[relationType.SourceObjectTypeID]
		targetObj := objectTypeMap[relationType.TargetObjectTypeID]

		if sourceObj != nil {
			relationType.SourceObjectType = interfaces.SimpleObjectType{
				OTID:   relationType.SourceObjectTypeID,
				OTName: sourceObj.OTName,
				Icon:   sourceObj.Icon,
				Color:  sourceObj.Color,
			}
		}
		if targetObj != nil {
			relationType.TargetObjectType = interfaces.SimpleObjectType{
				OTID:   relationType.TargetObjectTypeID,
				OTName: targetObj.OTName,
				Icon:   targetObj.Icon,
				Color:  targetObj.Color,
			}
		}
	}
	total := len(relationTypes)

	// limit = -1,则返回所有
	if query.Limit != -1 {

		// 分页
		// 检查起始位置是否越界
		if query.Offset < 0 || query.Offset >= len(relationTypes) {
			span.SetStatus(codes.Ok, "")
			return []*interfaces.RelationType{}, total, nil
		}
		// 计算结束位置
		end := query.Offset + query.Limit
		if end > len(relationTypes) {
			end = len(relationTypes)
		}
		relationTypes = relationTypes[query.Offset:end]
	}

	accountInfos := make([]*interfaces.AccountInfo, 0, len(relationTypes)*2)
	for _, relationType := range relationTypes {
		accountInfos = append(accountInfos, &relationType.Creator, &relationType.Updater)
	}

	err = rts.ums.GetAccountNames(ctx, accountInfos)
	if err != nil {
		span.SetStatus(codes.Error, "GetAccountNames error")

		return []*interfaces.RelationType{}, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_RelationType_InternalError).WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return relationTypes, total, nil
}

func (rts *relationTypeService) GetRelationTypesByIDs(ctx context.Context, knID string, branch string, rtIDs []string) ([]*interfaces.RelationType, error) {
	// 获取关系类
	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("查询关系类[%v]信息", rtIDs))
	defer span.End()

	// 判断userid是否有查看业务知识网络的权限
	err := rts.ps.CheckPermission(ctx, interfaces.Resource{
		Type: interfaces.RESOURCE_TYPE_KN,
		ID:   knID,
	}, []string{interfaces.OPERATION_TYPE_VIEW_DETAIL})
	if err != nil {
		return []*interfaces.RelationType{}, err
	}

	// id去重后再查
	rtIDs = common.DuplicateSlice(rtIDs)

	// 获取模型基本信息
	relationTypes, err := rts.rta.GetRelationTypesByIDs(ctx, knID, branch, rtIDs)
	if err != nil {
		logger.Errorf("GetRelationTypesByRTIDs error: %s", err.Error())
		span.SetStatus(codes.Error, fmt.Sprintf("Get relation types[%v] error: %v", rtIDs, err))

		return []*interfaces.RelationType{}, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_RelationType_InternalError_GetRelationTypesByIDsFailed).
			WithErrorDetails(err.Error())
	}

	if len(relationTypes) != len(rtIDs) {
		errStr := fmt.Sprintf("Exists any relation types not found, expect relation type nums is [%d], actual relation types num is [%d]", len(rtIDs), len(relationTypes))
		logger.Errorf(errStr)
		span.SetStatus(codes.Error, errStr)

		return []*interfaces.RelationType{}, rest.NewHTTPError(ctx, http.StatusNotFound,
			berrors.BknBackend_RelationType_RelationTypeNotFound).WithErrorDetails(errStr)
	}

	// 把起点终点对象类的名称拿到
	for _, relationType := range relationTypes {
		// 起点终点对象类的名称拿到
		objectTypeMap, err := rts.ots.GetObjectTypesMapByIDs(ctx, knID, branch,
			[]string{relationType.SourceObjectTypeID, relationType.TargetObjectTypeID}, true)
		if err != nil {
			return []*interfaces.RelationType{}, err
		}

		sourceObj := objectTypeMap[relationType.SourceObjectTypeID]
		targetObj := objectTypeMap[relationType.TargetObjectTypeID]

		// 映射字段的翻译
		switch relationType.Type {
		case interfaces.RELATION_TYPE_DIRECT:
			// 若都没有，不翻译，继续往下
			if sourceObj == nil && targetObj == nil {
				continue
			}

			// 源属性来自于源对象类。只绑数据属性，所以只需构造数据属性的map
			// 映射里的source字段名加上显示名
			for k, m := range relationType.MappingRules.([]interfaces.Mapping) {
				if sourceObj != nil {
					relationType.SourceObjectType = interfaces.SimpleObjectType{
						OTID:   relationType.SourceObjectTypeID,
						OTName: sourceObj.OTName,
						Icon:   sourceObj.Icon,
						Color:  sourceObj.Color,
					}
					// 映射里的source字段名加上显示名
					relationType.MappingRules.([]interfaces.Mapping)[k].SourceProp.DisplayName = sourceObj.PropertyMap[m.SourceProp.Name]
				}
				if targetObj != nil {
					relationType.TargetObjectType = interfaces.SimpleObjectType{
						OTID:   relationType.TargetObjectTypeID,
						OTName: targetObj.OTName,
						Icon:   targetObj.Icon,
						Color:  targetObj.Color,
					}
					// 映射里的target字段名加上显示名
					relationType.MappingRules.([]interfaces.Mapping)[k].TargetProp.DisplayName = targetObj.PropertyMap[m.TargetProp.Name]
				}
			}

		case interfaces.RELATION_TYPE_DATA_VIEW:
			// 查视图，翻译视图名称和视图字段显示名
			mappingRules := relationType.MappingRules.(interfaces.InDirectMapping)
			dataView, err := rts.dva.GetDataViewByID(ctx, mappingRules.BackingDataSource.ID)
			if err != nil {
				return []*interfaces.RelationType{}, rest.NewHTTPError(ctx, http.StatusInternalServerError,
					berrors.BknBackend_RelationType_InternalError_GetDataViewByIDFailed).
					WithErrorDetails(err.Error())
			}
			if dataView == nil {
				o11y.Warn(ctx, fmt.Sprintf("Relation type [%s]'s Backing Data view %s not found", relationType.RTID, mappingRules.BackingDataSource.ID))
				// 若都没有，不翻译，遍历下一个
				if sourceObj == nil && targetObj == nil {
					continue
				}
			} else {
				relationType.MappingRules.(interfaces.InDirectMapping).BackingDataSource.Name = dataView.ViewName
			}

			// 起点到视图
			for k, m := range relationType.MappingRules.(interfaces.InDirectMapping).SourceMappingRules {
				if sourceObj != nil {
					relationType.SourceObjectType = interfaces.SimpleObjectType{
						OTID:   relationType.SourceObjectTypeID,
						OTName: sourceObj.OTName,
						Icon:   sourceObj.Icon,
						Color:  sourceObj.Color,
					}
					// 映射里的source字段名加上显示名
					relationType.MappingRules.(interfaces.InDirectMapping).SourceMappingRules[k].
						SourceProp.DisplayName = sourceObj.PropertyMap[m.SourceProp.Name]
				}
				if dataView != nil {
					// 映射里的target字段名加上显示名
					relationType.MappingRules.(interfaces.InDirectMapping).SourceMappingRules[k].
						TargetProp.DisplayName = dataView.FieldsMap[m.TargetProp.Name].DisplayName
				}
			}

			// 视图到终点
			for k, m := range relationType.MappingRules.(interfaces.InDirectMapping).TargetMappingRules {
				if dataView != nil {
					// 映射里的target字段名加上显示名
					relationType.MappingRules.(interfaces.InDirectMapping).TargetMappingRules[k].
						SourceProp.DisplayName = dataView.FieldsMap[m.SourceProp.Name].DisplayName
				}
				if targetObj != nil {
					relationType.TargetObjectType = interfaces.SimpleObjectType{
						OTID:   relationType.TargetObjectTypeID,
						OTName: targetObj.OTName,
						Icon:   targetObj.Icon,
						Color:  targetObj.Color,
					}
					// 映射里的target字段名加上显示名
					relationType.MappingRules.(interfaces.InDirectMapping).TargetMappingRules[k].
						TargetProp.DisplayName = targetObj.PropertyMap[m.TargetProp.Name]
				}
			}
		}
	}

	span.SetStatus(codes.Ok, "")
	return relationTypes, nil
}

// 更新关系类
func (rts *relationTypeService) UpdateRelationType(ctx context.Context, tx *sql.Tx, relationType *interfaces.RelationType) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Update relation type")
	defer span.End()

	// 判断userid是否有修改业务知识网络的权限
	err := rts.ps.CheckPermission(ctx, interfaces.Resource{
		Type: interfaces.RESOURCE_TYPE_KN,
		ID:   relationType.KNID,
	}, []string{interfaces.OPERATION_TYPE_MODIFY})
	if err != nil {
		return err
	}

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}
	relationType.Updater = accountInfo

	currentTime := time.Now().UnixMilli() // 关系类的update_time是int类型
	relationType.UpdateTime = currentTime

	bknRel := logics.ToBKNRelationType(relationType)
	relationType.BKNRawContent = bknsdk.SerializeRelationType(bknRel)

	if tx == nil {
		// 0. 开始事务
		tx, err = rts.db.Begin()
		if err != nil {
			logger.Errorf("Begin transaction error: %s", err.Error())
			span.SetStatus(codes.Error, "事务开启失败")
			o11y.Error(ctx, fmt.Sprintf("Begin transaction error: %s", err.Error()))

			return rest.NewHTTPError(ctx, http.StatusInternalServerError,
				berrors.BknBackend_RelationType_InternalError_BeginTransactionFailed).
				WithErrorDetails(err.Error())
		}
		// 0.1 异常时
		defer func() {
			switch err {
			case nil:
				// 提交事务
				err = tx.Commit()
				if err != nil {
					logger.Errorf("UpdateRelationType Transaction Commit Failed:%v", err)
					span.SetStatus(codes.Error, "提交事务失败")
					o11y.Error(ctx, fmt.Sprintf("UpdateRelationType Transaction Commit Failed: %s", err.Error()))
				}
				logger.Infof("UpdateRelationType Transaction Commit Success:%v", relationType.RTName)
				o11y.Debug(ctx, fmt.Sprintf("UpdateRelationType Transaction Commit Success: %s", relationType.RTName))
			default:
				rollbackErr := tx.Rollback()
				if rollbackErr != nil {
					logger.Errorf("UpdateRelationType Transaction Rollback Error:%v", rollbackErr)
					span.SetStatus(codes.Error, "事务回滚失败")
					o11y.Error(ctx, fmt.Sprintf("UpdateRelationType Transaction Rollback Error: %s", err.Error()))
				}
			}
		}()
	}

	// 校验起点对象类、终点对象类非空时，需校验存在性
	// 更新操作默认进行依赖校验
	err = rts.validateDependency(ctx, tx, relationType, true)
	if err != nil {
		return err
	}

	// 更新模型信息
	err = rts.rta.UpdateRelationType(ctx, tx, relationType)
	if err != nil {
		logger.Errorf("relationType error: %s", err.Error())
		span.SetStatus(codes.Error, "修改关系类失败")

		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_RelationType_InternalError).
			WithErrorDetails(err.Error())
	}

	err = rts.InsertDatasetData(ctx, []*interfaces.RelationType{relationType})
	if err != nil {
		logger.Errorf("InsertDatasetData error: %s", err.Error())
		span.SetStatus(codes.Error, "关系类索引写入失败")

		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_RelationType_InternalError_InsertOpenSearchDataFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

func (rts *relationTypeService) DeleteRelationTypesByIDs(ctx context.Context, tx *sql.Tx, knID string, branch string, rtIDs []string) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Delete relation types")
	defer span.End()

	// 判断userid是否有修改业务知识网络的权限
	err := rts.ps.CheckPermission(ctx, interfaces.Resource{
		Type: interfaces.RESOURCE_TYPE_KN,
		ID:   knID,
	}, []string{interfaces.OPERATION_TYPE_MODIFY})
	if err != nil {
		return err
	}

	if tx == nil {
		// 0. 开始事务
		tx, err = rts.db.Begin()
		if err != nil {
			logger.Errorf("Begin transaction error: %s", err.Error())
			span.SetStatus(codes.Error, "事务开启失败")
			o11y.Error(ctx, fmt.Sprintf("Begin transaction error: %s", err.Error()))

			return rest.NewHTTPError(ctx, http.StatusInternalServerError,
				berrors.BknBackend_RelationType_InternalError_BeginTransactionFailed).
				WithErrorDetails(err.Error())
		}
		// 0.1 异常时
		defer func() {
			switch err {
			case nil:
				// 提交事务
				err = tx.Commit()
				if err != nil {
					logger.Errorf("DeleteRelationTypes Transaction Commit Failed:%v", err)
					span.SetStatus(codes.Error, "提交事务失败")
					o11y.Error(ctx, fmt.Sprintf("DeleteRelationTypes Transaction Commit Failed: %s", err.Error()))
				}
				logger.Infof("DeleteRelationTypes Transaction Commit Success: kn_id:%s,ot_ids:%v", knID, rtIDs)
				o11y.Debug(ctx, fmt.Sprintf("DeleteRelationTypes Transaction Commit Success: kn_id:%s,ot_ids:%v", knID, rtIDs))
			default:
				rollbackErr := tx.Rollback()
				if rollbackErr != nil {
					logger.Errorf("DeleteRelationTypes Transaction Rollback Error:%v", rollbackErr)
					span.SetStatus(codes.Error, "事务回滚失败")
					o11y.Error(ctx, fmt.Sprintf("DeleteRelationTypes Transaction Rollback Error: %s", rollbackErr.Error()))
				}
			}
		}()
	}

	// 删除指标模型
	rowsAffect, err := rts.rta.DeleteRelationTypesByIDs(ctx, tx, knID, branch, rtIDs)
	if err != nil {
		logger.Errorf("DeleteRelationTypes error: %s", err.Error())
		span.SetStatus(codes.Error, "删除关系类失败")

		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_RelationType_InternalError).WithErrorDetails(err.Error())
	}

	logger.Infof("DeleteRelationTypes: Rows affected is %v, request delete RTIDs is %v!", rowsAffect, len(rtIDs))
	if rowsAffect != int64(len(rtIDs)) {
		logger.Warnf("Delete relation types number %v not equal requerst relation types number %v!", rowsAffect, len(rtIDs))
		o11y.Warn(ctx, fmt.Sprintf("Delete relation types number %v not equal requerst relation types number %v!", rowsAffect, len(rtIDs)))
	}

	for _, rtID := range rtIDs {
		docid := interfaces.GenerateConceptDocuemtnID(knID, interfaces.MODULE_TYPE_RELATION_TYPE, rtID, branch)
		err = rts.vba.DeleteDatasetDocumentByID(ctx, interfaces.BKN_DATASET_ID, docid)
		if err != nil {
			logger.Errorf("DeleteDatasetDocumentByID error: %s", err.Error())
			span.SetStatus(codes.Error, "删除关系类概念索引失败")
			return err
		}
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 内部接口，根据业务知识网络ID删除所有关系类，不校验权限，tx必须传入
func (rts *relationTypeService) DeleteRelationTypesByKnID(ctx context.Context, tx *sql.Tx, knID string, branch string) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Delete relation types by kn_id")
	defer span.End()

	if tx == nil {
		logger.Errorf("missing transaction")
		o11y.Error(ctx, "missing transaction")
		span.SetStatus(codes.Error, "缺少事务")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_RelationType_InternalError_MissingTransaction).
			WithErrorDetails("missing transaction")
	}

	// 删除指标模型
	rowsAffect, err := rts.rta.DeleteRelationTypesByKnID(ctx, tx, knID, branch)
	if err != nil {
		logger.Errorf("DeleteRelationTypesByKnID error: %s", err.Error())
		span.SetStatus(codes.Error, "删除关系类失败")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_RelationType_InternalError).WithErrorDetails(err.Error())
	}

	logger.Infof("DeleteRelationTypesByKnID success, the kn_id is [%s], branch is [%s], rowsAffect is [%d]",
		knID, branch, rowsAffect)
	span.SetStatus(codes.Ok, "")
	return nil
}

func (rts *relationTypeService) handleRelationTypeImportMode(ctx context.Context, mode string,
	relationTypes []*interfaces.RelationType) ([]*interfaces.RelationType, []*interfaces.RelationType, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "relation type import mode logic")
	defer span.End()

	creates := []*interfaces.RelationType{}
	updates := []*interfaces.RelationType{}

	// 3. 校验 若模型的id不为空，则用请求体的id与现有模型ID的重复性
	for _, relationType := range relationTypes {
		creates = append(creates, relationType)
		_, idExist, err := rts.CheckRelationTypeExistByID(ctx, relationType.KNID, relationType.Branch, relationType.RTID)
		if err != nil {
			return creates, updates, err
		}

		// 根据mode来区别，若是ignore，就从结果集中忽略，若是overwrite，就调用update，若是normal就报错。
		if idExist {
			switch mode {
			case interfaces.ImportMode_Normal:
				errDetails := fmt.Sprintf("The relation type with id [%s] already exists!", relationType.RTID)
				logger.Error(errDetails)
				span.SetStatus(codes.Error, errDetails)
				return creates, updates, rest.NewHTTPError(ctx, http.StatusBadRequest,
					berrors.BknBackend_RelationType_RelationTypeIDExisted).
					WithErrorDetails(errDetails)

			case interfaces.ImportMode_Ignore:
				// ID 已存在则跳过，从create数组中删除
				creates = creates[:len(creates)-1]

			case interfaces.ImportMode_Overwrite:
				// ID 已存在则覆盖更新，从create数组中删除, 放到更新数组中
				creates = creates[:len(creates)-1]
				updates = append(updates, relationType)
			}
		}
	}
	span.SetStatus(codes.Ok, "")
	return creates, updates, nil
}

func (rts *relationTypeService) InsertDatasetData(ctx context.Context, relationTypes []*interfaces.RelationType) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "关系类索引写入")
	defer span.End()

	// 关系类索引写入
	if len(relationTypes) == 0 {
		return nil
	}

	if rts.appSetting.ServerSetting.DefaultSmallModelEnabled {
		words := []string{}
		for _, relationType := range relationTypes {
			arr := []string{relationType.RTName}
			arr = append(arr, relationType.Tags...)
			arr = append(arr, relationType.Comment, relationType.BKNRawContent)
			word := strings.Join(arr, "\n")
			words = append(words, word)
		}

		dftModel, err := rts.mfa.GetDefaultModel(ctx)
		if err != nil {
			logger.Errorf("GetDefaultModel error: %s", err.Error())
			span.SetStatus(codes.Error, "获取默认模型失败")
			return err
		}
		vectors, err := rts.mfa.GetVector(ctx, dftModel, words)
		if err != nil {
			logger.Errorf("GetVector error: %s", err.Error())
			span.SetStatus(codes.Error, "获取关系类向量失败")
			return err
		}

		if len(vectors) != len(relationTypes) {
			logger.Errorf("GetVector error: expect vectors num is [%d], actual vectors num is [%d]", len(relationTypes), len(vectors))
			span.SetStatus(codes.Error, "获取关系类向量失败")
			return fmt.Errorf("GetVector error: expect vectors num is [%d], actual vectors num is [%d]", len(relationTypes), len(vectors))
		}

		for i, relationType := range relationTypes {
			relationType.Vector = vectors[i].Vector
		}
	}

	documents := []map[string]any{}
	for _, relationType := range relationTypes {
		docid := interfaces.GenerateConceptDocuemtnID(relationType.KNID, interfaces.MODULE_TYPE_RELATION_TYPE,
			relationType.RTID, relationType.Branch)
		relationType.ModuleType = interfaces.MODULE_TYPE_RELATION_TYPE

		// Convert to map for dataset
		docBytes, err := sonic.Marshal(relationType)
		if err != nil {
			logger.Errorf("Failed to marshal RelationType: %s", err.Error())
			span.SetStatus(codes.Error, "序列化关系类失败")
			return err
		}

		var doc map[string]any
		if err := sonic.Unmarshal(docBytes, &doc); err != nil {
			logger.Errorf("Failed to unmarshal RelationType: %s", err.Error())
			span.SetStatus(codes.Error, "反序列化关系类失败")
			return err
		}

		// Set document ID
		doc["_id"] = docid
		documents = append(documents, doc)
	}

	err := rts.vba.WriteDatasetDocuments(ctx, interfaces.BKN_DATASET_ID, documents)
	if err != nil {
		logger.Errorf("WriteDatasetDocuments error: %s", err.Error())
		span.SetStatus(codes.Error, "关系类概念索引写入失败")
		return err
	}

	return nil
}

func (rts *relationTypeService) SearchRelationTypes(ctx context.Context,
	query *interfaces.ConceptsQuery) (interfaces.RelationTypes, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "业务知识网络关系类检索")
	defer span.End()

	response := interfaces.RelationTypes{}
	var err error

	// 判断userid是否有查看业务知识网络的权限
	err = rts.ps.CheckPermission(ctx, interfaces.Resource{
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
				if !rts.appSetting.ServerSetting.DefaultSmallModelEnabled {
					err = errors.New(cond.DEFAULT_SMALL_MODEL_ENABLED_FALSE_ERROR)
					span.SetStatus(codes.Error, err.Error())
					return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
						berrors.BknBackend_RelationType_InternalError).
						WithErrorDetails(err.Error())
				}
				dftModel, err := rts.mfa.GetDefaultModel(ctx)
				if err != nil {
					logger.Errorf("GetDefaultModel error: %s", err.Error())
					span.SetStatus(codes.Error, "获取默认模型失败")
					return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
						berrors.BknBackend_RelationType_InternalError).
						WithErrorDetails(err.Error())
				}
				result, err := rts.mfa.GetVector(ctx, dftModel, []string{word})
				if err != nil {
					logger.Errorf("GetVector error: %s", err.Error())
					span.SetStatus(codes.Error, "获取业务知识网络向量失败")
					return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
						berrors.BknBackend_RelationType_InternalError).
						WithErrorDetails(err.Error())
				}
				return result, nil
			})
		if err != nil {
			return response, rest.NewHTTPError(ctx, http.StatusBadRequest,
				berrors.BknBackend_RelationType_InvalidParameter_ConceptCondition).
				WithErrorDetails(fmt.Sprintf("failed to convert condition to filter condition, %s", err.Error()))
		}
	}

	// 1. 获取组下的关系类
	rtIDMap := map[string]bool{} // 分组下的对象类id
	rtIDs := []string{}          // 不同组下的对象类可以重叠，所以需要对对象类id的数组去重
	if len(query.ConceptGroups) > 0 {
		// 校验分组是否都存在，按分组id获取分组
		cgCnt, err := rts.cga.GetConceptGroupsTotal(ctx, interfaces.ConceptGroupsQueryParams{
			KNID:   query.KNID,
			Branch: query.Branch,
			CGIDs:  query.ConceptGroups,
		})
		if err != nil {
			logger.Errorf("GetConceptGroupsTotal in knowledge network[%s] error: %s", query.KNID, err.Error())
			span.SetStatus(codes.Error, fmt.Sprintf("GetConceptGroupsTotal in knowledge network[%s], error: %v", query.KNID, err))

			return response, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				berrors.BknBackend_RelationType_InternalError).WithErrorDetails(err.Error())
		}
		if cgCnt == 0 {
			errStr := fmt.Sprintf("all concept group not found, expect concept group nums is [%d], actual concept group num is [%d]",
				cgCnt, len(query.ConceptGroups))
			logger.Errorf(errStr)

			return response, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				berrors.BknBackend_RelationType_InternalError).
				WithErrorDetails(errStr)
		}
		// 在当前业务知识网络下查找属于请求的分组范围内的关系类ID
		rtIDArr, err := rts.cga.GetRelationTypeIDsFromConceptGroupRelation(ctx, interfaces.ConceptGroupRelationsQueryParams{
			KNID:        query.KNID,
			Branch:      query.Branch,
			ConceptType: interfaces.MODULE_TYPE_OBJECT_TYPE, // 概念与分组关系中的概念类型
			CGIDs:       query.ConceptGroups,
		})
		if err != nil {
			errStr := fmt.Sprintf("GetRelationTypeIDsFromConceptGroupRelation failed, kn_id:[%s],branch:[%s],cg_ids:[%v], error: %v",
				query.KNID, query.Branch, query.ConceptGroups, err)
			logger.Errorf(errStr)
			span.SetStatus(codes.Error, errStr)
			span.End()

			return response, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				berrors.BknBackend_RelationType_InternalError).WithErrorDetails(errStr)
		}
		// 概念分组下没有关系类,返回空
		if len(rtIDArr) == 0 {
			return response, nil
		}

		for _, rtID := range rtIDArr {
			if !rtIDMap[rtID] {
				rtIDMap[rtID] = true
				rtIDs = append(rtIDs, rtID)
			}
		}
	}

	// 根据NeedTotal参数决定是否查询total
	if query.NeedTotal {
		if len(rtIDMap) == 0 {
			// 未指定分组，直接搜索.总数从dataset的结果中读取
			params := &interfaces.DatasetQueryParams{
				FilterCondition: filterCondition,
				Offset:          0,
				Limit:           1, // 查询1条数据，获取total
				NeedTotal:       true,
			}
			datasetResp, err := rts.vba.QueryDatasetData(ctx, interfaces.BKN_DATASET_ID, params)
			if err != nil {
				logger.Errorf("QueryDatasetData error: %s", err.Error())
				span.SetStatus(codes.Error, "业务知识网络关系类检索查询总数失败")
				return response, rest.NewHTTPError(ctx, http.StatusInternalServerError,
					berrors.BknBackend_RelationType_InternalError).
					WithErrorDetails(err.Error())
			}
			response.TotalCount = datasetResp.TotalCount
		} else {
			// 指定了分组，需要查询分组内且符合条件的总数
			total, err := rts.GetTotalWithLargeRTIDs(ctx, filterCondition, rtIDs)
			if err != nil {
				return response, err
			}
			response.TotalCount = total
		}
	}

	// 4. 迭代查询直到获取足够数量或没有更多数据
	relationTypes := []*interfaces.RelationType{}
	var totalFilteredCount int64 = 0
	offset := 0
	limit := query.Limit
	if limit == 0 {
		limit = interfaces.SearchAfter_Limit
	}

	for {
		// 调用 dataset 查询
		params := &interfaces.DatasetQueryParams{
			FilterCondition: filterCondition,
			Offset:          offset,
			Limit:           limit,
			NeedTotal:       true,
			Sort:            query.Sort,
		}
		datasetResp, err := rts.vba.QueryDatasetData(ctx, interfaces.BKN_DATASET_ID, params)
		if err != nil {
			logger.Errorf("QueryDatasetData error: %s", err.Error())
			span.SetStatus(codes.Error, "业务知识网络关系类检索查询失败")
			return response, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				berrors.BknBackend_RelationType_InternalError).
				WithErrorDetails(err.Error())
		}

		// 如果没有数据了，跳出循环
		if len(datasetResp.Entries) == 0 {
			break
		}

		// 5. 处理查询结果
		for _, entry := range datasetResp.Entries {
			// 转成 relation type 的 struct
			jsonByte, err := json.Marshal(entry)
			if err != nil {
				return response, rest.NewHTTPError(ctx, http.StatusBadRequest,
					berrors.BknBackend_InternalError_MarshalDataFailed).
					WithErrorDetails(fmt.Sprintf("failed to Marshal dataset entry, %s", err.Error()))
			}
			var relationType interfaces.RelationType
			err = json.Unmarshal(jsonByte, &relationType)
			if err != nil {
				return response, rest.NewHTTPError(ctx, http.StatusBadRequest,
					berrors.BknBackend_InternalError_UnMarshalDataFailed).
					WithErrorDetails(fmt.Sprintf("failed to Unmarshal dataset entry to Relation Type, %s", err.Error()))
			}

			// 如果没有指定分组，或者关系类属于分组，则添加
			if len(rtIDMap) == 0 || rtIDMap[relationType.RTID] {
				// 提取 _score（如果有）
				if scoreVal, ok := entry["_score"]; ok {
					if scoreFloat, ok := scoreVal.(float64); ok {
						score := float64(scoreFloat)
						relationType.Score = &score
					}
				}
				relationType.Vector = nil
				relationTypes = append(relationTypes, &relationType)
				totalFilteredCount++

				// 如果已经收集到足够的数量，跳出循环
				if len(relationTypes) >= query.Limit && query.Limit > 0 {
					break
				}
			}
		}
		query.SearchAfter = datasetResp.SearchAfter

		// 如果已经收集到足够的数量或者没有更多数据了，跳出循环
		if (query.Limit > 0 && len(relationTypes) >= query.Limit) || len(datasetResp.Entries) < limit {
			break
		}

		// 更新 offset 用于下一次查询（当前使用 offset 分页作为临时方案）
		offset += limit
	}

	response.Entries = relationTypes
	response.SearchAfter = query.SearchAfter
	return response, nil
}

func (rts *relationTypeService) GetTotal(ctx context.Context, filterCondition map[string]any) (total int64, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: search relation type total ")
	defer span.End()

	params := &interfaces.DatasetQueryParams{
		FilterCondition: filterCondition,
		Offset:          0,
		Limit:           1, // 查询1条数据，获取total
		NeedTotal:       true,
	}
	datasetResp, err := rts.vba.QueryDatasetData(ctx, interfaces.BKN_DATASET_ID, params)
	if err != nil {
		span.SetStatus(codes.Error, "Search total documents count failed")
		return total, rest.NewHTTPError(ctx, http.StatusInternalServerError, berrors.BknBackend_RelationType_InternalError).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")

	if datasetResp == nil {
		return 0, nil
	}
	return datasetResp.TotalCount, nil
}

// 内部调用，不加权限校验
func (rts *relationTypeService) GetRelationTypeIDsByKnID(ctx context.Context, knID string, branch string) ([]string, error) {
	// 获取关系类
	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("按kn_id[%s]获取关系类IDs", knID))
	defer span.End()

	// 获取对象类基本信息
	rtIDs, err := rts.rta.GetRelationTypeIDsByKnID(ctx, knID, branch)
	if err != nil {
		logger.Errorf("GetRelationTypeIDsByKnID error: %s", err.Error())
		span.SetStatus(codes.Error, fmt.Sprintf("Get relation type ids by kn_id[%s] error: %v", knID, err))

		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_RelationType_InternalError_GetRelationTypesByIDsFailed).WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return rtIDs, nil
}

// 分批查询
func (rts *relationTypeService) GetTotalWithLargeRTIDs(ctx context.Context,
	filterCondition map[string]any,
	rtIDs []string) (int64, error) {

	total := int64(0)
	for i := 0; i < len(rtIDs); i += interfaces.GET_TOTAL_CONCEPTID_BATCH_SIZE {
		end := i + interfaces.GET_TOTAL_CONCEPTID_BATCH_SIZE
		if end > len(rtIDs) {
			end = len(rtIDs)
		}

		batchIDs := rtIDs[i:end]
		batchTotal, err := rts.GetTotalWithRTIDs(ctx, filterCondition, batchIDs)
		if err != nil {
			return 0, err
		}

		total += batchTotal
	}

	return total, nil
}

// 查询指定关系类ID列表的关系类总数
func (rts *relationTypeService) GetTotalWithRTIDs(ctx context.Context,
	filterCondition map[string]any,
	rtIDs []string) (int64, error) {

	// 构建包含 RTID 过滤的 filter condition
	rtIDCondition := map[string]any{
		"field":      "id",
		"operation":  "in",
		"value":      rtIDs,
		"value_from": "const",
	}

	var combinedCondition map[string]any
	if filterCondition == nil {
		combinedCondition = rtIDCondition
	} else {
		combinedCondition = map[string]any{
			"operation": "and",
			"sub_conditions": []map[string]any{
				filterCondition,
				rtIDCondition,
			},
		}
	}

	// 执行计数查询
	total, err := rts.GetTotal(ctx, combinedCondition)
	if err != nil {
		return total, err
	}

	return total, nil
}

// 校验关系类相关的对象类、数据视图存在性
func (rts *relationTypeService) validateDependency(ctx context.Context, tx *sql.Tx, relationType *interfaces.RelationType, validateDependency bool) error {
	var sourceObjectType *interfaces.ObjectType
	var targetObjectType *interfaces.ObjectType
	var err error
	if relationType.SourceObjectTypeID != "" {
		// 导入时未提交，在一个事务中get
		sourceObjectType, err = rts.ots.GetObjectTypeByID(ctx, tx, relationType.KNID, relationType.Branch, relationType.SourceObjectTypeID)
		if err != nil {
			return err
		}
	}
	if relationType.TargetObjectTypeID != "" {
		targetObjectType, err = rts.ots.GetObjectTypeByID(ctx, tx, relationType.KNID, relationType.Branch, relationType.TargetObjectTypeID)
		if err != nil {
			return err
		}
	}
	// 当关联关系非空时，校验起点对象类、终点对象类的属性存在性
	if relationType.MappingRules != nil {
		switch relationType.Type {
		case interfaces.RELATION_TYPE_DIRECT:
			directMappingRules := relationType.MappingRules.([]interfaces.Mapping)
			for _, mapping := range directMappingRules {
				if sourceObjectType != nil {
					// 检查起点属性是否在起点对象类的数据属性中存在
					if _, exist := sourceObjectType.PropertyMap[mapping.SourceProp.Name]; !exist {
						return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_RelationType_InvalidParameter).
							WithErrorDetails(fmt.Sprintf("起点关联属性[%s]在起点对象类[%s]中不存在", mapping.SourceProp.Name, sourceObjectType.OTName))
					}
				}

				if targetObjectType != nil {
					// 检查终点属性是否在终点对象类的数据属性中存在
					if _, exist := targetObjectType.PropertyMap[mapping.TargetProp.Name]; !exist {
						return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_RelationType_InvalidParameter).
							WithErrorDetails(fmt.Sprintf("终点关联属性[%s]在终点对象类[%s]中不存在", mapping.TargetProp.Name, targetObjectType.OTName))
					}
				}
			}
		case interfaces.RELATION_TYPE_DATA_VIEW:
			inDirectMappingRules := relationType.MappingRules.(interfaces.InDirectMapping)
			// validateDependency为true时才校验数据视图存在性,不需要校验时，跳过后续的字段校验
			if validateDependency && inDirectMappingRules.BackingDataSource != nil && inDirectMappingRules.BackingDataSource.ID != "" {
				dataView, err := rts.dva.GetDataViewByID(ctx, inDirectMappingRules.BackingDataSource.ID)
				if err != nil {
					return err
				}
				if validateDependency && dataView == nil {
					return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_RelationType_InvalidParameter).
						WithErrorDetails(fmt.Sprintf("关系类中的[%s]数据视图[%s]不存在", relationType.RTID, inDirectMappingRules.BackingDataSource.ID))
				}

				// 校验起点属性在起点对象类中存在，校验关联字段在视图中存在
				for _, mapping := range inDirectMappingRules.SourceMappingRules {
					if sourceObjectType != nil {
						// 检查起点属性是否在起点对象类的数据属性中存在
						if _, exist := sourceObjectType.PropertyMap[mapping.SourceProp.Name]; !exist {
							return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_RelationType_InvalidParameter).
								WithErrorDetails(fmt.Sprintf("起点关联属性[%s]在起点对象类[%s]中不存在", mapping.SourceProp.Name, sourceObjectType.OTName))
						}
					}

					// 检查关联字段是否在视图中存在
					if _, exist := dataView.FieldsMap[mapping.TargetProp.Name]; !exist {
						return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_RelationType_InvalidParameter).
							WithErrorDetails(fmt.Sprintf("中间关联字段[%s]在视图[%s]中不存在", mapping.SourceProp.Name, dataView.ViewName))
					}
				}
				for _, mapping := range inDirectMappingRules.TargetMappingRules {
					if targetObjectType != nil {
						// 检查中间关联字段是否在视图中存在
						if _, exist := dataView.FieldsMap[mapping.SourceProp.Name]; !exist {
							return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_RelationType_InvalidParameter).
								WithErrorDetails(fmt.Sprintf("中间关联字段[%s]在视图[%s]中不存在", mapping.SourceProp.Name, dataView.ViewName))
						}
						// 检查终点属性是否在终点对象类的数据属性中存在
						if _, exist := targetObjectType.PropertyMap[mapping.TargetProp.Name]; !exist {
							return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_RelationType_InvalidParameter).
								WithErrorDetails(fmt.Sprintf("终点关联属性[%s]在终点对象类[%s]中不存在", mapping.TargetProp.Name, targetObjectType.OTName))
						}
					}
				}
			}
		}
	}
	return nil
}

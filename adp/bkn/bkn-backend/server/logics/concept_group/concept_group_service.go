// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package concept_group

import (
	"context"
	"database/sql"
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
	berrors "bkn-backend/errors"
	"bkn-backend/interfaces"
	"bkn-backend/logics"
	"bkn-backend/logics/action_type"
	"bkn-backend/logics/batchindex"
	"bkn-backend/logics/object_type"
	"bkn-backend/logics/permission"
	"bkn-backend/logics/relation_type"
	"bkn-backend/logics/user_mgmt"
)

var (
	cgServiceOnce sync.Once
	cgService     interfaces.ConceptGroupService
)

type conceptGroupService struct {
	appSetting *common.AppSetting
	db         *sql.DB
	ata        interfaces.ActionTypeAccess
	ats        interfaces.ActionTypeService
	cga        interfaces.ConceptGroupAccess
	kna        interfaces.KNAccess
	mfa        interfaces.ModelFactoryAccess
	ota        interfaces.ObjectTypeAccess
	ots        interfaces.ObjectTypeService
	rta        interfaces.RelationTypeAccess
	ps         interfaces.PermissionService
	rts        interfaces.RelationTypeService
	ums        interfaces.UserMgmtService
	vba        interfaces.VegaBackendAccess
}

func NewConceptGroupService(appSetting *common.AppSetting) interfaces.ConceptGroupService {
	cgServiceOnce.Do(func() {
		cgService = &conceptGroupService{
			appSetting: appSetting,
			ata:        logics.ATA,
			ats:        action_type.NewActionTypeService(appSetting),
			db:         logics.DB,
			cga:        logics.CGA,
			kna:        logics.KNA,
			mfa:        logics.MFA,
			ota:        logics.OTA,
			ots:        object_type.NewObjectTypeService(appSetting),
			ps:         permission.NewPermissionService(appSetting),
			rta:        logics.RTA,
			rts:        relation_type.NewRelationTypeService(appSetting),
			ums:        user_mgmt.NewUserMgmtService(appSetting),
			vba:        logics.VBA,
		}
	})
	return cgService
}

func (cgs *conceptGroupService) CheckConceptGroupExistByID(ctx context.Context, knID string, branch string,
	cgID string) (string, bool, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("校验概念分组[%v]的存在性", cgID))
	defer span.End()

	cgName, exist, err := cgs.cga.CheckConceptGroupExistByID(ctx, knID, branch, cgID)
	if err != nil {
		logger.Errorf("CheckConceptGroupExistByID error: %s", err.Error())

		span.SetStatus(codes.Error, fmt.Sprintf("按ID[%v]获取概念分组失败", knID))
		// 记录处理的 sql 字符串
		o11y.Error(ctx, fmt.Sprintf("按ID[%v]获取概念分组失败: %v", knID, err))

		return "", exist, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_ConceptGroup_InternalError_CheckConceptGroupIfExistFailed).WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return cgName, exist, nil
}

func (cgs *conceptGroupService) CheckConceptGroupExistByName(ctx context.Context, knID string, branch string, cgName string) (string, bool, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("校验概念分组[%v]的存在性", cgName))
	defer span.End()

	cgID, exist, err := cgs.cga.CheckConceptGroupExistByName(ctx, knID, branch, cgName)
	if err != nil {
		logger.Errorf("CheckConceptGroupExistByName error: %s", err.Error())

		span.SetStatus(codes.Error, fmt.Sprintf("按名称[%v]获取概念分组失败", cgName))
		// 记录处理的 sql 字符串
		o11y.Error(ctx, fmt.Sprintf("按名称[%v]获取概念分组失败: %v", cgName, err))

		return cgID, exist, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_ConceptGroup_InternalError_CheckConceptGroupIfExistFailed).WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return cgID, exist, nil
}

// 创建概念分组
func (cgs *conceptGroupService) CreateConceptGroup(ctx context.Context, tx *sql.Tx,
	conceptGroup *interfaces.ConceptGroup, mode string, strictMode bool) (string, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "Create concept group")
	defer span.End()

	// 判断userid是否有创建概念分组的权限（策略决策）
	err := cgs.ps.CheckPermission(ctx, interfaces.PermissionResource{
		Type: interfaces.RESOURCE_TYPE_KN,
		ID:   conceptGroup.KNID,
	}, []string{interfaces.OPERATION_TYPE_MODIFY})
	if err != nil {
		return "", err
	}

	currentTime := time.Now().UnixMilli()
	// 若提交的模型id为空，生成分布式ID
	if conceptGroup.CGID == "" {
		conceptGroup.CGID = xid.New().String()
	}
	otIDs := []interfaces.ID{}
	bknOtMap := map[string]*bknsdk.BknObjectType{}
	for _, objectType := range conceptGroup.ObjectTypes {
		objectType.KNID = conceptGroup.KNID
		objectType.Branch = conceptGroup.Branch

		otIDs = append(otIDs, interfaces.ID{ID: objectType.OTID})
		bknOtMap[objectType.OTID] = logics.ToBKNObjectType(objectType)
	}
	for _, relationType := range conceptGroup.RelationTypes {
		relationType.KNID = conceptGroup.KNID
		relationType.Branch = conceptGroup.Branch
	}
	for _, actionType := range conceptGroup.ActionTypes {
		actionType.KNID = conceptGroup.KNID
		actionType.Branch = conceptGroup.Branch
	}

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}
	conceptGroup.Creator = accountInfo
	conceptGroup.Updater = accountInfo

	conceptGroup.CreateTime = currentTime
	conceptGroup.UpdateTime = currentTime

	bknCG := logics.ToBKNConceptGroup(conceptGroup)
	conceptGroup.BKNRawContent = bknsdk.SerializeConceptGroup(bknCG, bknOtMap)

	if tx == nil {
		// 0. 开始事务
		tx, err = cgs.db.Begin()
		if err != nil {
			logger.Errorf("Begin transaction error: %s", err.Error())
			span.SetStatus(codes.Error, "事务开启失败")
			o11y.Error(ctx, fmt.Sprintf("Begin transaction error: %s", err.Error()))

			return "", rest.NewHTTPError(ctx, http.StatusInternalServerError,
				berrors.BknBackend_ConceptGroup_InternalError_BeginTransactionFailed).
				WithErrorDetails(err.Error())
		}

		// 0.1 异常时
		defer func() {
			switch err {
			case nil:
				// 提交事务
				err = tx.Commit()
				if err != nil {
					logger.Errorf("CreateConceptGroup Transaction Commit Failed:%v", err)
					span.SetStatus(codes.Error, "提交事务失败")
					o11y.Error(ctx, fmt.Sprintf("CreateConceptGroup Transaction Commit Failed: %s", err.Error()))
					return
				}
				logger.Infof("CreateConceptGroup Transaction Commit Success")
				o11y.Debug(ctx, "CreateConceptGroup Transaction Commit Success")
			default:
				rollbackErr := tx.Rollback()
				if rollbackErr != nil {
					logger.Errorf("CreateConceptGroup Transaction Rollback Error:%v", rollbackErr)
					span.SetStatus(codes.Error, "事务回滚失败")
					o11y.Error(ctx, fmt.Sprintf("CreateConceptGroup Transaction Rollback Error: %s", err.Error()))
				}
			}
		}()
	}

	// 处理导入模式
	isCreate, isUpdate, err := cgs.handleConceptGroupImportMode(ctx, mode, conceptGroup)
	if err != nil {
		return "", err
	}

	// 处理创建情况
	if isCreate {
		err = cgs.cga.CreateConceptGroup(ctx, tx, conceptGroup)
		if err != nil {
			logger.Errorf("CreateConceptGroup error: %s", err.Error())
			span.SetStatus(codes.Error, "创建概念分组失败")

			return "", rest.NewHTTPError(ctx, http.StatusInternalServerError,
				berrors.BknBackend_ConceptGroup_InternalError_CreateConceptGroupFailed).
				WithErrorDetails(err.Error())
		}

		if len(conceptGroup.ObjectTypes) > 0 {
			_, err = cgs.ots.CreateObjectTypes(ctx, tx, conceptGroup.ObjectTypes, mode, false, strictMode)
			if err != nil {
				logger.Errorf("CreateObjectTypes error: %s", err.Error())
				span.SetStatus(codes.Error, "创建对象类失败")
				return "", rest.NewHTTPError(ctx, http.StatusInternalServerError,
					berrors.BknBackend_ConceptGroup_InternalError_CreateObjectTypesFailed).
					WithErrorDetails(err.Error())
			}

			//  导入部分：处理分组与本体对象的关系
			_, err = cgs.AddObjectTypesToConceptGroup(ctx, tx, conceptGroup.KNID, conceptGroup.Branch, conceptGroup.CGID, otIDs, mode, strictMode)
			if err != nil {
				logger.Errorf("AddObjectTypesToConceptGroup error: %s", err.Error())
				span.SetStatus(codes.Error, "创建概念分组与对象类的关系失败")
				return "", rest.NewHTTPError(ctx, http.StatusInternalServerError,
					berrors.BknBackend_ConceptGroup_InternalError_AddObjectTypesToConceptGroupFailed).
					WithErrorDetails(err.Error())
			}
		}

		if len(conceptGroup.RelationTypes) > 0 {
			_, err = cgs.rts.CreateRelationTypes(ctx, tx, conceptGroup.RelationTypes, mode, strictMode)
			if err != nil {
				logger.Errorf("CreateRelationTypes error: %s", err.Error())
				span.SetStatus(codes.Error, "创建关系类失败")
				return "", rest.NewHTTPError(ctx, http.StatusInternalServerError,
					berrors.BknBackend_ConceptGroup_InternalError_CreateRelationTypesFailed).
					WithErrorDetails(err.Error())
			}
		}

		if len(conceptGroup.ActionTypes) > 0 {
			_, err = cgs.ats.CreateActionTypes(ctx, tx, conceptGroup.ActionTypes, mode, strictMode)
			if err != nil {
				logger.Errorf("CreateActionTypes error: %s", err.Error())
				span.SetStatus(codes.Error, "创建概念分组动作类失败")
				return "", rest.NewHTTPError(ctx, http.StatusInternalServerError,
					berrors.BknBackend_ConceptGroup_InternalError_CreateActionTypesFailed).
					WithErrorDetails(err.Error())
			}
		}
	}

	// 处理更新情况
	if isUpdate {
		err = cgs.UpdateConceptGroup(ctx, tx, conceptGroup, strictMode)
		if err != nil {
			logger.Errorf("UpdateConceptGroup error: %s", err.Error())
			span.SetStatus(codes.Error, "修改概念分组失败")
			return "", rest.NewHTTPError(ctx, http.StatusInternalServerError,
				berrors.BknBackend_ConceptGroup_InternalError_UpdateConceptGroupFailed).
				WithErrorDetails(err.Error())
		}

		if len(conceptGroup.ObjectTypes) > 0 {
			// 写入对象类
			_, err = cgs.ots.CreateObjectTypes(ctx, tx, conceptGroup.ObjectTypes, mode, false, strictMode)
			if err != nil {
				logger.Errorf("CreateObjectTypes error: %s", err.Error())
				span.SetStatus(codes.Error, "创建对象类失败")
				return "", rest.NewHTTPError(ctx, http.StatusInternalServerError,
					berrors.BknBackend_ConceptGroup_InternalError_CreateObjectTypesFailed).
					WithErrorDetails(err.Error())
			}
			//  导入部分：处理分组与本体对象的关系,只创建本分组与当前对象类的关系
			//  更新分组话，需要做个全量同步
			_, err = cgs.AddObjectTypesToConceptGroup(ctx, tx, conceptGroup.KNID, conceptGroup.Branch, conceptGroup.CGID, otIDs, mode, strictMode)
			if err != nil {
				logger.Errorf("AddObjectTypesToConceptGroup error: %s", err.Error())
				span.SetStatus(codes.Error, "创建概念分组与对象类的关系失败")
				return "", rest.NewHTTPError(ctx, http.StatusInternalServerError,
					berrors.BknBackend_ConceptGroup_InternalError_AddObjectTypesToConceptGroupFailed).
					WithErrorDetails(err.Error())
			}
		}

		if len(conceptGroup.RelationTypes) > 0 {
			_, err = cgs.rts.CreateRelationTypes(ctx, tx, conceptGroup.RelationTypes, mode, strictMode)
			if err != nil {
				logger.Errorf("CreateRelationTypes error: %s", err.Error())
				span.SetStatus(codes.Error, "创建关系类失败")
				return "", rest.NewHTTPError(ctx, http.StatusInternalServerError,
					berrors.BknBackend_ConceptGroup_InternalError_CreateRelationTypesFailed).
					WithErrorDetails(err.Error())
			}
		}

		if len(conceptGroup.ActionTypes) > 0 {
			_, err = cgs.ats.CreateActionTypes(ctx, tx, conceptGroup.ActionTypes, mode, strictMode)
			if err != nil {
				logger.Errorf("CreateActionTypes error: %s", err.Error())
				span.SetStatus(codes.Error, "创建动作类失败")
				return "", rest.NewHTTPError(ctx, http.StatusInternalServerError,
					berrors.BknBackend_ConceptGroup_InternalError_CreateActionTypesFailed).
					WithErrorDetails(err.Error())
			}
		}
	}

	if isCreate || isUpdate {
		err = cgs.InsertDatasetData(ctx, conceptGroup)
		if err != nil {
			logger.Errorf("InsertDatasetData error: %s", err.Error())
			span.SetStatus(codes.Error, "概念分组概念索引写入失败")

			return "", rest.NewHTTPError(ctx, http.StatusInternalServerError,
				berrors.BknBackend_ConceptGroup_InternalError_InsertOpenSearchDataFailed).
				WithErrorDetails(err.Error())
		}
	}

	span.SetStatus(codes.Ok, "")
	return conceptGroup.CGID, nil
}

// ValidateConceptGroups checks concept-group dependency existence only; does not write to the database.
func (cgs *conceptGroupService) ValidateConceptGroups(ctx context.Context, knID string, branch string,
	conceptGroups []*interfaces.ConceptGroup, strictMode bool, parentBatch *interfaces.BatchIDIndex, mode string) error {

	ctx, span := ar_trace.Tracer.Start(ctx, "ValidateConceptGroups")
	defer span.End()

	if len(conceptGroups) == 0 {
		span.SetStatus(codes.Ok, "")
		return nil
	}

	err := cgs.ps.CheckPermission(ctx, interfaces.PermissionResource{
		Type: interfaces.RESOURCE_TYPE_KN,
		ID:   knID,
	}, []string{interfaces.OPERATION_TYPE_MODIFY})
	if err != nil {
		return err
	}

	effectiveBatch := parentBatch
	if effectiveBatch == nil {
		effectiveBatch, err = batchindex.CollectFromConceptGroups(knID, branch, conceptGroups)
		if err != nil {
			return rest.NewHTTPError(ctx, http.StatusBadRequest,
				berrors.BknBackend_ConceptGroup_InvalidParameter).
				WithErrorDetails(err.Error())
		}
	}

	for _, cg := range conceptGroups {
		cg.KNID = knID
		cg.Branch = branch
		if _, _, err := cgs.handleConceptGroupImportMode(ctx, mode, cg); err != nil {
			return err
		}
		if strictMode {
			// 与 CreateConceptGroup 落库路径对齐：嵌套概念同步做依赖预检
			if len(cg.ObjectTypes) > 0 {
				// 与 CreateObjectTypes(strict) 一致：数据视图/逻辑属性/绑定概念分组等，而非仅 ID 是否在库
				if err := cgs.ots.ValidateObjectTypes(ctx, knID, branch, cg.ObjectTypes, strictMode, effectiveBatch, mode); err != nil {
					return err
				}
			}
			if len(cg.RelationTypes) > 0 {
				if err := cgs.rts.ValidateRelationTypes(ctx, knID, branch, cg.RelationTypes, strictMode, effectiveBatch, mode); err != nil {
					return err
				}
			}
			if len(cg.ActionTypes) > 0 {
				if err := cgs.ats.ValidateActionTypes(ctx, knID, branch, cg.ActionTypes, strictMode, effectiveBatch, mode); err != nil {
					return err
				}
			}
		}
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

func (cgs *conceptGroupService) ListConceptGroups(ctx context.Context,
	query interfaces.ConceptGroupsQueryParams) ([]*interfaces.ConceptGroup, int, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "查询概念分组列表")
	defer span.End()

	// 判断userid是否有查看业务知识网络的权限
	err := cgs.ps.CheckPermission(ctx, interfaces.PermissionResource{
		Type: interfaces.RESOURCE_TYPE_KN,
		ID:   query.KNID,
	}, []string{interfaces.OPERATION_TYPE_VIEW_DETAIL})
	if err != nil {
		return []*interfaces.ConceptGroup{}, 0, err
	}

	//获取概念分组列表
	conceptGroups, err := cgs.cga.ListConceptGroups(ctx, query)
	if err != nil {
		logger.Errorf("ListConceptGroups error: %s", err.Error())
		span.SetStatus(codes.Error, "List concept groups error")

		return []*interfaces.ConceptGroup{}, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_ConceptGroup_InternalError).WithErrorDetails(err.Error())
	}
	if len(conceptGroups) == 0 {
		span.SetStatus(codes.Ok, "")
		return []*interfaces.ConceptGroup{}, 0, nil
	}

	total := len(conceptGroups)

	// limit = -1,则返回所有
	if query.Limit != -1 {
		// 分页
		// 检查起始位置是否越界
		if query.Offset < 0 || query.Offset >= len(conceptGroups) {
			span.SetStatus(codes.Ok, "")
			return []*interfaces.ConceptGroup{}, total, nil
		}
		// 计算结束位置
		end := query.Offset + query.Limit
		if end > len(conceptGroups) {
			end = len(conceptGroups)
		}
		conceptGroups = conceptGroups[query.Offset:end]
	}

	accountInfos := make([]*interfaces.AccountInfo, 0, len(conceptGroups)*2)
	for _, cg := range conceptGroups {
		accountInfos = append(accountInfos, &cg.Creator, &cg.Updater)
	}

	err = cgs.ums.GetAccountNames(ctx, accountInfos)
	if err != nil {
		span.SetStatus(codes.Error, "GetAccountNames error")

		return []*interfaces.ConceptGroup{}, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_ConceptGroup_InternalError).WithErrorDetails(err.Error())
	}

	// 分组列表为每个组生成本体对象统计信息
	for _, conceptGroup := range conceptGroups {
		stats, err := cgs.GetStatByConceptGroup(ctx, conceptGroup)
		if err != nil {
			return []*interfaces.ConceptGroup{}, 0, err
		}
		conceptGroup.Statistics = stats

		otIDs, err := cgs.cga.GetConceptIDsByConceptGroupIDs(ctx, conceptGroup.KNID,
			conceptGroup.Branch, []string{conceptGroup.CGID}, interfaces.MODULE_TYPE_OBJECT_TYPE)
		if err != nil {
			errStr := fmt.Sprintf("GetConceptIDsByConceptGroupIDs failed, kn_id:[%s],branch:[%s],cg_ids:[%s], error: %v",
				conceptGroup.KNID, conceptGroup.Branch, conceptGroup.CGID, err)
			logger.Errorf(errStr)
			span.SetStatus(codes.Error, errStr)

			return []*interfaces.ConceptGroup{}, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				berrors.BknBackend_ConceptGroup_InternalError).WithErrorDetails(err.Error())
		}
		conceptGroup.ObjectTypeIDs = otIDs
	}

	span.SetStatus(codes.Ok, "")
	return conceptGroups, total, nil
}

func (cgs *conceptGroupService) GetConceptGroupByID(ctx context.Context, knID string, branch string,
	cgID string, mode string) (*interfaces.ConceptGroup, error) {

	// 获取概念分组
	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("查询概念分组[%s]信息", knID))
	defer span.End()

	// 判断userid是否有查看业务知识网络的权限
	err := cgs.ps.CheckPermission(ctx, interfaces.PermissionResource{
		Type: interfaces.RESOURCE_TYPE_KN,
		ID:   knID,
	}, []string{interfaces.OPERATION_TYPE_VIEW_DETAIL})
	if err != nil {
		return &interfaces.ConceptGroup{}, err
	}

	// 获取模型基本信息
	conceptGroup, err := cgs.cga.GetConceptGroupByID(ctx, knID, branch, cgID)
	if err != nil {
		logger.Errorf("GetConceptGroupByID error: %s", err.Error())
		span.SetStatus(codes.Error, fmt.Sprintf("Get concept group[%s] error: %v", knID, err))
		span.End()

		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_ConceptGroup_InternalError_GetConceptGroupByIDFailed).WithErrorDetails(err.Error())
	}

	if conceptGroup == nil {
		errStr := fmt.Sprintf("Concept group[%s] not found in knowledge network [%s] branch [%s]", cgID, knID, branch)
		logger.Errorf(errStr)
		span.SetStatus(codes.Error, errStr)
		span.End()

		return nil, rest.NewHTTPError(ctx, http.StatusNotFound, berrors.BknBackend_ConceptGroup_ConceptGroupNotFound).
			WithErrorDetails(errStr)
	}

	otIDs, err := cgs.cga.GetConceptIDsByConceptGroupIDs(ctx, conceptGroup.KNID,
		conceptGroup.Branch, []string{conceptGroup.CGID}, interfaces.MODULE_TYPE_OBJECT_TYPE)
	if err != nil {
		errStr := fmt.Sprintf("GetConceptIDsByConceptGroupIDs failed, kn_id:[%s],branch:[%s],cg_ids:[%s], error: %v",
			conceptGroup.KNID, conceptGroup.Branch, conceptGroup.CGID, err)
		logger.Errorf(errStr)
		span.SetStatus(codes.Error, errStr)
		span.End()

		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_ConceptGroup_InternalError_GetConceptIDsByConceptGroupIDsFailed).WithErrorDetails(err.Error())
	}

	// 对象类不为空时才找对应的关系类
	if len(otIDs) > 0 {
		objectTypes, _, err := cgs.ots.ListObjectTypes(ctx, nil, interfaces.ObjectTypesQueryParams{
			PaginationQueryParameters: interfaces.PaginationQueryParameters{
				Limit: -1, // 等于-1，把数据库中查询到的都返回
			},
			KNID:   conceptGroup.KNID,
			Branch: conceptGroup.Branch,
			OTIDS:  otIDs,
		})
		if err != nil {
			return nil, err
		}
		conceptGroup.ObjectTypes = objectTypes

		relationTypes, _, err := cgs.rts.ListRelationTypes(ctx, interfaces.RelationTypesQueryParams{
			PaginationQueryParameters: interfaces.PaginationQueryParameters{
				Limit: -1,
			},
			KNID:                conceptGroup.KNID,
			Branch:              conceptGroup.Branch,
			SourceObjectTypeIDs: otIDs,
			TargetObjectTypeIDs: otIDs,
		})
		if err != nil {
			return nil, err
		}
		conceptGroup.RelationTypes = relationTypes

		actionTypes, _, err := cgs.ats.ListActionTypes(ctx, interfaces.ActionTypesQueryParams{
			PaginationQueryParameters: interfaces.PaginationQueryParameters{
				Limit: -1,
			},
			KNID:          conceptGroup.KNID,
			Branch:        conceptGroup.Branch,
			ObjectTypeIDs: otIDs,
		})
		if err != nil {
			return nil, err
		}
		conceptGroup.ActionTypes = actionTypes
	}

	span.SetStatus(codes.Ok, "")
	return conceptGroup, nil
}

func (cgs *conceptGroupService) GetConceptGroupIDsByKnID(ctx context.Context, knID string, branch string) ([]string, error) {
	// 获取概念分组
	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("查询概念分组[%s]信息", knID))
	defer span.End()

	// 获取模型基本信息
	cgIDs, err := cgs.cga.GetConceptGroupIDsByKnID(ctx, knID, branch)
	if err != nil {
		logger.Errorf("GetConceptGroupIDsByKnID error: %s", err.Error())
		span.SetStatus(codes.Error, fmt.Sprintf("Get concept group ids by kn_id[%s] error: %v", knID, err))
		span.End()

		return []string{}, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_ConceptGroup_InternalError).WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return cgIDs, nil
}

// 获取概念分组的统计信息
func (cgs *conceptGroupService) GetStatByConceptGroup(ctx context.Context, conceptGroup *interfaces.ConceptGroup) (*interfaces.Statistics, error) {
	// 获取概念分组
	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("查询概念分组[%s]信息", conceptGroup.KNID))
	defer span.End()

	//  数量从对象类、概念对象关系、概念分组表中联合查询得到
	// 获取概念分组下的对象类、关系类、行动类的数量

	otIDs, err := cgs.cga.GetConceptIDsByConceptGroupIDs(ctx, conceptGroup.KNID,
		conceptGroup.Branch, []string{conceptGroup.CGID}, interfaces.MODULE_TYPE_OBJECT_TYPE)
	if err != nil {
		errStr := fmt.Sprintf("GetConceptIDsByConceptGroupIDs failed, kn_id:[%s],branch:[%s],cg_ids:[%s], error: %v",
			conceptGroup.KNID, conceptGroup.Branch, conceptGroup.CGID, err)
		logger.Errorf(errStr)
		span.SetStatus(codes.Error, errStr)
		span.End()

		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_ConceptGroup_InternalError_GetConceptIDsByConceptGroupIDsFailed).WithErrorDetails(err.Error())
	}

	if len(otIDs) == 0 {
		return &interfaces.Statistics{
			OtTotal: 0,
			RtTotal: 0,
			AtTotal: 0,
		}, nil
	}

	// 关系类数量
	rtCnt, err := cgs.rta.GetRelationTypesTotal(ctx, interfaces.RelationTypesQueryParams{
		KNID:                conceptGroup.KNID,
		Branch:              conceptGroup.Branch,
		SourceObjectTypeIDs: otIDs,
		TargetObjectTypeIDs: otIDs,
	})
	if err != nil {
		logger.Errorf("GetRelationTypesTotal in concept group[%s] error: %s", conceptGroup.KNID, err.Error())
		span.SetStatus(codes.Error, fmt.Sprintf("GetRelationTypesTotal in concept group[%s], error: %v", conceptGroup.KNID, err))
		span.End()

		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_ConceptGroup_InternalError_GetRelationTypesTotalFailed).WithErrorDetails(err.Error())
	}

	// 行动类数量
	atCnt, err := cgs.ata.GetActionTypesTotal(ctx, interfaces.ActionTypesQueryParams{
		KNID:          conceptGroup.KNID,
		Branch:        conceptGroup.Branch,
		ObjectTypeIDs: otIDs,
	})
	if err != nil {
		logger.Errorf("GetActionTypesTotal in concept group[%s] error: %s", conceptGroup.KNID, err.Error())
		span.SetStatus(codes.Error, fmt.Sprintf("GetActionTypesTotal in concept group[%s], error: %v", conceptGroup.KNID, err))
		span.End()

		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_ConceptGroup_InternalError_GetRelationTypesTotalFailed).WithErrorDetails(err.Error())
	}

	statistics := &interfaces.Statistics{
		OtTotal: len(otIDs),
		RtTotal: rtCnt,
		AtTotal: atCnt,
	}

	span.SetStatus(codes.Ok, "")
	return statistics, nil
}

// 更新概念分组
func (cgs *conceptGroupService) UpdateConceptGroup(ctx context.Context, tx *sql.Tx, conceptGroup *interfaces.ConceptGroup, strictMode bool) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Update concept group")
	defer span.End()

	// 判断userid是否有创建概念分组的权限（策略决策）
	err := cgs.ps.CheckPermission(ctx, interfaces.PermissionResource{
		Type: interfaces.RESOURCE_TYPE_KN,
		ID:   conceptGroup.KNID,
	}, []string{interfaces.OPERATION_TYPE_MODIFY})
	if err != nil {
		return err
	}

	if strictMode {
		if err := cgs.ValidateConceptGroups(ctx, conceptGroup.KNID, conceptGroup.Branch,
			[]*interfaces.ConceptGroup{conceptGroup}, strictMode, nil, interfaces.ImportMode_Overwrite); err != nil {
			return err
		}
	}

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}
	conceptGroup.Updater = accountInfo

	currentTime := time.Now().UnixMilli() // 概念分组的update_time是int类型
	conceptGroup.UpdateTime = currentTime

	otIDs, err := cgs.cga.GetConceptIDsByConceptGroupIDs(ctx, conceptGroup.KNID,
		conceptGroup.Branch, []string{conceptGroup.CGID}, interfaces.MODULE_TYPE_OBJECT_TYPE)
	if err != nil {
		errStr := fmt.Sprintf("GetConceptIDsByConceptGroupIDs failed, kn_id:[%s],branch:[%s],cg_ids:[%s], error: %v",
			conceptGroup.KNID, conceptGroup.Branch, conceptGroup.CGID, err)
		logger.Errorf(errStr)
		span.SetStatus(codes.Error, errStr)

		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_ConceptGroup_InternalError_GetConceptIDsByConceptGroupIDsFailed).
			WithErrorDetails(err.Error())
	}

	// 对象类不为空时才找对应的关系类
	if len(otIDs) > 0 {
		objectTypes, _, err := cgs.ots.ListObjectTypes(ctx, nil, interfaces.ObjectTypesQueryParams{
			PaginationQueryParameters: interfaces.PaginationQueryParameters{
				Limit: -1, // 等于-1，把数据库中查询到的都返回
			},
			KNID:   conceptGroup.KNID,
			Branch: conceptGroup.Branch,
			OTIDS:  otIDs,
		})
		if err != nil {
			return err
		}
		conceptGroup.ObjectTypes = objectTypes
	}

	bknOtMaps := map[string]*bknsdk.BknObjectType{}
	for _, objectType := range conceptGroup.ObjectTypes {
		bknOtMaps[objectType.OTID] = logics.ToBKNObjectType(objectType)
	}

	bknCG := logics.ToBKNConceptGroup(conceptGroup)
	conceptGroup.BKNRawContent = bknsdk.SerializeConceptGroup(bknCG, bknOtMaps)

	if tx == nil {
		// 0. 开始事务
		tx, err = cgs.db.Begin()
		if err != nil {
			logger.Errorf("Begin transaction error: %s", err.Error())
			span.SetStatus(codes.Error, "事务开启失败")
			o11y.Error(ctx, fmt.Sprintf("Begin transaction error: %s", err.Error()))

			return rest.NewHTTPError(ctx, http.StatusInternalServerError,
				berrors.BknBackend_ConceptGroup_InternalError_BeginTransactionFailed).
				WithErrorDetails(err.Error())
		}
		// 0.1 异常时
		defer func() {
			switch err {
			case nil:
				// 提交事务
				err = tx.Commit()
				if err != nil {
					logger.Errorf("UpdateConceptGroup Transaction Commit Failed:%v", err)
					span.SetStatus(codes.Error, "提交事务失败")
					o11y.Error(ctx, fmt.Sprintf("UpdateConceptGroup Transaction Commit Failed: %s", err.Error()))
				}
				logger.Infof("UpdateConceptGroup Transaction Commit Success:%v", conceptGroup.CGName)
				o11y.Debug(ctx, fmt.Sprintf("UpdateConceptGroup Transaction Commit Success: %s", conceptGroup.CGName))
			default:
				rollbackErr := tx.Rollback()
				if rollbackErr != nil {
					logger.Errorf("UpdateConceptGroup Transaction Rollback Error:%v", rollbackErr)
					span.SetStatus(codes.Error, "事务回滚失败")
					o11y.Error(ctx, fmt.Sprintf("UpdateConceptGroup Transaction Rollback Error: %s", rollbackErr.Error()))
				}
			}
		}()
	}

	// 更新模型信息
	err = cgs.cga.UpdateConceptGroup(ctx, tx, conceptGroup)
	if err != nil {
		logger.Errorf("UpdateConceptGroup error: %s", err.Error())
		span.SetStatus(codes.Error, "修改概念分组失败")

		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_ConceptGroup_InternalError).
			WithErrorDetails(err.Error())
	}

	err = cgs.InsertDatasetData(ctx, conceptGroup)
	if err != nil {
		logger.Errorf("InsertDatasetData error: %s", err.Error())
		span.SetStatus(codes.Error, "概念分组概念索引写入失败")

		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_ConceptGroup_InternalError_InsertOpenSearchDataFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

func (cgs *conceptGroupService) DeleteConceptGroupByID(ctx context.Context, tx *sql.Tx, knID string, branch string, cgID string) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Delete concept group by id")
	defer span.End()

	// 判断userid是否有删除概念分组的权限
	err := cgs.ps.CheckPermission(ctx, interfaces.PermissionResource{
		Type: interfaces.RESOURCE_TYPE_KN,
		ID:   knID,
	}, []string{interfaces.OPERATION_TYPE_MODIFY})
	if err != nil {
		return err
	}

	if tx == nil {
		// 0. 开始事务
		tx, err = cgs.db.Begin()
		if err != nil {
			logger.Errorf("Begin transaction error: %s", err.Error())
			span.SetStatus(codes.Error, "事务开启失败")
			o11y.Error(ctx, fmt.Sprintf("Begin transaction error: %s", err.Error()))

			return rest.NewHTTPError(ctx, http.StatusInternalServerError,
				berrors.BknBackend_ConceptGroup_InternalError_BeginTransactionFailed).
				WithErrorDetails(err.Error())
		}
	}

	// 0.1 异常时
	defer func() {
		switch err {
		case nil:
			// 提交事务
			err = tx.Commit()
			if err != nil {
				logger.Errorf("DeleteConceptGroup Transaction Commit Failed:%v", err)
				span.SetStatus(codes.Error, "提交事务失败")
				o11y.Error(ctx, fmt.Sprintf("DeleteConceptGroup Transaction Commit Failed: %s", err.Error()))
				return
			}
			logger.Infof("DeleteConceptGroup Transaction Commit Success")
			o11y.Debug(ctx, "DeleteConceptGroup Transaction Commit Success")
		default:
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				logger.Errorf("DeleteConceptGroup Transaction Rollback Error:%v", rollbackErr)
				span.SetStatus(codes.Error, "事务回滚失败")
				o11y.Error(ctx, fmt.Sprintf("DeleteConceptGroup Transaction Rollback Error: %s", err.Error()))
			}
		}
	}()

	// 删除概念分组
	rowsAffect, err := cgs.cga.DeleteConceptGroupByID(ctx, tx, knID, branch, cgID)
	if err != nil {
		logger.Errorf("DeleteConceptGroupsByIDs error: %s", err.Error())
		span.SetStatus(codes.Error, "删除概念分组失败")

		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_ConceptGroup_InternalError).WithErrorDetails(err.Error())
	}
	logger.Infof("DeleteConceptGroupByID: Rows affected is %v, request delete CGID is %s in knowledge network [%s] branch [%s]!",
		rowsAffect, cgID, knID, branch)
	if rowsAffect != int64(1) {
		logger.Warnf("DeleteConceptGroupByID number %v not equal %v!", rowsAffect, 1)
		o11y.Warn(ctx, fmt.Sprintf("DeleteConceptGroupByID number %v not equal %v!", rowsAffect, 1))
	}

	// 删除组下所有的绑定关系
	cgrsRowsAffect, err := cgs.cga.DeleteObjectTypesFromGroup(ctx, tx, interfaces.ConceptGroupRelationsQueryParams{
		KNID:        knID,
		Branch:      branch,
		ConceptType: interfaces.MODULE_TYPE_OBJECT_TYPE,
		CGIDs:       []string{cgID},
	})
	if err != nil {
		logger.Errorf("DeleteObjectTypesFromGroup error: %s", err.Error())
		span.SetStatus(codes.Error, "删除概念与分组的关系失败")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_ConceptGroup_InternalError).WithErrorDetails(err.Error())
	}
	logger.Infof("DeleteObjectTypesFromGroup: Rows affected is %v, request delete cgID is %s!", cgrsRowsAffect, cgID)

	docid := interfaces.GenerateConceptDocuemtnID(knID,
		interfaces.MODULE_TYPE_CONCEPT_GROUP, cgID, branch)
	err = cgs.vba.DeleteDatasetDocumentByID(ctx, interfaces.BKN_DATASET_ID, docid)
	if err != nil {
		logger.Errorf("DeleteDatasetDocumentByID error: %s", err.Error())
		span.SetStatus(codes.Error, "删除概念分组概念索引失败")
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 内部方法，删除概念分组，不检查权限，tx必须传入
func (cgs *conceptGroupService) DeleteConceptGroupsByKnID(ctx context.Context, tx *sql.Tx, knID string, branch string) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Delete concept group by knID")
	defer span.End()

	if tx == nil {
		logger.Errorf("missing transaction")
		o11y.Error(ctx, "missing transaction")
		span.SetStatus(codes.Error, "缺少事务")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_ConceptGroup_InternalError_MissingTransaction).
			WithErrorDetails("missing transaction")
	}

	// 删除概念分组
	rowsAffect, err := cgs.cga.DeleteConceptGroupsByKnID(ctx, tx, knID, branch)
	if err != nil {
		logger.Errorf("DeleteConceptGroupsByKnID error: %s", err.Error())
		span.SetStatus(codes.Error, "删除概念分组失败")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_ConceptGroup_InternalError).WithErrorDetails(err.Error())
	}
	logger.Infof("DeleteConceptGroupsByKnID: Rows affected is %v, request delete knID is %s in knowledge network [%s] branch [%s]!",
		rowsAffect, knID, knID, branch)

	// 删除组下所有的绑定关系
	rowsAffect, err = cgs.cga.DeleteConceptGroupRelationsByKnID(ctx, tx, knID, branch)
	if err != nil {
		logger.Errorf("DeleteConceptGroupRelationsByKnID error: %s", err.Error())
		span.SetStatus(codes.Error, "删除概念与分组的关系失败")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_ConceptGroup_InternalError).WithErrorDetails(err.Error())
	}
	logger.Infof("DeleteConceptGroupRelationsByKnID: Rows affected is %v, request delete knID is %s in knowledge network [%s] branch [%s]!",
		rowsAffect, knID, knID, branch)

	span.SetStatus(codes.Ok, "")
	return nil
}

// 更新知识网络详情
func (cgs *conceptGroupService) UpdateConceptGroupDetail(ctx context.Context, knID string, branch string, cgID string, detail string) error {
	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("Update concept group detail[%s]", knID))
	defer span.End()

	// 更新知识网络详情
	err := cgs.cga.UpdateConceptGroupDetail(ctx, knID, branch, cgID, detail)
	if err != nil {
		logger.Errorf("UpdateConceptGroupDetail error: %s", err.Error())
		span.SetStatus(codes.Error, "修改知识网络详情失败")

		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_ConceptGroup_InternalError).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

func (cgs *conceptGroupService) handleConceptGroupImportMode(ctx context.Context, mode string,
	conceptGroup *interfaces.ConceptGroup) (isCreate bool, isUpdate bool, err error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "concept group import mode logic")
	defer span.End()

	isCreate = false
	isUpdate = false

	// 校验单个ConceptGroup的导入模式逻辑
	idExist := false
	_, idExist, err = cgs.CheckConceptGroupExistByID(ctx, conceptGroup.KNID, conceptGroup.Branch, conceptGroup.CGID)
	if err != nil {
		return false, false, err
	}

	// 校验请求体与现有模型名称的重复性
	existID, nameExist, err := cgs.CheckConceptGroupExistByName(ctx, conceptGroup.KNID, conceptGroup.Branch, conceptGroup.CGName)
	if err != nil {
		return false, false, err
	}

	// 根据mode来区别，若是ignore，就从结果集中忽略，若是overwrite，就调用update，若是normal就报错。
	if idExist || nameExist {
		switch mode {
		case interfaces.ImportMode_Normal:
			if idExist {
				errDetails := fmt.Sprintf("The concept group with id [%s] already exists in knowledge network [%s] branch [%s]!",
					conceptGroup.CGID, conceptGroup.KNID, conceptGroup.Branch)
				logger.Error(errDetails)
				span.SetStatus(codes.Error, errDetails)
				return false, false, rest.NewHTTPError(ctx, http.StatusBadRequest,
					berrors.BknBackend_ConceptGroup_ConceptGroupIDExisted).
					WithErrorDetails(errDetails)
			}

			if nameExist {
				errDetails := fmt.Sprintf("concept group name '%s' already exists in knowledge network [%s] branch [%s]",
					conceptGroup.CGName, conceptGroup.KNID, conceptGroup.Branch)
				logger.Error(errDetails)
				span.SetStatus(codes.Error, errDetails)
				return false, false, rest.NewHTTPError(ctx, http.StatusForbidden,
					berrors.BknBackend_ConceptGroup_ConceptGroupNameExisted).
					WithDescription(map[string]any{"cg_name": conceptGroup.CGName}).
					WithErrorDetails(errDetails)
			}

		case interfaces.ImportMode_Ignore:
			// 存在重复的就跳过，不创建也不更新
			return false, false, nil
		case interfaces.ImportMode_Overwrite:
			if idExist && nameExist {
				// 如果 id 和名称都存在，但是存在的名称对应的视图 id 和当前视图 id 不一样，则报错
				if existID != conceptGroup.CGID {
					errDetails := fmt.Sprintf("Concept group ID '%s' and name '%s' already exist in knowledge network [%s] branch [%s], but the exist concept group id is '%s'",
						conceptGroup.CGID, conceptGroup.CGName, conceptGroup.KNID, conceptGroup.Branch, existID)
					logger.Error(errDetails)
					span.SetStatus(codes.Error, errDetails)
					return false, false, rest.NewHTTPError(ctx, http.StatusForbidden,
						berrors.BknBackend_ConceptGroup_ConceptGroupNameExisted).
						WithErrorDetails(errDetails)
				} else {
					// 如果 id 和名称、度量名称都存在，存在的名称对应的模型 id 和当前模型 id 一样，则覆盖更新
					isUpdate = true
					return isCreate, isUpdate, nil
				}
			}

			// id 已存在，且名称不存在，覆盖更新
			if idExist && !nameExist {
				isUpdate = true
				return isCreate, isUpdate, nil
			}

			// 如果 id 不存在，name 存在，报错
			if !idExist && nameExist {
				errDetails := fmt.Sprintf("Concept Group ID '%s' does not exist, but name '%s' already exists in knowledge network [%s] branch [%s]",
					conceptGroup.CGID, conceptGroup.CGName, conceptGroup.KNID, conceptGroup.Branch)
				logger.Error(errDetails)
				span.SetStatus(codes.Error, errDetails)
				return false, false, rest.NewHTTPError(ctx, http.StatusForbidden,
					berrors.BknBackend_ConceptGroup_ConceptGroupNameExisted).
					WithErrorDetails(errDetails)
			}

			// 如果 id 不存在，name不存在，度量名称不存在，不需要做什么，创建
			// if !idExist && !nameExist {}
		}
	}

	// 默认情况：需要创建
	isCreate = true
	return isCreate, isUpdate, nil
}

func (cgs *conceptGroupService) InsertDatasetData(ctx context.Context, origConceptGroup *interfaces.ConceptGroup) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "概念分组概念索引写入")
	defer span.End()

	conceptGroup := &interfaces.ConceptGroup{
		CGID:       origConceptGroup.CGID,
		CGName:     origConceptGroup.CGName,
		CommonInfo: origConceptGroup.CommonInfo,
		KNID:       origConceptGroup.KNID,
		Branch:     origConceptGroup.Branch,
		Creator:    origConceptGroup.Creator,
		CreateTime: origConceptGroup.CreateTime,
		Updater:    origConceptGroup.Updater,
		UpdateTime: origConceptGroup.UpdateTime,
		ModuleType: interfaces.MODULE_TYPE_CONCEPT_GROUP,
	}

	if cgs.appSetting.ServerSetting.DefaultSmallModelEnabled {
		words := []string{conceptGroup.CGName}
		words = append(words, conceptGroup.Tags...)
		words = append(words, conceptGroup.Comment, conceptGroup.BKNRawContent)
		word := strings.Join(words, "\n")

		defaultModel, err := cgs.mfa.GetDefaultModel(ctx)
		if err != nil {
			logger.Errorf("GetDefaultModel error: %s", err.Error())
			span.SetStatus(codes.Error, "获取默认模型失败")
			return err
		}
		vectors, err := cgs.mfa.GetVector(ctx, defaultModel, []string{word})
		if err != nil {
			logger.Errorf("GetVector error: %s", err.Error())
			span.SetStatus(codes.Error, "获取概念分组向量失败")
			return err
		}

		conceptGroup.Vector = vectors[0].Vector
	}

	docid := interfaces.GenerateConceptDocuemtnID(conceptGroup.KNID, interfaces.MODULE_TYPE_CONCEPT_GROUP, conceptGroup.CGID, conceptGroup.Branch)

	// Convert to map for dataset
	docBytes, err := sonic.Marshal(conceptGroup)
	if err != nil {
		logger.Errorf("Failed to marshal ConceptGroup: %s", err.Error())
		span.SetStatus(codes.Error, "序列化概念分组失败")
		return err
	}

	var doc map[string]any
	if err := sonic.Unmarshal(docBytes, &doc); err != nil {
		logger.Errorf("Failed to unmarshal ConceptGroup: %s", err.Error())
		span.SetStatus(codes.Error, "反序列化概念分组失败")
		return err
	}

	// Set document ID
	doc["_id"] = docid

	err = cgs.vba.WriteDatasetDocuments(ctx, interfaces.BKN_DATASET_ID, []map[string]any{doc})
	if err != nil {
		logger.Errorf("WriteDatasetDocuments error: %s", err.Error())
		span.SetStatus(codes.Error, "概念分组概念索引写入失败")
		return err
	}

	return nil
}

// 添加对象类到指定概念分组中
func (cgs *conceptGroupService) AddObjectTypesToConceptGroup(ctx context.Context, tx *sql.Tx, knID string, branch string,
	cgID string, otIDs []interfaces.ID, importMode string, strictMode bool) ([]string, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "添加对象类到概念分组中")
	defer span.End()

	var err error
	if tx == nil {
		// 0. 开始事务
		tx, err = cgs.db.Begin()
		if err != nil {
			logger.Errorf("Begin transaction error: %s", err.Error())
			span.SetStatus(codes.Error, "事务开启失败")
			o11y.Error(ctx, fmt.Sprintf("Begin transaction error: %s", err.Error()))

			return []string{}, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				berrors.BknBackend_ConceptGroup_InternalError_BeginTransactionFailed).
				WithErrorDetails(err.Error())
		}
		// 0.1 异常时
		defer func() {
			switch err {
			case nil:
				// 提交事务
				err = tx.Commit()
				if err != nil {
					logger.Errorf("AddObjectTypesToConceptGroup Transaction Commit Failed:%v", err)
					span.SetStatus(codes.Error, "提交事务失败")
					o11y.Error(ctx, fmt.Sprintf("AddObjectTypesToConceptGroup Transaction Commit Failed: %s", err.Error()))
				}
				logger.Infof("AddObjectTypesToConceptGroup Transaction Commit Success:kn_id:%s,branch:%s,cg_id:%s,ot_ids:%v", knID, branch, cgID, otIDs)
				o11y.Debug(ctx, fmt.Sprintf("AddObjectTypesToConceptGroup Transaction Commit Success:kn_id:%s,branch:%s,cg_id:%s,ot_ids:%v", knID, branch, cgID, otIDs))
			default:
				rollbackErr := tx.Rollback()
				if rollbackErr != nil {
					logger.Errorf("AddObjectTypesToConceptGroup Transaction Rollback Error:%v", rollbackErr)
					span.SetStatus(codes.Error, "事务回滚失败")
					o11y.Error(ctx, fmt.Sprintf("AddObjectTypesToConceptGroup Transaction Rollback Error: %s", rollbackErr.Error()))
				}
			}
		}()
	}

	// id去重后再查
	otIDArr := interfaces.GetUniqueIDs(otIDs)

	// 1. When strictMode is true, validate all object type IDs exist in the KN/branch
	if strictMode {
		objectTypes, _, err := cgs.ots.ListObjectTypes(ctx, tx, interfaces.ObjectTypesQueryParams{
			PaginationQueryParameters: interfaces.PaginationQueryParameters{
				Limit: -1,
			},
			KNID:   knID,
			Branch: branch,
			OTIDS:  otIDArr,
		})
		if err != nil {
			return nil, err
		}
		if len(objectTypes) != len(otIDArr) {
			errStr := fmt.Sprintf("Exists any object types not found, expect object types nums is [%d], actual object types num is [%d]", len(otIDs), len(objectTypes))
			logger.Errorf(errStr)
			span.SetStatus(codes.Error, errStr)

			return []string{}, rest.NewHTTPError(ctx, http.StatusNotFound,
				berrors.BknBackend_ConceptGroup_ObjectTypeNotFound).WithErrorDetails(errStr)
		}
	}

	currentTime := time.Now().UnixMilli()

	// 2. 校验对象类是否已经在分组中，若存在对象类已在分组中，报错
	cgRelations, err := cgs.cga.ListConceptGroupRelations(ctx, tx, interfaces.ConceptGroupRelationsQueryParams{
		PaginationQueryParameters: interfaces.PaginationQueryParameters{
			Limit: -1,
		},
		KNID:        knID,
		Branch:      branch,
		CGIDs:       []string{cgID},
		ConceptType: interfaces.MODULE_TYPE_OBJECT_TYPE,
		OTIDs:       otIDArr,
	})
	if err != nil {
		errStr := fmt.Sprintf("ListConceptGroupRelations failed, the concept group is [%s], knowledge network is [%s], branch is [%s], object types is [%v]",
			cgID, knID, branch, otIDArr)
		logger.Errorf(errStr)
		span.SetStatus(codes.Error, errStr)

		return []string{}, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_ConceptGroup_InternalError).
			WithErrorDetails(err.Error())
	}

	groupsToAdd := make([]string, 0)
	if len(cgRelations) > 0 {
		switch importMode {
		case interfaces.ImportMode_Normal:
			// normal 请求下，关系已存在，报错
			errStr := fmt.Sprintf("Exists some object types in the concept group [%s] knowledge network [%s] branch [%s], expect relations num is [%d], actual relations num is [%d]",
				cgID, knID, branch, len(otIDs), len(cgRelations))
			logger.Errorf(errStr)
			span.SetStatus(codes.Error, errStr)

			return []string{}, rest.NewHTTPError(ctx, http.StatusNotFound,
				berrors.BknBackend_ConceptGroup_ConceptGroupRelationExisted).WithErrorDetails(errStr)

		case interfaces.ImportMode_Ignore, interfaces.ImportMode_Overwrite:
			// ignore 和 override 下，忽略重复（冲突）的关系，添加新的关系
			// 2. 计算需要添加(不冲突)的分组
			existingGroupIDs := make(map[string]bool)

			// 已建立关系的对象类
			for _, rel := range cgRelations {
				existingGroupIDs[rel.ConceptID] = true
			}

			// 当前请求期望建立关系的对象类
			newGroupIDs := make(map[string]bool)
			for _, otID := range otIDArr {
				newGroupIDs[otID] = true
			}

			// 计算差异
			for groupID := range newGroupIDs {
				if !existingGroupIDs[groupID] {
					groupsToAdd = append(groupsToAdd, groupID)
				}
			}
		}
	} else {
		groupsToAdd = otIDArr
	}

	// 3. 组装对应关系，保存对应关系数据
	otCGIDs := []string{}
	for _, otID := range groupsToAdd {
		cgRelationID := xid.New().String()

		err = cgs.cga.CreateConceptGroupRelation(ctx, tx, &interfaces.ConceptGroupRelation{
			ID:          cgRelationID,
			KNID:        knID,
			Branch:      branch,
			CGID:        cgID,
			ConceptType: interfaces.MODULE_TYPE_OBJECT_TYPE,
			ConceptID:   otID,
			CreateTime:  currentTime,
		})
		if err != nil {
			errStr := fmt.Sprintf("CreateConceptGroupRelation failed, the concept group is [%s], knowledge network is [%s], branch is [%s], object type is [%s]",
				cgID, knID, branch, otID)
			logger.Errorf(errStr)
			span.SetStatus(codes.Error, errStr)

			return []string{}, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				berrors.BknBackend_ConceptGroup_InternalError_CreateConceptGroupRelationFailed).
				WithErrorDetails(err.Error())
		}
		otCGIDs = append(otCGIDs, cgRelationID)
	}

	return otCGIDs, nil
}

// 获取分组与对象类的关系
func (cgs *conceptGroupService) ListConceptGroupRelations(ctx context.Context,
	query interfaces.ConceptGroupRelationsQueryParams) ([]interfaces.ConceptGroupRelation, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "查询概念与分组的关系列表")
	defer span.End()

	// 判断userid是否有查看业务知识网络的权限
	err := cgs.ps.CheckPermission(ctx, interfaces.PermissionResource{
		Type: interfaces.RESOURCE_TYPE_KN,
		ID:   query.KNID,
	}, []string{interfaces.OPERATION_TYPE_VIEW_DETAIL})
	if err != nil {
		return []interfaces.ConceptGroupRelation{}, err
	}

	// 0. 开始事务
	tx, err := cgs.db.Begin()
	if err != nil {
		logger.Errorf("Begin transaction error: %s", err.Error())
		span.SetStatus(codes.Error, "事务开启失败")
		o11y.Error(ctx, fmt.Sprintf("Begin transaction error: %s", err.Error()))

		return []interfaces.ConceptGroupRelation{}, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_ConceptGroup_InternalError_BeginTransactionFailed).
			WithErrorDetails(err.Error())
	}
	// 0.1 异常时
	defer func() {
		switch err {
		case nil:
			// 提交事务
			err = tx.Commit()
			if err != nil {
				logger.Errorf("ListConceptGroupRelations Transaction Commit Failed:%v", err)
				span.SetStatus(codes.Error, "提交事务失败")
				o11y.Error(ctx, fmt.Sprintf("ListConceptGroupRelations Transaction Commit Failed: %s", err.Error()))
			}
			logger.Infof("ListConceptGroupRelations Transaction Commit Success:%v", query)
			o11y.Debug(ctx, fmt.Sprintf("ListConceptGroupRelations Transaction Commit Success: %v", query))
		default:
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				logger.Errorf("ListConceptGroupRelations Transaction Rollback Error:%v", rollbackErr)
				span.SetStatus(codes.Error, "事务回滚失败")
				o11y.Error(ctx, fmt.Sprintf("ListConceptGroupRelations Transaction Rollback Error: %s", rollbackErr.Error()))
			}
		}
	}()

	//获取概念分组列表
	cgrArr, err := cgs.cga.ListConceptGroupRelations(ctx, tx, query)
	if err != nil {
		logger.Errorf("ListConceptGroupRelations error: %s", err.Error())
		span.SetStatus(codes.Error, "List concept group relations error")

		return []interfaces.ConceptGroupRelation{}, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_ConceptGroup_InternalError).WithErrorDetails(err.Error())
	}
	if len(cgrArr) == 0 {
		span.SetStatus(codes.Ok, "")
		return []interfaces.ConceptGroupRelation{}, nil
	}

	// limit = -1,则返回所有
	if query.Limit == -1 {
		span.SetStatus(codes.Ok, "")
		return cgrArr, nil
	}
	// 分页
	// 检查起始位置是否越界
	if query.Offset < 0 || query.Offset >= len(cgrArr) {
		span.SetStatus(codes.Ok, "")
		return []interfaces.ConceptGroupRelation{}, nil
	}
	// 计算结束位置
	end := query.Offset + query.Limit
	if end > len(cgrArr) {
		end = len(cgrArr)
	}

	cgrArr = cgrArr[query.Offset:end]

	span.SetStatus(codes.Ok, "")
	return cgrArr, nil

}

// 从概念分组中移除对象类
func (cgs *conceptGroupService) DeleteObjectTypesFromGroup(ctx context.Context, tx *sql.Tx, knID string, branch string, cgID string, otIDs []string) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Delete concept group relations")
	defer span.End()

	// 判断userid是否有修改业务知识网络的权限
	err := cgs.ps.CheckPermission(ctx, interfaces.PermissionResource{
		Type: interfaces.RESOURCE_TYPE_KN,
		ID:   knID,
	}, []string{interfaces.OPERATION_TYPE_MODIFY})
	if err != nil {
		return err
	}

	if tx == nil {
		// 0. 开始事务
		tx, err = cgs.db.Begin()
		if err != nil {
			logger.Errorf("Begin transaction error: %s", err.Error())
			span.SetStatus(codes.Error, "事务开启失败")
			o11y.Error(ctx, fmt.Sprintf("Begin transaction error: %s", err.Error()))

			return rest.NewHTTPError(ctx, http.StatusInternalServerError,
				berrors.BknBackend_ConceptGroup_InternalError_BeginTransactionFailed).
				WithErrorDetails(err.Error())
		}
		// 0.1 异常时
		defer func() {
			switch err {
			case nil:
				// 提交事务
				err = tx.Commit()
				if err != nil {
					logger.Errorf("DeleteObjectTypesFromGroup Transaction Commit Failed:%v", err)
					span.SetStatus(codes.Error, "提交事务失败")
					o11y.Error(ctx, fmt.Sprintf("DeleteObjectTypesFromGroup Transaction Commit Failed: %s", err.Error()))
				}
				logger.Infof("DeleteObjectTypesFromGroup Transaction Commit Success: kn_id:%s,branch:%s,cg_id:%s,ot_ids:%v", knID, branch, cgID, otIDs)
				o11y.Debug(ctx, fmt.Sprintf("DeleteObjectTypesFromGroup Transaction Commit Success: kn_id:%s,branch:%s,cg_id:%s,ot_ids:%v", knID, branch, cgID, otIDs))
			default:
				rollbackErr := tx.Rollback()
				if rollbackErr != nil {
					logger.Errorf("DeleteObjectTypesFromGroup Transaction Rollback Error:%v", rollbackErr)
					span.SetStatus(codes.Error, "事务回滚失败")
					o11y.Error(ctx, fmt.Sprintf("DeleteObjectTypesFromGroup Transaction Rollback Error: %s", rollbackErr.Error()))
				}
			}
		}()
	}

	// 删除对象类与分组的绑定关系
	rowsAffect, err := cgs.cga.DeleteObjectTypesFromGroup(ctx, tx, interfaces.ConceptGroupRelationsQueryParams{
		KNID:        knID,
		Branch:      branch,
		CGIDs:       []string{cgID},
		ConceptType: interfaces.MODULE_TYPE_OBJECT_TYPE,
		OTIDs:       otIDs,
	})
	if err != nil {
		logger.Errorf("DeleteObjectTypesFromGroup error: %s", err.Error())
		span.SetStatus(codes.Error, "删除概念与分组的关系失败")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_ConceptGroup_InternalError).WithErrorDetails(err.Error())
	}

	logger.Infof("DeleteObjectTypesFromGroup: Rows affected is %v, request delete ATIDs is %v!", rowsAffect, len(otIDs))
	if rowsAffect != int64(len(otIDs)) {
		logger.Warnf("Delete action types number %v not equal requerst action types number %v!", rowsAffect, len(otIDs))
		o11y.Warn(ctx, fmt.Sprintf("Delete action types number %v not equal requerst action types number %v!", rowsAffect, len(otIDs)))
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

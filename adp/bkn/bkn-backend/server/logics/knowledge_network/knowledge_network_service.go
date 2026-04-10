// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package knowledge_network

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
	"bkn-backend/logics/business_system"
	"bkn-backend/logics/concept_group"
	"bkn-backend/logics/job"
	"bkn-backend/logics/object_type"
	"bkn-backend/logics/permission"
	"bkn-backend/logics/relation_type"
	"bkn-backend/logics/risk_type"
	"bkn-backend/logics/user_mgmt"
)

var (
	knServiceOnce sync.Once
	knService     interfaces.KNService
)

type knowledgeNetworkService struct {
	appSetting *common.AppSetting
	db         *sql.DB
	ata        interfaces.ActionTypeAccess
	ats        interfaces.ActionTypeService
	bss        interfaces.BusinessSystemService
	cga        interfaces.ConceptGroupAccess
	cgs        interfaces.ConceptGroupService
	js         interfaces.JobService
	kna        interfaces.KNAccess
	mfa        interfaces.ModelFactoryAccess
	ota        interfaces.ObjectTypeAccess
	ots        interfaces.ObjectTypeService
	rta        interfaces.RelationTypeAccess
	riskTypeA  interfaces.RiskTypeAccess
	riskTypeS  interfaces.RiskTypeService
	ps         interfaces.PermissionService
	rts        interfaces.RelationTypeService
	ums        interfaces.UserMgmtService
	vba        interfaces.VegaBackendAccess
}

func NewKNService(appSetting *common.AppSetting) interfaces.KNService {
	knServiceOnce.Do(func() {
		knService = &knowledgeNetworkService{
			appSetting: appSetting,
			ata:        logics.ATA,
			ats:        action_type.NewActionTypeService(appSetting),
			bss:        business_system.NewBusinessSystemService(appSetting),
			cga:        logics.CGA,
			cgs:        concept_group.NewConceptGroupService(appSetting),
			db:         logics.DB,
			js:         job.NewJobService(appSetting),
			kna:        logics.KNA,
			mfa:        logics.MFA,
			ota:        logics.OTA,
			ots:        object_type.NewObjectTypeService(appSetting),
			ps:         permission.NewPermissionService(appSetting),
			rta:        logics.RTA,
			riskTypeA:  logics.RiskTypeAccess,
			riskTypeS:  risk_type.NewRiskTypeService(appSetting),
			rts:        relation_type.NewRelationTypeService(appSetting),
			ums:        user_mgmt.NewUserMgmtService(appSetting),
			vba:        logics.VBA,
		}
	})
	return knService
}

func (kns *knowledgeNetworkService) CheckKNExistByID(ctx context.Context, KNID string, branch string) (string, bool, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("校验业务知识网络[%v]的存在性", KNID))
	defer span.End()

	otName, exist, err := kns.kna.CheckKNExistByID(ctx, KNID, branch)
	if err != nil {
		logger.Errorf("CheckKNExistByID error: %s", err.Error())

		span.SetStatus(codes.Error, fmt.Sprintf("按ID[%v]获取业务知识网络失败", KNID))
		// 记录处理的 sql 字符串
		o11y.Error(ctx, fmt.Sprintf("按ID[%v]获取业务知识网络失败: %v", KNID, err))

		return "", exist, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_KnowledgeNetwork_InternalError_CheckKNIfExistFailed).WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return otName, exist, nil
}

func (kns *knowledgeNetworkService) CheckKNExistByName(ctx context.Context, knName string, branch string) (string, bool, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("校验业务知识网络[%v]的存在性", knName))
	defer span.End()

	KNID, exist, err := kns.kna.CheckKNExistByName(ctx, knName, branch)
	if err != nil {
		logger.Errorf("CheckKNExistByName error: %s", err.Error())

		span.SetStatus(codes.Error, fmt.Sprintf("按名称[%v]获取业务知识网络失败", knName))
		// 记录处理的 sql 字符串
		o11y.Error(ctx, fmt.Sprintf("按名称[%v]获取业务知识网络失败: %v", knName, err))

		return KNID, exist, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_KnowledgeNetwork_InternalError_CheckKNIfExistFailed).WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return KNID, exist, nil
}

func (kns *knowledgeNetworkService) CreateKN(ctx context.Context, kn *interfaces.KN, mode string, strictMode bool) (string, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "Create knowledge network")
	defer span.End()

	// 判断userid是否有创建业务知识网络的权限（策略决策）
	err := kns.ps.CheckPermission(ctx, interfaces.PermissionResource{
		Type: interfaces.RESOURCE_TYPE_KN,
		ID:   interfaces.RESOURCE_ID_ALL,
	}, []string{interfaces.OPERATION_TYPE_CREATE})
	if err != nil {
		return "", err
	}

	currentTime := time.Now().UnixMilli()
	// 若提交的模型id为空，生成分布式ID
	if kn.KNID == "" {
		kn.KNID = xid.New().String()
	}
	for _, conceptGroup := range kn.ConceptGroups {
		conceptGroup.KNID = kn.KNID
	}
	for _, objectType := range kn.ObjectTypes {
		objectType.KNID = kn.KNID
	}
	for _, relationType := range kn.RelationTypes {
		relationType.KNID = kn.KNID
	}
	for _, actionType := range kn.ActionTypes {
		actionType.KNID = kn.KNID
	}
	for _, riskType := range kn.RiskTypes {
		riskType.KNID = kn.KNID
	}

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}
	kn.Creator = accountInfo
	kn.Updater = accountInfo

	kn.CreateTime = currentTime
	kn.UpdateTime = currentTime

	bknNetwork := logics.ToBKNNetWork(kn)
	kn.BKNRawContent = bknsdk.SerializeBknNetwork(bknNetwork)

	// 0. 开始事务
	tx, err := kns.db.Begin()
	if err != nil {
		logger.Errorf("Begin transaction error: %s", err.Error())
		span.SetStatus(codes.Error, "事务开启失败")
		o11y.Error(ctx, fmt.Sprintf("Begin transaction error: %s", err.Error()))

		return "", rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_KnowledgeNetwork_InternalError_BeginTransactionFailed).
			WithErrorDetails(err.Error())
	}

	// 0.1 异常时
	defer func() {
		switch err {
		case nil:
			// 提交事务
			err = tx.Commit()
			if err != nil {
				logger.Errorf("CreateKN Transaction Commit Failed:%v", err)
				span.SetStatus(codes.Error, "提交事务失败")
				o11y.Error(ctx, fmt.Sprintf("CreateKN Transaction Commit Failed: %s", err.Error()))
				return
			}
			logger.Infof("CreateKN Transaction Commit Success")
			o11y.Debug(ctx, "CreateKN Transaction Commit Success")
		default:
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				logger.Errorf("CreateKN Transaction Rollback Error:%v", rollbackErr)
				span.SetStatus(codes.Error, "事务回滚失败")
				o11y.Error(ctx, fmt.Sprintf("CreateKN Transaction Rollback Error: %s", err.Error()))
			}
		}
	}()

	// 处理导入模式
	isCreate, isUpdate, err := kns.handleKNImportMode(ctx, mode, kn)
	if err != nil {
		return "", err
	}

	// 处理创建情况
	if isCreate {
		err = kns.kna.CreateKN(ctx, tx, kn)
		if err != nil {
			logger.Errorf("CreateKN error: %s", err.Error())
			span.SetStatus(codes.Error, "创建业务知识网络失败")

			return "", rest.NewHTTPError(ctx, http.StatusInternalServerError,
				berrors.BknBackend_KnowledgeNetwork_InternalError_CreateKNFailed).
				WithErrorDetails(err.Error())
		}

		// 导入概念分组
		if len(kn.ConceptGroups) > 0 {
			for _, cg := range kn.ConceptGroups {
				_, err = kns.cgs.CreateConceptGroup(ctx, tx, cg, mode, strictMode)
				if err != nil {
					logger.Errorf("CreateObjectTypes error: %s", err.Error())
					span.SetStatus(codes.Error, "创建业务知识网络概念分组失败")
					return "", rest.NewHTTPError(ctx, http.StatusInternalServerError,
						berrors.BknBackend_KnowledgeNetwork_InternalError_CreateObjectTypesFailed).
						WithErrorDetails(err.Error())
				}
			}
		}

		if len(kn.ObjectTypes) > 0 {
			_, err = kns.ots.CreateObjectTypes(ctx, tx, kn.ObjectTypes, mode, true, strictMode)
			if err != nil {
				logger.Errorf("CreateObjectTypes error: %s", err.Error())
				span.SetStatus(codes.Error, "创建业务知识网络对象类失败")
				return "", rest.NewHTTPError(ctx, http.StatusInternalServerError,
					berrors.BknBackend_KnowledgeNetwork_InternalError_CreateObjectTypesFailed).
					WithErrorDetails(err.Error())
			}
		}

		if len(kn.RelationTypes) > 0 {
			_, err = kns.rts.CreateRelationTypes(ctx, tx, kn.RelationTypes, mode, strictMode)
			if err != nil {
				logger.Errorf("CreateRelationTypes error: %s", err.Error())
				span.SetStatus(codes.Error, "创建业务知识网络关系类失败")
				return "", rest.NewHTTPError(ctx, http.StatusInternalServerError,
					berrors.BknBackend_KnowledgeNetwork_InternalError_CreateRelationTypesFailed).
					WithErrorDetails(err.Error())
			}
		}

		if len(kn.ActionTypes) > 0 {
			_, err = kns.ats.CreateActionTypes(ctx, tx, kn.ActionTypes, mode, strictMode)
			if err != nil {
				logger.Errorf("CreateActionTypes error: %s", err.Error())
				span.SetStatus(codes.Error, "创建业务知识网络动作类失败")
				return "", rest.NewHTTPError(ctx, http.StatusInternalServerError,
					berrors.BknBackend_KnowledgeNetwork_InternalError_CreateActionTypesFailed).
					WithErrorDetails(err.Error())
			}
		}

		if len(kn.RiskTypes) > 0 {
			_, err = kns.riskTypeS.CreateRiskTypes(ctx, tx, kn.RiskTypes, mode)
			if err != nil {
				logger.Errorf("CreateRiskTypes error: %s", err.Error())
				span.SetStatus(codes.Error, "创建业务知识网络风险类失败")
				return "", rest.NewHTTPError(ctx, http.StatusInternalServerError,
					berrors.BknBackend_RiskType_InternalError).
					WithErrorDetails(err.Error())
			}
		}
	}

	// 处理更新情况
	if isUpdate {
		// todo: 提交的已存在，需要更新，则版本号+1
		err = kns.UpdateKN(ctx, tx, kn, strictMode)
		if err != nil {
			logger.Errorf("UpdateKN error: %s", err.Error())
			span.SetStatus(codes.Error, "修改业务知识网络失败")
			return "", rest.NewHTTPError(ctx, http.StatusInternalServerError,
				berrors.BknBackend_KnowledgeNetwork_InternalError_UpdateKNFailed).
				WithErrorDetails(err.Error())
		}

		if len(kn.ConceptGroups) > 0 {
			for _, cg := range kn.ConceptGroups {
				_, err = kns.cgs.CreateConceptGroup(ctx, tx, cg, mode, strictMode)
				if err != nil {
					logger.Errorf("CreateObjectTypes error: %s", err.Error())
					span.SetStatus(codes.Error, "创建业务知识网络概念分组失败")
					return "", rest.NewHTTPError(ctx, http.StatusInternalServerError,
						berrors.BknBackend_KnowledgeNetwork_InternalError_CreateObjectTypesFailed).
						WithErrorDetails(err.Error())
				}
			}
		}

		if len(kn.ObjectTypes) > 0 {
			_, err = kns.ots.CreateObjectTypes(ctx, tx, kn.ObjectTypes, mode, true, strictMode)
			if err != nil {
				logger.Errorf("CreateObjectTypes error: %s", err.Error())
				span.SetStatus(codes.Error, "创建业务知识网络对象类失败")
				return "", rest.NewHTTPError(ctx, http.StatusInternalServerError,
					berrors.BknBackend_KnowledgeNetwork_InternalError_CreateObjectTypesFailed).
					WithErrorDetails(err.Error())
			}
		}

		if len(kn.RelationTypes) > 0 {
			_, err = kns.rts.CreateRelationTypes(ctx, tx, kn.RelationTypes, mode, strictMode)
			if err != nil {
				logger.Errorf("CreateRelationTypes error: %s", err.Error())
				span.SetStatus(codes.Error, "创建业务知识网络关系类失败")
				return "", rest.NewHTTPError(ctx, http.StatusInternalServerError,
					berrors.BknBackend_KnowledgeNetwork_InternalError_CreateRelationTypesFailed).
					WithErrorDetails(err.Error())
			}
		}

		if len(kn.ActionTypes) > 0 {
			_, err = kns.ats.CreateActionTypes(ctx, tx, kn.ActionTypes, mode, strictMode)
			if err != nil {
				logger.Errorf("CreateActionTypes error: %s", err.Error())
				span.SetStatus(codes.Error, "创建业务知识网络动作类失败")
				return "", rest.NewHTTPError(ctx, http.StatusInternalServerError,
					berrors.BknBackend_KnowledgeNetwork_InternalError_CreateActionTypesFailed).
					WithErrorDetails(err.Error())
			}
		}

		if len(kn.RiskTypes) > 0 {
			_, err = kns.riskTypeS.CreateRiskTypes(ctx, tx, kn.RiskTypes, mode)
			if err != nil {
				logger.Errorf("CreateRiskTypes error: %s", err.Error())
				span.SetStatus(codes.Error, "创建业务知识网络风险类失败")
				return "", rest.NewHTTPError(ctx, http.StatusInternalServerError,
					berrors.BknBackend_RiskType_InternalError).
					WithErrorDetails(err.Error())
			}
		}
	}

	if isCreate || isUpdate {
		err = kns.InsertDatasetData(ctx, kn)
		if err != nil {
			logger.Errorf("InsertDatasetData error: %s", err.Error())
			span.SetStatus(codes.Error, "业务知识网络概念索引写入失败")

			return "", rest.NewHTTPError(ctx, http.StatusInternalServerError,
				berrors.BknBackend_KnowledgeNetwork_InternalError_InsertOpenSearchDataFailed).
				WithErrorDetails(err.Error())
		}
	}

	// 最后才绑定业务域，创建才绑业务域
	if isCreate {
		// 注册资源策略
		err = kns.ps.CreateResources(ctx, []interfaces.PermissionResource{{
			ID:   kn.KNID,
			Type: interfaces.RESOURCE_TYPE_KN,
			Name: kn.KNName,
		}}, interfaces.COMMON_OPERATIONS)
		if err != nil {
			logger.Errorf("CreateResources error: %s", err.Error())
			span.SetStatus(codes.Error, "创建业务知识网络资源失败")
			return "", rest.NewHTTPError(ctx, http.StatusInternalServerError,
				berrors.BknBackend_KnowledgeNetwork_InternalError_CreateResourcesFailed).
				WithErrorDetails(err.Error())
		}

		// 绑定业务域
		err = kns.bss.BindResource(ctx, kn.BusinessDomain, kn.KNID, interfaces.MODULE_TYPE_KN)
		if err != nil {
			logger.Errorf("BindResource error: %s", err.Error())
			span.SetStatus(codes.Error, "绑定业务知识网络业务系统失败")
			return "", rest.NewHTTPError(ctx, http.StatusInternalServerError,
				berrors.BknBackend_KnowledgeNetwork_InternalError_BindBusinessDomainFailed).
				WithErrorDetails(err.Error())
		}
	}

	span.SetStatus(codes.Ok, "")
	return kn.KNID, nil
}

// ValidateKN checks whole-KN dependency existence only; does not write to the database.
func (kns *knowledgeNetworkService) ValidateKN(ctx context.Context, kn *interfaces.KN, strictMode bool, mode string) error {

	ctx, span := ar_trace.Tracer.Start(ctx, "ValidateKN")
	defer span.End()

	if kn == nil {
		span.SetStatus(codes.Ok, "")
		return nil
	}

	knID := kn.KNID
	branch := kn.Branch
	if branch == "" {
		branch = interfaces.MAIN_BRANCH
	}
	kn.Branch = branch

	// 处理导入模式
	_, _, err := kns.handleKNImportMode(ctx, mode, kn)
	if err != nil {
		return err
	}

	batch, err := batchindex.CollectKNFromPayload(kn)
	if err != nil {
		return rest.NewHTTPError(ctx, http.StatusBadRequest,
			berrors.BknBackend_KnowledgeNetwork_InvalidParameter).
			WithErrorDetails(err.Error())
	}

	if len(kn.ConceptGroups) > 0 {
		if err := kns.cgs.ValidateConceptGroups(ctx, knID, branch, kn.ConceptGroups, strictMode, batch, mode); err != nil {
			return err
		}
	}
	if len(kn.ObjectTypes) > 0 {
		if err := kns.ots.ValidateObjectTypes(ctx, knID, branch, kn.ObjectTypes, strictMode, batch, mode); err != nil {
			return err
		}
	}
	if len(kn.RelationTypes) > 0 {
		if err := kns.rts.ValidateRelationTypes(ctx, knID, branch, kn.RelationTypes, strictMode, batch, mode); err != nil {
			return err
		}
	}
	if len(kn.ActionTypes) > 0 {
		if err := kns.ats.ValidateActionTypes(ctx, knID, branch, kn.ActionTypes, strictMode, batch, mode); err != nil {
			return err
		}
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

func (kns *knowledgeNetworkService) ListKNs(ctx context.Context, parameter interfaces.KNsQueryParams) ([]*interfaces.KN, int, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "查询业务知识网络列表")
	defer span.End()

	//获取业务知识网络列表
	KNArr, err := kns.kna.ListKNs(ctx, parameter)
	if err != nil {
		logger.Errorf("ListKNs error: %s", err.Error())
		span.SetStatus(codes.Error, "List knowledge networks error")

		return []*interfaces.KN{}, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_KnowledgeNetwork_InternalError).WithErrorDetails(err.Error())
	}
	if len(KNArr) == 0 {
		span.SetStatus(codes.Ok, "")
		return []*interfaces.KN{}, 0, nil
	}

	// 处理资源id
	KNIDs := make([]string, 0)
	for _, m := range KNArr {
		KNIDs = append(KNIDs, m.KNID)
	}

	// 根据权限过滤有查看权限的对象，过滤后的数组的总长度就是总数，无需再请求总数
	matchResoucesMap, err := kns.ps.FilterResources(ctx, interfaces.RESOURCE_TYPE_KN, KNIDs,
		[]string{interfaces.OPERATION_TYPE_VIEW_DETAIL}, true, interfaces.COMMON_OPERATIONS)
	if err != nil {
		span.SetStatus(codes.Error, "Filter resources error")
		return []*interfaces.KN{}, 0, err
	}

	KNs := make([]*interfaces.KN, 0)
	for _, kn := range KNArr {
		// 只留下有权限的模型
		if resrc, exist := matchResoucesMap[kn.KNID]; exist {
			kn.Operations = resrc.Operations // 用户当前有权限的操作
			KNs = append(KNs, kn)
		}
	}
	total := len(KNs)

	// limit = -1,则返回所有
	if parameter.Limit != -1 {
		// 分页
		// 检查起始位置是否越界
		if parameter.Offset < 0 || parameter.Offset >= len(KNs) {
			span.SetStatus(codes.Ok, "")
			return []*interfaces.KN{}, total, nil
		}
		// 计算结束位置
		end := parameter.Offset + parameter.Limit
		if end > len(KNs) {
			end = len(KNs)
		}

		KNs = KNs[parameter.Offset:end]
	}

	accountInfos := make([]*interfaces.AccountInfo, 0, len(KNs)*2)
	for _, kn := range KNs {
		accountInfos = append(accountInfos, &kn.Creator, &kn.Updater)
	}

	err = kns.ums.GetAccountNames(ctx, accountInfos)
	if err != nil {
		span.SetStatus(codes.Error, "GetAccountNames error")

		return []*interfaces.KN{}, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_KnowledgeNetwork_InternalError).WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return KNs, total, nil
}

func (kns *knowledgeNetworkService) GetKNByID(ctx context.Context, knID string, branch string, mode string) (*interfaces.KN, error) {

	// 获取业务知识网络
	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("查询业务知识网络[%s]信息", knID))
	defer span.End()

	// 获取模型基本信息
	kn, err := kns.kna.GetKNByID(ctx, knID, branch)
	if err != nil {
		logger.Errorf("GetKNByID error: %s", err.Error())
		span.SetStatus(codes.Error, fmt.Sprintf("Get knowledge network[%s] error: %v", knID, err))
		span.End()

		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_KnowledgeNetwork_InternalError_GetKNByIDFailed).WithErrorDetails(err.Error())
	}

	if kn == nil {
		errStr := fmt.Sprintf("Knowledge network[%s] not found", knID)
		logger.Errorf(errStr)
		span.SetStatus(codes.Error, errStr)
		span.End()

		return nil, rest.NewHTTPError(ctx, http.StatusNotFound,
			berrors.BknBackend_KnowledgeNetwork_NotFound).
			WithErrorDetails(errStr)
	}

	// 根据权限过滤有查看权限的对象，过滤后的数组的总长度就是总数，无需再请求总数
	matchResoucesMap, err := kns.ps.FilterResources(ctx, interfaces.RESOURCE_TYPE_KN, []string{kn.KNID},
		[]string{interfaces.OPERATION_TYPE_VIEW_DETAIL}, true, interfaces.COMMON_OPERATIONS)
	if err != nil {
		span.SetStatus(codes.Error, "Filter resources error")
		return nil, err
	}

	if resrc, exist := matchResoucesMap[kn.KNID]; exist {
		kn.Operations = resrc.Operations // 用户当前有权限的操作
	} else {
		return nil, rest.NewHTTPError(ctx, http.StatusForbidden, rest.PublicError_Forbidden).
			WithErrorDetails(fmt.Sprintf("Access denied: insufficient permissions for[%v]", interfaces.OPERATION_TYPE_VIEW_DETAIL))
	}

	accountInfos := []*interfaces.AccountInfo{&kn.Creator, &kn.Updater}
	err = kns.ums.GetAccountNames(ctx, accountInfos)
	if err != nil {
		span.SetStatus(codes.Error, "GetAccountNames error")

		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_KnowledgeNetwork_InternalError).WithErrorDetails(err.Error())
	}

	if mode == interfaces.Mode_Export {
		conceptGroups, _, err := kns.cgs.ListConceptGroups(ctx, interfaces.ConceptGroupsQueryParams{
			PaginationQueryParameters: interfaces.PaginationQueryParameters{
				Limit: -1,
			},
			KNID:   kn.KNID,
			Branch: kn.Branch,
		})
		if err != nil {
			return nil, err
		}
		kn.ConceptGroups = conceptGroups

		objectTypes, _, err := kns.ots.ListObjectTypes(ctx, nil, interfaces.ObjectTypesQueryParams{
			PaginationQueryParameters: interfaces.PaginationQueryParameters{
				Limit: -1,
			},
			KNID: kn.KNID,
		})
		if err != nil {
			return nil, err
		}
		kn.ObjectTypes = objectTypes

		relationTypes, _, err := kns.rts.ListRelationTypes(ctx, interfaces.RelationTypesQueryParams{
			PaginationQueryParameters: interfaces.PaginationQueryParameters{
				Limit: -1,
			},
			KNID: kn.KNID,
		})
		if err != nil {
			return nil, err
		}
		kn.RelationTypes = relationTypes

		actionTypes, _, err := kns.ats.ListActionTypes(ctx, interfaces.ActionTypesQueryParams{
			PaginationQueryParameters: interfaces.PaginationQueryParameters{
				Limit: -1,
			},
			KNID: kn.KNID,
		})
		if err != nil {
			return nil, err
		}
		kn.ActionTypes = actionTypes

		if kns.riskTypeA != nil {
			riskTypes, err := kns.riskTypeA.GetAllRiskTypesByKnID(ctx, kn.KNID, kn.Branch)
			if err != nil {
				return nil, err
			}
			kn.RiskTypes = riskTypes
		}
	}

	span.SetStatus(codes.Ok, "")
	return kn, nil
}

func (kns *knowledgeNetworkService) GetStatByKN(ctx context.Context, kn *interfaces.KN) (*interfaces.Statistics, error) {
	// 获取业务知识网络
	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("查询业务知识网络[%s]信息", kn.KNID))
	defer span.End()

	// 获取业务知识网络下的对象类、关系类、行动类的数量
	otCnt, err := kns.ota.GetObjectTypesTotal(ctx, interfaces.ObjectTypesQueryParams{
		KNID:   kn.KNID,
		Branch: kn.Branch,
	})
	if err != nil {
		logger.Errorf("GetObjectTypesTotal in knowledge network[%s] error: %s", kn.KNID, err.Error())
		span.SetStatus(codes.Error, fmt.Sprintf("GetObjectTypesTotal in knowledge network[%s], error: %v", kn.KNID, err))
		span.End()

		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_KnowledgeNetwork_InternalError_GetObjectTypesTotalFailed).WithErrorDetails(err.Error())
	}

	// 关系类数量
	rtCnt, err := kns.rta.GetRelationTypesTotal(ctx, interfaces.RelationTypesQueryParams{
		KNID:   kn.KNID,
		Branch: kn.Branch,
	})
	if err != nil {
		logger.Errorf("GetRelationTypesTotal in knowledge network[%s] error: %s", kn.KNID, err.Error())
		span.SetStatus(codes.Error, fmt.Sprintf("GetRelationTypesTotal in knowledge network[%s], error: %v", kn.KNID, err))
		span.End()

		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_KnowledgeNetwork_InternalError_GetRelationTypesTotalFailed).WithErrorDetails(err.Error())
	}

	// 行动类数量
	atCnt, err := kns.ata.GetActionTypesTotal(ctx, interfaces.ActionTypesQueryParams{
		KNID:   kn.KNID,
		Branch: kn.Branch,
	})
	if err != nil {
		logger.Errorf("GetActionTypesTotal in knowledge network[%s] error: %s", kn.KNID, err.Error())
		span.SetStatus(codes.Error, fmt.Sprintf("GetActionTypesTotal in knowledge network[%s], error: %v", kn.KNID, err))
		span.End()

		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_KnowledgeNetwork_InternalError_GetRelationTypesTotalFailed).WithErrorDetails(err.Error())
	}

	// 概念分组数量
	cgCnt, err := kns.cga.GetConceptGroupsTotal(ctx, interfaces.ConceptGroupsQueryParams{
		KNID:   kn.KNID,
		Branch: kn.Branch,
	})
	if err != nil {
		logger.Errorf("GetConceptGroupsTotal in knowledge network[%s] error: %s", kn.KNID, err.Error())
		span.SetStatus(codes.Error, fmt.Sprintf("GetConceptGroupsTotal in knowledge network[%s], error: %v", kn.KNID, err))
		span.End()

		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_KnowledgeNetwork_InternalError_GetRelationTypesTotalFailed).WithErrorDetails(err.Error())
	}

	// riskTypes 数量
	riskTypeCnt, err := kns.riskTypeA.GetRiskTypesTotal(ctx, interfaces.RiskTypesQueryParams{
		KNID:   kn.KNID,
		Branch: kn.Branch,
	})
	if err != nil {
		logger.Errorf("GetRiskTypesTotal in knowledge network[%s] error: %s", kn.KNID, err.Error())
		span.SetStatus(codes.Error, fmt.Sprintf("GetRiskTypesTotal in knowledge network[%s], error: %v", kn.KNID, err))
		span.End()

		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_KnowledgeNetwork_InternalError_GetRiskTypesTotalFailed).WithErrorDetails(err.Error())
	}

	statistics := &interfaces.Statistics{
		CgTotal:       cgCnt,
		OtTotal:       otCnt,
		RtTotal:       rtCnt,
		AtTotal:       atCnt,
		RiskTypeTotal: riskTypeCnt,
	}

	span.SetStatus(codes.Ok, "")
	return statistics, nil
}

// 更新业务知识网络
func (kns *knowledgeNetworkService) UpdateKN(ctx context.Context, tx *sql.Tx, kn *interfaces.KN, strictMode bool) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Update knowledge network")
	defer span.End()

	// 判断userid是否有创建业务知识网络的权限（策略决策）
	err := kns.ps.CheckPermission(ctx, interfaces.PermissionResource{
		Type: interfaces.RESOURCE_TYPE_KN,
		ID:   kn.KNID,
	}, []string{interfaces.OPERATION_TYPE_MODIFY})
	if err != nil {
		return err
	}

	if strictMode {
		if err := kns.ValidateKN(ctx, kn, strictMode, interfaces.ImportMode_Overwrite); err != nil {
			return err
		}
	}

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}
	kn.Updater = accountInfo

	currentTime := time.Now().UnixMilli() // 业务知识网络的update_time是int类型
	kn.UpdateTime = currentTime

	bknNetwork := logics.ToBKNNetWork(kn)
	kn.BKNRawContent = bknsdk.SerializeBknNetwork(bknNetwork)

	if tx == nil {
		// 0. 开始事务
		tx, err = kns.db.Begin()
		if err != nil {
			logger.Errorf("Begin transaction error: %s", err.Error())
			span.SetStatus(codes.Error, "事务开启失败")
			o11y.Error(ctx, fmt.Sprintf("Begin transaction error: %s", err.Error()))

			return rest.NewHTTPError(ctx, http.StatusInternalServerError,
				berrors.BknBackend_KnowledgeNetwork_InternalError_BeginTransactionFailed).
				WithErrorDetails(err.Error())
		}
		// 0.1 异常时
		defer func() {
			switch err {
			case nil:
				// 提交事务
				err = tx.Commit()
				if err != nil {
					logger.Errorf("UpdateKN Transaction Commit Failed:%v", err)
					span.SetStatus(codes.Error, "提交事务失败")
					o11y.Error(ctx, fmt.Sprintf("UpdateKN Transaction Commit Failed: %s", err.Error()))
				}
				logger.Infof("UpdateKN Transaction Commit Success:%v", kn.KNName)
				o11y.Debug(ctx, fmt.Sprintf("UpdateKN Transaction Commit Success: %s", kn.KNName))
			default:
				rollbackErr := tx.Rollback()
				if rollbackErr != nil {
					logger.Errorf("UpdateKN Transaction Rollback Error:%v", rollbackErr)
					span.SetStatus(codes.Error, "事务回滚失败")
					o11y.Error(ctx, fmt.Sprintf("UpdateKN Transaction Rollback Error: %s", rollbackErr.Error()))
				}
			}
		}()
	}

	// 更新模型信息
	err = kns.kna.UpdateKN(ctx, tx, kn)
	if err != nil {
		logger.Errorf("UpdateKN error: %s", err.Error())
		span.SetStatus(codes.Error, "修改业务知识网络失败")

		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_KnowledgeNetwork_InternalError).
			WithErrorDetails(err.Error())
	}

	err = kns.InsertDatasetData(ctx, kn)
	if err != nil {
		logger.Errorf("InsertDatasetData error: %s", err.Error())
		span.SetStatus(codes.Error, "业务知识网络概念索引写入失败")

		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_KnowledgeNetwork_InternalError_InsertOpenSearchDataFailed).
			WithErrorDetails(err.Error())
	}

	// 请求更新资源名称的接口，更新资源的名称
	if kn.IfNameModify {
		err = kns.ps.UpdateResource(ctx, interfaces.PermissionResource{
			ID:   kn.KNID,
			Type: interfaces.RESOURCE_TYPE_KN,
			Name: kn.KNName,
		})
		if err != nil {
			return err
		}
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

func (kns *knowledgeNetworkService) DeleteKN(ctx context.Context, kn *interfaces.KN) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "Delete knowledge network")
	defer span.End()

	// 判断userid是否有删除业务知识网络的权限
	err := kns.ps.CheckPermission(ctx, interfaces.PermissionResource{
		Type: interfaces.RESOURCE_TYPE_KN,
		ID:   kn.KNID,
	}, []string{interfaces.OPERATION_TYPE_DELETE})
	if err != nil {
		return err
	}

	// 获取业务知识网络下的所有任务
	jobs, _, err := kns.js.ListJobs(ctx, interfaces.JobsQueryParams{
		PaginationQueryParameters: interfaces.PaginationQueryParameters{
			Sort:      interfaces.JOB_SORT["create_time"],
			Direction: interfaces.DESC_DIRECTION,
		},
		KNID:   kn.KNID,
		Branch: kn.Branch,
		State:  []interfaces.JobState{interfaces.JobStateRunning},
	})
	if err != nil {
		return err
	}
	if len(jobs) > 0 {
		// 有运行中的任务，不允许删除
		return rest.NewHTTPError(ctx, http.StatusForbidden,
			berrors.BknBackend_KnowledgeNetwork_Forbidden_HasRunningJob).
			WithErrorDetails("业务知识网络下有运行中的任务，不允许删除")
	}

	// 0. 开始事务
	tx, err := kns.db.Begin()
	if err != nil {
		logger.Errorf("Begin transaction error: %s", err.Error())
		span.SetStatus(codes.Error, "事务开启失败")
		o11y.Error(ctx, fmt.Sprintf("Begin transaction error: %s", err.Error()))

		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_KnowledgeNetwork_InternalError_BeginTransactionFailed).
			WithErrorDetails(err.Error())
	}

	// 0.1 异常时
	defer func() {
		switch err {
		case nil:
			// 提交事务
			err = tx.Commit()
			if err != nil {
				logger.Errorf("DeleteKN Transaction Commit Failed:%v", err)
				span.SetStatus(codes.Error, "提交事务失败")
				o11y.Error(ctx, fmt.Sprintf("CreateKN Transaction Commit Failed: %s", err.Error()))
				return
			}
			logger.Infof("DeleteKN Transaction Commit Success")
			o11y.Debug(ctx, "DeleteKN Transaction Commit Success")
		default:
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				logger.Errorf("DeleteKN Transaction Rollback Error:%v", rollbackErr)
				span.SetStatus(codes.Error, "事务回滚失败")
				o11y.Error(ctx, fmt.Sprintf("CreateKN Transaction Rollback Error: %s", err.Error()))
			}
		}
	}()

	// 删除业务知识网络
	rowsAffect, err := kns.kna.DeleteKN(ctx, tx, kn.KNID, kn.Branch)
	if err != nil {
		logger.Errorf("DeleteKN error: %s", err.Error())
		span.SetStatus(codes.Error, "删除业务知识网络失败")

		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_KnowledgeNetwork_InternalError).WithErrorDetails(err.Error())
	}
	logger.Infof("DeleteKN: Rows affected is %v, request delete KNID is %s!", rowsAffect, kn.KNID)
	if rowsAffect != 1 {
		logger.Warnf("Delete kns number %v not equal 1!", rowsAffect)
		o11y.Warn(ctx, fmt.Sprintf("Delete kns number %v not equal 1!", rowsAffect))
	}

	// 删除业务知识网络下的所有对象类、关系类、行动类, 概念分组, 任务
	// 获取业务知识网络下的对象类id
	err = kns.ots.DeleteObjectTypesByKnID(ctx, tx, kn.KNID, kn.Branch)
	if err != nil {
		logger.Errorf("DeleteObjectTypesByKnID error: %s", err.Error())
		span.SetStatus(codes.Error, "删除业务知识网络对象类失败")
		return err
	}

	// 删除业务知识网络下的所有关系类
	err = kns.rts.DeleteRelationTypesByKnID(ctx, tx, kn.KNID, kn.Branch)
	if err != nil {
		logger.Errorf("DeleteRelationTypesByKnID error: %s", err.Error())
		span.SetStatus(codes.Error, "删除业务知识网络关系类失败")
		return err
	}

	// 删除业务知识网络下的所有行动类
	err = kns.ats.DeleteActionTypesByKnID(ctx, tx, kn.KNID, kn.Branch)
	if err != nil {
		logger.Errorf("DeleteActionTypesByKnID error: %s", err.Error())
		span.SetStatus(codes.Error, "删除业务知识网络动作类失败")
		return err
	}

	// 删除业务知识网络下的所有风险类
	err = kns.riskTypeS.DeleteRiskTypesByKnID(ctx, tx, kn.KNID, kn.Branch)
	if err != nil {
		logger.Errorf("DeleteRiskTypesByKnID error: %s", err.Error())
		span.SetStatus(codes.Error, "删除业务知识网络风险类失败")
		return err
	}

	// 删除业务知识网络下的所有概念分组
	err = kns.cgs.DeleteConceptGroupsByKnID(ctx, tx, kn.KNID, kn.Branch)
	if err != nil {
		logger.Errorf("DeleteConceptGroupsByKnID error: %s", err.Error())
		span.SetStatus(codes.Error, "删除业务知识网络概念分组失败")
		return err
	}

	// 删除业务知识网络下的所有任务
	err = kns.js.DeleteJobsByKnID(ctx, tx, kn.KNID, kn.Branch)
	if err != nil {
		logger.Errorf("DeleteJobsByKnID error: %s", err.Error())
		span.SetStatus(codes.Error, "删除业务知识网络任务失败")
		return err
	}

	docid := interfaces.GenerateConceptDocuemtnID(kn.KNID,
		interfaces.MODULE_TYPE_KN, kn.KNID, kn.Branch)
	err = kns.vba.DeleteDatasetDocumentByID(ctx, interfaces.BKN_DATASET_ID, docid)
	if err != nil {
		logger.Errorf("DeleteDatasetDocumentByID error: %s", err.Error())
		span.SetStatus(codes.Error, "删除业务知识网络概念失败")
		return err
	}

	// Delete all concepts under this KN by query condition
	filterCondition := map[string]any{
		"operation": "and",
		"sub_conditions": []map[string]any{
			{
				"field":      "kn_id",
				"operation":  "==",
				"value":      kn.KNID,
				"value_from": "const",
			},
			{
				"field":      "branch",
				"operation":  "==",
				"value":      kn.Branch,
				"value_from": "const",
			},
		},
	}
	err = kns.vba.DeleteDatasetDocumentsByQuery(ctx, interfaces.BKN_DATASET_ID, filterCondition)
	if err != nil {
		logger.Errorf("DeleteDatasetDocumentsByQuery error: %s", err.Error())
		span.SetStatus(codes.Error, "删除业务知识网络概念失败")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_KnowledgeNetwork_InternalError).
			WithErrorDetails(err.Error())
	}

	//  清除资源策略
	err = kns.ps.DeleteResources(ctx, interfaces.RESOURCE_TYPE_KN, []string{kn.KNID})
	if err != nil {
		logger.Errorf("DeleteResources error: %s", err.Error())
		span.SetStatus(codes.Error, "删除业务知识网络资源策略失败")
		return err
	}
	// 最后再解绑业务域
	err = kns.bss.UnbindResource(ctx, kn.BusinessDomain, kn.KNID, interfaces.RESOURCE_TYPE_KN)
	if err != nil {
		logger.Errorf("UnbindResource error: %s", err.Error())
		span.SetStatus(codes.Error, "解绑业务知识网络业务域失败")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_KnowledgeNetwork_InternalError_UnbindBusinessDomainFailed).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 更新知识网络详情
func (kns *knowledgeNetworkService) UpdateKNDetail(ctx context.Context, knID string, branch string, detail string) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "UpdateKNDetail")
	defer span.End()

	// 更新知识网络详情
	err := kns.kna.UpdateKNDetail(ctx, knID, branch, detail)
	if err != nil {
		logger.Errorf("UpdateKNDetail error: %s", err.Error())
		span.SetStatus(codes.Error, "修改知识网络详情失败")

		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_KnowledgeNetwork_InternalError).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

func (kns *knowledgeNetworkService) handleKNImportMode(ctx context.Context, mode string,
	kn *interfaces.KN) (isCreate bool, isUpdate bool, err error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "knowledge network import mode logic")
	defer span.End()

	isCreate = false
	isUpdate = false

	// 校验单个KN的导入模式逻辑
	idExist := false
	_, idExist, err = kns.CheckKNExistByID(ctx, kn.KNID, kn.Branch)
	if err != nil {
		return false, false, err
	}

	// 校验请求体与现有模型名称的重复性
	existID, nameExist, err := kns.CheckKNExistByName(ctx, kn.KNName, kn.Branch)
	if err != nil {
		return false, false, err
	}

	// 根据mode来区别，若是ignore，就从结果集中忽略，若是overwrite，就调用update，若是normal就报错。
	if idExist || nameExist {
		switch mode {
		case interfaces.ImportMode_Normal:
			if idExist {
				errDetails := fmt.Sprintf("The knowledge network with id [%s] already exists!", kn.KNID)
				logger.Error(errDetails)
				span.SetStatus(codes.Error, errDetails)
				return false, false, rest.NewHTTPError(ctx, http.StatusBadRequest,
					berrors.BknBackend_KnowledgeNetwork_KNIDExisted).
					WithErrorDetails(errDetails)
			}

			if nameExist {
				errDetails := fmt.Sprintf("knowledge network name '%s' already exists", kn.KNName)
				logger.Error(errDetails)
				span.SetStatus(codes.Error, errDetails)
				return false, false, rest.NewHTTPError(ctx, http.StatusForbidden,
					berrors.BknBackend_KnowledgeNetwork_KNNameExisted).
					WithDescription(map[string]any{"kn_name": kn.KNName}).
					WithErrorDetails(errDetails)
			}

		case interfaces.ImportMode_Ignore:
			// 存在重复的就跳过，不创建也不更新
			return false, false, nil
		case interfaces.ImportMode_Overwrite:
			if idExist && nameExist {
				// 如果 id 和名称都存在，但是存在的名称对应的视图 id 和当前视图 id 不一样，则报错
				if existID != kn.KNID {
					errDetails := fmt.Sprintf("KN ID '%s' and name '%s' already exist, but the exist knowledge network id is '%s'",
						kn.KNID, kn.KNName, existID)
					logger.Error(errDetails)
					span.SetStatus(codes.Error, errDetails)
					return false, false, rest.NewHTTPError(ctx, http.StatusForbidden,
						berrors.BknBackend_KnowledgeNetwork_KNNameExisted).
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
				errDetails := fmt.Sprintf("KN ID '%s' does not exist, but name '%s' already exists",
					kn.KNID, kn.KNName)
				logger.Error(errDetails)
				span.SetStatus(codes.Error, errDetails)
				return false, false, rest.NewHTTPError(ctx, http.StatusForbidden,
					berrors.BknBackend_KnowledgeNetwork_KNNameExisted).
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

func (kns *knowledgeNetworkService) InsertDatasetData(ctx context.Context, origKN *interfaces.KN) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "业务知识网络概念索引写入")
	defer span.End()

	kn := &interfaces.KN{
		KNID:           origKN.KNID,
		KNName:         origKN.KNName,
		Tags:           origKN.Tags,
		Comment:        origKN.Comment,
		Icon:           origKN.Icon,
		Color:          origKN.Color,
		BKNRawContent:  origKN.BKNRawContent,
		Branch:         origKN.Branch,
		BusinessDomain: origKN.BusinessDomain,
		Creator:        origKN.Creator,
		CreateTime:     origKN.CreateTime,
		Updater:        origKN.Updater,
		UpdateTime:     origKN.UpdateTime,
		ModuleType:     interfaces.MODULE_TYPE_KN,
	}

	if kns.appSetting.ServerSetting.DefaultSmallModelEnabled {
		words := []string{kn.KNName}
		words = append(words, kn.Tags...)
		words = append(words, kn.Comment, kn.BKNRawContent)
		word := strings.Join(words, "\n")

		defaultModel, err := kns.mfa.GetDefaultModel(ctx)
		if err != nil {
			logger.Errorf("GetDefaultModel error: %s", err.Error())
			span.SetStatus(codes.Error, "获取默认模型失败")
			return err
		}
		vectors, err := kns.mfa.GetVector(ctx, defaultModel, []string{word})
		if err != nil {
			logger.Errorf("GetVector error: %s", err.Error())
			span.SetStatus(codes.Error, "获取业务知识网络向量失败")
			return err
		}

		kn.Vector = vectors[0].Vector
	}

	docid := interfaces.GenerateConceptDocuemtnID(kn.KNID, interfaces.MODULE_TYPE_KN, kn.KNID, kn.Branch)

	// Convert to map for dataset
	docBytes, err := sonic.Marshal(kn)
	if err != nil {
		logger.Errorf("Failed to marshal KN: %s", err.Error())
		span.SetStatus(codes.Error, "序列化业务知识网络失败")
		return err
	}

	var doc map[string]any
	if err := sonic.Unmarshal(docBytes, &doc); err != nil {
		logger.Errorf("Failed to unmarshal KN: %s", err.Error())
		span.SetStatus(codes.Error, "反序列化业务知识网络失败")
		return err
	}

	// Set document ID
	doc["_id"] = docid

	err = kns.vba.WriteDatasetDocuments(ctx, interfaces.BKN_DATASET_ID, []map[string]any{doc})
	if err != nil {
		logger.Errorf("WriteDatasetDocuments error: %s", err.Error())
		span.SetStatus(codes.Error, "业务知识网络概念索引写入失败")
		return err
	}

	return nil
}

// 批量查询的中间状态
type batchQueryState struct {
	visited   map[string]bool
	batchSize int
}

// 根据起点对象类，方向，长度获取路径
func (kns *knowledgeNetworkService) GetRelationTypePaths(ctx context.Context,
	query interfaces.RelationTypePathsBaseOnSource) ([]interfaces.RelationTypePath, error) {
	// 1. 获取起点对象类

	allPaths := []interfaces.RelationTypePath{}

	// 使用BFS进行路径搜索
	queue := []interfaces.RelationTypePath{
		{
			ObjectTypes: []interfaces.ObjectTypeWithKeyField{
				{
					OTID: query.SourceObjecTypeId,
				},
			},
			Length: 0,
		},
	}

	// 初始化状态
	state := &batchQueryState{
		visited: map[string]bool{}, // 用于防止循环路径
		// objectTypeCache: map[string]interfaces.ObjectType{},
		batchSize: 50, // 每批查询的节点数量
	}
	for len(queue) > 0 {
		currentLevelSize := len(queue)
		var nextLevelNodes []string
		currentLevelPaths := make([]interfaces.RelationTypePath, 0, currentLevelSize)

		// 处理当前层的所有路径
		for i := 0; i < currentLevelSize; i++ {
			currentPath := queue[i]
			currentNode := currentPath.ObjectTypes[len(currentPath.ObjectTypes)-1]
			// 获取当前节点的信息（按需查询）
			if currentNode.OTName == "" {
				// 若currentNode.OTID不存在，此函数会报错： objetc type not found
				objectType, err := kns.ots.GetObjectTypeByID(ctx, nil, query.KNID, query.Branch, currentNode.OTID)
				if err != nil {
					return nil, err
				}
				currentNode = interfaces.ObjectTypeWithKeyField{
					OTID:            objectType.OTID,
					OTName:          objectType.OTName,
					DataSource:      objectType.DataSource,
					DataProperties:  objectType.DataProperties,
					LogicProperties: objectType.LogicProperties,
					PrimaryKeys:     objectType.PrimaryKeys,
					DisplayKey:      objectType.DisplayKey,
					IncrementalKey:  objectType.IncrementalKey,
				}
				currentPath.ObjectTypes[len(currentPath.ObjectTypes)-1] = currentNode
			}

			// 如果达到最大深度，保存路径
			if currentPath.Length >= query.PathLength {
				allPaths = append(allPaths, currentPath)
				continue
			}

			// 收集需要查询邻居的节点ID
			nextLevelNodes = append(nextLevelNodes, currentNode.OTID)
			currentLevelPaths = append(currentLevelPaths, currentPath)
		}

		// 批量查询下一层节点的邻居
		if len(nextLevelNodes) > 0 {
			neighborPathsMap, err := kns.getNeighborsBatch(ctx, nextLevelNodes, query, state)
			if err != nil {
				return nil, err
			}

			// 为每个当前层的路径扩展新路径
			for i, currentPath := range currentLevelPaths {

				currentNodeID := nextLevelNodes[i]
				neighborPaths := neighborPathsMap[currentNodeID]

				// 如果没有邻居节点，保存当前路径
				if len(neighborPaths) == 0 {
					allPaths = append(allPaths, currentPath)
					continue
				}

				// 为每个邻居创建新路径. 当前起点指向的路径需要重置，不完整。
				for _, neighbor := range neighborPaths {
					// 构建路径键来检测循环
					// 这个一度的路径，第二个对象类是终点
					pathKey := buildPathKey(currentPath, neighbor)
					if state.visited[pathKey] {
						continue // 跳过已访问的路径
					}
					state.visited[pathKey] = true

					newPath := interfaces.RelationTypePath{
						ObjectTypes: make([]interfaces.ObjectTypeWithKeyField, len(currentPath.ObjectTypes)),
						TypeEdges:   make([]interfaces.TypeEdge, len(currentPath.TypeEdges)),
						Length:      currentPath.Length + 1,
					}
					copy(newPath.ObjectTypes, currentPath.ObjectTypes)
					copy(newPath.TypeEdges, currentPath.TypeEdges)
					newPath.ObjectTypes = append(newPath.ObjectTypes, neighbor.ObjectTypes[1])
					newPath.TypeEdges = append(newPath.TypeEdges, neighbor.TypeEdges...)

					queue = append(queue, newPath)
				}
			}
		}
		// 移除已处理的当前层路径
		if currentLevelSize > 0 {
			queue = queue[currentLevelSize:]
		}
	}
	// 添加队列中剩余的路径（如果未达到限制）
	for i := 0; i < len(queue); i++ {
		allPaths = append(allPaths, queue[i])
	}

	return allPaths, nil
}

// 批量查询相邻节点 - 核心优化方法
func (kns *knowledgeNetworkService) getNeighborsBatch(ctx context.Context, objectClassIDs []string,
	query interfaces.RelationTypePathsBaseOnSource, state *batchQueryState) (map[string][]interfaces.RelationTypePath, error) {

	if len(objectClassIDs) == 0 {
		return nil, nil
	}

	// 分批处理，避免SQL参数过多
	batchSize := state.batchSize
	neighborPathsMap := map[string][]interfaces.RelationTypePath{}

	for start := 0; start < len(objectClassIDs); start += batchSize {
		// 遍历当前节点的邻居路径
		end := start + batchSize
		if end > len(objectClassIDs) {
			end = len(objectClassIDs)
		}

		batchIDs := objectClassIDs[start:end]
		batchNeighborPathsMap, err := kns.kna.GetNeighborPathsBatch(ctx, batchIDs, query)
		if err != nil {
			return nil, err
		}

		// 合并结果
		for k, v := range batchNeighborPathsMap {
			neighborPathsMap[k] = append(neighborPathsMap[k], v...)
		}
	}

	return neighborPathsMap, nil
}

// 构建路径键用于循环检测
func buildPathKey(path interfaces.RelationTypePath, neighborPath interfaces.RelationTypePath) string {

	key := ""
	for i := 1; i < len(path.ObjectTypes); i++ {
		key += fmt.Sprintf("%s:%s->%s", path.TypeEdges[i-1].RelationTypeId, path.ObjectTypes[i-1].OTID, path.ObjectTypes[i].OTID)
	}
	key += fmt.Sprintf("%s:%s->%s",
		neighborPath.TypeEdges[0].RelationTypeId, neighborPath.ObjectTypes[0].OTID, neighborPath.ObjectTypes[1].OTID)
	return key
}

// 获取业务知识网络资源列表
func (kns *knowledgeNetworkService) ListKnSrcs(ctx context.Context,
	parameter interfaces.KNsQueryParams) ([]interfaces.PermissionResource, int, error) {

	listCtx, listSpan := ar_trace.Tracer.Start(ctx, "查询业务知识网络实例列表")
	defer listSpan.End()

	//获取业务知识网络列表（不分页，获取所有的业务知识网络)
	knList, err := kns.kna.ListKnSrcs(listCtx, parameter)
	emptyResources := []interfaces.PermissionResource{}
	if err != nil {
		logger.Errorf("ListSimpleKns error: %s", err.Error())
		listSpan.SetStatus(codes.Error, "List simple knowledge networks error")
		listSpan.End()
		return emptyResources, 0, rest.NewHTTPError(listCtx, http.StatusInternalServerError,
			berrors.BknBackend_KnowledgeNetwork_InternalError).WithErrorDetails(err.Error())
	}
	if len(knList) == 0 {
		return emptyResources, 0, nil
	}

	// 根据权限过滤有查看权限的对象，过滤后的数组的总长度就是总数，无需再请求总数
	// 处理资源id
	resMids := make([]string, 0)
	for _, m := range knList {
		resMids = append(resMids, m.ID)
	}
	// 校验权限管理的操作权限
	matchResoucesMap, err := kns.ps.FilterResources(ctx, interfaces.RESOURCE_TYPE_KN, resMids,
		[]string{interfaces.OPERATION_TYPE_VIEW_DETAIL}, false, interfaces.COMMON_OPERATIONS)
	if err != nil {
		return emptyResources, 0, err
	}

	// 遍历对象
	results := make([]interfaces.PermissionResource, 0)
	for _, knSrc := range knList {
		if _, exist := matchResoucesMap[knSrc.ID]; exist {
			results = append(results, knSrc)
		}
	}

	// limit = -1,则返回所有
	if parameter.Limit == -1 {
		return results, len(results), nil
	}

	// 分页
	// 检查起始位置是否越界
	if parameter.Offset < 0 || parameter.Offset >= len(results) {
		return nil, len(results), nil
	}
	// 计算结束位置
	end := parameter.Offset + parameter.Limit
	if end > len(results) {
		end = len(results)
	}

	listSpan.SetStatus(codes.Ok, "")
	return results[parameter.Offset:end], len(results), nil
}

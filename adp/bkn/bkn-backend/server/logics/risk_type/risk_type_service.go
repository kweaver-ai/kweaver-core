// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package risk_type

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
	"bkn-backend/logics/permission"
)

var (
	rtsOnce sync.Once
	rts     interfaces.RiskTypeService
)

type riskTypeService struct {
	appSetting *common.AppSetting
	db         *sql.DB
	rta        interfaces.RiskTypeAccess
	ps         interfaces.PermissionService
	uma        interfaces.UserMgmtAccess
	vba        interfaces.VegaBackendAccess
	mfa        interfaces.ModelFactoryAccess
}

func NewRiskTypeService(appSetting *common.AppSetting) interfaces.RiskTypeService {
	rtsOnce.Do(func() {
		rts = &riskTypeService{
			appSetting: appSetting,
			db:         logics.DB,
			rta:        logics.RiskTypeAccess,
			ps:         permission.NewPermissionService(appSetting),
			uma:        logics.UMA,
			vba:        logics.VBA,
			mfa:        logics.MFA,
		}
	})
	return rts
}

func (rts *riskTypeService) CheckRiskTypeExistByID(ctx context.Context, knID string, branch string, rtID string) (string, bool, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "CheckRiskTypeExistByID")
	defer span.End()

	rtName, exist, err := rts.rta.CheckRiskTypeExistByID(ctx, knID, branch, rtID)
	if err != nil {
		logger.Errorf("CheckRiskTypeExistByID error: %s", err.Error())
		o11y.Error(ctx, fmt.Sprintf("按ID[%s]获取风险类失败: %v", rtID, err))
		span.SetStatus(codes.Error, fmt.Sprintf("按ID[%s]获取风险类失败", rtID))
		return "", false, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_RiskType_InternalError_CheckRiskTypeIfExistFailed).WithErrorDetails(err.Error())
	}
	span.SetStatus(codes.Ok, "")
	return rtName, exist, nil
}

func (rts *riskTypeService) CheckRiskTypeExistByName(ctx context.Context, knID string, branch string, rtName string) (string, bool, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "CheckRiskTypeExistByName")
	defer span.End()

	rtID, exist, err := rts.rta.CheckRiskTypeExistByName(ctx, knID, branch, rtName)
	if err != nil {
		logger.Errorf("CheckRiskTypeExistByName error: %s", err.Error())
		o11y.Error(ctx, fmt.Sprintf("按名称[%s]获取风险类失败: %v", rtName, err))
		span.SetStatus(codes.Error, fmt.Sprintf("按名称[%s]获取风险类失败", rtName))
		return "", false, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_RiskType_InternalError_CheckRiskTypeIfExistFailed).WithErrorDetails(err.Error())
	}
	span.SetStatus(codes.Ok, "")
	return rtID, exist, nil
}

func (rts *riskTypeService) CreateRiskTypes(ctx context.Context, tx *sql.Tx, riskTypes []*interfaces.RiskType, mode string) ([]string, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "CreateRiskTypes")
	defer span.End()

	err := rts.ps.CheckPermission(ctx, interfaces.PermissionResource{
		Type: interfaces.RESOURCE_TYPE_KN,
		ID:   riskTypes[0].KNID,
	}, []string{interfaces.OPERATION_TYPE_MODIFY})
	if err != nil {
		return nil, err
	}

	currentTime := time.Now().UnixMilli()
	for _, rt := range riskTypes {
		if rt.RTID == "" {
			rt.RTID = xid.New().String()
		}
		accountInfo := interfaces.AccountInfo{}
		if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
			accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
		}
		rt.Creator = accountInfo
		rt.Updater = accountInfo
		rt.CreateTime = currentTime
		rt.UpdateTime = currentTime
		rt.ModuleType = interfaces.MODULE_TYPE_RISK_TYPE
	}

	if tx == nil {
		tx, err = rts.db.Begin()
		if err != nil {
			logger.Errorf("Begin transaction error: %s", err.Error())
			span.SetStatus(codes.Error, "事务开启失败")
			return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				berrors.BknBackend_RiskType_InternalError).WithErrorDetails(err.Error())
		}
		defer func() {
			if err != nil {
				_ = tx.Rollback()
			} else {
				_ = tx.Commit()
			}
		}()
	}

	createList, updateList, err := rts.handleImportMode(ctx, mode, riskTypes)
	if err != nil {
		return nil, err
	}

	rtIDs := []string{}
	for _, rt := range createList {
		rtIDs = append(rtIDs, rt.RTID)
		if err = rts.rta.CreateRiskType(ctx, tx, rt); err != nil {
			logger.Errorf("CreateRiskType error: %s", err.Error())
			return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				berrors.BknBackend_RiskType_InternalError).WithErrorDetails(err.Error())
		}
	}
	for _, rt := range updateList {
		rtIDs = append(rtIDs, rt.RTID)
		if err = rts.UpdateRiskType(ctx, tx, rt); err != nil {
			return nil, err
		}
	}

	insertList := append(createList, updateList...)
	if len(insertList) > 0 {
		err = rts.InsertDatasetData(ctx, insertList)
		if err != nil {
			logger.Errorf("InsertDatasetData error: %s", err.Error())
			span.SetStatus(codes.Error, "风险类索引写入失败")
			return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				berrors.BknBackend_RiskType_InternalError).
				WithErrorDetails(err.Error())
		}
	}

	span.SetStatus(codes.Ok, "")
	return rtIDs, nil
}

func (rts *riskTypeService) handleImportMode(ctx context.Context, mode string, riskTypes []*interfaces.RiskType) (createList, updateList []*interfaces.RiskType, err error) {
	if mode != interfaces.ImportMode_Normal && mode != interfaces.ImportMode_Ignore && mode != interfaces.ImportMode_Overwrite {
		return nil, nil, rest.NewHTTPError(ctx, http.StatusBadRequest,
			berrors.BknBackend_InvalidParameter_ImportMode).WithErrorDetails("invalid import_mode")
	}

	ctx, span := ar_trace.Tracer.Start(ctx, "handleRiskTypeImportMode")
	defer span.End()

	creates := []*interfaces.RiskType{}
	updates := []*interfaces.RiskType{}

	for _, rt := range riskTypes {
		creates = append(creates, rt)

		_, idExist, e := rts.rta.CheckRiskTypeExistByID(ctx, rt.KNID, rt.Branch, rt.RTID)
		if e != nil {
			return nil, nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				berrors.BknBackend_RiskType_InternalError_CheckRiskTypeIfExistFailed).WithErrorDetails(e.Error())
		}

		existID, nameExist, e := rts.rta.CheckRiskTypeExistByName(ctx, rt.KNID, rt.Branch, rt.RTName)
		if e != nil {
			return nil, nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				berrors.BknBackend_RiskType_InternalError_CheckRiskTypeIfExistFailed).WithErrorDetails(e.Error())
		}

		if idExist || nameExist {
			switch mode {
			case interfaces.ImportMode_Normal:
				if idExist {
					return nil, nil, rest.NewHTTPError(ctx, http.StatusBadRequest,
						berrors.BknBackend_RiskType_RiskTypeIDExisted).
						WithErrorDetails(fmt.Sprintf("RiskType ID '%s' already exists", rt.RTID))
				}
				if nameExist {
					errDetails := fmt.Sprintf("risk type name '%s' already exists", rt.RTName)
					logger.Error(errDetails)
					span.SetStatus(codes.Error, errDetails)
					return nil, nil, rest.NewHTTPError(ctx, http.StatusBadRequest,
						berrors.BknBackend_RiskType_RiskTypeNameExisted).
						WithDescription(map[string]any{"name": rt.RTName}).
						WithErrorDetails(errDetails)
				}

			case interfaces.ImportMode_Ignore:
				creates = creates[:len(creates)-1]

			case interfaces.ImportMode_Overwrite:
				if idExist && nameExist {
					if existID != rt.RTID {
						errDetails := fmt.Sprintf("RiskType ID '%s' and name '%s' already exist, but the exist risk type id is '%s'",
							rt.RTID, rt.RTName, existID)
						logger.Error(errDetails)
						span.SetStatus(codes.Error, errDetails)
						return nil, nil, rest.NewHTTPError(ctx, http.StatusBadRequest,
							berrors.BknBackend_RiskType_RiskTypeNameExisted).
							WithErrorDetails(errDetails)
					}
					creates = creates[:len(creates)-1]
					updates = append(updates, rt)
				}
				if idExist && !nameExist {
					creates = creates[:len(creates)-1]
					updates = append(updates, rt)
				}
				if !idExist && nameExist {
					errDetails := fmt.Sprintf("RiskType ID '%s' does not exist, but name '%s' already exists",
						rt.RTID, rt.RTName)
					logger.Error(errDetails)
					span.SetStatus(codes.Error, errDetails)
					return nil, nil, rest.NewHTTPError(ctx, http.StatusBadRequest,
						berrors.BknBackend_RiskType_RiskTypeNameExisted).
						WithErrorDetails(errDetails)
				}
			}
		}
	}

	return creates, updates, nil
}

func (rts *riskTypeService) ListRiskTypes(ctx context.Context, query interfaces.RiskTypesQueryParams) ([]*interfaces.RiskType, int, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "ListRiskTypes")
	defer span.End()

	err := rts.ps.CheckPermission(ctx, interfaces.PermissionResource{
		Type: interfaces.RESOURCE_TYPE_KN,
		ID:   query.KNID,
	}, []string{interfaces.OPERATION_TYPE_VIEW_DETAIL})
	if err != nil {
		return nil, 0, err
	}

	list, err := rts.rta.ListRiskTypes(ctx, query)
	if err != nil {
		logger.Errorf("ListRiskTypes error: %s", err.Error())
		return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_RiskType_InternalError).WithErrorDetails(err.Error())
	}

	total, err := rts.rta.GetRiskTypesTotal(ctx, query)
	if err != nil {
		logger.Errorf("GetRiskTypesTotal error: %s", err.Error())
		return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_RiskType_InternalError).WithErrorDetails(err.Error())
	}

	// limit = -1 则返回所有
	if query.Limit != -1 {
		if query.Offset < 0 || query.Offset >= len(list) {
			span.SetStatus(codes.Ok, "")
			return []*interfaces.RiskType{}, total, nil
		}
		end := query.Offset + query.Limit
		if end > len(list) {
			end = len(list)
		}
		list = list[query.Offset:end]
	}

	if len(list) > 0 && rts.uma != nil {
		accountInfos := make([]*interfaces.AccountInfo, 0, len(list)*2)
		for _, rt := range list {
			accountInfos = append(accountInfos, &rt.Creator, &rt.Updater)
		}
		_ = rts.uma.GetAccountNames(ctx, accountInfos)
	}

	span.SetStatus(codes.Ok, "")
	return list, total, nil
}

func (rts *riskTypeService) GetRiskTypesByIDs(ctx context.Context, knID string, branch string, rtIDs []string) ([]*interfaces.RiskType, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "GetRiskTypesByIDs")
	defer span.End()

	rtIDs = common.DuplicateSlice(rtIDs)
	list, err := rts.rta.GetRiskTypesByIDs(ctx, knID, branch, rtIDs)
	if err != nil {
		logger.Errorf("GetRiskTypesByIDs error: %s", err.Error())
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_RiskType_InternalError_GetRiskTypesByIDsFailed).WithErrorDetails(err.Error())
	}
	span.SetStatus(codes.Ok, "")
	return list, nil
}

func (rts *riskTypeService) UpdateRiskType(ctx context.Context, tx *sql.Tx, riskType *interfaces.RiskType) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "UpdateRiskType")
	defer span.End()

	err := rts.ps.CheckPermission(ctx, interfaces.PermissionResource{
		Type: interfaces.RESOURCE_TYPE_KN,
		ID:   riskType.KNID,
	}, []string{interfaces.OPERATION_TYPE_MODIFY})
	if err != nil {
		return err
	}

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}
	riskType.Updater = accountInfo
	riskType.UpdateTime = time.Now().UnixMilli()

	if tx == nil {
		tx, err = rts.db.Begin()
		if err != nil {
			return rest.NewHTTPError(ctx, http.StatusInternalServerError,
				berrors.BknBackend_RiskType_InternalError).WithErrorDetails(err.Error())
		}
		defer func() {
			if err != nil {
				_ = tx.Rollback()
			} else {
				_ = tx.Commit()
			}
		}()
	}

	if err = rts.rta.UpdateRiskType(ctx, tx, riskType); err != nil {
		logger.Errorf("UpdateRiskType error: %s", err.Error())
		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_RiskType_InternalError).WithErrorDetails(err.Error())
	}

	err = rts.InsertDatasetData(ctx, []*interfaces.RiskType{riskType})
	if err != nil {
		logger.Errorf("InsertDatasetData error: %s", err.Error())
		span.SetStatus(codes.Error, "风险类索引写入失败")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_RiskType_InternalError).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

func (rts *riskTypeService) DeleteRiskTypesByIDs(ctx context.Context, tx *sql.Tx, knID string, branch string, rtIDs []string) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "DeleteRiskTypesByIDs")
	defer span.End()

	err := rts.ps.CheckPermission(ctx, interfaces.PermissionResource{
		Type: interfaces.RESOURCE_TYPE_KN,
		ID:   knID,
	}, []string{interfaces.OPERATION_TYPE_MODIFY})
	if err != nil {
		return err
	}

	if tx == nil {
		tx, err = rts.db.Begin()
		if err != nil {
			return rest.NewHTTPError(ctx, http.StatusInternalServerError,
				berrors.BknBackend_RiskType_InternalError).WithErrorDetails(err.Error())
		}
		defer func() {
			if err != nil {
				_ = tx.Rollback()
			} else {
				_ = tx.Commit()
			}
		}()
	}

	_, err = rts.rta.DeleteRiskTypesByIDs(ctx, tx, knID, branch, rtIDs)
	if err != nil {
		logger.Errorf("DeleteRiskTypesByIDs error: %s", err.Error())
		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_RiskType_InternalError).WithErrorDetails(err.Error())
	}

	for _, rtID := range rtIDs {
		docid := interfaces.GenerateConceptDocuemtnID(knID, interfaces.MODULE_TYPE_RISK_TYPE, rtID, branch)
		err = rts.vba.DeleteDatasetDocumentByID(ctx, interfaces.BKN_DATASET_ID, docid)
		if err != nil {
			logger.Errorf("DeleteDatasetDocumentByID error: %s", err.Error())
			span.SetStatus(codes.Error, "删除风险类概念索引失败")
			return err
		}
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

func (rts *riskTypeService) InsertDatasetData(ctx context.Context, riskTypes []*interfaces.RiskType) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "风险类索引写入")
	defer span.End()

	if len(riskTypes) == 0 {
		return nil
	}

	if rts.appSetting.ServerSetting.DefaultSmallModelEnabled && rts.mfa != nil {
		words := []string{}
		for _, riskType := range riskTypes {
			arr := []string{riskType.RTName}
			arr = append(arr, riskType.Tags...)
			arr = append(arr, riskType.Comment, riskType.BKNRawContent)
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
			span.SetStatus(codes.Error, "获取风险类向量失败")
			return err
		}

		if len(vectors) != len(riskTypes) {
			logger.Errorf("GetVector error: expect vectors num is [%d], actual vectors num is [%d]", len(riskTypes), len(vectors))
			span.SetStatus(codes.Error, "获取风险类向量失败")
			return fmt.Errorf("GetVector error: expect vectors num is [%d], actual vectors num is [%d]", len(riskTypes), len(vectors))
		}

		for i, riskType := range riskTypes {
			riskType.Vector = vectors[i].Vector
		}
	}

	documents := make([]map[string]any, 0, len(riskTypes))
	for _, riskType := range riskTypes {
		docid := interfaces.GenerateConceptDocuemtnID(riskType.KNID, interfaces.MODULE_TYPE_RISK_TYPE,
			riskType.RTID, riskType.Branch)
		riskType.ModuleType = interfaces.MODULE_TYPE_RISK_TYPE

		docBytes, err := sonic.Marshal(riskType)
		if err != nil {
			logger.Errorf("Failed to marshal RiskType: %s", err.Error())
			span.SetStatus(codes.Error, "序列化风险类失败")
			return err
		}

		var doc map[string]any
		if err := sonic.Unmarshal(docBytes, &doc); err != nil {
			logger.Errorf("Failed to unmarshal RiskType: %s", err.Error())
			span.SetStatus(codes.Error, "反序列化风险类失败")
			return err
		}

		doc["_id"] = docid
		documents = append(documents, doc)
	}

	err := rts.vba.WriteDatasetDocuments(ctx, interfaces.BKN_DATASET_ID, documents)
	if err != nil {
		logger.Errorf("WriteDatasetDocuments error: %s", err.Error())
		span.SetStatus(codes.Error, "风险类概念索引写入失败")
		return err
	}

	return nil
}

func (rts *riskTypeService) SearchRiskTypes(ctx context.Context, query *interfaces.ConceptsQuery) (interfaces.RiskTypes, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "SearchRiskTypes")
	defer span.End()

	response := interfaces.RiskTypes{}
	var err error

	err = rts.ps.CheckPermission(ctx, interfaces.PermissionResource{
		Type: interfaces.RESOURCE_TYPE_KN,
		ID:   query.KNID,
	}, []string{interfaces.OPERATION_TYPE_VIEW_DETAIL})
	if err != nil {
		return response, err
	}

	var filterCondition map[string]any
	if query.ActualCondition != nil {
		filterCondition, err = cond.ConvertCondCfgToFilterCondition(ctx, query.ActualCondition,
			interfaces.CONCPET_QUERY_FIELD,
			func(ctx context.Context, word string) ([]*cond.VectorResp, error) {
				if !rts.appSetting.ServerSetting.DefaultSmallModelEnabled || rts.mfa == nil {
					err = errors.New(cond.DEFAULT_SMALL_MODEL_ENABLED_FALSE_ERROR)
					span.SetStatus(codes.Error, err.Error())
					return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
						berrors.BknBackend_RiskType_InternalError).
						WithErrorDetails(err.Error())
				}
				dftModel, err := rts.mfa.GetDefaultModel(ctx)
				if err != nil {
					logger.Errorf("GetDefaultModel error: %s", err.Error())
					span.SetStatus(codes.Error, "获取默认模型失败")
					return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
						berrors.BknBackend_RiskType_InternalError).
						WithErrorDetails(err.Error())
				}
				result, err := rts.mfa.GetVector(ctx, dftModel, []string{word})
				if err != nil {
					logger.Errorf("GetVector error: %s", err.Error())
					span.SetStatus(codes.Error, "获取风险类向量失败")
					return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
						berrors.BknBackend_RiskType_InternalError).
						WithErrorDetails(err.Error())
				}
				return result, nil
			})
		if err != nil {
			return response, rest.NewHTTPError(ctx, http.StatusBadRequest,
				berrors.BknBackend_RiskType_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("failed to convert condition to filter condition, %s", err.Error()))
		}
	}

	if query.NeedTotal {
		params := &interfaces.DatasetQueryParams{
			FilterCondition: filterCondition,
			Offset:          0,
			Limit:           1,
			NeedTotal:       true,
		}
		datasetResp, err := rts.vba.QueryDatasetData(ctx, interfaces.BKN_DATASET_ID, params)
		if err != nil {
			logger.Errorf("QueryDatasetData error: %s", err.Error())
			span.SetStatus(codes.Error, "风险类检索查询总数失败")
			return response, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				berrors.BknBackend_RiskType_InternalError).
				WithErrorDetails(err.Error())
		}
		response.TotalCount = datasetResp.TotalCount
	}

	riskTypes := []*interfaces.RiskType{}
	offset := 0
	limit := query.Limit
	if limit == 0 {
		limit = interfaces.SearchAfter_Limit
	}

	for {
		params := &interfaces.DatasetQueryParams{
			FilterCondition: filterCondition,
			Offset:          offset,
			Limit:           limit,
			NeedTotal:       false,
			Sort:            query.Sort,
		}
		datasetResp, err := rts.vba.QueryDatasetData(ctx, interfaces.BKN_DATASET_ID, params)
		if err != nil {
			logger.Errorf("QueryDatasetData error: %s", err.Error())
			span.SetStatus(codes.Error, "风险类检索查询失败")
			return response, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				berrors.BknBackend_RiskType_InternalError).
				WithErrorDetails(err.Error())
		}

		if len(datasetResp.Entries) == 0 {
			break
		}

		for _, entry := range datasetResp.Entries {
			jsonByte, err := json.Marshal(entry)
			if err != nil {
				return response, rest.NewHTTPError(ctx, http.StatusBadRequest,
					berrors.BknBackend_InternalError_MarshalDataFailed).
					WithErrorDetails(fmt.Sprintf("failed to Marshal dataset entry, %s", err.Error()))
			}
			var riskType interfaces.RiskType
			err = json.Unmarshal(jsonByte, &riskType)
			if err != nil {
				return response, rest.NewHTTPError(ctx, http.StatusBadRequest,
					berrors.BknBackend_InternalError_UnMarshalDataFailed).
					WithErrorDetails(fmt.Sprintf("failed to Unmarshal dataset entry to RiskType, %s", err.Error()))
			}

			if scoreVal, ok := entry["_score"]; ok {
				if scoreFloat, ok := scoreVal.(float64); ok {
					score := float64(scoreFloat)
					riskType.Score = &score
				}
			}
			riskType.Vector = nil
			riskTypes = append(riskTypes, &riskType)

			if query.Limit > 0 && len(riskTypes) >= query.Limit {
				break
			}
		}

		query.SearchAfter = datasetResp.SearchAfter

		if (query.Limit > 0 && len(riskTypes) >= query.Limit) || len(datasetResp.Entries) < limit {
			break
		}

		offset += limit
	}

	response.Entries = riskTypes
	response.SearchAfter = query.SearchAfter
	span.SetStatus(codes.Ok, "")
	return response, nil
}

func (rts *riskTypeService) GetAllRiskTypesByKnID(ctx context.Context, knID string, branch string) ([]*interfaces.RiskType, error) {
	return rts.rta.GetAllRiskTypesByKnID(ctx, knID, branch)
}

func (rts *riskTypeService) DeleteRiskTypesByKnID(ctx context.Context, tx *sql.Tx, knID string, branch string) error {
	_, err := rts.rta.DeleteRiskTypesByKnID(ctx, tx, knID, branch)
	return err
}

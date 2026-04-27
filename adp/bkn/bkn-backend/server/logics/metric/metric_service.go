// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package metric

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
)

var (
	metricServiceOnce sync.Once
	metricServiceInst interfaces.MetricService
)

type metricService struct {
	appSetting *common.AppSetting
	db         *sql.DB
	ma         interfaces.MetricAccess
	ps         interfaces.PermissionService
	uma        interfaces.UserMgmtService
	vba        interfaces.VegaBackendAccess
	mfa        interfaces.ModelFactoryAccess
	ots        interfaces.ObjectTypeService
}

func NewMetricService(appSetting *common.AppSetting) interfaces.MetricService {
	metricServiceOnce.Do(func() {
		metricServiceInst = &metricService{
			appSetting: appSetting,
			db:         logics.DB,
			ma:         logics.MA,
			ps:         permission.NewPermissionService(appSetting),
			uma:        logics.UMA,
			vba:        logics.VBA,
			mfa:        logics.MFA,
			ots:        object_type.NewObjectTypeService(appSetting),
		}
	})
	return metricServiceInst
}

func buildMetricBKNRawContent(def *interfaces.MetricDefinition) string {
	cf, _ := json.Marshal(def.CalculationFormula)
	return strings.Join([]string{def.Name, def.Comment, string(cf)}, "\n")
}

func (ms *metricService) InsertDatasetData(ctx context.Context, metrics []*interfaces.MetricDefinition) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "metric index write")
	defer span.End()

	if len(metrics) == 0 {
		return nil
	}

	if ms.appSetting.ServerSetting.DefaultSmallModelEnabled && ms.mfa != nil {
		words := make([]string, 0, len(metrics))
		for _, m := range metrics {
			m.BKNRawContent = buildMetricBKNRawContent(m)
			words = append(words, m.BKNRawContent)
		}
		dftModel, err := ms.mfa.GetDefaultModel(ctx)
		if err != nil {
			logger.Errorf("GetDefaultModel error: %s", err.Error())
			span.SetStatus(codes.Error, "获取默认模型失败")
			return err
		}
		vectors, err := ms.mfa.GetVector(ctx, dftModel, words)
		if err != nil {
			logger.Errorf("GetVector error: %s", err.Error())
			span.SetStatus(codes.Error, "获取指标向量失败")
			return err
		}
		if len(vectors) != len(metrics) {
			return fmt.Errorf("GetVector: expect %d vectors, got %d", len(metrics), len(vectors))
		}
		for i := range metrics {
			metrics[i].Vector = vectors[i].Vector
		}
	}

	documents := make([]map[string]any, 0, len(metrics))
	for _, def := range metrics {
		if def.BKNRawContent == "" {
			def.BKNRawContent = buildMetricBKNRawContent(def)
		}
		docid := interfaces.GenerateConceptDocuemtnID(def.KnID, interfaces.MODULE_TYPE_METRIC, def.ID, def.Branch)
		def.ModuleType = interfaces.MODULE_TYPE_METRIC

		docBytes, err := sonic.Marshal(def)
		if err != nil {
			return err
		}
		var doc map[string]any
		if err := sonic.Unmarshal(docBytes, &doc); err != nil {
			return err
		}
		doc["_id"] = docid
		documents = append(documents, doc)
	}

	if err := ms.vba.WriteDatasetDocuments(ctx, interfaces.BKN_DATASET_ID, documents); err != nil {
		logger.Errorf("WriteDatasetDocuments error: %s", err.Error())
		span.SetStatus(codes.Error, "指标概念索引写入失败")
		return err
	}
	return nil
}

func (ms *metricService) deleteDatasetDocs(ctx context.Context, knID string, branch string, metricIDs []string) {
	for _, id := range metricIDs {
		docid := interfaces.GenerateConceptDocuemtnID(knID, interfaces.MODULE_TYPE_METRIC, id, branch)
		if err := ms.vba.DeleteDatasetDocumentByID(ctx, interfaces.BKN_DATASET_ID, docid); err != nil {
			logger.Errorf("DeleteDatasetDocumentByID metric %s: %v", id, err)
		}
	}
}

func (ms *metricService) CreateMetrics(ctx context.Context, tx *sql.Tx, entries []*interfaces.MetricDefinition, strictMode bool, importMode string) (ids []string, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "CreateMetrics")
	defer span.End()

	if len(entries) == 0 {
		return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_InvalidParameter_RequestBody).
			WithErrorDetails("No metric entries were passed in")
	}

	err = ms.ps.CheckPermission(ctx, interfaces.PermissionResource{
		Type: interfaces.RESOURCE_TYPE_KN,
		ID:   entries[0].KnID,
	}, []string{interfaces.OPERATION_TYPE_MODIFY})
	if err != nil {
		return nil, err
	}

	// 0. 开始事务
	if tx == nil {
		tx, err = ms.db.Begin()
		if err != nil {
			logger.Errorf("Begin transaction error: %s", err.Error())
			span.SetStatus(codes.Error, "事务开启失败")

			return []string{}, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				berrors.BknBackend_Metric_InternalError_BeginTransactionFailed).
				WithErrorDetails(err.Error())
		}
		// 0.1 异常时
		defer func() {
			switch err {
			case nil:
				// 提交事务
				err = tx.Commit()
				if err != nil {
					logger.Errorf("CreateMetrics Transaction Commit Failed:%v", err)
					span.SetStatus(codes.Error, "提交事务失败")
					return
				}
				logger.Infof("CreateMetrics Transaction Commit Success")
			default:
				rollbackErr := tx.Rollback()
				if rollbackErr != nil {
					logger.Errorf("CreateMetrics Transaction Rollback Error:%v", rollbackErr)
					span.SetStatus(codes.Error, "事务回滚失败")
				}
			}
		}()
	}

	currentTime := time.Now().UnixMilli()
	var accountInfo interfaces.AccountInfo
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}

	for _, m := range entries {
		if strings.TrimSpace(m.ID) == "" {
			m.ID = xid.New().String()
		} else {
			m.ID = strings.TrimSpace(m.ID)
		}
		m.Creator = accountInfo
		m.Updater = accountInfo
		m.CreateTime = currentTime
		m.UpdateTime = currentTime
		m.ModuleType = interfaces.MODULE_TYPE_METRIC

		if strictMode {
			if err := ms.validateMetricStrictExternalDeps(ctx, tx, m); err != nil {
				return []string{}, err
			}
		}

		m.BKNRawContent = buildMetricBKNRawContent(m)
	}

	var creates []*interfaces.MetricDefinition
	creates, err = ms.handleMetricImportMode(ctx, importMode, entries)
	if err != nil {
		return nil, err
	}
	if len(creates) == 0 {
		return []string{}, nil
	}

	ids = make([]string, 0, len(creates))
	for _, def := range creates {
		err = ms.ma.CreateMetric(ctx, tx, def)
		if err != nil {
			logger.Errorf("CreateMetric error: %s", err.Error())
			span.SetStatus(codes.Error, "创建指标失败")
			return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, berrors.BknBackend_Metric_InternalError).WithErrorDetails(err.Error())
		}
		ids = append(ids, def.ID)
	}

	err = ms.InsertDatasetData(ctx, creates)
	if err != nil {
		logger.Errorf("InsertDatasetData error: %s", err.Error())
		span.SetStatus(codes.Error, "指标概念索引写入失败")
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, berrors.BknBackend_Metric_InternalError).WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return ids, nil
}

// handleMetricImportMode 按 import_mode 过滤待创建的指标（与 handleObjectTypeImportMode 对齐；overwrite 暂未支持）。
func (ms *metricService) handleMetricImportMode(ctx context.Context, mode string, metrics []*interfaces.MetricDefinition) ([]*interfaces.MetricDefinition, error) {
	if mode == interfaces.ImportMode_Overwrite {
		return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_Metric_InvalidParameter).
			WithErrorDetails("import_mode overwrite is not supported for metrics yet")
	}

	creates := make([]*interfaces.MetricDefinition, 0, len(metrics))
	for _, m := range metrics {
		knID, branch := m.KnID, m.Branch
		id := strings.TrimSpace(m.ID)

		var idExist, nameExist bool
		var existNameByID, existIDByName string
		var qerr error
		if id != "" {
			existNameByID, idExist, qerr = ms.ma.CheckMetricExistByID(ctx, knID, branch, id)
			if qerr != nil {
				return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
					berrors.BknBackend_Metric_InternalError_CheckMetricIfExistFailed).WithErrorDetails(qerr.Error())
			}
		}
		existIDByName, nameExist, qerr = ms.ma.CheckMetricExistByName(ctx, knID, branch, m.Name)
		if qerr != nil {
			return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				berrors.BknBackend_Metric_InternalError_CheckMetricIfExistFailed).WithErrorDetails(qerr.Error())
		}

		if idExist || nameExist {
			switch mode {
			case interfaces.ImportMode_Normal:
				if idExist {
					return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_Metric_InvalidParameter).
						WithErrorDetails(fmt.Sprintf("metric id '%s' already exists (name in DB: %s)", id, existNameByID))
				}
				if nameExist && existIDByName != id {
					return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_Metric_Duplicated_Name).
						WithErrorDetails(fmt.Sprintf("metric name '%s' already exists", m.Name))
				}
			case interfaces.ImportMode_Ignore:
				continue
			default:
				return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_InvalidParameter_ImportMode).
					WithErrorDetails("unsupported import_mode")
			}
		}

		creates = append(creates, m)
	}
	return creates, nil
}

func (ms *metricService) ListMetrics(ctx context.Context, query interfaces.MetricsListQueryParams) (*interfaces.MetricsList, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "ListMetrics")
	defer span.End()

	err := ms.ps.CheckPermission(ctx, interfaces.PermissionResource{
		Type: interfaces.RESOURCE_TYPE_KN,
		ID:   query.KNID,
	}, []string{interfaces.OPERATION_TYPE_VIEW_DETAIL})
	if err != nil {
		return nil, err
	}

	list, err := ms.ma.ListMetrics(ctx, query)
	if err != nil {
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, berrors.BknBackend_Metric_InternalError).WithErrorDetails(err.Error())
	}
	total, err := ms.ma.GetMetricsTotal(ctx, query)
	if err != nil {
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, berrors.BknBackend_Metric_InternalError).WithErrorDetails(err.Error())
	}

	if query.Limit != -1 {
		if query.Offset < 0 || int(query.Offset) >= len(list) {
			return &interfaces.MetricsList{Entries: []*interfaces.MetricDefinition{}, TotalCount: int64(total)}, nil
		}
		end := int(query.Offset) + query.Limit
		if end > len(list) {
			end = len(list)
		}
		list = list[query.Offset:end]
	}

	if len(list) > 0 && ms.uma != nil {
		infos := make([]*interfaces.AccountInfo, 0, len(list)*2)
		for _, m := range list {
			infos = append(infos, &m.Creator, &m.Updater)
		}
		_ = ms.uma.GetAccountNames(ctx, infos)
	}

	span.SetStatus(codes.Ok, "")
	return &interfaces.MetricsList{Entries: list, TotalCount: int64(total)}, nil
}

func (ms *metricService) GetMetricByID(ctx context.Context, knID, branch, metricID string) (*interfaces.MetricDefinition, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "GetMetricByID")
	defer span.End()

	err := ms.ps.CheckPermission(ctx, interfaces.PermissionResource{
		Type: interfaces.RESOURCE_TYPE_KN,
		ID:   knID,
	}, []string{interfaces.OPERATION_TYPE_VIEW_DETAIL})
	if err != nil {
		return nil, err
	}

	def, err := ms.ma.GetMetricByID(ctx, knID, branch, metricID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, rest.NewHTTPError(ctx, http.StatusNotFound, berrors.BknBackend_Metric_NotFound)
		}
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, berrors.BknBackend_Metric_InternalError).WithErrorDetails(err.Error())
	}
	span.SetStatus(codes.Ok, "")
	return def, nil
}

func (ms *metricService) GetMetricsByIDs(ctx context.Context, knID, branch string, metricIDs []string) ([]*interfaces.MetricDefinition, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "GetMetricsByIDs")
	defer span.End()

	metricIDs = common.DuplicateSlice(metricIDs)
	list, err := ms.ma.GetMetricsByIDs(ctx, knID, branch, metricIDs)
	if err != nil {
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_Metric_InternalError_GetMetricsByIDsFailed).WithErrorDetails(err.Error())
	}
	span.SetStatus(codes.Ok, "")
	return list, nil
}

func (ms *metricService) CheckMetricExistByID(ctx context.Context, knID, branch, metricID string) (string, bool, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("校验指标[%s]的存在性", metricID))
	defer span.End()

	name, exist, err := ms.ma.CheckMetricExistByID(ctx, knID, branch, metricID)
	if err != nil {
		logger.Errorf("CheckMetricExistByID error: %s", err.Error())
		span.SetStatus(codes.Error, "check metric existence by id failed")
		return "", exist, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_Metric_InternalError_CheckMetricIfExistFailed).
			WithErrorDetails(err.Error())
	}
	span.SetStatus(codes.Ok, "")
	return name, exist, nil
}

func (ms *metricService) CheckMetricExistByName(ctx context.Context, knID, branch, name string) (string, bool, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, fmt.Sprintf("校验指标名称[%s]的存在性", name))
	defer span.End()

	id, exist, err := ms.ma.CheckMetricExistByName(ctx, knID, branch, name)
	if err != nil {
		logger.Errorf("CheckMetricExistByName error: %s", err.Error())
		span.SetStatus(codes.Error, "check metric existence by name failed")
		return "", exist, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			berrors.BknBackend_Metric_InternalError_CheckMetricIfExistFailed).
			WithErrorDetails(err.Error())
	}
	span.SetStatus(codes.Ok, "")
	return id, exist, nil
}

func (ms *metricService) ValidateMetrics(ctx context.Context, entries []*interfaces.MetricDefinition, strictMode bool, importMode string) error {
	_ = importMode
	if len(entries) == 0 {
		return nil
	}
	for _, e := range entries {
		if strictMode {
			if err := ms.validateMetricStrictExternalDeps(ctx, nil, e); err != nil {
				return err
			}
		}
	}
	return nil
}

func (ms *metricService) UpdateMetric(ctx context.Context, tx *sql.Tx, req *interfaces.MetricDefinition, strictMode bool) (err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "UpdateMetric")
	defer span.End()

	knID := strings.TrimSpace(req.KnID)
	branch := strings.TrimSpace(req.Branch)
	metricID := strings.TrimSpace(req.ID)
	if knID == "" || branch == "" || metricID == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_Metric_InvalidParameter).
			WithErrorDetails("missing kn_id, branch or id in metric definition")
	}

	err = ms.ps.CheckPermission(ctx, interfaces.PermissionResource{
		Type: interfaces.RESOURCE_TYPE_KN,
		ID:   knID,
	}, []string{interfaces.OPERATION_TYPE_MODIFY})
	if err != nil {
		return err
	}

	if tx == nil {
		// 0. 开始事务
		tx, err = ms.db.Begin()
		if err != nil {
			logger.Errorf("Begin transaction error: %s", err.Error())
			span.SetStatus(codes.Error, "事务开启失败")

			return rest.NewHTTPError(ctx, http.StatusInternalServerError,
				berrors.BknBackend_Metric_InternalError_BeginTransactionFailed).
				WithErrorDetails(err.Error())
		}
		// 0.1 异常时
		defer func() {
			switch err {
			case nil:
				// 提交事务
				err = tx.Commit()
				if err != nil {
					logger.Errorf("UpdateMetric Transaction Commit Failed:%v", err)
					span.SetStatus(codes.Error, "提交事务失败")
				}
				logger.Infof("UpdateMetric Transaction Commit Success:%v", metricID)
			default:
				rollbackErr := tx.Rollback()
				if rollbackErr != nil {
					logger.Errorf("UpdateMetric Transaction Rollback Error:%v", rollbackErr)
					span.SetStatus(codes.Error, "事务回滚失败")
				}
			}
		}()
	}

	if strictMode {
		if err := ms.validateMetricStrictExternalDeps(ctx, tx, req); err != nil {
			return err
		}
	}

	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		req.Updater = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}
	req.UpdateTime = time.Now().UnixMilli()

	err = ms.ma.UpdateMetric(ctx, tx, req)
	if err != nil {
		logger.Errorf("UpdateMetric error: %s", err.Error())
		span.SetStatus(codes.Error, "修改指标失败")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, berrors.BknBackend_Metric_InternalError).WithErrorDetails(err.Error())
	}

	err = ms.InsertDatasetData(ctx, []*interfaces.MetricDefinition{req})
	if err != nil {
		logger.Errorf("InsertDatasetData after update: %s", err.Error())
		span.SetStatus(codes.Error, "指标概念索引写入失败")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, berrors.BknBackend_Metric_InternalError).WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

func (ms *metricService) DeleteMetricsByIDs(ctx context.Context, tx *sql.Tx, knID, branch string, metricIDs []string) (err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "DeleteMetricsByIDs")
	defer span.End()

	if len(metricIDs) == 0 {
		return nil
	}

	err = ms.ps.CheckPermission(ctx, interfaces.PermissionResource{
		Type: interfaces.RESOURCE_TYPE_KN,
		ID:   knID,
	}, []string{interfaces.OPERATION_TYPE_MODIFY})
	if err != nil {
		return err
	}

	if tx == nil {
		// 0. 开始事务
		tx, err = ms.db.Begin()
		if err != nil {
			logger.Errorf("Begin transaction error: %s", err.Error())
			span.SetStatus(codes.Error, "事务开启失败")

			return rest.NewHTTPError(ctx, http.StatusInternalServerError,
				berrors.BknBackend_Metric_InternalError_BeginTransactionFailed).
				WithErrorDetails(err.Error())
		}
		// 0.1 异常时
		defer func() {
			switch err {
			case nil:
				// 提交事务
				err = tx.Commit()
				if err != nil {
					logger.Errorf("DeleteMetricsByIDs Transaction Commit Failed:%v", err)
					span.SetStatus(codes.Error, "提交事务失败")
				}
				logger.Infof("DeleteMetricsByIDs Transaction Commit Success: kn_id:%s,metric_ids:%v", knID, metricIDs)
			default:
				rollbackErr := tx.Rollback()
				if rollbackErr != nil {
					logger.Errorf("DeleteMetricsByIDs Transaction Rollback Error:%v", rollbackErr)
					span.SetStatus(codes.Error, "事务回滚失败")
				}
			}
		}()
	}

	dErr := ms.ma.DeleteMetricsByIDs(ctx, tx, knID, branch, metricIDs)
	if dErr != nil {
		logger.Errorf("DeleteMetricsByIDs error: %s", dErr.Error())
		span.SetStatus(codes.Error, "删除指标失败")
		err = rest.NewHTTPError(ctx, http.StatusInternalServerError, berrors.BknBackend_Metric_InternalError).WithErrorDetails(dErr.Error())
		return err
	}

	ms.deleteDatasetDocs(ctx, knID, branch, metricIDs)
	return nil
}

// DeleteMetricsByKnID 内部接口，不校验权限；tx 必须传（与 DeleteActionTypesByKnID 一致）。
func (ms *metricService) DeleteMetricsByKnID(ctx context.Context, tx *sql.Tx, knID, branch string) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "DeleteMetricsByKnID")
	defer span.End()

	if tx == nil {
		logger.Errorf("DeleteMetricsByKnID: missing transaction")
		span.SetStatus(codes.Error, "missing transaction")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, berrors.BknBackend_Metric_InternalError).
			WithErrorDetails("missing transaction")
	}

	ids, err := ms.ma.GetMetricIDsByKnID(ctx, knID, branch)
	if err != nil {
		logger.Errorf("GetMetricIDsByKnID error: %v", err)
		span.SetStatus(codes.Error, "list metric ids failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, berrors.BknBackend_Metric_InternalError).WithErrorDetails(err.Error())
	}

	rowsAff, err := ms.ma.DeleteMetricsByKnID(ctx, tx, knID, branch)
	if err != nil {
		logger.Errorf("DeleteMetricsByKnID access error: %v", err)
		span.SetStatus(codes.Error, "delete metrics by kn failed")
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, berrors.BknBackend_Metric_InternalError).WithErrorDetails(err.Error())
	}

	ms.deleteDatasetDocs(ctx, knID, branch, ids)
	logger.Infof("DeleteMetricsByKnID success, kn_id=%s branch=%s rows=%d metric_docs=%d", knID, branch, rowsAff, len(ids))
	span.SetStatus(codes.Ok, "")
	return nil
}

func (ms *metricService) SearchMetrics(ctx context.Context, query *interfaces.ConceptsQuery) (interfaces.MetricSearchResult, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "SearchMetrics")
	defer span.End()

	response := interfaces.MetricSearchResult{
		Type:   interfaces.MODULE_TYPE_METRIC,
		Groups: []any{},
	}

	err := ms.ps.CheckPermission(ctx, interfaces.PermissionResource{
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
				if !ms.appSetting.ServerSetting.DefaultSmallModelEnabled {
					err = errors.New(cond.DEFAULT_SMALL_MODEL_ENABLED_FALSE_ERROR)
					span.SetStatus(codes.Error, err.Error())
					return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
						berrors.BknBackend_Metric_InternalError).
						WithErrorDetails(err.Error())
				}
				dftModel, err := ms.mfa.GetDefaultModel(ctx)
				if err != nil {
					return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
						berrors.BknBackend_Metric_InternalError).
						WithErrorDetails(err.Error())
				}
				result, err := ms.mfa.GetVector(ctx, dftModel, []string{word})
				if err != nil {
					return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
						berrors.BknBackend_Metric_InternalError).
						WithErrorDetails(err.Error())
				}
				return result, nil
			})
		if err != nil {
			return response, rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_InvalidParameter_Condition).
				WithErrorDetails(fmt.Sprintf("failed to convert condition to filter condition: %s", err.Error()))
		}
	}

	if query.NeedTotal {
		params := &interfaces.ResourceDataQueryParams{
			FilterCondition: filterCondition,
			Offset:          0,
			Limit:           1,
			NeedTotal:       true,
		}
		datasetResp, err := ms.vba.QueryResourceData(ctx, interfaces.BKN_DATASET_ID, params)
		if err != nil {
			return response, rest.NewHTTPError(ctx, http.StatusInternalServerError, berrors.BknBackend_Metric_InternalError).WithErrorDetails(err.Error())
		}
		response.TotalCount = datasetResp.TotalCount
	}

	offset := 0
	limit := query.Limit
	if limit == 0 {
		limit = interfaces.SearchAfter_Limit
	}

	var entries []*interfaces.MetricDefinition

	for {
		params := &interfaces.ResourceDataQueryParams{
			FilterCondition: filterCondition,
			Offset:          offset,
			Limit:           limit,
			NeedTotal:       false,
			Sort:            query.Sort,
		}
		datasetResp, err := ms.vba.QueryResourceData(ctx, interfaces.BKN_DATASET_ID, params)
		if err != nil {
			return response, rest.NewHTTPError(ctx, http.StatusInternalServerError, berrors.BknBackend_Metric_InternalError).WithErrorDetails(err.Error())
		}
		if len(datasetResp.Entries) == 0 {
			break
		}
		for _, entry := range datasetResp.Entries {
			jsonByte, err := json.Marshal(entry)
			if err != nil {
				return response, rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_InternalError_MarshalDataFailed).WithErrorDetails(err.Error())
			}
			var m interfaces.MetricDefinition
			if err := json.Unmarshal(jsonByte, &m); err != nil {
				return response, rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_InternalError_UnMarshalDataFailed).WithErrorDetails(err.Error())
			}
			if scoreVal, ok := entry["_score"]; ok {
				if scoreFloat, ok := scoreVal.(float64); ok {
					s := float64(scoreFloat)
					m.Score = &s
				}
			}
			m.Vector = nil
			entries = append(entries, &m)
			if query.Limit > 0 && len(entries) >= query.Limit {
				break
			}
		}
		query.SearchAfter = datasetResp.SearchAfter
		if (query.Limit > 0 && len(entries) >= query.Limit) || len(datasetResp.Entries) < limit {
			break
		}
		offset += limit
	}

	response.Entries = entries
	response.SearchAfter = query.SearchAfter
	span.SetStatus(codes.Ok, "")
	return response, nil
}

func (ms *metricService) validateMetricStrictExternalDeps(ctx context.Context, tx *sql.Tx, metric *interfaces.MetricDefinition) error {

	// 校验统计主体
	scopeRef := strings.TrimSpace(metric.ScopeRef)
	if scopeRef == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_Metric_InvalidParameter).
			WithErrorDetails("scope_ref is required (service dependency check)")
	}

	ot, err := ms.ots.GetObjectTypeByID(ctx, tx, metric.KnID, metric.Branch, scopeRef)
	if err != nil {
		return err
	}
	if ot == nil {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_Metric_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("metric[%s]'s scope_ref[%s] object type not found", metric.ID, scopeRef))
	}

	ds := ot.DataSource
	if ds == nil || strings.TrimSpace(ds.ID) == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_Metric_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("metric[%s]'s object type data_source resource id is required", metric.ID))
	}
	if ds.Type == interfaces.DATA_SOURCE_TYPE_DATA_VIEW {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_Metric_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("metric[%s] is not supported for object types backed by data_view; use resource-backed object types", metric.ID))
	}
	if ds.Type != interfaces.DATA_SOURCE_TYPE_RESOURCE {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_Metric_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("only resource-backed object types can define metric[%s]", metric.ID))
	}

	propertyMap := map[string]*interfaces.DataProperty{}
	for _, prop := range ot.DataProperties {
		propertyMap[prop.Name] = prop
	}

	// 校验指标定义中使用的各个属性是属于统计主体对象类的
	// 	1. 时间维度非空时，其property需是属于对象类的属性
	if metric.TimeDimension != nil {
		if p := strings.TrimSpace(metric.TimeDimension.Property); p != "" {
			if _, ok := propertyMap[p]; !ok {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_Metric_InvalidParameter).
					WithErrorDetails(fmt.Sprintf("metric[%s]'s time_dimension.property[%s] is not a data property of the scope object type[%s]",
						metric.ID, p, scopeRef))
			}
		}
	}

	// 2. 分析维度非空时，其property需是属于对象类的属性
	for i, ad := range metric.AnalysisDimensions {
		if n := strings.TrimSpace(ad.Name); n != "" {
			if _, ok := propertyMap[n]; !ok {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_Metric_InvalidParameter).
					WithErrorDetails(fmt.Sprintf("metric[%s]'s analysis_dimensions[%d].name[%s] is not a data property of the scope object type[%s]",
						metric.ID, i, n, scopeRef))
			}
		}
	}

	// 3. 计算公式的 condition 的字段需在对象类中存在
	if metric.CalculationFormula == nil {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_Metric_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("metric[%s]'s calculation_formula is required", metric.ID))
	}

	// 3. calculation_formula.condition：递归校验 CondCfg 中出现的字段（含 and/or/knn、multi_match.fields、叶子条件 field）均须为对象类数据属性
	if metric.CalculationFormula.Condition != nil {
		if err := validateConditionFieldsReferenceObjectType(ctx, metric.CalculationFormula.Condition, propertyMap, metric.ID); err != nil {
			return err
		}
	}

	// 4. aggregation使用的聚合属性需在对象类中存在
	if metric.CalculationFormula.Aggregation.Aggr != "" {
		if _, ok := propertyMap[metric.CalculationFormula.Aggregation.Property]; !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_Metric_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("metric[%s]'s calculation_formula.aggregation.property[%s] is not a data property of the scope object type[%s]",
					metric.ID, metric.CalculationFormula.Aggregation.Property, scopeRef))
		}
	}

	// 5. 分组字段使用的属性需在对象类中存在
	if metric.CalculationFormula.GroupBy != nil {
		for i, g := range metric.CalculationFormula.GroupBy {
			if _, ok := propertyMap[g.Property]; !ok {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_Metric_InvalidParameter).
					WithErrorDetails(fmt.Sprintf("metric[%s]'s calculation_formula.group_by[%d].property[%s] is not a data property of the scope object type[%s]",
						metric.ID, i, g.Property, scopeRef))
			}
		}
	}
	// 6. 排序字段属于分组字段
	if metric.CalculationFormula.OrderBy != nil {
		for i, o := range metric.CalculationFormula.OrderBy {
			if _, ok := propertyMap[o.Property]; !ok {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_Metric_InvalidParameter).
					WithErrorDetails(fmt.Sprintf("metric[%s]'s calculation_formula.order_by[%d].property[%s] is not a data property of the scope object type[%s]",
						metric.ID, i, o.Property, scopeRef))
			}
		}
	}

	// 7.having过滤只能是__value字段
	if metric.CalculationFormula.Having != nil {
		if metric.CalculationFormula.Having.Field != interfaces.MetricHavingFieldValue {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_Metric_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("metric[%s]'s calculation_formula.having.field must be %q, got %q", metric.ID, interfaces.MetricHavingFieldValue, metric.CalculationFormula.Having.Field))
		}
	}
	return nil
}

func validateConditionFieldsReferenceObjectType(ctx context.Context, cfg *cond.CondCfg, propertyMap map[string]*interfaces.DataProperty, metricID string) error {
	if cfg == nil {
		return nil
	}

	switch cfg.Operation {
	case cond.OperationAnd, cond.OperationOr:
		for _, s := range cfg.SubConds {
			if err := validateConditionFieldsReferenceObjectType(ctx, s, propertyMap, metricID); err != nil {
				return err
			}
		}
		return nil
	default:
		n := strings.TrimSpace(cfg.Field)
		if n == "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_Metric_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("metric[%s]: condition property is required", metricID))
		}

		if _, ok := propertyMap[n]; !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_Metric_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("metric[%s]: condition property [%s] is not a data property of the scope object type", metricID, n))
		}
		return nil
	}
}

// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-go-lib/audit"
	"github.com/kweaver-ai/kweaver-go-lib/hydra"
	"github.com/kweaver-ai/kweaver-go-lib/otel/oteltrace"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	attr "go.opentelemetry.io/otel/attribute"

	"bkn-backend/common"
	"bkn-backend/common/visitor"
	berrors "bkn-backend/errors"
	"bkn-backend/interfaces"
)

func (r *restHandler) HandleMetricGetOverrideByIn(c *gin.Context) {
	switch c.GetHeader(interfaces.HTTP_HEADER_METHOD_OVERRIDE) {
	case "", http.MethodPost:
		r.CreateMetricsByIn(c)
	case http.MethodGet:
		r.SearchMetricsByIn(c)
	default:
		httpErr := rest.NewHTTPError(rest.GetLanguageCtx(c), http.StatusBadRequest,
			berrors.BknBackend_InvalidParameter_OverrideMethod)
		rest.ReplyError(c, httpErr)
	}
}

func (r *restHandler) HandleMetricGetOverrideByEx(c *gin.Context) {
	switch c.GetHeader(interfaces.HTTP_HEADER_METHOD_OVERRIDE) {
	case "", http.MethodPost:
		r.CreateMetricsByEx(c)
	case http.MethodGet:
		r.SearchMetricsByEx(c)
	default:
		httpErr := rest.NewHTTPError(rest.GetLanguageCtx(c), http.StatusBadRequest,
			berrors.BknBackend_InvalidParameter_OverrideMethod)
		rest.ReplyError(c, httpErr)
	}
}

func (r *restHandler) CreateMetricsByIn(c *gin.Context) {
	v := visitor.GenerateVisitor(c)
	r.CreateMetrics(c, v)
}

func (r *restHandler) CreateMetricsByEx(c *gin.Context) {
	visitor, err := r.verifyOAuth(rest.GetLanguageCtx(c), c)
	if err != nil {
		return
	}
	r.CreateMetrics(c, visitor)
}

func (r *restHandler) CreateMetrics(c *gin.Context, vis hydra.Visitor) {
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	accountInfo := interfaces.AccountInfo{ID: vis.ID, Type: string(vis.Type)}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)
	oteltrace.AddHttpAttrs4API(span, oteltrace.GetAttrsByGinCtx(c))

	mode := c.DefaultQuery(interfaces.QueryParam_ImportMode, interfaces.ImportMode_Normal)
	if httpErr := validateImportMode(ctx, mode); httpErr != nil {
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	strictModeStr := c.DefaultQuery(interfaces.QueryParam_StrictMode, "true")
	strictMode, err := strconv.ParseBool(strictModeStr)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_Metric_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("Invalid strict_mode parameter: %s", strictModeStr))
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	knID := c.Param("kn_id")
	branch := c.DefaultQuery("branch", interfaces.MAIN_BRANCH)
	span.SetAttributes(attr.Key("kn_id").String(knID), attr.Key("branch").String(branch))

	_, exist, err := r.kns.CheckKNExistByID(ctx, knID, branch)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if !exist {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, berrors.BknBackend_KnowledgeNetwork_NotFound)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	var body struct {
		Entries []*interfaces.MetricDefinition `json:"entries"`
	}
	if err = c.ShouldBindJSON(&body); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_Metric_InvalidParameter).
			WithErrorDetails("Binding Parameter Failed: " + err.Error())
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if len(body.Entries) == 0 {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_InvalidParameter_RequestBody).
			WithErrorDetails("No metric was passed in")
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	metrics := body.Entries

	if err := ValidateMetricRequests(ctx, metrics, strictMode); err != nil {
		var httpErr *rest.HTTPError
		if errors.As(err, &httpErr) {
			oteltrace.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
		oteltrace.AddHttpAttrs4Error(span, http.StatusInternalServerError, "", err.Error())
		rest.ReplyError(c, err)
		return
	}

	// request来的actionTypes的branch都用url里的branch
	for i := range metrics {
		metrics[i].KnID = knID
		metrics[i].Branch = branch
	}

	ids, err := r.ms.CreateMetrics(ctx, nil, metrics, strictMode, mode)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	for _, id := range ids {
		audit.NewInfoLog(audit.OPERATION, audit.CREATE, audit.TransforOperator(vis),
			interfaces.GenerateMetricAuditObject(id, ""), "")
	}

	result := []any{}
	for _, id := range ids {
		result = append(result, map[string]any{"id": id})
	}
	oteltrace.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusCreated, result)
}

func (r *restHandler) ValidateMetricsByIn(c *gin.Context) {
	r.ValidateMetrics(c, visitor.GenerateVisitor(c))
}

func (r *restHandler) ValidateMetricsByEx(c *gin.Context) {
	vis, err := r.verifyOAuth(rest.GetLanguageCtx(c), c)
	if err != nil {
		return
	}
	r.ValidateMetrics(c, vis)
}

func (r *restHandler) ValidateMetrics(c *gin.Context, vis hydra.Visitor) {
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	accountInfo := interfaces.AccountInfo{ID: vis.ID, Type: string(vis.Type)}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)
	oteltrace.AddHttpAttrs4API(span, oteltrace.GetAttrsByGinCtx(c))

	mode := c.DefaultQuery(interfaces.QueryParam_ImportMode, interfaces.ImportMode_Normal)
	if httpErr := validateImportMode(ctx, mode); httpErr != nil {
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	strictModeStr := c.DefaultQuery(interfaces.QueryParam_StrictMode, "true")
	strictMode, err := strconv.ParseBool(strictModeStr)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_Metric_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("Invalid strict_mode parameter: %s", strictModeStr))
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	knID := c.Param("kn_id")
	branch := c.DefaultQuery("branch", interfaces.MAIN_BRANCH)
	span.SetAttributes(attr.Key("kn_id").String(knID), attr.Key("branch").String(branch))

	_, exist, err := r.kns.CheckKNExistByID(ctx, knID, branch)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if !exist {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, berrors.BknBackend_KnowledgeNetwork_NotFound)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	var body struct {
		Entries []*interfaces.MetricDefinition `json:"entries"`
	}
	if err = c.ShouldBindJSON(&body); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_Metric_InvalidParameter).
			WithErrorDetails("Binding Parameter Failed: " + err.Error())
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if len(body.Entries) == 0 {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_InvalidParameter_RequestBody).
			WithErrorDetails("No metric was passed in")
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	metrics := body.Entries

	// request来的actionTypes的branch都用url里的branch
	for i := range metrics {
		metrics[i].KnID = knID
		metrics[i].Branch = branch
	}

	if err := ValidateMetricRequests(ctx, metrics, strictMode); err != nil {
		var httpErr *rest.HTTPError
		if errors.As(err, &httpErr) {
			oteltrace.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
		oteltrace.AddHttpAttrs4Error(span, http.StatusInternalServerError, "", err.Error())
		rest.ReplyError(c, err)
		return
	}
	if err := r.ms.ValidateMetrics(ctx, metrics, strictMode, mode, nil); err != nil {
		httpErr := err.(*rest.HTTPError)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	oteltrace.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, nil)
}

func (r *restHandler) ListMetricsByIn(c *gin.Context) {
	r.ListMetrics(c, visitor.GenerateVisitor(c))
}

func (r *restHandler) ListMetricsByEx(c *gin.Context) {
	vis, err := r.verifyOAuth(rest.GetLanguageCtx(c), c)
	if err != nil {
		return
	}
	r.ListMetrics(c, vis)
}

func (r *restHandler) ListMetrics(c *gin.Context, vis hydra.Visitor) {
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	accountInfo := interfaces.AccountInfo{ID: vis.ID, Type: string(vis.Type)}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)
	oteltrace.AddHttpAttrs4API(span, oteltrace.GetAttrsByGinCtx(c))

	knID := c.Param("kn_id")
	branch := c.DefaultQuery("branch", interfaces.MAIN_BRANCH)
	span.SetAttributes(attr.Key("kn_id").String(knID), attr.Key("branch").String(branch))

	_, exist, err := r.kns.CheckKNExistByID(ctx, knID, branch)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if !exist {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, berrors.BknBackend_KnowledgeNetwork_NotFound)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	namePattern := c.Query("name_pattern")
	tag := strings.Trim(c.Query("tag"), " ")
	offset := c.DefaultQuery("offset", interfaces.DEFAULT_OFFEST)
	limit := c.DefaultQuery("limit", interfaces.DEFAULT_LIMIT)
	sort := c.DefaultQuery("sort", "update_time")
	direction := c.DefaultQuery("direction", interfaces.DESC_DIRECTION)

	pageParam, err := validatePaginationQueryParameters(ctx, offset, limit, sort, direction, interfaces.MetricSort)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	query := interfaces.MetricsListQueryParams{
		PaginationQueryParameters: interfaces.PaginationQueryParameters{
			Offset:    pageParam.Offset,
			Limit:     pageParam.Limit,
			Sort:      pageParam.Sort,
			Direction: pageParam.Direction,
		},
		NamePattern: namePattern,
		Tag:         tag,
		Branch:      branch,
		KNID:        knID,
	}

	list, err := r.ms.ListMetrics(ctx, query)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	oteltrace.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, list)
}

func (r *restHandler) GetMetricsByIDsByEx(c *gin.Context) {
	vis, err := r.verifyOAuth(rest.GetLanguageCtx(c), c)
	if err != nil {
		return
	}
	r.GetMetricsByIDs(c, vis)
}

func (r *restHandler) GetMetricsByIDsByIn(c *gin.Context) {
	r.GetMetricsByIDs(c, visitor.GenerateVisitor(c))
}

func (r *restHandler) GetMetricsByIDs(c *gin.Context, vis hydra.Visitor) {
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	accountInfo := interfaces.AccountInfo{ID: vis.ID, Type: string(vis.Type)}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	knID := c.Param("kn_id")
	metricIDsStr := c.Param("metric_ids")
	branch := c.DefaultQuery("branch", interfaces.MAIN_BRANCH)
	span.SetAttributes(attr.Key("kn_id").String(knID), attr.Key("metric_ids").String(metricIDsStr))

	_, exist, err := r.kns.CheckKNExistByID(ctx, knID, branch)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if !exist {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, berrors.BknBackend_KnowledgeNetwork_NotFound)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	ids := common.StringToStringSlice(metricIDsStr)
	list, err := r.ms.GetMetricsByIDs(ctx, knID, branch, ids)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if len(list) != len(ids) {
		found := make(map[string]bool)
		for _, m := range list {
			found[m.ID] = true
		}
		var missing []string
		for _, id := range ids {
			if !found[id] {
				missing = append(missing, id)
			}
		}
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, berrors.BknBackend_Metric_NotFound).
			WithErrorDetails(fmt.Sprintf("metrics not found: %v", missing))
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	oteltrace.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, map[string]any{"entries": list})
}

func (r *restHandler) UpdateMetricByIn(c *gin.Context) {
	r.UpdateMetric(c, visitor.GenerateVisitor(c))
}

func (r *restHandler) UpdateMetricByEx(c *gin.Context) {
	vis, err := r.verifyOAuth(rest.GetLanguageCtx(c), c)
	if err != nil {
		return
	}
	r.UpdateMetric(c, vis)
}

func (r *restHandler) UpdateMetric(c *gin.Context, vis hydra.Visitor) {
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	accountInfo := interfaces.AccountInfo{ID: vis.ID, Type: string(vis.Type)}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)
	oteltrace.AddHttpAttrs4API(span, oteltrace.GetAttrsByGinCtx(c))

	knID := c.Param("kn_id")
	metricID := c.Param("metric_ids")
	branch := c.DefaultQuery("branch", interfaces.MAIN_BRANCH)
	strictModeStr := c.DefaultQuery(interfaces.QueryParam_StrictMode, "true")
	strictMode, err := strconv.ParseBool(strictModeStr)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_Metric_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("Invalid strict_mode parameter: %s", strictModeStr))
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	_, exist, err := r.kns.CheckKNExistByID(ctx, knID, branch)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if !exist {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, berrors.BknBackend_KnowledgeNetwork_NotFound)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 先按 id 校验指标是否存在（与 UpdateObjectType 中 CheckObjectTypeExistByID 一致；在解析 body 前失败可尽早返回）
	oldMetricName, metricExist, err := r.ms.CheckMetricExistByID(ctx, knID, branch, metricID)
	if err != nil {
		var httpErr *rest.HTTPError
		if errors.As(err, &httpErr) {
			oteltrace.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
		oteltrace.AddHttpAttrs4Error(span, http.StatusInternalServerError, "", err.Error())
		rest.ReplyError(c, err)
		return
	}
	if !metricExist {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound,
			berrors.BknBackend_Metric_NotFound)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	var req interfaces.MetricDefinition
	if err = c.ShouldBindJSON(&req); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_Metric_InvalidParameter).
			WithErrorDetails(err.Error())
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	req.ID = metricID
	req.KnID = knID
	req.Branch = branch

	if err := ValidateMetricRequest(ctx, &req, strictMode); err != nil {
		var httpErr *rest.HTTPError
		if errors.As(err, &httpErr) {
			oteltrace.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
		oteltrace.AddHttpAttrs4Error(span, http.StatusInternalServerError, "", err.Error())
		rest.ReplyError(c, err)
		return
	}

	// 名称变更时校验新名称未被其它指标占用
	newName := strings.TrimSpace(req.Name)
	if newName != "" && newName != strings.TrimSpace(oldMetricName) {
		existID, nameExist, err := r.ms.CheckMetricExistByName(ctx, knID, branch, newName)
		if err != nil {
			httpErr := err.(*rest.HTTPError)
			oteltrace.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
		if nameExist && existID != metricID {
			httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_Metric_Duplicated_Name).
				WithErrorDetails(fmt.Sprintf("metric name %q already exists", newName))
			oteltrace.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
	}

	if err = r.ms.UpdateMetric(ctx, nil, &req, strictMode); err != nil {
		rest.ReplyError(c, err.(*rest.HTTPError))
		return
	}

	audit.NewInfoLog(audit.OPERATION, audit.UPDATE, audit.TransforOperator(vis),
		interfaces.GenerateMetricAuditObject(metricID, ""), "")
	oteltrace.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusNoContent, nil)
}

func (r *restHandler) DeleteMetricsByIDsByEx(c *gin.Context) {
	vis, err := r.verifyOAuth(rest.GetLanguageCtx(c), c)
	if err != nil {
		return
	}
	r.DeleteMetricsByIDs(c, vis)
}

func (r *restHandler) DeleteMetricsByIDsByIn(c *gin.Context) {
	r.DeleteMetricsByIDs(c, visitor.GenerateVisitor(c))
}

func (r *restHandler) DeleteMetricsByIDs(c *gin.Context, vis hydra.Visitor) {
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	accountInfo := interfaces.AccountInfo{ID: vis.ID, Type: string(vis.Type)}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	knID := c.Param("kn_id")
	metricIDsStr := c.Param("metric_ids")
	branch := c.DefaultQuery("branch", interfaces.MAIN_BRANCH)

	_, exist, err := r.kns.CheckKNExistByID(ctx, knID, branch)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if !exist {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, berrors.BknBackend_KnowledgeNetwork_NotFound)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	ids := common.StringToStringSlice(metricIDsStr)
	if err = r.ms.DeleteMetricsByIDs(ctx, nil, knID, branch, ids); err != nil {
		httpErr := err.(*rest.HTTPError)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	for _, id := range ids {
		audit.NewInfoLog(audit.OPERATION, audit.DELETE, audit.TransforOperator(vis),
			interfaces.GenerateMetricAuditObject(id, ""), "")
	}
	oteltrace.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusNoContent, nil)
}

func (r *restHandler) SearchMetricsByIn(c *gin.Context) {
	r.SearchMetrics(c, visitor.GenerateVisitor(c))
}

func (r *restHandler) SearchMetricsByEx(c *gin.Context) {
	vis, err := r.verifyOAuth(rest.GetLanguageCtx(c), c)
	if err != nil {
		return
	}
	r.SearchMetrics(c, vis)
}

func (r *restHandler) SearchMetrics(c *gin.Context, vis hydra.Visitor) {
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	accountInfo := interfaces.AccountInfo{ID: vis.ID, Type: string(vis.Type)}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)
	oteltrace.AddHttpAttrs4API(span, oteltrace.GetAttrsByGinCtx(c))

	knID := c.Param("kn_id")
	branch := c.DefaultQuery("branch", interfaces.MAIN_BRANCH)
	span.SetAttributes(attr.Key("kn_id").String(knID), attr.Key("branch").String(branch))

	_, exist, err := r.kns.CheckKNExistByID(ctx, knID, branch)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if !exist {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, berrors.BknBackend_KnowledgeNetwork_NotFound)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	query := interfaces.ConceptsQuery{}
	if err = c.ShouldBindJSON(&query); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_Metric_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("Binding Concept Query Paramter Failed:%s", err.Error()))
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	query.KNID = knID
	query.Branch = branch
	query.ModuleType = interfaces.MODULE_TYPE_METRIC

	if query.Limit == 0 {
		query.Limit = interfaces.DEFAULT_CONCEPT_SEARCH_LIMIT
	}
	if query.Sort == nil {
		query.Sort = []*interfaces.SortParams{
			{Field: interfaces.OPENSEARCH_SCORE_FIELD, Direction: interfaces.DESC_DIRECTION},
			{Field: interfaces.CONCEPT_ID_FIELD, Direction: interfaces.ASC_DIRECTION},
		}
	}

	if err = validateConceptsQuery(ctx, &query); err != nil {
		httpErr := err.(*rest.HTTPError)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	result, err := r.ms.SearchMetrics(ctx, &query)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	oteltrace.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, result)
}

// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/audit"
	"github.com/kweaver-ai/kweaver-go-lib/hydra"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	attr "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"bkn-backend/common"
	"bkn-backend/common/visitor"
	berrors "bkn-backend/errors"
	"bkn-backend/interfaces"
)

func (r *restHandler) HandleRiskTypeGetOverrideByEx(c *gin.Context) {
	switch c.GetHeader(interfaces.HTTP_HEADER_METHOD_OVERRIDE) {
	case "", http.MethodPost:
		r.CreateRiskTypesByEx(c)
	case http.MethodGet:
		r.SearchRiskTypesByEx(c)
	default:
		httpErr := rest.NewHTTPError(rest.GetLanguageCtx(c), http.StatusBadRequest,
			berrors.BknBackend_InvalidParameter_OverrideMethod)
		rest.ReplyError(c, httpErr)
	}
}

func (r *restHandler) HandleRiskTypeGetOverrideByIn(c *gin.Context) {
	switch c.GetHeader(interfaces.HTTP_HEADER_METHOD_OVERRIDE) {
	case "", http.MethodPost:
		r.CreateRiskTypesByIn(c)
	case http.MethodGet:
		r.SearchRiskTypesByIn(c)
	default:
		httpErr := rest.NewHTTPError(rest.GetLanguageCtx(c), http.StatusBadRequest,
			berrors.BknBackend_InvalidParameter_OverrideMethod)
		rest.ReplyError(c, httpErr)
	}
}

func (r *restHandler) CreateRiskTypesByIn(c *gin.Context) {
	visitor := visitor.GenerateVisitor(c)
	r.CreateRiskTypes(c, visitor)
}

func (r *restHandler) CreateRiskTypesByEx(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c), "创建风险类", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.CreateRiskTypes(c, visitor)
}

func (r *restHandler) CreateRiskTypes(c *gin.Context, visitor hydra.Visitor) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c), "创建风险类", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	accountInfo := interfaces.AccountInfo{ID: visitor.ID, Type: string(visitor.Type)}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)
	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	knID := c.Param("kn_id")
	branch := c.DefaultQuery("branch", interfaces.MAIN_BRANCH)
	span.SetAttributes(attr.Key("kn_id").String(knID), attr.Key("branch").String(branch))

	_, exist, err := r.kns.CheckKNExistByID(ctx, knID, branch)
	if err != nil {
		rest.ReplyError(c, err.(*rest.HTTPError))
		return
	}
	if !exist {
		rest.ReplyError(c, rest.NewHTTPError(ctx, http.StatusForbidden, berrors.BknBackend_KnowledgeNetwork_NotFound))
		return
	}

	mode := c.DefaultQuery(interfaces.QueryParam_ImportMode, interfaces.ImportMode_Normal)
	if httpErr := validateImportMode(ctx, mode); httpErr != nil {
		rest.ReplyError(c, httpErr)
		return
	}

	var requestData struct {
		Entries []*interfaces.RiskType `json:"entries"`
	}
	if err = c.ShouldBindJSON(&requestData); err != nil {
		rest.ReplyError(c, rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_RiskType_InvalidParameter).
			WithErrorDetails("Binding Parameter Failed:"+err.Error()))
		return
	}
	riskTypes := requestData.Entries
	if len(riskTypes) == 0 {
		rest.ReplyError(c, rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_InvalidParameter_RequestBody).
			WithErrorDetails("No risk type was passed in"))
		return
	}

	// request来的riskTypes的branch都用url里的branch
	for i := range riskTypes {
		riskTypes[i].KNID = knID
		riskTypes[i].Branch = branch
	}

	if err = ValidateRiskTypes(ctx, knID, riskTypes); err != nil {
		rest.ReplyError(c, err.(*rest.HTTPError))
		return
	}

	rtIDs, err := r.rtsRisk.CreateRiskTypes(ctx, nil, riskTypes, mode)
	if err != nil {
		rest.ReplyError(c, err.(*rest.HTTPError))
		return
	}

	for _, rt := range riskTypes {
		audit.NewInfoLog(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
			interfaces.GenerateRiskTypeAuditObject(rt.RTID, rt.RTName), "")
	}

	result := []any{}
	for _, id := range rtIDs {
		result = append(result, map[string]any{"id": id})
	}
	rest.ReplyOK(c, http.StatusCreated, result)
}

func (r *restHandler) UpdateRiskTypeByIn(c *gin.Context) {
	visitor := visitor.GenerateVisitor(c)
	r.UpdateRiskType(c, visitor)
}

func (r *restHandler) UpdateRiskTypeByEx(c *gin.Context) {
	visitor, err := r.verifyOAuth(rest.GetLanguageCtx(c), c)
	if err != nil {
		return
	}
	r.UpdateRiskType(c, visitor)
}

func (r *restHandler) UpdateRiskType(c *gin.Context, visitor hydra.Visitor) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c), "修改风险类", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	accountInfo := interfaces.AccountInfo{ID: visitor.ID, Type: string(visitor.Type)}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)
	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	knID := c.Param("kn_id")
	rtID := c.Param("rt_id")
	branch := c.DefaultQuery("branch", interfaces.MAIN_BRANCH)
	span.SetAttributes(attr.Key("kn_id").String(knID), attr.Key("rt_id").String(rtID), attr.Key("branch").String(branch))

	_, exist, err := r.kns.CheckKNExistByID(ctx, knID, branch)
	if err != nil {
		rest.ReplyError(c, err.(*rest.HTTPError))
		return
	}
	if !exist {
		rest.ReplyError(c, rest.NewHTTPError(ctx, http.StatusForbidden, berrors.BknBackend_KnowledgeNetwork_NotFound))
		return
	}

	var riskType interfaces.RiskType
	if err = c.ShouldBindJSON(&riskType); err != nil {
		rest.ReplyError(c, rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_RiskType_InvalidParameter).
			WithErrorDetails("Binding Parameter Failed:"+err.Error()))
		return
	}
	riskType.RTID = rtID
	riskType.KNID = knID
	riskType.Branch = branch

	oldName, exist, err := r.rtsRisk.CheckRiskTypeExistByID(ctx, knID, branch, rtID)
	if err != nil {
		rest.ReplyError(c, err.(*rest.HTTPError))
		return
	}
	if !exist {
		rest.ReplyError(c, rest.NewHTTPError(ctx, http.StatusForbidden, berrors.BknBackend_RiskType_RiskTypeNotFound))
		return
	}

	if err = ValidateRiskType(ctx, &riskType); err != nil {
		rest.ReplyError(c, err.(*rest.HTTPError))
		return
	}

	if oldName != riskType.RTName {
		_, exist, err = r.rtsRisk.CheckRiskTypeExistByName(ctx, knID, branch, riskType.RTName)
		if err != nil {
			rest.ReplyError(c, err.(*rest.HTTPError))
			return
		}
		if exist {
			errDetails := fmt.Sprintf("risk type name '%s' already exists", riskType.RTName)
			rest.ReplyError(c, rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_RiskType_RiskTypeNameExisted).
				WithDescription(map[string]any{"name": riskType.RTName}).
				WithErrorDetails(errDetails))
			return
		}
	}

	if err = r.rtsRisk.UpdateRiskType(ctx, nil, &riskType); err != nil {
		rest.ReplyError(c, err.(*rest.HTTPError))
		return
	}

	audit.NewInfoLog(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
		interfaces.GenerateRiskTypeAuditObject(rtID, riskType.RTName), "")
	rest.ReplyOK(c, http.StatusNoContent, nil)
}

func (r *restHandler) DeleteRiskTypes(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c), "删除风险类", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	accountInfo := interfaces.AccountInfo{ID: visitor.ID, Type: string(visitor.Type)}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)
	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	knID := c.Param("kn_id")
	rtIDsStr := c.Param("rt_ids")
	branch := c.DefaultQuery("branch", interfaces.MAIN_BRANCH)
	span.SetAttributes(attr.Key("kn_id").String(knID), attr.Key("rt_ids").String(rtIDsStr), attr.Key("branch").String(branch))

	_, exist, err := r.kns.CheckKNExistByID(ctx, knID, branch)
	if err != nil {
		rest.ReplyError(c, err.(*rest.HTTPError))
		return
	}
	if !exist {
		rest.ReplyError(c, rest.NewHTTPError(ctx, http.StatusForbidden, berrors.BknBackend_KnowledgeNetwork_NotFound))
		return
	}

	rtIDs := common.StringToStringSlice(rtIDsStr)
	var riskTypes []*interfaces.RiskType
	for _, rtID := range rtIDs {
		rtName, exist, e := r.rtsRisk.CheckRiskTypeExistByID(ctx, knID, branch, rtID)
		if e != nil {
			rest.ReplyError(c, e.(*rest.HTTPError))
			return
		}
		if !exist {
			rest.ReplyError(c, rest.NewHTTPError(ctx, http.StatusForbidden, berrors.BknBackend_RiskType_RiskTypeNotFound))
			return
		}
		riskTypes = append(riskTypes, &interfaces.RiskType{RTID: rtID, RTName: rtName})
	}

	if err = r.rtsRisk.DeleteRiskTypesByIDs(ctx, nil, knID, branch, rtIDs); err != nil {
		rest.ReplyError(c, err.(*rest.HTTPError))
		return
	}

	for _, rt := range riskTypes {
		audit.NewWarnLog(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
			interfaces.GenerateRiskTypeAuditObject(rt.RTID, rt.RTName), audit.SUCCESS, "")
	}
	rest.ReplyOK(c, http.StatusNoContent, nil)
}

func (r *restHandler) ListRiskTypesByIn(c *gin.Context) {
	visitor := visitor.GenerateVisitor(c)
	r.ListRiskTypes(c, visitor)
}

func (r *restHandler) ListRiskTypesByEx(c *gin.Context) {
	visitor, err := r.verifyOAuth(rest.GetLanguageCtx(c), c)
	if err != nil {
		return
	}
	r.ListRiskTypes(c, visitor)
}

func (r *restHandler) ListRiskTypes(c *gin.Context, visitor hydra.Visitor) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c), "分页获取风险类列表", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	accountInfo := interfaces.AccountInfo{ID: visitor.ID, Type: string(visitor.Type)}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)
	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	knID := c.Param("kn_id")
	branch := c.DefaultQuery("branch", interfaces.MAIN_BRANCH)
	span.SetAttributes(attr.Key("kn_id").String(knID), attr.Key("branch").String(branch))

	_, exist, err := r.kns.CheckKNExistByID(ctx, knID, branch)
	if err != nil {
		rest.ReplyError(c, err.(*rest.HTTPError))
		return
	}
	if !exist {
		rest.ReplyError(c, rest.NewHTTPError(ctx, http.StatusForbidden, berrors.BknBackend_KnowledgeNetwork_NotFound))
		return
	}

	namePattern := c.Query("name_pattern")
	tag := strings.Trim(c.Query("tag"), " ")
	offset := c.DefaultQuery("offset", interfaces.DEFAULT_OFFEST)
	limit := c.DefaultQuery("limit", interfaces.DEFAULT_LIMIT)
	sort := c.DefaultQuery("sort", "update_time")
	direction := c.DefaultQuery("direction", interfaces.DESC_DIRECTION)

	pageParam, err := validatePaginationQueryParameters(ctx, offset, limit, sort, direction, interfaces.RiskTypeSort)
	if err != nil {
		rest.ReplyError(c, err.(*rest.HTTPError))
		return
	}

	query := interfaces.RiskTypesQueryParams{
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

	list, total, err := r.rtsRisk.ListRiskTypes(ctx, query)
	if err != nil {
		rest.ReplyError(c, err.(*rest.HTTPError))
		return
	}

	logger.Debug("Handler ListRiskTypes Success")
	rest.ReplyOK(c, http.StatusOK, map[string]any{"entries": list, "total_count": total})
}

func (r *restHandler) GetRiskTypesByEx(c *gin.Context) {
	visitor, err := r.verifyOAuth(rest.GetLanguageCtx(c), c)
	if err != nil {
		return
	}
	r.GetRiskTypes(c, visitor)
}

func (r *restHandler) GetRiskTypes(c *gin.Context, visitor hydra.Visitor) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c), "按ID获取风险类", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	accountInfo := interfaces.AccountInfo{ID: visitor.ID, Type: string(visitor.Type)}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)
	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	knID := c.Param("kn_id")
	rtIDsStr := c.Param("rt_ids")
	branch := c.DefaultQuery("branch", interfaces.MAIN_BRANCH)
	span.SetAttributes(attr.Key("kn_id").String(knID), attr.Key("rt_ids").String(rtIDsStr), attr.Key("branch").String(branch))

	_, exist, err := r.kns.CheckKNExistByID(ctx, knID, branch)
	if err != nil {
		rest.ReplyError(c, err.(*rest.HTTPError))
		return
	}
	if !exist {
		rest.ReplyError(c, rest.NewHTTPError(ctx, http.StatusForbidden, berrors.BknBackend_KnowledgeNetwork_NotFound))
		return
	}

	rtIDs := common.StringToStringSlice(rtIDsStr)
	list, err := r.rtsRisk.GetRiskTypesByIDs(ctx, knID, branch, rtIDs)
	if err != nil {
		rest.ReplyError(c, err.(*rest.HTTPError))
		return
	}

	if len(list) != len(rtIDs) {
		foundIDs := make(map[string]bool)
		for _, rt := range list {
			foundIDs[rt.RTID] = true
		}
		var missing []string
		for _, id := range rtIDs {
			if !foundIDs[id] {
				missing = append(missing, id)
			}
		}
		rest.ReplyError(c, rest.NewHTTPError(ctx, http.StatusNotFound, berrors.BknBackend_RiskType_RiskTypeNotFound).
			WithErrorDetails(fmt.Sprintf("Risk types not found: %v", missing)))
		return
	}

	rest.ReplyOK(c, http.StatusOK, map[string]any{"entries": list})
}

// GetRiskTypesByInWithPath 内部 API：按 path 中的 rt_ids 获取风险类
func (r *restHandler) GetRiskTypesByInWithPath(c *gin.Context) {
	r.GetRiskTypes(c, visitor.GenerateVisitor(c))
}

// GetRiskTypesByIn 内部 API：按 risk_type_ids、branch 批量获取风险类（供 ontology-query 调用）
func (r *restHandler) GetRiskTypesByIn(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c), "内部获取风险类", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	knID := c.Param("kn_id")
	rtIDsStr := c.Query("risk_type_ids")
	branch := c.DefaultQuery("branch", interfaces.MAIN_BRANCH)
	span.SetAttributes(attr.Key("kn_id").String(knID), attr.Key("risk_type_ids").String(rtIDsStr), attr.Key("branch").String(branch))

	if rtIDsStr == "" {
		rest.ReplyOK(c, http.StatusOK, map[string]any{"entries": []*interfaces.RiskType{}})
		return
	}

	rtIDs := common.StringToStringSlice(rtIDsStr)
	list, err := r.rtsRisk.GetRiskTypesByIDs(ctx, knID, branch, rtIDs)
	if err != nil {
		rest.ReplyError(c, err.(*rest.HTTPError))
		return
	}

	rest.ReplyOK(c, http.StatusOK, map[string]any{"entries": list})
}

// 检索风险类（内部）
func (r *restHandler) SearchRiskTypesByIn(c *gin.Context) {
	logger.Debug("Handler SearchRiskTypesByIn Start")
	visitor := visitor.GenerateVisitor(c)
	r.SearchRiskTypes(c, visitor)
}

// 检索风险类（外部）
func (r *restHandler) SearchRiskTypesByEx(c *gin.Context) {
	logger.Debug("Handler SearchRiskTypesByEx Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"检索风险类", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.SearchRiskTypes(c, visitor)
}

// 检索风险类
func (r *restHandler) SearchRiskTypes(c *gin.Context, visitor hydra.Visitor) {
	logger.Debug("SearchRiskTypes Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"检索风险类", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))
	o11y.Info(ctx, fmt.Sprintf("检索风险类请求参数: [%s]", c.Request.RequestURI))

	knID := c.Param("kn_id")
	branch := c.DefaultQuery("branch", interfaces.MAIN_BRANCH)
	span.SetAttributes(
		attr.Key("kn_id").String(knID),
		attr.Key("branch").String(branch),
	)

	_, exist, err := r.kns.CheckKNExistByID(ctx, knID, branch)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if !exist {
		httpErr := rest.NewHTTPError(ctx, http.StatusForbidden,
			berrors.BknBackend_KnowledgeNetwork_NotFound)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	query := interfaces.ConceptsQuery{}
	err = c.ShouldBindJSON(&query)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_RiskType_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("Binding Concept Query Paramter Failed:%s", err.Error()))
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails))
		rest.ReplyError(c, httpErr)
		return
	}
	query.KNID = knID
	query.Branch = branch
	query.ModuleType = interfaces.MODULE_TYPE_RISK_TYPE

	if query.Limit == 0 {
		query.Limit = interfaces.DEFAULT_CONCEPT_SEARCH_LIMIT
	}

	if query.Sort == nil {
		query.Sort = []*interfaces.SortParams{
			{
				Field:     interfaces.OPENSEARCH_SCORE_FIELD,
				Direction: interfaces.DESC_DIRECTION,
			},
			{
				Field:     interfaces.CONCEPT_ID_FIELD,
				Direction: interfaces.ASC_DIRECTION,
			},
		}
	}

	err = validateConceptsQuery(ctx, &query)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails))
		rest.ReplyError(c, httpErr)
		return
	}

	result, err := r.rtsRisk.SearchRiskTypes(ctx, &query)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	logger.Debug("Handler SearchRiskTypes Success")
	rest.ReplyOK(c, http.StatusOK, result)
}

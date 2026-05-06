// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package driveradapters provides HTTP handlers (primary adapters).
package driveradapters

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"vega-backend/common"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/audit"
	"github.com/kweaver-ai/kweaver-go-lib/hydra"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	verrors "vega-backend/errors"
	"vega-backend/interfaces"
)

// ========== ListResources ==========

// ListResourcesByEx handles GET /api/vega-backend/v1/resources (External)
func (r *restHandler) ListResourcesByEx(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"ListResourcesByEx", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 外网接口：校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.listResources(c, ctx, span, visitor)
}

// ListResourcesByIn handles GET /api/vega-backend/in/v1/resources (Internal)
func (r *restHandler) ListResourcesByIn(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"ListResourcesByIn", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 内网接口：user_id从header中取
	visitor := GenerateVisitor(c)
	r.listResources(c, ctx, span, visitor)
}

// listResources is the shared implementation
func (r *restHandler) listResources(c *gin.Context, ctx context.Context, span trace.Span, visitor hydra.Visitor) {
	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	catalogID := c.Query("catalog_id")
	category := c.Query("category")
	status := c.Query("status")
	database := c.Query("database")
	offset := common.GetQueryOrDefault(c, "offset", interfaces.DEFAULT_OFFSET)
	limit := common.GetQueryOrDefault(c, "limit", interfaces.DEFAULT_LIMIT)
	sort := common.GetQueryOrDefault(c, "sort", "update_time")
	direction := common.GetQueryOrDefault(c, "direction", interfaces.DESC_DIRECTION)

	// 校验分页查询参数
	pageParam, err := validatePaginationQueryParams(ctx,
		offset, limit, sort, direction, interfaces.RESOURCE_SORT)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails))
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	params := interfaces.ResourcesQueryParams{
		PaginationQueryParams: pageParam,
		CatalogID:             catalogID,
		Category:              category,
		Status:                status,
		Database:              database,
	}

	entries, total, err := r.rs.List(ctx, params)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	result := map[string]any{
		"entries":     entries,
		"total_count": total,
	}

	logger.Debug("Handler ListResources Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, result)
}

// ========== CreateResource ==========

// CreateResourceByEx handles POST /api/vega-backend/v1/resources (External)
func (r *restHandler) CreateResourceByEx(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"CreateResourceByEx", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 外网接口：校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.createResource(c, ctx, span, visitor)
}

// CreateResourceByIn handles POST /api/vega-backend/in/v1/resources (Internal)
func (r *restHandler) CreateResourceByIn(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"CreateResourceByIn", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 内网接口：user_id从header中取
	visitor := GenerateVisitor(c)
	r.createResource(c, ctx, span, visitor)
}

// createResource is the shared implementation
func (r *restHandler) createResource(c *gin.Context, ctx context.Context, span trace.Span, visitor hydra.Visitor) {
	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	var req interfaces.ResourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest,
			verrors.VegaBackend_InvalidParameter_RequestBody).WithErrorDetails(err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	if err := ValidateResourceRequest(ctx, &req); err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// Check catelog exists
	csExists, csErr := r.cs.CheckExistByID(ctx, req.CatalogID)
	if csErr != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError,
			verrors.VegaBackend_Resource_InternalError).WithErrorDetails(csErr.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if !csExists {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Resource_CatalogNotFound).
			WithErrorDetails(fmt.Sprintf("catalog %s not found", req.CatalogID))
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// Check if name exists
	exists, err := r.rs.CheckExistByName(ctx, req.CatalogID, req.Name)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError,
			verrors.VegaBackend_Resource_InternalError).WithErrorDetails(err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if exists {
		httpErr := rest.NewHTTPError(ctx, http.StatusConflict, verrors.VegaBackend_Resource_NameExists)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// Check if id exists if provided
	if req.ID != "" {
		exists, err := r.rs.CheckExistByID(ctx, req.ID)
		if err != nil {
			httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError,
				verrors.VegaBackend_Resource_InternalError).WithErrorDetails(err.Error())
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
		if exists {
			httpErr := rest.NewHTTPError(ctx, http.StatusConflict, verrors.VegaBackend_Resource_IDExists).
				WithErrorDetails(fmt.Sprintf("id %s already exists", req.ID))
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
	}

	resource, err := r.rs.Create(ctx, &req)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 成功创建记录审计日志
	audit.NewInfoLog(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
		interfaces.GenerateResourceAuditObject(resource.ID, req.Name), "")

	result := map[string]any{"id": resource.ID}

	logger.Debug("Handler CreateResource Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusCreated, result)
}

// ========== GetResources ==========

// GetResourcesByEx handles GET /api/vega-backend/v1/resources/:ids (External)
func (r *restHandler) GetResourcesByEx(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"GetResourcesByEx", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 外网接口：校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.getResources(c, ctx, span, visitor)
}

// GetResourcesByIn handles GET /api/vega-backend/in/v1/resources/:ids (Internal)
func (r *restHandler) GetResourcesByIn(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"GetResourcesByIn", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 内网接口：user_id从header中取
	visitor := GenerateVisitor(c)
	r.getResources(c, ctx, span, visitor)
}

// getResources is the shared implementation
func (r *restHandler) getResources(c *gin.Context, ctx context.Context, span trace.Span, visitor hydra.Visitor) {
	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	ids := strings.Split(c.Param("ids"), ",")

	resources, err := r.rs.GetByIDs(ctx, ids)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	if len(resources) != len(ids) {
		for _, id := range ids {
			found := false
			for _, resource := range resources {
				if resource.ID == id {
					found = true
					break
				}
			}
			if !found {
				httpErr := rest.NewHTTPError(ctx, http.StatusNotFound,
					verrors.VegaBackend_Resource_NotFound).WithErrorDetails(fmt.Sprintf("id %s not found", id))
				o11y.AddHttpAttrs4HttpError(span, httpErr)
				rest.ReplyError(c, httpErr)
				return
			}
		}
	}

	result := map[string]any{"entries": resources}

	logger.Debug("Handler GetResource Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, result)
}

// ========== UpdateResource ==========

// UpdateResourceByEx handles PUT /api/vega-backend/v1/resources/:id (External)
func (r *restHandler) UpdateResourceByEx(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"UpdateResourceByEx", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 外网接口：校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.updateResource(c, ctx, span, visitor)
}

// UpdateResourceByIn handles PUT /api/vega-backend/in/v1/resources/:id (Internal)
func (r *restHandler) UpdateResourceByIn(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"UpdateResourceByIn", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 内网接口：user_id从header中取
	visitor := GenerateVisitor(c)
	r.updateResource(c, ctx, span, visitor)
}

// updateResource is the shared implementation
func (r *restHandler) updateResource(c *gin.Context, ctx context.Context, span trace.Span, visitor hydra.Visitor) {
	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	id := c.Param("id")

	var req interfaces.ResourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest,
			verrors.VegaBackend_InvalidParameter_RequestBody).WithErrorDetails(err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	if err := ValidateResourceRequest(ctx, &req); err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// Check if id exists
	resource, err := r.rs.GetByID(ctx, id)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	req.OriginResource = resource

	// Apply updates
	if req.Name != resource.Name {
		exists, err := r.rs.CheckExistByName(ctx, req.CatalogID, req.Name)
		if err != nil {
			httpErr := err.(*rest.HTTPError)
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
		if exists {
			span.SetStatus(codes.Error, "Resource name exists")
			httpErr := rest.NewHTTPError(ctx, http.StatusConflict, verrors.VegaBackend_Resource_NameExists)
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
		req.IfNameModify = true
	}

	if err := r.rs.Update(ctx, id, &req); err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	audit.NewInfoLog(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
		interfaces.GenerateResourceAuditObject(id, req.Name), "")

	logger.Debug("Handler UpdateResource Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusNoContent)
	rest.ReplyOK(c, http.StatusNoContent, nil)
}

// ========== DeleteResources ==========

// DeleteResourcesByEx handles DELETE /api/vega-backend/v1/resources/:ids (External)
func (r *restHandler) DeleteResourcesByEx(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"DeleteResourcesByEx", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 外网接口：校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.deleteResources(c, ctx, span, visitor)
}

// DeleteResourcesByIn handles DELETE /api/vega-backend/in/v1/resources/:ids (Internal)
func (r *restHandler) DeleteResourcesByIn(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"DeleteResourcesByIn", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 内网接口：user_id从header中取
	visitor := GenerateVisitor(c)
	r.deleteResources(c, ctx, span, visitor)
}

// deleteResources is the shared implementation
func (r *restHandler) deleteResources(c *gin.Context, ctx context.Context, span trace.Span, visitor hydra.Visitor) {
	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	ids := strings.Split(c.Param("ids"), ",")

	// Check if ids exists
	for _, id := range ids {
		exists, err := r.rs.CheckExistByID(ctx, id)
		if err != nil {
			httpErr := err.(*rest.HTTPError)
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
		if !exists {
			httpErr := rest.NewHTTPError(ctx, http.StatusNotFound,
				verrors.VegaBackend_Resource_NotFound).WithErrorDetails(fmt.Sprintf("id %s not found", id))
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
	}

	if err := r.rs.DeleteByIDs(ctx, ids); err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	for _, id := range ids {
		audit.NewWarnLog(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
			interfaces.GenerateResourceAuditObject(id, ""), audit.SUCCESS, "")
	}

	logger.Debug("Handler DeleteResource Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusNoContent)
	rest.ReplyOK(c, http.StatusNoContent, nil)
}

// ========== ListResourceSrcs ==========

// ListResourceSrcsByEx resource source list (External)
func (r *restHandler) ListResourceSrcsByEx(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"ListResourceSrcsByEx", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 外网接口：校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.listResourceSrcs(c, ctx, span, visitor)
}

// listResourceSrcs is the shared implementation
func (r *restHandler) listResourceSrcs(c *gin.Context, ctx context.Context, span trace.Span, visitor hydra.Visitor) {
	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	// 获取查询参数
	id := c.Query("id")
	keyword := strings.TrimSpace(c.Query("keyword"))
	offset := common.GetQueryOrDefault(c, "offset", interfaces.DEFAULT_OFFSET)
	limit := common.GetQueryOrDefault(c, "limit", "50")
	sort := common.GetQueryOrDefault(c, "sort", "update_time")
	direction := common.GetQueryOrDefault(c, "direction", interfaces.DESC_DIRECTION)

	// 校验分页查询参数
	pageParam, err := validatePaginationQueryParams(ctx,
		offset, limit, sort, direction, interfaces.RESOURCE_SORT)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails))
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	params := interfaces.ListResourcesQueryParams{
		PaginationQueryParams: pageParam,
		ID:                    id,
		Keyword:               keyword,
	}

	entries, total, err := r.rs.ListResourceSrcs(ctx, params)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	result := map[string]any{
		"entries":     entries,
		"total_count": total,
	}

	logger.Debug("Handler ListResourceSrcs Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, result)
}

// 分页获取资源列表
func (r *restHandler) ListResources(c *gin.Context) {
	logger.Debug("ListResources Start")

	ctx := rest.GetLanguageCtx(c)
	resourceType := strings.TrimSpace(c.Query("resource_type"))
	switch resourceType {
	case interfaces.RESOURCE_TYPE_CATALOG:
		r.ListCatalogSrcsByEx(c)
	case interfaces.RESOURCE_TYPE_RESOURCE:
		r.ListResourceSrcsByEx(c)
	case "":
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest,
			verrors.VegaBackend_InvalidParameter_RequestBody).
			WithErrorDetails("resource_type is required; valid values: catalog, resource")
		rest.ReplyError(c, httpErr)
	default:
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest,
			verrors.VegaBackend_Resource_InvalidParameter).
			WithErrorDetails("resource_type is invalid; valid values: catalog, resource")
		rest.ReplyError(c, httpErr)
	}

}

// ========== Resource Build ==========

// CreateBuildTaskByEx handles POST /api/vega-backend/v1/resources/:id/build (External)
func (r *restHandler) CreateBuildTaskByEx(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"CreateBuildTaskByEx", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 外网接口：校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.createBuildTask(c, ctx, span, visitor)
}

// CreateBuildTaskByIn handles POST /api/vega-backend/in/v1/resources/:id/build (Internal)
func (r *restHandler) CreateBuildTaskByIn(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"CreateBuildTaskByIn", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 内网接口：user_id从header中取
	visitor := GenerateVisitor(c)
	r.createBuildTask(c, ctx, span, visitor)
}

// createBuildTask is the shared implementation
func (r *restHandler) createBuildTask(c *gin.Context, ctx context.Context, span trace.Span, visitor hydra.Visitor) {
	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	id := c.Param("id")

	resource, err := r.rs.GetByID(ctx, id)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// Parse request body to get task mode
	var req interfaces.BuildTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_RequestBody).
			WithErrorDetails(err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// check req.BuildKeyFields and req.EmbeddingFields
	if req.BuildKeyFields == "" && req.Mode == interfaces.BuildTaskModeBatch {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_RequestBody).
			WithErrorDetails("build_key_fields is required")
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// Create a map of schema fields for quick lookup
	schemaFields := make(map[string]bool)
	if resource.SchemaDefinition != nil {
		for _, prop := range resource.SchemaDefinition {
			schemaFields[prop.Name] = true
		}
	}

	// Check build key fields
	buildKeys := strings.Split(req.BuildKeyFields, ",")
	for _, key := range buildKeys {
		key = strings.TrimSpace(key)
		if key != "" && !schemaFields[key] {
			httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_RequestBody).
				WithErrorDetails(fmt.Sprintf("build_key_field '%s' not found in resource schema", key))
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
	}

	// Check embedding fields if provided
	if req.EmbeddingFields != "" {
		embeddingFields := strings.Split(req.EmbeddingFields, ",")
		for _, field := range embeddingFields {
			field = strings.TrimSpace(field)
			if field != "" && !schemaFields[field] {
				httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_RequestBody).
					WithErrorDetails(fmt.Sprintf("embedding_field '%s' not found in resource schema", field))
				o11y.AddHttpAttrs4HttpError(span, httpErr)
				rest.ReplyError(c, httpErr)
				return
			}
		}
	}

	taskID, err := r.rs.CreateBuildTask(ctx, id, &req)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	audit.NewInfoLog(audit.OPERATION, "build", audit.TransforOperator(visitor),
		interfaces.GenerateResourceAuditObject(id, ""), "")

	logger.Debug("Handler CreateBuildTaskset Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusCreated)
	rest.ReplyOK(c, http.StatusCreated, gin.H{"task_id": taskID})
}

// GetBuildTaskByEx handles get build task request (External)
func (r *restHandler) GetBuildTaskByEx(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"GetBuildTaskByEx", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}

	r.getBuildTask(c, ctx, span, visitor)
}

// GetBuildTaskByIn handles get build task request (Internal)
func (r *restHandler) GetBuildTaskByIn(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"GetBuildTaskByIn", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 内网接口：user_id从header中取
	visitor := GenerateVisitor(c)
	r.getBuildTask(c, ctx, span, visitor)
}

// getBuildTask is the shared implementation for getting build task
func (r *restHandler) getBuildTask(c *gin.Context, ctx context.Context, span trace.Span, visitor hydra.Visitor) {
	// Get task ID from path parameter
	id := c.Param("id")
	taskID := c.Param("taskid")
	if taskID == "" {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, "VegaBackend.InvalidRequestParameter.TaskID")
		rest.ReplyError(c, httpErr)
		return
	}

	// Check if resource exists
	exists, err := r.rs.CheckExistByID(ctx, id)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if !exists {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound,
			verrors.VegaBackend_Resource_NotFound).WithErrorDetails(fmt.Sprintf("resource id %s not found", id))
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// Get build task
	buildTask, err := r.rs.GetBuildTaskByID(ctx, taskID)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}

	rest.ReplyOK(c, http.StatusOK, buildTask)
}

// UpdateBuildTaskStatusByEx handles update build task status request (External)
func (r *restHandler) UpdateBuildTaskStatusByEx(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"UpdateBuildTaskStatusByEx", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}

	r.updateBuildTaskStatus(c, ctx, span, visitor)
}

// UpdateBuildTaskStatusByIn handles update build task status request (Internal)
func (r *restHandler) UpdateBuildTaskStatusByIn(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"UpdateBuildTaskStatusByIn", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 内网接口：user_id从header中取
	visitor := GenerateVisitor(c)
	r.updateBuildTaskStatus(c, ctx, span, visitor)
}

// updateBuildTaskStatus is the shared implementation for updating build task status
func (r *restHandler) updateBuildTaskStatus(c *gin.Context, ctx context.Context, span trace.Span, visitor hydra.Visitor) {
	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	// Get task ID from path parameter
	id := c.Param("id")
	taskID := c.Param("taskid")
	if taskID == "" {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, "VegaBackend.InvalidRequestParameter.TaskID")
		rest.ReplyError(c, httpErr)
		return
	}

	// Check if resource exists
	exists, err := r.rs.CheckExistByID(ctx, id)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if !exists {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound,
			verrors.VegaBackend_Resource_NotFound).WithErrorDetails(fmt.Sprintf("resource id %s not found", id))
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// Parse request body
	var req interfaces.UpdateBuildTaskStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_RequestBody).
			WithErrorDetails(err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// Update build task status
	if err := r.rs.UpdateBuildTaskStatus(ctx, taskID, &req); err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	logger.Debug("Handler UpdateBuildTaskStatus Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, nil)
}

// ListBuildTasksByEx handles list build tasks request (External)
func (r *restHandler) ListBuildTasksByEx(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"ListBuildTasksByEx", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}

	r.listBuildTasks(c, ctx, span, visitor)
}

// ListBuildTasksByIn handles list build tasks request (Internal)
func (r *restHandler) ListBuildTasksByIn(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"ListBuildTasksByIn", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 内网接口：user_id从header中取
	visitor := GenerateVisitor(c)
	r.listBuildTasks(c, ctx, span, visitor)
}

// listBuildTasks is the shared implementation for listing build tasks
func (r *restHandler) listBuildTasks(c *gin.Context, ctx context.Context, span trace.Span, visitor hydra.Visitor) {
	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	// 解析分页参数
	offset := 0
	limit := 100
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if val, err := strconv.Atoi(offsetStr); err == nil {
			offset = val
		}
	}
	if limitStr := c.Query("limit"); limitStr != "" {
		if val, err := strconv.Atoi(limitStr); err == nil {
			limit = val
		}
	}

	// Get build tasks with pagination
	buildTasks, totalCount, err := r.rs.GetBuildTasks(ctx, offset, limit)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 返回包含 entries 和 total_count 字段的对象
	rest.ReplyOK(c, http.StatusOK, map[string]any{
		"entries":     buildTasks,
		"total_count": totalCount,
	})
}

// DeleteBuildTasksByEx handles delete build tasks request (External)
func (r *restHandler) DeleteBuildTasksByEx(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"DeleteBuildTasksByEx", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}

	r.deleteBuildTasks(c, ctx, span, visitor)
}

// DeleteBuildTasksByIn handles delete build tasks request (Internal)
func (r *restHandler) DeleteBuildTasksByIn(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"DeleteBuildTasksByIn", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 内网接口：user_id从header中取
	visitor := GenerateVisitor(c)
	r.deleteBuildTasks(c, ctx, span, visitor)
}

// deleteBuildTasks is the shared implementation for deleting build tasks
func (r *restHandler) deleteBuildTasks(c *gin.Context, ctx context.Context, span trace.Span, visitor hydra.Visitor) {
	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	// Get task IDs from path parameter
	taskIDs := c.Param("taskids")
	if taskIDs == "" {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, "VegaBackend.InvalidRequestParameter.TaskIDs")
		rest.ReplyError(c, httpErr)
		return
	}

	// Split task IDs
	taskIDList := strings.Split(taskIDs, ",")

	// Delete build tasks
	for _, taskID := range taskIDList {
		if taskID != "" {
			if err := r.rs.DeleteBuildTask(ctx, taskID); err != nil {
				httpErr := err.(*rest.HTTPError)
				o11y.AddHttpAttrs4HttpError(span, httpErr)
				rest.ReplyError(c, httpErr)
				return
			}
		}
	}

	logger.Debug("Handler DeleteBuildTasks Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusNoContent)
	rest.ReplyOK(c, http.StatusNoContent, nil)
}

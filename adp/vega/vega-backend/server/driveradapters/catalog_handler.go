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
	"reflect"
	"strings"
	"vega-backend/common"

	"github.com/robfig/cron/v3"

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

// Helper function to validate strategies array
func validateStrategies(strategies []string) error {
	validStrategies := map[string]bool{
		"insert": true,
		"delete": true,
		"update": true,
	}
	for _, strategy := range strategies {
		if !validStrategies[strategy] {
			return fmt.Errorf("invalid strategy: %s, must be one of: insert, delete, update", strategy)
		}
	}
	return nil
}

// ========== ListCatalogs ==========

// ListCatalogsByEx handles GET /api/vega-backend/v1/catalogs (External)
func (r *restHandler) ListCatalogsByEx(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"ListCatalogsByEx", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 外网接口：校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.listCatalogs(c, ctx, span, visitor)
}

// ListCatalogsByIn handles GET /api/vega-backend/in/v1/catalogs (Internal)
func (r *restHandler) ListCatalogsByIn(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"ListCatalogsByIn", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 内网接口：user_id从header中取
	visitor := GenerateVisitor(c)
	r.listCatalogs(c, ctx, span, visitor)
}

// listCatalogs is the shared implementation
func (r *restHandler) listCatalogs(c *gin.Context, ctx context.Context, span trace.Span, visitor hydra.Visitor) {
	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	// 获取查询参数
	tag := strings.TrimSpace(c.Query("tag"))
	typ := c.Query("type")
	healthCheckStatus := c.Query("health_check_status")
	offset := common.GetQueryOrDefault(c, "offset", interfaces.DEFAULT_OFFSET)
	limit := common.GetQueryOrDefault(c, "limit", interfaces.DEFAULT_LIMIT)
	sort := common.GetQueryOrDefault(c, "sort", "update_time")
	direction := common.GetQueryOrDefault(c, "direction", interfaces.DESC_DIRECTION)

	// 校验分页查询参数
	pageParam, err := validatePaginationQueryParams(ctx,
		offset, limit, sort, direction, interfaces.CATALOG_SORT)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails))
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	params := interfaces.CatalogsQueryParams{
		PaginationQueryParams: pageParam,
		Tag:                   tag,
		Type:                  typ,
		HealthCheckStatus:     healthCheckStatus,
	}

	entries, total, err := r.cs.List(ctx, params)
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

	logger.Debug("Handler ListCatalogs Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, result)
}

// ========== CreateCatalog ==========

// CreateCatalogByEx handles POST /api/vega-backend/v1/catalogs (External)
func (r *restHandler) CreateCatalogByEx(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"CreateCatalogByEx", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 外网接口：校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.createCatalog(c, ctx, span, visitor)
}

// CreateCatalogByIn handles POST /api/vega-backend/in/v1/catalogs (Internal)
func (r *restHandler) CreateCatalogByIn(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"CreateCatalogByIn", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 内网接口：user_id从header中取
	visitor := GenerateVisitor(c)
	r.createCatalog(c, ctx, span, visitor)
}

// createCatalog is the shared implementation
func (r *restHandler) createCatalog(c *gin.Context, ctx context.Context, span trace.Span, visitor hydra.Visitor) {
	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	var req interfaces.CatalogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_RequestBody).
			WithErrorDetails(err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	if err := ValidateCatalogRequest(ctx, &req); err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// Check if name exists
	exists, err := r.cs.CheckExistByName(ctx, req.Name)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Catalog_InternalError).
			WithErrorDetails(err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if exists {
		httpErr := rest.NewHTTPError(ctx, http.StatusConflict, verrors.VegaBackend_Catalog_NameExists)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// Check if id exists if provided
	if req.ID != "" {
		exists, err := r.cs.CheckExistByID(ctx, req.ID)
		if err != nil {
			httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Catalog_InternalError).
				WithErrorDetails(err.Error())
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
		if exists {
			httpErr := rest.NewHTTPError(ctx, http.StatusConflict, verrors.VegaBackend_Catalog_IDExists).
				WithErrorDetails(fmt.Sprintf("id %s already exists", req.ID))
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
	}

	id, err := r.cs.Create(ctx, &req)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 成功创建记录审计日志
	audit.NewInfoLog(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
		interfaces.GenerateCatalogAuditObject(id, req.Name), "")

	result := map[string]any{"id": id}

	logger.Debug("Handler CreateCatalog Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusCreated, result)
}

// ========== GetCatalogs ==========

// GetCatalogsByEx handles GET /api/vega-backend/v1/catalogs/:ids (External)
func (r *restHandler) GetCatalogsByEx(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"GetCatalogsByEx", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 外网接口：校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.getCatalogs(c, ctx, span, visitor)
}

// GetCatalogsByIn handles GET /api/vega-backend/in/v1/catalogs/:ids (Internal)
func (r *restHandler) GetCatalogsByIn(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"GetCatalogsByIn", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 内网接口：user_id从header中取
	visitor := GenerateVisitor(c)
	r.getCatalogs(c, ctx, span, visitor)
}

// getCatalogs is the shared implementation
func (r *restHandler) getCatalogs(c *gin.Context, ctx context.Context, span trace.Span, visitor hydra.Visitor) {
	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	ids := strings.Split(c.Param("ids"), ",")

	catalogs, err := r.cs.GetByIDs(ctx, ids)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	if len(catalogs) != len(ids) {
		for _, id := range ids {
			found := false
			for _, catalog := range catalogs {
				if catalog.ID == id {
					found = true
					break
				}
			}
			if !found {
				httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Catalog_NotFound).
					WithErrorDetails(fmt.Sprintf("id %s not found", id))
				o11y.AddHttpAttrs4HttpError(span, httpErr)
				rest.ReplyError(c, httpErr)
				return
			}
		}
	}

	result := map[string]any{"entries": catalogs}

	logger.Debug("Handler GetCatalogs Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, result)
}

// ========== UpdateCatalog ==========

// UpdateCatalogByEx handles PUT /api/vega-backend/v1/catalogs/:id (External)
func (r *restHandler) UpdateCatalogByEx(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"UpdateCatalogByEx", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 外网接口：校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.updateCatalog(c, ctx, span, visitor)
}

// UpdateCatalogByIn handles PUT /api/vega-backend/in/v1/catalogs/:id (Internal)
func (r *restHandler) UpdateCatalogByIn(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"UpdateCatalogByIn", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 内网接口：user_id从header中取
	visitor := GenerateVisitor(c)
	r.updateCatalog(c, ctx, span, visitor)
}

// updateCatalog is the shared implementation
func (r *restHandler) updateCatalog(c *gin.Context, ctx context.Context, span trace.Span, visitor hydra.Visitor) {
	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	id := c.Param("id")

	var req interfaces.CatalogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_RequestBody).
			WithErrorDetails(err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	if req.ID == "" {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Catalog_InvalidParameter_ID).
			WithErrorDetails("body field 'id' is required and must equal path parameter")
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if req.ID != id {
		httpErr := rest.NewHTTPError(ctx, http.StatusConflict, verrors.VegaBackend_Catalog_IDMismatch).
			WithErrorDetails(fmt.Sprintf("path id %q != body id %q", id, req.ID))
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	if err := ValidateCatalogRequest(ctx, &req); err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// Check if id exists
	catalog, err := r.cs.GetByID(ctx, id, false)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	req.OriginCatalog = catalog

	// Validate immutable fields
	// connector_type cannot be modified
	if req.ConnectorType != catalog.ConnectorType {
		span.SetStatus(codes.Error, "Connector type cannot be modified")
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Catalog_InvalidParameter_ConnectorType).
			WithErrorDetails("connector_type cannot be modified")
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// connector_config immutable fields: host, port, database, databases, schemas, paths, protocol
	// These fields cannot be modified or removed if they exist in the original catalog
	immutableFields := []string{"host", "port", "database", "databases", "schemas", "paths", "protocol", "concurrent"}
	for _, field := range immutableFields {
		if _, existsInCatalog := catalog.ConnectorCfg[field]; existsInCatalog {
			if _, existsInReq := req.ConnectorCfg[field]; existsInReq {
				// Field exists in both, check if it's being modified
				// Handle different types: string, number, array
				catalogValue := catalog.ConnectorCfg[field]
				reqValue := req.ConnectorCfg[field]

				var isModified bool
				switch v := catalogValue.(type) {
				case []interface{}:
					// Compare arrays using reflect.DeepEqual
					isModified = !reflect.DeepEqual(v, reqValue)
				default:
					// Compare other types (string, number, etc.)
					isModified = (reqValue != catalogValue)
				}

				if isModified {
					span.SetStatus(codes.Error, fmt.Sprintf("Connector config field %s cannot be modified", field))
					httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Catalog_InvalidParameter_ConnectorConfig).
						WithErrorDetails(fmt.Sprintf("connector_config.%s cannot be modified", field))
					o11y.AddHttpAttrs4HttpError(span, httpErr)
					rest.ReplyError(c, httpErr)
					return
				}
			} else {
				// Field exists in catalog but not in request - cannot remove immutable fields
				span.SetStatus(codes.Error, fmt.Sprintf("Connector config field %s cannot be removed", field))
				httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Catalog_InvalidParameter_ConnectorConfig).
					WithErrorDetails(fmt.Sprintf("connector_config.%s cannot be removed", field))
				o11y.AddHttpAttrs4HttpError(span, httpErr)
				rest.ReplyError(c, httpErr)
				return
			}
		}
	}

	// Apply updates
	if req.Name != catalog.Name {
		exists, err := r.cs.CheckExistByName(ctx, req.Name)
		if err != nil {
			httpErr := err.(*rest.HTTPError)
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
		if exists {
			span.SetStatus(codes.Error, "Catalog name exists")
			httpErr := rest.NewHTTPError(ctx, http.StatusConflict, verrors.VegaBackend_Catalog_NameExists)
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
		req.IfNameModify = true
	}

	if err := r.cs.Update(ctx, id, &req); err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	audit.NewInfoLog(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
		interfaces.GenerateCatalogAuditObject(id, req.Name), "")

	logger.Debug("Handler UpdateCatalog Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusNoContent)
	rest.ReplyOK(c, http.StatusNoContent, nil)
}

// ========== DeleteCatalogs ==========

// DeleteCatalogsByEx handles DELETE /api/vega-backend/v1/catalogs/:ids (External)
func (r *restHandler) DeleteCatalogsByEx(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"DeleteCatalogsByEx", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 外网接口：校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.deleteCatalogs(c, ctx, span, visitor)
}

// DeleteCatalogsByIn handles DELETE /api/vega-backend/in/v1/catalogs/:ids (Internal)
func (r *restHandler) DeleteCatalogsByIn(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"DeleteCatalogsByIn", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 内网接口：user_id从header中取
	visitor := GenerateVisitor(c)
	r.deleteCatalogs(c, ctx, span, visitor)
}

// deleteCatalogs is the shared implementation
func (r *restHandler) deleteCatalogs(c *gin.Context, ctx context.Context, span trace.Span, visitor hydra.Visitor) {
	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	ids := strings.Split(c.Param("ids"), ",")

	// Check if ids exists
	for _, id := range ids {
		exists, err := r.cs.CheckExistByID(ctx, id)
		if err != nil {
			httpErr := err.(*rest.HTTPError)
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
		if !exists {
			httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Catalog_NotFound).
				WithErrorDetails(fmt.Sprintf("id %s not found", id))
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}

		// check if catalog discover tasks exists
		exists, err = r.dts.CheckExistByStatuses(ctx, id, []string{interfaces.DiscoverTaskStatusPending, interfaces.DiscoverTaskStatusRunning})
		if err != nil {
			httpErr := err.(*rest.HTTPError)
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
		if exists {
			httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Catalog_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("catalog %s contains tasks in the pending or running statuses and cannot be deleted.", id))
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}

		// check if catalog resources exists
		exists, err = r.rs.CheckExistByCategories(ctx, id, []string{interfaces.ResourceCategoryDataset, interfaces.ResourceCategoryLogicView})
		if err != nil {
			httpErr := err.(*rest.HTTPError)
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
		if exists {
			httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Catalog_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("catalog %s contains data from dataset or logicview class resources and cannot be deleted.", id))
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}

	}

	if err := r.cs.DeleteByIDs(ctx, ids); err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	for _, id := range ids {
		audit.NewWarnLog(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
			interfaces.GenerateCatalogAuditObject(id, ""), audit.SUCCESS, "")
	}

	logger.Debug("Handler DeleteCatalog Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusNoContent)
	rest.ReplyOK(c, http.StatusNoContent, nil)
}

// ========== GetCatalogHealthStatus ==========

// GetCatalogHealthStatusByEx handles GET /api/vega-backend/v1/catalogs/:ids/health-status (External)
func (r *restHandler) GetCatalogHealthStatusByEx(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"GetCatalogHealthStatusByEx", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 外网接口：校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.getCatalogHealthStatus(c, ctx, span, visitor)
}

// GetCatalogHealthStatusByIn handles GET /api/vega-backend/in/v1/catalogs/:ids/health-status (Internal)
func (r *restHandler) GetCatalogHealthStatusByIn(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"GetCatalogHealthStatusByIn", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 内网接口：user_id从header中取
	visitor := GenerateVisitor(c)
	r.getCatalogHealthStatus(c, ctx, span, visitor)
}

// getCatalogHealthStatus is the shared implementation
func (r *restHandler) getCatalogHealthStatus(c *gin.Context, ctx context.Context, span trace.Span, visitor hydra.Visitor) {
	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	id := c.Param("ids")

	catalog, err := r.cs.GetByID(ctx, id, false)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	result := map[string]any{
		"id":                  catalog.ID,
		"health_check_status": catalog.HealthCheckStatus,
		"last_check_time":     catalog.LastCheckTime,
		"health_check_result": catalog.HealthCheckResult,
	}

	logger.Debug("Handler GetCatalogsHealthStatus Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, result)
}

// ========== TestConnection ==========

// TestConnectionByEx handles POST /api/vega-backend/v1/catalogs/:id/test-connection (External)
func (r *restHandler) TestConnectionByEx(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"TestConnectionByEx", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 外网接口：校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.testConnection(c, ctx, span, visitor)
}

// TestConnectionByIn handles POST /api/vega-backend/in/v1/catalogs/:id/test-connection (Internal)
func (r *restHandler) TestConnectionByIn(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"TestConnectionByIn", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 内网接口：user_id从header中取
	visitor := GenerateVisitor(c)
	r.testConnection(c, ctx, span, visitor)
}

// testConnection is the shared implementation
func (r *restHandler) testConnection(c *gin.Context, ctx context.Context, span trace.Span, visitor hydra.Visitor) {
	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	id := c.Param("id")

	// Check if id exists
	catalog, err := r.cs.GetByID(ctx, id, false)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	status, err := r.cs.TestConnection(ctx, catalog)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 映射缓存的健康状态为对外契约：
	// 严格 healthy = success=true，其它（degraded / unhealthy / offline / disabled）= false。
	result := map[string]any{
		"success": status.HealthCheckStatus == interfaces.CatalogHealthStatusHealthy,
		"message": status.HealthCheckResult,
	}

	logger.Debug("Handler TestConnection Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, result)
}

// ========== DiscoverCatalogResources ==========

// DiscoverCatalogResourcesByEx handles POST /api/vega-backend/v1/catalogs/:id/discover (External)
func (r *restHandler) DiscoverCatalogResourcesByEx(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"DiscoverCatalogResourcesByEx", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 外网接口：校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.discoverCatalogResources(c, ctx, span, visitor)
}

// DiscoverCatalogResourcesByIn handles POST /api/vega-backend/in/v1/catalogs/:id/discover (Internal)
func (r *restHandler) DiscoverCatalogResourcesByIn(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"DiscoverCatalogResourcesByIn", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 内网接口：user_id从header中取
	visitor := GenerateVisitor(c)
	r.discoverCatalogResources(c, ctx, span, visitor)
}

// discoverCatalogResources is the shared implementation
func (r *restHandler) discoverCatalogResources(c *gin.Context, ctx context.Context, span trace.Span, visitor hydra.Visitor) {
	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	id := c.Param("id")

	// Get catalog to verify it exists
	catalog, err := r.cs.GetByID(ctx, id, false)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if catalog == nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Catalog_NotFound)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// Create discover task (async)
	taskID, err := r.dts.Create(ctx, catalog.ID)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Catalog_InternalError).
			WithErrorDetails(err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	result := map[string]any{
		"id": taskID,
	}

	logger.Debug("Handler DiscoverCatalogResources Success - Task Created")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, result)
}

// ========== ListCatalogSrcs ==========

// ListCatalogSrcsByEx catalog resource list (External)
func (r *restHandler) ListCatalogSrcsByEx(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"ListCatalogSrcsByEx", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 外网接口：校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.listCatalogSrcs(c, ctx, span, visitor)
}

// listCatalogSrcs is the shared implementation
func (r *restHandler) listCatalogSrcs(c *gin.Context, ctx context.Context, span trace.Span, visitor hydra.Visitor) {
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
		offset, limit, sort, direction, interfaces.CATALOG_SORT)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails))
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	params := interfaces.ListCatalogsQueryParams{
		PaginationQueryParams: pageParam,
		ID:                    id,
		Keyword:               keyword,
	}

	entries, total, err := r.cs.ListCatalogSrcs(ctx, params)
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

	logger.Debug("Handler ListCatalogSrcs Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, result)
}

// ========== ListScheduledDiscoverTasks ==========

// ListScheduledDiscoverTasksByEx handles GET /api/vega-backend/v1/catalogs/:id/scheduled-discover (External)
func (r *restHandler) ListScheduledDiscoverTasksByEx(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"ListScheduledDiscoverTasksByEx", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 外网接口：校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.listScheduledDiscoverTasks(c, ctx, span, visitor)
}

// ListScheduledDiscoverTasksByIn handles GET /api/vega-backend/in/v1/catalogs/:id/scheduled-discover (Internal)
func (r *restHandler) ListScheduledDiscoverTasksByIn(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"ListScheduledDiscoverTasksByIn", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 内网接口：user_id从header中取
	visitor := GenerateVisitor(c)
	r.listScheduledDiscoverTasks(c, ctx, span, visitor)
}

// listScheduledDiscoverTasks is the shared implementation
func (r *restHandler) listScheduledDiscoverTasks(c *gin.Context, ctx context.Context, span trace.Span, visitor hydra.Visitor) {
	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	catalogID := c.Query("id")

	// Verify catalog exists if catalogID is provided
	if catalogID != "" {
		catalog, err := r.cs.GetByID(ctx, catalogID, false)
		if err != nil {
			httpErr := err.(*rest.HTTPError)
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
		if catalog == nil {
			httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Catalog_NotFound)
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
	}

	// Get query parameters
	enabledStr := c.Query("enabled")
	var enabled *bool
	if enabledStr != "" {
		enabledValue := enabledStr == "true"
		enabled = &enabledValue
	}

	offset := common.GetQueryOrDefault(c, "offset", interfaces.DEFAULT_OFFSET)
	limit := common.GetQueryOrDefault(c, "limit", interfaces.DEFAULT_LIMIT)
	sort := common.GetQueryOrDefault(c, "sort", "create_time")
	direction := common.GetQueryOrDefault(c, "direction", interfaces.DESC_DIRECTION)

	// Validate pagination parameters
	pageParam, err := validatePaginationQueryParams(ctx,
		offset, limit, sort, direction, interfaces.CATALOG_SORT)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails))
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	params := interfaces.ScheduledDiscoverTaskQueryParams{
		PaginationQueryParams: pageParam,
		Enabled:               enabled,
	}
	// 如果catalogID不为空，则添加过滤条件
	if catalogID != "" {
		params.CatalogID = catalogID
	}

	tasks, total, err := r.sdtService.List(ctx, params)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Catalog_InternalError).
			WithErrorDetails(err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	result := map[string]any{
		"entries":     tasks,
		"total_count": total,
	}

	logger.Debug("Handler ListScheduledDiscoverTasks Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, result)
}

// ========== ScheduledDiscoverCatalogResources ==========

// ScheduledDiscoverCatalogResourcesByEx handles POST /api/vega-backend/v1/catalogs/:id/scheduled-discover (External)
func (r *restHandler) ScheduledDiscoverCatalogResourcesByEx(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"ScheduledDiscoverCatalogResourcesByEx", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 外网接口：校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.scheduledDiscoverCatalogResources(c, ctx, span, visitor)
}

// ScheduledDiscoverCatalogResourcesByIn handles POST /api/vega-backend/in/v1/catalogs/:id/scheduled-discover (Internal)
func (r *restHandler) ScheduledDiscoverCatalogResourcesByIn(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"ScheduledDiscoverCatalogResourcesByIn", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 内网接口：user_id从header中取
	visitor := GenerateVisitor(c)
	r.scheduledDiscoverCatalogResources(c, ctx, span, visitor)
}

// scheduledDiscoverCatalogResources is the shared implementation
func (r *restHandler) scheduledDiscoverCatalogResources(c *gin.Context, ctx context.Context, span trace.Span, visitor hydra.Visitor) {
	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	id := c.Param("id")

	// Get catalog to verify it exists
	catalog, err := r.cs.GetByID(ctx, id, false)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if catalog == nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Catalog_NotFound)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	var req interfaces.ScheduledDiscoverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_RequestBody).
			WithErrorDetails(err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// Validate cron expression
	if req.CronExpr == "" {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_RequestBody).
			WithErrorDetails("cron_expr is required")
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	// 校验cron表达式是否合法
	if _, err := cron.ParseStandard(req.CronExpr); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_RequestBody).
			WithErrorDetails(fmt.Sprintf("invalid cron expression: %v", err))
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	// Validate strategies if provided
	if len(req.Strategies) > 0 {
		if err := validateStrategies(req.Strategies); err != nil {
			httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_RequestBody).
				WithErrorDetails(err.Error())
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
	}
	// Create scheduled discover task
	task := &interfaces.ScheduledDiscoverTask{
		CatalogID:  catalog.ID,
		CronExpr:   req.CronExpr,
		StartTime:  req.StartTime,
		EndTime:    req.EndTime,
		Strategies: req.Strategies,
		Enabled:    true,
	}
	// Create scheduled discover task
	scheduleID, err := r.sdtService.Create(ctx, task)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Catalog_InternalError).
			WithErrorDetails(err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// Schedule the task in the scheduler
	if err := r.scheduler.ScheduleTask(scheduleID); err != nil {
		logger.Errorf("Failed to schedule task %s: %v", scheduleID, err)
		// Note: We don't return error here because the task is already created in database
		// The scheduler will pick it up on next reload
	}

	result := map[string]any{
		"schedule_id": scheduleID,
		"message":     "Scheduled discover task created successfully",
	}

	logger.Debug("Handler ScheduledDiscoverCatalogResources Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, result)
}

// StopScheduledDiscoverTaskByEx handles POST /api/vega-backend/v1/catalogs/:id/scheduled-discover/:task_id/stop (External)
func (r *restHandler) StopScheduledDiscoverTaskByEx(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"StopScheduledDiscoverTaskByEx", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 外网接口：校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.stopScheduledDiscoverTask(c, ctx, span, visitor)
}

// StopScheduledDiscoverTaskByIn handles POST /api/vega-backend/in/v1/catalogs/:id/scheduled-discover/:task_id/stop (Internal)
func (r *restHandler) StopScheduledDiscoverTaskByIn(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"StopScheduledDiscoverTaskByIn", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 内网接口：user_id从header中取
	visitor := GenerateVisitor(c)
	r.stopScheduledDiscoverTask(c, ctx, span, visitor)
}

// StartScheduledDiscoverTaskByEx handles POST /api/vega-backend/v1/catalogs/:id/scheduled-discover/:task_id/start (External)
func (r *restHandler) StartScheduledDiscoverTaskByEx(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"StartScheduledDiscoverTaskByEx", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 外网接口：校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.startScheduledDiscoverTask(c, ctx, span, visitor)
}

// StartScheduledDiscoverTaskByIn handles POST /api/vega-backend/in/v1/catalogs/:id/scheduled-discover/:task_id/start (Internal)
func (r *restHandler) StartScheduledDiscoverTaskByIn(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"StartScheduledDiscoverTaskByIn", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 内网接口：user_id从header中取
	visitor := GenerateVisitor(c)
	r.startScheduledDiscoverTask(c, ctx, span, visitor)
}

// stopScheduledDiscoverTask is the shared implementation
func (r *restHandler) stopScheduledDiscoverTask(c *gin.Context, ctx context.Context, span trace.Span, visitor hydra.Visitor) {
	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	catalogID := c.Param("id")
	taskID := c.Param("task_id")

	// Verify catalog exists
	catalog, err := r.cs.GetByID(ctx, catalogID, false)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if catalog == nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Catalog_NotFound)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// Verify task exists and belongs to the catalog
	task, err := r.sdtService.GetByID(ctx, taskID)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Catalog_InternalError).
			WithErrorDetails(err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if task == nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Catalog_NotFound).
			WithErrorDetails("Scheduled discover task not found")
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if task.CatalogID != catalogID {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_RequestBody).
			WithErrorDetails("Task does not belong to the specified catalog")
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// Disable the task
	if err := r.sdtService.Disable(ctx, taskID); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Catalog_InternalError).
			WithErrorDetails(err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// Unschedule the task from scheduler
	if err := r.scheduler.UnscheduleTask(taskID); err != nil {
		logger.Errorf("Failed to unschedule task %s: %v", taskID, err)
		// Note: We don't return error here because the task is already disabled in database
	}

	result := map[string]any{
		"schedule_id": taskID,
		"message":     "Scheduled discover task stopped successfully",
	}

	logger.Debug("Handler StopScheduledDiscoverTask Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, result)
}

// startScheduledDiscoverTask is the shared implementation
func (r *restHandler) startScheduledDiscoverTask(c *gin.Context, ctx context.Context, span trace.Span, visitor hydra.Visitor) {
	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	catalogID := c.Param("id")
	taskID := c.Param("task_id")

	// Verify catalog exists
	catalog, err := r.cs.GetByID(ctx, catalogID, false)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if catalog == nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Catalog_NotFound)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// Verify task exists and belongs to the catalog
	task, err := r.sdtService.GetByID(ctx, taskID)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Catalog_InternalError).
			WithErrorDetails(err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if task == nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Catalog_NotFound).
			WithErrorDetails("Scheduled discover task not found")
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if task.CatalogID != catalogID {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_RequestBody).
			WithErrorDetails("Task does not belong to the specified catalog")
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// Check if task is already enabled
	if task.Enabled {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_RequestBody).
			WithErrorDetails("Task is already enabled")
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// Enable the task
	if err := r.sdtService.Enable(ctx, taskID); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Catalog_InternalError).
			WithErrorDetails(err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// Schedule the task in the scheduler
	if err := r.scheduler.ScheduleTask(taskID); err != nil {
		logger.Errorf("Failed to schedule task %s: %v", taskID, err)
		// Note: We don't return error here because the task is already enabled in database
		// The scheduler will pick it up on next reload
	}

	result := map[string]any{
		"schedule_id": taskID,
		"message":     "Scheduled discover task started successfully",
	}

	logger.Debug("Handler StartScheduledDiscoverTask Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, result)
}

// UpdateScheduledDiscoverTaskByEx handles PUT /api/vega-backend/v1/catalogs/:id/scheduled-discover/:task_id (External)
func (r *restHandler) UpdateScheduledDiscoverTaskByEx(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"UpdateScheduledDiscoverTaskByEx", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 外网接口：校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.updateScheduledDiscoverTask(c, ctx, span, visitor)
}

// UpdateScheduledDiscoverTaskByIn handles PUT /api/vega-backend/in/v1/catalogs/:id/scheduled-discover/:task_id (Internal)
func (r *restHandler) UpdateScheduledDiscoverTaskByIn(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"UpdateScheduledDiscoverTaskByIn", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 内网接口：user_id从header中取
	visitor := GenerateVisitor(c)
	r.updateScheduledDiscoverTask(c, ctx, span, visitor)
}

// updateScheduledDiscoverTask is the shared implementation
func (r *restHandler) updateScheduledDiscoverTask(c *gin.Context, ctx context.Context, span trace.Span, visitor hydra.Visitor) {
	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	catalogID := c.Param("id")
	taskID := c.Param("task_id")

	// Verify catalog exists
	catalog, err := r.cs.GetByID(ctx, catalogID, false)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if catalog == nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Catalog_NotFound)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// Verify task exists and belongs to the catalog
	task, err := r.sdtService.GetByID(ctx, taskID)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Catalog_InternalError).
			WithErrorDetails(err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if task == nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Catalog_NotFound).
			WithErrorDetails("Scheduled discover task not found")
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if task.CatalogID != catalogID {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_RequestBody).
			WithErrorDetails("Task does not belong to the specified catalog")
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// Parse request body
	var req interfaces.ScheduledDiscoverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_RequestBody).
			WithErrorDetails(err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// Validate cron expression if provided
	if req.CronExpr != "" {
		if _, err := cron.ParseStandard(req.CronExpr); err != nil {
			httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_RequestBody).
				WithErrorDetails(fmt.Sprintf("invalid cron expression: %v", err))
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
	}

	// Validate strategies if provided
	if len(req.Strategies) > 0 {
		if err := validateStrategies(req.Strategies); err != nil {
			httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_RequestBody).
				WithErrorDetails(err.Error())
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
	}

	// Build update request
	updateTask := &interfaces.ScheduledDiscoverTask{
		ID:         taskID,
		CatalogID:  catalogID,
		CronExpr:   req.CronExpr,
		StartTime:  req.StartTime,
		EndTime:    req.EndTime,
		Strategies: req.Strategies,
	}

	// Update task in database
	if err := r.sdtService.Update(ctx, taskID, updateTask); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Catalog_InternalError).
			WithErrorDetails(err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// If task is enabled, reschedule it with new cron expression
	if task.Enabled {
		// Unschedule old task
		if err := r.scheduler.UnscheduleTask(taskID); err != nil {
			logger.Errorf("Failed to unschedule task %s: %v", taskID, err)
		}
		// Schedule with new configuration
		if err := r.scheduler.ScheduleTask(taskID); err != nil {
			logger.Errorf("Failed to reschedule task %s: %v", taskID, err)
		}
	}

	result := map[string]any{
		"task_id": taskID,
		"message": "Scheduled discover task updated successfully",
	}

	logger.Debug("Handler UpdateScheduledDiscoverTask Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, result)
}

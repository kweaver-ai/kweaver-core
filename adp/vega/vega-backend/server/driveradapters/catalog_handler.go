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
	"vega-backend/common/visitor"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-go-lib/audit"
	"github.com/kweaver-ai/kweaver-go-lib/hydra"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/otel/otellog"
	"github.com/kweaver-ai/kweaver-go-lib/otel/oteltrace"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"go.opentelemetry.io/otel/codes"

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
	// 外网接口：校验token
	visitor, err := r.verifyOAuth(rest.GetLanguageCtx(c), c)
	if err != nil {
		return
	}
	r.listCatalogs(c, visitor)
}

// ListCatalogsByIn handles GET /api/vega-backend/in/v1/catalogs (Internal)
func (r *restHandler) ListCatalogsByIn(c *gin.Context) {
	// 内网接口：user_id从header中取
	visitor := visitor.GenerateVisitor(c)
	r.listCatalogs(c, visitor)
}

// listCatalogs is the shared implementation
func (r *restHandler) listCatalogs(c *gin.Context, visitor hydra.Visitor) {
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	oteltrace.AddHttpAttrs4API(span, oteltrace.GetAttrsByGinCtx(c))

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
		otellog.LogError(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails), nil)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	extKeys := c.QueryArray("extension_key")
	extVals := c.QueryArray("extension_value")
	includeExt := strings.EqualFold(strings.TrimSpace(c.Query("include_extensions")), "true")
	includeExtKeys := strings.TrimSpace(c.Query("include_extension_keys"))

	params := interfaces.CatalogsQueryParams{
		PaginationQueryParams: pageParam,
		Tag:                   tag,
		Type:                  typ,
		HealthCheckStatus:     healthCheckStatus,
		ExtensionKeys:         extKeys,
		ExtensionValues:       extVals,
		IncludeExtensions:     includeExt,
		IncludeExtensionKeys:  includeExtKeys,
	}

	if err := ValidateCatalogListQueryParams(ctx, params); err != nil {
		httpErr := err.(*rest.HTTPError)
		otellog.LogError(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails), nil)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	entries, total, err := r.cs.List(ctx, params)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	result := map[string]any{
		"entries":     entries,
		"total_count": total,
	}

	logger.Debug("Handler ListCatalogs Success")
	oteltrace.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, result)
}

// ========== CreateCatalog ==========

// CreateCatalogByEx handles POST /api/vega-backend/v1/catalogs (External)
func (r *restHandler) CreateCatalogByEx(c *gin.Context) {
	// 外网接口：校验token
	visitor, err := r.verifyOAuth(rest.GetLanguageCtx(c), c)
	if err != nil {
		return
	}
	r.createCatalog(c, visitor)
}

// CreateCatalogByIn handles POST /api/vega-backend/in/v1/catalogs (Internal)
func (r *restHandler) CreateCatalogByIn(c *gin.Context) {
	// 内网接口：user_id从header中取
	visitor := visitor.GenerateVisitor(c)
	r.createCatalog(c, visitor)
}

// createCatalog is the shared implementation
func (r *restHandler) createCatalog(c *gin.Context, visitor hydra.Visitor) {
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	oteltrace.AddHttpAttrs4API(span, oteltrace.GetAttrsByGinCtx(c))

	var req interfaces.CatalogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_RequestBody).
			WithErrorDetails(err.Error())
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	if err := ValidateCatalogRequest(ctx, &req); err != nil {
		httpErr := err.(*rest.HTTPError)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// Check if name exists
	exists, err := r.cs.CheckExistByName(ctx, req.Name)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Catalog_InternalError).
			WithErrorDetails(err.Error())
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if exists {
		httpErr := rest.NewHTTPError(ctx, http.StatusConflict, verrors.VegaBackend_Catalog_NameExists)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// Check if id exists if provided
	if req.ID != "" {
		exists, err := r.cs.CheckExistByID(ctx, req.ID)
		if err != nil {
			httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Catalog_InternalError).
				WithErrorDetails(err.Error())
			oteltrace.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
		if exists {
			httpErr := rest.NewHTTPError(ctx, http.StatusConflict, verrors.VegaBackend_Catalog_IDExists).
				WithErrorDetails(fmt.Sprintf("id %s already exists", req.ID))
			oteltrace.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
	}

	id, err := r.cs.Create(ctx, &req)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 成功创建记录审计日志
	audit.NewInfoLog(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
		interfaces.GenerateCatalogAuditObject(id, req.Name), "")

	result := map[string]any{"id": id}

	logger.Debug("Handler CreateCatalog Success")
	oteltrace.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusCreated, result)
}

// ========== GetCatalogs ==========

// GetCatalogsByEx handles GET /api/vega-backend/v1/catalogs/:ids (External)
func (r *restHandler) GetCatalogsByEx(c *gin.Context) {
	// 外网接口：校验token
	visitor, err := r.verifyOAuth(rest.GetLanguageCtx(c), c)
	if err != nil {
		return
	}
	r.getCatalogs(c, visitor)
}

// GetCatalogsByIn handles GET /api/vega-backend/in/v1/catalogs/:ids (Internal)
func (r *restHandler) GetCatalogsByIn(c *gin.Context) {
	// 内网接口：user_id从header中取
	visitor := visitor.GenerateVisitor(c)
	r.getCatalogs(c, visitor)
}

// getCatalogs is the shared implementation
func (r *restHandler) getCatalogs(c *gin.Context, visitor hydra.Visitor) {
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	oteltrace.AddHttpAttrs4API(span, oteltrace.GetAttrsByGinCtx(c))

	ids := strings.Split(c.Param("id"), ",")

	catalogs, err := r.cs.GetByIDs(ctx, ids)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
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
				oteltrace.AddHttpAttrs4HttpError(span, httpErr)
				rest.ReplyError(c, httpErr)
				return
			}
		}
	}

	result := map[string]any{"entries": catalogs}

	logger.Debug("Handler GetCatalogs Success")
	oteltrace.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, result)
}

// ========== UpdateCatalog ==========

// UpdateCatalogByEx handles PUT /api/vega-backend/v1/catalogs/:id (External)
func (r *restHandler) UpdateCatalogByEx(c *gin.Context) {
	// 外网接口：校验token
	visitor, err := r.verifyOAuth(rest.GetLanguageCtx(c), c)
	if err != nil {
		return
	}
	r.updateCatalog(c, visitor)
}

// UpdateCatalogByIn handles PUT /api/vega-backend/in/v1/catalogs/:id (Internal)
func (r *restHandler) UpdateCatalogByIn(c *gin.Context) {
	// 内网接口：user_id从header中取
	visitor := visitor.GenerateVisitor(c)
	r.updateCatalog(c, visitor)
}

// updateCatalog is the shared implementation
func (r *restHandler) updateCatalog(c *gin.Context, visitor hydra.Visitor) {
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	oteltrace.AddHttpAttrs4API(span, oteltrace.GetAttrsByGinCtx(c))

	id := c.Param("id")

	var req interfaces.CatalogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_RequestBody).
			WithErrorDetails(err.Error())
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	if req.ID == "" {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Catalog_InvalidParameter_ID).
			WithErrorDetails("body field 'id' is required and must equal path parameter")
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if req.ID != id {
		httpErr := rest.NewHTTPError(ctx, http.StatusConflict, verrors.VegaBackend_Catalog_IDMismatch).
			WithErrorDetails(fmt.Sprintf("path id %q != body id %q", id, req.ID))
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	if err := ValidateCatalogRequest(ctx, &req); err != nil {
		httpErr := err.(*rest.HTTPError)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// Check if id exists
	catalog, err := r.cs.GetByID(ctx, id, false)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
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
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
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
					oteltrace.AddHttpAttrs4HttpError(span, httpErr)
					rest.ReplyError(c, httpErr)
					return
				}
			} else {
				// Field exists in catalog but not in request - cannot remove immutable fields
				span.SetStatus(codes.Error, fmt.Sprintf("Connector config field %s cannot be removed", field))
				httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Catalog_InvalidParameter_ConnectorConfig).
					WithErrorDetails(fmt.Sprintf("connector_config.%s cannot be removed", field))
				oteltrace.AddHttpAttrs4HttpError(span, httpErr)
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
			oteltrace.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
		if exists {
			span.SetStatus(codes.Error, "Catalog name exists")
			httpErr := rest.NewHTTPError(ctx, http.StatusConflict, verrors.VegaBackend_Catalog_NameExists)
			oteltrace.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
		req.IfNameModify = true
	}

	if err := r.cs.Update(ctx, id, &req); err != nil {
		httpErr := err.(*rest.HTTPError)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	audit.NewInfoLog(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
		interfaces.GenerateCatalogAuditObject(id, req.Name), "")

	logger.Debug("Handler UpdateCatalog Success")
	oteltrace.AddHttpAttrs4Ok(span, http.StatusNoContent)
	rest.ReplyOK(c, http.StatusNoContent, nil)
}

// ========== DeleteCatalogs ==========

// DeleteCatalogsByEx handles DELETE /api/vega-backend/v1/catalogs/:ids (External)
func (r *restHandler) DeleteCatalogsByEx(c *gin.Context) {
	// 外网接口：校验token
	visitor, err := r.verifyOAuth(rest.GetLanguageCtx(c), c)
	if err != nil {
		return
	}
	r.deleteCatalogs(c, visitor)
}

// DeleteCatalogsByIn handles DELETE /api/vega-backend/in/v1/catalogs/:ids (Internal)
func (r *restHandler) DeleteCatalogsByIn(c *gin.Context) {
	// 内网接口：user_id从header中取
	visitor := visitor.GenerateVisitor(c)
	r.deleteCatalogs(c, visitor)
}

// deleteCatalogs is the shared implementation
func (r *restHandler) deleteCatalogs(c *gin.Context, visitor hydra.Visitor) {
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	oteltrace.AddHttpAttrs4API(span, oteltrace.GetAttrsByGinCtx(c))

	ids := strings.Split(c.Param("id"), ",")

	// Check if ids exists
	for _, id := range ids {
		exists, err := r.cs.CheckExistByID(ctx, id)
		if err != nil {
			httpErr := err.(*rest.HTTPError)
			oteltrace.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
		if !exists {
			httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Catalog_NotFound).
				WithErrorDetails(fmt.Sprintf("id %s not found", id))
			oteltrace.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}

		// check if catalog discover tasks exists
		exists, err = r.dts.CheckExistByStatuses(ctx, id, []string{interfaces.DiscoverTaskStatusPending, interfaces.DiscoverTaskStatusRunning})
		if err != nil {
			httpErr := err.(*rest.HTTPError)
			oteltrace.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
		if exists {
			httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Catalog_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("catalog %s contains tasks in the pending or running statuses and cannot be deleted.", id))
			oteltrace.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}

		// check if catalog resources exists
		exists, err = r.rs.CheckExistByCategories(ctx, id, []string{interfaces.ResourceCategoryDataset, interfaces.ResourceCategoryLogicView})
		if err != nil {
			httpErr := err.(*rest.HTTPError)
			oteltrace.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
		if exists {
			httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Catalog_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("catalog %s contains data from dataset or logicview class resources and cannot be deleted.", id))
			oteltrace.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}

	}

	if err := r.cs.DeleteByIDs(ctx, ids); err != nil {
		httpErr := err.(*rest.HTTPError)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	for _, id := range ids {
		audit.NewWarnLog(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
			interfaces.GenerateCatalogAuditObject(id, ""), audit.SUCCESS, "")
	}

	logger.Debug("Handler DeleteCatalog Success")
	oteltrace.AddHttpAttrs4Ok(span, http.StatusNoContent)
	rest.ReplyOK(c, http.StatusNoContent, nil)
}

// ========== GetCatalogHealthStatus ==========

// GetCatalogHealthStatusByEx handles GET /api/vega-backend/v1/catalogs/:id/health-status (External)
func (r *restHandler) GetCatalogHealthStatusByEx(c *gin.Context) {
	// 外网接口：校验token
	visitor, err := r.verifyOAuth(rest.GetLanguageCtx(c), c)
	if err != nil {
		return
	}
	r.getCatalogHealthStatus(c, visitor)
}

// GetCatalogHealthStatusByIn handles GET /api/vega-backend/in/v1/catalogs/:id/health-status (Internal)
func (r *restHandler) GetCatalogHealthStatusByIn(c *gin.Context) {
	// 内网接口：user_id从header中取
	visitor := visitor.GenerateVisitor(c)
	r.getCatalogHealthStatus(c, visitor)
}

// getCatalogHealthStatus is the shared implementation
func (r *restHandler) getCatalogHealthStatus(c *gin.Context, visitor hydra.Visitor) {
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	oteltrace.AddHttpAttrs4API(span, oteltrace.GetAttrsByGinCtx(c))

	id := c.Param("id")

	catalog, err := r.cs.GetByID(ctx, id, false)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
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
	oteltrace.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, result)
}

// ========== TestConnection ==========

// TestConnectionByEx handles POST /api/vega-backend/v1/catalogs/:id/test-connection (External)
func (r *restHandler) TestConnectionByEx(c *gin.Context) {
	// 外网接口：校验token
	visitor, err := r.verifyOAuth(rest.GetLanguageCtx(c), c)
	if err != nil {
		return
	}
	r.testConnection(c, visitor)
}

// TestConnectionByIn handles POST /api/vega-backend/in/v1/catalogs/:id/test-connection (Internal)
func (r *restHandler) TestConnectionByIn(c *gin.Context) {
	// 内网接口：user_id从header中取
	visitor := visitor.GenerateVisitor(c)
	r.testConnection(c, visitor)
}

// testConnection is the shared implementation
func (r *restHandler) testConnection(c *gin.Context, visitor hydra.Visitor) {
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	oteltrace.AddHttpAttrs4API(span, oteltrace.GetAttrsByGinCtx(c))

	id := c.Param("id")

	// Check if id exists
	catalog, err := r.cs.GetByID(ctx, id, false)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	status, err := r.cs.TestConnection(ctx, catalog)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
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
	oteltrace.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, result)
}

// ========== DiscoverCatalogResources ==========

// DiscoverCatalogResourcesByEx handles POST /api/vega-backend/v1/catalogs/:id/discover (External)
func (r *restHandler) DiscoverCatalogResourcesByEx(c *gin.Context) {
	// 外网接口：校验token
	visitor, err := r.verifyOAuth(rest.GetLanguageCtx(c), c)
	if err != nil {
		return
	}
	r.discoverCatalogResources(c, visitor)
}

// DiscoverCatalogResourcesByIn handles POST /api/vega-backend/in/v1/catalogs/:id/discover (Internal)
func (r *restHandler) DiscoverCatalogResourcesByIn(c *gin.Context) {
	// 内网接口：user_id从header中取
	visitor := visitor.GenerateVisitor(c)
	r.discoverCatalogResources(c, visitor)
}

// discoverCatalogResources is the shared implementation
func (r *restHandler) discoverCatalogResources(c *gin.Context, visitor hydra.Visitor) {
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	oteltrace.AddHttpAttrs4API(span, oteltrace.GetAttrsByGinCtx(c))

	id := c.Param("id")

	// Get catalog to verify it exists
	catalog, err := r.cs.GetByID(ctx, id, false)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if catalog == nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Catalog_NotFound)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// Create discover task (async)
	taskID, err := r.dts.Create(ctx, catalog.ID)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Catalog_InternalError).
			WithErrorDetails(err.Error())
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	result := map[string]any{
		"id": taskID,
	}

	logger.Debug("Handler DiscoverCatalogResources Success - Task Created")
	oteltrace.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, result)
}

// ========== ListCatalogSrcs ==========

// ListCatalogSrcsByEx catalog resource list (External)
func (r *restHandler) ListCatalogSrcsByEx(c *gin.Context) {
	// 外网接口：校验token
	visitor, err := r.verifyOAuth(rest.GetLanguageCtx(c), c)
	if err != nil {
		return
	}
	r.listCatalogSrcs(c, visitor)
}

// listCatalogSrcs is the shared implementation
func (r *restHandler) listCatalogSrcs(c *gin.Context, visitor hydra.Visitor) {
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	oteltrace.AddHttpAttrs4API(span, oteltrace.GetAttrsByGinCtx(c))

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
		otellog.LogError(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails), nil)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
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
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	result := map[string]any{
		"entries":     entries,
		"total_count": total,
	}

	logger.Debug("Handler ListCatalogSrcs Success")
	oteltrace.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, result)
}

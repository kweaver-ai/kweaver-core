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

// ========== ListResources ==========

// ListResourcesByEx handles GET /api/vega-backend/v1/resources (External)
func (r *restHandler) ListResourcesByEx(c *gin.Context) {
	// 外网接口：校验token
	visitor, err := r.verifyOAuth(rest.GetLanguageCtx(c), c)
	if err != nil {
		return
	}
	r.listResources(c, visitor)
}

// ListResourcesByIn handles GET /api/vega-backend/in/v1/resources (Internal)
func (r *restHandler) ListResourcesByIn(c *gin.Context) {
	// 内网接口：user_id从header中取
	visitor := visitor.GenerateVisitor(c)
	r.listResources(c, visitor)
}

// listResources is the shared implementation
func (r *restHandler) listResources(c *gin.Context, visitor hydra.Visitor) {
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	oteltrace.AddHttpAttrs4API(span, oteltrace.GetAttrsByGinCtx(c))

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

	params := interfaces.ResourcesQueryParams{
		PaginationQueryParams: pageParam,
		CatalogID:             catalogID,
		Category:              category,
		Status:                status,
		Database:              database,
		ExtensionKeys:         extKeys,
		ExtensionValues:       extVals,
		IncludeExtensions:     includeExt,
		IncludeExtensionKeys:  includeExtKeys,
	}

	if err := ValidateResourceListQueryParams(ctx, params); err != nil {
		httpErr := err.(*rest.HTTPError)
		otellog.LogError(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails), nil)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	entries, total, err := r.rs.List(ctx, params)
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

	logger.Debug("Handler ListResources Success")
	oteltrace.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, result)
}

// ========== CreateResource ==========

// CreateResourceByEx handles POST /api/vega-backend/v1/resources (External)
func (r *restHandler) CreateResourceByEx(c *gin.Context) {
	// 外网接口：校验token
	visitor, err := r.verifyOAuth(rest.GetLanguageCtx(c), c)
	if err != nil {
		return
	}
	r.createResource(c, visitor)
}

// CreateResourceByIn handles POST /api/vega-backend/in/v1/resources (Internal)
func (r *restHandler) CreateResourceByIn(c *gin.Context) {
	// 内网接口：user_id从header中取
	visitor := visitor.GenerateVisitor(c)
	r.createResource(c, visitor)
}

// createResource is the shared implementation
func (r *restHandler) createResource(c *gin.Context, visitor hydra.Visitor) {
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	oteltrace.AddHttpAttrs4API(span, oteltrace.GetAttrsByGinCtx(c))

	var req interfaces.ResourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest,
			verrors.VegaBackend_InvalidParameter_RequestBody).WithErrorDetails(err.Error())
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	if err := ValidateResourceRequest(ctx, &req); err != nil {
		httpErr := err.(*rest.HTTPError)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// Check catelog exists
	csExists, csErr := r.cs.CheckExistByID(ctx, req.CatalogID)
	if csErr != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError,
			verrors.VegaBackend_Resource_InternalError).WithErrorDetails(csErr.Error())
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if !csExists {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Resource_CatalogNotFound).
			WithErrorDetails(fmt.Sprintf("catalog %s not found", req.CatalogID))
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// Check if name exists
	exists, err := r.rs.CheckExistByName(ctx, req.CatalogID, req.Name)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError,
			verrors.VegaBackend_Resource_InternalError).WithErrorDetails(err.Error())
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if exists {
		httpErr := rest.NewHTTPError(ctx, http.StatusConflict, verrors.VegaBackend_Resource_NameExists)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// Check if id exists if provided
	if req.ID != "" {
		exists, err := r.rs.CheckExistByID(ctx, req.ID)
		if err != nil {
			httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError,
				verrors.VegaBackend_Resource_InternalError).WithErrorDetails(err.Error())
			oteltrace.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
		if exists {
			httpErr := rest.NewHTTPError(ctx, http.StatusConflict, verrors.VegaBackend_Resource_IDExists).
				WithErrorDetails(fmt.Sprintf("id %s already exists", req.ID))
			oteltrace.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
	}

	resource, err := r.rs.Create(ctx, &req)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 成功创建记录审计日志
	audit.NewInfoLog(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
		interfaces.GenerateResourceAuditObject(resource.ID, req.Name), "")

	result := map[string]any{"id": resource.ID}

	logger.Debug("Handler CreateResource Success")
	oteltrace.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusCreated, result)
}

// ========== GetResources ==========

// GetResourcesByEx handles GET /api/vega-backend/v1/resources/:ids (External)
func (r *restHandler) GetResourcesByEx(c *gin.Context) {
	// 外网接口：校验token
	visitor, err := r.verifyOAuth(rest.GetLanguageCtx(c), c)
	if err != nil {
		return
	}
	r.getResources(c, visitor)
}

// GetResourcesByIn handles GET /api/vega-backend/in/v1/resources/:ids (Internal)
func (r *restHandler) GetResourcesByIn(c *gin.Context) {
	// 内网接口：user_id从header中取
	visitor := visitor.GenerateVisitor(c)
	r.getResources(c, visitor)
}

// getResources is the shared implementation
func (r *restHandler) getResources(c *gin.Context, visitor hydra.Visitor) {
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	oteltrace.AddHttpAttrs4API(span, oteltrace.GetAttrsByGinCtx(c))

	ids := strings.Split(c.Param("id"), ",")

	resources, err := r.rs.GetByIDs(ctx, ids)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
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
				oteltrace.AddHttpAttrs4HttpError(span, httpErr)
				rest.ReplyError(c, httpErr)
				return
			}
		}
	}

	result := map[string]any{"entries": resources}

	logger.Debug("Handler GetResource Success")
	oteltrace.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, result)
}

// ========== UpdateResource ==========

// UpdateResourceByEx handles PUT /api/vega-backend/v1/resources/:id (External)
func (r *restHandler) UpdateResourceByEx(c *gin.Context) {
	// 外网接口：校验token
	visitor, err := r.verifyOAuth(rest.GetLanguageCtx(c), c)
	if err != nil {
		return
	}
	r.updateResource(c, visitor)
}

// UpdateResourceByIn handles PUT /api/vega-backend/in/v1/resources/:id (Internal)
func (r *restHandler) UpdateResourceByIn(c *gin.Context) {
	// 内网接口：user_id从header中取
	visitor := visitor.GenerateVisitor(c)
	r.updateResource(c, visitor)
}

// updateResource is the shared implementation
func (r *restHandler) updateResource(c *gin.Context, visitor hydra.Visitor) {
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	oteltrace.AddHttpAttrs4API(span, oteltrace.GetAttrsByGinCtx(c))

	id := c.Param("id")

	var req interfaces.ResourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest,
			verrors.VegaBackend_InvalidParameter_RequestBody).WithErrorDetails(err.Error())
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	if err := ValidateResourceRequest(ctx, &req); err != nil {
		httpErr := err.(*rest.HTTPError)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// Check if id exists
	resource, err := r.rs.GetByID(ctx, id)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	req.OriginResource = resource

	// Apply updates
	if req.Name != resource.Name {
		exists, err := r.rs.CheckExistByName(ctx, req.CatalogID, req.Name)
		if err != nil {
			httpErr := err.(*rest.HTTPError)
			oteltrace.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
		if exists {
			span.SetStatus(codes.Error, "Resource name exists")
			httpErr := rest.NewHTTPError(ctx, http.StatusConflict, verrors.VegaBackend_Resource_NameExists)
			oteltrace.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
		req.IfNameModify = true
	}

	if err := r.rs.Update(ctx, id, &req); err != nil {
		httpErr := err.(*rest.HTTPError)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	audit.NewInfoLog(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
		interfaces.GenerateResourceAuditObject(id, req.Name), "")

	logger.Debug("Handler UpdateResource Success")
	oteltrace.AddHttpAttrs4Ok(span, http.StatusNoContent)
	rest.ReplyOK(c, http.StatusNoContent, nil)
}

// ========== DeleteResources ==========

// DeleteResourcesByEx handles DELETE /api/vega-backend/v1/resources/:ids (External)
func (r *restHandler) DeleteResourcesByEx(c *gin.Context) {
	// 外网接口：校验token
	visitor, err := r.verifyOAuth(rest.GetLanguageCtx(c), c)
	if err != nil {
		return
	}
	r.deleteResources(c, visitor)
}

// DeleteResourcesByIn handles DELETE /api/vega-backend/in/v1/resources/:ids (Internal)
func (r *restHandler) DeleteResourcesByIn(c *gin.Context) {
	// 内网接口：user_id从header中取
	visitor := visitor.GenerateVisitor(c)
	r.deleteResources(c, visitor)
}

// deleteResources is the shared implementation
func (r *restHandler) deleteResources(c *gin.Context, visitor hydra.Visitor) {
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	oteltrace.AddHttpAttrs4API(span, oteltrace.GetAttrsByGinCtx(c))

	rawIDs := strings.Split(c.Param("id"), ",")
	ignoreMissing := strings.EqualFold(c.Query("ignore_missing"), "true")

	// Pre-validate existence; collect ids to delete based on ignore_missing.
	idsToDelete := make([]string, 0, len(rawIDs))
	for _, id := range rawIDs {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		exists, err := r.rs.CheckExistByID(ctx, id)
		if err != nil {
			httpErr := err.(*rest.HTTPError)
			oteltrace.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
		if !exists {
			if ignoreMissing {
				continue
			}
			httpErr := rest.NewHTTPError(ctx, http.StatusNotFound,
				verrors.VegaBackend_Resource_NotFound).WithErrorDetails(fmt.Sprintf("id %s not found", id))
			oteltrace.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
		idsToDelete = append(idsToDelete, id)
	}

	if len(idsToDelete) > 0 {
		if err := r.rs.DeleteByIDs(ctx, idsToDelete); err != nil {
			httpErr := err.(*rest.HTTPError)
			oteltrace.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
	}

	for _, id := range idsToDelete {
		audit.NewWarnLog(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
			interfaces.GenerateResourceAuditObject(id, ""), audit.SUCCESS, "")
	}

	logger.Debug("Handler DeleteResource Success")
	oteltrace.AddHttpAttrs4Ok(span, http.StatusNoContent)
	rest.ReplyOK(c, http.StatusNoContent, nil)
}

// ========== ListResourceSrcs ==========

// ListResourceSrcsByEx resource source list (External)
func (r *restHandler) ListResourceSrcsByEx(c *gin.Context) {
	// 外网接口：校验token
	visitor, err := r.verifyOAuth(rest.GetLanguageCtx(c), c)
	if err != nil {
		return
	}
	r.listResourceSrcs(c, visitor)
}

// listResourceSrcs is the shared implementation
func (r *restHandler) listResourceSrcs(c *gin.Context, visitor hydra.Visitor) {
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
		offset, limit, sort, direction, interfaces.RESOURCE_SORT)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		otellog.LogError(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails), nil)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
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
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	result := map[string]any{
		"entries":     entries,
		"total_count": total,
	}

	logger.Debug("Handler ListResourceSrcs Success")
	oteltrace.AddHttpAttrs4Ok(span, http.StatusOK)
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

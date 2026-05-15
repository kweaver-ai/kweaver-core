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
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-go-lib/hydra"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/otel/otellog"
	"github.com/kweaver-ai/kweaver-go-lib/otel/oteltrace"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/mitchellh/mapstructure"
	"go.opentelemetry.io/otel/trace"

	"vega-backend/common/visitor"
	verrors "vega-backend/errors"
	"vega-backend/interfaces"
)

// PostResourceDataByEx handles POST /api/vega-backend/v1/resources/:id/data (External).
// Dispatches based on X-HTTP-Method-Override header:
//
//	GET    → query (any category; returns entries + optional total_count)
//	POST   → batch create documents (dataset category only)
//	DELETE → delete documents by filter (dataset category only)
func (r *restHandler) PostResourceDataByEx(c *gin.Context) {
	visitor, err := r.verifyOAuth(rest.GetLanguageCtx(c), c)
	if err != nil {
		return
	}
	r.postResourceData(c, visitor)
}

// PostResourceDataByIn handles POST /api/vega-backend/in/v1/resources/:id/data (Internal).
func (r *restHandler) PostResourceDataByIn(c *gin.Context) {
	visitor := visitor.GenerateVisitor(c)
	r.postResourceData(c, visitor)
}

// postResourceData dispatches POST /resources/:id/data to the right branch based on
// X-HTTP-Method-Override header.
func (r *restHandler) postResourceData(c *gin.Context, visitor hydra.Visitor) {
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	accountInfo := interfaces.AccountInfo{ID: visitor.ID, Type: string(visitor.Type)}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)
	oteltrace.AddHttpAttrs4API(span, oteltrace.GetAttrsByGinCtx(c))

	override := strings.ToUpper(c.GetHeader(interfaces.HTTP_HEADER_METHOD_OVERRIDE))
	switch override {
	case http.MethodGet:
		r.queryResourceData(c, ctx, span)
	case http.MethodPost:
		r.createResourceData(c, ctx, span)
	case http.MethodDelete:
		r.deleteResourceDataByQuery(c, ctx, span)
	default:
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_OverrideMethod)
		otellog.LogError(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails), httpErr)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
	}
}

// queryResourceData handles POST /resources/:id/data + Override: GET.
// Generic resource-data query, supports any category.
func (r *restHandler) queryResourceData(c *gin.Context, ctx context.Context, span trace.Span) {
	start := time.Now()

	resourceID := c.Param("id")

	var params interfaces.ResourceDataQueryParams
	if err := c.ShouldBindJSON(&params); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_RequestBody).
			WithErrorDetails(err.Error())
		otellog.LogError(ctx, "Bind resource data query request failed", httpErr)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	if err := ValidateResourceDataQueryParams(ctx, &params); err != nil {
		httpErr := err.(*rest.HTTPError)
		otellog.LogError(ctx, "Validate resource data query params failed", httpErr)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	resource, err := r.rs.GetByID(ctx, resourceID)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		otellog.LogError(ctx, "Get resource failed", httpErr)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if resource == nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Resource_NotFound)
		otellog.LogError(ctx, "Resource not found", httpErr)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	entries, total, err := r.rds.Query(ctx, resource, &params)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		otellog.LogError(ctx, "Query resource data failed", httpErr)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	resultData := map[string]any{
		"entries": entries,
	}
	if params.NeedTotal {
		resultData["total_count"] = total
	}

	logger.Debug("Handler queryResourceData Success")
	oteltrace.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOkWithHeaders(c, http.StatusOK, resultData, map[string]string{
		interfaces.X_REQUEST_TOOK: time.Since(start).String(),
	})
}

// createResourceData handles POST /resources/:id/data + Override: POST.
// Batch create documents; dataset category only.
func (r *restHandler) createResourceData(c *gin.Context, ctx context.Context, span trace.Span) {
	resource, ok := r.requireDatasetResource(c, ctx, span, c.Param("id"))
	if !ok {
		return
	}

	var documents []map[string]any
	if err := c.ShouldBindJSON(&documents); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_RequestBody).
			WithErrorDetails(err.Error())
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	docIDs, err := r.ds.CreateDocuments(ctx, resource.ID, documents)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	logger.Debug("Handler createResourceData Success")
	oteltrace.AddHttpAttrs4Ok(span, http.StatusCreated)
	rest.ReplyOK(c, http.StatusCreated, map[string]any{"ids": docIDs})
}

// deleteResourceDataByQuery handles POST /resources/:id/data + Override: DELETE.
// Delete documents by filter; dataset category only. Body must carry a non-empty filter.
func (r *restHandler) deleteResourceDataByQuery(c *gin.Context, ctx context.Context, span trace.Span) {
	start := time.Now()

	resource, ok := r.requireDatasetResource(c, ctx, span, c.Param("id"))
	if !ok {
		return
	}

	var params interfaces.ResourceDataQueryParams
	if err := c.ShouldBindJSON(&params); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_RequestBody).
			WithErrorDetails(err.Error())
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	if params.FilterCondition == nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_RequestBody).
			WithErrorDetails("filter is required for delete-by-query (empty filter would delete all documents)")
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	var actualCond *interfaces.FilterCondCfg
	if err := mapstructure.Decode(params.FilterCondition, &actualCond); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_FilterCondition).
			WithErrorDetails(fmt.Sprintf("mapstructure decode filters failed: %s", err.Error()))
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	params.FilterCondCfg = actualCond

	if err := r.ds.DeleteDocumentsByQuery(ctx, resource.ID, resource, &params); err != nil {
		httpErr := err.(*rest.HTTPError)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	logger.Debug("Handler deleteResourceDataByQuery Success")
	oteltrace.AddHttpAttrs4Ok(span, http.StatusNoContent)
	rest.ReplyOkWithHeaders(c, http.StatusNoContent, nil, map[string]string{
		interfaces.X_REQUEST_TOOK: time.Since(start).String(),
	})
}

// =========================== PUT /resources/:id/data ===========================

// PutResourceDataByEx handles PUT /api/vega-backend/v1/resources/:id/data (External).
// Batch upsert documents; dataset category only. Each document must carry an `id` field.
func (r *restHandler) PutResourceDataByEx(c *gin.Context) {
	visitor, err := r.verifyOAuth(rest.GetLanguageCtx(c), c)
	if err != nil {
		return
	}
	r.putResourceData(c, visitor)
}

// PutResourceDataByIn handles PUT /api/vega-backend/in/v1/resources/:id/data (Internal).
func (r *restHandler) PutResourceDataByIn(c *gin.Context) {
	visitor := visitor.GenerateVisitor(c)
	r.putResourceData(c, visitor)
}

func (r *restHandler) putResourceData(c *gin.Context, visitor hydra.Visitor) {
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	accountInfo := interfaces.AccountInfo{ID: visitor.ID, Type: string(visitor.Type)}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)
	oteltrace.AddHttpAttrs4API(span, oteltrace.GetAttrsByGinCtx(c))

	resource, ok := r.requireDatasetResource(c, ctx, span, c.Param("id"))
	if !ok {
		return
	}

	var documents []map[string]any
	if err := c.ShouldBindJSON(&documents); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_RequestBody).
			WithErrorDetails(err.Error())
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if len(documents) == 0 {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_RequestBody).
			WithErrorDetails("documents array cannot be empty")
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// Each document must carry id; collect violators.
	invalidIndexes := make([]int, 0)
	for i, doc := range documents {
		if id, ok := doc["id"].(string); !ok || id == "" {
			invalidIndexes = append(invalidIndexes, i)
		}
	}
	if len(invalidIndexes) > 0 {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_RequestBody).
			WithErrorDetails(map[string]any{
				"message":         "every document must carry a non-empty `id` field for PUT (update)",
				"invalid_indexes": invalidIndexes,
			})
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	docIDs, err := r.ds.UpsertDocuments(ctx, resource.ID, documents)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	logger.Debug("Handler putResourceData Success")
	oteltrace.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, map[string]any{"ids": docIDs})
}

// =========================== GET /resources/:id/data/:doc_id ===========================

// GetResourceDataDocByEx handles GET /api/vega-backend/v1/resources/:id/data/:doc_id (External).
func (r *restHandler) GetResourceDataDocByEx(c *gin.Context) {
	visitor, err := r.verifyOAuth(rest.GetLanguageCtx(c), c)
	if err != nil {
		return
	}
	r.getResourceDataDoc(c, visitor)
}

// GetResourceDataDocByIn handles GET /api/vega-backend/in/v1/resources/:id/data/:doc_id (Internal).
func (r *restHandler) GetResourceDataDocByIn(c *gin.Context) {
	visitor := visitor.GenerateVisitor(c)
	r.getResourceDataDoc(c, visitor)
}

func (r *restHandler) getResourceDataDoc(c *gin.Context, visitor hydra.Visitor) {
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	accountInfo := interfaces.AccountInfo{ID: visitor.ID, Type: string(visitor.Type)}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)
	oteltrace.AddHttpAttrs4API(span, oteltrace.GetAttrsByGinCtx(c))

	resource, ok := r.requireDatasetResource(c, ctx, span, c.Param("id"))
	if !ok {
		return
	}

	docID := c.Param("doc_id")
	doc, err := r.ds.GetDocument(ctx, resource.ID, docID)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if doc == nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Resource_NotFound).
			WithErrorDetails(fmt.Sprintf("document %s not found", docID))
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	oteltrace.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, doc)
}

// =========================== PUT /resources/:id/data/:doc_id ===========================

// PutResourceDataDocByEx handles PUT /api/vega-backend/v1/resources/:id/data/:doc_id (External).
// Single-document update; doc_id from path takes precedence over any `id` field in body.
func (r *restHandler) PutResourceDataDocByEx(c *gin.Context) {
	visitor, err := r.verifyOAuth(rest.GetLanguageCtx(c), c)
	if err != nil {
		return
	}
	r.putResourceDataDoc(c, visitor)
}

// PutResourceDataDocByIn handles PUT /api/vega-backend/in/v1/resources/:id/data/:doc_id (Internal).
func (r *restHandler) PutResourceDataDocByIn(c *gin.Context) {
	visitor := visitor.GenerateVisitor(c)
	r.putResourceDataDoc(c, visitor)
}

func (r *restHandler) putResourceDataDoc(c *gin.Context, visitor hydra.Visitor) {
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	accountInfo := interfaces.AccountInfo{ID: visitor.ID, Type: string(visitor.Type)}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)
	oteltrace.AddHttpAttrs4API(span, oteltrace.GetAttrsByGinCtx(c))

	resource, ok := r.requireDatasetResource(c, ctx, span, c.Param("id"))
	if !ok {
		return
	}

	docID := c.Param("doc_id")

	var doc map[string]any
	if err := c.ShouldBindJSON(&doc); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_RequestBody).
			WithErrorDetails(err.Error())
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if doc == nil {
		doc = map[string]any{}
	}
	if bodyID, exists := doc["id"]; exists {
		if s, ok := bodyID.(string); !ok || s != docID {
			logger.Warnf("PutResourceDataDoc: body.id (%v) overridden by path doc_id (%s)", bodyID, docID)
		}
	}
	doc["id"] = docID

	if _, err := r.ds.UpsertDocuments(ctx, resource.ID, []map[string]any{doc}); err != nil {
		httpErr := err.(*rest.HTTPError)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	logger.Debug("Handler putResourceDataDoc Success")
	oteltrace.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, map[string]any{"id": docID})
}

// =========================== DELETE /resources/:id/data/:doc_ids ===========================

// DeleteResourceDataByEx handles DELETE /api/vega-backend/v1/resources/:id/data/:doc_ids (External).
// Best-effort batch delete by IDs; missing IDs are silently skipped.
func (r *restHandler) DeleteResourceDataByEx(c *gin.Context) {
	visitor, err := r.verifyOAuth(rest.GetLanguageCtx(c), c)
	if err != nil {
		return
	}
	r.deleteResourceData(c, visitor)
}

// DeleteResourceDataByIn handles DELETE /api/vega-backend/in/v1/resources/:id/data/:doc_ids (Internal).
func (r *restHandler) DeleteResourceDataByIn(c *gin.Context) {
	visitor := visitor.GenerateVisitor(c)
	r.deleteResourceData(c, visitor)
}

func (r *restHandler) deleteResourceData(c *gin.Context, visitor hydra.Visitor) {
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	accountInfo := interfaces.AccountInfo{ID: visitor.ID, Type: string(visitor.Type)}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)
	oteltrace.AddHttpAttrs4API(span, oteltrace.GetAttrsByGinCtx(c))

	resource, ok := r.requireDatasetResource(c, ctx, span, c.Param("id"))
	if !ok {
		return
	}

	// service expects comma-separated string; pass through.
	docIDs := c.Param("doc_ids")
	if err := r.ds.DeleteDocuments(ctx, resource.ID, docIDs); err != nil {
		httpErr := err.(*rest.HTTPError)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	logger.Debug("Handler deleteResourceData Success")
	oteltrace.AddHttpAttrs4Ok(span, http.StatusNoContent)
	rest.ReplyOK(c, http.StatusNoContent, nil)
}

// =========================== helpers ===========================

// requireDatasetResource loads resource by id and verifies it exists with category=dataset.
// On failure replies with the appropriate HTTP error and returns ok=false.
func (r *restHandler) requireDatasetResource(c *gin.Context, ctx context.Context, span trace.Span, id string) (*interfaces.Resource, bool) {
	resource, err := r.rs.GetByID(ctx, id)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return nil, false
	}
	if resource == nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Resource_NotFound)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return nil, false
	}
	if resource.Category != interfaces.ResourceCategoryDataset {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_Resource_InternalError_InvalidCategory).
			WithErrorDetails(fmt.Sprintf("operation requires resource category=dataset, got: %s", resource.Category))
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return nil, false
	}
	return resource, true
}

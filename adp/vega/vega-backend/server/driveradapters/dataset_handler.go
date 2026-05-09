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
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/hydra"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/mitchellh/mapstructure"
	"go.opentelemetry.io/otel/trace"

	verrors "vega-backend/errors"
	"vega-backend/interfaces"
)

// ========== CreateDatasetDocuments ==========

// CreateDatasetDocumentsByEx handles POST /api/vega-backend/v1/resources/dataset/:id/docs (External)
func (r *restHandler) CreateDatasetDocumentsByEx(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"CreateDatasetDocumentsByEx", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 外网接口：校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.createDatasetDocuments(c, ctx, span, visitor)
}

// CreateDatasetDocumentsByIn handles POST /api/vega-backend/in/v1/resources/dataset/:id/docs (Internal)
func (r *restHandler) CreateDatasetDocumentsByIn(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"CreateDatasetDocumentsByIn", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 内网接口：user_id从header中取
	visitor := GenerateVisitor(c)
	r.createDatasetDocuments(c, ctx, span, visitor)
}

// createDatasetDocuments is the shared implementation
func (r *restHandler) createDatasetDocuments(c *gin.Context, ctx context.Context, span trace.Span, visitor hydra.Visitor) {
	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	datasetID := c.Param("id")

	// 获取请求体
	var documents []map[string]any
	if err := c.ShouldBindJSON(&documents); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_RequestBody).
			WithErrorDetails(err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 调用 dataset 服务批量创建文档
	docIDs, err := r.ds.CreateDocuments(ctx, datasetID, documents)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	resultData := map[string]any{"ids": docIDs}

	logger.Debug("Handler CreateDatasetDocuments Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusCreated)
	rest.ReplyOK(c, http.StatusCreated, resultData)
}

// ========== UpdateDatasetDocuments ==========

// UpdateDatasetDocumentsByEx handles PUT /api/vega-backend/v1/resources/dataset/:id/docs (External)
func (r *restHandler) UpdateDatasetDocumentsByEx(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"UpdateDatasetDocumentsByEx", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 外网接口：校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.updateDatasetDocuments(c, ctx, span, visitor)
}

// UpdateDatasetDocumentsByIn handles PUT /api/vega-backend/in/v1/resources/dataset/:id/docs (Internal)
func (r *restHandler) UpdateDatasetDocumentsByIn(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"UpdateDatasetDocumentsByIn", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 内网接口：user_id从header中取
	visitor := GenerateVisitor(c)
	r.updateDatasetDocuments(c, ctx, span, visitor)
}

// updateDatasetDocuments is the shared implementation
func (r *restHandler) updateDatasetDocuments(c *gin.Context, ctx context.Context, span trace.Span, visitor hydra.Visitor) {
	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	datasetID := c.Param("id")

	// 获取请求体，支持批量更新
	var updateRequests []map[string]any
	if err := c.ShouldBindJSON(&updateRequests); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_RequestBody).
			WithErrorDetails(err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 调用 dataset 服务批量更新文档
	successDocIDs, err := r.ds.UpsertDocuments(ctx, datasetID, updateRequests)
	if len(successDocIDs) == 0 {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Resource_InternalError).
			WithErrorDetails(err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	logger.Debug("Handler UpdateDatasetDocuments Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, map[string]any{"ids": successDocIDs})
}

// ========== DeleteDatasetDocuments ==========

// DeleteDatasetDocumentsByEx handles DELETE /api/vega-backend/v1/resources/dataset/:id/docs/:ids (External)
func (r *restHandler) DeleteDatasetDocumentsByEx(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"DeleteDatasetDocumentsByEx", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 外网接口：校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.deleteDatasetDocuments(c, ctx, span, visitor)
}

// DeleteDatasetDocumentsByIn handles DELETE /api/vega-backend/in/v1/resources/dataset/:id/docs/:ids (Internal)
func (r *restHandler) DeleteDatasetDocumentsByIn(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"DeleteDatasetDocumentsByIn", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 内网接口：user_id从header中取
	visitor := GenerateVisitor(c)
	r.deleteDatasetDocuments(c, ctx, span, visitor)
}

// deleteDatasetDocuments is the shared implementation
func (r *restHandler) deleteDatasetDocuments(c *gin.Context, ctx context.Context, span trace.Span, visitor hydra.Visitor) {
	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	datasetID := c.Param("id")
	docIDs := c.Param("ids")

	// 调用 dataset 服务批量删除文档
	if err := r.ds.DeleteDocuments(ctx, datasetID, docIDs); err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	logger.Debug("Handler DeleteDatasetDocuments Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusNoContent)
	rest.ReplyOK(c, http.StatusNoContent, nil)
}

// ========== DeleteDatasetDocumentsByQuery ==========

// DeleteDatasetDocumentsByQueryByEx handles POST /api/vega-backend/v1/resources/dataset/:id/docs/query (External)
func (r *restHandler) DeleteDatasetDocumentsByQueryByEx(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"DeleteDatasetDocumentsByQueryByEx", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 外网接口：校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.deleteDatasetDocumentsByQuery(c, ctx, span, visitor)
}

// DeleteDatasetDocumentsByQueryByIn handles POST /api/vega-backend/in/v1/resources/dataset/:id/docs/query (Internal)
func (r *restHandler) DeleteDatasetDocumentsByQueryByIn(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"DeleteDatasetDocumentsByQueryByIn", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 内网接口：user_id从header中取
	visitor := GenerateVisitor(c)
	r.deleteDatasetDocumentsByQuery(c, ctx, span, visitor)
}

// deleteDatasetDocumentsByQuery is the shared implementation
func (r *restHandler) deleteDatasetDocumentsByQuery(c *gin.Context, ctx context.Context, span trace.Span, visitor hydra.Visitor) {
	start := time.Now()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	// 1. check重载请求头
	method := c.GetHeader(interfaces.HTTP_HEADER_METHOD_OVERRIDE)
	if method != http.MethodDelete {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_OverrideMethod)
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyErrorWithHeaders(c, httpErr, map[string]string{
			interfaces.X_REQUEST_TOOK: time.Since(start).String(),
		})
		return
	}

	datasetID := c.Param("id")

	// 解析请求体
	var params interfaces.ResourceDataQueryParams
	if err := c.ShouldBindJSON(&params); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_RequestBody).
			WithErrorDetails(err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 过滤条件用map接，然后再decode到condCfg中
	var actualCond *interfaces.FilterCondCfg
	err := mapstructure.Decode(params.FilterCondition, &actualCond)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_FilterCondition).
			WithErrorDetails(fmt.Sprintf("mapstructure decode filters failed: %s", err.Error()))
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	params.FilterCondCfg = actualCond

	resource, err := r.rs.GetByID(ctx, datasetID)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if resource == nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Resource_NotFound)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 调用 dataset 服务批量删除文档
	if err := r.ds.DeleteDocumentsByQuery(ctx, resource.ID, resource, &params); err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	logger.Debug("Handler DeleteDatasetDocumentsByQuery Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusNoContent)
	rest.ReplyOK(c, http.StatusNoContent, nil)
}

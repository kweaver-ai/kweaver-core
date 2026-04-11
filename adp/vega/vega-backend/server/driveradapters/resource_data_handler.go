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
	"go.opentelemetry.io/otel/trace"

	verrors "vega-backend/errors"
	"vega-backend/interfaces"
)

// QueryResourceDataByEx handles POST /api/vega-backend/v1/resources/:id/data (External)
func (r *restHandler) QueryResourceDataByEx(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"QueryResourceDataByEx", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 外网接口：校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.queryResourceData(c, ctx, span, visitor)
}

// QueryResourceDataByIn handles POST /api/vega-backend/in/v1/resources/:id/data (Internal)
func (r *restHandler) QueryResourceDataByIn(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"QueryResourceDataByIn", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 内网接口：user_id从header中取
	visitor := GenerateVisitor(c)
	r.queryResourceData(c, ctx, span, visitor)
}

// queryResourceData is the shared implementation
func (r *restHandler) queryResourceData(c *gin.Context, ctx context.Context, span trace.Span, visitor hydra.Visitor) {
	start := time.Now()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	// 1. check重载请求头
	method := c.GetHeader(interfaces.HTTP_HEADER_METHOD_OVERRIDE)
	if method != http.MethodGet {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_OverrideMethod)
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyErrorWithHeaders(c, httpErr, map[string]string{
			interfaces.X_REQUEST_TOOK: time.Since(start).String(),
		})
		return
	}

	resourceID := c.Param("id")

	// 解析请求体
	var params interfaces.ResourceDataQueryParams
	if err := c.ShouldBindJSON(&params); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_RequestBody).
			WithErrorDetails(err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 资源查询的参数校验
	err := ValidateResourceDataQueryParams(ctx, &params)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	resource, err := r.rs.GetByID(ctx, resourceID)
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

	entries, total, err := r.rds.Query(ctx, resource, &params)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	resultData := map[string]any{
		"entries": entries,
	}
	if params.NeedTotal {
		resultData["total_count"] = total
	}

	logger.Debug("Handler QueryResourceData Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, resultData)
}

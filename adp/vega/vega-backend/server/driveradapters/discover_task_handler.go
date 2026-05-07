// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package driveradapters provides HTTP handlers.
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
	"go.opentelemetry.io/otel/trace"

	verrors "vega-backend/errors"
	"vega-backend/interfaces"
)

// =========================== GET /discover-tasks ===========================

// ListDiscoverTasks handles GET /api/vega-backend/v1/discover-tasks (External)
func (r *restHandler) ListDiscoverTasks(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"ListDiscoverTasks", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.listDiscoverTasks(c, ctx, span, visitor)
}

// ListDiscoverTasksByIn handles GET /api/vega-backend/in/v1/discover-tasks (Internal)
func (r *restHandler) ListDiscoverTasksByIn(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"ListDiscoverTasksByIn", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	visitor := GenerateVisitor(c)
	r.listDiscoverTasks(c, ctx, span, visitor)
}

func (r *restHandler) listDiscoverTasks(c *gin.Context, ctx context.Context, span trace.Span, visitor hydra.Visitor) {
	accountInfo := interfaces.AccountInfo{ID: visitor.ID, Type: string(visitor.Type)}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)
	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	var params interfaces.DiscoverTaskQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_RequestBody).
			WithErrorDetails(err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if params.Limit == 0 {
		params.Limit = 20
	}

	if params.Status != "" && !isValidDiscoverTaskStatus(params.Status) {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_DiscoverTask_InvalidStatus).
			WithErrorDetails(fmt.Sprintf("invalid status: %s", params.Status))
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	tasks, total, err := r.dts.List(ctx, params)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_DiscoverTask_InternalError_GetFailed).
			WithErrorDetails(err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	logger.Debug("Handler ListDiscoverTasks Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, gin.H{
		"entries":     tasks,
		"total_count": total,
	})
}

// =========================== GET /discover-tasks/:id ===========================

// GetDiscoverTask handles GET /api/vega-backend/v1/discover-tasks/:id (External)
func (r *restHandler) GetDiscoverTask(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"GetDiscoverTask", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.getDiscoverTask(c, ctx, span, visitor)
}

// GetDiscoverTaskByIn handles GET /api/vega-backend/in/v1/discover-tasks/:id (Internal)
func (r *restHandler) GetDiscoverTaskByIn(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"GetDiscoverTaskByIn", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	visitor := GenerateVisitor(c)
	r.getDiscoverTask(c, ctx, span, visitor)
}

func (r *restHandler) getDiscoverTask(c *gin.Context, ctx context.Context, span trace.Span, visitor hydra.Visitor) {
	accountInfo := interfaces.AccountInfo{ID: visitor.ID, Type: string(visitor.Type)}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)
	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	taskID := c.Param("id")

	task, err := r.dts.GetByID(ctx, taskID)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_DiscoverTask_InternalError_GetFailed).
			WithErrorDetails(err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if task == nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_DiscoverTask_NotFound)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, task)
}

// =========================== DELETE /discover-tasks/:ids ===========================

// DeleteDiscoverTasks handles DELETE /api/vega-backend/v1/discover-tasks/:ids (External).
// `ids` is comma-separated. Optional query: ?ignore_missing=true
func (r *restHandler) DeleteDiscoverTasks(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"DeleteDiscoverTasks", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.deleteDiscoverTasks(c, ctx, span, visitor)
}

// DeleteDiscoverTasksByIn handles DELETE /api/vega-backend/in/v1/discover-tasks/:ids (Internal)
func (r *restHandler) DeleteDiscoverTasksByIn(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"DeleteDiscoverTasksByIn", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	visitor := GenerateVisitor(c)
	r.deleteDiscoverTasks(c, ctx, span, visitor)
}

func (r *restHandler) deleteDiscoverTasks(c *gin.Context, ctx context.Context, span trace.Span, visitor hydra.Visitor) {
	accountInfo := interfaces.AccountInfo{ID: visitor.ID, Type: string(visitor.Type)}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)
	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	idsStr := c.Param("ids")
	ids := make([]string, 0)
	for _, id := range strings.Split(idsStr, ",") {
		id = strings.TrimSpace(id)
		if id != "" {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_RequestBody).
			WithErrorDetails("ids path parameter is required")
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	ignoreMissing := strings.EqualFold(c.Query("ignore_missing"), "true")

	if err := r.dts.Delete(ctx, ids, ignoreMissing); err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	for _, id := range ids {
		audit.NewWarnLog(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
			interfaces.GenerateResourceAuditObject(id, ""), audit.SUCCESS, "")
	}

	logger.Debug("Handler DeleteDiscoverTasks Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusNoContent)
	rest.ReplyOK(c, http.StatusNoContent, nil)
}

// =========================== helpers ===========================

func isValidDiscoverTaskStatus(s string) bool {
	switch s {
	case interfaces.DiscoverTaskStatusPending,
		interfaces.DiscoverTaskStatusRunning,
		interfaces.DiscoverTaskStatusCompleted,
		interfaces.DiscoverTaskStatusFailed:
		return true
	}
	return false
}

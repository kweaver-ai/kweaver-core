// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package driveradapters provides HTTP handlers.
package driveradapters

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"go.opentelemetry.io/otel/trace"

	verrors "vega-backend/errors"
	"vega-backend/interfaces"
)

// GetDiscoverTask handles GET /api/vega-backend/v1/discover-tasks/:id
func (r *restHandler) GetDiscoverTask(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"GetDiscoverTask", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	taskID := c.Param("id")
	// Get task
	task, err := r.dts.GetByID(ctx, taskID)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Catalog_InternalError).
			WithErrorDetails(err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if task == nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Task_NotFound)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	logger.Debug("Handler GetDiscoverTask Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, task)
}

// GetDiscoverTaskByScheduleId handles GET /api/vega-backend/v1/discover-tasks/:scheduleId
func (r *restHandler) GetDiscoverTaskByScheduleId(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"GetDiscoverTaskByScheduleId", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	scheduleId := c.Param("scheduleId")
	// Get tasks by scheduled ID
	tasks, err := r.dts.GetByScheduledID(ctx, scheduleId)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Catalog_InternalError).
			WithErrorDetails(err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if len(tasks) == 0 {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Task_NotFound)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	logger.Debug("Handler GetDiscoverTaskByScheduleId Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	result := map[string]any{
		"entries":     tasks,
		"total_count": len(tasks),
	}
	rest.ReplyOK(c, http.StatusOK, result)
}

// ListDiscoverTasks handles GET /api/vega-backend/v1/discover-tasks
func (r *restHandler) ListDiscoverTasks(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"ListDiscoverTasks", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	// Parse query params
	params := interfaces.DiscoverTaskQueryParams{
		CatalogID:   c.Query("catalog_id"),
		Status:      c.Query("status"),
		TriggerType: c.Query("trigger_type"),
	}
	if err := c.ShouldBindQuery(&params.PaginationQueryParams); err == nil {
		if params.Limit == 0 {
			params.Limit = 10
		}
	}

	// Verify catalog exists if catalog_id is provided
	if params.CatalogID != "" {
		catalog, err := r.cs.GetByID(ctx, params.CatalogID, false)
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

	// List tasks
	tasks, total, err := r.dts.List(ctx, params)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Catalog_InternalError).
			WithErrorDetails(err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	logger.Debug("Handler ListDiscoverTasks Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, gin.H{
		"entries": tasks,
		"total":   total,
	})
}

// ListDiscoverTasksByIn handles GET /api/vega-backend/in/v1/discover-tasks
func (r *restHandler) ListDiscoverTasksByIn(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"ListDiscoverTasksByIn", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// Generate visitor from request headers (for internal APIs)
	visitor := GenerateVisitor(c)
	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	// Parse query params
	params := interfaces.DiscoverTaskQueryParams{
		CatalogID:   c.Query("catalog_id"),
		Status:      c.Query("status"),
		TriggerType: c.Query("trigger_type"),
	}
	if err := c.ShouldBindQuery(&params.PaginationQueryParams); err == nil {
		if params.Limit == 0 {
			params.Limit = 10
		}
	}

	// List tasks
	tasks, total, err := r.dts.List(ctx, params)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Catalog_InternalError).
			WithErrorDetails(err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	logger.Debug("Handler ListDiscoverTasksByIn Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, gin.H{
		"entries": tasks,
		"total":   total,
	})
}

// GetDiscoverTaskByIn handles GET /api/vega-backend/in/v1/discover-tasks/:id
func (r *restHandler) GetDiscoverTaskByIn(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"GetDiscoverTaskByIn", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// Generate visitor from request headers (for internal APIs)
	visitor := GenerateVisitor(c)
	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	taskID := c.Param("id")
	// Get task
	task, err := r.dts.GetByID(ctx, taskID)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Catalog_InternalError).
			WithErrorDetails(err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if task == nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Task_NotFound)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	logger.Debug("Handler GetDiscoverTaskByIn Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, task)
}

// GetDiscoverTaskByScheduleIdByIn handles GET /api/vega-backend/in/v1/discover-tasks/by-schedule/:scheduleId
func (r *restHandler) GetDiscoverTaskByScheduleIdByIn(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"GetDiscoverTaskByScheduleIdByIn", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// Generate visitor from request headers (for internal APIs)
	visitor := GenerateVisitor(c)
	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	scheduleId := c.Param("scheduleId")
	// Get tasks by scheduled ID
	tasks, err := r.dts.GetByScheduledID(ctx, scheduleId)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_Catalog_InternalError).
			WithErrorDetails(err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if len(tasks) == 0 {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Task_NotFound)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	logger.Debug("Handler GetDiscoverTaskByScheduleIdByIn Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	result := map[string]any{
		"entries":     tasks,
		"total_count": len(tasks),
	}
	rest.ReplyOK(c, http.StatusOK, result)
}

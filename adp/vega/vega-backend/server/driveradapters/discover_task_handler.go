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
	"github.com/kweaver-ai/kweaver-go-lib/audit"
	"github.com/kweaver-ai/kweaver-go-lib/hydra"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/otel/oteltrace"
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	"vega-backend/common/visitor"
	verrors "vega-backend/errors"
	"vega-backend/interfaces"
)

// =========================== GET /discover-tasks ===========================

// ListDiscoverTasksByEx handles GET /api/vega-backend/v1/discover-tasks (External)
func (r *restHandler) ListDiscoverTasksByEx(c *gin.Context) {
	visitor, err := r.verifyOAuth(rest.GetLanguageCtx(c), c)
	if err != nil {
		return
	}
	r.listDiscoverTasks(c, visitor)
}

// ListDiscoverTasksByIn handles GET /api/vega-backend/in/v1/discover-tasks (Internal)
func (r *restHandler) ListDiscoverTasksByIn(c *gin.Context) {
	visitor := visitor.GenerateVisitor(c)
	r.listDiscoverTasks(c, visitor)
}

func (r *restHandler) listDiscoverTasks(c *gin.Context, visitor hydra.Visitor) {
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	accountInfo := interfaces.AccountInfo{ID: visitor.ID, Type: string(visitor.Type)}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)
	oteltrace.AddHttpAttrs4API(span, oteltrace.GetAttrsByGinCtx(c))

	var params interfaces.DiscoverTaskQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_RequestBody).
			WithErrorDetails(err.Error())
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if params.Limit == 0 {
		params.Limit = 20
	}

	if params.Status != "" && !isValidDiscoverTaskStatus(params.Status) {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_DiscoverTask_InvalidStatus).
			WithErrorDetails(fmt.Sprintf("invalid status: %s", params.Status))
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	tasks, total, err := r.dts.List(ctx, params)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_DiscoverTask_InternalError_GetFailed).
			WithErrorDetails(err.Error())
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	logger.Debug("Handler ListDiscoverTasksByEx Success")
	oteltrace.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, gin.H{
		"entries":     tasks,
		"total_count": total,
	})
}

// =========================== GET /discover-tasks/:id ===========================

// GetDiscoverTaskByEx handles GET /api/vega-backend/v1/discover-tasks/:id (External)
func (r *restHandler) GetDiscoverTaskByEx(c *gin.Context) {
	visitor, err := r.verifyOAuth(rest.GetLanguageCtx(c), c)
	if err != nil {
		return
	}
	r.getDiscoverTask(c, visitor)
}

// GetDiscoverTaskByIn handles GET /api/vega-backend/in/v1/discover-tasks/:id (Internal)
func (r *restHandler) GetDiscoverTaskByIn(c *gin.Context) {
	visitor := visitor.GenerateVisitor(c)
	r.getDiscoverTask(c, visitor)
}

func (r *restHandler) getDiscoverTask(c *gin.Context, visitor hydra.Visitor) {
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	accountInfo := interfaces.AccountInfo{ID: visitor.ID, Type: string(visitor.Type)}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)
	oteltrace.AddHttpAttrs4API(span, oteltrace.GetAttrsByGinCtx(c))

	taskID := c.Param("id")

	task, err := r.dts.GetByID(ctx, taskID)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_DiscoverTask_InternalError_GetFailed).
			WithErrorDetails(err.Error())
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if task == nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_DiscoverTask_NotFound)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	oteltrace.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, task)
}

// =========================== DELETE /discover-tasks/:ids ===========================

// DeleteDiscoverTasksByEx handles DELETE /api/vega-backend/v1/discover-tasks/:ids (External).
// `ids` is comma-separated. Optional query: ?ignore_missing=true
func (r *restHandler) DeleteDiscoverTasksByEx(c *gin.Context) {
	visitor, err := r.verifyOAuth(rest.GetLanguageCtx(c), c)
	if err != nil {
		return
	}
	r.deleteDiscoverTasks(c, visitor)
}

// DeleteDiscoverTasksByIn handles DELETE /api/vega-backend/in/v1/discover-tasks/:ids (Internal)
func (r *restHandler) DeleteDiscoverTasksByIn(c *gin.Context) {
	visitor := visitor.GenerateVisitor(c)
	r.deleteDiscoverTasks(c, visitor)
}

func (r *restHandler) deleteDiscoverTasks(c *gin.Context, visitor hydra.Visitor) {
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	accountInfo := interfaces.AccountInfo{ID: visitor.ID, Type: string(visitor.Type)}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)
	oteltrace.AddHttpAttrs4API(span, oteltrace.GetAttrsByGinCtx(c))

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
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	ignoreMissing := strings.EqualFold(c.Query("ignore_missing"), "true")

	if err := r.dts.Delete(ctx, ids, ignoreMissing); err != nil {
		httpErr := err.(*rest.HTTPError)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	for _, id := range ids {
		audit.NewWarnLog(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
			interfaces.GenerateResourceAuditObject(id, ""), audit.SUCCESS, "")
	}

	logger.Debug("Handler DeleteDiscoverTasksByEx Success")
	oteltrace.AddHttpAttrs4Ok(span, http.StatusNoContent)
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

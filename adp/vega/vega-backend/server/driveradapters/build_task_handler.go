// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

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

// buildTaskListQuery captures filter + pagination from query string.
type buildTaskListQuery struct {
	interfaces.PaginationQueryParams
	ResourceID string `form:"resource_id"`
	CatalogID  string `form:"catalog_id"`
	Status     string `form:"status"`
	Mode       string `form:"mode"`
}

// =========================== POST /build-tasks ===========================

// CreateBuildTaskByEx handles POST /api/vega-backend/v1/build-tasks (External).
func (r *restHandler) CreateBuildTaskByEx(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"CreateBuildTaskByEx", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.createBuildTask(c, ctx, span, visitor)
}

// CreateBuildTaskByIn handles POST /api/vega-backend/in/v1/build-tasks (Internal).
func (r *restHandler) CreateBuildTaskByIn(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"CreateBuildTaskByIn", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	visitor := GenerateVisitor(c)
	r.createBuildTask(c, ctx, span, visitor)
}

func (r *restHandler) createBuildTask(c *gin.Context, ctx context.Context, span trace.Span, visitor hydra.Visitor) {
	accountInfo := interfaces.AccountInfo{ID: visitor.ID, Type: string(visitor.Type)}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)
	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	var req interfaces.CreateBuildTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_RequestBody).
			WithErrorDetails(err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	resource, err := r.rs.GetByID(ctx, req.ResourceID)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	if req.BuildKeyFields == "" && req.Mode == interfaces.BuildTaskModeBatch {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_RequestBody).
			WithErrorDetails("build_key_fields is required for batch mode")
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	schemaFields := make(map[string]bool, len(resource.SchemaDefinition))
	for _, prop := range resource.SchemaDefinition {
		schemaFields[prop.Name] = true
	}
	if req.BuildKeyFields != "" {
		for _, key := range strings.Split(req.BuildKeyFields, ",") {
			key = strings.TrimSpace(key)
			if key != "" && !schemaFields[key] {
				httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_RequestBody).
					WithErrorDetails(fmt.Sprintf("build_key_field '%s' not found in resource schema", key))
				o11y.AddHttpAttrs4HttpError(span, httpErr)
				rest.ReplyError(c, httpErr)
				return
			}
		}
	}
	if req.EmbeddingFields != "" {
		for _, field := range strings.Split(req.EmbeddingFields, ",") {
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

	taskID, err := r.bts.CreateBuildTask(ctx, req.ResourceID, &req.BuildTaskRequest)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	audit.NewInfoLog(audit.OPERATION, "build", audit.TransforOperator(visitor),
		interfaces.GenerateResourceAuditObject(req.ResourceID, ""), "")

	logger.Debug("Handler CreateBuildTask Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusCreated)
	rest.ReplyOK(c, http.StatusCreated, gin.H{
		"id":          taskID,
		"resource_id": req.ResourceID,
		"status":      interfaces.BuildTaskStatusInit,
	})
}

// =========================== GET /build-tasks/:id ===========================

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

func (r *restHandler) GetBuildTaskByIn(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"GetBuildTaskByIn", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	visitor := GenerateVisitor(c)
	r.getBuildTask(c, ctx, span, visitor)
}

func (r *restHandler) getBuildTask(c *gin.Context, ctx context.Context, span trace.Span, visitor hydra.Visitor) {
	accountInfo := interfaces.AccountInfo{ID: visitor.ID, Type: string(visitor.Type)}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)
	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	taskID := c.Param("id")
	buildTask, err := r.bts.GetBuildTaskByID(ctx, taskID)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, buildTask)
}

// =========================== GET /build-tasks ===========================

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

func (r *restHandler) ListBuildTasksByIn(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"ListBuildTasksByIn", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	visitor := GenerateVisitor(c)
	r.listBuildTasks(c, ctx, span, visitor)
}

func (r *restHandler) listBuildTasks(c *gin.Context, ctx context.Context, span trace.Span, visitor hydra.Visitor) {
	accountInfo := interfaces.AccountInfo{ID: visitor.ID, Type: string(visitor.Type)}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)
	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	var q buildTaskListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_RequestBody).
			WithErrorDetails(err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if q.Limit == 0 {
		q.Limit = 20
	}
	if q.Sort == "" {
		q.Sort = "update_time"
	}
	if q.Direction == "" {
		q.Direction = interfaces.DESC_DIRECTION
	}

	if q.Status != "" && !isValidBuildTaskStatus(q.Status) {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_BuildTask_InvalidStatus).
			WithErrorDetails(fmt.Sprintf("invalid status: %s", q.Status))
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if q.Mode != "" && !isValidBuildTaskMode(q.Mode) {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_BuildTask_InvalidParameter_Mode).
			WithErrorDetails(fmt.Sprintf("invalid mode: %s", q.Mode))
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	params := interfaces.BuildTasksQueryParams{
		PaginationQueryParams: q.PaginationQueryParams,
		ResourceID:            q.ResourceID,
		CatalogID:             q.CatalogID,
		Status:                q.Status,
		Mode:                  q.Mode,
	}
	tasks, total, err := r.bts.ListBuildTasks(ctx, params)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, gin.H{"entries": tasks, "total_count": total})
}

// =========================== DELETE /build-tasks/:id ===========================

func (r *restHandler) DeleteBuildTaskByEx(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"DeleteBuildTaskByEx", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.deleteBuildTask(c, ctx, span, visitor)
}

func (r *restHandler) DeleteBuildTaskByIn(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"DeleteBuildTaskByIn", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	visitor := GenerateVisitor(c)
	r.deleteBuildTask(c, ctx, span, visitor)
}

func (r *restHandler) deleteBuildTask(c *gin.Context, ctx context.Context, span trace.Span, visitor hydra.Visitor) {
	accountInfo := interfaces.AccountInfo{ID: visitor.ID, Type: string(visitor.Type)}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)
	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	taskID := c.Param("id")
	if err := r.bts.DeleteBuildTask(ctx, taskID); err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	audit.NewWarnLog(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
		interfaces.GenerateResourceAuditObject(taskID, ""), audit.SUCCESS, "")

	o11y.AddHttpAttrs4Ok(span, http.StatusNoContent)
	rest.ReplyOK(c, http.StatusNoContent, nil)
}

// =========================== DELETE /build-tasks?ids=... ===========================

func (r *restHandler) BatchDeleteBuildTasksByEx(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"BatchDeleteBuildTasksByEx", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.batchDeleteBuildTasks(c, ctx, span, visitor)
}

func (r *restHandler) BatchDeleteBuildTasksByIn(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"BatchDeleteBuildTasksByIn", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	visitor := GenerateVisitor(c)
	r.batchDeleteBuildTasks(c, ctx, span, visitor)
}

func (r *restHandler) batchDeleteBuildTasks(c *gin.Context, ctx context.Context, span trace.Span, visitor hydra.Visitor) {
	accountInfo := interfaces.AccountInfo{ID: visitor.ID, Type: string(visitor.Type)}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)
	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	idsStr := c.Query("ids")
	if idsStr == "" {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_RequestBody).
			WithErrorDetails("ids query parameter is required")
		rest.ReplyError(c, httpErr)
		return
	}

	deleted := make([]string, 0)
	failed := make([]interfaces.BatchDeleteFailedItem, 0)
	for _, taskID := range strings.Split(idsStr, ",") {
		taskID = strings.TrimSpace(taskID)
		if taskID == "" {
			continue
		}
		if err := r.bts.DeleteBuildTask(ctx, taskID); err != nil {
			code, msg := extractHTTPErrorParts(err)
			failed = append(failed, interfaces.BatchDeleteFailedItem{ID: taskID, Code: code, Message: msg})
			continue
		}
		deleted = append(deleted, taskID)
	}

	for _, id := range deleted {
		audit.NewWarnLog(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
			interfaces.GenerateResourceAuditObject(id, ""), audit.SUCCESS, "")
	}

	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, gin.H{
		"affected": len(deleted),
		"ids":      deleted,
		"failed":   failed,
	})
}

// =========================== POST /build-tasks/:id/start ===========================

func (r *restHandler) StartBuildTaskByEx(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"StartBuildTaskByEx", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.startBuildTask(c, ctx, span, visitor)
}

func (r *restHandler) StartBuildTaskByIn(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"StartBuildTaskByIn", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	visitor := GenerateVisitor(c)
	r.startBuildTask(c, ctx, span, visitor)
}

func (r *restHandler) startBuildTask(c *gin.Context, ctx context.Context, span trace.Span, visitor hydra.Visitor) {
	accountInfo := interfaces.AccountInfo{ID: visitor.ID, Type: string(visitor.Type)}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)
	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	taskID := c.Param("id")
	var req interfaces.StartBuildTaskRequest
	// body is optional; bind errors are tolerated
	_ = c.ShouldBindJSON(&req)

	buildTask, err := r.bts.StartBuildTask(ctx, taskID, req.ExecuteType)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	audit.NewInfoLog(audit.OPERATION, "start", audit.TransforOperator(visitor),
		interfaces.GenerateResourceAuditObject(taskID, ""), "")

	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, buildTask)
}

// =========================== POST /build-tasks/:id/stop ===========================

func (r *restHandler) StopBuildTaskByEx(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"StopBuildTaskByEx", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.stopBuildTask(c, ctx, span, visitor)
}

func (r *restHandler) StopBuildTaskByIn(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"StopBuildTaskByIn", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	visitor := GenerateVisitor(c)
	r.stopBuildTask(c, ctx, span, visitor)
}

func (r *restHandler) stopBuildTask(c *gin.Context, ctx context.Context, span trace.Span, visitor hydra.Visitor) {
	accountInfo := interfaces.AccountInfo{ID: visitor.ID, Type: string(visitor.Type)}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)
	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	taskID := c.Param("id")
	buildTask, err := r.bts.StopBuildTask(ctx, taskID)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	audit.NewInfoLog(audit.OPERATION, "stop", audit.TransforOperator(visitor),
		interfaces.GenerateResourceAuditObject(taskID, ""), "")

	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, buildTask)
}

// =========================== GET /resources/:id/build-tasks ===========================

func (r *restHandler) ListResourceBuildTasksByEx(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"ListResourceBuildTasksByEx", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.listResourceBuildTasks(c, ctx, span, visitor)
}

func (r *restHandler) ListResourceBuildTasksByIn(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"ListResourceBuildTasksByIn", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	visitor := GenerateVisitor(c)
	r.listResourceBuildTasks(c, ctx, span, visitor)
}

func (r *restHandler) listResourceBuildTasks(c *gin.Context, ctx context.Context, span trace.Span, visitor hydra.Visitor) {
	accountInfo := interfaces.AccountInfo{ID: visitor.ID, Type: string(visitor.Type)}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)
	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	resourceID := c.Param("id")
	exists, err := r.rs.CheckExistByID(ctx, resourceID)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if !exists {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Resource_NotFound)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	var q buildTaskListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_RequestBody).
			WithErrorDetails(err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if q.Limit == 0 {
		q.Limit = 20
	}
	if q.Sort == "" {
		q.Sort = "update_time"
	}
	if q.Direction == "" {
		q.Direction = interfaces.DESC_DIRECTION
	}

	params := interfaces.BuildTasksQueryParams{
		PaginationQueryParams: q.PaginationQueryParams,
		ResourceID:            resourceID,
		Status:                q.Status,
		Mode:                  q.Mode,
	}
	tasks, total, err := r.bts.ListBuildTasks(ctx, params)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, gin.H{"entries": tasks, "total_count": total})
}

// =========================== helpers ===========================

func isValidBuildTaskStatus(s string) bool {
	switch s {
	case interfaces.BuildTaskStatusInit,
		interfaces.BuildTaskStatusRunning,
		interfaces.BuildTaskStatusStopping,
		interfaces.BuildTaskStatusStopped,
		interfaces.BuildTaskStatusCompleted,
		interfaces.BuildTaskStatusFailed:
		return true
	}
	return false
}

func isValidBuildTaskMode(m string) bool {
	switch m {
	case interfaces.BuildTaskModeStreaming,
		interfaces.BuildTaskModeBatch,
		interfaces.BuildTaskModeEmbedding:
		return true
	}
	return false
}

// extractHTTPErrorParts pulls the errcode + message from a *rest.HTTPError; falls back to err.Error().
func extractHTTPErrorParts(err error) (string, string) {
	if httpErr, ok := err.(*rest.HTTPError); ok {
		msg := httpErr.BaseError.Description
		if details, ok := httpErr.BaseError.ErrorDetails.(string); ok && details != "" {
			msg = details
		}
		return httpErr.BaseError.ErrorCode, msg
	}
	return verrors.VegaBackend_BuildTask_InternalError_DeleteFailed, err.Error()
}

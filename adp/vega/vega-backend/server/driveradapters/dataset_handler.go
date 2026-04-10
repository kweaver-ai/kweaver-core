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
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/audit"
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
	if err := r.ds.UpdateDocuments(ctx, datasetID, updateRequests); err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	logger.Debug("Handler UpdateDatasetDocuments Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusNoContent)
	rest.ReplyOK(c, http.StatusNoContent, nil)
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
	if err := r.ds.DeleteDocumentsByQuery(ctx, resource, &params); err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	logger.Debug("Handler DeleteDatasetDocumentsByQuery Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusNoContent)
	rest.ReplyOK(c, http.StatusNoContent, nil)
}

// BuildDataByEx handles POST /api/vega-backend/v1/resources/:id/build (External)
func (r *restHandler) BuildDataByEx(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"BuildDataByEx", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 外网接口：校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.buildDataset(c, ctx, span, visitor)
}

// BuildDataByIn handles POST /api/vega-backend/in/v1/resources/:id/build (Internal)
func (r *restHandler) BuildDataByIn(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"BuildDataByIn", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 内网接口：user_id从header中取
	visitor := GenerateVisitor(c)
	r.buildDataset(c, ctx, span, visitor)
}

// buildDataset is the shared implementation
func (r *restHandler) buildDataset(c *gin.Context, ctx context.Context, span trace.Span, visitor hydra.Visitor) {
	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	id := c.Param("id")

	// Check if resource exists
	exists, err := r.rs.CheckExistByID(ctx, id)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if !exists {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound,
			verrors.VegaBackend_Resource_NotFound).WithErrorDetails(fmt.Sprintf("resource id %s not found", id))
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// Parse request body to get task mode
	var req interfaces.BuildTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_RequestBody).
			WithErrorDetails(err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	taskID, err := r.ds.CreateBuildTask(ctx, id, &req)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	audit.NewInfoLog(audit.OPERATION, "build", audit.TransforOperator(visitor),
		interfaces.GenerateResourceAuditObject(id, ""), "")

	logger.Debug("Handler BuildDataset Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, gin.H{"task_id": taskID})
}

// GetBuildTaskByEx handles get build task request (External)
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

// GetBuildTaskByIn handles get build task request (Internal)
func (r *restHandler) GetBuildTaskByIn(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"GetBuildTaskByIn", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 内网接口：user_id从header中取
	visitor := GenerateVisitor(c)
	r.getBuildTask(c, ctx, span, visitor)
}

// getBuildTask is the shared implementation for getting build task
func (r *restHandler) getBuildTask(c *gin.Context, ctx context.Context, span trace.Span, visitor hydra.Visitor) {
	// Get task ID from path parameter
	id := c.Param("id")
	taskID := c.Param("taskid")
	if taskID == "" {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, "VegaBackend.InvalidRequestParameter.TaskID")
		rest.ReplyError(c, httpErr)
		return
	}

	// Check if resource exists
	exists, err := r.rs.CheckExistByID(ctx, id)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if !exists {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound,
			verrors.VegaBackend_Resource_NotFound).WithErrorDetails(fmt.Sprintf("resource id %s not found", id))
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// Get build task
	buildTask, err := r.ds.GetBuildTaskByID(ctx, taskID)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}

	rest.ReplyOK(c, http.StatusOK, buildTask)
}

// UpdateBuildTaskStatusByEx handles update build task status request (External)
func (r *restHandler) UpdateBuildTaskStatusByEx(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"UpdateBuildTaskStatusByEx", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}

	r.updateBuildTaskStatus(c, ctx, span, visitor)
}

// UpdateBuildTaskStatusByIn handles update build task status request (Internal)
func (r *restHandler) UpdateBuildTaskStatusByIn(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"UpdateBuildTaskStatusByIn", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 内网接口：user_id从header中取
	visitor := GenerateVisitor(c)
	r.updateBuildTaskStatus(c, ctx, span, visitor)
}

// updateBuildTaskStatus is the shared implementation for updating build task status
func (r *restHandler) updateBuildTaskStatus(c *gin.Context, ctx context.Context, span trace.Span, visitor hydra.Visitor) {
	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	// Get task ID from path parameter
	id := c.Param("id")
	taskID := c.Param("taskid")
	if taskID == "" {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, "VegaBackend.InvalidRequestParameter.TaskID")
		rest.ReplyError(c, httpErr)
		return
	}

	// Check if resource exists
	exists, err := r.rs.CheckExistByID(ctx, id)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if !exists {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound,
			verrors.VegaBackend_Resource_NotFound).WithErrorDetails(fmt.Sprintf("resource id %s not found", id))
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// Parse request body
	var req interfaces.UpdateBuildTaskStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_RequestBody).
			WithErrorDetails(err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// Update build task status
	if err := r.ds.UpdateBuildTaskStatus(ctx, taskID, &req); err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	logger.Debug("Handler UpdateBuildTaskStatus Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, nil)
}

// ListBuildTasksByEx handles list build tasks request (External)
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

// ListBuildTasksByIn handles list build tasks request (Internal)
func (r *restHandler) ListBuildTasksByIn(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"ListBuildTasksByIn", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 内网接口：user_id从header中取
	visitor := GenerateVisitor(c)
	r.listBuildTasks(c, ctx, span, visitor)
}

// listBuildTasks is the shared implementation for listing build tasks
func (r *restHandler) listBuildTasks(c *gin.Context, ctx context.Context, span trace.Span, visitor hydra.Visitor) {
	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	// 解析分页参数
	offset := 0
	limit := 100
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if val, err := strconv.Atoi(offsetStr); err == nil {
			offset = val
		}
	}
	if limitStr := c.Query("limit"); limitStr != "" {
		if val, err := strconv.Atoi(limitStr); err == nil {
			limit = val
		}
	}

	// Get build tasks with pagination
	buildTasks, totalCount, err := r.ds.GetBuildTasks(ctx, offset, limit)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 返回包含 entries 和 total_count 字段的对象
	rest.ReplyOK(c, http.StatusOK, map[string]any{
		"entries":     buildTasks,
		"total_count": totalCount,
	})
}

// DeleteBuildTasksByEx handles delete build tasks request (External)
func (r *restHandler) DeleteBuildTasksByEx(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"DeleteBuildTasksByEx", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}

	r.deleteBuildTasks(c, ctx, span, visitor)
}

// DeleteBuildTasksByIn handles delete build tasks request (Internal)
func (r *restHandler) DeleteBuildTasksByIn(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"DeleteBuildTasksByIn", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 内网接口：user_id从header中取
	visitor := GenerateVisitor(c)
	r.deleteBuildTasks(c, ctx, span, visitor)
}

// deleteBuildTasks is the shared implementation for deleting build tasks
func (r *restHandler) deleteBuildTasks(c *gin.Context, ctx context.Context, span trace.Span, visitor hydra.Visitor) {
	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	// Get task IDs from path parameter
	taskIDs := c.Param("taskids")
	if taskIDs == "" {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, "VegaBackend.InvalidRequestParameter.TaskIDs")
		rest.ReplyError(c, httpErr)
		return
	}

	// Split task IDs
	taskIDList := strings.Split(taskIDs, ",")

	// Delete build tasks
	for _, taskID := range taskIDList {
		if taskID != "" {
			if err := r.ds.DeleteBuildTask(ctx, taskID); err != nil {
				httpErr := err.(*rest.HTTPError)
				o11y.AddHttpAttrs4HttpError(span, httpErr)
				rest.ReplyError(c, httpErr)
				return
			}
		}
	}

	logger.Debug("Handler DeleteBuildTasks Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, nil)
}

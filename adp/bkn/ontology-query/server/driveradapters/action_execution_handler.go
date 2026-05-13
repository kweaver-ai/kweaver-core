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
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-go-lib/hydra"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/otel/otellog"
	"github.com/kweaver-ai/kweaver-go-lib/otel/oteltrace"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	attr "go.opentelemetry.io/otel/attribute"

	"ontology-query/common/visitor"
	oerrors "ontology-query/errors"
	"ontology-query/interfaces"
)

// ExecuteActionByIn handles action execution request (internal)
func (r *restHandler) ExecuteActionByIn(c *gin.Context) {
	logger.Debug("Handler ExecuteActionByIn Start")
	visitor := visitor.GenerateVisitor(c)
	r.ExecuteAction(c, visitor)
}

// ExecuteActionByEx handles action execution request (external)
func (r *restHandler) ExecuteActionByEx(c *gin.Context) {
	logger.Debug("Handler ExecuteActionByEx Start")
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.ExecuteAction(c, visitor)
}

// ExecuteAction handles the action execution request
func (r *restHandler) ExecuteAction(c *gin.Context, visitor hydra.Visitor) {
	logger.Debug("Handler ExecuteAction Start")
	startTime := time.Now()

	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	// Pass x-business-domain header to context for MCP execution
	businessDomain := c.GetHeader(interfaces.HTTP_HEADER_BUSINESS_DOMAIN)
	ctx = context.WithValue(ctx, interfaces.BUSINESS_DOMAIN_KEY, businessDomain)

	oteltrace.AddHttpAttrs4API(span, oteltrace.GetAttrsByGinCtx(c))
	otellog.LogInfo(ctx, fmt.Sprintf("行动执行请求参数: [%s,%v]", c.Request.RequestURI, c.Request.Body))

	// Get path parameters
	knID := c.Param("kn_id")
	atID := c.Param("at_id")
	branch := c.DefaultQuery("branch", interfaces.MAIN_BRANCH)
	span.SetAttributes(
		attr.Key("kn_id").String(knID),
		attr.Key("at_id").String(atID),
		attr.Key("branch").String(branch),
	)

	// Bind request body
	req := interfaces.ActionExecutionRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_ActionExecution_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("Binding Parameter Failed: %s", err.Error()))

		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		otellog.LogError(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails), nil)
		rest.ReplyError(c, httpErr)
		return
	}

	req.KNID = knID
	req.Branch = branch
	req.ActionTypeID = atID

	// Note: _instance_identities is optional
	// If not provided, the action will apply to all entities matching the action type's conditions

	// Execute action
	result, err := r.ass.ExecuteAction(ctx, &req)
	if err != nil {
		httpErr, ok := err.(*rest.HTTPError)
		if !ok {
			httpErr = rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.OntologyQuery_InternalError).
				WithErrorDetails(err.Error())
		}

		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		otellog.LogError(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails), nil)
		rest.ReplyError(c, httpErr)
		return
	}

	oteltrace.AddHttpAttrs4Ok(span, http.StatusAccepted)
	logger.Debugf("ExecuteAction completed in %dms", time.Since(startTime).Milliseconds())
	rest.ReplyOK(c, http.StatusAccepted, result)
}

// GetActionExecutionByIn handles get execution status request (internal)
func (r *restHandler) GetActionExecutionByIn(c *gin.Context) {
	logger.Debug("Handler GetActionExecutionByIn Start")
	visitor := visitor.GenerateVisitor(c)
	r.GetActionExecution(c, visitor)
}

// GetActionExecutionByEx handles get execution status request (external)
func (r *restHandler) GetActionExecutionByEx(c *gin.Context) {
	logger.Debug("Handler GetActionExecutionByEx Start")
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.GetActionExecution(c, visitor)
}

// GetActionExecution handles the get execution status request
func (r *restHandler) GetActionExecution(c *gin.Context, visitor hydra.Visitor) {
	logger.Debug("Handler GetActionExecution Start")
	startTime := time.Now()

	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	oteltrace.AddHttpAttrs4API(span, oteltrace.GetAttrsByGinCtx(c))

	// Get path parameters
	knID := c.Param("kn_id")
	executionID := c.Param("execution_id")
	span.SetAttributes(
		attr.Key("kn_id").String(knID),
		attr.Key("execution_id").String(executionID),
	)

	// Get execution
	result, err := r.ass.GetExecution(ctx, knID, executionID)
	if err != nil {
		httpErr, ok := err.(*rest.HTTPError)
		if !ok {
			httpErr = rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.OntologyQuery_InternalError).
				WithErrorDetails(err.Error())
		}

		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		otellog.LogError(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails), nil)
		rest.ReplyError(c, httpErr)
		return
	}

	oteltrace.AddHttpAttrs4Ok(span, http.StatusOK)
	logger.Debugf("GetActionExecution completed in %dms", time.Since(startTime).Milliseconds())
	rest.ReplyOK(c, http.StatusOK, result)
}

// QueryActionLogsByIn handles query action logs request (internal)
func (r *restHandler) QueryActionLogsByIn(c *gin.Context) {
	logger.Debug("Handler QueryActionLogsByIn Start")
	visitor := visitor.GenerateVisitor(c)
	r.QueryActionLogs(c, visitor)
}

// QueryActionLogsByEx handles query action logs request (external)
func (r *restHandler) QueryActionLogsByEx(c *gin.Context) {
	logger.Debug("Handler QueryActionLogsByEx Start")
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.QueryActionLogs(c, visitor)
}

// QueryActionLogs handles the query action logs request (GET with query parameters)
func (r *restHandler) QueryActionLogs(c *gin.Context, visitor hydra.Visitor) {
	logger.Debug("Handler QueryActionLogs Start")
	startTime := time.Now()

	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	oteltrace.AddHttpAttrs4API(span, oteltrace.GetAttrsByGinCtx(c))
	otellog.LogInfo(ctx, fmt.Sprintf("行动日志查询请求参数: [%s]", c.Request.RequestURI))

	// Get path parameters
	knID := c.Param("kn_id")
	span.SetAttributes(attr.Key("kn_id").String(knID))

	// Bind query parameters
	query := interfaces.ActionLogQuery{}
	if err := c.ShouldBindQuery(&query); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_ActionExecution_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("Binding Parameter Failed: %s", err.Error()))

		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		otellog.LogError(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails), nil)
		rest.ReplyError(c, httpErr)
		return
	}

	query.KNID = knID

	// Convert GET query params to internal format
	if query.StartTimeFrom > 0 || query.StartTimeTo > 0 {
		query.StartTimeRange = []int64{query.StartTimeFrom, query.StartTimeTo}
	}

	// Parse search_after from comma-separated string
	if query.SearchAfterStr != "" {
		parts := strings.Split(query.SearchAfterStr, ",")
		query.SearchAfter = make([]any, len(parts))
		for i, p := range parts {
			query.SearchAfter[i] = strings.TrimSpace(p)
		}
	}

	// Set default limit
	if query.Limit <= 0 {
		query.Limit = 20
	}
	if query.Limit > 1000 {
		query.Limit = 1000
	}

	// Query executions
	result, err := r.als.QueryExecutions(ctx, &query)
	if err != nil {
		httpErr, ok := err.(*rest.HTTPError)
		if !ok {
			httpErr = rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.OntologyQuery_ActionExecution_QueryExecutionsFailed).
				WithErrorDetails(err.Error())
		}

		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		otellog.LogError(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails), nil)
		rest.ReplyError(c, httpErr)
		return
	}

	oteltrace.AddHttpAttrs4Ok(span, http.StatusOK)
	logger.Debugf("QueryActionLogs completed in %dms", time.Since(startTime).Milliseconds())
	rest.ReplyOK(c, http.StatusOK, result)
}

// GetActionLogByIn handles get single action log request (internal)
func (r *restHandler) GetActionLogByIn(c *gin.Context) {
	logger.Debug("Handler GetActionLogByIn Start")
	visitor := visitor.GenerateVisitor(c)
	r.GetActionLog(c, visitor)
}

// GetActionLogByEx handles get single action log request (external)
func (r *restHandler) GetActionLogByEx(c *gin.Context) {
	logger.Debug("Handler GetActionLogByEx Start")
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.GetActionLog(c, visitor)
}

// GetActionLog handles the get single action log request
func (r *restHandler) GetActionLog(c *gin.Context, visitor hydra.Visitor) {
	logger.Debug("Handler GetActionLog Start")
	startTime := time.Now()

	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	oteltrace.AddHttpAttrs4API(span, oteltrace.GetAttrsByGinCtx(c))

	// Get path parameters
	knID := c.Param("kn_id")
	logID := c.Param("log_id")
	span.SetAttributes(
		attr.Key("kn_id").String(knID),
		attr.Key("log_id").String(logID),
	)

	// Bind query parameters for results pagination
	query := interfaces.ActionLogDetailQuery{}
	if err := c.ShouldBindQuery(&query); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_ActionExecution_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("Binding Parameter Failed: %s", err.Error()))

		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		otellog.LogError(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails), nil)
		rest.ReplyError(c, httpErr)
		return
	}

	query.KNID = knID
	query.LogID = logID

	// Set default values for results pagination
	if query.ResultsLimit <= 0 {
		query.ResultsLimit = 100
	}
	if query.ResultsLimit > 1000 {
		query.ResultsLimit = 1000
	}
	if query.ResultsOffset < 0 {
		query.ResultsOffset = 0
	}

	// Get execution log with pagination
	result, err := r.als.GetExecution(ctx, &query)
	if err != nil {
		httpErr, ok := err.(*rest.HTTPError)
		if !ok {
			httpErr = rest.NewHTTPError(ctx, http.StatusNotFound, oerrors.OntologyQuery_ActionExecution_ExecutionNotFound).
				WithErrorDetails(err.Error())
		}

		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		otellog.LogError(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails), nil)
		rest.ReplyError(c, httpErr)
		return
	}

	oteltrace.AddHttpAttrs4Ok(span, http.StatusOK)
	logger.Debugf("GetActionLog completed in %dms", time.Since(startTime).Milliseconds())
	rest.ReplyOK(c, http.StatusOK, result)
}

// CancelActionLogByIn handles cancel action execution request (internal)
func (r *restHandler) CancelActionLogByIn(c *gin.Context) {
	logger.Debug("Handler CancelActionLogByIn Start")
	visitor := visitor.GenerateVisitor(c)
	r.CancelActionLog(c, visitor)
}

// CancelActionLogByEx handles cancel action execution request (external)
func (r *restHandler) CancelActionLogByEx(c *gin.Context) {
	logger.Debug("Handler CancelActionLogByEx Start")
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.CancelActionLog(c, visitor)
}

// CancelActionLog handles the cancel action execution request
func (r *restHandler) CancelActionLog(c *gin.Context, visitor hydra.Visitor) {
	logger.Debug("Handler CancelActionLog Start")
	startTime := time.Now()

	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	oteltrace.AddHttpAttrs4API(span, oteltrace.GetAttrsByGinCtx(c))

	// Get path parameters
	knID := c.Param("kn_id")
	logID := c.Param("log_id")
	span.SetAttributes(
		attr.Key("kn_id").String(knID),
		attr.Key("log_id").String(logID),
	)

	// Bind request body (optional)
	req := interfaces.CancelExecutionRequest{}
	// Ignore binding errors since request body is optional
	_ = c.ShouldBindJSON(&req)

	// Cancel execution
	result, err := r.als.CancelExecution(ctx, knID, logID, req.Reason)
	if err != nil {
		httpErr, ok := err.(*rest.HTTPError)
		if !ok {
			// Check if it's a "not found" error
			if strings.Contains(err.Error(), "not found") {
				httpErr = rest.NewHTTPError(ctx, http.StatusNotFound, oerrors.OntologyQuery_ActionExecution_ExecutionNotFound).
					WithErrorDetails(err.Error())
			} else if strings.Contains(err.Error(), "cannot be cancelled") {
				httpErr = rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_ActionExecution_InvalidParameter).
					WithErrorDetails(err.Error())
			} else {
				httpErr = rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.OntologyQuery_ActionExecution_CancelExecutionFailed).
					WithErrorDetails(err.Error())
			}
		}

		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		otellog.LogError(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails), nil)
		rest.ReplyError(c, httpErr)
		return
	}

	oteltrace.AddHttpAttrs4Ok(span, http.StatusOK)
	logger.Debugf("CancelActionLog completed in %dms", time.Since(startTime).Milliseconds())
	rest.ReplyOK(c, http.StatusOK, result)
}

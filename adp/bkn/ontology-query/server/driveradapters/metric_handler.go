// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/hydra"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"go.opentelemetry.io/otel/trace"

	"ontology-query/common/visitor"
	oerrors "ontology-query/errors"
	"ontology-query/interfaces"
)

// PostMetricDataByEx 外部：POST .../metrics/:metric_id/data
func (r *restHandler) PostMetricDataByEx(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c), "PostMetricDataByEx", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	vis, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.postMetricData(c, vis)
}

// PostMetricDataByIn 内部：POST .../metrics/:metric_id/data
func (r *restHandler) PostMetricDataByIn(c *gin.Context) {
	r.postMetricData(c, visitor.GenerateVisitor(c))
}

func (r *restHandler) postMetricData(c *gin.Context, vis hydra.Visitor) {
	start := time.Now()
	ctx := rest.GetLanguageCtx(c)
	accountInfo := interfaces.AccountInfo{ID: vis.ID, Type: string(vis.Type)}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	knID := c.Param("kn_id")
	metricID := c.Param("metric_id")
	branch := c.DefaultQuery("branch", interfaces.MAIN_BRANCH)

	// fill_null 范围查询时,对于缺失的步长点是否补空
	fillNullStr := c.DefaultQuery("fill_null", interfaces.DefaultFillNullQuery)
	fillNull, err := strconv.ParseBool(fillNullStr)
	if err != nil {
		rest.ReplyError(c, rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_Metric_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("The fill_null:%s is invalid", fillNullStr)))
		return
	}

	var body interfaces.MetricQueryRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_Metric_InvalidParameter).
			WithErrorDetails("bind json: " + err.Error())
		rest.ReplyError(c, httpErr)
		return
	}
	if err := validateMetricQueryRequest(ctx, &body); err != nil {
		if httpErr, ok := err.(*rest.HTTPError); ok {
			rest.ReplyError(c, httpErr)
			return
		}
		rest.ReplyError(c, rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_Metric_InvalidParameter).WithErrorDetails(err.Error()))
		return
	}

	body.FillNull = fillNull

	out, err := r.ms.QueryMetricData(ctx, knID, branch, metricID, &body)
	if err != nil {
		if httpErr, ok := err.(*rest.HTTPError); ok {
			rest.ReplyError(c, httpErr)
			return
		}
		rest.ReplyError(c, rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.OntologyQuery_Metric_InternalError_QueryFailed).
			WithErrorDetails(err.Error()))
		return
	}

	logger.Infof("PostMetricData done in %dms", time.Since(start).Milliseconds())
	rest.ReplyOK(c, http.StatusOK, out)
}

// PostMetricDryRunByEx 外部：POST .../metrics/dry-run
func (r *restHandler) PostMetricDryRunByEx(c *gin.Context) {
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c), "PostMetricDryRunByEx", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	vis, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.postMetricDryRun(c, vis)
}

// PostMetricDryRunByIn 内部：POST .../metrics/dry-run
func (r *restHandler) PostMetricDryRunByIn(c *gin.Context) {
	r.postMetricDryRun(c, visitor.GenerateVisitor(c))
}

func (r *restHandler) postMetricDryRun(c *gin.Context, vis hydra.Visitor) {
	start := time.Now()
	ctx := rest.GetLanguageCtx(c)
	accountInfo := interfaces.AccountInfo{ID: vis.ID, Type: string(vis.Type)}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	knID := c.Param("kn_id")
	branch := c.DefaultQuery("branch", interfaces.MAIN_BRANCH)

	// fill_null 范围查询时,对于缺失的步长点是否补空
	fillNullStr := c.DefaultQuery("fill_null", interfaces.DefaultFillNullQuery)
	fillNull, err := strconv.ParseBool(fillNullStr)
	if err != nil {
		rest.ReplyError(c, rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_Metric_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("The fill_null:%s is invalid", fillNullStr)))
		return
	}

	var body interfaces.MetricDryRunRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_Metric_InvalidParameter).
			WithErrorDetails("bind json: " + err.Error())
		rest.ReplyError(c, httpErr)
		return
	}

	err = validateMetricDryRunForExecution(ctx, &body)
	if err != nil {
		if httpErr, ok := err.(*rest.HTTPError); ok {
			rest.ReplyError(c, httpErr)
			return
		}
		rest.ReplyError(c, rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_Metric_InvalidParameter).WithErrorDetails(err.Error()))
		return
	}
	body.FillNull = fillNull

	out, err := r.ms.DryRunMetricData(ctx, knID, branch, &body)
	if err != nil {
		if httpErr, ok := err.(*rest.HTTPError); ok {
			rest.ReplyError(c, httpErr)
			return
		}
		rest.ReplyError(c, rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.OntologyQuery_Metric_InternalError_QueryFailed).
			WithErrorDetails(err.Error()))
		return
	}

	logger.Infof("PostMetricDryRun done in %dms", time.Since(start).Milliseconds())
	rest.ReplyOK(c, http.StatusOK, out)
}

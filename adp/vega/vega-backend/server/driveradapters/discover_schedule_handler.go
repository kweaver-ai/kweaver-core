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

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-go-lib/audit"
	"github.com/kweaver-ai/kweaver-go-lib/hydra"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/otel/oteltrace"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/robfig/cron/v3"
	"go.opentelemetry.io/otel/trace"

	"vega-backend/common/visitor"
	verrors "vega-backend/errors"
	"vega-backend/interfaces"
)

// =========================== POST /discover-schedules ===========================

// CreateDiscoverScheduleByEx handles POST /api/vega-backend/v1/discover-schedules (External).
func (r *restHandler) CreateDiscoverScheduleByEx(c *gin.Context) {
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.createDiscoverSchedule(c, ctx, span, visitor)
}

// CreateDiscoverScheduleByIn handles POST /api/vega-backend/in/v1/discover-schedules (Internal).
func (r *restHandler) CreateDiscoverScheduleByIn(c *gin.Context) {
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	visitor := visitor.GenerateVisitor(c)
	r.createDiscoverSchedule(c, ctx, span, visitor)
}

func (r *restHandler) createDiscoverSchedule(c *gin.Context, ctx context.Context, span trace.Span, visitor hydra.Visitor) {
	accountInfo := interfaces.AccountInfo{ID: visitor.ID, Type: string(visitor.Type)}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)
	oteltrace.AddHttpAttrs4API(span, oteltrace.GetAttrsByGinCtx(c))

	var req interfaces.DiscoverScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_RequestBody).
			WithErrorDetails(err.Error())
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	if req.CatalogID == "" {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_RequestBody).
			WithErrorDetails("catalog_id is required")
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	catalog, err := r.cs.GetByID(ctx, req.CatalogID, false)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if catalog == nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Catalog_NotFound)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	if err := validateCronExprAndStrategies(ctx, span, c, req.CronExpr, req.Strategies); err != nil {
		return
	}

	scheduleID, err := r.dss.Create(ctx, &req)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_DiscoverSchedule_InternalError_CreateFailed).
			WithErrorDetails(err.Error())
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	if req.Enabled {
		if err := r.sw.Schedule(scheduleID); err != nil {
			logger.Errorf("Failed to schedule schedule %s: %v", scheduleID, err)
		}
	}

	audit.NewInfoLog(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
		interfaces.GenerateCatalogAuditObject(req.CatalogID, ""), "")

	logger.Debug("Handler CreateDiscoverSchedule Success")
	oteltrace.AddHttpAttrs4Ok(span, http.StatusCreated)
	rest.ReplyOK(c, http.StatusCreated, gin.H{"id": scheduleID})
}

// =========================== GET /discover-schedules ===========================

// ListDiscoverSchedulesByEx handles GET /api/vega-backend/v1/discover-schedules (External).
func (r *restHandler) ListDiscoverSchedulesByEx(c *gin.Context) {
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.listDiscoverSchedules(c, ctx, span, visitor)
}

// ListDiscoverSchedulesByIn handles GET /api/vega-backend/in/v1/discover-schedules (Internal).
func (r *restHandler) ListDiscoverSchedulesByIn(c *gin.Context) {
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	visitor := visitor.GenerateVisitor(c)
	r.listDiscoverSchedules(c, ctx, span, visitor)
}

func (r *restHandler) listDiscoverSchedules(c *gin.Context, ctx context.Context, span trace.Span, visitor hydra.Visitor) {
	accountInfo := interfaces.AccountInfo{ID: visitor.ID, Type: string(visitor.Type)}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)
	oteltrace.AddHttpAttrs4API(span, oteltrace.GetAttrsByGinCtx(c))

	params := interfaces.DiscoverScheduleQueryParams{
		CatalogID: c.Query("catalog_id"),
	}
	if enabledStr := c.Query("enabled"); enabledStr != "" {
		v, err := strconv.ParseBool(enabledStr)
		if err != nil {
			httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_RequestBody).
				WithErrorDetails(fmt.Sprintf("invalid enabled value: %s", enabledStr))
			oteltrace.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
		params.Enabled = &v
	}

	entries, total, err := r.dss.List(ctx, params)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_DiscoverSchedule_InternalError_GetFailed).
			WithErrorDetails(err.Error())
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	oteltrace.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, gin.H{"entries": entries, "total_count": total})
}

// =========================== GET /discover-schedules/:id ===========================

// GetDiscoverScheduleByEx handles GET /api/vega-backend/v1/discover-schedules/:id (External).
func (r *restHandler) GetDiscoverScheduleByEx(c *gin.Context) {
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.getDiscoverSchedule(c, ctx, span, visitor)
}

// GetDiscoverScheduleByIn handles GET /api/vega-backend/in/v1/discover-schedules/:id (Internal).
func (r *restHandler) GetDiscoverScheduleByIn(c *gin.Context) {
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	visitor := visitor.GenerateVisitor(c)
	r.getDiscoverSchedule(c, ctx, span, visitor)
}

func (r *restHandler) getDiscoverSchedule(c *gin.Context, ctx context.Context, span trace.Span, visitor hydra.Visitor) {
	accountInfo := interfaces.AccountInfo{ID: visitor.ID, Type: string(visitor.Type)}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)
	oteltrace.AddHttpAttrs4API(span, oteltrace.GetAttrsByGinCtx(c))

	id := c.Param("id")
	schedule, err := r.dss.GetByID(ctx, id)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_DiscoverSchedule_InternalError_GetFailed).
			WithErrorDetails(err.Error())
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if schedule == nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_DiscoverSchedule_NotFound)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	oteltrace.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, schedule)
}

// =========================== PUT /discover-schedules/:id ===========================

// UpdateDiscoverScheduleByEx handles PUT /api/vega-backend/v1/discover-schedules/:id (External).
func (r *restHandler) UpdateDiscoverScheduleByEx(c *gin.Context) {
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.updateDiscoverSchedule(c, ctx, span, visitor)
}

// UpdateDiscoverScheduleByIn handles PUT /api/vega-backend/in/v1/discover-schedules/:id (Internal).
func (r *restHandler) UpdateDiscoverScheduleByIn(c *gin.Context) {
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	visitor := visitor.GenerateVisitor(c)
	r.updateDiscoverSchedule(c, ctx, span, visitor)
}

func (r *restHandler) updateDiscoverSchedule(c *gin.Context, ctx context.Context, span trace.Span, visitor hydra.Visitor) {
	accountInfo := interfaces.AccountInfo{ID: visitor.ID, Type: string(visitor.Type)}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)
	oteltrace.AddHttpAttrs4API(span, oteltrace.GetAttrsByGinCtx(c))

	id := c.Param("id")

	current, err := r.dss.GetByID(ctx, id)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_DiscoverSchedule_InternalError_GetFailed).
			WithErrorDetails(err.Error())
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if current == nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_DiscoverSchedule_NotFound)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	var req interfaces.DiscoverSchedule
	if err := c.ShouldBindJSON(&req); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_RequestBody).
			WithErrorDetails(err.Error())
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// Strict checks: id / catalog_id / enabled are read-only here.
	if req.ID != "" && req.ID != id {
		httpErr := rest.NewHTTPError(ctx, http.StatusConflict, verrors.VegaBackend_DiscoverSchedule_IdMismatch).
			WithErrorDetails(fmt.Sprintf("body.id=%s does not match path id=%s", req.ID, id))
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if req.CatalogID != "" && req.CatalogID != current.CatalogID {
		httpErr := rest.NewHTTPError(ctx, http.StatusConflict, verrors.VegaBackend_DiscoverSchedule_CatalogMismatch).
			WithErrorDetails(fmt.Sprintf("catalog_id is read-only; current=%s, body=%s", current.CatalogID, req.CatalogID))
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if req.Enabled != current.Enabled {
		httpErr := rest.NewHTTPError(ctx, http.StatusConflict, verrors.VegaBackend_DiscoverSchedule_EnabledFieldNotAllowed).
			WithErrorDetails("use POST /discover-schedules/{id}/enable or /disable to change enabled state")
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	if err := validateCronExprAndStrategies(ctx, span, c, req.CronExpr, req.Strategies); err != nil {
		return
	}

	// Force authoritative fields from path / current state.
	req.ID = id
	req.CatalogID = current.CatalogID
	req.Enabled = current.Enabled

	if err := r.dss.Update(ctx, id, &req); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_DiscoverSchedule_InternalError_UpdateFailed).
			WithErrorDetails(err.Error())
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	if current.Enabled {
		if err := r.sw.Schedule(id); err != nil {
			logger.Errorf("Failed to reschedule schedule %s after update: %v", id, err)
		}
	}

	audit.NewInfoLog(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
		interfaces.GenerateCatalogAuditObject(current.CatalogID, ""), "")

	logger.Debug("Handler UpdateDiscoverSchedule Success")
	oteltrace.AddHttpAttrs4Ok(span, http.StatusNoContent)
	rest.ReplyOK(c, http.StatusNoContent, nil)
}

// =========================== DELETE /discover-schedules/:id ===========================

// DeleteDiscoverScheduleByEx handles DELETE /api/vega-backend/v1/discover-schedules/:id (External).
func (r *restHandler) DeleteDiscoverScheduleByEx(c *gin.Context) {
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.deleteDiscoverSchedule(c, ctx, span, visitor)
}

// DeleteDiscoverScheduleByIn handles DELETE /api/vega-backend/in/v1/discover-schedules/:id (Internal).
func (r *restHandler) DeleteDiscoverScheduleByIn(c *gin.Context) {
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	visitor := visitor.GenerateVisitor(c)
	r.deleteDiscoverSchedule(c, ctx, span, visitor)
}

func (r *restHandler) deleteDiscoverSchedule(c *gin.Context, ctx context.Context, span trace.Span, visitor hydra.Visitor) {
	accountInfo := interfaces.AccountInfo{ID: visitor.ID, Type: string(visitor.Type)}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)
	oteltrace.AddHttpAttrs4API(span, oteltrace.GetAttrsByGinCtx(c))

	id := c.Param("id")

	current, err := r.dss.GetByID(ctx, id)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_DiscoverSchedule_InternalError_GetFailed).
			WithErrorDetails(err.Error())
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if current == nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_DiscoverSchedule_NotFound)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// Unschedule first; ignore error since DB delete is the source of truth.
	_ = r.sw.Unschedule(id)

	if err := r.dss.Delete(ctx, id); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_DiscoverSchedule_InternalError_DeleteFailed).
			WithErrorDetails(err.Error())
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	audit.NewWarnLog(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
		interfaces.GenerateCatalogAuditObject(current.CatalogID, ""), audit.SUCCESS, "")

	oteltrace.AddHttpAttrs4Ok(span, http.StatusNoContent)
	rest.ReplyOK(c, http.StatusNoContent, nil)
}

// =========================== POST /discover-schedules/:id/enable ===========================

// EnableDiscoverScheduleByEx handles POST /api/vega-backend/v1/discover-schedules/:id/enable (External).
func (r *restHandler) EnableDiscoverScheduleByEx(c *gin.Context) {
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.toggleDiscoverSchedule(c, ctx, span, visitor, true)
}

// EnableDiscoverScheduleByIn handles POST /api/vega-backend/in/v1/discover-schedules/:id/enable (Internal).
func (r *restHandler) EnableDiscoverScheduleByIn(c *gin.Context) {
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	visitor := visitor.GenerateVisitor(c)
	r.toggleDiscoverSchedule(c, ctx, span, visitor, true)
}

// DisableDiscoverScheduleByEx handles POST /api/vega-backend/v1/discover-schedules/:id/disable (External).
func (r *restHandler) DisableDiscoverScheduleByEx(c *gin.Context) {
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.toggleDiscoverSchedule(c, ctx, span, visitor, false)
}

// DisableDiscoverScheduleByIn handles POST /api/vega-backend/in/v1/discover-schedules/:id/disable (Internal).
func (r *restHandler) DisableDiscoverScheduleByIn(c *gin.Context) {
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	visitor := visitor.GenerateVisitor(c)
	r.toggleDiscoverSchedule(c, ctx, span, visitor, false)
}

// toggleDiscoverSchedule shared logic for enable / disable.
// Idempotent: re-enable an enabled schedule (or re-disable a disabled one) returns 204 without error.
func (r *restHandler) toggleDiscoverSchedule(c *gin.Context, ctx context.Context, span trace.Span, visitor hydra.Visitor, enable bool) {
	accountInfo := interfaces.AccountInfo{ID: visitor.ID, Type: string(visitor.Type)}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)
	oteltrace.AddHttpAttrs4API(span, oteltrace.GetAttrsByGinCtx(c))

	id := c.Param("id")

	current, err := r.dss.GetByID(ctx, id)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_DiscoverSchedule_InternalError_GetFailed).
			WithErrorDetails(err.Error())
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if current == nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_DiscoverSchedule_NotFound)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	if current.Enabled == enable {
		// Idempotent no-op.
		oteltrace.AddHttpAttrs4Ok(span, http.StatusNoContent)
		rest.ReplyOK(c, http.StatusNoContent, nil)
		return
	}

	if enable {
		if err := r.dss.Enable(ctx, id); err != nil {
			httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_DiscoverSchedule_InternalError_UpdateFailed).
				WithErrorDetails(err.Error())
			oteltrace.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
		if err := r.sw.Schedule(id); err != nil {
			logger.Errorf("Failed to schedule schedule %s: %v", id, err)
		}
	} else {
		if err := r.dss.Disable(ctx, id); err != nil {
			httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, verrors.VegaBackend_DiscoverSchedule_InternalError_UpdateFailed).
				WithErrorDetails(err.Error())
			oteltrace.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
		_ = r.sw.Unschedule(id)
	}

	op := audit.UPDATE
	_ = op
	audit.NewInfoLog(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
		interfaces.GenerateCatalogAuditObject(current.CatalogID, ""), "")

	oteltrace.AddHttpAttrs4Ok(span, http.StatusNoContent)
	rest.ReplyOK(c, http.StatusNoContent, nil)
}

// =========================== helpers ===========================

// validateCronExprAndStrategies validates cron expression and strategies; on failure replies error and returns non-nil.
func validateCronExprAndStrategies(ctx context.Context, span trace.Span, c *gin.Context, cronExpr string, strategies []string) error {
	if cronExpr == "" {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_DiscoverSchedule_InvalidCronExpr).
			WithErrorDetails("cron_expr is required")
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return httpErr
	}
	if _, err := cron.ParseStandard(cronExpr); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_DiscoverSchedule_InvalidCronExpr).
			WithErrorDetails(fmt.Sprintf("invalid cron expression: %v", err))
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return httpErr
	}
	if len(strategies) > 0 {
		if err := validateStrategies(strategies); err != nil {
			httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_DiscoverSchedule_InvalidStrategies).
				WithErrorDetails(err.Error())
			oteltrace.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return httpErr
		}
	}
	return nil
}

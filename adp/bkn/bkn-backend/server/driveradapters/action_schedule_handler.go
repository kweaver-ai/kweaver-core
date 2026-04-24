// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-go-lib/audit"
	"github.com/kweaver-ai/kweaver-go-lib/hydra"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	attr "go.opentelemetry.io/otel/attribute"

	"bkn-backend/common"
	"bkn-backend/common/visitor"
	berrors "bkn-backend/errors"
	"bkn-backend/infra/otel/otellog"
	"bkn-backend/infra/otel/oteltrace"
	"bkn-backend/interfaces"
)

// CreateActionScheduleByIn creates a new action schedule (internal)
func (r *restHandler) CreateActionScheduleByIn(c *gin.Context) {
	logger.Debug("Handler CreateActionScheduleByIn Start")
	visitor := visitor.GenerateVisitor(c)
	r.CreateActionSchedule(c, visitor)
}

// CreateActionScheduleByEx creates a new action schedule (external)
func (r *restHandler) CreateActionScheduleByEx(c *gin.Context) {
	logger.Debug("Handler CreateActionScheduleByEx Start")
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.CreateActionSchedule(c, visitor)
}

// CreateActionSchedule creates a new action schedule (shared logic)
func (r *restHandler) CreateActionSchedule(c *gin.Context, visitor hydra.Visitor) {
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	oteltrace.AddHttpAttrs4API(span, oteltrace.GetAttrsByGinCtx(c))

	knID := c.Param("kn_id")
	branch := c.DefaultQuery("branch", interfaces.MAIN_BRANCH)
	span.SetAttributes(
		attr.Key("kn_id").String(knID),
		attr.Key("branch").String(branch),
	)

	// Verify KN exists
	_, exist, err := r.kns.CheckKNExistByID(ctx, knID, branch)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if !exist {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, berrors.BknBackend_KnowledgeNetwork_NotFound)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// Bind request
	var reqBody interfaces.ActionScheduleCreateRequest
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ActionSchedule_InvalidParameter).
			WithErrorDetails("Binding Parameter Failed: " + err.Error())
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	otellog.LogInfo(ctx, fmt.Sprintf("Create action schedule request: [%s,%v]", c.Request.RequestURI, reqBody))

	// Validate request
	if err := ValidateActionScheduleCreate(ctx, &reqBody); err != nil {
		httpErr := err.(*rest.HTTPError)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// Build schedule object
	schedule := &interfaces.ActionSchedule{
		Name:               reqBody.Name,
		KNID:               knID,
		Branch:             branch,
		ActionTypeID:       reqBody.ActionTypeID,
		CronExpression:     reqBody.CronExpression,
		InstanceIdentities: reqBody.InstanceIdentities,
		DynamicParams:      reqBody.DynamicParams,
		Status:             reqBody.Status,
		Creator:            accountInfo,
		Updater:            accountInfo,
	}

	scheduleID, err := r.ass.CreateSchedule(ctx, schedule)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	audit.NewInfoLog(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
		interfaces.GenerateScheduleAuditObject(scheduleID, reqBody.Name), "")

	result := map[string]any{"id": scheduleID}
	logger.Debug("Handler CreateActionSchedule Success")
	oteltrace.AddHttpAttrs4Ok(span, http.StatusCreated)
	rest.ReplyOK(c, http.StatusCreated, result)
}

// UpdateActionScheduleByIn updates an existing action schedule (internal)
func (r *restHandler) UpdateActionScheduleByIn(c *gin.Context) {
	logger.Debug("Handler UpdateActionScheduleByIn Start")
	visitor := visitor.GenerateVisitor(c)
	r.UpdateActionSchedule(c, visitor)
}

// UpdateActionScheduleByEx updates an existing action schedule (external)
func (r *restHandler) UpdateActionScheduleByEx(c *gin.Context) {
	logger.Debug("Handler UpdateActionScheduleByEx Start")
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.UpdateActionSchedule(c, visitor)
}

// UpdateActionSchedule updates an existing action schedule (shared logic)
func (r *restHandler) UpdateActionSchedule(c *gin.Context, visitor hydra.Visitor) {
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	oteltrace.AddHttpAttrs4API(span, oteltrace.GetAttrsByGinCtx(c))

	knID := c.Param("kn_id")
	branch := c.DefaultQuery("branch", interfaces.MAIN_BRANCH)
	scheduleID := c.Param("schedule_id")
	span.SetAttributes(
		attr.Key("kn_id").String(knID),
		attr.Key("branch").String(branch),
		attr.Key("schedule_id").String(scheduleID),
	)

	// Verify schedule exists and belongs to this KN
	schedule, err := r.ass.GetSchedule(ctx, scheduleID)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if schedule.KNID != knID || schedule.Branch != branch {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, berrors.BknBackend_ActionSchedule_NotFound)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// Bind request
	var reqBody interfaces.ActionScheduleUpdateRequest
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ActionSchedule_InvalidParameter).
			WithErrorDetails("Binding Parameter Failed: " + err.Error())
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	otellog.LogInfo(ctx, fmt.Sprintf("Update action schedule request: [%s,%v]", c.Request.RequestURI, reqBody))

	// Validate request
	if err := ValidateActionScheduleUpdate(ctx, &reqBody); err != nil {
		httpErr := err.(*rest.HTTPError)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	if err := r.ass.UpdateSchedule(ctx, scheduleID, &reqBody); err != nil {
		httpErr := err.(*rest.HTTPError)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	audit.NewInfoLog(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
		interfaces.GenerateScheduleAuditObject(scheduleID, schedule.Name), "")

	logger.Debug("Handler UpdateActionSchedule Success")
	oteltrace.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, nil)
}

// UpdateActionScheduleStatusByIn updates the status of an action schedule (internal)
func (r *restHandler) UpdateActionScheduleStatusByIn(c *gin.Context) {
	logger.Debug("Handler UpdateActionScheduleStatusByIn Start")
	visitor := visitor.GenerateVisitor(c)
	r.UpdateActionScheduleStatus(c, visitor)
}

// UpdateActionScheduleStatusByEx updates the status of an action schedule (external)
func (r *restHandler) UpdateActionScheduleStatusByEx(c *gin.Context) {
	logger.Debug("Handler UpdateActionScheduleStatusByEx Start")
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.UpdateActionScheduleStatus(c, visitor)
}

// UpdateActionScheduleStatus updates the status of an action schedule (shared logic)
func (r *restHandler) UpdateActionScheduleStatus(c *gin.Context, visitor hydra.Visitor) {
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	oteltrace.AddHttpAttrs4API(span, oteltrace.GetAttrsByGinCtx(c))

	knID := c.Param("kn_id")
	branch := c.DefaultQuery("branch", interfaces.MAIN_BRANCH)
	scheduleID := c.Param("schedule_id")
	span.SetAttributes(
		attr.Key("kn_id").String(knID),
		attr.Key("branch").String(branch),
		attr.Key("schedule_id").String(scheduleID),
	)

	// Verify schedule exists
	schedule, err := r.ass.GetSchedule(ctx, scheduleID)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if schedule.KNID != knID || schedule.Branch != branch {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, berrors.BknBackend_ActionSchedule_NotFound)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// Bind request
	var reqBody interfaces.ActionScheduleStatusRequest
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ActionSchedule_InvalidParameter).
			WithErrorDetails("Binding Parameter Failed: " + err.Error())
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	if err := r.ass.UpdateScheduleStatus(ctx, scheduleID, reqBody.Status); err != nil {
		httpErr := err.(*rest.HTTPError)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	audit.NewInfoLog(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
		interfaces.GenerateScheduleAuditObject(scheduleID, schedule.Name), fmt.Sprintf("status: %s", reqBody.Status))

	logger.Debug("Handler UpdateActionScheduleStatus Success")
	oteltrace.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, nil)
}

// DeleteActionSchedulesByIn deletes action schedules (internal)
func (r *restHandler) DeleteActionSchedulesByIn(c *gin.Context) {
	logger.Debug("Handler DeleteActionSchedulesByIn Start")
	visitor := visitor.GenerateVisitor(c)
	r.DeleteActionSchedules(c, visitor)
}

// DeleteActionSchedulesByEx deletes action schedules (external)
func (r *restHandler) DeleteActionSchedulesByEx(c *gin.Context) {
	logger.Debug("Handler DeleteActionSchedulesByEx Start")
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.DeleteActionSchedules(c, visitor)
}

// DeleteActionSchedules deletes action schedules (shared logic)
func (r *restHandler) DeleteActionSchedules(c *gin.Context, visitor hydra.Visitor) {
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	oteltrace.AddHttpAttrs4API(span, oteltrace.GetAttrsByGinCtx(c))

	knID := c.Param("kn_id")
	branch := c.DefaultQuery("branch", interfaces.MAIN_BRANCH)
	scheduleIDsStr := c.Param("schedule_ids")
	span.SetAttributes(
		attr.Key("kn_id").String(knID),
		attr.Key("branch").String(branch),
		attr.Key("schedule_ids").String(scheduleIDsStr),
	)

	scheduleIDs := common.StringToStringSlice(scheduleIDsStr)

	// Get schedules for audit log
	schedules, err := r.ass.GetSchedules(ctx, scheduleIDs)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	if err := r.ass.DeleteSchedules(ctx, knID, branch, scheduleIDs); err != nil {
		httpErr := err.(*rest.HTTPError)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	for _, schedule := range schedules {
		audit.NewWarnLog(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
			interfaces.GenerateScheduleAuditObject(schedule.ID, schedule.Name), audit.SUCCESS, "")
	}

	logger.Debug("Handler DeleteActionSchedules Success")
	oteltrace.AddHttpAttrs4Ok(span, http.StatusNoContent)
	rest.ReplyOK(c, http.StatusNoContent, nil)
}

// ListActionSchedulesByIn lists action schedules (internal)
func (r *restHandler) ListActionSchedulesByIn(c *gin.Context) {
	logger.Debug("Handler ListActionSchedulesByIn Start")
	visitor := visitor.GenerateVisitor(c)
	r.ListActionSchedules(c, visitor)
}

// ListActionSchedulesByEx lists action schedules (external)
func (r *restHandler) ListActionSchedulesByEx(c *gin.Context) {
	logger.Debug("Handler ListActionSchedulesByEx Start")
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.ListActionSchedules(c, visitor)
}

// ListActionSchedules lists action schedules (shared logic)
func (r *restHandler) ListActionSchedules(c *gin.Context, visitor hydra.Visitor) {
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	oteltrace.AddHttpAttrs4API(span, oteltrace.GetAttrsByGinCtx(c))

	knID := c.Param("kn_id")
	branch := c.DefaultQuery("branch", interfaces.MAIN_BRANCH)
	span.SetAttributes(
		attr.Key("kn_id").String(knID),
		attr.Key("branch").String(branch),
	)

	// Verify KN exists
	_, exist, err := r.kns.CheckKNExistByID(ctx, knID, branch)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if !exist {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, berrors.BknBackend_KnowledgeNetwork_NotFound)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// Get query params
	namePattern := c.Query("name_pattern")
	actionTypeID := c.Query("action_type_id")
	status := c.Query("status")
	offset := c.DefaultQuery("offset", interfaces.DEFAULT_OFFEST)
	limit := c.DefaultQuery("limit", interfaces.DEFAULT_LIMIT)
	sort := c.DefaultQuery("sort", "create_time")
	direction := c.DefaultQuery("direction", interfaces.DESC_DIRECTION)

	pageParam, err := validatePaginationQueryParameters(ctx, offset, limit, sort, direction, interfaces.ACTION_SCHEDULE_SORT)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// Validate status if provided
	if status != "" && status != interfaces.ScheduleStatusActive && status != interfaces.ScheduleStatusInactive {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ActionSchedule_InvalidStatus).
			WithErrorDetails(fmt.Sprintf("Invalid status: %s", status))
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	queryParams := interfaces.ActionScheduleQueryParams{
		KNID:         knID,
		Branch:       branch,
		NamePattern:  namePattern,
		ActionTypeID: actionTypeID,
		Status:       status,
	}
	queryParams.Sort = pageParam.Sort
	queryParams.Direction = pageParam.Direction
	queryParams.Limit = pageParam.Limit
	queryParams.Offset = pageParam.Offset

	schedules, total, err := r.ass.ListSchedules(ctx, queryParams)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	result := map[string]any{
		"entries":     schedules,
		"total_count": total,
	}

	logger.Debug("Handler ListActionSchedules Success")
	oteltrace.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, result)
}

// GetActionScheduleByIn gets a single action schedule (internal)
func (r *restHandler) GetActionScheduleByIn(c *gin.Context) {
	logger.Debug("Handler GetActionScheduleByIn Start")
	visitor := visitor.GenerateVisitor(c)
	r.GetActionSchedule(c, visitor)
}

// GetActionScheduleByEx gets a single action schedule (external)
func (r *restHandler) GetActionScheduleByEx(c *gin.Context) {
	logger.Debug("Handler GetActionScheduleByEx Start")
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.GetActionSchedule(c, visitor)
}

// GetActionSchedule gets a single action schedule (shared logic)
func (r *restHandler) GetActionSchedule(c *gin.Context, visitor hydra.Visitor) {
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	oteltrace.AddHttpAttrs4API(span, oteltrace.GetAttrsByGinCtx(c))

	knID := c.Param("kn_id")
	branch := c.DefaultQuery("branch", interfaces.MAIN_BRANCH)
	scheduleID := c.Param("schedule_id")
	span.SetAttributes(
		attr.Key("kn_id").String(knID),
		attr.Key("branch").String(branch),
		attr.Key("schedule_id").String(scheduleID),
	)

	schedule, err := r.ass.GetSchedule(ctx, scheduleID)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	if schedule.KNID != knID || schedule.Branch != branch {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, berrors.BknBackend_ActionSchedule_NotFound)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	logger.Debug("Handler GetActionSchedule Success")
	oteltrace.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, schedule)
}

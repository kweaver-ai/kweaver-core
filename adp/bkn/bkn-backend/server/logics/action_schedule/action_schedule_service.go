// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package action_schedule

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/otel/otellog"
	"github.com/kweaver-ai/kweaver-go-lib/otel/oteltrace"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/robfig/cron/v3"
	"github.com/rs/xid"
	"go.opentelemetry.io/otel/codes"

	"bkn-backend/common"
	berrors "bkn-backend/errors"
	"bkn-backend/interfaces"
	"bkn-backend/logics"
)

var (
	assOnce    sync.Once
	assService interfaces.ActionScheduleService
)

type actionScheduleService struct {
	appSetting *common.AppSetting
	asa        interfaces.ActionScheduleAccess
	ata        interfaces.ActionTypeAccess
	//db         interface{ Begin() (interface{}, error) }

	cronParser cron.Parser
}

// NewActionScheduleService creates a singleton instance of ActionScheduleService
func NewActionScheduleService(appSetting *common.AppSetting) interfaces.ActionScheduleService {
	assOnce.Do(func() {
		assService = &actionScheduleService{
			appSetting: appSetting,
			asa:        logics.ASA,
			ata:        logics.ATA,
			// Standard 5-field cron parser (minute, hour, day of month, month, day of week)
			cronParser: cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow),
		}
	})
	return assService
}

// CreateSchedule creates a new action schedule
func (s *actionScheduleService) CreateSchedule(ctx context.Context, schedule *interfaces.ActionSchedule) (string, error) {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "CreateSchedule")
	defer span.End()

	// Validate cron expression
	if err := s.ValidateCronExpression(schedule.CronExpression); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ActionSchedule_InvalidCronExpression).
			WithErrorDetails(err.Error())
		otellog.LogError(ctx, "Validate cron expression failed", httpErr)
		return "", httpErr
	}

	// Validate action type exists
	actionTypes, err := s.ata.GetActionTypesByIDs(ctx, schedule.KNID, schedule.Branch, []string{schedule.ActionTypeID})
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, berrors.BknBackend_ActionSchedule_GetActionTypeFailed).
			WithErrorDetails(err.Error())
		otellog.LogError(ctx, "Failed to get action type", httpErr)
		return "", httpErr
	}
	if len(actionTypes) == 0 {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, berrors.BknBackend_ActionSchedule_ActionTypeNotFound).
			WithErrorDetails(fmt.Sprintf("Action type not found: %s", schedule.ActionTypeID))
		otellog.LogError(ctx, "Action type not found", httpErr)
		return "", httpErr
	}

	// Generate ID and set defaults
	schedule.ID = xid.New().String()
	now := time.Now().UnixMilli()
	schedule.CreateTime = now
	schedule.UpdateTime = now

	if schedule.Status == "" {
		schedule.Status = interfaces.ScheduleStatusInactive
	}

	// Calculate next run time if status is active
	if schedule.Status == interfaces.ScheduleStatusActive {
		nextRunTime, err := s.CalculateNextRunTime(schedule.CronExpression, now)
		if err != nil {
			httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ActionSchedule_InvalidCronExpression).
				WithErrorDetails(err.Error())
			otellog.LogError(ctx, "Calculate next run time failed", httpErr)
			return "", httpErr
		}
		schedule.NextRunTime = nextRunTime
	}

	// Create in database
	if err := s.asa.CreateSchedule(ctx, nil, schedule); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, berrors.BknBackend_ActionSchedule_CreateFailed).
			WithErrorDetails(err.Error())
		otellog.LogError(ctx, "Failed to create schedule", httpErr)
		return "", httpErr
	}

	logger.Infof("Created schedule: %s", schedule.ID)
	span.SetStatus(codes.Ok, "")
	return schedule.ID, nil
}

// UpdateSchedule updates an existing action schedule
func (s *actionScheduleService) UpdateSchedule(ctx context.Context, scheduleID string, req *interfaces.ActionScheduleUpdateRequest) error {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "UpdateSchedule")
	defer span.End()

	// Check if schedule exists
	existing, err := s.asa.GetSchedule(ctx, scheduleID)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, berrors.BknBackend_ActionSchedule_GetFailed).
			WithErrorDetails(err.Error())
		otellog.LogError(ctx, "Failed to get schedule", httpErr)
		return httpErr
	}
	if existing == nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, berrors.BknBackend_ActionSchedule_NotFound)
		otellog.LogError(ctx, "Schedule not found", httpErr)
		return httpErr
	}

	// Validate cron expression if provided
	cronExpr := existing.CronExpression
	if req.CronExpression != "" {
		if err := s.ValidateCronExpression(req.CronExpression); err != nil {
			httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ActionSchedule_InvalidCronExpression).
				WithErrorDetails(err.Error())
			otellog.LogError(ctx, "Validate cron expression failed", httpErr)
			return httpErr
		}
		cronExpr = req.CronExpression
	}

	// Build update object
	now := time.Now().UnixMilli()
	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}

	update := &interfaces.ActionSchedule{
		ID:             scheduleID,
		Name:           req.Name,
		CronExpression: req.CronExpression,
		Updater:        accountInfo,
		UpdateTime:     now,
	}

	if req.InstanceIdentities != nil {
		update.InstanceIdentities = req.InstanceIdentities
	}
	if req.DynamicParams != nil {
		update.DynamicParams = req.DynamicParams
	}

	// Recalculate next run time if cron changed and schedule is active
	if req.CronExpression != "" && existing.Status == interfaces.ScheduleStatusActive {
		nextRunTime, err := s.CalculateNextRunTime(cronExpr, now)
		if err != nil {
			httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ActionSchedule_InvalidCronExpression).
				WithErrorDetails(err.Error())
			otellog.LogError(ctx, fmt.Sprintf("Failed to calculate next run time for schedule %s", scheduleID), httpErr)
			return httpErr
		}
		update.NextRunTime = nextRunTime
	}

	if err := s.asa.UpdateSchedule(ctx, nil, update); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, berrors.BknBackend_ActionSchedule_UpdateFailed).
			WithErrorDetails(err.Error())
		otellog.LogError(ctx, "Failed to update schedule", httpErr)
		return httpErr
	}

	logger.Infof("Updated schedule: %s", scheduleID)
	span.SetStatus(codes.Ok, "")
	return nil
}

// UpdateScheduleStatus updates the status of a schedule
func (s *actionScheduleService) UpdateScheduleStatus(ctx context.Context, scheduleID string, status string) error {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "UpdateScheduleStatus")
	defer span.End()

	// Validate status
	if status != interfaces.ScheduleStatusActive && status != interfaces.ScheduleStatusInactive {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ActionSchedule_InvalidStatus).
			WithErrorDetails(fmt.Sprintf("Invalid status: %s. Must be 'active' or 'inactive'", status))
		otellog.LogError(ctx, "Invalid schedule status", httpErr)
		return httpErr
	}

	// Check if schedule exists
	existing, err := s.asa.GetSchedule(ctx, scheduleID)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, berrors.BknBackend_ActionSchedule_GetFailed).
			WithErrorDetails(err.Error())
		otellog.LogError(ctx, "Failed to get schedule", httpErr)
		return httpErr
	}
	if existing == nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, berrors.BknBackend_ActionSchedule_NotFound)
		otellog.LogError(ctx, "Schedule not found", httpErr)
		return httpErr
	}

	// Calculate next run time when activating
	var nextRunTime int64
	if status == interfaces.ScheduleStatusActive {
		now := time.Now().UnixMilli()
		nextRunTime, err = s.CalculateNextRunTime(existing.CronExpression, now)
		if err != nil {
			httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ActionSchedule_InvalidCronExpression).
				WithErrorDetails(err.Error())
			otellog.LogError(ctx, "Calculate next run time failed", httpErr)
			return httpErr
		}
	}

	if err := s.asa.UpdateScheduleStatus(ctx, scheduleID, status, nextRunTime); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, berrors.BknBackend_ActionSchedule_UpdateFailed).
			WithErrorDetails(err.Error())
		otellog.LogError(ctx, "Failed to update schedule status", httpErr)
		return httpErr
	}

	logger.Infof("Updated schedule %s status to %s", scheduleID, status)
	span.SetStatus(codes.Ok, "")
	return nil
}

// DeleteSchedules deletes schedules by IDs
func (s *actionScheduleService) DeleteSchedules(ctx context.Context, knID, branch string, scheduleIDs []string) error {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "DeleteSchedules")
	defer span.End()

	if len(scheduleIDs) == 0 {
		span.SetStatus(codes.Ok, "")
		return nil
	}

	// Verify all schedules exist and belong to the kn/branch
	schedules, err := s.asa.GetSchedules(ctx, scheduleIDs)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, berrors.BknBackend_ActionSchedule_GetFailed).
			WithErrorDetails(err.Error())
		otellog.LogError(ctx, "Failed to get schedules", httpErr)
		return httpErr
	}

	for _, id := range scheduleIDs {
		schedule, exists := schedules[id]
		if !exists {
			httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, berrors.BknBackend_ActionSchedule_NotFound).
				WithErrorDetails(fmt.Sprintf("Schedule not found: %s", id))
			otellog.LogError(ctx, "Schedule not found", httpErr)
			return httpErr
		}
		if schedule.KNID != knID || schedule.Branch != branch {
			httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ActionSchedule_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("Schedule %s does not belong to kn %s branch %s", id, knID, branch))
			otellog.LogError(ctx, "Schedule does not belong to request scope", httpErr)
			return httpErr
		}
	}

	if err := s.asa.DeleteSchedules(ctx, nil, scheduleIDs); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, berrors.BknBackend_ActionSchedule_DeleteFailed).
			WithErrorDetails(err.Error())
		otellog.LogError(ctx, "Failed to delete schedules", httpErr)
		return httpErr
	}

	logger.Infof("Deleted schedules: %v", scheduleIDs)
	span.SetStatus(codes.Ok, "")
	return nil
}

// GetSchedule gets a single schedule by ID
func (s *actionScheduleService) GetSchedule(ctx context.Context, scheduleID string) (*interfaces.ActionSchedule, error) {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "GetSchedule")
	defer span.End()

	schedule, err := s.asa.GetSchedule(ctx, scheduleID)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, berrors.BknBackend_ActionSchedule_GetFailed).
			WithErrorDetails(err.Error())
		otellog.LogError(ctx, "Failed to get schedule", httpErr)
		return nil, httpErr
	}
	if schedule == nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, berrors.BknBackend_ActionSchedule_NotFound)
		otellog.LogError(ctx, "Schedule not found", httpErr)
		return nil, httpErr
	}

	span.SetStatus(codes.Ok, "")
	return schedule, nil
}

// GetSchedules gets schedules by IDs
func (s *actionScheduleService) GetSchedules(ctx context.Context, scheduleIDs []string) (map[string]*interfaces.ActionSchedule, error) {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "GetSchedules")
	defer span.End()

	schedules, err := s.asa.GetSchedules(ctx, scheduleIDs)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, berrors.BknBackend_ActionSchedule_GetFailed).
			WithErrorDetails(err.Error())
		otellog.LogError(ctx, "Failed to get schedules", httpErr)
		return nil, httpErr
	}

	span.SetStatus(codes.Ok, "")
	return schedules, nil
}

// ListSchedules lists schedules with pagination
func (s *actionScheduleService) ListSchedules(ctx context.Context, queryParams interfaces.ActionScheduleQueryParams) ([]*interfaces.ActionSchedule, int64, error) {
	ctx, span := oteltrace.StartNamedInternalSpan(ctx, "ListSchedules")
	defer span.End()

	schedules, err := s.asa.ListSchedules(ctx, queryParams)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, berrors.BknBackend_ActionSchedule_GetFailed).
			WithErrorDetails(err.Error())
		otellog.LogError(ctx, "Failed to list schedules", httpErr)
		return nil, 0, httpErr
	}

	total, err := s.asa.GetSchedulesTotal(ctx, queryParams)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusInternalServerError, berrors.BknBackend_ActionSchedule_GetFailed).
			WithErrorDetails(err.Error())
		otellog.LogError(ctx, "Failed to get schedules total", httpErr)
		return nil, 0, httpErr
	}

	span.SetStatus(codes.Ok, "")
	return schedules, total, nil
}

// ValidateCronExpression validates a cron expression
func (s *actionScheduleService) ValidateCronExpression(cronExpr string) error {
	_, err := s.cronParser.Parse(cronExpr)
	if err != nil {
		return fmt.Errorf("invalid cron expression '%s': %v", cronExpr, err)
	}
	return nil
}

// CalculateNextRunTime calculates the next run time based on cron expression
func (s *actionScheduleService) CalculateNextRunTime(cronExpr string, from int64) (int64, error) {
	schedule, err := s.cronParser.Parse(cronExpr)
	if err != nil {
		return 0, fmt.Errorf("invalid cron expression: %v", err)
	}

	fromTime := time.UnixMilli(from)
	nextTime := schedule.Next(fromTime)
	return nextTime.UnixMilli(), nil
}

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

	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/robfig/cron/v3"
	"github.com/rs/xid"
	"go.opentelemetry.io/otel/trace"

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
	ctx, span := ar_trace.Tracer.Start(ctx, "CreateSchedule", trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()

	// Validate cron expression
	if err := s.ValidateCronExpression(schedule.CronExpression); err != nil {
		return "", rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ActionSchedule_InvalidCronExpression).
			WithErrorDetails(err.Error())
	}

	// Validate action type exists
	actionTypes, err := s.ata.GetActionTypesByIDs(ctx, schedule.KNID, schedule.Branch, []string{schedule.ActionTypeID})
	if err != nil {
		logger.Errorf("Failed to get action type: %v", err)
		return "", rest.NewHTTPError(ctx, http.StatusInternalServerError, berrors.BknBackend_ActionSchedule_GetActionTypeFailed).
			WithErrorDetails(err.Error())
	}
	if len(actionTypes) == 0 {
		return "", rest.NewHTTPError(ctx, http.StatusNotFound, berrors.BknBackend_ActionSchedule_ActionTypeNotFound).
			WithErrorDetails(fmt.Sprintf("Action type not found: %s", schedule.ActionTypeID))
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
			return "", rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ActionSchedule_InvalidCronExpression).
				WithErrorDetails(err.Error())
		}
		schedule.NextRunTime = nextRunTime
	}

	// Create in database
	if err := s.asa.CreateSchedule(ctx, nil, schedule); err != nil {
		logger.Errorf("Failed to create schedule: %v", err)
		return "", rest.NewHTTPError(ctx, http.StatusInternalServerError, berrors.BknBackend_ActionSchedule_CreateFailed).
			WithErrorDetails(err.Error())
	}

	logger.Infof("Created schedule: %s", schedule.ID)
	return schedule.ID, nil
}

// UpdateSchedule updates an existing action schedule
func (s *actionScheduleService) UpdateSchedule(ctx context.Context, scheduleID string, req *interfaces.ActionScheduleUpdateRequest) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "UpdateSchedule", trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()

	// Check if schedule exists
	existing, err := s.asa.GetSchedule(ctx, scheduleID)
	if err != nil {
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, berrors.BknBackend_ActionSchedule_GetFailed).
			WithErrorDetails(err.Error())
	}
	if existing == nil {
		return rest.NewHTTPError(ctx, http.StatusNotFound, berrors.BknBackend_ActionSchedule_NotFound)
	}

	// Validate cron expression if provided
	cronExpr := existing.CronExpression
	if req.CronExpression != "" {
		if err := s.ValidateCronExpression(req.CronExpression); err != nil {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ActionSchedule_InvalidCronExpression).
				WithErrorDetails(err.Error())
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
			logger.Errorf("Failed to calculate next run time for schedule %s: %v", scheduleID, err)
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ActionSchedule_InvalidCronExpression).
				WithErrorDetails(err.Error())
		}
		update.NextRunTime = nextRunTime
	}

	if err := s.asa.UpdateSchedule(ctx, nil, update); err != nil {
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, berrors.BknBackend_ActionSchedule_UpdateFailed).
			WithErrorDetails(err.Error())
	}

	logger.Infof("Updated schedule: %s", scheduleID)
	return nil
}

// UpdateScheduleStatus updates the status of a schedule
func (s *actionScheduleService) UpdateScheduleStatus(ctx context.Context, scheduleID string, status string) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "UpdateScheduleStatus", trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()

	// Validate status
	if status != interfaces.ScheduleStatusActive && status != interfaces.ScheduleStatusInactive {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ActionSchedule_InvalidStatus).
			WithErrorDetails(fmt.Sprintf("Invalid status: %s. Must be 'active' or 'inactive'", status))
	}

	// Check if schedule exists
	existing, err := s.asa.GetSchedule(ctx, scheduleID)
	if err != nil {
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, berrors.BknBackend_ActionSchedule_GetFailed).
			WithErrorDetails(err.Error())
	}
	if existing == nil {
		return rest.NewHTTPError(ctx, http.StatusNotFound, berrors.BknBackend_ActionSchedule_NotFound)
	}

	// Calculate next run time when activating
	var nextRunTime int64
	if status == interfaces.ScheduleStatusActive {
		now := time.Now().UnixMilli()
		nextRunTime, err = s.CalculateNextRunTime(existing.CronExpression, now)
		if err != nil {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ActionSchedule_InvalidCronExpression).
				WithErrorDetails(err.Error())
		}
	}

	if err := s.asa.UpdateScheduleStatus(ctx, scheduleID, status, nextRunTime); err != nil {
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, berrors.BknBackend_ActionSchedule_UpdateFailed).
			WithErrorDetails(err.Error())
	}

	logger.Infof("Updated schedule %s status to %s", scheduleID, status)
	return nil
}

// DeleteSchedules deletes schedules by IDs
func (s *actionScheduleService) DeleteSchedules(ctx context.Context, knID, branch string, scheduleIDs []string) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "DeleteSchedules", trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()

	if len(scheduleIDs) == 0 {
		return nil
	}

	// Verify all schedules exist and belong to the kn/branch
	schedules, err := s.asa.GetSchedules(ctx, scheduleIDs)
	if err != nil {
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, berrors.BknBackend_ActionSchedule_GetFailed).
			WithErrorDetails(err.Error())
	}

	for _, id := range scheduleIDs {
		schedule, exists := schedules[id]
		if !exists {
			return rest.NewHTTPError(ctx, http.StatusNotFound, berrors.BknBackend_ActionSchedule_NotFound).
				WithErrorDetails(fmt.Sprintf("Schedule not found: %s", id))
		}
		if schedule.KNID != knID || schedule.Branch != branch {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ActionSchedule_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("Schedule %s does not belong to kn %s branch %s", id, knID, branch))
		}
	}

	if err := s.asa.DeleteSchedules(ctx, nil, scheduleIDs); err != nil {
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, berrors.BknBackend_ActionSchedule_DeleteFailed).
			WithErrorDetails(err.Error())
	}

	logger.Infof("Deleted schedules: %v", scheduleIDs)
	return nil
}

// GetSchedule gets a single schedule by ID
func (s *actionScheduleService) GetSchedule(ctx context.Context, scheduleID string) (*interfaces.ActionSchedule, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "GetSchedule", trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()

	schedule, err := s.asa.GetSchedule(ctx, scheduleID)
	if err != nil {
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, berrors.BknBackend_ActionSchedule_GetFailed).
			WithErrorDetails(err.Error())
	}
	if schedule == nil {
		return nil, rest.NewHTTPError(ctx, http.StatusNotFound, berrors.BknBackend_ActionSchedule_NotFound)
	}

	return schedule, nil
}

// GetSchedules gets schedules by IDs
func (s *actionScheduleService) GetSchedules(ctx context.Context, scheduleIDs []string) (map[string]*interfaces.ActionSchedule, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "GetSchedules", trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()

	schedules, err := s.asa.GetSchedules(ctx, scheduleIDs)
	if err != nil {
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, berrors.BknBackend_ActionSchedule_GetFailed).
			WithErrorDetails(err.Error())
	}

	return schedules, nil
}

// ListSchedules lists schedules with pagination
func (s *actionScheduleService) ListSchedules(ctx context.Context, queryParams interfaces.ActionScheduleQueryParams) ([]*interfaces.ActionSchedule, int64, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "ListSchedules", trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()

	schedules, err := s.asa.ListSchedules(ctx, queryParams)
	if err != nil {
		return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, berrors.BknBackend_ActionSchedule_GetFailed).
			WithErrorDetails(err.Error())
	}

	total, err := s.asa.GetSchedulesTotal(ctx, queryParams)
	if err != nil {
		return nil, 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, berrors.BknBackend_ActionSchedule_GetFailed).
			WithErrorDetails(err.Error())
	}

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

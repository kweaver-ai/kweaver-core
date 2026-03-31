// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import "context"

// ActionScheduleService defines the business logic interface for action schedules
//
//go:generate mockgen -source ../interfaces/action_schedule_service.go -destination ../interfaces/mock/mock_action_schedule_service.go
type ActionScheduleService interface {
	// CRUD operations
	CreateSchedule(ctx context.Context, schedule *ActionSchedule) (string, error)
	UpdateSchedule(ctx context.Context, scheduleID string, req *ActionScheduleUpdateRequest) error
	UpdateScheduleStatus(ctx context.Context, scheduleID string, status string) error
	DeleteSchedules(ctx context.Context, knID, branch string, scheduleIDs []string) error
	GetSchedule(ctx context.Context, scheduleID string) (*ActionSchedule, error)
	GetSchedules(ctx context.Context, scheduleIDs []string) (map[string]*ActionSchedule, error)
	ListSchedules(ctx context.Context, queryParams ActionScheduleQueryParams) ([]*ActionSchedule, int64, error)

	// Cron expression validation and next run time calculation
	ValidateCronExpression(cronExpr string) error
	CalculateNextRunTime(cronExpr string, from int64) (int64, error)
}

// ScheduleExecutor defines the interface for executing scheduled actions
type ScheduleExecutor interface {
	// ExecuteSchedule executes a scheduled action
	ExecuteSchedule(ctx context.Context, schedule *ActionSchedule) (string, error)
}

// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
	"database/sql"
)

// ActionScheduleAccess defines the database access interface for action schedules
//
//go:generate mockgen -source ../interfaces/action_schedule_access.go -destination ../interfaces/mock/mock_action_schedule_access.go
type ActionScheduleAccess interface {
	// CRUD operations
	CreateSchedule(ctx context.Context, tx *sql.Tx, schedule *ActionSchedule) error
	UpdateSchedule(ctx context.Context, tx *sql.Tx, schedule *ActionSchedule) error
	UpdateScheduleStatus(ctx context.Context, scheduleID, status string, nextRunTime int64) error
	DeleteSchedules(ctx context.Context, tx *sql.Tx, scheduleIDs []string) error
	GetSchedule(ctx context.Context, scheduleID string) (*ActionSchedule, error)
	GetSchedules(ctx context.Context, scheduleIDs []string) (map[string]*ActionSchedule, error)
	ListSchedules(ctx context.Context, queryParams ActionScheduleQueryParams) ([]*ActionSchedule, error)
	GetSchedulesTotal(ctx context.Context, queryParams ActionScheduleQueryParams) (int64, error)

	// Lock operations for distributed execution
	// TryAcquireLock attempts to acquire execution lock for a schedule
	// Returns rows affected (1 = success, 0 = failed)
	TryAcquireLock(ctx context.Context, scheduleID, podID string, now, lockTimeout int64) (int64, error)

	// ReleaseLock releases the execution lock and updates run times
	ReleaseLock(ctx context.Context, scheduleID, podID string, lastRunTime, nextRunTime int64) error

	// GetDueSchedules returns schedules that are due for execution
	GetDueSchedules(ctx context.Context, now int64) ([]*ActionSchedule, error)

	// UpdateNextRunTime updates the next run time for a schedule
	UpdateNextRunTime(ctx context.Context, scheduleID string, nextRunTime int64) error
}

// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import "context"

// DiscoverScheduleAccess defines data access interface for scheduled discover schedules.
//
//go:generate mockgen -source ../interfaces/discover_schedule_access.go -destination ../interfaces/mock/mock_discover_schedule_access.go
type DiscoverScheduleAccess interface {
	// Create creates a new discover schedule in database.
	Create(ctx context.Context, schedule *DiscoverSchedule) error
	// GetByID retrieves a discover schedule by ID.
	GetByID(ctx context.Context, id string) (*DiscoverSchedule, error)
	// List lists discover schedules with filters.
	List(ctx context.Context, params DiscoverScheduleQueryParams) ([]*DiscoverSchedule, int64, error)
	// Update updates a discover schedule.
	Update(ctx context.Context, schedule *DiscoverSchedule) error
	// Delete deletes a discover schedule by ID.
	Delete(ctx context.Context, id string) error
	// GetEnabledSchedules retrieves all enabled discover schedules.
	GetEnabledSchedules(ctx context.Context) ([]*DiscoverSchedule, error)
	// UpdateLastRun updates the last run time.
	UpdateLastRun(ctx context.Context, id string, lastRun int64) error
}

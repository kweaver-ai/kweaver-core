// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import "context"

// DiscoverScheduleService defines discover schedule business logic interface.
//
//go:generate mockgen -source ../interfaces/discover_schedule_service.go -destination ../interfaces/mock/mock_discover_schedule_service.go
type DiscoverScheduleService interface {
	// Create creates a new discover schedule.
	Create(ctx context.Context, req *DiscoverScheduleRequest) (string, error)
	// GetByID retrieves a discover schedule by ID.
	GetByID(ctx context.Context, id string) (*DiscoverSchedule, error)
	// List lists discover schedules.
	List(ctx context.Context, params DiscoverScheduleQueryParams) ([]*DiscoverSchedule, int64, error)
	// Update updates a discover schedule.
	Update(ctx context.Context, id string, req *DiscoverSchedule) error
	// Delete deletes a discover schedule by ID.
	Delete(ctx context.Context, id string) error
	// Enable enables a discover schedule.
	Enable(ctx context.Context, id string) error
	// Disable disables a discover schedule.
	Disable(ctx context.Context, id string) error
	// GetEnabledSchedules retrieves all enabled discover schedules.
	GetEnabledSchedules(ctx context.Context) ([]*DiscoverSchedule, error)
	// UpdateLastRun updates the last run time and calculates next run time.
	UpdateLastRun(ctx context.Context, id string, lastRun int64) error
	// ExecuteSchedule executes a discover schedule.
	ExecuteSchedule(ctx context.Context, schedule *DiscoverSchedule) error
}

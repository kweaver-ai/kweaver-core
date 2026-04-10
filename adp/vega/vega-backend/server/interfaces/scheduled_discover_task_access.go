// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import "context"

// ScheduledDiscoverTaskAccess defines data access interface for scheduled discover tasks.
//
//go:generate mockgen -source ../interfaces/scheduled_discover_task_access.go -destination ../interfaces/mock/mock_scheduled_discover_task_access.go
type ScheduledDiscoverTaskAccess interface {
	// Create creates a new scheduled discover task in database.
	Create(ctx context.Context, task *ScheduledDiscoverTask) error
	// GetByID retrieves a scheduled discover task by ID.
	GetByID(ctx context.Context, id string) (*ScheduledDiscoverTask, error)
	// List lists scheduled discover tasks with filters.
	List(ctx context.Context, params ScheduledDiscoverTaskQueryParams) ([]*ScheduledDiscoverTask, int64, error)
	// Update updates a scheduled discover task.
	Update(ctx context.Context, task *ScheduledDiscoverTask) error
	// Delete deletes a scheduled discover task by ID.
	Delete(ctx context.Context, id string) error
	// GetEnabledTasks retrieves all enabled scheduled discover tasks.
	GetEnabledTasks(ctx context.Context) ([]*ScheduledDiscoverTask, error)
	// UpdateLastRun updates the last run time.
	UpdateLastRun(ctx context.Context, id string, lastRun int64) error
}

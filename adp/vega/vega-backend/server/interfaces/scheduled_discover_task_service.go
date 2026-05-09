// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import "context"

// ScheduledDiscoverTask represents a scheduled discover task configuration.
type ScheduledDiscoverTask struct {
	ID         string   `json:"id"`
	CatalogID  string   `json:"catalog_id"`
	CronExpr   string   `json:"cron_expr"`
	StartTime  int64    `json:"start_time"` // Unix timestamp in milliseconds
	EndTime    int64    `json:"end_time"`   // Unix timestamp in milliseconds, 0 means no end time
	Enabled    bool     `json:"enabled"`
	Strategies []string `json:"strategies"` // Strategies for discover: can be one or more of ["insert", "delete", "update"], or empty for all
	LastRun    int64    `json:"last_run"`   // Unix timestamp in milliseconds of last execution
	NextRun    int64    `json:"next_run"`   // Unix timestamp in milliseconds of next scheduled execution

	Creator    AccountInfo `json:"creator"`
	CreateTime int64       `json:"create_time"`
	Updater    AccountInfo `json:"updater"`
	UpdateTime int64       `json:"update_time"`
}

// ScheduledDiscoverTaskQueryParams holds query parameters for scheduled discover tasks.
type ScheduledDiscoverTaskQueryParams struct {
	PaginationQueryParams
	CatalogID string `json:"catalog_id"`
	Enabled   *bool  `json:"enabled"`
}

// ScheduledDiscoverTaskService defines scheduled discover task business logic interface.
//
//go:generate mockgen -source ../interfaces/scheduled_discover_task_service.go -destination ../interfaces/mock/mock_scheduled_discover_task_service.go
type ScheduledDiscoverTaskService interface {
	// Create creates a new scheduled discover task.
	Create(ctx context.Context, req *ScheduledDiscoverTask) (string, error)
	// GetByID retrieves a scheduled discover task by ID.
	GetByID(ctx context.Context, id string) (*ScheduledDiscoverTask, error)
	// List lists scheduled discover tasks.
	List(ctx context.Context, params ScheduledDiscoverTaskQueryParams) ([]*ScheduledDiscoverTask, int64, error)
	// Update updates a scheduled discover task.
	Update(ctx context.Context, id string, req *ScheduledDiscoverTask) error
	// Delete deletes a scheduled discover task by ID.
	Delete(ctx context.Context, id string) error
	// Enable enables a scheduled discover task.
	Enable(ctx context.Context, id string) error
	// Disable disables a scheduled discover task.
	Disable(ctx context.Context, id string) error
	// GetEnabledTasks retrieves all enabled scheduled discover tasks.
	GetEnabledTasks(ctx context.Context) ([]*ScheduledDiscoverTask, error)
	// UpdateLastRun updates the last run time and calculates next run time.
	UpdateLastRun(ctx context.Context, id string, lastRun int64) error
	// ExecuteTask executes a scheduled discover task by creating a discover task.
	ExecuteTask(ctx context.Context, task *ScheduledDiscoverTask) error
}

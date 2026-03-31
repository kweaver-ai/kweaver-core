// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package interfaces defines entities, DTOs, and service interfaces.
package interfaces

import "context"

// DiscoverTaskAccess defines discover task data access interface.
//
//go:generate mockgen -source ../interfaces/discover_task_access.go -destination ../interfaces/mock/mock_discover_task_access.go
type DiscoverTaskAccess interface {
	// Create creates a new DiscoverTask.
	Create(ctx context.Context, task *DiscoverTask) error
	// GetByID retrieves a DiscoverTask by ID.
	GetByID(ctx context.Context, id string) (*DiscoverTask, error)
	// List lists DiscoverTasks with filters.
	List(ctx context.Context, params DiscoverTaskQueryParams) ([]*DiscoverTask, int64, error)
	// UpdateStatus updates a DiscoverTask's status and message.
	UpdateStatus(ctx context.Context, id, status, message string, stime int64) error
	// UpdateProgress updates a DiscoverTask's progress.
	UpdateProgress(ctx context.Context, id string, progress int) error
	// UpdateResult updates a DiscoverTask's result and sets status to completed.
	UpdateResult(ctx context.Context, id string, result *DiscoverResult, stime int64) error

	// CheckExistByStatuses checks if DiscoverTasks exists by catalog ID and statuses.
	CheckExistByStatuses(ctx context.Context, catalogID string, statuses []string) (bool, error)
}

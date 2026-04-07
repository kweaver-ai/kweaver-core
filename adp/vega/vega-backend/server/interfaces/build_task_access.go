// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import "context"

// BuildTaskAccess defines build task data access interface.
type BuildTaskAccess interface {
	// Create creates a new build task.
	Create(ctx context.Context, buildTask *BuildTask) error
	// GetByID retrieves a build task by ID.
	GetByID(ctx context.Context, id string) (*BuildTask, error)
	// GetByResourceID retrieves build tasks by resource ID.
	GetByResourceID(ctx context.Context, resourceID string) ([]*BuildTask, error)
	// CheckResourceHasUncompletedTasks checks if resource has uncompleted build tasks by resource ID.
	CheckResourceHasUncompletedTasks(ctx context.Context, resourceID string) (bool, error)
	// UpdateStatus updates a build task's status and other fields.
	UpdateStatus(ctx context.Context, id string, updates map[string]interface{}) error
	// GetUncompletedTasks retrieves uncompleted build tasks (pending and running) with limit.
	GetUncompletedTasks(ctx context.Context, limit int) ([]*BuildTask, error)
	// GetLastBuildTask retrieves the last completed build task with non-empty synced mark by resource ID.
	GetLastBuildTask(ctx context.Context, resourceID string) (*BuildTask, error)
}

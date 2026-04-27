// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import "context"

//go:generate mockgen -source ../interfaces/build_task_access.go -destination ../interfaces/mock/mock_build_task_access.go

// BuildTaskAccess defines build task data access interface.
type BuildTaskAccess interface {
	// Create creates a new build task.
	Create(ctx context.Context, buildTask *BuildTask) error
	// GetByID retrieves a build task by ID.
	GetByID(ctx context.Context, id string) (*BuildTask, error)
	// GetByResourceID retrieves a build task by resource ID.
	GetByResourceID(ctx context.Context, resourceID string) (*BuildTask, error)
	// GetAll retrieves all build tasks with pagination.
	GetAll(ctx context.Context, offset, limit int) ([]*BuildTask, int64, error)
	// UpdateStatus updates a build task's status and other fields.
	UpdateStatus(ctx context.Context, id string, updates map[string]interface{}) error
	// GetStatus retrieves the status of a build task by ID.
	GetStatus(ctx context.Context, id string) (string, error)
	// Delete deletes a build task by ID.
	Delete(ctx context.Context, id string) error
}

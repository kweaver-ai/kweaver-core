// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import "context"

//go:generate mockgen -source ../interfaces/build_task_service.go -destination ../interfaces/mock/mock_build_task_service.go

// BuildTaskService defines build task business logic interface.
type BuildTaskService interface {
	// CreateBuildTask creates a new build task. resource_id and mode come from req.
	CreateBuildTask(ctx context.Context, req *CreateBuildTaskRequest) (string, error)
	// GetBuildTaskByID retrieves a build task by ID.
	GetBuildTaskByID(ctx context.Context, id string) (*BuildTask, error)
	// GetBuildTaskByResourceID retrieves a build task by resource ID.
	GetBuildTaskByResourceID(ctx context.Context, resourceID string) (*BuildTask, error)
	// ListBuildTasks retrieves build tasks with filters and pagination.
	ListBuildTasks(ctx context.Context, params BuildTasksQueryParams) ([]*BuildTask, int64, error)
	// StartBuildTask transitions a task from {init, stopped} to running (asynchronous; status persisted by worker).
	StartBuildTask(ctx context.Context, taskID string, executeType string) (*BuildTask, error)
	// StopBuildTask transitions a task from running to stopping (asynchronous; status persisted by worker).
	StopBuildTask(ctx context.Context, taskID string) (*BuildTask, error)
	// DeleteBuildTasks atomically deletes build tasks by IDs.
	// Pre-validates: any missing id returns 404 unless ignoreMissing=true; any running/stopping id returns 409 (cannot be skipped).
	DeleteBuildTasks(ctx context.Context, ids []string, ignoreMissing bool) error
}

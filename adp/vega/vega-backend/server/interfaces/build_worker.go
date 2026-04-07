// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"

	"github.com/hibiken/asynq"
)

// BuildWorker interface defines build execution functionality.
// This worker is called by the task management service to execute the actual build.
//
//go:generate mockgen -source ../interfaces/build_task.go -destination ../interfaces/mock/mock_build_worker.go

// BuildTaskMessage represents a build task message.
type BuildTaskMessage struct {
	TaskID     string `json:"task_id"`
}

type BuildWorker interface {
	Start()

	Run(ctx context.Context) error
	ProcessTask(ctx context.Context, event *asynq.Task) error
}

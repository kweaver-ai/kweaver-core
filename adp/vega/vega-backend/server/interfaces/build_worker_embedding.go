// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"

	"github.com/hibiken/asynq"
)

// EmbeddingBuildWorker interface defines embedding execution functionality.
// This worker is called by the task management service to execute the actual embedding.
//
//go:generate mockgen -source ../interfaces/embedding_build_task.go -destination ../interfaces/mock/mock_embedding_build_worker.go

// EmbeddingBuildTaskMessage represents an embedding task message.
type EmbeddingBuildTaskMessage struct {
	TaskID string `json:"task_id"`
}

type EmbeddingBuildWorker interface {
	Start()

	Run(ctx context.Context) error
	ProcessTask(ctx context.Context, event *asynq.Task) error
}

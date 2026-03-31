// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"

	"github.com/hibiken/asynq"
)

// EmbeddingWorker interface defines embedding execution functionality.
// This worker is called by the task management service to execute the actual embedding.
//
//go:generate mockgen -source ../interfaces/embedding_task.go -destination ../interfaces/mock/mock_embedding_worker.go

// EmbeddingTaskMessage represents an embedding task message.
type EmbeddingTaskMessage struct {
	TaskID     string `json:"task_id"`
	ResourceID string `json:"resource_id"`
}

type EmbeddingWorker interface {
	Start()

	Run(ctx context.Context) error
	ProcessTask(ctx context.Context, event *asynq.Task) error
}

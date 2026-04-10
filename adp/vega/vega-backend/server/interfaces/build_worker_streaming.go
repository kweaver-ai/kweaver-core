// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"

	"github.com/hibiken/asynq"
)

var ConnectorClassMapping = map[string]string{
	"mysql": "io.debezium.connector.mysql.MySqlConnector",
}

// StreamingBuildWorker interface defines streaming execution functionality.
// This worker is called by the task management service to execute the actual streaming.
//
//go:generate mockgen -source ../interfaces/build_task.go -destination ../interfaces/mock/mock_streaming_build_worker.go

// StreamingBuildTaskMessage represents a streaming task message.
type StreamingBuildTaskMessage struct {
	TaskID string `json:"task_id"`
}

type StreamingBuildWorker interface {
	Start()

	Run(ctx context.Context) error
	ProcessTask(ctx context.Context, event *asynq.Task) error
}

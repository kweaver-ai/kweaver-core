// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package interfaces defines interfaces for VEGA Manager.
package interfaces

import (
	"context"

	"github.com/hibiken/asynq"
)

// TaskHandler defines the interface for task handlers.
type TaskHandler interface {
	// HandleTask handles a task from the queue.
	HandleTask(ctx context.Context, task *asynq.Task) error
}

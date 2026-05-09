// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package worker provides background workers for VEGA Manager.
package worker

import (
	"context"
	"sync"

	"github.com/hibiken/asynq"

	"vega-backend/common"
	"vega-backend/interfaces"
)

var (
	stWorkerOnce sync.Once
	stWorker     interfaces.StreamingBuildWorker
)

// streamingBuildWorker provides build functionality.
type streamingBuildWorker struct {
	appSetting    *common.AppSetting
	taskWorkerMgr *TaskWorkerManger
}

// NewStreamingBuildWorker creates or returns the singleton StreamingBuildWorker.
func NewStreamingBuildWorker(appSetting *common.AppSetting) interfaces.StreamingBuildWorker {
	stWorkerOnce.Do(func() {
		stWorker = &streamingBuildWorker{
			appSetting:    appSetting,
			taskWorkerMgr: NewTaskWorkerManager(appSetting),
		}
	})
	return stWorker
}

func (bw *streamingBuildWorker) Start() {
	// Start the unified task worker
	bw.taskWorkerMgr.Start()
}

func (bw *streamingBuildWorker) Run(ctx context.Context) error {
	// Delegate to the unified task worker
	return bw.taskWorkerMgr.Run(ctx)
}

func (bw *streamingBuildWorker) ProcessTask(ctx context.Context, task *asynq.Task) error {
	// Delegate to the unified task worker
	return bw.taskWorkerMgr.ProcessTask(ctx, task)
}

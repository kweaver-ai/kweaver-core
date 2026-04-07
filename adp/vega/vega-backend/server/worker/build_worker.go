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
	bWorkerOnce sync.Once
	bWorker     interfaces.BuildWorker
)

// buildWorker provides build functionality.
type buildWorker struct {
	appSetting *common.AppSetting
	taskWorker *TaskWorker
}

// NewBuildWorker creates or returns the singleton BuildWorker.
func NewBuildWorker(appSetting *common.AppSetting) interfaces.BuildWorker {
	bWorkerOnce.Do(func() {
		bWorker = &buildWorker{
			appSetting: appSetting,
			taskWorker: NewTaskWorker(appSetting),
		}
	})
	return bWorker
}

func (bw *buildWorker) Start() {
	// Start the unified task worker
	bw.taskWorker.Start()
}

func (bw *buildWorker) Run(ctx context.Context) error {
	// Delegate to the unified task worker
	return bw.taskWorker.Run(ctx)
}

func (bw *buildWorker) ProcessTask(ctx context.Context, task *asynq.Task) error {
	// Delegate to the unified task worker
	return bw.taskWorker.ProcessTask(ctx, task)
}

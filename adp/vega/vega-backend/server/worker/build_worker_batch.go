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
	bWorker     interfaces.BatchBuildWorker
)

// batchBuildWorker provides build functionality.
type batchBuildWorker struct {
	appSetting    *common.AppSetting
	taskWorkerMgr *TaskWorkerManger
}

// NewBatchBuildWorker creates or returns the singleton BatchBuildWorker.
func NewBatchBuildWorker(appSetting *common.AppSetting) interfaces.BatchBuildWorker {
	bWorkerOnce.Do(func() {
		bWorker = &batchBuildWorker{
			appSetting:    appSetting,
			taskWorkerMgr: NewTaskWorkerManager(appSetting),
		}
	})
	return bWorker
}

func (bw *batchBuildWorker) Start() {
	// Start the unified task worker
	bw.taskWorkerMgr.Start()
}

func (bw *batchBuildWorker) Run(ctx context.Context) error {
	// Delegate to the unified task worker
	return bw.taskWorkerMgr.Run(ctx)
}

func (bw *batchBuildWorker) ProcessTask(ctx context.Context, task *asynq.Task) error {
	// Delegate to the unified task worker
	return bw.taskWorkerMgr.ProcessTask(ctx, task)
}

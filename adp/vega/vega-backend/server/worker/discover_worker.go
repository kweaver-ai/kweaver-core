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
	dWorkerOnce sync.Once
	dWorker     interfaces.DiscoverWorker
)

// discoverWorker provides resource discover functionality.
type discoverWorker struct {
	appSetting *common.AppSetting
	taskWorker *TaskWorker
}

// NewDiscoverWorker creates or returns the singleton DiscoverWorker.
func NewDiscoverWorker(appSetting *common.AppSetting) interfaces.DiscoverWorker {
	dWorkerOnce.Do(func() {
		dWorker = &discoverWorker{
			appSetting: appSetting,
			taskWorker: NewTaskWorker(appSetting),
		}
	})
	return dWorker
}

func (dw *discoverWorker) Start() {
	// Start the unified task worker
	dw.taskWorker.Start()
}

func (dw *discoverWorker) Run(ctx context.Context) error {
	// Delegate to the unified task worker
	return dw.taskWorker.Run(ctx)
}

func (dw *discoverWorker) ProcessTask(ctx context.Context, task *asynq.Task) error {
	// Delegate to the unified task worker
	return dw.taskWorker.ProcessTask(ctx, task)
}

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
	eWorkerOnce sync.Once
	eWorker     interfaces.EmbeddingBuildWorker
)

// embeddingWorker provides embedding functionality.
type embeddingWorker struct {
	appSetting    *common.AppSetting
	taskWorkerMgr *TaskWorkerManger
}

// NewEmbeddingWorker creates or returns the singleton EmbeddingBuildWorker.
func NewEmbeddingWorker(appSetting *common.AppSetting) interfaces.EmbeddingBuildWorker {
	eWorkerOnce.Do(func() {
		eWorker = &embeddingWorker{
			appSetting:    appSetting,
			taskWorkerMgr: NewTaskWorkerManager(appSetting),
		}
	})
	return eWorker
}

func (ew *embeddingWorker) Start() {
	// Start the unified task worker
	ew.taskWorkerMgr.Start()
}

func (ew *embeddingWorker) Run(ctx context.Context) error {
	// Delegate to the unified task worker
	return ew.taskWorkerMgr.Run(ctx)
}

func (ew *embeddingWorker) ProcessTask(ctx context.Context, task *asynq.Task) error {
	// Delegate to the unified task worker
	return ew.taskWorkerMgr.ProcessTask(ctx, task)
}

// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package worker provides background workers for VEGA Manager.
package worker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/hibiken/asynq"
	"github.com/kweaver-ai/kweaver-go-lib/logger"

	"vega-backend/common"
	asynq_access "vega-backend/drivenadapters/asynq"
	"vega-backend/interfaces"
)

var (
	taskWorkerOnce sync.Once
	taskWorker     *TaskWorker
)

// TaskWorker provides unified task processing functionality.
type TaskWorker struct {
	appSetting       *common.AppSetting
	aqa              interfaces.AsynqAccess
	discoverHandler  *discoverHandler
	buildHandler     *buildHandler
	embeddingHandler *embeddingHandler
}

// NewTaskWorker creates or returns the singleton TaskWorker.
func NewTaskWorker(appSetting *common.AppSetting) *TaskWorker {
	taskWorkerOnce.Do(func() {
		taskWorker = &TaskWorker{
			appSetting:       appSetting,
			aqa:              asynq_access.NewAsynqAccess(appSetting),
			discoverHandler:  NewDiscoverHandler(appSetting),
			buildHandler:     NewBuildHandler(appSetting),
			embeddingHandler: NewEmbeddingHandler(appSetting),
		}
	})
	return taskWorker
}

// Start starts the task worker.
func (tw *TaskWorker) Start() {
	// Start server in a goroutine
	go func() {
		for {
			if err := tw.Run(context.Background()); err != nil {
				logger.Errorf("Task worker failed: %v", err)
			}
			time.Sleep(1 * time.Second)
		}
	}()
}

// Run runs the task worker.
func (tw *TaskWorker) Run(ctx context.Context) error {
	defer func() {
		if err := recover(); err != nil {
			logger.Errorf("Task worker failed: %v", err)
		}
	}()

	srv := tw.aqa.CreateServer(ctx)

	// Register task handlers
	mux := asynq.NewServeMux()
	mux.Handle(interfaces.DiscoverTaskType, tw)
	mux.Handle(interfaces.BuildTaskType, tw)
	mux.Handle(interfaces.EmbeddingTaskType, tw)

	logger.Infof("Task worker starting, listening for task types: %s, %s, %s", interfaces.DiscoverTaskType, interfaces.BuildTaskType, interfaces.EmbeddingTaskType)
	if err := srv.Run(mux); err != nil {
		logger.Errorf("Task worker failed: %v", err)
		return err
	}
	return nil
}

// ProcessTask processes a task from the queue.
func (tw *TaskWorker) ProcessTask(ctx context.Context, task *asynq.Task) error {
	switch task.Type() {
	case interfaces.DiscoverTaskType:
		return tw.discoverHandler.HandleTask(ctx, task)
	case interfaces.BuildTaskType:
		return tw.buildHandler.HandleTask(ctx, task)
	case interfaces.EmbeddingTaskType:
		return tw.embeddingHandler.HandleTask(ctx, task)
	default:
		return fmt.Errorf("unknown task type: %s", task.Type())
	}
}

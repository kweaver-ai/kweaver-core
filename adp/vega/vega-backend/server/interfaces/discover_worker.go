// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"

	"github.com/hibiken/asynq"
)

// DiscoverResult represents the result of a discover operation.
type DiscoverResult struct {
	CatalogID      string `json:"catalog_id"`
	NewCount       int    `json:"new_count"`
	StaleCount     int    `json:"stale_count"`
	UnchangedCount int    `json:"unchanged_count"`
	Message        string `json:"message"`
}

// DiscoverWorker interface defines discover execution functionality.
// This worker is called by the task management service to execute the actual discover.
//
//go:generate mockgen -source ../interfaces/discover_worker.go -destination ../interfaces/mock/mock_discover_worker.go
type DiscoverWorker interface {
	Start()

	Run(ctx context.Context) error
	ProcessTask(ctx context.Context, event *asynq.Task) error
}

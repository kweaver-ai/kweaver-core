// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package interfaces defines entities, DTOs, and service interfaces.
package interfaces

import (
	"context"

	"github.com/hibiken/asynq"
)

const (
	HighQueue    = "vega-backend-high"
	DefaultQueue = "vega-backend-default"
	LowQueue     = "vega-backend-low"
)

// AsynqAccess defines the interface for creating Asynq client and server.
//
//go:generate mockgen -source ../interfaces/asynq_access.go -destination ../interfaces/mock/mock_asynq_access.go
type AsynqAccess interface {
	// CreateClient creates and returns the Asynq client for enqueueing tasks.
	CreateClient(ctx context.Context) *asynq.Client
	// CreateServer creates and returns the Asynq server for processing tasks.
	CreateServer(ctx context.Context) *asynq.Server
}

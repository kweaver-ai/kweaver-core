// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package interfaces defines entities, DTOs, and service interfaces.
package interfaces

import "context"

// DiscoverTaskService defines discover task business logic interface.
//
//go:generate mockgen -source ../interfaces/discover_task_service.go -destination ../interfaces/mock/mock_discover_task_service.go
type DiscoverTaskService interface {
	// Create creates a new DiscoverTask and sends message to Kafka.
	// taskType is optional, defaults to DiscoverTaskTriggerManual if not provided.
	Create(ctx context.Context, catalogID string, taskType ...string) (string, error)
	// GetByID retrieves a DiscoverTask by ID.
	GetByID(ctx context.Context, id string) (*DiscoverTask, error)
	// GetByScheduledID retrieves DiscoverTasks by scheduled ID.
	GetByScheduledID(ctx context.Context, scheduledID string) ([]*DiscoverTask, error)
	// List lists DiscoverTasks for a catalog.
	List(ctx context.Context, params DiscoverTaskQueryParams) ([]*DiscoverTask, int64, error)
	// UpdateStatus updates a DiscoverTask's status.
	UpdateStatus(ctx context.Context, id string, status string, message string, stime int64) error
	// UpdateResult updates a DiscoverTask's result.
	UpdateResult(ctx context.Context, id string, result *DiscoverResult, stime int64) error

	// CheckExistByStatuses  checks if DiscoverTasks exists by catalog ID and statuses.
	CheckExistByStatuses(ctx context.Context, catalogID string, statuses []string) (bool, error)
}

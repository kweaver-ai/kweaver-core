// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import "context"

// ActionLogsService defines the interface for managing action execution logs
//
//go:generate mockgen -source ../interfaces/action_logs_service.go -destination ../interfaces/mock/mock_action_logs_service.go
type ActionLogsService interface {
	// CreateExecution creates a new execution record
	CreateExecution(ctx context.Context, exec *ActionExecution) error

	// UpdateExecution updates an existing execution record
	// The updates map should contain field names as keys and new values as values
	UpdateExecution(ctx context.Context, knID, execID string, updates map[string]any) error

	// GetExecution retrieves a single execution by ID with optional results pagination
	GetExecution(ctx context.Context, query *ActionLogDetailQuery) (*ActionExecution, error)

	// QueryExecutions queries executions based on filter criteria
	QueryExecutions(ctx context.Context, query *ActionLogQuery) (*ActionExecutionList, error)

	// CancelExecution cancels a running or pending execution
	CancelExecution(ctx context.Context, knID, execID, reason string) (*CancelExecutionResponse, error)
}

// OpenSearch index name pattern for action executions
const ActionExecutionIndexPrefix = "ontology_action_executions_"

// GetActionExecutionIndex returns the OpenSearch index name for a knowledge network
func GetActionExecutionIndex(knID string) string {
	return ActionExecutionIndexPrefix + knID
}

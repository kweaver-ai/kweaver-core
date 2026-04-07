// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package action_logs

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	attr "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"ontology-query/common"
	"ontology-query/interfaces"
	"ontology-query/logics"
)

var (
	alsOnce    sync.Once
	alsService interfaces.ActionLogsService
)

type actionLogsService struct {
	appSetting *common.AppSetting
	osAccess   interfaces.OpenSearchAccess
}

// NewActionLogsService creates a singleton instance of ActionLogsService
func NewActionLogsService(appSetting *common.AppSetting) interfaces.ActionLogsService {
	alsOnce.Do(func() {
		alsService = &actionLogsService{
			appSetting: appSetting,
			osAccess:   logics.OSA,
		}
	})
	return alsService
}

// CreateExecution creates a new execution record in OpenSearch
func (s *actionLogsService) CreateExecution(ctx context.Context, exec *interfaces.ActionExecution) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "CreateExecution", trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()

	span.SetAttributes(
		attr.Key("execution_id").String(exec.ID),
		attr.Key("kn_id").String(exec.KNID),
		attr.Key("action_type_id").String(exec.ActionTypeID),
	)

	indexName := interfaces.GetActionExecutionIndex(exec.KNID)

	// Ensure index exists
	if err := s.ensureIndexExists(ctx, indexName); err != nil {
		logger.Errorf("Failed to ensure index exists: %v", err)
		return fmt.Errorf("failed to ensure index exists: %w", err)
	}

	// Insert the execution record
	if err := s.osAccess.InsertData(ctx, indexName, exec.ID, exec); err != nil {
		logger.Errorf("Failed to insert execution record: %v", err)
		return fmt.Errorf("failed to insert execution record: %w", err)
	}

	logger.Debugf("Created execution record: %s", exec.ID)
	return nil
}

// UpdateExecution updates an existing execution record
func (s *actionLogsService) UpdateExecution(ctx context.Context, knID, execID string, updates map[string]any) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "UpdateExecution", trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()

	span.SetAttributes(
		attr.Key("execution_id").String(execID),
		attr.Key("kn_id").String(knID),
	)

	// Get the current execution (without pagination for full data)
	query := &interfaces.ActionLogDetailQuery{
		KNID:         knID,
		LogID:        execID,
		ResultsLimit: 10000, // Get all results for update
	}
	exec, err := s.GetExecution(ctx, query)
	if err != nil {
		return err
	}

	// Apply updates
	execMap := structToMap(exec)
	for k, v := range updates {
		execMap[k] = v
	}

	indexName := interfaces.GetActionExecutionIndex(knID)

	// Re-insert with updated values (OpenSearch index API is upsert)
	if err := s.osAccess.InsertData(ctx, indexName, execID, execMap); err != nil {
		logger.Errorf("Failed to update execution record: %v", err)
		return fmt.Errorf("failed to update execution record: %w", err)
	}

	logger.Debugf("Updated execution record: %s", execID)
	return nil
}

// GetExecution retrieves a single execution by ID with optional results pagination
func (s *actionLogsService) GetExecution(ctx context.Context, query *interfaces.ActionLogDetailQuery) (*interfaces.ActionExecution, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "GetExecution", trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()

	span.SetAttributes(
		attr.Key("execution_id").String(query.LogID),
		attr.Key("kn_id").String(query.KNID),
	)

	indexName := interfaces.GetActionExecutionIndex(query.KNID)

	// Build query to get by ID
	osQuery := map[string]any{
		"query": map[string]any{
			"term": map[string]any{
				"id": query.LogID,
			},
		},
		"size": 1,
	}

	hits, err := s.osAccess.SearchData(ctx, indexName, osQuery)
	if err != nil {
		logger.Errorf("Failed to search execution: %v", err)
		return nil, fmt.Errorf("failed to search execution: %w", err)
	}

	if len(hits) == 0 {
		return nil, fmt.Errorf("execution not found: %s", query.LogID)
	}

	exec, err := mapToActionExecution(hits[0].Source)
	if err != nil {
		return nil, fmt.Errorf("failed to parse execution: %w", err)
	}

	// Apply results pagination and filtering
	allResults := exec.Results
	resultsTotal := len(allResults)

	// Filter by status if specified
	if query.ResultsStatus != "" {
		filteredResults := make([]interfaces.ObjectExecutionResult, 0)
		for _, r := range allResults {
			if r.Status == query.ResultsStatus {
				filteredResults = append(filteredResults, r)
			}
		}
		allResults = filteredResults
		resultsTotal = len(allResults)
	}

	// Apply pagination
	resultsLimit := query.ResultsLimit
	if resultsLimit <= 0 {
		resultsLimit = 100
	}
	if resultsLimit > 1000 {
		resultsLimit = 1000
	}

	resultsOffset := query.ResultsOffset
	if resultsOffset < 0 {
		resultsOffset = 0
	}

	// Slice the results based on pagination
	startIdx := resultsOffset
	endIdx := resultsOffset + resultsLimit

	if startIdx >= len(allResults) {
		exec.Results = []interfaces.ObjectExecutionResult{}
	} else {
		if endIdx > len(allResults) {
			endIdx = len(allResults)
		}
		exec.Results = allResults[startIdx:endIdx]
	}

	// Set pagination metadata
	exec.ResultsTotal = resultsTotal
	exec.ResultsOffset = resultsOffset
	exec.ResultsLimit = resultsLimit

	return exec, nil
}

// QueryExecutions queries executions based on filter criteria
func (s *actionLogsService) QueryExecutions(ctx context.Context, query *interfaces.ActionLogQuery) (*interfaces.ActionExecutionList, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "QueryExecutions", trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()

	span.SetAttributes(attr.Key("kn_id").String(query.KNID))

	indexName := interfaces.GetActionExecutionIndex(query.KNID)

	// Build the must conditions
	mustConditions := []map[string]any{}

	if query.ActionTypeID != "" {
		mustConditions = append(mustConditions, map[string]any{
			"term": map[string]any{
				"action_type_id": query.ActionTypeID,
			},
		})
	}

	if query.Status != "" {
		mustConditions = append(mustConditions, map[string]any{
			"term": map[string]any{
				"status": query.Status,
			},
		})
	}

	if query.TriggerType != "" {
		mustConditions = append(mustConditions, map[string]any{
			"term": map[string]any{
				"trigger_type": query.TriggerType,
			},
		})
	}

	if len(query.StartTimeRange) == 2 {
		mustConditions = append(mustConditions, map[string]any{
			"range": map[string]any{
				"start_time": map[string]any{
					"gte": query.StartTimeRange[0],
					"lte": query.StartTimeRange[1],
				},
			},
		})
	}

	// Build the query
	offset := query.Offset
	if offset < 0 {
		offset = 0
	}

	limit := query.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 1000 {
		limit = 1000
	}

	osQuery := map[string]any{
		"query": map[string]any{
			"bool": map[string]any{
				"must": mustConditions,
			},
		},
		"from": offset,
		"size": limit,
		"sort": []map[string]any{
			{"start_time": map[string]any{"order": "desc"}},
			{"id": map[string]any{"order": "asc"}},
		},
	}

	if len(query.SearchAfter) > 0 {
		osQuery["search_after"] = query.SearchAfter
	}

	hits, err := s.osAccess.SearchData(ctx, indexName, osQuery)
	if err != nil {
		logger.Errorf("Failed to query executions: %v", err)
		return nil, fmt.Errorf("failed to query executions: %w", err)
	}

	// Convert hits to executions
	executions := make([]interfaces.ActionExecution, 0, len(hits))
	var lastSort []any

	for _, hit := range hits {
		exec, err := mapToActionExecution(hit.Source)
		if err != nil {
			logger.Warnf("Failed to parse execution, skipping: %v", err)
			continue
		}
		executions = append(executions, *exec)
		lastSort = hit.Sort
	}

	result := &interfaces.ActionExecutionList{
		Entries:     executions,
		SearchAfter: lastSort,
	}

	// Get total count if needed
	if query.NeedTotal {
		countQuery := map[string]any{
			"query": map[string]any{
				"bool": map[string]any{
					"must": mustConditions,
				},
			},
		}
		countBytes, err := s.osAccess.Count(ctx, indexName, countQuery)
		if err == nil {
			var countResult struct {
				Count int `json:"count"`
			}
			if json.Unmarshal(countBytes, &countResult) == nil {
				result.TotalCount = countResult.Count
			}
		}
	}

	return result, nil
}

// ensureIndexExists creates the index if it doesn't exist
// This function is safe for concurrent calls - if multiple requests try to create
// the same index simultaneously, only one will succeed and others will detect the
// index already exists.
func (s *actionLogsService) ensureIndexExists(ctx context.Context, indexName string) error {
	exists, err := s.osAccess.IndexExists(ctx, indexName)
	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	// Create the index with mappings
	indexBody := map[string]any{
		"settings": map[string]any{
			"number_of_shards":   1,
			"number_of_replicas": 0,
		},
		"mappings": map[string]any{
			"properties": map[string]any{
				"id":                 map[string]any{"type": "keyword"},
				"kn_id":              map[string]any{"type": "keyword"},
				"action_type_id":     map[string]any{"type": "keyword"},
				"action_type_name":   map[string]any{"type": "keyword"},
				"action_source_type": map[string]any{"type": "keyword"},
				"object_type_id":     map[string]any{"type": "keyword"},
				"trigger_type":       map[string]any{"type": "keyword"},
				"status":             map[string]any{"type": "keyword"},
				"total_count":        map[string]any{"type": "integer"},
				"success_count":      map[string]any{"type": "integer"},
				"failed_count":       map[string]any{"type": "integer"},
				"executor_id":        map[string]any{"type": "keyword"},
				"executor": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"id":   map[string]any{"type": "keyword"},
						"type": map[string]any{"type": "keyword"},
						"name": map[string]any{"type": "keyword"},
					},
				},
				"start_time":           map[string]any{"type": "long"},
				"end_time":             map[string]any{"type": "long"},
				"duration_ms":          map[string]any{"type": "long"},
				"results":              map[string]any{"type": "nested"},
				"dynamic_params":       map[string]any{"type": "object", "enabled": false},
				"action_source":        map[string]any{"type": "object", "enabled": false},
				"action_type_snapshot": map[string]any{"type": "object", "enabled": false},
			},
		},
	}

	if err := s.osAccess.CreateIndex(ctx, indexName, indexBody); err != nil {
		// Handle concurrent creation: if creation fails, check if another request created it
		existsAfter, checkErr := s.osAccess.IndexExists(ctx, indexName)
		if checkErr == nil && existsAfter {
			logger.Debugf("Index %s was created by another request", indexName)
			return nil
		}
		return fmt.Errorf("failed to create index: %w", err)
	}

	logger.Infof("Created index: %s", indexName)

	// Wait a bit for the index to be ready
	time.Sleep(100 * time.Millisecond)

	return nil
}

// CancelExecution cancels a running or pending execution
func (s *actionLogsService) CancelExecution(ctx context.Context, knID, execID, reason string) (*interfaces.CancelExecutionResponse, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "CancelExecution", trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()

	span.SetAttributes(
		attr.Key("execution_id").String(execID),
		attr.Key("kn_id").String(knID),
		attr.Key("reason").String(reason),
	)

	// Get the current execution (without pagination for full data)
	query := &interfaces.ActionLogDetailQuery{
		KNID:         knID,
		LogID:        execID,
		ResultsLimit: 10000, // Get all results
	}
	exec, err := s.GetExecution(ctx, query)
	if err != nil {
		return nil, err
	}

	// Check if execution can be cancelled
	if exec.Status == interfaces.ExecutionStatusCompleted ||
		exec.Status == interfaces.ExecutionStatusFailed ||
		exec.Status == interfaces.ExecutionStatusCancelled {
		return nil, fmt.Errorf("execution %s cannot be cancelled, current status: %s", execID, exec.Status)
	}

	// Count and update pending objects
	cancelledCount := 0
	completedCount := 0
	for i := range exec.Results {
		if exec.Results[i].Status == interfaces.ObjectStatusPending {
			exec.Results[i].Status = interfaces.ObjectStatusCancelled
			exec.Results[i].ErrorMessage = "cancelled by user"
			if reason != "" {
				exec.Results[i].ErrorMessage = fmt.Sprintf("cancelled: %s", reason)
			}
			cancelledCount++
		} else if exec.Results[i].Status == interfaces.ObjectStatusSuccess {
			completedCount++
		}
	}

	// Update execution status
	exec.Status = interfaces.ExecutionStatusCancelled
	exec.EndTime = time.Now().UnixMilli()
	if exec.StartTime > 0 {
		exec.DurationMs = exec.EndTime - exec.StartTime
	}

	// Save the updated execution
	indexName := interfaces.GetActionExecutionIndex(knID)
	execMap := structToMap(exec)
	if err := s.osAccess.InsertData(ctx, indexName, execID, execMap); err != nil {
		logger.Errorf("Failed to update cancelled execution: %v", err)
		return nil, fmt.Errorf("failed to update cancelled execution: %w", err)
	}

	logger.Infof("Cancelled execution %s, cancelled_count=%d, completed_count=%d", execID, cancelledCount, completedCount)

	return &interfaces.CancelExecutionResponse{
		ExecutionID:    execID,
		Status:         interfaces.ExecutionStatusCancelled,
		Message:        "任务已取消",
		CancelledCount: cancelledCount,
		CompletedCount: completedCount,
	}, nil
}

// structToMap converts a struct to a map
func structToMap(v any) map[string]any {
	data, _ := json.Marshal(v)
	var result map[string]any
	_ = json.Unmarshal(data, &result)
	return result
}

// mapToActionExecution converts a map to ActionExecution
func mapToActionExecution(m map[string]any) (*interfaces.ActionExecution, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	var exec interfaces.ActionExecution
	if err := json.Unmarshal(data, &exec); err != nil {
		return nil, err
	}

	return &exec, nil
}

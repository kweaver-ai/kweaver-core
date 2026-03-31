// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package action_scheduler

import (
	"context"
	"fmt"
	"strings"

	"github.com/kweaver-ai/kweaver-go-lib/logger"

	"ontology-query/interfaces"
)

// ExecuteTool executes a tool-based action through tool-box API
// API: POST /tool-box/{box_id}/proxy/{tool_id}
func ExecuteTool(ctx context.Context, aoAccess interfaces.AgentOperatorAccess, actionType *interfaces.ActionType, params map[string]any) (any, error) {
	source := actionType.ActionSource

	// Validate tool configuration
	if source.BoxID == "" || source.ToolID == "" {
		return nil, fmt.Errorf("tool execution requires box_id and tool_id")
	}

	// Build tool execution request using ActionType.Parameters configuration
	execRequest := buildToolExecutionRequest(actionType.Parameters, params)
	execRequest.Timeout = 300 // 5 minutes timeout

	logger.Debugf("Executing tool: box_id=%s, tool_id=%s, request=%+v", source.BoxID, source.ToolID, execRequest)

	// Execute through tool-box API
	result, err := aoAccess.ExecuteTool(ctx, source.BoxID, source.ToolID, execRequest)
	if err != nil {
		logger.Errorf("Tool execution failed: %v", err)
		return nil, fmt.Errorf("tool execution failed: %w", err)
	}

	logger.Debugf("Tool execution completed successfully")
	return result, nil
}

// buildToolExecutionRequest builds ToolExecutionRequest based on ActionType.Parameters configuration
// Note: params already contains processed values from buildExecutionParams
func buildToolExecutionRequest(configParams []interfaces.Parameter, params map[string]any) interfaces.ToolExecutionRequest {
	request := interfaces.ToolExecutionRequest{
		Header: map[string]any{},
		Query:  map[string]any{},
		Body:   map[string]any{},
		Path:   map[string]any{},
	}

	// If no parameters configured, put all params in body (backward compatible)
	if len(configParams) == 0 {
		request.Body = params
		return request
	}

	// Process each configured parameter - get value from params (already processed by buildExecutionParams)
	// and assign to the appropriate location based on Source.
	// Uses getNestedValue/setNestedValue to support dot-separated nested parameter names (e.g. "props.headers").
	for _, param := range configParams {
		value := getNestedValue(params, param.Name)
		if value == nil {
			continue
		}

		switch strings.ToLower(param.Source) {
		case interfaces.PARAMETER_HEADER:
			setNestedValue(request.Header, param.Name, value)
		case interfaces.PARAMETER_QUERY:
			setNestedValue(request.Query, param.Name, value)
		case interfaces.PARAMETER_BODY:
			setNestedValue(request.Body, param.Name, value)
		case interfaces.PARAMETER_PATH:
			setNestedValue(request.Path, param.Name, value)
		default:
			setNestedValue(request.Body, param.Name, value)
		}
	}

	return request
}

// getNestedValue retrieves a value from a map using a dot-separated key for nested access.
func getNestedValue(data map[string]any, key string) any {
	if data == nil {
		return nil
	}

	if strings.Contains(key, ".") {
		parts := strings.Split(key, ".")
		current := data

		for i, part := range parts {
			if i == len(parts)-1 {
				return current[part]
			}
			if next, ok := current[part].(map[string]any); ok {
				current = next
			} else {
				return nil
			}
		}
	}

	return data[key]
}

// setNestedValue sets a value in a map using a dot-separated key, creating intermediate maps as needed.
func setNestedValue(target map[string]any, key string, value any) {
	if value == nil {
		return
	}

	if strings.Contains(key, ".") {
		parts := strings.Split(key, ".")
		current := target

		for i, part := range parts {
			if i == len(parts)-1 {
				current[part] = value
				return
			}
			if _, exists := current[part]; !exists {
				current[part] = make(map[string]any)
			}
			if next, ok := current[part].(map[string]any); ok {
				current = next
			} else {
				current[part] = make(map[string]any)
				current = current[part].(map[string]any)
			}
		}
		return
	}

	target[key] = value
}

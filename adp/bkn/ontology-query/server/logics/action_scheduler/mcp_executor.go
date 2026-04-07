// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package action_scheduler

import (
	"context"
	"fmt"

	"github.com/kweaver-ai/kweaver-go-lib/logger"

	"ontology-query/interfaces"
)

// ExecuteMCP executes an MCP-based action through agent-operator-integration
// API: POST /mcp/proxy/{mcp_id}/tool/call
func ExecuteMCP(ctx context.Context, aoAccess interfaces.AgentOperatorAccess, actionType *interfaces.ActionType, params map[string]any) (any, error) {
	source := actionType.ActionSource

	// Validate MCP configuration
	if source.McpID == "" {
		return nil, fmt.Errorf("MCP execution requires mcp_id")
	}

	toolName := source.ToolName
	if toolName == "" {
		toolName = source.ToolID
	}

	// Build MCP execution request using ActionType.Parameters configuration
	mcpParams := buildMCPParameters(actionType.Parameters, params)

	mcpRequest := interfaces.MCPExecutionRequest{
		McpID:      source.McpID,
		ToolName:   toolName,
		Parameters: mcpParams,
		Timeout:    60, // Default 60 seconds timeout
	}

	mcpID := source.McpID

	logger.Debugf("Executing MCP: mcp_id=%s, tool_name=%s, params=%+v", mcpID, toolName, mcpParams)

	// Execute through agent-operator-integration MCP endpoint
	result, err := aoAccess.ExecuteMCP(ctx, mcpID, toolName, mcpRequest)
	if err != nil {
		logger.Errorf("MCP execution failed: %v", err)
		return nil, fmt.Errorf("MCP execution failed: %w", err)
	}

	logger.Debugf("MCP execution completed successfully")
	return result, nil
}

// buildMCPParameters builds parameters for MCP execution based on ActionType.Parameters configuration
// Note: params already contains processed values from buildExecutionParams
func buildMCPParameters(configParams []interfaces.Parameter, params map[string]any) map[string]any {
	// If no parameters configured, use params directly (backward compatible)
	if len(configParams) == 0 {
		return params
	}

	// For MCP, params already contains the processed values from buildExecutionParams
	// Just return params directly since MCP doesn't distinguish between header/body/query/path
	return params
}

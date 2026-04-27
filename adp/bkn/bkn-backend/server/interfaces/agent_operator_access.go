// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import "context"

// AgentOperator describes an operator returned from agent-operator-integration market API.
type AgentOperator struct {
	OperatorId string `json:"operator_id"`
	Name       string `json:"name"`
}

//go:generate mockgen -source ../interfaces/agent_operator_access.go -destination ../interfaces/mock/mock_agent_operator_access.go -package mock_interfaces
type AgentOperatorAccess interface {
	GetAgentOperatorByID(ctx context.Context, operatorID string) (AgentOperator, error)
	// GetToolByID verifies the tool exists in the tool-box via internal GET .../tool-box/{box_id}/tool/{tool_id}.
	GetToolByID(ctx context.Context, boxID, toolID string) error
	// GetMcpToolByName verifies the MCP server exposes a tool with the given name (internal GET .../mcp/proxy/{mcp_id}/tools).
	GetMcpToolByName(ctx context.Context, mcpID, toolName string) error
}

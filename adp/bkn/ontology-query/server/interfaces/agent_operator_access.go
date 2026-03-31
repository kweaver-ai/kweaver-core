// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import "context"

const (
	PARAMETER_HEADER = "header"
	PARAMETER_BODY   = "body"
	PARAMETER_QUERY  = "query"
	PARAMETER_PATH   = "path"
)

type AgentOperator struct {
	Name       string `json:"name"`
	OperatorId string `json:"operator_id"`
	Version    string `json:"version"`
	Status     string `json:"status"`
	// MetadataType string       `json:"metadata_type"`
	// Metadata     Metadata     `json:"metadata"`
	OperatorInfo OperatorInfo `json:"operator_info"`
}

// type Metadata struct {
// 	Openapi Openapi `json:"openapi"`
// }

// type Openapi struct {
// 	Summary     string   `json:"summary"`
// 	Path        string   `json:"path"`
// 	Method      string   `json:"method"`
// 	Description string   `json:"description"`
// 	Server_url  []string `json:"server_url"`
// 	ApiSpec     ApiSpec  `json:"api_spec"`
// }

// type ApiSpec struct {
// 	Parameters OperatorParameters `json:"parameters"`
// 	// RequestBody OperatorParameters `json:"request_body"`
// 	// RequestBody OperatorParameters `json:"request_body"`
// }

// type OperatorParameters struct {
// 	Name            string          `json:"name"`
// 	In              string          `json:"in"`
// 	Description     string          `json:"description"`
// 	Required        bool            `json:"required"`
// 	ParameterSchema ParameterSchema `json:"schema"`
// }

// type ParameterSchema struct {
// 	Type   string `json:"type"`
// 	Format string `json:"format"`
// }

type OperatorInfo struct {
	OperatorType  string `json:"operator_type"`
	ExecutionMode string `json:"execution_mode"`
}

// 算子执行请求体
type OperatorExecutionRequest struct {
	Header  map[string]any `json:"header"`
	Body    map[string]any `json:"body"`
	Query   map[string]any `json:"query"`
	Path    map[string]any `json:"path"`
	Timeout int64          `json:"timeout"` // 超时时间，单位秒
}

// ToolExecutionRequest represents the request to execute a tool via tool-box API
type ToolExecutionRequest struct {
	Header  map[string]any `json:"header,omitempty"`
	Body    map[string]any `json:"body,omitempty"`
	Query   map[string]any `json:"query,omitempty"`
	Path    map[string]any `json:"path,omitempty"`
	Timeout int64          `json:"timeout,omitempty"` // 超时时间，单位秒
}

//go:generate mockgen -source ../interfaces/agent_operator_access.go -destination ../interfaces/mock/mock_agent_operator_access.go
type AgentOperatorAccess interface {
	GetAgentOperatorByID(ctx context.Context, operatorID string) (AgentOperator, error)
	ExecuteOperator(ctx context.Context, operatorID string, execRequest OperatorExecutionRequest) (any, error)
	// ExecuteTool executes a tool via tool-box API
	// API: POST /tool-box/{box_id}/proxy/{tool_id}
	ExecuteTool(ctx context.Context, boxID string, toolID string, execRequest ToolExecutionRequest) (any, error)
	// ExecuteMCP executes an MCP-based action through agent-operator-integration
	// API: POST /mcp/proxy/{mcp_id}/tool/call
	ExecuteMCP(ctx context.Context, mcpID string, toolName string, execRequest MCPExecutionRequest) (any, error)
}

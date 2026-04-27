// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

// Action execution status constants
const (
	ExecutionStatusPending   = "pending"
	ExecutionStatusRunning   = "running"
	ExecutionStatusCompleted = "completed"
	ExecutionStatusFailed    = "failed"
	ExecutionStatusCancelled = "cancelled"
)

// Object execution status constants
const (
	ObjectStatusPending   = "pending"
	ObjectStatusSuccess   = "success"
	ObjectStatusFailed    = "failed"
	ObjectStatusCancelled = "cancelled"
)

// Trigger type constants
const (
	TriggerTypeManual    = "manual"
	TriggerTypeScheduled = "scheduled"
)

// Action source type constants
const (
	ActionSourceTypeTool = "tool"
	ActionSourceTypeMCP  = "mcp"
)

// ActionExecutionRequest represents the request to execute an action
type ActionExecutionRequest struct {
	KNID               string           `json:"-"`
	Branch             string           `json:"-"`
	ActionTypeID       string           `json:"-"`
	TriggerType        string           `json:"trigger_type,omitempty"` // "manual" or "scheduled", defaults to "manual"
	InstanceIdentities []map[string]any `json:"_instance_identities"`
	DynamicParams      map[string]any   `json:"dynamic_params,omitempty"`

	Instances []ObjectSystemInfo `json:"-"`
	ObjDatas  []map[string]any   `json:"-"`
}

// ActionExecutionResponse represents the immediate response after submitting execution
type ActionExecutionResponse struct {
	ExecutionID string `json:"execution_id"`
	Status      string `json:"status"`
	Message     string `json:"message"`
	CreatedAt   int64  `json:"created_at"`
}

// ActionExecution represents a single execution request (may contain multiple objects)
type ActionExecution struct {
	ID                 string                  `json:"id"` // execution_id
	KNID               string                  `json:"kn_id"`
	ActionTypeID       string                  `json:"action_type_id"`
	ActionTypeName     string                  `json:"action_type_name"`
	ActionSourceType   string                  `json:"action_source_type"` // "tool" | "mcp"
	ActionSource       ActionSource            `json:"action_source"`
	ObjectTypeID       string                  `json:"object_type_id"`
	TriggerType        string                  `json:"trigger_type"` // "manual" | "scheduled"
	Status             string                  `json:"status"`       // "pending" | "running" | "completed" | "failed"
	TotalCount         int                     `json:"total_count"`
	SuccessCount       int                     `json:"success_count"`
	FailedCount        int                     `json:"failed_count"`
	Results            []ObjectExecutionResult `json:"results"`
	ResultsTotal       int                     `json:"results_total,omitempty"`  // total count of results (for pagination)
	ResultsOffset      int                     `json:"results_offset,omitempty"` // current offset of results
	ResultsLimit       int                     `json:"results_limit,omitempty"`  // current limit of results
	DynamicParams      map[string]any          `json:"dynamic_params,omitempty"`
	ExecutorID         string                  `json:"executor_id"`                    // user ID who triggered (deprecated, use Executor instead)
	Executor           AccountInfo             `json:"executor"`                       // user info who triggered the execution
	StartTime          int64                   `json:"start_time"`                     // execution start time (Unix milliseconds)
	EndTime            int64                   `json:"end_time,omitempty"`             // execution end time (Unix milliseconds)
	DurationMs         int64                   `json:"duration_ms,omitempty"`          // execution duration in milliseconds
	ActionTypeSnapshot map[string]any          `json:"action_type_snapshot,omitempty"` // 执行时的行动类配置快照（与 manager 返回一致）
}

// ObjectExecutionResult represents execution result for a single object
type ObjectExecutionResult struct {
	ObjectSystemInfo
	Status       string         `json:"status"` // "pending" | "success" | "failed"
	Parameters   map[string]any `json:"parameters,omitempty"`
	Result       any            `json:"result,omitempty"`
	ErrorMessage string         `json:"error_message,omitempty"`
	StartTime    int64          `json:"start_time,omitempty"`
	EndTime      int64          `json:"end_time,omitempty"`
	DurationMs   int64          `json:"duration_ms,omitempty"`
}

// ActionLogQuery represents query parameters for execution logs (supports both GET query params and JSON body)
type ActionLogQuery struct {
	KNID           string  `json:"-" form:"-"`
	ActionTypeID   string  `json:"action_type_id,omitempty" form:"action_type_id"`
	Status         string  `json:"status,omitempty" form:"status"`
	TriggerType    string  `json:"trigger_type,omitempty" form:"trigger_type"`
	StartTimeRange []int64 `json:"start_time_range,omitempty"` // [start, end] for JSON body
	StartTimeFrom  int64   `json:"-" form:"start_time_from"`   // for GET query params
	StartTimeTo    int64   `json:"-" form:"start_time_to"`     // for GET query params
	Offset         int     `json:"offset,omitempty" form:"offset"`
	Limit          int     `json:"limit,omitempty" form:"limit"`
	NeedTotal      bool    `json:"need_total,omitempty" form:"need_total"`
	SearchAfter    []any   `json:"search_after,omitempty"`
	SearchAfterStr string  `json:"-" form:"search_after"` // comma-separated string for GET query params
}

// ActionLogDetailQuery represents query parameters for single execution log detail
type ActionLogDetailQuery struct {
	KNID          string `form:"-"`
	LogID         string `form:"-"`
	ResultsLimit  int    `form:"results_limit"`  // pagination limit for results, default 100, max 1000
	ResultsOffset int    `form:"results_offset"` // pagination offset for results, default 0
	ResultsStatus string `form:"results_status"` // filter results by status: "success" | "failed"
}

// ActionExecutionList represents a list of action executions with pagination
type ActionExecutionList struct {
	Entries     []ActionExecution `json:"entries"`
	TotalCount  int               `json:"total_count,omitempty"`
	SearchAfter []any             `json:"search_after,omitempty"`
}

// MCPExecutionRequest represents the request to execute an MCP action
type MCPExecutionRequest struct {
	McpID      string         `json:"mcp_id"`
	ToolName   string         `json:"tool_name"`
	Parameters map[string]any `json:"parameters"`
	Timeout    int64          `json:"timeout"` // timeout in seconds
}

// CancelExecutionRequest represents the request to cancel an execution
type CancelExecutionRequest struct {
	Reason string `json:"reason,omitempty"`
}

// CancelExecutionResponse represents the response after cancelling an execution
type CancelExecutionResponse struct {
	ExecutionID    string `json:"execution_id"`
	Status         string `json:"status"`
	Message        string `json:"message"`
	CancelledCount int    `json:"cancelled_count"` // number of objects that were pending and now cancelled
	CompletedCount int    `json:"completed_count"` // number of objects that were already completed before cancel
}

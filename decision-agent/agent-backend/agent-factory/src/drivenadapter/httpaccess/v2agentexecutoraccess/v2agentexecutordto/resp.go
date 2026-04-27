package v2agentexecutordto

// V2AgentCallResp v2 版本的 Agent 调用响应
type V2AgentCallResp struct {
	Answer        interface{}        `json:"answer"`                   // Agent 输出
	Status        string             `json:"status"`                   // 状态
	AgentRunID    string             `json:"agent_run_id,omitempty"`   // Agent 运行 ID
	InterruptInfo *ToolInterruptInfo `json:"interrupt_info,omitempty"` // 中断信息
}

// V2AgentDebugResp v2 版本的 Agent Debug 响应
type V2AgentDebugResp struct {
	Answer        interface{}        `json:"answer"`                   // Agent 输出
	Status        string             `json:"status"`                   // 状态
	AgentRunID    string             `json:"agent_run_id,omitempty"`   // Agent 运行 ID
	InterruptInfo *ToolInterruptInfo `json:"interrupt_info,omitempty"` // 中断信息
}

package v2agentexecutordto

// AgentTerminateReq Agent 终止执行请求
type AgentTerminateReq struct {
	AgentRunID string `json:"agent_run_id"` // Agent运行ID
}

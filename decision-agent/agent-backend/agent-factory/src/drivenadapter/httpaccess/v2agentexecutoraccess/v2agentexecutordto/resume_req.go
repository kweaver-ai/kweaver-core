package v2agentexecutordto

// AgentResumeReq Agent 恢复执行请求
type AgentResumeReq struct {
	AgentRunID string           `json:"agent_run_id"` // Agent运行ID
	ResumeInfo *AgentResumeInfo `json:"resume_info"`  // 恢复执行信息
}

// AgentResumeInfo 恢复执行信息
type AgentResumeInfo struct {
	ResumeHandle *InterruptHandle `json:"resume_handle"` // 复用 InterruptHandle
	Action       string           `json:"action"`        // 操作类型: confirm | skip
	ModifiedArgs []ModifiedArg    `json:"modified_args"` // 修改后的参数
	Data         *InterruptData   `json:"data"`          // 中断详情数据（从响应透传）
}

// ModifiedArg 修改后的参数
type ModifiedArg struct {
	Key   string      `json:"key"`   // 参数名称
	Value interface{} `json:"value"` // 参数值
}

package v2agentexecutordto

// ToolInterruptInfo 工具中断信息
type ToolInterruptInfo struct {
	Handle *InterruptHandle `json:"handle"` // 恢复句柄
	Data   *InterruptData   `json:"data"`   // 中断详情
}

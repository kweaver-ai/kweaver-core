package v2agentexecutordto

// InterruptData 中断详情
type InterruptData struct {
	ToolName        string           `json:"tool_name"`        // 工具名称
	ToolDescription string           `json:"tool_description"` // 工具描述
	ToolArgs        []ToolArg        `json:"tool_args"`        // 工具参数列表
	InterruptConfig *InterruptConfig `json:"interrupt_config"` // 中断配置
}

// ToolArg 工具参数
type ToolArg struct {
	Key   string      `json:"key"`   // 参数名称
	Value interface{} `json:"value"` // 参数值
	Type  string      `json:"type"`  // 参数类型
}

// InterruptConfig 中断配置
type InterruptConfig struct {
	RequiresConfirmation bool   `json:"requires_confirmation"` // 是否需要用户确认
	ConfirmationMessage  string `json:"confirmation_message"`  // 确认提示消息
}

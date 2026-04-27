package agentexecutordto

// NOTE： agent-executor 返回给 agent-app 的响应结果
type AgentCallResp struct {
	Answer interface{} `json:"answer"`          // dolphin执行输出的结果。返回结果与dolphin语句的内容有关
	Status string      `json:"status"`          // 执行状态("True": 流式信息已结束, "False": 流式信息未结束, "Error": 执行错误)
	Ask    interface{} `json:"ask,omitempty"`   // 当出现中断时返回该字段
	Error  interface{} `json:"error,omitempty"` // 发生错误时返回该字段。包含错误码、错误信息等。
	// UserDefine map[string]interface{} `json:"user_define,omitempty"` // 透传中断的messageid
}

package agentreq

// TerminateReq 终止聊天请求
type TerminateReq struct {
	ConversationID                string `json:"conversation_id" binding:"required"` // 会话ID
	AgentRunID                    string `json:"agent_run_id"`                       // Agent运行ID（用于终止 Executor 中的 Agent）
	InterruptedAssistantMessageID string `json:"interrupted_assistant_message_id"`   // 中断的助手消息ID
}

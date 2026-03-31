package agentreq

// ResumeReq 恢复聊天请求（Session恢复）
type ResumeReq struct {
	ConversationID string `json:"conversation_id" binding:"required"` // 会话ID
}

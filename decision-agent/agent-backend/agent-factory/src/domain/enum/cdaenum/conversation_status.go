package cdaenum

// ConversationStatus 会话状态（基于最新消息状态的简化映射）
type ConversationStatus string

const (
	// 处理中
	ConvStatusProcessing ConversationStatus = "processing"
	// 已完成
	ConvStatusCompleted ConversationStatus = "completed"
	// 已取消
	ConvStatusCancelled ConversationStatus = "cancelled"
	// 失败
	ConvStatusFailed ConversationStatus = "failed"
)

package observabilityresp

import "github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/agentrespvo"

// POST /api/agent-app/v1/observability/agent/:agent_id/conversation/:conversation_id/session/:session_id/run/:run_id/detail
// 获取指定run的详情信息
// 返回信息是单个run的完整详情
type RunDetailResp struct {
	RunID string `json:"run_id"`

	AgentID             string                 `json:"agent_id"`
	AgentVersion        string                 `json:"agent_version"`
	ConversationID      string                 `json:"conversation_id"` // 对话id
	SessionID           string                 `json:"session_id"`      // 会话id
	UserID              string                 `json:"user_id"`
	CallType            string                 `json:"call_type"`  // 调用类型 chat/debugchat/apichat
	StartTime           int64                  `json:"start_time"` // 毫秒时间戳
	EndTime             int64                  `json:"end_time"`
	TTFT                int                    `json:"ttft"`                   // 单位ms 首token响应时间
	TotalTime           int64                  `json:"total_time"`             // 单位ms 运行时长
	TotalTokens         int64                  `json:"total_tokens"`           // 总token数
	InputMessage        string                 `json:"input_message"`          // 输入query
	ToolCallCount       int                    `json:"tool_call_count"`        // 工具调用次数
	ToolCallFailedCount int                    `json:"tool_call_failed_count"` // 工具调用失败次数
	Progress            []agentrespvo.Progress `json:"progress"`               // run运行过程
	Status              string                 `json:"status"`                 // 运行状态 Success / Failed
}

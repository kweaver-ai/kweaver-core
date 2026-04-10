package conversationmsgvo

import (
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/agentresperr"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/httpaccess/v2agentexecutoraccess/v2agentexecutordto"
)

// MessageExt 消息扩展字段
type MessageExt struct {
	InterruptInfo  *v2agentexecutordto.ToolInterruptInfo `json:"interrupt_info,omitempty"`  // 中断信息
	RelatedQueries []string                              `json:"related_queries,omitempty"` // 相关问题（query 列表）
	TotalTime      float64                               `json:"total_time,omitempty"`      // 总耗时（秒）
	TotalTokens    int64                                 `json:"total_tokens,omitempty"`    // 总 token 数
	TTFT           int64                                 `json:"ttft,omitempty"`            // 首 token 时间（毫秒）
	AgentRunID     string                                `json:"agent_run_id,omitempty"`    // Agent 运行 ID
	Error          *agentresperr.RespError               `json:"error,omitempty"`           // 错误信息
}

func (m *MessageExt) IsInterrupted() bool {
	return m.InterruptInfo != nil
}

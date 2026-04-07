package observabilityreq

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
)

// POST /api/agent-app/v1/observability/agent/:agent_id/conversation/:conversation_id/session/:session_id/run
// 获取指定对话的session下的run列表
// 返回信息是每个run的详情列表
type RunListReq struct {
	AgentID        string `json:"agent_id"`
	AgentVersion   string `json:"agent_version"`
	ConversationID string `json:"conversation_id"`
	SessionID      string `json:"session_id"`
	StartTime      int64  `json:"start_time"`
	EndTime        int64  `json:"end_time"`
	Page           int    `json:"page"` // 默认值为1
	Size           int    `json:"size"` // 默认值为10

	XAccountID   string            `json:"-"`
	XAccountType cenum.AccountType `json:"-"`
}

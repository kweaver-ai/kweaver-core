package observabilityreq

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
)

// POST /api/agent-app/v1/observability/agent/:agent_id/conversation/:conversation_id/session
// 获取指定对话的session列表
// 返回值： 包括session列表，每个session 的大概信息还是详情呢，如果是详情直接就是把run信息也返回了
type SessionListReq struct {
	AgentID        string `json:"agent_id"`
	AgentVersion   string `json:"agent_version"`
	ConversationID string `json:"conversation_id"`
	StartTime      int64  `json:"start_time"`
	EndTime        int64  `json:"end_time"`
	Size           int    `json:"size"` // 默认值为10
	Page           int    `json:"page"` // 默认值为1

	XAccountID   string            `json:"-"`
	XAccountType cenum.AccountType `json:"-"`
}

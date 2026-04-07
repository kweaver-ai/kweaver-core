package observabilityreq

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
)

// POST /api/agent-app/v1/observability/agent/:agent_id/conversation/:conversation_id/session/:session_id/run/:run_id/detail
// 获取指定run的详情信息
// 返回信息是单个run的完整详情
type RunDetailReq struct {
	AgentID        string `json:"agent_id"`
	AgentVersion   string `json:"agent_version"`
	ConversationID string `json:"conversation_id"`
	SessionID      string `json:"session_id"`
	RunID          string `json:"run_id"`
	StartTime      int64  `json:"start_time"`
	EndTime        int64  `json:"end_time"`

	XAccountID   string            `json:"-"`
	XAccountType cenum.AccountType `json:"-"`
}

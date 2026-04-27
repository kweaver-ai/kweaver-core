package conversationresp

import "github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cenum"

type InitConversationResp struct {
	ID               string            `json:"id"`
	SandboxSessionID string            `json:"sandbox_session_id"` // 沙箱会话ID
	XAccountID       string            `json:"-"`                  // 用户ID
	XAccountType     cenum.AccountType `json:"-"`                  // 用户类型 app/user/anonymous
}

package agentreq

import (
	"errors"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
)

type ConversationSessionInitReq struct {
	ConversationID        string `json:"conversation_id"`         // 对话ID
	ConversationSessionID string `json:"conversation_session_id"` // 对话session ID。说明：当对话session ID为空时，表示是需要创建新的对话session。如果对话session ID不为空，则表示是获取对话session的信息或延长对话session的有效期。

	AgentID      string `json:"agent_id"`      // agentID
	AgentVersion string `json:"agent_version"` // agentVersion

	UserID            string            `json:"-"`
	XAccountID        string            `json:"-"` // 用户ID
	XAccountType      cenum.AccountType `json:"-"` // 用户类型 app/user/anonymous
	XBusinessDomainID string            `json:"-"`
}

func (req *ConversationSessionInitReq) Check() (err error) {
	if req.ConversationID == "" {
		err = errors.New("conversation_id is empty")
		return
	}

	// 当agent_id提供时,此字段为required
	if req.AgentID != "" {
		if req.AgentVersion == "" {
			err = errors.New("agent_version cannot be empty when agent_id is provided")
			return
		}
	}

	return
}

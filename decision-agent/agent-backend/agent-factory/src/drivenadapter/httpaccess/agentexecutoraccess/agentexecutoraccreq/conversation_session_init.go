package agentexecutoraccreq

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/daconfvalobj"
)

type ConversationSessionInitReq struct {
	ConversationID string              `json:"conversation_id"`
	AgentID        string              `json:"agent_id"`
	AgentVersion   string              `json:"agent_version"`
	AgentConfig    daconfvalobj.Config `json:"agent_config"`
}

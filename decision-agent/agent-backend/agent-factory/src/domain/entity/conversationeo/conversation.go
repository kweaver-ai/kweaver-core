package conversationeo

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
)

// DataAgent 数据智能体配置实体对象
type Conversation struct {
	*dapo.ConversationPO

	Messages []*dapo.ConversationMsgPO
}

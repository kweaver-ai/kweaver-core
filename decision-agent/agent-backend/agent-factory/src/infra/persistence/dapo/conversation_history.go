package dapo

type ConversationHistoryLatestVisitAgentPO struct {
	AgentId string `db:"f_bot_id"`
	// AgentVersion string `db:"f_bot_version"`
	LastModifyAt int64 `db:"last_modified_at"`
}

func (po *ConversationHistoryLatestVisitAgentPO) TableName() string {
	return "tb_conversation_history_v2"
}

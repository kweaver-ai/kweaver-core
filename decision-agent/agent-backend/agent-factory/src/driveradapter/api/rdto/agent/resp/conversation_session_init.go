package agentresp

type ConversationSessionInitResp struct {
	ConversationSessionID string `json:"conversation_session_id"`
	TTL                   int    `json:"ttl"`
}

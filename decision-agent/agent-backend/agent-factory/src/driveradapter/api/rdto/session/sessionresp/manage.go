package sessionresp

type ManageResp struct {
	ConversationSessionID string `json:"conversation_session_id"`
	TTL                   int    `json:"ttl"`
	StartTime             int64  `json:"start_time"`
}

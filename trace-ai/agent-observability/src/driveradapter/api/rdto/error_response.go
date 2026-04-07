package rdto

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type TraceSearchByConversationResponse struct {
	ConversationID string `json:"conversation_id" example:"conv-1001"`
}

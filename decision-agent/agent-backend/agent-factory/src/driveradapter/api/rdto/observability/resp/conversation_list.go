package observabilityresp

import "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/conversation/conversationresp"

type ObservabilityConversationDetail struct {
	Conversation conversationresp.ConversationDetail `json:"conversation"`
	SessionCount int                                 `json:"session_count"`
}

type ConversationListResp struct {
	Entries    []ObservabilityConversationDetail `json:"entries"`
	TotalCount int64                             `json:"total_count"`
}

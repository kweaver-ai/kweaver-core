package observabilityreq

// POST /api/agent-app/v1/observability/agent/:agent_id/conversation
// 查询Agent下的所有对话列表，支持分页和根据对话title模糊查询

type ConversationListReq struct {
	AgentID      string `json:"agent_id"`
	AgentVersion string `json:"agent_version"`
	Title        string `json:"title"` // 默认为空查询全部。模糊查询
	Size         int    `json:"size"`  // 默认值为10
	Page         int    `json:"page"`  // 默认值为1
	StartTime    int64  `json:"start_time"`
	EndTime      int64  `json:"end_time"`
}

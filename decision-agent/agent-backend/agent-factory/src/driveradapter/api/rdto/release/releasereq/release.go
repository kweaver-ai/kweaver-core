package releasereq

type PermissionRange []PermissionRangeObject

type PermissionRangeObject struct {
	ObjectID   string `json:"obj_id"`
	ObjectType string `json:"obj_type"`
}

type HistoryListResp []HistoryListItemResp

type HistoryListItemResp struct {
	HistoryID    string `json:"history_id"`
	AgentID      string `json:"agent_id"`
	AgentVersion string `json:"agent_version"`
	AgentDesc    string `json:"agent_desc"`
	CreateTime   int64  `json:"create_time"`
}

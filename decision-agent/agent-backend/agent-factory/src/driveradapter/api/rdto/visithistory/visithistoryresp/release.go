package releaseresp

type PublishResp struct {
	ReleaseId string `json:"release_id"`
	Version   string `json:"version"`
}

type HistoryListResp []HistoryListItemResp

type HistoryListItemResp struct {
	HistoryId    string `json:"history_id"`
	AgentId      string `json:"agent_id"`
	AgentVersion string `json:"agent_version"`
	AgentDesc    string `json:"agent_desc"`
	CreateTime   int64  `json:"create_time"`
}

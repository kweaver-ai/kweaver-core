package observabilityresp

type SessionListItem struct {
	SessionID       string `json:"session_id"`
	StartTime       int64  `json:"start_time"`
	EndTime         int64  `json:"end_time"`
	SessionDuration int    `json:"session_duration"` // 毫秒
}

type SessionListResp struct {
	Entries    []SessionListItem `json:"entries"`
	TotalCount int               `json:"total_count"` // 不分页情况下的总数量
}

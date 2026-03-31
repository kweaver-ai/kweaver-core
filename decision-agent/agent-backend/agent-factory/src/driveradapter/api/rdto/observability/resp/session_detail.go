package observabilityresp

type SessionDetailResp struct {
	SessionID             string `json:"session_id"`
	StartTime             int64  `json:"start_time"`
	EndTime               int64  `json:"end_time"`
	SessionRunCount       int    `json:"session_run_count"`
	SessionDuration       int    `json:"session_duration"`         // 毫秒
	AvgRunExecuteDuration int    `json:"avg_run_execute_duration"` // 毫秒
	AvgRunTTFTDuration    int    `json:"avg_run_ttft_duration"`    // 毫秒
	RunErrorCount         int    `json:"run_error_count"`
	ToolFailCount         int    `json:"tool_fail_count"`
}

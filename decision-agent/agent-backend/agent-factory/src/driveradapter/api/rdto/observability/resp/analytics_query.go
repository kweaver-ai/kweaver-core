package observabilityresp

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/agentrespvo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/daconfvalobj"
)

// AnalyticsQueryResp 查询可观测数据响应
// POST /analytics/query
// 根据analysis_level返回对应的数据

type AnalyticsQueryResp struct {
	Success bool   `json:"success"` // 请求是否成功
	Data    any    `json:"data"`    // 根据analysis_level返回对应的数据
	Error   string `json:"error"`   // 错误信息（如果有）
}

// AgentMetrics Agent级指标数据
type AgentMetrics struct {
	AgentConfig  daconfvalobj.Config `json:"agent_config"`  // Agent配置信息
	AgentMetrics AgentMetric         `json:"agent_metrics"` // Agent级指标数据
	SessionList  []Session           `json:"session_list"`  // 会话列表
	TrendData    TrendData           `json:"trend_data"`    // 趋势数据
}

// AgentMetric Agent级指标数据
type AgentMetric struct {
	TotalRequests      int     `json:"total_requests"`       // 总请求数
	TotalSessions      int     `json:"total_sessions"`       // 总会话数
	AvgSessionRounds   int     `json:"avg_session_rounds"`   // 平均会话轮次
	RunSuccessRate     float32 `json:"run_success_rate"`     // Run成功率
	AvgExecuteDuration int     `json:"avg_execute_duration"` // 平均执行时长
	AvgTTFTDuration    int     `json:"avg_ttft_duration"`    // 平均首token响应时间
	ToolSuccessRate    float32 `json:"tool_success_rate"`    // 工具调用成功率
}

// Session 会话信息
type Session struct {
	SessionID        string `json:"session_id"`
	SessionStartTime string `json:"session_start_time"` // ISO 8601格式
	SessionEndTime   string `json:"session_end_time"`   // ISO 8601格式
	SessionDuration  int64  `json:"session_duration"`   // 会话时长，单位毫秒
}

// TrendData 趋势数据
type TrendData struct {
	Last7Days   []any `json:"last_7_days"`   // 过去7天趋势数据
	Last24Hours []any `json:"last_24_hours"` // 过去24小时趋势数据
}

// SessionMetrics 会话级指标数据
type SessionMetrics struct {
	SessionMetrics SessionMetric       `json:"session_metrics"` // 会话指标数据
	AgentConfig    daconfvalobj.Config `json:"agent_config"`    // Agent配置信息
	RunList        []RunItem           `json:"run_list"`        // Run列表信息
}

// SessionMetric 会话指标
type SessionMetric struct {
	SessionRunCount       int `json:"session_run_count"`        // Session运行次数
	SessionDuration       int `json:"session_duration"`         // Session时长，单位毫秒
	AvgRunExecuteDuration int `json:"avg_run_execute_duration"` // 平均运行执行耗时（毫秒）
	AvgRunTTFTDuration    int `json:"avg_run_ttft_duration"`    // 平均运行首token响应耗时（毫秒）
	RunErrorCount         int `json:"run_error_count"`          // 运行错误次数
	ToolFailCount         int `json:"tool_fail_count"`          // 工具失败次数
}

// RunItem Run列表项
type RunItem struct {
	RunID        string `json:"run_id"`        // Run ID
	ResponseTime int    `json:"response_time"` // 响应时间（毫秒）
	Status       string `json:"status"`        // 状态
}

// RunMetrics Run级指标数据
type RunMetrics struct {
	RunID               string                 `json:"run_id"`                 // Run ID
	InputMessage        string                 `json:"input_message"`          // 输入query
	StartTime           int64                  `json:"start_time"`             // 开始时间，毫秒时间戳
	EndTime             int64                  `json:"end_time"`               // 结束时间
	TTFT                int                    `json:"ttft"`                   // 首token响应时间，单位ms
	TotalTime           int64                  `json:"total_time"`             // 总时间，单位ms
	TotalTokens         int64                  `json:"total_tokens"`           // 总token数
	ToolCallCount       int                    `json:"tool_call_count"`        // 工具调用次数
	ToolCallFailedCount int                    `json:"tool_call_failed_count"` // 工具调用失败次数
	Progress            []agentrespvo.Progress `json:"progress"`               // Progress链路信息
	Status              string                 `json:"status"`                 // 运行状态 Success / Failed
}

package observabilityresp

import "github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/daconfvalobj"

type AgentResp struct {
	Agent              Agent   `json:"agent"`
	TotalRequests      int     `json:"total_requests"`
	TotalSessions      int     `json:"total_sessions"`
	AvgSessionRounds   int     `json:"avg_session_rounds"`
	RunSuccessRate     float32 `json:"run_success_rate"`
	AvgExecuteDuration int     `json:"avg_execute_duration"`
	AvgTTFTDuration    int     `json:"avg_ttft_duration"`
	ToolSuccessRate    float32 `json:"tool_success_rate"`
}

type Agent struct {
	ID           string              `json:"id"`
	Version      string              `json:"version"`
	Key          string              `json:"key"`
	IsBuiltIn    int                 `json:"is_built_in"`
	Name         string              `json:"name"`
	CategoryID   string              `json:"category_id"`
	CategoryName string              `json:"category_name"`
	Profile      string              `json:"profile"`
	Config       daconfvalobj.Config `json:"config"`
	AvatarType   int                 `json:"avatar_type"`
	Avatar       string              `json:"avatar"`
	ProductID    int                 `json:"product_id"`
	ProductName  string              `json:"product_name"`
	PublishInfo  PublishInfo         `json:"publish_info"`
}

type PublishInfo struct {
	IsAPIAgent      int `json:"is_api_agent"`
	IsSDKAgent      int `json:"is_sdk_agent"`
	IsSkillAgent    int `json:"is_skill_agent"`
	IsDataFlowAgent int `json:"is_data_flow_agent"`
}

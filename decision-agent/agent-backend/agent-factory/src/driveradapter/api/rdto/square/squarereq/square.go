package squarereq

import (
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/common"
)

// Agent 详情请求对象
type AgentInfoReq struct {
	UserID       string
	AgentID      string
	AgentVersion string
	IsVisit      bool
}

// Agent 应用广场请求对象
type AgentSquareAgentReq struct {
	Name        string              `json:"name"`
	CategoryID  string              `json:"category_id"`
	ReleaseIDS  []string            `json:"release_ids"`
	PublishToBe cdaenum.PublishToBe `json:"publish_to_be"`
	common.PageSize
}

// 个人空间 Agent 请求对象
type AgentSquareMyAgentReq struct {
	UserID                    string `json:"user_id"`
	Name                      string `json:"name"`
	ShouldContainBuiltInAgent bool
	common.PageSize
}

// 最近访问 Agent 请求对象
type AgentSquareRecentAgentReq struct {
	UserID    string
	Name      string `json:"name"`
	StartTime int64  `json:"start_time"`
	EndTime   int64  `json:"end_time"`
	common.PageSize
}

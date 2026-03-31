package observabilityreq

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
)

// POST /api/agent-app/v1/observability/agent/:agent_id/detail
// 查询指定Agent的可观测信息，返回结果包含 Agent 的配置信息以及 可观测性的指标
// 支持有个时间的过滤信息

type AgentReq struct {
	AgentID       string `json:"agent_id"`
	AgentVersion  string `json:"agent_version"`
	IncludeConfig bool   `json:"include_config"` // 是否包含agent 配置信息 默认值为 false
	StartTime     int64  `json:"start_time"`     // 时间戳 单位精确到毫秒。 例如: 1646360670123
	EndTime       int64  `json:"end_time"`       // 时间戳 单位精确到毫秒。 例如: 1646360670123

	XAccountID   string            `json:"-"`
	XAccountType cenum.AccountType `json:"-"`
}

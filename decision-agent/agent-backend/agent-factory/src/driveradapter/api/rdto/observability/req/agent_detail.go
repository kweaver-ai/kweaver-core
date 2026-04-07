package observabilityreq

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
)

// POST /observability/agent/{agent_id}/detail
// 查询Agent可观测信息
// 获取Agent级别的指标数据和配置信息

type AgentDetailReq struct {
	AgentID       string `json:"agent_id"`
	AgentVersion  string `json:"agent_version"`  // Agent版本
	IncludeConfig bool   `json:"include_config"` // 是否包含配置信息
	StartTime     int64  `json:"start_time"`     // 开始时间（Unix时间戳）
	EndTime       int64  `json:"end_time"`       // 结束时间（Unix时间戳）

	XAccountID   string            `json:"-"`
	XAccountType cenum.AccountType `json:"-"`
}

package observabilityreq

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
)

// POST /analytics/query
// 查询可观测数据
// 根据分析类型和时间范围查询相应的metrics和配置数据

type AnalyticsQueryReq struct {
	AnalysisLevel string `json:"analysis_level"` // 分析类型: agent, session, run
	ID            string `json:"id"`             // 分析对象ID（Agent ID / Session ID / Run ID）
	StartTime     int64  `json:"start_time"`     // 开始时间（Unix时间戳）
	EndTime       int64  `json:"end_time"`       // 结束时间（Unix时间戳）

	XAccountID   string            `json:"-"`
	XAccountType cenum.AccountType `json:"-"`
}

package iportdriver

import (
	"context"

	observabilityreq "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/observability/req"
	observabilityresp "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/observability/resp"
)

// IObservability 可观测性服务接口
type IObservability interface {
	// AnalyticsQuery 查询可观测数据
	AnalyticsQuery(ctx context.Context, req *observabilityreq.AnalyticsQueryReq) (*observabilityresp.AnalyticsQueryResp, error)

	// AgentDetail 查询Agent可观测信息
	AgentDetail(ctx context.Context, req *observabilityreq.AgentDetailReq) (*observabilityresp.AgentResp, error)

	// SessionList 查询Session列表
	SessionList(ctx context.Context, req *observabilityreq.SessionListReq) (*observabilityresp.SessionListResp, error)

	// SessionDetail 查询Session详情
	SessionDetail(ctx context.Context, req *observabilityreq.SessionDetailReq) (*observabilityresp.SessionDetailResp, error)

	// RunList 查询Run列表
	RunList(ctx context.Context, req *observabilityreq.RunListReq) (*observabilityresp.RunListResp, error)

	// RunDetail 查询Run详情
	RunDetail(ctx context.Context, req *observabilityreq.RunDetailReq) (*observabilityresp.RunDetailResp, error)

	// GetSessionCountsByConversationIDs 批量查询会话数量
	GetSessionCountsByConversationIDs(ctx context.Context, agentID string, conversationIDs []string, startTime, endTime int64, xAccountID string, xAccountType string) (map[string]int, error)
}

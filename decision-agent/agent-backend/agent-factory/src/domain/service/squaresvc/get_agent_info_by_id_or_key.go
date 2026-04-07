package squaresvc

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/square/squarereq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/square/squareresp"
)

// GetAgentInfoByIDOrKey
func (svc *squareSvc) GetAgentInfoByIDOrKey(ctx context.Context, agentInfoReq *squarereq.AgentInfoReq) (res *squareresp.AgentMarketAgentInfoResp, err error) {
	// 1. 检查并获取agentID
	agentID, err := svc.CheckAndGetID(ctx, agentInfoReq.AgentID)
	if err != nil {
		return
	}

	// 2. 设置agentID并获取agent信息
	agentInfoReq.AgentID = agentID
	res, err = svc.GetAgentInfo(ctx, agentInfoReq)

	return
}

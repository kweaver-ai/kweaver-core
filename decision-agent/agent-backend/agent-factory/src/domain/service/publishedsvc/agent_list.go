package publishedsvc

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/p2e/publishedp2e"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/published/pubedreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/published/pubedresp"
	"github.com/pkg/errors"
)

// GetPublishedAgentList 获取已发布智能体列表
func (svc *publishedSvc) GetPublishedAgentList(ctx context.Context, req *pubedreq.PubedAgentListReq) (res *pubedresp.PubedAgentListResp, err error) {
	res, err = svc.getPublishedAgentList(ctx, req)
	if err != nil {
		return
	}

	return
}

// GetPublishedAgentList 获取已发布智能体列表
func (svc *publishedSvc) getPublishedAgentList(ctx context.Context, req *pubedreq.PubedAgentListReq) (res *pubedresp.PubedAgentListResp, err error) {
	res = pubedresp.NewPAListResp()

	// 1. 获取有使用权限的已发布智能体pos
	pos, agentID2BdIDMap, isLastPage, err := svc.getPmsAgentPos(ctx, req)
	if err != nil {
		return
	}

	if len(pos) == 0 {
		return
	}

	// 2. pos to eos

	eos, err := publishedp2e.PublishedAgents(ctx, pos, svc.umHttp, false)
	if err != nil {
		err = errors.Wrapf(err, "[publishedSvc][GetPublishedAgentList]: convert published agent list failed")
		return
	}

	// 3. 转换为响应格式

	// res.Total = totalCount

	res.IsLastPage = isLastPage

	err = res.LoadFromEos(eos, agentID2BdIDMap)
	if err != nil {
		err = errors.Wrapf(err, "[publishedSvc][GetPublishedAgentList]: load from eos failed")
		return
	}

	return
}

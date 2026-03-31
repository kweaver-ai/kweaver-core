package publishedsvc

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/p2e/publishedp2e"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/dbaccess/pubedagentdbacc/padbarg"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/published/pubedreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/published/pubedresp"
	"github.com/pkg/errors"
)

func (svc *publishedSvc) GetPubedAgentInfoList(ctx context.Context, req *pubedreq.PAInfoListReq) (res *pubedresp.PAInfoListResp, err error) {
	res = pubedresp.NewPublishedAgentInfoListResp()

	// 从数据库获取已发布智能体列表
	arg := padbarg.NewGetPaPoListByKeyArg(req.AgentKeys, nil)

	ret, err := svc.pubedAgentRepo.GetPubedListByXx(ctx, arg)
	if err != nil {
		err = errors.Wrapf(err, "[publishedSvc][GetPubedAgentInfoList]: get published agent list failed")
		return
	}

	pos := ret.JoinPos

	if len(pos) == 0 {
		return
	}
	// pos to eos

	eos, err := publishedp2e.PublishedAgents(ctx, pos, svc.umHttp, true)
	if err != nil {
		err = errors.Wrapf(err, "[publishedSvc][GetPubedAgentInfoList]: convert published agent list failed")
		return
	}

	// 转换为响应格式

	err = res.LoadFromEos(eos, req.NeedConfigFields)

	return
}

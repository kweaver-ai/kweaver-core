package v3agentconfigsvc

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/dbaccess/pubedagentdbacc/padbarg"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/dbaccess/pubedagentdbacc/padbret"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/apierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
)

// getPublishedAgentPo 获取已发布的Agent版本
func (s *dataAgentConfigSvc) getPublishedAgentPo(ctx context.Context, agentID string) (po *dapo.DataAgentPo, err error) {
	arg := padbarg.NewGetPAPoListByIDArg([]string{agentID}, nil)

	var ret *padbret.GetPaPoMapByXxRet

	ret, err = s.pubedAgentRepo.GetPubedPoMapByXx(ctx, arg)
	if err != nil {
		return
	}

	joinPo := ret.JoinPosID2PoMap[agentID]
	if joinPo == nil {
		err = capierr.NewCustom404Err(ctx, apierr.ReleaseNotFound, "此已发布的Agent不存在")
		return
	}

	po = &joinPo.DataAgentPo

	return
}

func (s *dataAgentConfigSvc) getAgentPoForCopy(ctx context.Context, agentID string) (sourcePo *dapo.DataAgentPo, err error) {
	sourcePo, err = s.agentConfRepo.GetByID(ctx, agentID)
	if err != nil {
		if chelper.IsSqlNotFound(err) {
			err = capierr.NewCustom404Err(ctx, apierr.DataAgentConfigNotFound, "源Agent不存在")
		}

		return
	}

	// if sourcePo.Status == cdaenum.StatusPublished {
	//	// 获取已发布版本
	//	sourcePo, err = s.getPublishedAgentPo(ctx, agentID)
	//	if err != nil {
	//		return
	//	}
	//}

	return
}

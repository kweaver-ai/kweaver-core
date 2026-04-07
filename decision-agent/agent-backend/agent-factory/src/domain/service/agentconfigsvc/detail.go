package v3agentconfigsvc

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/p2e/daconfp2e"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/dbaccess/pubedagentdbacc/padbarg"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/dbaccess/pubedagentdbacc/padbret"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent_config/agentconfigresp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/apierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/pkg/errors"
)

func (s *dataAgentConfigSvc) Detail(ctx context.Context, id, key string) (res *agentconfigresp.DetailRes, err error) {
	var po *dapo.DataAgentPo

	// 1. 获取数据
	if id != "" {
		po, err = s.agentConfRepo.GetByID(ctx, id)
	} else {
		po, err = s.agentConfRepo.GetByKey(ctx, key)
	}

	if err != nil {
		if chelper.IsSqlNotFound(err) {
			err = capierr.NewCustom404Err(ctx, apierr.DataAgentConfigNotFound, "数据智能体配置不存在")
			return
		}

		return
	}

	isPrivate := chelper.IsInternalAPIFromCtx(ctx)
	uid := chelper.GetUserIDFromCtx(ctx)

	// 2. 权限检查
	err = s.detailPmsCheck(ctx, po, isPrivate, uid)
	if err != nil {
		return
	}

	// 3. PO转EO
	eo, err := daconfp2e.DataAgent(ctx, po)
	if err != nil {
		return
	}

	// 4. 标记技能配置中的Agent
	if !isPrivate {
		err = s.markSkillAgentPmsForDetail(ctx, eo, uid)
		if err != nil {
			return
		}
	}

	// 5. 判断是否发布过
	isPublished := eo.Status.IsPublished()
	if !isPublished {
		// 当前状态不是已发布，需要查询是否曾经发布过
		arg := padbarg.NewGetPAPoListByIDArg([]string{po.ID}, nil)

		var pubedRet *padbret.GetPaPoMapByXxRet

		pubedRet, err = s.pubedAgentRepo.GetPubedPoMapByXx(ctx, arg)
		if err != nil {
			err = errors.Wrap(err, "[dataAgentConfigSvc][Detail]: call pubedAgentRepo.GetPubedPoMapByXx失败")
			return
		}

		_, isPublished = pubedRet.JoinPosID2PoMap[po.ID]
	}

	// 6. 转换为响应DTO
	res = agentconfigresp.NewDetailRes()

	err = res.LoadFromEo(eo)
	if err != nil {
		return
	}

	res.IsPublished = isPublished

	return
}

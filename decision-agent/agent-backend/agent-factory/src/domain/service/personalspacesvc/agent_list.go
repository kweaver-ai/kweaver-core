package personalspacesvc

import (
	"context"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/p2e/personalspacep2e"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/dbaccess/personalspacedbacc/psdbarg"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/dbaccess/pubedagentdbacc/padbarg"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/dbaccess/pubedagentdbacc/padbret"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/personal_space/personalspacereq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/personal_space/personalspaceresp"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/global"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/pkg/errors"
)

// AgentList 获取个人空间Agent列表
func (s *PersonalSpaceService) AgentList(ctx context.Context, req *personalspacereq.AgentListReq) (resp *personalspaceresp.AgentListResp, err error) {
	resp = personalspaceresp.NewAgentListResp()

	// 1. 参数验证
	if req == nil {
		err = capierr.New400Err(ctx, "[PersonalSpaceService][AgentList]: 请求参数不能为空（req == nil）")
		return
	}

	// 2. 获取当前用户ID
	uid := chelper.GetUserIDFromCtx(ctx)
	if uid == "" {
		err = capierr.New401Err(ctx, "用户未登录")
		return
	}

	// 2.1. 检查用户是否有内置Agent管理权限
	hasBuiltInAgentMgmtPermission, err := s.isHasBuiltInAgentMgmtPermission(ctx)
	if err != nil {
		err = errors.Wrapf(err, "[PersonalSpaceService][AgentList]: get agent list from repo failed")
		return
	}

	// 3. 从数据库获取Agent列表
	needSize := req.Size
	req.Size++

	var agentIDsByBizDomain []string

	if !global.GConfig.IsBizDomainDisabled() {
		// 3.1. 获取当前用户当前的业务域ID
		bdIDs := []string{
			chelper.GetBizDomainIDFromCtx(ctx),
		}
		// 3.2. 获取当前用户所属的业务域ID下的Agent ID列表
		agentIDsByBizDomain, _, err = s.bizDomainHttp.GetAllAgentIDList(ctx, bdIDs)
		if err != nil {
			err = errors.Wrapf(err, "[PersonalSpaceService][AgentList]: get agent list from repo failed")
			return
		}

		// 如果此业务域下没有agent，直接返回
		if len(agentIDsByBizDomain) == 0 {
			return
		}
	}

	// 3.3. 构建参数
	argDto := psdbarg.NewAgentListArg(req, uid, hasBuiltInAgentMgmtPermission, agentIDsByBizDomain)

	// 3.4. 从数据库获取Agent列表
	pos, err := s.personalSpaceRepo.ListPersonalSpaceAgent(ctx, argDto)
	if err != nil {
		err = errors.Wrapf(err, "[PersonalSpaceService][AgentList]: get agent list from repo failed")
		return
	}

	// 4. 构建响应

	if len(pos) == 0 {
		resp.IsLastPage = true
		return
	}

	// 4.1. 如果pos长度大于needSize，说明还有下一页
	if len(pos) > needSize {
		pos = pos[:needSize]
		resp.IsLastPage = false
	} else {
		resp.IsLastPage = true
	}

	// 4.1. 汇总Agent ids
	agentIDs := make([]string, 0, len(pos))
	for _, po := range pos {
		agentIDs = append(agentIDs, po.ID)
	}

	eos, err := personalspacep2e.AgentsListForPersonalSpaces(ctx, pos, s.umHttp)
	if err != nil {
		err = errors.Wrapf(err, "[PersonalSpaceService][AgentList]: convert pos to eos failed")
		return
	}

	// 4.2. 批量获取已发布Agent信息
	releaseAgentPoMap := make(map[string]*dapo.PublishedJoinPo)

	if len(agentIDs) > 0 {
		arg := padbarg.NewGetPAPoListByIDArg(agentIDs, nil)

		var ret *padbret.GetPaPoMapByXxRet

		ret, err = s.pubedAgentRepo.GetPubedPoMapByXx(ctx, arg)
		if err != nil {
			err = errors.Wrap(err, "[PersonalSpaceService][AgentList]: call pubedAgentRepo.GetPubedPoMap失败")
			return
		}

		releaseAgentPoMap = ret.JoinPosID2PoMap
	}

	err = resp.LoadFromEos(eos, releaseAgentPoMap)
	if err != nil {
		err = errors.Wrapf(err, "[PersonalSpaceService][AgentList]: load from eos failed")
		return
	}

	return
}

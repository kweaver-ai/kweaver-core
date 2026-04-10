package personalspacesvc

import (
	"context"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/p2e/tplp2e"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/dbaccess/personalspacedbacc/psdbarg"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/personal_space/personalspacereq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/personal_space/personalspaceresp"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/global"
	"github.com/pkg/errors"
)

// AgentTplList 获取个人空间Agent模板列表
func (s *PersonalSpaceService) AgentTplList(ctx context.Context, req *personalspacereq.AgentTplListReq) (resp *personalspaceresp.AgentTplListResp, err error) {
	resp = personalspaceresp.NewAgentTplListResp()

	// 1. 参数验证
	if req == nil {
		err = capierr.New400Err(ctx, "[PersonalSpaceService][AgentTplList]: 请求参数不能为空（req == nil）")
		return
	}

	// 2. 获取当前用户ID
	uid := chelper.GetUserIDFromCtx(ctx)
	if uid == "" {
		err = capierr.New401Err(ctx, "用户未登录")
		return
	}

	// 3. 从数据库获取模板列表

	needSize := req.Size
	req.Size++

	// 3.1. 构建argDto
	var tplIDsByBdID []string

	if !global.GConfig.IsBizDomainDisabled() {
		bdID := chelper.GetBizDomainIDFromCtx(ctx)

		tplIDsByBdID, err = s.bizDomainHttp.GetAllAgentTplIDList(ctx, []string{bdID})
		if err != nil {
			err = errors.Wrapf(err, "[PersonalSpaceService][AgentTplList]: get all agent tpl id list failed")
			return
		}
		// 如果此业务域下没有agent tpl，直接返回
		if len(tplIDsByBdID) == 0 {
			return
		}
	}

	argDto := psdbarg.NewTplListArg(req, uid, tplIDsByBdID)

	// 3.2. 从数据库获取模板列表
	pos, err := s.personalSpaceRepo.ListPersonalSpaceTpl(ctx, argDto)
	if err != nil {
		err = errors.Wrapf(err, "[PersonalSpaceService][AgentTplList]: get agent template list from repo failed")
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

	// 4.2. pos 转换为eos
	eos, err := tplp2e.AgentTplListEos(ctx, pos, s.umHttp)
	if err != nil {
		err = errors.Wrapf(err, "[PersonalSpaceService][AgentTplList]: convert pos to eos failed")
		return
	}

	// 4.3. eos 转换为resp
	err = resp.LoadFromEos(eos)
	if err != nil {
		err = errors.Wrapf(err, "[PersonalSpaceService][AgentTplList]: load from eos failed")
		return
	}

	return
}

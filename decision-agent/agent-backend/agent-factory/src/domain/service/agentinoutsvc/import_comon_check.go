package agentinoutsvc

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/entity/daconfeo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/p2e/daconfp2e"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent_config/agentconfigreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent_inout/agentinoutresp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/pkg/errors"
)

func (s *agentInOutSvc) importCheck(ctx context.Context, exportData *agentinoutresp.ExportResp, resp *agentinoutresp.ImportResp) (err error) {
	// 1. 检查配置是否合法
	s.checkAgentConfigValid(ctx, exportData, resp)

	// 2. 检查是否有创建系统Agent权限
	err = s.checkSystemAgentCreatePermission(ctx, exportData, resp)
	if err != nil {
		return
	}

	return
}

// checkAgentConfigValid 检查Agent配置是否合法
func (s *agentInOutSvc) checkAgentConfigValid(ctx context.Context, exportData *agentinoutresp.ExportResp, resp *agentinoutresp.ImportResp) {
	for _, agent := range exportData.Agents {
		// 1.1 po转eo
		var eo *daconfeo.DataAgent

		var _err error

		eo, _err = daconfp2e.DataAgentSimple(ctx, agent.DataAgentPo)
		if _err != nil {
			resp.AddConfigInvalid(agent.Key, agent.Name)
			continue
		}

		// 1.2 eo转dto
		createReq := &agentconfigreq.CreateReq{}

		_err = cutil.CopyStructUseJSON(createReq, eo)
		if _err != nil {
			resp.AddConfigInvalid(agent.Key, agent.Name)
			continue
		}

		// 1.3 校验dto
		err1 := createReq.UpdateReq.ReqCheckWithCtx(ctx)
		err2 := createReq.UpdateReq.Validate()

		if err1 != nil || err2 != nil {
			resp.AddConfigInvalid(agent.Key, agent.Name)
			continue
		}
	}
}

// checkSystemAgentCreatePermission 检查是否有创建系统Agent权限
func (s *agentInOutSvc) checkSystemAgentCreatePermission(ctx context.Context, exportData *agentinoutresp.ExportResp, resp *agentinoutresp.ImportResp) (err error) {
	sysAgentFailItems := exportData.GetSystemAgentFailItems()
	if len(sysAgentFailItems) == 0 {
		return
	}

	// 检查是否有创建系统Agent权限
	var hasPms bool

	hasPms, err = s.isHasSystemAgentCreatePermission(ctx)
	if err != nil {
		err = errors.Wrapf(err, "check system agent create permission failed")
		return
	}

	if !hasPms {
		resp.NoCreateSystemAgentPms = sysAgentFailItems
	}

	return
}

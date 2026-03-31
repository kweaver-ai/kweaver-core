package othersvc

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/other/otherreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/other/otherresp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/apierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/pkg/errors"
)

func (s *otherSvc) DolphinTplList(ctx context.Context, req *otherreq.DolphinTplListReq) (resp *otherresp.DolphinTplListResp, err error) {
	resp = otherresp.NewDolphinTplListResp()

	// 1. 如果 builtInAgentKey 不为空，检查 agent 是否存在，且是内置 agent
	builtInAgentKey := req.BuiltInAgentKey.String()

	if builtInAgentKey != "" {
		var agentPo *dapo.DataAgentPo

		agentPo, err = s.agentConfRepo.GetByKey(ctx, builtInAgentKey)
		if err != nil {
			if chelper.IsSqlNotFound(err) {
				err = capierr.NewCustom404Err(ctx, apierr.DataAgentConfigNotFound, "agent不存在")
				return
			}

			err = errors.Wrapf(err, "[DolphinTplList]: get agent by key %s", builtInAgentKey)

			return
		}

		if !agentPo.IsBuiltInBool() {
			err = capierr.New400Err(ctx, "built_in_agent_key 不合法，不是内置 agent")
			return
		}
	}

	// 2. 生成 resp
	err = resp.LoadFromConfig(req)

	return
}

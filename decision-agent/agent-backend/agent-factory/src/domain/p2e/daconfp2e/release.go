package daconfp2e

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/entity/releaseeo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/daconfvalobj"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/pkg/errors"
)

func ReleaseDAConfEoSimple(ctx context.Context, _po *dapo.ReleasePO) (eo *releaseeo.ReleaseDAConfWrapperEO, err error) {
	eo = &releaseeo.ReleaseDAConfWrapperEO{
		Config: &daconfvalobj.Config{},
	}

	err = cutil.CopyStructUseJSON(&eo.ReleaseEO, _po)
	if err != nil {
		return
	}

	// 1. 解析配置
	if _po.AgentConfig != "" {
		// 1.1. _po.AgentConfig -> DataAgentPo
		agentConfPo := &dapo.DataAgentPo{}

		err = cutil.JSON().UnmarshalFromString(_po.AgentConfig, agentConfPo)
		if err != nil {
			err = errors.Wrapf(err, "ReleaseSimple unmarshal1 config error")
			return
		}

		// 1.2. agentConfPo.Config -> eo.Config
		err = cutil.JSON().UnmarshalFromString(agentConfPo.Config, &eo.Config)
		if err != nil {
			err = errors.Wrapf(err, "ReleaseSimple unmarshal2 config error")
			return
		}
	}

	return
}

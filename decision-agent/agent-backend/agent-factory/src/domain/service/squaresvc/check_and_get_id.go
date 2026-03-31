package squaresvc

import (
	"context"
	"database/sql"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/apierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capierr"

	"github.com/pkg/errors"
)

func (svc *squareSvc) CheckAndGetID(ctx context.Context, agentID string) (newAgentID string, err error) {
	exists, err := svc.agentConfRepo.ExistsByID(ctx, agentID)
	if err != nil {
		err = errors.Wrapf(err, "svc.agentConfRepo.ExistsByID(ctx, %s)", agentID)
		return
	}

	if exists {
		newAgentID = agentID
		return
	}

	agentV0CfgPo, err := svc.agentConfRepo.GetByKey(ctx, agentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = capierr.NewCustom404Err(ctx, apierr.DataAgentConfigNotFound, "agent not found")
			return
		}

		return
	}

	newAgentID = agentV0CfgPo.ID

	return
}

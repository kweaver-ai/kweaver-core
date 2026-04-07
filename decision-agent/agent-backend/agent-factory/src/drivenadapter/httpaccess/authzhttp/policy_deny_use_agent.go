package authzhttp

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/authzhttp/authzhttpreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
)

func (a *authZHttpAcc) DenyAgentUsePmsForAllAccessor(ctx context.Context, agentID string, agentName string) (err error) {
	accessors := []*authzhttpreq.PolicyAccessor{
		{
			ID:   "*",
			Type: cenum.PmsTargetObjTypeRole,
		},
		{
			ID:   "*",
			Type: cenum.PmsTargetObjTypeUser,
		},
		{
			ID:   "*",
			Type: cenum.PmsTargetObjTypeUserGroup,
		},
		{
			ID:   "*",
			Type: cenum.PmsTargetObjTypeAppAccount,
		},
		{
			ID:   "*",
			Type: cenum.PmsTargetObjTypeDep,
		},
	}

	reqs, err := authzhttpreq.NewDenyAgentUseReqs(accessors, agentID, agentName)
	if err != nil {
		return
	}

	err = a.CreatePolicy(ctx, reqs)
	if err != nil {
		return
	}

	return
}

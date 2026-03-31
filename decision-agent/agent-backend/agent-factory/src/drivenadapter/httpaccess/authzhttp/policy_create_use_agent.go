package authzhttp

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/constant/cdapmsconstant"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/authzhttp/authzhttpreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
)

// GrantAgentUsePmsForAppAdmin 给应用管理员授予Agent使用权限
func (a *authZHttpAcc) GrantAgentUsePmsForAppAdmin(ctx context.Context) (err error) {
	accessors := []*authzhttpreq.PolicyAccessor{
		{
			ID:   cdapmsconstant.InnerRoleAppAdminID,
			Type: cenum.PmsTargetObjTypeRole,
		},
	}

	// 1. 给应用管理员授予agent资源类型的使用权限
	agentID := "*"
	agentName := "所有Agent"

	reqs, err := authzhttpreq.NewGrantAgentUseReqs(accessors, agentID, agentName)
	if err != nil {
		return
	}

	err = a.CreatePolicy(ctx, reqs)
	if err != nil {
		return
	}

	return
}

func (a *authZHttpAcc) GrantAgentUsePmsForSingleAccessor(ctx context.Context, accessor *authzhttpreq.PolicyAccessor, agentID string, agentName string) (err error) {
	req := authzhttpreq.NewGrantAgentUseReq(accessor, agentID, agentName)

	err = a.CreatePolicy(ctx, []*authzhttpreq.CreatePolicyReq{req})
	if err != nil {
		return
	}

	return
}

func (a *authZHttpAcc) GrantAgentUsePmsForAccessors(ctx context.Context, accessors []*authzhttpreq.PolicyAccessor, agentID string, agentName string) (err error) {
	reqs, err := authzhttpreq.NewGrantAgentUseReqs(accessors, agentID, agentName)
	if err != nil {
		return
	}

	err = a.CreatePolicy(ctx, reqs)
	if err != nil {
		return
	}

	return
}

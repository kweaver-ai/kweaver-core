package authzhttp

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/constant/cdapmsconstant"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdapmsenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/authzhttp/authzhttpreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
)

// GrantMgmtPmsForAppAdmin 给应用管理员授予应用管理权限
func (a *authZHttpAcc) GrantMgmtPmsForAppAdmin(ctx context.Context) (err error) {
	accessors := []*authzhttpreq.PolicyAccessor{
		{
			ID:   cdapmsconstant.InnerRoleAppAdminID,
			Type: cenum.PmsTargetObjTypeRole,
		},
	}

	// 1. 给应用管理员授予agent资源类型的管理权限
	agentID := "*"
	agentName := "所有Agent"
	operations := cdapmsenum.GetAllAgentMgmtOperator()

	reqs, err := authzhttpreq.NewGrantAgentMgmtReqs(accessors, agentID, agentName, operations)
	if err != nil {
		return
	}

	err = a.CreatePolicy(ctx, reqs)
	if err != nil {
		return
	}

	//// 2. 给应用管理员授予“自定义空间”资源类型的管理权限
	//customSpaceID := "*"
	//customSpaceName := "所有自定义空间"
	//operations = cdapmsenum.GetAllCustomSpaceMgmtOperator()
	//
	//reqs, err = authzhttpreq.NewGrantCustomSpaceMgmtReqs(accessors, customSpaceID, customSpaceName, operations)
	//if err != nil {
	//	return
	//}
	//
	//err = a.CreatePolicy(ctx, reqs)
	//if err != nil {
	//	return
	//}

	// 3. 给应用管理员授予“agent模板”资源类型的使用权限
	agentTplID := "*"
	agentTplName := "所有Agent模板"
	operations = cdapmsenum.GetAllAgentTplOperator()

	reqs, err = authzhttpreq.NewGrantAgentTplMgmtReqs(accessors, agentTplID, agentTplName, operations)
	if err != nil {
		return
	}

	err = a.CreatePolicy(ctx, reqs)
	if err != nil {
		return
	}

	return
}

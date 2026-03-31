package releasesvc

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/authzhttp/authzhttpreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/cenvhelper"
	"github.com/pkg/errors"
)

func (svc *releaseSvc) removeUsePmsByHTTPAcc(ctx context.Context, agentID string) (err error) {
	if cenvhelper.IsLocalDev() {
		return
	}

	// 删除Agent使用权限
	err = svc.authZHttp.DeleteAgentPolicy(ctx, agentID)
	if err != nil {
		err = errors.Wrapf(err, "remove use pms failed")
		return
	}

	return
}

// pmsMap-> key: cenum.PmsTargetObjType, value: []string（对象ID slice）
func (svc *releaseSvc) grantUsePms(ctx context.Context, agentID, agentName string, pmsMap map[cenum.PmsTargetObjType][]string) (err error) {
	accessors := make([]*authzhttpreq.PolicyAccessor, 0)

	for pmsTargetObjType, objIDs := range pmsMap {
		for _, objID := range objIDs {
			accessors = append(accessors, &authzhttpreq.PolicyAccessor{
				ID:   objID,
				Type: pmsTargetObjType,
			})
		}
	}

	if len(accessors) == 0 {
		return
	}

	if cenvhelper.IsLocalDev() {
		return
	}

	// 授予Agent使用权限
	err = svc.authZHttp.GrantAgentUsePmsForAccessors(ctx, accessors, agentID, agentName)
	if err != nil {
		err = errors.Wrapf(err, "grant use pms failed")
		return
	}

	return
}

// func (svc *releaseSvc) grantUsePmsToAll(ctx context.Context, agentID string, agentName string) (err error) {
//
//	// 授予Agent使用权限 to all user
//	accessor := &authzhttpreq.PolicyAccessor{
//		ID: "*",
//		Type: cenum.PmsTargetObjTypeUser,
//	}
//
//	err = svc.authZHttp.GrantAgentUsePmsForSingleAccessor(ctx, accessor, agentID, agentName)
//	if err != nil {
//		err = errors.Wrapf(err, "grant use pms to all user failed")
//		return
//	}
//
//	return
//}

package permissionsvc

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/rdto/agent_permission/cpmsreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/rdto/agent_permission/cpmsresp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/apierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/global"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/pkg/errors"
)

// CheckUsePermission 检查非个人空间下的某个agent是否有使用（运行）权限
func (svc *permissionSvc) CheckUsePermission(ctx context.Context, req *cpmsreq.CheckAgentRunReq) (resp *cpmsresp.CheckRunResp, err error) {
	resp = &cpmsresp.CheckRunResp{}

	if global.GConfig.SwitchFields.DisablePmsCheck {
		resp.IsAllowed = true
		return
	}

	// 1. uid or appAccountID

	uid := req.UserID
	if uid == "" {
		uid = chelper.GetUserIDFromCtx(ctx)
	}

	appAccountID := req.AppAccountID

	if uid == "" && appAccountID == "" {
		err = errors.New("user id or app account id cannot be all empty")
		return
	}

	// 2. 默认没有权限
	resp.IsAllowed = false

	// 3. 获取agent信息
	agentPo, err := svc.agentConfRepo.GetByID(ctx, req.AgentID)
	if err != nil {
		if chelper.IsSqlNotFound(err) {
			err = capierr.NewCustom404Err(ctx, apierr.DataAgentConfigNotFound, "agent不存在")
			return
		}

		err = errors.Wrapf(err, "[CheckUsePermission][GetByID]: get agent by id %s", req.AgentID)

		return
	}

	// 4. 如果不是创建者，检查用户是否有权限使用该agent
	hasPms, err := svc.checkUserPms(ctx, agentPo, uid, appAccountID)
	if err != nil {
		return
	}

	if !hasPms {
		return
	}

	// 5. 检查通过
	resp.IsAllowed = true

	return
}

func (svc *permissionSvc) checkUserPms(ctx context.Context, agentPo *dapo.DataAgentPo, uid string, appAccountID string) (hasPms bool, err error) {
	// 1 如果是创建者，有权限
	if agentPo.CreatedBy == uid {
		hasPms = true
		return
	}

	// 2 获取release信息
	var releasePo *dapo.ReleasePO

	releasePo, err = svc.releaseRepo.GetByAgentID(ctx, agentPo.ID)
	if err != nil {
		err = errors.Wrapf(err, "[checkUserPms][GetByAgentID]: get release by agent id %s", agentPo.ID)
		return
	}

	// 3. 如果release存在，表示有已发布版本
	if releasePo != nil {
		// 3.1 如果开启了权限控制
		if releasePo.IsPmsCtrlBool() {
			hasPms, err = svc.checkByPmsPlatform(ctx, agentPo, uid, appAccountID)
			if err != nil {
				return
			}
		} else {
			hasPms = true
		}
	} else {
		// 3.2 如果release不存在，表示没有已发布版本
		// 这时不是owner，需要查询下权限平台（因为可能是有“管理内置agent”权限的，这时候也是有使用权限的）
		hasPms, err = svc.checkByPmsPlatform(ctx, agentPo, uid, appAccountID)
		if err != nil {
			return
		}
	}

	return
}

// 通过“权限平台”查询用户是否有权限使用该agent
func (svc *permissionSvc) checkByPmsPlatform(ctx context.Context, agentPo *dapo.DataAgentPo, uid string, appAccountID string) (hasPms bool, err error) {
	var ok bool

	// 1. 获取访问者ID和类型
	var (
		accessorID   string
		accessorType cenum.PmsTargetObjType
	)

	if appAccountID != "" {
		accessorID = appAccountID
		accessorType = cenum.PmsTargetObjTypeAppAccount
	} else if uid != "" {
		accessorID = uid
		accessorType = cenum.PmsTargetObjTypeUser
	} else {
		err = errors.New("user id or app account id cannot be all empty")
		return
	}

	// 2. 查询权限平台
	ok, err = svc.authZHttp.SingleAgentUseCheck(ctx, accessorID, accessorType, agentPo.ID)
	if err != nil {
		err = errors.Wrapf(err, "[checkByPmsPlatform][SingleAgentUseCheck]: single agent use check failed")
		return
	}

	// 3. 设置hasPms
	if !ok {
		return
	}

	hasPms = true

	return
}

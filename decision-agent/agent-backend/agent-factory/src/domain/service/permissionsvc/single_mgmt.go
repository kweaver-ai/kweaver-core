package permissionsvc

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdapmsenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/cenvhelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/global"
	"github.com/pkg/errors"
)

func (svc *permissionSvc) GetSingleMgmtPermission(ctx context.Context, resourceType cdaenum.ResourceType, operator cdapmsenum.Operator) (allAllowed bool, err error) {
	// 1. 检查是否禁用权限检查
	if global.GConfig.SwitchFields.DisablePmsCheck {
		allAllowed = true
		return
	}

	if cenvhelper.IsLocalDev() {
		allAllowed = false
		return
	}

	// 2. 获取当前用户ID
	uid := chelper.GetUserIDFromCtx(ctx)
	if uid == "" {
		err = errors.New("user id is empty")
		return
	}

	// 3. 获取用户权限
	var m map[cdapmsenum.Operator]bool

	switch resourceType {
	case cdaenum.ResourceTypeDataAgent:
		m, err = svc.authZHttp.GetAgentResourceOpsByUid(ctx, uid)
	case cdaenum.ResourceTypeDataAgentTpl:
		m, err = svc.authZHttp.GetAgentTplResourceOpsByUid(ctx, uid)
	default:
		err = errors.New("invalid resource type")
		return
	}

	// 4. 返回权限
	allAllowed = m[operator]

	return
}

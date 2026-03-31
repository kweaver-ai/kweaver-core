package permissionsvc

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdapmsenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/rdto/agent_permission/cpmsresp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/global"
	"github.com/pkg/errors"
)

// GetUserStatus 获取用户拥有的管理权限状态
func (svc *permissionSvc) GetUserStatus(ctx context.Context) (resp *cpmsresp.UserStatusResp, err error) {
	resp = cpmsresp.NewUserStatusResp()

	if global.GConfig.SwitchFields.DisablePmsCheck {
		resp = cpmsresp.NewUserStatusRespAllAllowed()
		return
	}

	// 1. 获取当前用户ID
	userID := chelper.GetUserIDFromCtx(ctx)
	if userID == "" {
		err = errors.New("user id is empty")
		return
	}

	// 2. 获取用户权限
	// 2.1 获取Agent权限
	agentOpMap, err := svc.authZHttp.GetAgentResourceOpsByUid(ctx, userID)
	if err != nil {
		err = errors.Wrapf(err, "get agent resource ops by uid failed")
		return
	}

	// 2.2 获取Agent模板权限
	agentTplOpMap, err := svc.authZHttp.GetAgentTplResourceOpsByUid(ctx, userID)
	if err != nil {
		err = errors.Wrapf(err, "get agent tpl resource ops by uid failed")
		return
	}

	// 3. 设置权限响应

	// 3.1 设置Agent权限
	resp.Agent = cpmsresp.AgentPermission{
		Publish:                  agentOpMap[cdapmsenum.AgentPublish],
		Unpublish:                agentOpMap[cdapmsenum.AgentUnpublish],
		UnpublishOtherUserAgent:  agentOpMap[cdapmsenum.AgentUnpublishOtherUserAgent],
		PublishToBeSkillAgent:    agentOpMap[cdapmsenum.AgentPublishToBeSkillAgent],
		PublishToBeWebSdkAgent:   agentOpMap[cdapmsenum.AgentPublishToBeWebSdkAgent],
		PublishToBeApiAgent:      agentOpMap[cdapmsenum.AgentPublishToBeApiAgent],
		PublishToBeDataFlowAgent: agentOpMap[cdapmsenum.AgentPublishToBeDataFlowAgent],
		CreateSystemAgent:        agentOpMap[cdapmsenum.AgentCreateSystemAgent],
		MgntBuiltInAgent:         agentOpMap[cdapmsenum.AgentBuiltInAgentMgmt],
		SeeTrajectoryAnalysis:    agentOpMap[cdapmsenum.AgentSeeTrajectoryAnalysis],
	}

	// 3.2 设置Agent模板权限
	resp.AgentTpl = cpmsresp.AgentTplPermission{
		Publish:                    agentTplOpMap[cdapmsenum.AgentTplPublish],
		Unpublish:                  agentTplOpMap[cdapmsenum.AgentTplUnpublish],
		UnpublishOtherUserAgentTpl: agentTplOpMap[cdapmsenum.AgentTplUnpublishOtherUserAgentTpl],
	}

	return
}

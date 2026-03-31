package v3agentconfigsvc

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdapmsenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/apierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/pkg/errors"
)

func (s *dataAgentConfigSvc) isHasTplPublishPermission(ctx context.Context) (has bool, err error) {
	has, err = s.pmsSvc.GetSingleMgmtPermission(ctx, cdaenum.ResourceTypeDataAgentTpl, cdapmsenum.AgentTplPublish)
	return
}

func (s *dataAgentConfigSvc) isHasBuiltInAgentMgmtPermission(ctx context.Context) (has bool, err error) {
	has, err = s.pmsSvc.GetSingleMgmtPermission(ctx, cdaenum.ResourceTypeDataAgent, cdapmsenum.AgentBuiltInAgentMgmt)
	return
}

func (s *dataAgentConfigSvc) isHasSystemAgentCreatePermission(ctx context.Context) (has bool, err error) {
	has, err = s.pmsSvc.GetSingleMgmtPermission(ctx, cdaenum.ResourceTypeDataAgent, cdapmsenum.AgentCreateSystemAgent)
	return
}

func (s *dataAgentConfigSvc) checkUseAgentPms(ctx context.Context, m map[string]*dapo.PublishedJoinPo, uid string) (hasPmsMap map[string]struct{}, err error) {
	hasPmsMap = make(map[string]struct{})

	var agentIds []string

	for _, po := range m {
		if po.IsPmsCtrlBool() {
			agentIds = append(agentIds, po.ID)
		}
	}

	// 3.2 获取过滤后的、有权限的Agent ID
	var filteredAgentIdMap map[string]struct{}

	filteredAgentIdMap, err = s.authZHttp.FilterCanUseAgentIDMap(ctx, uid, agentIds)
	if err != nil {
		err = errors.Wrapf(err, "[publishedSvc][GetPublishedAgentList]: filter can use agent ids failed")
		return
	}

	// 3.3 标记有权限的Agent

	for _, po := range m {
		if !po.IsPmsCtrlBool() {
			hasPmsMap[po.Key] = struct{}{}
		} else {
			if _, ok := filteredAgentIdMap[po.ID]; ok {
				hasPmsMap[po.Key] = struct{}{}
			}
		}
	}

	return
}

func (s *dataAgentConfigSvc) isOwnerOrHasBuiltInAgentMgmtPermission(ctx context.Context, po *dapo.DataAgentPo, uid string) (err error) {
	// 1. owner，直接返回
	if po.CreatedBy == uid {
		return
	}

	// 2. 不是owner时判断是否是内置Agent，并是否有内置Agent管理权限
	var hasBuiltInAgentMgmtPermission bool

	isBuiltIn := po.IsBuiltIn.IsBuiltIn()

	if isBuiltIn {
		hasBuiltInAgentMgmtPermission, err = s.isHasBuiltInAgentMgmtPermission(ctx)
		if err != nil {
			return
		}
	}

	// 如果不是内置Agent，或者没有内置Agent管理权限，返回403
	if !isBuiltIn || !hasBuiltInAgentMgmtPermission {
		err = capierr.NewCustom403Err(ctx, apierr.AgentFactoryPermissionForbidden, "不是owner。不是内置Agent或者没有内置Agent管理权限")
		return
	}

	return
}

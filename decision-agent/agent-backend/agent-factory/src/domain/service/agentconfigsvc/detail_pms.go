package v3agentconfigsvc

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/entity/daconfeo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/daconfvalobj/skillvalobj"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/dbaccess/pubedagentdbacc/padbarg"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
)

func (s *dataAgentConfigSvc) detailPmsCheck(ctx context.Context, po *dapo.DataAgentPo, isPrivate bool, uid string) (err error) {
	// 1. 私有API，不检查
	if isPrivate {
		return
	}

	// 2. 检查owner或内置Agent管理权限
	err = s.isOwnerOrHasBuiltInAgentMgmtPermission(ctx, po, uid)
	if err != nil {
		return
	}

	return
}

func (s *dataAgentConfigSvc) markSkillAgentPmsForDetail(ctx context.Context, eo *daconfeo.DataAgent, uid string) (err error) {
	skillAgents := make([]*skillvalobj.SkillAgent, 0)

	// 1. 获取技能配置中的Agent
	if eo.Config.Skill != nil && len(eo.Config.Skill.Agents) > 0 {
		skillAgents = eo.Config.Skill.Agents
	}

	if len(skillAgents) == 0 {
		return
	}

	// 2. 获取技能配置中的Agent Key
	agentKeys := make([]string, 0)
	for _, skillAgent := range skillAgents {
		agentKeys = append(agentKeys, skillAgent.AgentKey)
	}

	// 3. 获取技能配置中的已发布Agent
	ret, err := s.pubedAgentRepo.GetPubedPoMapByXx(ctx, padbarg.NewGetPaPoListByKeyArg(agentKeys, nil))
	if err != nil {
		return
	}

	pubedAgentMap := ret.JoinPosKey2PoMap

	// 4. 检查技能配置中的Agent是否有权限
	hasPmsMap, err := s.checkUseAgentPms(ctx, pubedAgentMap, uid)
	if err != nil {
		return
	}

	// 5. 标记技能配置中的Agent
	for _, skillAgent := range skillAgents {
		if _, ok := pubedAgentMap[skillAgent.AgentKey]; !ok {
			skillAgent.CurrentIsExistsAndPublished = false
			continue
		}

		skillAgent.CurrentIsExistsAndPublished = true

		if _, ok := hasPmsMap[skillAgent.AgentKey]; ok {
			skillAgent.CurrentPmsCheckStatus = skillvalobj.CurrentPmsCheckStatusSuccess
		} else {
			skillAgent.CurrentPmsCheckStatus = skillvalobj.CurrentPmsCheckStatusFailed
		}
	}

	return
}

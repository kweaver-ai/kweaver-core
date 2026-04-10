package agentinoutsvc

import (
	"context"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/agent_inout/agentinoutresp"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/global"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/pkg/errors"
)

func (s *agentInOutSvc) importByUpsertCheck(ctx context.Context, exportData *agentinoutresp.ExportResp, uid string, resp *agentinoutresp.ImportResp) (existingAgentMap map[string]*dapo.DataAgentPo, err error) {
	// 1. 检查业务域冲突
	err = s.checkBizDomainConflict(ctx, exportData, resp)
	if err != nil {
		return
	}
	// 这里有问题时，直接返回，这样逻辑可能比较简单
	if resp.HasFail() {
		return
	}

	// 2. 检查是否有重复的agent key
	existingAgentMap, err = s.upsertCheckRepeatAgentKey(ctx, exportData, uid, resp)
	if err != nil {
		return
	}

	return
}

// checkBizDomainConflict 检查业务域冲突
// 检查导入的agent key是否有不在当前业务域中的
func (s *agentInOutSvc) checkBizDomainConflict(ctx context.Context, exportData *agentinoutresp.ExportResp, resp *agentinoutresp.ImportResp) (err error) {
	if global.GConfig.IsBizDomainDisabled() {
		return nil
	}

	// 1. 从header获取当前业务域的id
	bdID := chelper.GetBizDomainIDFromCtx(ctx)

	// 2. 使用此id获取当前业务域的所有agent id
	agentIDsByBdID, _, err := s.bizDomainHttp.GetAllAgentIDList(ctx, []string{bdID})
	if err != nil {
		err = errors.Wrapf(err, "get all agent id list by biz domain id failed")
		return
	}

	// 将业务域下的agent id列表转为map，方便查找
	bdAgentIDSet := make(map[string]struct{}, len(agentIDsByBdID))
	for _, agentID := range agentIDsByBdID {
		bdAgentIDSet[agentID] = struct{}{}
	}

	// 3. 汇总导入的agent key
	agentKeys := make([]string, 0, len(exportData.Agents))
	agentKeyMap := make(map[string]*agentinoutresp.ExportAgentItem, len(exportData.Agents))

	for _, agent := range exportData.Agents {
		agentKeys = append(agentKeys, agent.Key)
		agentKeyMap[agent.Key] = agent
	}

	// 4. 查询导入的agent key对应的agent id列表
	pos, err := s.agentConfRepo.GetByKeys(ctx, agentKeys)
	if err != nil {
		err = errors.Wrapf(err, "get agent by keys failed")
		return
	}

	// 5. 比较导入的agent id列表和当前业务域的agent id列表
	// 如果导入的agent id有不在当前业务域的agent id列表中，则将导入的agent key添加到失败列表中
	for _, po := range pos {
		if _, ok := bdAgentIDSet[po.ID]; !ok {
			// 该agent不在当前业务域中
			if agent, exists := agentKeyMap[po.Key]; exists {
				resp.AddBizDomainConflict(agent.Key, agent.Name)
			}
		}
	}

	return
}

func (s *agentInOutSvc) upsertCheckRepeatAgentKey(ctx context.Context, exportData *agentinoutresp.ExportResp, uid string, resp *agentinoutresp.ImportResp) (existingAgentMap map[string]*dapo.DataAgentPo, err error) {
	// 1. 检查是否有重复的agent key
	agentKeys := make([]string, 0)
	for _, agent := range exportData.Agents {
		agentKeys = append(agentKeys, agent.Key)
	}

	pos, err := s.agentConfRepo.GetByKeys(ctx, agentKeys)
	if err != nil {
		return
	}

	existingAgentMap = make(map[string]*dapo.DataAgentPo, len(pos))

	for _, po := range pos {
		if po.CreatedBy != uid {
			resp.AddAgentKeyConflict(po.Key, po.Name)
			continue
		}

		existingAgentMap[po.Key] = po
	}

	return
}

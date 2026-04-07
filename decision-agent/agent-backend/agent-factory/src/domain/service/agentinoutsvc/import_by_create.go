package agentinoutsvc

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/bizdomainhttp/bizdomainhttpreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent_inout/agentinoutresp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/global"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
)

func (s *agentInOutSvc) importByCreate(ctx context.Context, exportData *agentinoutresp.ExportResp, uid string, resp *agentinoutresp.ImportResp) (err error) {
	// 1. 检查导入数据
	err = s.importByCreateCheck(ctx, exportData, resp)
	if err != nil {
		return
	}

	if resp.HasFail() {
		return
	}

	// 2. 开启事务
	tx, err := s.agentConfRepo.BeginTx(ctx)
	if err != nil {
		return
	}
	defer chelper.TxRollbackOrCommit(tx, &err, s.logger)

	// 3. 构建批量导入的agent pos
	now := cutil.GetCurrentMSTimestamp()

	pos := make([]*dapo.DataAgentPo, 0, len(exportData.Agents))

	for _, agentItem := range exportData.Agents {
		// 生成新的ID
		newID := cutil.UlidMake()

		// 设置导入相关字段
		agentItem.DataAgentPo.ResetForImport()
		agentItem.DataAgentPo.ID = newID
		agentItem.DataAgentPo.CreatedBy = uid
		agentItem.DataAgentPo.UpdatedBy = uid
		agentItem.DataAgentPo.CreatedAt = now
		agentItem.DataAgentPo.UpdatedAt = now

		pos = append(pos, agentItem.DataAgentPo)
	}

	// 4. 批量导入agent
	err = s.agentConfRepo.CreateBatch(ctx, tx, pos)
	if err != nil {
		return
	}

	if global.GConfig.IsBizDomainDisabled() {
		return
	}

	// 5. 关联业务域（先写入本地关联表，再调用HTTP）
	bdID := chelper.GetBizDomainIDFromCtx(ctx)

	// 5.1 构建本地关联表数据
	bdRelPos := make([]*dapo.BizDomainAgentRelPo, 0, len(pos))
	agentIDs := make([]string, 0, len(pos))

	for _, po := range pos {
		bdRelPos = append(bdRelPos, &dapo.BizDomainAgentRelPo{
			BizDomainID: bdID,
			AgentID:     po.ID,
			CreatedAt:   now,
		})

		agentIDs = append(agentIDs, po.ID)
	}

	// 5.2 写入本地关联表
	err = s.bdAgentRelRepo.BatchCreate(ctx, tx, bdRelPos)
	if err != nil {
		return
	}

	// 5.3 调用HTTP接口批量关联
	batchReq := make(bizdomainhttpreq.AssociateResourceBatchReq, 0, len(agentIDs))
	for _, agentID := range agentIDs {
		batchReq = append(batchReq, &bizdomainhttpreq.AssociateResourceItem{
			BdID: cenum.BizDomainID(bdID),
			ID:   agentID,
			Type: cdaenum.ResourceTypeDataAgent,
		})
	}

	err = s.bizDomainHttp.AssociateResourceBatch(ctx, batchReq)
	if err != nil {
		return
	}

	return
}

func (s *agentInOutSvc) importByCreateCheck(ctx context.Context, exportData *agentinoutresp.ExportResp, resp *agentinoutresp.ImportResp) (err error) {
	// 1. 检查是否有重复的agent key
	existingKeys := make([]string, 0)
	for _, agent := range exportData.Agents {
		existingKeys = append(existingKeys, agent.Key)
	}

	conflictAgentPos, err := s.agentConfRepo.GetByKeys(ctx, existingKeys)
	if err != nil {
		return
	}

	if len(conflictAgentPos) > 0 {
		// 构建冲突响应
		for _, po := range conflictAgentPos {
			resp.AddAgentKeyConflict(po.Key, po.Name)
		}
	}

	return
}

package agentinoutsvc

import (
	"context"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/httpaccess/bizdomainhttp/bizdomainhttpreq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/agent_inout/agentinoutresp"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/global"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
)

func (s *agentInOutSvc) importByUpsert(ctx context.Context, exportData *agentinoutresp.ExportResp, uid string, resp *agentinoutresp.ImportResp) (err error) {
	// 1. 导入前验证
	existingAgentMap, err := s.importByUpsertCheck(ctx, exportData, uid, resp)
	if err != nil {
		return
	}

	if resp.HasFail() {
		return
	}

	// 2. upsert agent

	tx, err := s.agentConfRepo.BeginTx(ctx)
	if err != nil {
		return
	}
	defer chelper.TxRollbackOrCommit(tx, &err, s.logger)

	createPos := make([]*dapo.DataAgentPo, 0)
	updatePos := make([]*dapo.DataAgentPo, 0)

	now := cutil.GetCurrentMSTimestamp()

	for _, agent := range exportData.Agents {
		if _, ok := existingAgentMap[agent.Key]; ok {
			// 更新
			// 设置导入相关字段
			agent.DataAgentPo.UpdatedBy = uid
			agent.DataAgentPo.UpdatedAt = now

			agent.DataAgentPo.Status = cdaenum.StatusUnpublished
			agent.DataAgentPo.SetIsBuiltIn(cdaenum.BuiltInNo)

			updatePos = append(updatePos, agent.DataAgentPo)
		} else {
			// 插入
			// 生成新的ID
			newID := cutil.UlidMake()

			// 设置导入相关字段
			agent.DataAgentPo.ResetForImport()
			agent.DataAgentPo.ID = newID
			agent.DataAgentPo.CreatedBy = uid
			agent.DataAgentPo.UpdatedBy = uid
			agent.DataAgentPo.CreatedAt = now
			agent.DataAgentPo.UpdatedAt = now

			createPos = append(createPos, agent.DataAgentPo)
		}
	}

	if len(createPos) > 0 {
		if err = s.agentConfRepo.CreateBatch(ctx, tx, createPos); err != nil {
			return
		}
	}

	if len(updatePos) > 0 {
		for _, po := range updatePos {
			if err = s.agentConfRepo.UpdateByKey(ctx, tx, po); err != nil {
				return
			}
		}
	}

	// 3. 关联业务域（只对新创建的agent添加关联，更新的agent已有关联）
	if len(createPos) > 0 && !global.GConfig.IsBizDomainDisabled() {
		bdID := chelper.GetBizDomainIDFromCtx(ctx)

		// 3.1 构建本地关联表数据
		bdRelPos := make([]*dapo.BizDomainAgentRelPo, 0, len(createPos))

		for _, po := range createPos {
			bdRelPos = append(bdRelPos, &dapo.BizDomainAgentRelPo{
				BizDomainID: bdID,
				AgentID:     po.ID,
				CreatedAt:   now,
			})
		}

		// 3.2 写入本地关联表
		err = s.bdAgentRelRepo.BatchCreate(ctx, tx, bdRelPos)
		if err != nil {
			return
		}

		// 3.3 调用HTTP接口批量关联
		batchReq := make(bizdomainhttpreq.AssociateResourceBatchReq, 0, len(createPos))
		for _, po := range createPos {
			batchReq = append(batchReq, &bizdomainhttpreq.AssociateResourceItem{
				BdID: cenum.BizDomainID(bdID),
				ID:   po.ID,
				Type: cdaenum.ResourceTypeDataAgent,
			})
		}

		err = s.bizDomainHttp.AssociateResourceBatch(ctx, batchReq)
		if err != nil {
			return
		}
	}

	return
}

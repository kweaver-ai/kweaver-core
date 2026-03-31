package bizdomainsvc

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/bizdomainhttp/bizdomainhttpreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess"
	"github.com/pkg/errors"
)

// InitBizDomainAgentRel 初始化业务域与agent的关联关系
// 如果关联表已有数据，则跳过初始化
func (s *BizDomainSvc) InitBizDomainAgentRel(
	ctx context.Context,
	agentRepo idbaccess.IDataAgentConfigRepo,
	bdAgentRelRepo idbaccess.IBizDomainAgentRelRepo,
) (err error) {
	bdID := cenum.BizDomainPublic.ToString()

	// 1. 开启事务
	tx, err := bdAgentRelRepo.BeginTx(ctx)
	if err != nil {
		return errors.Wrap(err, "begin tx failed")
	}

	defer chelper.TxRollback(tx, &err, s.logger)

	// 2. 查询关联表是否已有数据
	existingRels, err := bdAgentRelRepo.GetByBizDomainID(ctx, tx, bdID)
	if err != nil {
		return errors.Wrap(err, "get existing agent rels failed")
	}

	// 如果已有数据，跳过初始化
	if len(existingRels) > 0 {
		s.logger.Infof("[InitBizDomainAgentRel] 关联表已有 %d 条数据，跳过初始化", len(existingRels))
		// 回滚事务（因为没有任何修改）
		_ = tx.Rollback()

		return nil
	}

	// 3. 获取所有agent ID
	agentIDs, err := agentRepo.GetAllIDs(ctx)
	if err != nil {
		return errors.Wrap(err, "get all agent ids failed")
	}

	if len(agentIDs) == 0 {
		s.logger.Infoln("[InitBizDomainAgentRel] 没有agent数据，跳过初始化")

		_ = tx.Rollback()

		return nil
	}

	s.logger.Infof("[InitBizDomainAgentRel] 准备初始化 %d 个agent的业务域关联", len(agentIDs))

	// 4. 先写入本地关联表
	currentTs := cutil.GetCurrentMSTimestamp()
	pos := make([]*dapo.BizDomainAgentRelPo, 0, len(agentIDs))

	for _, agentID := range agentIDs {
		pos = append(pos, &dapo.BizDomainAgentRelPo{
			BizDomainID: bdID,
			AgentID:     agentID,
			CreatedAt:   currentTs,
		})
	}

	err = bdAgentRelRepo.BatchCreate(ctx, tx, pos)
	if err != nil {
		return errors.Wrap(err, "batch create agent rels failed")
	}

	// 5. 调用HTTP接口批量关联
	httpReq := bizdomainhttpreq.NewInitAllAgentToPublicBusinessDomainReq(agentIDs)

	err = s.bizDomainHttp.AssociateResourceBatch(ctx, httpReq)
	if err != nil {
		return errors.Wrap(err, "associate resource batch failed")
	}

	// 6. 提交事务
	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "commit tx failed")
	}

	s.logger.Infof("[InitBizDomainAgentRel] 成功初始化 %d 个agent的业务域关联", len(agentIDs))

	return nil
}

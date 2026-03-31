package bizdomainsvc

import (
	"context"
	"strconv"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/bizdomainhttp/bizdomainhttpreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess"
	"github.com/pkg/errors"
)

// InitBizDomainAgentTplRel 初始化业务域与agent模板的关联关系
// 如果关联表已有数据，则跳过初始化
func (s *BizDomainSvc) InitBizDomainAgentTplRel(
	ctx context.Context,
	agentTplRepo idbaccess.IDataAgentTplRepo,
	bdAgentTplRelRepo idbaccess.IBizDomainAgentTplRelRepo,
) (err error) {
	bdID := cenum.BizDomainPublic.ToString()

	// 1. 开启事务
	tx, err := bdAgentTplRelRepo.BeginTx(ctx)
	if err != nil {
		return errors.Wrap(err, "begin tx failed")
	}

	defer chelper.TxRollback(tx, &err, s.logger)

	// 2. 查询关联表是否已有数据
	existingRels, err := bdAgentTplRelRepo.GetByBizDomainID(ctx, tx, bdID)
	if err != nil {
		return errors.Wrap(err, "get existing agent tpl rels failed")
	}

	// 如果已有数据，跳过初始化
	if len(existingRels) > 0 {
		s.logger.Infof("[InitBizDomainAgentTplRel] 关联表已有 %d 条数据，跳过初始化", len(existingRels))
		// 回滚事务（因为没有任何修改）
		_ = tx.Rollback()

		return nil
	}

	// 3. 获取所有agent模板ID
	agentTplIDs, err := agentTplRepo.GetAllIDs(ctx)
	if err != nil {
		return errors.Wrap(err, "get all agent tpl ids failed")
	}

	if len(agentTplIDs) == 0 {
		s.logger.Infoln("[InitBizDomainAgentTplRel] 没有agent模板数据，跳过初始化")

		_ = tx.Rollback()

		return nil
	}

	s.logger.Infof("[InitBizDomainAgentTplRel] 准备初始化 %d 个agent模板的业务域关联", len(agentTplIDs))

	// 4. 先写入本地关联表
	currentTs := cutil.GetCurrentMSTimestamp()
	pos := make([]*dapo.BizDomainAgentTplRelPo, 0, len(agentTplIDs))

	for _, agentTplID := range agentTplIDs {
		pos = append(pos, &dapo.BizDomainAgentTplRelPo{
			BizDomainID: bdID,
			AgentTplID:  agentTplID,
			CreatedAt:   currentTs,
		})
	}

	err = bdAgentTplRelRepo.BatchCreate(ctx, tx, pos)
	if err != nil {
		return errors.Wrap(err, "batch create agent tpl rels failed")
	}

	// 5. 调用HTTP接口批量关联（需要将int64转为string）
	agentTplIDStrs := make([]string, 0, len(agentTplIDs))
	for _, id := range agentTplIDs {
		agentTplIDStrs = append(agentTplIDStrs, strconv.FormatInt(id, 10))
	}

	httpReq := bizdomainhttpreq.NewInitAllAgentTplToPublicBusinessDomainReq(agentTplIDStrs)

	err = s.bizDomainHttp.AssociateResourceBatch(ctx, httpReq)
	if err != nil {
		return errors.Wrap(err, "associate resource batch failed")
	}

	// 6. 提交事务
	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "commit tx failed")
	}

	s.logger.Infof("[InitBizDomainAgentTplRel] 成功初始化 %d 个agent模板的业务域关联", len(agentTplIDs))

	return nil
}

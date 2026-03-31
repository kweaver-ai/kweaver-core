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

// FixMissingAgentTplRelResp 修复缺失的agent模板业务域关联响应
type FixMissingAgentTplRelResp struct {
	FixedCount int     `json:"fixed_count"` // 修复的数量
	FixedIDs   []int64 `json:"fixed_ids"`   // 修复的agent模板ID列表
}

// FixMissingAgentTplRel 修复缺失的agent模板业务域关联
// 查找在t_data_agent_config_tpl表但不在t_biz_domain_agent_tpl_rel表中的数据
// 然后为这些数据建立业务域关联
func (s *BizDomainSvc) FixMissingAgentTplRel(
	ctx context.Context,
	agentTplRepo idbaccess.IDataAgentTplRepo,
	bdAgentTplRelRepo idbaccess.IBizDomainAgentTplRelRepo,
) (resp *FixMissingAgentTplRelResp, err error) {
	resp = &FixMissingAgentTplRelResp{
		FixedIDs: make([]int64, 0),
	}

	bdID := cenum.BizDomainPublic.ToString()

	// 1. 获取所有agent模板ID
	allAgentTplIDs, err := agentTplRepo.GetAllIDs(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get all agent tpl ids failed")
	}

	if len(allAgentTplIDs) == 0 {
		s.logger.Infoln("[FixMissingAgentTplRel] 没有agent模板数据，无需修复")
		return resp, nil
	}

	// 2. 开启事务
	tx, err := bdAgentTplRelRepo.BeginTx(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "begin tx failed")
	}

	defer chelper.TxRollback(tx, &err, s.logger)

	// 3. 获取已有关联的agent模板ID
	existingRels, err := bdAgentTplRelRepo.GetByBizDomainID(ctx, tx, bdID)
	if err != nil {
		return nil, errors.Wrap(err, "get existing agent tpl rels failed")
	}

	// 构建已关联的agent模板ID集合
	existingAgentTplIDSet := make(map[int64]struct{}, len(existingRels))
	for _, rel := range existingRels {
		existingAgentTplIDSet[rel.AgentTplID] = struct{}{}
	}

	// 4. 找出缺失关联的agent模板ID
	missingAgentTplIDs := make([]int64, 0)

	for _, agentTplID := range allAgentTplIDs {
		if _, exists := existingAgentTplIDSet[agentTplID]; !exists {
			missingAgentTplIDs = append(missingAgentTplIDs, agentTplID)
		}
	}

	if len(missingAgentTplIDs) == 0 {
		s.logger.Infoln("[FixMissingAgentTplRel] 没有缺失业务域关联的agent模板，无需修复")

		_ = tx.Rollback()

		return resp, nil
	}

	s.logger.Infof("[FixMissingAgentTplRel] 发现 %d 个缺失业务域关联的agent模板，准备修复", len(missingAgentTplIDs))

	// 5. 写入本地关联表
	currentTs := cutil.GetCurrentMSTimestamp()
	pos := make([]*dapo.BizDomainAgentTplRelPo, 0, len(missingAgentTplIDs))

	for _, agentTplID := range missingAgentTplIDs {
		pos = append(pos, &dapo.BizDomainAgentTplRelPo{
			BizDomainID: bdID,
			AgentTplID:  agentTplID,
			CreatedAt:   currentTs,
		})
	}

	err = bdAgentTplRelRepo.BatchCreate(ctx, tx, pos)
	if err != nil {
		return nil, errors.Wrap(err, "batch create agent tpl rels failed")
	}

	// 6. 调用HTTP接口批量关联（需要将int64转为string）
	agentTplIDStrs := make([]string, 0, len(missingAgentTplIDs))
	for _, id := range missingAgentTplIDs {
		agentTplIDStrs = append(agentTplIDStrs, strconv.FormatInt(id, 10))
	}

	httpReq := bizdomainhttpreq.NewInitAllAgentTplToPublicBusinessDomainReq(agentTplIDStrs)

	err = s.bizDomainHttp.AssociateResourceBatch(ctx, httpReq)
	if err != nil {
		return nil, errors.Wrap(err, "associate resource batch failed")
	}

	// 7. 提交事务
	err = tx.Commit()
	if err != nil {
		return nil, errors.Wrap(err, "commit tx failed")
	}

	resp.FixedCount = len(missingAgentTplIDs)
	resp.FixedIDs = missingAgentTplIDs

	s.logger.Infof("[FixMissingAgentTplRel] 成功修复 %d 个agent模板的业务域关联", len(missingAgentTplIDs))

	return resp, nil
}

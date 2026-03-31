package v3agentconfigsvc

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/bizdomainhttp/bizdomainhttpreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/auditlogdto"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/apierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/global"
)

// Delete 通过id删除数据智能体配置
//
// 1. 检查数据智能体配置是否存在
// 2.1 如果是发布状态，抛出409错误
// 2.2 如果不是私有接口，检查访问者是否是创建者，如果不是，抛出403错误
//
// 3. PO转EO（如果需要发送审计日志）
// 4. 调用repo层删除数据
// 5. 触发向量索引
// 6. 发送审计日志
func (s *dataAgentConfigSvc) Delete(ctx context.Context, id, uid string, isPrivate bool) (auditLogInfo auditlogdto.AgentDeleteAuditLogInfo, err error) {
	// 1. 检查是否存在
	po, err := s.agentConfRepo.GetByID(ctx, id)
	if err != nil {
		if chelper.IsSqlNotFound(err) {
			err = capierr.NewCustom404Err(ctx, apierr.DataAgentConfigNotFound, "数据智能体配置不存在")
		}

		return
	}

	auditLogInfo = auditlogdto.AgentDeleteAuditLogInfo{
		ID:   id,
		Name: po.Name,
	}

	// 2. 删除检查
	// 2.1 检查是否是发布状态
	if po.Status == cdaenum.StatusPublished {
		err = capierr.NewCustom409Err(ctx, apierr.DataAgentConfigPublishedCannotBeDeleted, "数据智能体配置已发布，无法删除。可以取消发布后再删除")
		return
	}

	// 2.2 检查访问者是否是创建者
	if !isPrivate {
		if po.CreatedBy != uid {
			err = capierr.NewCustom403Err(ctx, apierr.DataAgentConfigForbiddenNotOwner, "访问者不是创建者，无法删除")
			return
		}

		if po.IsBuiltIn.IsBuiltIn() {
			err = capierr.New403Err(ctx, "内置数据智能体不可删除")
			return
		}
	}

	// 注释掉审计日志相关代码
	// 3. PO转EO（如果需要发送审计日志）
	// origEo, err := daconfp2e.DataAgent(origPo)
	// if err != nil {
	// 	return
	// }

	// 4. 开始事务
	tx, err := s.agentConfRepo.BeginTx(ctx)
	if err != nil {
		return
	}

	defer chelper.TxRollbackOrCommit(tx, &err, s.logger)

	// 5. 调用repo层删除数据
	err = s.agentConfRepo.Delete(ctx, tx, id)
	if err != nil {
		return
	}

	if global.GConfig.IsBizDomainDisabled() {
		return
	}

	// 7. 删除空间下资源的关联关系
	//err = s.spaceResourceRepo.DeleteByAgentID(ctx, tx, id)
	//if err != nil {
	//	return
	//}

	// 8. 解除业务域关联（先删除本地关联表，再调用HTTP）
	bdID := chelper.GetBizDomainIDFromCtx(ctx)

	// 8.1 删除本地关联表
	err = s.bdAgentRelRepo.DeleteByAgentID(ctx, tx, id)
	if err != nil {
		return
	}

	// 8.2 调用HTTP接口解除关联
	err = s.bizDomainHttp.DisassociateResource(ctx, &bizdomainhttpreq.DisassociateResourceReq{
		ID:   id,
		BdID: bdID,
		Type: cdaenum.ResourceTypeDataAgent,
	})
	if err != nil {
		return
	}

	// 9. 发送审计日志
	// err = s.sendAuditLog(ctx, origEo, persrecenums.MngLogOpTypeDelete, tx)

	return
}

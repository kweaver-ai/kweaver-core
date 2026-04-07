package releasesvc

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/auditlogdto"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/apierr"
	"github.com/pkg/errors"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
)

// UnPublish implements iv3portdriver.IReleaseSvc.
func (svc *releaseSvc) UnPublish(ctx context.Context, agentID string) (auditloginfo auditlogdto.AgentUnPublishAuditLogInfo, err error) {
	agentCfgPo, err := svc.agentConfigRepo.GetByID(ctx, agentID)
	if err != nil {
		if chelper.IsSqlNotFound(err) {
			err = capierr.NewCustom404Err(ctx, apierr.DataAgentConfigNotFound, "agent not found")
			return
		}

		err = errors.Wrapf(err, "[releaseSvc.UnPublish]: get agent config by id failed")

		return
	}

	auditloginfo = auditlogdto.AgentUnPublishAuditLogInfo{
		ID:   agentCfgPo.ID,
		Name: agentCfgPo.Name,
	}

	// 检查取消发布权限
	hasPms, err := svc.isHasUnPublishPermission(ctx, agentCfgPo)
	if err != nil {
		err = errors.Wrapf(err, "check unpublish permission failed")
		return
	}

	if !hasPms {
		err = capierr.NewCustom403Err(ctx, apierr.AgentFactoryPermissionForbidden, "do not have unpublish permission")
		return
	}

	userID := chelper.GetUserIDFromCtx(ctx)
	if agentCfgPo.CreatedBy != userID {
		var b bool

		b, err = svc.isHasUnpublishOtherUserAgentPermission(ctx)
		if err != nil {
			return
		}

		if !b {
			err = capierr.NewCustom403Err(ctx, apierr.AgentFactoryPermissionForbidden, "无取消发布的权限：非创建人且无取消发布别人已发布Agent的权限")
			return
		}
		// return
	}

	// 1. 检查Agent是否存在
	releasePo, err := svc.releaseRepo.GetByAgentID(ctx, agentID)
	if err != nil {
		return auditloginfo, errors.Wrapf(err, "[releaseSvc.UnPublish]: get release by agent id failed")
	}

	if releasePo == nil {
		return
	}

	// 2. 开启事务
	tx, err := svc.releaseRepo.BeginTx(ctx)
	if err != nil {
		err = errors.Wrapf(err, "begin transaction failed")
		return
	}

	defer chelper.TxRollback(tx, &err, svc.Logger)

	// 3. 删除发布记录
	err = svc.releaseRepo.DeleteByAgentID(ctx, tx, agentID)
	if err != nil {
		err = errors.Wrapf(err, "delete release by agent id failed")
		return
	}

	// 4. 删除分类绑定关系
	err = svc.releaseCategoryRelRepo.DelByReleaseID(ctx, tx, releasePo.ID)
	if err != nil {
		err = errors.Wrapf(err, "delete release category rel by release id failed")
		return
	}

	// 5. 删除权限控制关联关系
	err = svc.releasePermissionRepo.DelByReleaseID(ctx, tx, releasePo.ID)
	if err != nil {
		err = errors.Wrapf(err, "delete release permission by release id failed")
		return
	}

	// 6. 更新Agent状态
	err = svc.agentConfigRepo.UpdateStatus(ctx, tx, cdaenum.StatusUnpublished, agentID, "")
	if err != nil {
		err = errors.Wrapf(err, "update agent status to unpublished failed")
		return
	}

	// 7. 删除空间资源关联关系
	//err = svc.spaceResourceRepo.DeleteByAgentID(ctx, tx, agentID)
	//if err != nil {
	//	err = errors.Wrapf(err, "delete space resource by agent id failed")
	//	return
	//}

	// 9. 从“权限平台”删除Agent使用权限
	err = svc.removeUsePmsByHTTPAcc(ctx, agentID)
	if err != nil {
		err = errors.Wrapf(err, "remove use pms failed")
		return
	}

	// 10. 提交事务
	err = tx.Commit()
	if err != nil {
		err = errors.Wrapf(err, "commit tx failed")
		return
	}

	return
}

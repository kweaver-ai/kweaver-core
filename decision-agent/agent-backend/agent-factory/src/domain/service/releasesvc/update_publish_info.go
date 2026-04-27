package releasesvc

import (
	"context"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/auditlogdto"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/release/releasereq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/release/releaseresp"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/apierr"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/pkg/errors"
)

// UpdatePublishInfo 更新发布信息
func (svc *releaseSvc) UpdatePublishInfo(ctx context.Context, agentID string, req *releasereq.UpdatePublishInfoReq) (resp *releaseresp.PublishUpsertResp,
	auditloginfo auditlogdto.AgentModifyPublishAuditLogInfo, err error,
) {
	defer func() {
		if err != nil {
			resp = &releaseresp.PublishUpsertResp{}
		}
	}()

	// 1. 检查Agent是否存在
	var po *dapo.DataAgentPo

	po, err = svc.agentConfigRepo.GetByID(ctx, agentID)
	if err != nil {
		if chelper.IsSqlNotFound(err) {
			err = capierr.NewCustom404Err(ctx, apierr.DataAgentConfigNotFound, "agent not found")
			return
		}

		return
	}

	auditloginfo = auditlogdto.AgentModifyPublishAuditLogInfo{
		ID:   po.ID,
		Name: po.Name,
	}

	// 2. 检查发布权限
	var hasPms bool

	hasPms, err = svc.isHasPublishPermission(ctx, po)
	if err != nil {
		return
	}

	if !hasPms {
		err = capierr.NewCustom403Err(ctx, apierr.AgentFactoryPermissionForbidden, "do not have publish permission")
		return
	}

	// 3. 开始事务
	tx, err := svc.releaseRepo.BeginTx(ctx)
	if err != nil {
		err = errors.Wrapf(err, "begin transaction failed")
		return
	}
	defer chelper.TxRollback(tx, &err, svc.Logger)

	// 4. 获取发布记录
	releasePo, err := svc.releaseRepo.GetByAgentID(ctx, agentID)
	if err != nil {
		err = errors.Wrapf(err, "get release by agent id failed, agentID: %s", agentID)
		return
	}

	if releasePo == nil {
		err = capierr.NewCustom404Err(ctx, apierr.ReleaseNotFound, "the release of agent not found")
		return
	}

	// 4. 更新发布信息
	// 4.1 更新发布描述
	releasePo.AgentDesc = req.Description

	// 4.2 设置发布为标识&发布到标识
	releasePo.SetPublishToBes(req.PublishToBes)
	releasePo.SetPublishToWhere(req.PublishToWhere)

	// 4.3 设置是否进行“使用权限”的管控
	releasePo.SetIsPmsCtrl(req.PmsControl != nil)

	//// 4.4 设置更新时间和更新人
	// releasePo.UpdateTime = cutil.GetCurrentMSTimestamp()
	//releasePo.UpdateBy = chelper.GetUserIDFromCtx(ctx)

	// 5. 更新发布记录
	err = svc.releaseRepo.Update(ctx, tx, releasePo)
	if err != nil {
		err = errors.Wrapf(err, "update release failed")
		return
	}

	// 6. 返回响应
	resp = &releaseresp.PublishUpsertResp{
		ReleaseId:   releasePo.ID,
		Version:     releasePo.AgentVersion,
		PublishedAt: releasePo.UpdateTime,
		PublishedBy: releasePo.UpdateBy,
	}

	err = resp.FillPublishedByName(ctx, svc.umHttp)
	if err != nil {
		err = errors.Wrapf(err, "fill published by name failed")
		return
	}

	// 7. 更新分类关联
	err = svc.handleCategory(ctx, req.CategoryIDs, releasePo.ID, tx)
	if err != nil {
		err = errors.Wrapf(err, "handle category failed")
		return
	}

	// 8. 更新权限控制
	err = svc.handlePmsCtrl(ctx, req.PmsControl, releasePo.ID, releasePo.AgentID, tx)
	if err != nil {
		err = errors.Wrapf(err, "handle pms ctrl failed")
		return
	}

	// 9. 当req.PublishToWhere不包含"custom_space"时，删除自定义空间关联
	//if !slices.Contains(req.PublishToWhere, daenum.PublishToWhereCustomSpace) {
	//	err = svc.spaceResourceRepo.DeleteByAgentID(ctx, tx, agentID)
	//	if err != nil {
	//		err = errors.Wrapf(err, "delete custom space relations failed")
	//		return
	//	}
	//}

	// 10. 提交事务
	err = tx.Commit()
	if err != nil {
		err = errors.Wrapf(err, "commit transaction failed")
		return
	}

	return
}

package tplsvc

import (
	"context"
	"strconv"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/auditlogdto"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/apierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/pkg/errors"
)

func (s *dataAgentTplSvc) Unpublish(ctx context.Context, tplID int64) (auditloginfo auditlogdto.AgentTemplateUnpublishAuditLogInfo, err error) {
	var hasPms bool

	hasPms, err = s.isHasUnPublishPermission(ctx)
	if err != nil {
		return
	}

	if !hasPms {
		err = capierr.NewCustom403Err(ctx, apierr.AgentFactoryPermissionForbidden, "do not have unpublish permission")
		return
	}

	// 1. 获取模板信息进行权限检查
	po, err := s.agentTplRepo.GetByID(ctx, tplID)
	if err != nil {
		if chelper.IsSqlNotFound(err) {
			err = capierr.NewCustom404Err(ctx, apierr.AgentTplNotFound, "模板不存在")
			return
		}

		err = errors.Wrapf(err, "[Unpublish]: get template by id %d", tplID)

		return
	}

	auditloginfo = auditlogdto.AgentTemplateUnpublishAuditLogInfo{
		ID:   strconv.FormatInt(tplID, 10),
		Name: po.Name,
	}

	// 2. 获取已发布模板信息
	pubedPo, err := s.publishedTplRepo.GetByTplID(ctx, tplID)
	if err != nil {
		if chelper.IsSqlNotFound(err) {
			err = capierr.NewCustom404Err(ctx, apierr.PublishedTplNotFound, "此已发布模板不存在")
			return
		}

		err = errors.Wrapf(err, "[Unpublish]: get template by id %d", tplID)

		return
	}

	// if po.Status != cdaenum.StatusPublished {
	//	err = capierr.NewCustom409Err(ctx, apierr.AgentTplIsUnpublished, "模板未发布")
	//	return
	//}

	// 3. 权限检查
	userID := chelper.GetUserIDFromCtx(ctx)

	if po.CreatedBy != userID {
		var b bool

		b, err = s.isHasUnpublishOtherUserAgentTplPermission(ctx)
		if err != nil {
			return
		}

		if !b {
			err = capierr.NewCustom403Err(ctx, apierr.AgentFactoryPermissionForbidden, "无取消发布的权限：非创建人且无取消发布别人已发布模板的权限")
			return
		}

		return
	}

	// 4. 开启事务
	tx, err := s.agentTplRepo.BeginTx(ctx)
	if err != nil {
		err = errors.Wrapf(err, "[Unpublish]: begin transaction")
		return
	}

	defer chelper.TxRollbackOrCommit(tx, &err, s.Logger)

	// 5. 删除分类关联
	err = s.publishedTplRepo.DelCategoryAssocByTplID(ctx, tx, tplID)
	if err != nil {
		err = errors.Wrapf(err, "[Unpublish]: delete category assoc by tpl id %d", tplID)
		return
	}

	// 6. 更新状态为未发布

	err = s.agentTplRepo.UpdateStatus(ctx, tx, cdaenum.StatusUnpublished, tplID, "", 0)
	if err != nil {
		err = errors.Wrapf(err, "[Unpublish]: update template status to unpublished")
		return
	}

	// 7. 删除已发布模板
	err = s.publishedTplRepo.Delete(ctx, tx, pubedPo.ID)
	if err != nil {
		err = errors.Wrapf(err, "[Unpublish]: delete published template")
		return
	}

	return
}

package tplsvc

import (
	"context"
	"strconv"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/auditlogdto"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent_tpl/agenttplreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent_tpl/agenttplresp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/apierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/pkg/errors"
)

func (s *dataAgentTplSvc) UpdatePublishInfo(ctx context.Context, req *agenttplreq.UpdatePublishInfoReq,
	tplID int64,
) (resp *agenttplresp.PublishUpsertResp, auditloginfo auditlogdto.AgentTemplateModifyPublishAuditLogInfo, err error) {
	var hasPms bool

	hasPms, err = s.isHasPublishPermission(ctx)
	if err != nil {
		return
	}

	if !hasPms {
		err = capierr.New403Err(ctx, "do not have publish permission")
		return
	}

	// 1. 获取模板信息进行权限检查
	po, err := s.agentTplRepo.GetByID(ctx, tplID)
	if err != nil {
		if chelper.IsSqlNotFound(err) {
			err = capierr.NewCustom404Err(ctx, apierr.AgentTplNotFound, "模板不存在")
			return
		}

		err = errors.Wrapf(err, "[UpdatePublishInfo]: get template by id %d", tplID)

		return
	}

	// 2. 获取已发布模板信息
	pubedPo, err := s.publishedTplRepo.GetByTplID(ctx, tplID)
	if err != nil {
		if chelper.IsSqlNotFound(err) {
			err = capierr.NewCustom404Err(ctx, apierr.PublishedTplNotFound, "此已发布模板不存在")
			return
		}

		err = errors.Wrapf(err, "[UpdatePublishInfo]: get published template by tpl id %d", tplID)

		return
	}

	auditloginfo = auditlogdto.AgentTemplateModifyPublishAuditLogInfo{
		ID:   strconv.FormatInt(tplID, 10),
		Name: po.Name,
	}

	publishedTplID := pubedPo.ID

	// 3. 权限检查
	userID := chelper.GetUserIDFromCtx(ctx)
	if po.CreatedBy != userID {
		err = capierr.NewCustom403Err(ctx, apierr.AgentTplForbiddenNotOwner, "无权限发布，非创建人")
		return
	}

	// 4. 开启事务
	tx, err := s.agentTplRepo.BeginTx(ctx)
	if err != nil {
		err = errors.Wrapf(err, "begin transaction")
		return
	}

	defer chelper.TxRollbackOrCommit(tx, &err, s.Logger)

	// 5. 处理分类关联
	err = s.handleCategory(ctx, req.CategoryIDs, publishedTplID, tx)
	if err != nil {
		err = errors.Wrapf(err, "handle category")
		return
	}

	// 6. 设置响应
	resp = &agenttplresp.PublishUpsertResp{
		AgentTplId:  po.ID,
		PublishedAt: po.GetPublishedAtInt64(),
		PublishedBy: po.GetPublishedByString(),
	}

	err = resp.FillPublishedByName(ctx, s.umHttp)
	if err != nil {
		err = errors.Wrapf(err, "fill published by name failed")
		return
	}

	return
}

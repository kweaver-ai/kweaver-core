package tplsvc

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/auditlogdto"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent_tpl/agenttplreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent_tpl/agenttplresp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/apierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/pkg/errors"
)

func (s *dataAgentTplSvc) Publish(ctx context.Context, tx *sql.Tx, req *agenttplreq.PublishReq, id int64,
	isFromCopy2TplAndPublish bool,
) (resp *agenttplresp.PublishUpsertResp, auditloginfo auditlogdto.AgentTemplatePublishAuditLogInfo, err error) {
	if !isFromCopy2TplAndPublish {
		var hasPms bool

		hasPms, err = s.isHasPublishPermission(ctx)
		if err != nil {
			return
		}

		if !hasPms {
			err = capierr.NewCustom403Err(ctx, apierr.AgentFactoryPermissionForbidden, "do not have publish permission")
			return
		}
	}

	// 3. 开启事务
	if tx == nil {
		tx, err = s.agentTplRepo.BeginTx(ctx)
		if err != nil {
			err = errors.Wrapf(err, "begin transaction")
			return
		}

		defer chelper.TxRollbackOrCommit(tx, &err, s.Logger)
	}

	// 1. 获取模板信息进行权限检查
	po, err := s.agentTplRepo.GetByIDWithTx(ctx, tx, id)
	if err != nil {
		if chelper.IsSqlNotFound(err) {
			err = capierr.NewCustom404Err(ctx, apierr.AgentTplNotFound, "模板不存在")
			return
		}

		err = errors.Wrapf(err, "get template by id")

		return
	}

	auditloginfo = auditlogdto.AgentTemplatePublishAuditLogInfo{
		ID:   strconv.FormatInt(id, 10),
		Name: po.Name,
	}

	// 2. 权限检查
	userID := chelper.GetUserIDFromCtx(ctx)

	if po.CreatedBy != userID {
		err = capierr.NewCustom403Err(ctx, apierr.AgentTplForbiddenNotOwner, "无权限发布，非创建人")
		return
	}

	// 3. 生成已发布模板信息

	uid := chelper.GetUserIDFromCtx(ctx)
	currentTs := cutil.GetCurrentMSTimestamp()
	publishedAt := currentTs
	publishedBy := uid

	var publishedID int64

	publishedID, err = s.genPublishedPo(ctx, tx, po, publishedAt, publishedBy)
	if err != nil {
		err = errors.Wrapf(err, "[Publish]: gen published po")
		return
	}

	// 4. 处理分类关联
	err = s.handleCategory(ctx, req.CategoryIDs, publishedID, tx)
	if err != nil {
		err = errors.Wrapf(err, "[Publish]: handle category")
		return
	}

	// 5. 更新状态为已发布

	err = s.agentTplRepo.UpdateStatus(ctx, tx, cdaenum.StatusPublished, id, publishedBy, publishedAt)
	if err != nil {
		err = errors.Wrapf(err, "[Publish]: update template status to published")
		return
	}

	// 6. 设置响应
	resp = &agenttplresp.PublishUpsertResp{
		AgentTplId:  publishedID,
		PublishedAt: publishedAt,
		PublishedBy: publishedBy,
	}

	err = resp.FillPublishedByName(ctx, s.umHttp)
	if err != nil {
		err = errors.Wrapf(err, "[Publish]: fill published by name failed")
		return
	}

	return
}

func (s *dataAgentTplSvc) handleCategory(ctx context.Context, categoryIDs []string, pubedTplID int64, tx *sql.Tx) (err error) {
	if len(categoryIDs) > 0 {
		// 1. 验证分类是否存在
		var categoryNameMap map[string]string

		categoryNameMap, err = s.categoryRepo.GetIDNameMap(ctx, categoryIDs)
		if err != nil {
			err = errors.Wrapf(err, "[handleCategory]: get category name map failed")
			return
		}

		for _, categoryID := range categoryIDs {
			if categoryNameMap[categoryID] == "" {
				detail := fmt.Sprintf("分类[%s]不存在", categoryID)
				err = capierr.NewCustom404Err(ctx, apierr.CategoryNotFound, detail)

				return
			}
		}

		// 2. 先删除现有的分类关联
		err = s.publishedTplRepo.DelCategoryAssocByTplID(ctx, tx, pubedTplID)
		if err != nil {
			err = errors.Wrapf(err, "[handleCategory]: delete category relations failed")
			return
		}

		// 3. 添加新的分类关联
		categoryRels := make([]*dapo.PubTplCatAssocPo, 0)

		for _, categoryID := range categoryIDs {
			categoryID = strings.TrimSpace(categoryID)
			if categoryID != "" {
				categoryRel := &dapo.PubTplCatAssocPo{
					PublishedTplID: pubedTplID,
					CategoryID:     categoryID,
				}
				categoryRels = append(categoryRels, categoryRel)
			}
		}

		// 4. 批量创建分类关联
		err = s.publishedTplRepo.BatchCreateCategoryAssoc(ctx, tx, categoryRels)
		if err != nil {
			err = errors.Wrapf(err, "[handleCategory]: batch create category relations failed")
			return
		}
	}

	return
}

func (s *dataAgentTplSvc) genPublishedPo(ctx context.Context, tx *sql.Tx, po *dapo.DataAgentTplPo, publishedAt int64, publishedBy string) (id int64, err error) {
	publishedPo := &dapo.PublishedTplPo{}

	err = cutil.CopyStructUseJSON(publishedPo, po)
	if err != nil {
		err = errors.Wrapf(err, "[genPublishedPo]: copy struct use json")
		return
	}

	publishedPo.ID = 0
	publishedPo.TplID = po.ID

	// 设置发布时间
	publishedPo.PublishedAt = publishedAt
	publishedPo.PublishedBy = publishedBy

	// 1. 删除已发布模板
	err = s.publishedTplRepo.DeleteByTplID(ctx, tx, po.ID)
	if err != nil {
		err = errors.Wrapf(err, "[genPublishedPo]: delete published tpl failed")
		return
	}

	// 2. 创建已发布模板
	id, err = s.publishedTplRepo.Create(ctx, tx, publishedPo)
	if err != nil {
		err = errors.Wrapf(err, "[genPublishedPo]: create published tpl failed")
		return
	}

	return
}

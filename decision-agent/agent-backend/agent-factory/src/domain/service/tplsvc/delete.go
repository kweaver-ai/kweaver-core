package tplsvc

import (
	"context"
	"fmt"
	"strconv"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/httpaccess/bizdomainhttp/bizdomainhttpreq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/auditlogdto"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/apierr"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/global"
	"github.com/pkg/errors"
)

func (s *dataAgentTplSvc) Delete(ctx context.Context, id int64, uid string, isPrivate bool) (auditloginfo auditlogdto.AgentTemplateDeleteAuditLogInfo, err error) {
	// 1. 参数验证
	if id == 0 {
		err = capierr.NewCustom400Err(ctx, apierr.AgentFactory_InvalidParameter_RequestBody, "模板ID不能为空")
		return
	}

	// 2. 检查模板是否存在
	exists, err := s.agentTplRepo.ExistsByID(ctx, id)
	if err != nil {
		err = errors.Wrapf(err, "check template exists")
		return
	}

	if !exists {
		err = capierr.NewCustom404Err(ctx, apierr.AgentTplNotFound, "模板不存在")
		return
	}

	// 3. 获取模板信息进行权限检查
	po, err := s.agentTplRepo.GetByID(ctx, id)
	if err != nil {
		err = errors.Wrapf(err, "get template by id")
		return
	}

	auditloginfo = auditlogdto.AgentTemplateDeleteAuditLogInfo{
		ID:   strconv.FormatInt(id, 10),
		Name: po.Name,
	}
	// 3.1 检查是否是发布状态
	if po.Status == cdaenum.StatusPublished {
		err = capierr.NewCustom409Err(ctx, apierr.AgentTplPublishedCannotBeDeleted, "模板已发布，无法删除。可以取消发布后再删除")
		return
	}

	// 4. 权限检查
	if !isPrivate {
		if po.CreatedBy != uid {
			err = capierr.NewCustom403Err(ctx, apierr.AgentTplForbiddenNotOwner, "无权限删除，非创建人")
			return
		}

		if po.IsBuiltIn.IsBuiltIn() {
			err = capierr.New403Err(ctx, "内置数据智能体模板不可删除")
			return
		}
	}

	// 5. 开启事务
	tx, err := s.agentTplRepo.BeginTx(ctx)
	if err != nil {
		err = errors.Wrapf(err, "begin transaction")
		return
	}

	defer chelper.TxRollbackOrCommit(tx, &err, s.Logger)

	// 6. 执行软删除
	err = s.agentTplRepo.Delete(ctx, tx, id)
	if err != nil {
		err = errors.Wrapf(err, "delete template")
		return
	}

	// 7. 删除已发布模板
	err = s.publishedTplRepo.DeleteByTplID(ctx, tx, id)
	if err != nil {
		err = errors.Wrapf(err, "delete published template")
		return
	}

	if global.GConfig.IsBizDomainDisabled() {
		return
	}

	// 8. 解除业务域关联（先删除本地关联表，再调用HTTP）
	bdID := chelper.GetBizDomainIDFromCtx(ctx)

	// 8.1 删除本地关联表
	err = s.bdAgentTplRelRepo.DeleteByAgentTplID(ctx, tx, id)
	if err != nil {
		return
	}

	// 8.2 调用HTTP接口解除关联
	err = s.bizDomainHttp.DisassociateResource(ctx, &bizdomainhttpreq.DisassociateResourceReq{
		ID:   fmt.Sprintf("%d", id),
		BdID: bdID,
		Type: cdaenum.ResourceTypeDataAgentTpl,
	})
	if err != nil {
		return
	}

	return
}

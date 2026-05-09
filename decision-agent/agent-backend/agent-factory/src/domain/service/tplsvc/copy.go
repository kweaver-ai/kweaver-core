package tplsvc

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/daenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/httpaccess/bizdomainhttp/bizdomainhttpreq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/auditlogdto"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/agent_tpl/agenttplresp"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/global"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/pkg/errors"
)

func (s *dataAgentTplSvc) Copy(ctx context.Context, id int64) (res *agenttplresp.CopyResp, auditLogInfo auditlogdto.AgentTemplateCopyAuditLogInfo, err error) {
	// 1. 获取源模板
	sourcePo, err := s.agentTplRepo.GetByID(ctx, id)
	if err != nil {
		// 检查是否是记录不存在的错误
		if chelper.IsSqlNotFound(err) {
			err = capierr.New404Err(ctx, "源模板不存在")
		}

		return
	}

	auditLogInfo = auditlogdto.AgentTemplateCopyAuditLogInfo{
		ID:   strconv.FormatInt(id, 10),
		Name: sourcePo.Name,
	}

	// 2. 生成模板名称（如果没有提供）
	templateName := sourcePo.Name + "_副本"

	// // 3. 检查新模板名称是否已存在
	// exists, err := s.agentTplRepo.ExistsByName(ctx, templateName)
	// if err != nil {
	// 	err = errors.Wrapf(err, "check template name exists")
	// 	return
	// }

	// if exists {
	// 	err = capierr.NewCustom409Err(ctx, apierr.DataAgentConfigNameExists, "模板名称已存在")
	// 	return
	// }

	// 3. 开启事务
	tx, err := s.agentTplRepo.BeginTx(ctx)
	if err != nil {
		err = errors.Wrapf(err, "begin transaction")
		return
	}
	defer chelper.TxRollbackOrCommit(tx, &err, s.logger)

	// 4. 创建新的PO（复制配置）
	newPo := &dapo.DataAgentTplPo{}

	// 5. 设置新模板的基本信息
	id, err = s.copyPo(ctx, tx, newPo, sourcePo, templateName)
	if err != nil {
		err = errors.Wrapf(err, "copy po")
		return
	}

	if global.GConfig.IsBizDomainDisabled() {
		res = &agenttplresp.CopyResp{
			ID:   id,
			Name: templateName,
			Key:  newPo.Key,
		}

		return
	}

	// 6. 关联业务域（先写入本地关联表，再调用HTTP）
	bdID := chelper.GetBizDomainIDFromCtx(ctx)

	// 6.1 写入本地关联表
	bdRelPo := &dapo.BizDomainAgentTplRelPo{
		BizDomainID: bdID,
		AgentTplID:  id,
		CreatedAt:   cutil.GetCurrentMSTimestamp(),
	}

	err = s.bdAgentTplRelRepo.BatchCreate(ctx, tx, []*dapo.BizDomainAgentTplRelPo{bdRelPo})
	if err != nil {
		return
	}

	// 6.2 调用HTTP接口关联
	err = s.bizDomainHttp.AssociateResource(ctx, &bizdomainhttpreq.AssociateResourceReq{
		ID:   fmt.Sprintf("%d", id),
		BdID: bdID,
		Type: cdaenum.ResourceTypeDataAgentTpl,
	})
	if err != nil {
		return
	}

	// 7. 返回结果
	res = &agenttplresp.CopyResp{
		ID:   id,
		Name: templateName,
		Key:  newPo.Key,
	}

	return
}

func (s *dataAgentTplSvc) copyPo(ctx context.Context, tx *sql.Tx, newPo *dapo.DataAgentTplPo, sourcePo *dapo.DataAgentTplPo, templateName string) (id int64, err error) {
	err = cutil.CopyStructUseJSON(newPo, sourcePo)
	if err != nil {
		return
	}

	// 设置新模板的基本信息
	newPo.ID = 0
	newPo.Name = templateName

	// 生成Key（如果没有提供）
	newPo.Key = cutil.UlidMake()

	// 设置分类ID
	// newPo.CategoryID = ""

	// 设置状态为未发布
	newPo.Status = cdaenum.StatusUnpublished

	// 设置是否内置
	newPo.SetIsBuiltIn(cdaenum.BuiltInNo)

	// 设置时间戳
	currentTs := cutil.GetCurrentMSTimestamp()
	newPo.CreatedAt = currentTs
	newPo.UpdatedAt = currentTs

	// 设置创建者
	userID := chelper.GetUserIDFromCtx(ctx)
	newPo.CreatedBy = userID
	newPo.UpdatedBy = userID

	newPo.CreatedType = daenum.AgentTplCreatedTypeCopyFromTpl

	// 清空发布时间和发布人
	newPo.SetPublishedAt(0)
	newPo.SetPublishedBy("")

	// 保存到数据库
	err = s.agentTplRepo.Create(ctx, tx, newPo)
	if err != nil {
		return
	}

	// get po by key
	po, err := s.agentTplRepo.GetByKeyWithTx(ctx, tx, newPo.Key)
	if err != nil {
		return
	}

	id = po.ID

	return
}

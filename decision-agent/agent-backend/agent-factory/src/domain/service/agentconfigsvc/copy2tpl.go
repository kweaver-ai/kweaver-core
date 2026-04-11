package v3agentconfigsvc

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/daenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/httpaccess/bizdomainhttp/bizdomainhttpreq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/auditlogdto"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/agent_config/agentconfigreq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/agent_config/agentconfigresp"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/apierr"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cconstant"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/global"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/util"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/pkg/errors"
)

// Copy2Tpl 复制Agent为模板
func (s *dataAgentConfigSvc) Copy2Tpl(ctx context.Context, agentID string, req *agentconfigreq.Copy2TplReq, tx *sql.Tx) (res *agentconfigresp.Copy2TplResp, auditLogInfo auditlogdto.AgentCopy2TplAuditLogInfo, err error) {
	// 1. 获取源Agent
	sourcePo, err := s.getAgentPoForCopy(ctx, agentID)
	if err != nil {
		return
	}

	// 1.1 检查owner或内置Agent管理权限
	err = s.isOwnerOrHasBuiltInAgentMgmtPermission(ctx, sourcePo, chelper.GetUserIDFromCtx(ctx))
	if err != nil {
		return
	}

	auditLogInfo = auditlogdto.AgentCopy2TplAuditLogInfo{
		ID:   agentID,
		Name: sourcePo.Name,
	}

	// 2. 生成新模板名称
	newName, err := s.getNewNameForAgentCopy2Tpl(ctx, req, sourcePo)
	if err != nil {
		return
	}

	// 3. 开启事务
	if tx == nil {
		tx, err = s.agentTplRepo.BeginTx(ctx)
		if err != nil {
			err = errors.Wrapf(err, "开启事务失败")
			return
		}

		// 由tx创建者负责提交事务
		defer chelper.TxRollbackOrCommit(tx, &err, s.logger)
	}

	// 4. 调用模板服务创建模板
	res, err = s.createTemplateFromAgent(ctx, sourcePo, newName, tx)
	if err != nil {
		err = errors.Wrapf(err, "[dataAgentConfigSvc][Copy2Tpl]创建模板失败")
		return
	}

	if global.GConfig.IsBizDomainDisabled() {
		return
	}

	// 5. 关联业务域（先写入本地关联表，再调用HTTP）
	bdID := chelper.GetBizDomainIDFromCtx(ctx)

	// 5.1 写入本地关联表
	bdRelPo := &dapo.BizDomainAgentTplRelPo{
		BizDomainID: bdID,
		AgentTplID:  res.ID,
		CreatedAt:   cutil.GetCurrentMSTimestamp(),
	}

	err = s.bdAgentTplRelRepo.BatchCreate(ctx, tx, []*dapo.BizDomainAgentTplRelPo{bdRelPo})
	if err != nil {
		return
	}

	// 5.2 调用HTTP接口关联
	bdReq := &bizdomainhttpreq.AssociateResourceReq{
		ID:   fmt.Sprintf("%d", res.ID),
		BdID: bdID,
		Type: cdaenum.ResourceTypeDataAgentTpl,
	}

	err = s.bizDomainHttp.AssociateResource(ctx, bdReq)
	if err != nil {
		return
	}

	return
}

// createTemplateFromAgent 从Agent创建模板
func (s *dataAgentConfigSvc) createTemplateFromAgent(ctx context.Context, sourcePo *dapo.DataAgentPo, templateName string, tx *sql.Tx) (res *agentconfigresp.Copy2TplResp, err error) {
	// 1. 生成模板ID和Key
	// templateID := cutil.UlidMake()
	templateKey := cutil.UlidMake()

	// 2. 创建模板PO
	newPo := &dapo.DataAgentTplPo{}
	tmpNewPo := &dapo.DataAgentTplIDStrPo{}

	err = cutil.CopyStructUseJSON(tmpNewPo, sourcePo)
	if err != nil {
		return nil, errors.Wrapf(err, "复制Agent PO数据失败")
	}

	newPo = &tmpNewPo.DataAgentTplPo

	// newPo.ID = templateID
	newPo.ID = 0
	newPo.Name = templateName
	newPo.Key = templateKey

	newPo.Status = cdaenum.StatusUnpublished

	timeMs := cutil.GetCurrentMSTimestamp()

	newPo.CreatedAt = timeMs
	newPo.CreatedBy = chelper.GetUserIDFromCtx(ctx)

	newPo.UpdatedAt = timeMs
	newPo.UpdatedBy = chelper.GetUserIDFromCtx(ctx)

	newPo.DeletedAt = 0
	newPo.DeletedBy = ""

	newPo.SetIsBuiltIn(cdaenum.BuiltInNo)

	newPo.CreatedType = daenum.AgentTplCreatedTypeCopyFromAgent // 从Agent复制创建

	newPo.SetPublishedAt(0)
	newPo.SetPublishedBy("")

	// 2.1 清除数据源
	err = s.removeDataSourceFromConfig(newPo)
	if err != nil {
		return nil, errors.Wrapf(err, "[createTemplateFromAgent]: 清除数据源失败")
	}

	// 3. 保存模板
	err = s.agentTplRepo.Create(ctx, tx, newPo)
	if err != nil {
		return nil, errors.Wrapf(err, "[createTemplateFromAgent]: 保存模板失败")
	}

	// get po by key
	po, err := s.agentTplRepo.GetByKeyWithTx(ctx, tx, templateKey)
	if err != nil {
		return nil, errors.Wrapf(err, "[createTemplateFromAgent]: 获取模板PO失败")
	}

	// 4. 构建响应
	res = &agentconfigresp.Copy2TplResp{
		ID:   po.ID,
		Name: templateName,
		Key:  templateKey,
	}

	return
}

func (s *dataAgentConfigSvc) removeDataSourceFromConfig(newPo *dapo.DataAgentTplPo) (err error) {
	err = newPo.RemoveDataSourceFromConfig(true)

	return
}

func (s *dataAgentConfigSvc) getNewNameForAgentCopy2Tpl(ctx context.Context, req *agentconfigreq.Copy2TplReq, sourcePo *dapo.DataAgentPo) (newName string, err error) {
	// 1. 检查用户提供的名称是否已存在
	newName = req.Name
	if newName != "" {
		var exists bool

		exists, err = s.agentTplRepo.ExistsByName(ctx, newName)
		if err != nil {
			err = errors.Wrapf(err, "检查模板名称是否存在失败")
			return
		}

		if exists {
			err = capierr.NewCustom409Err(ctx, apierr.AgentTplNameExists, "模板名称已存在")
			return
		}
	}

	// 2. 生成新Agent模板名称

	newName = util.LeftTrimEllipsisSize(sourcePo.Name, cconstant.NameMaxLength-3) + "_模板"

	// retryNum := 0
	//for retryNum < 5 {
	//	var exists bool
	//
	//	exists, err = s.agentTplRepo.ExistsByName(ctx, newName)
	//	if err != nil {
	//		err = errors.Wrapf(err, "检查Agent模板名称是否存在失败")
	//		return
	//	}
	//
	//	if exists {
	//		// 如果检查失败，使用时间戳生成一个唯一的名称
	//		newName += "_" + fmt.Sprintf("%d", cutil.GetCurrentMSTimestamp())
	//
	//		time.Sleep(time.Millisecond * 2)
	//
	//		retryNum++
	//
	//		continue
	//	}
	//
	//	break
	//}
	//
	//// 3. 检查是否生成成功
	//if retryNum >= 5 {
	//	err = capierr.New409Err(ctx, "复制失败，重试生成唯一名称失败")
	//	return
	//}

	return
}

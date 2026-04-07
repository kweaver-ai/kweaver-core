package v3agentconfigsvc

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/constant/daconstant"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/daenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/bizdomainhttp/bizdomainhttpreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/auditlogdto"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent_config/agentconfigreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent_config/agentconfigresp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/apierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cconstant"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/global"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/util"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/pkg/errors"
)

// Copy 复制Agent
func (s *dataAgentConfigSvc) Copy(ctx context.Context, agentID string, req *agentconfigreq.CopyReq) (res *agentconfigresp.CopyResp, auditLogInfo auditlogdto.AgentCopyAuditLogInfo, err error) {
	// 1. 获取源Agent
	sourcePo, err := s.getAgentPoForCopy(ctx, agentID)
	if err != nil {
		return
	}

	auditLogInfo = auditlogdto.AgentCopyAuditLogInfo{
		ID:   agentID,
		Name: sourcePo.Name,
	}

	// 2. 获取新Agent名称
	newName, err := s.getNewNameForAgentCopy(ctx, req, sourcePo)
	if err != nil {
		return
	}

	// 3. 开启事务
	tx, err := s.agentConfRepo.BeginTx(ctx)
	if err != nil {
		err = errors.Wrapf(err, "开启事务失败")
		return
	}

	defer chelper.TxRollbackOrCommit(tx, &err, s.logger)

	// 4. 生成新的Agent ID和Key
	newID := cutil.UlidMake()
	newKey := cutil.UlidMake()

	// 5. 创建新的Agent PO
	newPo := &dapo.DataAgentPo{}

	err = s.copyAgentPo(ctx, newPo, sourcePo, newID, newKey, newName)
	if err != nil {
		err = errors.Wrapf(err, "复制Agent数据失败")
		return
	}

	// 6. 保存新Agent
	err = s.agentConfRepo.Create(ctx, tx, newID, newPo)
	if err != nil {
		err = errors.Wrapf(err, "保存新Agent失败")
		return
	}

	if global.GConfig.IsBizDomainDisabled() {
		res = &agentconfigresp.CopyResp{
			ID:      newID,
			Name:    newName,
			Key:     newKey,
			Version: daconstant.AgentVersionUnpublished,
		}

		return
	}

	// 8. 关联业务域（先写入本地关联表，再调用HTTP）
	bdID := chelper.GetBizDomainIDFromCtx(ctx)

	// 8.1 写入本地关联表
	bdRelPo := &dapo.BizDomainAgentRelPo{
		BizDomainID: bdID,
		AgentID:     newID,
		CreatedAt:   cutil.GetCurrentMSTimestamp(),
	}

	err = s.bdAgentRelRepo.BatchCreate(ctx, tx, []*dapo.BizDomainAgentRelPo{bdRelPo})
	if err != nil {
		return
	}

	// 8.2 调用HTTP接口关联
	err = s.bizDomainHttp.AssociateResource(ctx, &bizdomainhttpreq.AssociateResourceReq{
		ID:   newID,
		BdID: bdID,
		Type: cdaenum.ResourceTypeDataAgent,
	})
	if err != nil {
		return
	}

	// 9. 构建响应
	res = &agentconfigresp.CopyResp{
		ID:      newID,
		Name:    newName,
		Key:     newKey,
		Version: daconstant.AgentVersionUnpublished,
	}

	return
}

func (s *dataAgentConfigSvc) getNewNameForAgentCopy(ctx context.Context, req *agentconfigreq.CopyReq, sourcePo *dapo.DataAgentPo) (newName string, err error) {
	// 1. 检查用户提供的名称是否已存在
	newName = req.Name

	if newName != "" {
		var exists bool

		exists, err = s.agentConfRepo.ExistsByName(ctx, newName)
		if err != nil {
			err = errors.Wrapf(err, "检查Agent名称是否存在失败")
			return
		}

		if exists {
			err = capierr.NewCustom409Err(ctx, apierr.DataAgentConfigNameExists, "Agent名称已存在")
			return
		}
	}

	// 2. 生成新Agent名称
	newName = util.LeftTrimEllipsisSize(sourcePo.Name, cconstant.NameMaxLength-3) + "_副本"

	// retryNum := 0
	//for retryNum < 5 {
	//	var exists bool
	//
	//	exists, err = s.agentConfRepo.ExistsByName(ctx, newName)
	//	if err != nil {
	//		err = errors.Wrapf(err, "检查Agent名称是否存在失败")
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

	//// 3. 检查是否生成成功
	// if retryNum >= 5 {
	//	err = capierr.New409Err(ctx, "复制失败，重试生成唯一名称失败")
	//	return
	//}

	return
}

// copyAgentPo 复制Agent PO数据
func (s *dataAgentConfigSvc) copyAgentPo(ctx context.Context, newPo, sourcePo *dapo.DataAgentPo, newID, newKey, newName string) (err error) {
	// 复制源Agent的所有字段
	err = cutil.CopyStructUseJSON(newPo, sourcePo)
	if err != nil {
		err = errors.Wrapf(err, "[dataAgentConfigSvc][copyAgentPo]: 复制Agent PO数据失败")
		return
	}

	// 设置新的基本信息
	newPo.ID = newID
	newPo.Key = newKey
	newPo.Name = newName

	// 设置状态为草稿
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

	// 清除删除相关字段
	newPo.DeletedAt = 0
	newPo.DeletedBy = ""

	// 设置创建类型为复制
	newPo.CreatedType = daenum.AgentCreatedTypeCopy

	return
}

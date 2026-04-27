package v3agentconfigsvc

import (
	"context"
	"database/sql"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/e2p/daconfe2p"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/auditlogdto"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/agent_config/agentconfigreq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/apierr"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper/grhelper"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
)

func (s *dataAgentConfigSvc) Update(ctx context.Context, req *agentconfigreq.UpdateReq, id string) (auditLogInfo auditlogdto.AgentUpdateAuditLogInfo, err error) {
	// 1. 检查产品是否存在
	exists, err := s.productRepo.ExistsByKey(ctx, req.ProductKey)
	if err != nil {
		return
	}

	if !exists {
		err = capierr.NewCustom409Err(ctx, apierr.ProductNotFound, "产品不存在")
		return
	}

	// 2. 检查agent是否存在
	oldPo, err := s.agentConfRepo.GetByID(ctx, id)
	if err != nil {
		if chelper.IsSqlNotFound(err) {
			err = capierr.NewCustom404Err(ctx, apierr.DataAgentConfigNotFound, "数据智能体配置不存在")
			return
		}

		return
	}

	auditLogInfo = auditlogdto.AgentUpdateAuditLogInfo{
		ID:      id,
		OldName: oldPo.Name,
		NewName: req.Name,
	}

	// 2.1 如果不是内部接口，且没有变化，直接返回
	if !req.IsInternalAPI && !req.IsChanged(oldPo) {
		return
	}

	// 2.2 名称变化时发送MQ，同步更新
	defer func() {
		if err == nil {
			if req.Name != oldPo.Name {
				grhelper.GoSafe(s.logger, func() error {
					return s.handleUpdateNameMq(id, req.Name)
				})
			}
		}
	}()

	// // 2. 检查名称是否重复（如果名称有变化）
	// existsByName, err := s.agentConfRepo.ExistsByNameExcludeID(ctx, req.Name, id)
	// if err != nil {
	// 	return
	// }

	// if existsByName {
	// 	err = capierr.NewCustom409Err(ctx, apierr.DataAgentConfigNameExists, "名称已存在")
	// 	return
	// }

	// 3. DTO 转 EO
	eo, err := req.D2e()
	if err != nil {
		return auditLogInfo, err
	}

	eo.ID = id

	// 4. 开始事务
	tx, err := s.agentConfRepo.BeginTx(ctx)
	if err != nil {
		return
	}

	defer chelper.TxRollbackOrCommit(tx, &err, s.logger)

	// 6. EO转PO
	po, err := daconfe2p.DataAgent(eo)
	if err != nil {
		return
	}

	// 7. 调用repo层更新数据
	err = s.updatePo(ctx, tx, req, po, oldPo)
	if err != nil {
		return
	}

	// 8. 发送审计日志
	// err = s.sendAuditLog(ctx, eo, persrecenums.MngLogOpTypeUpdate, tx)

	return
}

func (s *dataAgentConfigSvc) updatePo(ctx context.Context, tx *sql.Tx, req *agentconfigreq.UpdateReq, po *dapo.DataAgentPo, oldPo *dapo.DataAgentPo) (err error) {
	po.Status = cdaenum.StatusUnpublished
	currentTs := cutil.GetCurrentMSTimestamp()
	po.UpdatedAt = currentTs

	if req.IsInternalAPI {
		po.UpdatedBy = req.UpdatedBy
	} else {
		po.UpdatedBy = chelper.GetUserIDFromCtx(ctx)

		// 不是owner时判断是否是内置Agent，并是否有内置Agent管理权限
		if oldPo.CreatedBy != po.UpdatedBy {
			var hasBuiltInAgentMgmtPermission bool

			isBuiltIn := oldPo.IsBuiltIn.IsBuiltIn()

			if isBuiltIn {
				hasBuiltInAgentMgmtPermission, err = s.isHasBuiltInAgentMgmtPermission(ctx)
				if err != nil {
					return
				}
			}

			// 如果不是内置Agent，或者没有内置Agent管理权限，返回403
			if !isBuiltIn || !hasBuiltInAgentMgmtPermission {
				err = capierr.NewCustom403Err(ctx, apierr.AgentFactoryPermissionForbidden, "不是owner。不是内置Agent或者没有内置Agent管理权限")
				return
			}
		}
	}

	err = s.agentConfRepo.Update(ctx, tx, po)

	return
}

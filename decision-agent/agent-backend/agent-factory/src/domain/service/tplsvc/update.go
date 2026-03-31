package tplsvc

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/e2p/daconfe2p"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/auditlogdto"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent_tpl/agenttplreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/apierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/pkg/errors"
)

func (s *dataAgentTplSvc) Update(ctx context.Context, req *agenttplreq.UpdateReq, id int64) (auditloginfo auditlogdto.AgentTemplateUpdateAuditLogInfo, err error) {
	// 1. 检查模板是否存在
	oldPo, err := s.agentTplRepo.GetByID(ctx, id)
	if err != nil {
		if chelper.IsSqlNotFound(err) {
			err = capierr.NewCustom404Err(ctx, apierr.AgentTplNameExists, "模板不存在")
			return
		}

		err = errors.Wrapf(err, "get template by id")

		return
	}

	auditloginfo = auditlogdto.AgentTemplateUpdateAuditLogInfo{
		ID:   strconv.FormatInt(id, 10),
		Name: oldPo.Name,
	}

	// 2. 检查名称是否与其他模板冲突
	var isNameExists bool

	isNameExists, err = s.agentTplRepo.ExistsByNameExcludeID(ctx, req.Name, id)
	if err != nil {
		err = errors.Wrapf(err, "check template name exists")
		return
	}

	if isNameExists {
		err = capierr.NewCustom409Err(ctx, apierr.AgentTplNameExists, "模板名称已存在")
		return
	}

	// 3. 权限检查
	userID := chelper.GetUserIDFromCtx(ctx)

	if oldPo.CreatedBy != userID {
		err = capierr.NewCustom403Err(ctx, apierr.AgentTplForbiddenNotOwner, "无权限更新，非创建人")
		return
	}

	// 4. DTO 转 EO
	eo, err := req.D2e()
	if err != nil {
		return auditloginfo, err
	}

	eo.ID = id

	// 5. 开启事务
	tx, err := s.agentTplRepo.BeginTx(ctx)
	if err != nil {
		err = errors.Wrapf(err, "begin transaction")
		return
	}

	defer chelper.TxRollbackOrCommit(tx, &err, s.Logger)

	// 6. 构建PO对象
	po, err := daconfe2p.DataAgentTpl(eo)
	if err != nil {
		err = errors.Wrapf(err, "convert entity to po")
		return
	}

	// 7. 更新PO
	err = s.updatePo(ctx, tx, po)
	if err != nil {
		err = errors.Wrapf(err, "update po")
		return
	}

	return
}

func (s *dataAgentTplSvc) updatePo(ctx context.Context, tx *sql.Tx, po *dapo.DataAgentTplPo) (err error) {
	// 设置更新信息
	currentTs := cutil.GetCurrentMSTimestamp()
	po.UpdatedAt = currentTs

	userID := chelper.GetUserIDFromCtx(ctx)
	po.UpdatedBy = userID

	// 编辑后变为未发布
	po.Status = cdaenum.StatusUnpublished

	// 设置为最后一个（详情可见init.sql中改字段的注释）
	// po.SetIsLastOne(cenum.YesNoInt8Yes)

	err = s.agentTplRepo.Update(ctx, tx, po)
	if err != nil {
		return
	}

	return
}

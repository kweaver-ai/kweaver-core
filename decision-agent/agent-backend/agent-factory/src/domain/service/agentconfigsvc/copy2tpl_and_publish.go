package v3agentconfigsvc

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/auditlogdto"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent_config/agentconfigreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent_tpl/agenttplreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent_tpl/agenttplresp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/pkg/errors"
)

// Copy2Tpl 复制Agent为模板
func (s *dataAgentConfigSvc) Copy2TplAndPublish(ctx context.Context, agentID string, req *agenttplreq.PublishReq) (res *agenttplresp.PublishUpsertResp, auditLogInfo auditlogdto.AgentCopy2TplAndPublishAuditLogInfo, err error) {
	// 检查是否有模板发布权限
	hasPms, err := s.isHasTplPublishPermission(ctx)
	if err != nil {
		err = errors.Wrapf(err, "check tpl publish permission failed")
		return
	}

	if !hasPms {
		err = capierr.New403Err(ctx, "do not have tpl publish permission")
		return
	}

	// 1. 开启事务
	tx, err := s.getTx(ctx)
	if err != nil {
		err = errors.Wrapf(err, "[Copy2TplAndPublish][getTx]开启事务失败")
		return
	}
	defer chelper.TxRollbackOrCommit(tx, &err, s.logger)

	// 2. 复制Agent为模板
	copyReq := &agentconfigreq.Copy2TplReq{}

	copyRes, info, err := s.Copy2Tpl(ctx, agentID, copyReq, tx)
	if err != nil {
		err = errors.Wrapf(err, "[Copy2TplAndPublish][Copy2Tpl]复制Agent为模板失败")
		return
	}

	auditLogInfo = auditlogdto.AgentCopy2TplAndPublishAuditLogInfo{
		ID:   agentID,
		Name: info.Name,
	}

	// 3. 发布模板
	res, _, err = s.tplSvc.Publish(ctx, tx, req, copyRes.ID, true)
	if err != nil {
		err = errors.Wrapf(err, "[Copy2TplAndPublish][Publish]发布模板失败")
		return
	}

	return
}

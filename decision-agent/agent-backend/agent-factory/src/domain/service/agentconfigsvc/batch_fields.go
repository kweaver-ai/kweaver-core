package v3agentconfigsvc

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent_config/agentconfigreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent_config/agentconfigresp"
	"github.com/pkg/errors"
)

// BatchFields 批量获取agent指定字段
func (s *dataAgentConfigSvc) BatchFields(ctx context.Context, req *agentconfigreq.BatchFieldsReq) (resp *agentconfigresp.BatchFieldsResp, err error) {
	resp = agentconfigresp.NewBatchFieldsResp()

	// 1. 调用repository获取agent数据
	agentPOs, err := s.agentConfRepo.GetByIDS(ctx, req.AgentIDs)
	if err != nil {
		err = errors.Wrapf(err, "[BatchFields] 获取agent数据失败")
		return
	}

	// 2. 转换为响应格式
	err = resp.LoadFromAgentPOs(agentPOs, req.Fields)
	if err != nil {
		err = errors.Wrapf(err, "[BatchFields] 转换响应数据失败")
		return
	}

	return
}

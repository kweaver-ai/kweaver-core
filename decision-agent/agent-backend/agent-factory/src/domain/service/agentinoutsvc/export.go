package agentinoutsvc

import (
	"context"
	"fmt"
	"time"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent_inout/agentinoutreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent_inout/agentinoutresp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/apierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
)

// Export 导出agent数据
func (s *agentInOutSvc) Export(ctx context.Context, req *agentinoutreq.ExportReq) (resp *agentinoutresp.ExportResp, filename string, err error) {
	resp = agentinoutresp.NewExportResp()

	// 1. 获取用户ID
	uid := chelper.GetUserIDFromCtx(ctx)
	if uid == "" {
		err = capierr.New400Err(ctx, "无法获取用户ID")
		return
	}

	// 2. 检查agent是否存在且属于用户自己的
	agentPOs, err := s.agentConfRepo.GetByIDsAndCreatedBy(ctx, req.AgentIDs, uid)
	if err != nil {
		return
	}

	// 3. 检查是否有不存在的agent
	foundAgentMap := make(map[string]bool)
	for _, po := range agentPOs {
		foundAgentMap[po.ID] = true
	}

	var notFoundAgents []string

	for _, agentID := range req.AgentIDs {
		if !foundAgentMap[agentID] {
			notFoundAgents = append(notFoundAgents, agentID)
		}
	}

	if len(notFoundAgents) > 0 {
		detail := struct {
			NotFoundAgents []string `json:"not_found_agents"`
		}{
			NotFoundAgents: notFoundAgents,
		}

		err = capierr.NewCustom404Err(ctx, apierr.DataAgentConfigNotFound, detail)

		return
	}

	// 4. 转换为实体
	for _, po := range agentPOs {
		resp.AddAgent(po)
	}

	// 5. 生成文件名
	timestamp := time.Now().Format("20060102_150405")
	filename = fmt.Sprintf("agent_export_%s.json", timestamp)

	return
}

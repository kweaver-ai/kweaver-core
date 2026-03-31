package iv3portdriver

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent_inout/agentinoutreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent_inout/agentinoutresp"
)

// IAgentInOutSvc Agent导入导出服务接口
type IAgentInOutSvc interface {
	// Export 导出agent数据
	Export(ctx context.Context, req *agentinoutreq.ExportReq) (resp *agentinoutresp.ExportResp, filename string, err error)

	// Import 导入agent数据
	Import(ctx context.Context, req *agentinoutreq.ImportReq) (resp *agentinoutresp.ImportResp, err error)
}

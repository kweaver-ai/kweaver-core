package agentinoutresp

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
)

// ExportResp 导出agent响应
type ExportResp struct {
	Agents []*ExportAgentItem `json:"agents"`
}

// ExportAgentItem 导出的agent项
type ExportAgentItem struct {
	*dapo.DataAgentPo `json:",inline"`
}

func NewExportResp() *ExportResp {
	return &ExportResp{
		Agents: make([]*ExportAgentItem, 0),
	}
}

// AddAgent 添加agent到导出列表
func (r *ExportResp) AddAgent(po *dapo.DataAgentPo) {
	// 1. 清除数据源
	err := po.RemoveDataSourceFromConfig(true)
	if err != nil {
		return
	}

	// 2. 添加到导出列表
	item := &ExportAgentItem{
		DataAgentPo: po,
	}
	r.Agents = append(r.Agents, item)
}

func (r *ExportResp) GetSystemAgentFailItems() (items []*ImportFailItem) {
	for _, agent := range r.Agents {
		if agent.IsSystemAgent != nil && agent.IsSystemAgent.Bool() {
			items = append(items, &ImportFailItem{
				AgentKey:  agent.Key,
				AgentName: agent.Name,
			})
		}
	}

	return
}

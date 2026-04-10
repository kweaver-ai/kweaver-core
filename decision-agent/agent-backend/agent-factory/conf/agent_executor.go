package conf

import (
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/cconf"
)

type AgentExecutorConf struct {
	PrivateSvc cconf.SvcConf `yaml:"private_svc"`
	// UseV2 控制是否使用 v2 版本的 Agent Executor 接口
	// true: 使用 v2 接口 (agent_id, agent_config, agent_input)
	// false: 使用 v1 接口 (id, config, input)
	UseV2 bool `yaml:"use_v2"`
}

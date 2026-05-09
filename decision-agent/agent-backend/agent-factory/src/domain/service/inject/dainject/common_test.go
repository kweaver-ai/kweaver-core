package dainject

import (
	"sync"
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/cconf"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/conf"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/global"
)

func initInjectGlobalConfig(t *testing.T) {
	t.Helper()

	oldCfg := global.GConfig
	global.GConfig = &conf.Config{
		Config: cconf.BaseDefConfig(),
		AgentFactoryConf: &conf.AgentFactoryConf{
			PrivateSvc: cconf.SvcConf{Protocol: "http", Host: "127.0.0.1", Port: 1},
		},
		AgentExecutorConf: &conf.AgentExecutorConf{
			PrivateSvc: cconf.SvcConf{Protocol: "http", Host: "127.0.0.1", Port: 1},
		},
		UniqueryConf: &conf.UniqueryConf{
			PrivateSvc: cconf.SvcConf{Protocol: "http", Host: "127.0.0.1", Port: 1},
		},
		SandboxPlatformConf: &conf.SandboxPlatformConf{
			PrivateSvc: cconf.SvcConf{Protocol: "http", Host: "127.0.0.1", Port: 1},
		},
		SwitchFields: conf.NewSwitchFields(),
	}

	t.Cleanup(func() {
		global.GConfig = oldCfg
	})
}

func resetInjectSingletons() {
	agentSvcOnce = sync.Once{}
	agentSvcImpl = nil

	conversationSvcOnce = sync.Once{}
	conversationSvcImpl = nil

	sessionSvcOnce = sync.Once{}
	sessionSvcImpl = nil

	squareSvcOnce = sync.Once{}
	squareSvcImpl = nil
}

package httpinject

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/cconf"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/conf"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/global"
)

func resetHttpInjectGlobals() {
	agentExecutorOnce = sync.Once{}
	agentExecutorV1Impl = nil
	agentExecutorV2Impl = nil

	modelApiOnce = sync.Once{}
	modelApiImpl = nil

	sandboxPlatformOnce = sync.Once{}
	sandboxPlatformImpl = nil

	uniqueryOnce = sync.Once{}
	uniqueryImpl = nil
}

func baseHttpInjectConfig(mockSandbox bool) *conf.Config {
	cfg := &conf.Config{
		Config: &cconf.Config{
			ModelFactory: &cconf.ModelFactoryConf{ModelApiSvc: cconf.SvcConf{Host: "127.0.0.1", Port: 19085, Protocol: "http"}},
		},
		SwitchFields: conf.NewSwitchFields(),
		AgentExecutorConf: &conf.AgentExecutorConf{
			PrivateSvc: cconf.SvcConf{Host: "127.0.0.1", Port: 19080, Protocol: "http"},
		},
		UniqueryConf: &conf.UniqueryConf{
			PrivateSvc: cconf.SvcConf{Host: "127.0.0.1", Port: 19081, Protocol: "http"},
			PublicSvc:  cconf.SvcConf{Host: "127.0.0.1", Port: 19082, Protocol: "http"},
		},
		SandboxPlatformConf: &conf.SandboxPlatformConf{
			PrivateSvc: cconf.SvcConf{Host: "127.0.0.1", Port: 19083, Protocol: "http"},
			PublicSvc:  cconf.SvcConf{Host: "127.0.0.1", Port: 19084, Protocol: "http"},
		},
	}
	cfg.SwitchFields.Mock.MockSandboxPlatform = mockSandbox

	return cfg
}

func TestNewAgentExecutorHttpAcc(t *testing.T) {
	// t.Parallel() - 移除：此测试调用单例初始化函数 initAgentExecutors，在并发环境下会导致 sync.Once 死锁
	oldCfg := global.GConfig

	t.Cleanup(func() {
		global.GConfig = oldCfg

		resetHttpInjectGlobals()
	})

	resetHttpInjectGlobals()

	global.GConfig = baseHttpInjectConfig(false)

	v1a := NewAgentExecutorV1HttpAcc()
	require.NotNil(t, v1a)

	v2a := NewAgentExecutorV2HttpAcc()
	require.NotNil(t, v2a)

	v1b := NewAgentExecutorV1HttpAcc()
	v2b := NewAgentExecutorV2HttpAcc()

	assert.Same(t, v1a, v1b)
	assert.Same(t, v2a, v2b)
}

func TestNewModelApiAcc(t *testing.T) {
	// t.Parallel() - 移除：此测试修改全局配置 global.GConfig，在并发环境下不安全
	oldCfg := global.GConfig

	t.Cleanup(func() {
		global.GConfig = oldCfg

		resetHttpInjectGlobals()
	})

	resetHttpInjectGlobals()

	global.GConfig = baseHttpInjectConfig(false)

	m1 := NewModelApiAcc()
	require.NotNil(t, m1)

	m2 := NewModelApiAcc()
	assert.Same(t, m1, m2)
}

func TestNewUniqueryHttpAcc(t *testing.T) {
	// t.Parallel() - 移除：此测试修改全局配置 global.GConfig，在并发环境下不安全
	oldCfg := global.GConfig

	t.Cleanup(func() {
		global.GConfig = oldCfg

		resetHttpInjectGlobals()
	})

	resetHttpInjectGlobals()

	global.GConfig = baseHttpInjectConfig(false)

	u1 := NewUniqueryHttpAcc()
	require.NotNil(t, u1)

	u2 := NewUniqueryHttpAcc()
	assert.Same(t, u1, u2)
}

func TestNewSandboxPlatformHttpAcc(t *testing.T) {
	// t.Parallel() - 移除：此测试修改全局配置 global.GConfig，在并发环境下不安全
	oldCfg := global.GConfig

	t.Cleanup(func() {
		global.GConfig = oldCfg

		resetHttpInjectGlobals()
	})

	t.Run("mock", func(t *testing.T) {
		// t.Parallel() - 移除：子测试修改全局配置
		resetHttpInjectGlobals()

		global.GConfig = baseHttpInjectConfig(true)

		s1 := NewSandboxPlatformHttpAcc()
		require.NotNil(t, s1)

		s2 := NewSandboxPlatformHttpAcc()
		assert.Same(t, s1, s2)
	})

	t.Run("real", func(t *testing.T) {
		// t.Parallel() - 移除：子测试修改全局配置
		resetHttpInjectGlobals()

		global.GConfig = baseHttpInjectConfig(false)

		s1 := NewSandboxPlatformHttpAcc()
		require.NotNil(t, s1)

		s2 := NewSandboxPlatformHttpAcc()
		assert.Same(t, s1, s2)
	})
}

package chttpinject

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/cconf"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/conf"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cglobal"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/global"
)

func resetCHttpInjectGlobals() {
	authZOnce = sync.Once{}
	authZImpl = nil
	bizDomainOnce = sync.Once{}
	bizDomainImpl = nil
	umOnce = sync.Once{}
	umImpl = nil
	userManagementOnce = sync.Once{}
	userManagementImpl = nil
}

func baseConfig(mockAuthz, mockBizDomain bool) (*conf.Config, *cconf.Config) {
	cfg := &conf.Config{SwitchFields: conf.NewSwitchFields()}
	cfg.SwitchFields.Mock.MockAuthZ = mockAuthz
	cfg.SwitchFields.Mock.MockBizDomain = mockBizDomain

	ccfg := &cconf.Config{
		Hydra: cconf.HydraCfg{UserMgnt: cconf.UserMgntCfg{Host: "127.0.0.1", Port: 18083, Protocol: "http"}},
		Authorization: &cconf.AuthzCfg{
			PrivateSvc: &cconf.SvcConf{Host: "127.0.0.1", Port: 18080, Protocol: "http"},
			PublicSvc:  &cconf.SvcConf{Host: "127.0.0.1", Port: 18081, Protocol: "http"},
		},
		BizDomain: &cconf.BizDomainConf{
			PrivateSvc: &cconf.BizDomainSvcConf{Host: "127.0.0.1", Port: 18082, Protocol: "http"},
		},
	}

	return cfg, ccfg
}

func TestNewAuthZHttpAcc(t *testing.T) {
	// 不使用 t.Parallel(): 修改 global.GConfig/cglobal.GConfig 和 singleton once 变量
	oldCfg := global.GConfig
	oldCCfg := cglobal.GConfig

	t.Cleanup(func() {
		global.GConfig = oldCfg
		cglobal.GConfig = oldCCfg

		resetCHttpInjectGlobals()
	})

	t.Run("mock branch", func(t *testing.T) {
		// 不使用 t.Parallel(): 修改全局配置
		resetCHttpInjectGlobals()

		cfg, ccfg := baseConfig(true, false)
		global.GConfig = cfg
		cglobal.GConfig = ccfg

		a1 := NewAuthZHttpAcc()
		require.NotNil(t, a1)

		a2 := NewAuthZHttpAcc()
		assert.Same(t, a1, a2)
	})

	t.Run("real branch", func(t *testing.T) {
		// 不使用 t.Parallel(): 修改全局配置
		resetCHttpInjectGlobals()

		cfg, ccfg := baseConfig(false, false)
		global.GConfig = cfg
		cglobal.GConfig = ccfg

		a1 := NewAuthZHttpAcc()
		require.NotNil(t, a1)

		a2 := NewAuthZHttpAcc()
		assert.Same(t, a1, a2)
	})
}

func TestNewBizDomainHttpAcc(t *testing.T) {
	// 不使用 t.Parallel(): 修改 global.GConfig/cglobal.GConfig 和 singleton once 变量
	oldCfg := global.GConfig
	oldCCfg := cglobal.GConfig

	t.Cleanup(func() {
		global.GConfig = oldCfg
		cglobal.GConfig = oldCCfg

		resetCHttpInjectGlobals()
	})

	t.Run("mock branch", func(t *testing.T) {
		// 不使用 t.Parallel(): 修改全局配置
		resetCHttpInjectGlobals()

		cfg, ccfg := baseConfig(false, true)
		global.GConfig = cfg
		cglobal.GConfig = ccfg

		b1 := NewBizDomainHttpAcc()
		require.NotNil(t, b1)

		b2 := NewBizDomainHttpAcc()
		assert.Same(t, b1, b2)
	})

	t.Run("real branch", func(t *testing.T) {
		// 不使用 t.Parallel(): 修改全局配置
		resetCHttpInjectGlobals()

		cfg, ccfg := baseConfig(false, false)
		global.GConfig = cfg
		cglobal.GConfig = ccfg

		b1 := NewBizDomainHttpAcc()
		require.NotNil(t, b1)

		b2 := NewBizDomainHttpAcc()
		assert.Same(t, b1, b2)
	})
}

func TestNewUmHttpAcc(t *testing.T) {
	// 不使用 t.Parallel(): 修改 global.GConfig/cglobal.GConfig 和 singleton once 变量
	oldCfg := global.GConfig
	oldCCfg := cglobal.GConfig

	t.Cleanup(func() {
		global.GConfig = oldCfg
		cglobal.GConfig = oldCCfg

		resetCHttpInjectGlobals()
	})

	resetCHttpInjectGlobals()

	cfg, ccfg := baseConfig(false, false)
	global.GConfig = cfg
	cglobal.GConfig = ccfg

	u1 := NewUmHttpAcc()
	require.NotNil(t, u1)

	u2 := NewUmHttpAcc()
	assert.Same(t, u1, u2)
}

func TestNewUmHttpAcc_MockUserManagerModule(t *testing.T) {
	oldCfg := global.GConfig
	oldCCfg := cglobal.GConfig

	t.Cleanup(func() {
		global.GConfig = oldCfg
		cglobal.GConfig = oldCCfg

		resetCHttpInjectGlobals()
	})

	resetCHttpInjectGlobals()

	cfg, ccfg := baseConfig(false, false)
	cfg.SwitchFields.Mock.MockUserManagerModule = true
	global.GConfig = cfg
	cglobal.GConfig = ccfg

	umAcc := NewUmHttpAcc()
	require.NotNil(t, umAcc)

	ret, err := umAcc.GetUserIDNameMap(t.Context(), []string{"user-1"})
	require.NoError(t, err)
	assert.Equal(t, "user-1", ret["user-1"])
}

func TestNewUserManagementClient_MockUserManagerModule(t *testing.T) {
	oldCfg := global.GConfig
	oldCCfg := cglobal.GConfig

	t.Cleanup(func() {
		global.GConfig = oldCfg
		cglobal.GConfig = oldCCfg

		resetCHttpInjectGlobals()
	})

	resetCHttpInjectGlobals()

	cfg, ccfg := baseConfig(false, false)
	cfg.SwitchFields.Mock.MockUserManagerModule = true
	global.GConfig = cfg
	cglobal.GConfig = ccfg

	client := NewUserManagementClient()
	require.NotNil(t, client)

	users, err := client.GetUserInfoByUserID(t.Context(), []string{"user-1"}, []string{"name"})
	require.NoError(t, err)
	require.Contains(t, users, "user-1")
	assert.Equal(t, "user-1", users["user-1"].Name)
}

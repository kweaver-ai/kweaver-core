package dainject

import (
	"sync"
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/cconf"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/conf"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cglobal"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/global"
	"github.com/stretchr/testify/assert"
)

func TestGetModelApiUrlPrefix(t *testing.T) {
	t.Parallel()

	t.Run("returns url prefix with http protocol", func(t *testing.T) {
		t.Parallel()

		conf := &cconf.ModelFactoryConf{
			ModelApiSvc: cconf.SvcConf{
				Protocol: "http",
				Host:     "localhost",
				Port:     8080,
			},
		}

		result := getModelApiUrlPrefix(conf)

		assert.Contains(t, result, "http://localhost:8080/api/private/mf-model-api/v1")
	})

	t.Run("returns url prefix with https protocol", func(t *testing.T) {
		t.Parallel()

		conf := &cconf.ModelFactoryConf{
			ModelApiSvc: cconf.SvcConf{
				Protocol: "https",
				Host:     "api.example.com",
				Port:     443,
			},
		}

		result := getModelApiUrlPrefix(conf)

		assert.Contains(t, result, "https://api.example.com:443/api/private/mf-model-api/v1")
	})

	t.Run("handles empty config", func(t *testing.T) {
		t.Parallel()

		conf := &cconf.ModelFactoryConf{
			ModelApiSvc: cconf.SvcConf{
				Protocol: "",
				Host:     "",
				Port:     0,
			},
		}

		result := getModelApiUrlPrefix(conf)

		assert.Contains(t, result, "://:0/api/private/mf-model-api/v1")
	})
}

func initV3InjectGlobalConfig(t *testing.T) {
	t.Helper()

	oldCfg := global.GConfig
	oldCGlobalCfg := cglobal.GConfig

	baseCfg := cconf.BaseDefConfig()
	cglobal.GConfig = baseCfg
	global.GConfig = &conf.Config{
		Config: baseCfg,
		SwitchFields: &conf.SwitchFields{
			Mock: &conf.MockSwitchFields{
				MockAuthZ:     true,
				MockBizDomain: true,
			},
		},
	}

	t.Cleanup(func() {
		global.GConfig = oldCfg
		cglobal.GConfig = oldCGlobalCfg
	})
}

func resetV3InjectSingletons() {
	agentInOutSvcOnce = sync.Once{}
	agentInOutSvcImpl = nil

	daTplSvcOnce = sync.Once{}
	daTplSvcImpl = nil

	bizDomainSvcOnce = sync.Once{}
	bizDomainSvcImpl = nil

	permissionSvcOnce = sync.Once{}
	permissionSvcImpl = nil

	personalSpaceSvcOnce = sync.Once{}
	personalSpaceSvcImpl = nil

	publishedSvcOnce = sync.Once{}
	publishedSvcImpl = nil

	releaseSvcOnce = sync.Once{}
	releaseSvcImpl = nil
}

package capimiddleware

import (
	"context"
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/cconf"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/conf"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/global"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestMockHydra_VerifyToken_UsesConfiguredMockUserID(t *testing.T) {
	oldCfg := global.GConfig

	t.Cleanup(func() {
		global.GConfig = oldCfg
	})

	cfg := &conf.Config{
		Config:       cconf.BaseDefConfig(),
		SwitchFields: conf.NewSwitchFields(),
	}

	err := yaml.Unmarshal([]byte(`
switch_fields:
  mock:
    mock_user_id: configured-user-id
`), cfg)
	assert.NoError(t, err)

	global.GConfig = cfg

	visitor, err := (&MockHydra{}).VerifyToken(context.Background(), nil)
	assert.NoError(t, err)
	assert.Equal(t, "configured-user-id", visitor.ID)
}

func TestMockHydra_VerifyToken_UsesDefaultUserIDWhenConfigEmpty(t *testing.T) {
	oldCfg := global.GConfig

	t.Cleanup(func() {
		global.GConfig = oldCfg
	})

	global.GConfig = &conf.Config{
		Config:       cconf.BaseDefConfig(),
		SwitchFields: conf.NewSwitchFields(),
	}

	visitor, err := (&MockHydra{}).VerifyToken(context.Background(), nil)
	assert.NoError(t, err)
	assert.Equal(t, defaultMockUserID, visitor.ID)
}

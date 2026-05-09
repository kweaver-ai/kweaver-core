package conf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func boolPtr(v bool) *bool {
	return &v
}

func TestConfig_IsAuthEnabled_DefaultTrueWhenUnset(t *testing.T) {
	t.Parallel()

	cfg := &Config{}

	assert.True(t, cfg.IsAuthEnabled())
}

func TestConfig_NormalizeAuthRelatedSwitches_DisabledForcesRelevantSwitches(t *testing.T) {
	t.Parallel()

	cfg := &Config{
		AuthEnable:   boolPtr(false),
		SwitchFields: nil,
	}

	cfg.normalizeAuthRelatedSwitches()

	assert.NotNil(t, cfg.SwitchFields)
	assert.NotNil(t, cfg.SwitchFields.Mock)
	assert.True(t, cfg.SwitchFields.DisablePmsCheck)
	assert.True(t, cfg.SwitchFields.Mock.MockHydra)
	assert.True(t, cfg.SwitchFields.Mock.MockAuthZ)
	assert.True(t, cfg.SwitchFields.Mock.MockUserManagerModule)
}

func TestConfig_NormalizeAuthRelatedSwitches_EnabledDoesNotClearExistingSwitches(t *testing.T) {
	t.Parallel()

	cfg := &Config{
		AuthEnable: boolPtr(true),
		SwitchFields: &SwitchFields{
			DisablePmsCheck: true,
			Mock: &MockSwitchFields{
				MockHydra:             true,
				MockAuthZ:             true,
				MockUserManagerModule: true,
			},
		},
	}

	cfg.normalizeAuthRelatedSwitches()

	assert.True(t, cfg.SwitchFields.DisablePmsCheck)
	assert.True(t, cfg.SwitchFields.Mock.MockHydra)
	assert.True(t, cfg.SwitchFields.Mock.MockAuthZ)
	assert.True(t, cfg.SwitchFields.Mock.MockUserManagerModule)
}

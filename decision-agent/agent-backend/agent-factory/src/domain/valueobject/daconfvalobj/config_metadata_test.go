package daconfvalobj

import (
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum/agentconfigenum"
	"github.com/stretchr/testify/assert"
)

func TestConfigMetadata_SetConfigTplVersion_Valid(t *testing.T) {
	t.Parallel()

	metadata := &ConfigMetadata{}

	assert.NotPanics(t, func() {
		metadata.SetConfigTplVersion(agentconfigenum.ConfigTplVersionV1)
	})

	assert.Equal(t, agentconfigenum.ConfigTplVersionV1, metadata.ConfigTplVersion)
}

func TestConfigMetadata_SetConfigTplVersion_Invalid(t *testing.T) {
	t.Parallel()

	metadata := &ConfigMetadata{}

	assert.Panics(t, func() {
		metadata.SetConfigTplVersion("invalid_version")
	})
}

func TestConfigMetadata_GetConfigTplVersion(t *testing.T) {
	t.Parallel()

	metadata := &ConfigMetadata{
		ConfigTplVersion: agentconfigenum.ConfigTplVersionV1,
	}

	result := metadata.GetConfigTplVersion()
	assert.Equal(t, agentconfigenum.ConfigTplVersionV1, result)
}

func TestConfigMetadata_SetConfigLastSetTimestamp(t *testing.T) {
	t.Parallel()

	metadata := &ConfigMetadata{}
	before := metadata.ConfigLastSetTimestamp

	metadata.SetConfigLastSetTimestamp()

	after := metadata.ConfigLastSetTimestamp
	assert.Greater(t, after, before)
	assert.Greater(t, after, uint64(0))
}

func TestConfigMetadata_GetConfigLastSetTimestamp(t *testing.T) {
	t.Parallel()

	expectedTimestamp := uint64(1234567890)
	metadata := &ConfigMetadata{
		ConfigLastSetTimestamp: expectedTimestamp,
	}

	result := metadata.GetConfigLastSetTimestamp()
	assert.Equal(t, expectedTimestamp, result)
}

func TestConfigMetadata_Fields(t *testing.T) {
	t.Parallel()

	metadata := &ConfigMetadata{
		ConfigTplVersion:       agentconfigenum.ConfigTplVersionV1,
		ConfigLastSetTimestamp: 9876543210,
	}

	assert.Equal(t, agentconfigenum.ConfigTplVersionV1, metadata.ConfigTplVersion)
	assert.Equal(t, uint64(9876543210), metadata.ConfigLastSetTimestamp)
}

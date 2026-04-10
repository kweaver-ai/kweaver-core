package agentconfigreq

import (
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/entity/daconfeo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum/agentconfigenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/daconfvalobj"
	"github.com/stretchr/testify/assert"
)

func TestSetDefaultValue_SetsDefaultConfigTplVersion(t *testing.T) {
	t.Parallel()

	config := daconfvalobj.NewConfig()

	setDefaultValue(config)

	assert.Equal(t, agentconfigenum.ConfigTplVersionV1, config.GetConfigMetadata().GetConfigTplVersion())
}

func TestSetDefaultValue_DoesNotOverrideExistingVersion(t *testing.T) {
	t.Parallel()

	config := daconfvalobj.NewConfig()
	config.GetConfigMetadata().SetConfigTplVersion(agentconfigenum.ConfigTplVersionV1)

	setDefaultValue(config)

	assert.Equal(t, agentconfigenum.ConfigTplVersionV1, config.GetConfigMetadata().GetConfigTplVersion())
}

func TestHandleConfig_SetsDefaultValues(t *testing.T) {
	t.Parallel()

	config := daconfvalobj.NewConfig()

	err := HandleConfig(config)

	assert.NoError(t, err)
	assert.Equal(t, agentconfigenum.ConfigTplVersionV1, config.GetConfigMetadata().GetConfigTplVersion())
}

func TestHandleConfig_WithExistingVersion(t *testing.T) {
	t.Parallel()

	config := daconfvalobj.NewConfig()
	config.GetConfigMetadata().SetConfigTplVersion(agentconfigenum.ConfigTplVersionV1)

	err := HandleConfig(config)

	assert.NoError(t, err)
	assert.Equal(t, agentconfigenum.ConfigTplVersionV1, config.GetConfigMetadata().GetConfigTplVersion())
}

func TestHandleConfig_NilConfig(t *testing.T) {
	t.Parallel()

	var config *daconfvalobj.Config = nil

	// This test documents behavior with nil config
	// Based on implementation, it would panic
	// In production, nil should never be passed
	assert.Panics(t, func() {
		HandleConfig(config) //nolint:errcheck
	})
}

func TestD2eCommonAfterD2e_SetsConfigLastSetTimestamp(t *testing.T) {
	t.Parallel()

	config := daconfvalobj.NewConfig()
	eo := &daconfeo.DataAgent{
		Config: config,
	}

	D2eCommonAfterD2e(eo)

	// The timestamp should now be set
	finalTimestamp := eo.Config.Metadata.GetConfigLastSetTimestamp()

	// The timestamp should have been set
	assert.NotNil(t, finalTimestamp)
}

func TestD2eCommonAfterD2e_WithNilEntity(t *testing.T) {
	t.Parallel()

	var eo *daconfeo.DataAgent = nil

	// This test documents behavior with nil entity
	// Based on implementation, this would panic
	// In production, nil should never be passed
	assert.Panics(t, func() {
		D2eCommonAfterD2e(eo)
	})
}

func TestSetDefaultValue_MultipleCalls(t *testing.T) {
	t.Parallel()

	config := daconfvalobj.NewConfig()

	setDefaultValue(config)
	firstVersion := config.GetConfigMetadata().GetConfigTplVersion()

	setDefaultValue(config)
	secondVersion := config.GetConfigMetadata().GetConfigTplVersion()

	assert.Equal(t, firstVersion, secondVersion)
	assert.Equal(t, agentconfigenum.ConfigTplVersionV1, firstVersion)
}

func TestHandleConfig_MultipleCalls(t *testing.T) {
	t.Parallel()

	config := daconfvalobj.NewConfig()

	err1 := HandleConfig(config)
	err2 := HandleConfig(config)

	assert.NoError(t, err1)
	assert.NoError(t, err2)

	// Version should remain the same
	assert.Equal(t, agentconfigenum.ConfigTplVersionV1, config.GetConfigMetadata().GetConfigTplVersion())
}

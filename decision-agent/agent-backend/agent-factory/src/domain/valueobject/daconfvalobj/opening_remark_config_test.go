package daconfvalobj

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOpeningRemarkConfig_ValObjCheck_FixedType(t *testing.T) {
	t.Parallel()

	config := &OpeningRemarkConfig{
		Type:               "fixed",
		FixedOpeningRemark: "Hello! How can I help you today?",
	}

	err := config.ValObjCheck()
	assert.NoError(t, err)
}

func TestOpeningRemarkConfig_ValObjCheck_DynamicType(t *testing.T) {
	t.Parallel()

	config := &OpeningRemarkConfig{
		Type:                       "dynamic",
		DynamicOpeningRemarkPrompt: "Generate a greeting based on context",
	}

	err := config.ValObjCheck()
	assert.NoError(t, err)
}

func TestOpeningRemarkConfig_ValObjCheck_EmptyType(t *testing.T) {
	t.Parallel()

	config := &OpeningRemarkConfig{}

	err := config.ValObjCheck()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "type is required")
}

func TestOpeningRemarkConfig_ValObjCheck_InvalidType(t *testing.T) {
	t.Parallel()

	config := &OpeningRemarkConfig{
		Type:               "invalid_type",
		FixedOpeningRemark: "Test",
	}

	err := config.ValObjCheck()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "type must be fixed or dynamic")
}

func TestOpeningRemarkConfig_ValObjCheck_FixedTypeWithoutRemark(t *testing.T) {
	t.Parallel()

	config := &OpeningRemarkConfig{
		Type: "fixed",
	}

	err := config.ValObjCheck()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "fixed_opening_remark is required when type is fixed")
}

func TestOpeningRemarkConfig_ValObjCheck_DynamicTypeWithoutPrompt(t *testing.T) {
	t.Parallel()

	config := &OpeningRemarkConfig{
		Type: "dynamic",
	}

	err := config.ValObjCheck()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "dynamic_opening_remark_prompt is required when type is dynamic")
}

func TestOpeningRemarkConfig_Fields(t *testing.T) {
	t.Parallel()

	config := &OpeningRemarkConfig{
		Type:                       "fixed",
		FixedOpeningRemark:         "Fixed greeting",
		DynamicOpeningRemarkPrompt: "Dynamic prompt",
	}

	assert.Equal(t, "fixed", config.Type)
	assert.Equal(t, "Fixed greeting", config.FixedOpeningRemark)
	assert.Equal(t, "Dynamic prompt", config.DynamicOpeningRemarkPrompt)
}

func TestOpeningRemarkConfig_Empty(t *testing.T) {
	t.Parallel()

	config := &OpeningRemarkConfig{}

	assert.Empty(t, config.Type)
	assert.Empty(t, config.FixedOpeningRemark)
	assert.Empty(t, config.DynamicOpeningRemarkPrompt)
}

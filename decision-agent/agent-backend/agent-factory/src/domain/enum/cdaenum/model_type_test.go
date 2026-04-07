package cdaenum

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestModelType_Constants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, ModelType("llm"), ModelTypeLlm)
	assert.Equal(t, ModelType("rlm"), ModelTypeRlm)
}

func TestModelType_EnumCheck_Valid(t *testing.T) {
	t.Parallel()

	validTypes := []ModelType{
		ModelTypeLlm,
		ModelTypeRlm,
	}

	for _, modelType := range validTypes {
		t.Run(string(modelType), func(t *testing.T) {
			t.Parallel()

			err := modelType.EnumCheck()
			assert.NoError(t, err)
		})
	}
}

func TestModelType_EnumCheck_Invalid(t *testing.T) {
	t.Parallel()

	invalidType := ModelType("invalid_type")
	err := invalidType.EnumCheck()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid model type")
}

func TestModelType_EnumCheck_Empty(t *testing.T) {
	t.Parallel()

	emptyType := ModelType("")
	err := emptyType.EnumCheck()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid model type")
}

func TestModelType_AllUnique(t *testing.T) {
	t.Parallel()

	modelTypes := []ModelType{
		ModelTypeLlm,
		ModelTypeRlm,
	}

	uniqueTypes := make(map[ModelType]bool)
	for _, mt := range modelTypes {
		assert.False(t, uniqueTypes[mt], "Duplicate model type found: %s", mt)
		uniqueTypes[mt] = true
	}
}

package cdaenum

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToolType_Constants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, ToolType("tool"), ToolTypeTool)
	assert.Equal(t, ToolType("agent"), ToolTypeAgent)
}

func TestToolType_EnumCheck_Valid(t *testing.T) {
	t.Parallel()

	validTypes := []ToolType{
		ToolTypeTool,
		ToolTypeAgent,
	}

	for _, toolType := range validTypes {
		t.Run(string(toolType), func(t *testing.T) {
			t.Parallel()

			err := toolType.EnumCheck()
			assert.NoError(t, err)
		})
	}
}

func TestToolType_EnumCheck_Invalid(t *testing.T) {
	t.Parallel()

	invalidType := ToolType("invalid_type")
	err := invalidType.EnumCheck()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid tool type")
}

func TestToolType_EnumCheck_Empty(t *testing.T) {
	t.Parallel()

	emptyType := ToolType("")
	err := emptyType.EnumCheck()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid tool type")
}

func TestToolType_AllUnique(t *testing.T) {
	t.Parallel()

	toolTypes := []ToolType{
		ToolTypeTool,
		ToolTypeAgent,
	}

	uniqueTypes := make(map[ToolType]bool)
	for _, tt := range toolTypes {
		assert.False(t, uniqueTypes[tt], "Duplicate tool type found: %s", tt)
		uniqueTypes[tt] = true
	}
}

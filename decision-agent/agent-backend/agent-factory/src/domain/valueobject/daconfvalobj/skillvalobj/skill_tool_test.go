package skillvalobj

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSkillTool_ValObjCheck_Valid(t *testing.T) {
	t.Parallel()

	tool := &SkillTool{
		ToolID:    "tool-123",
		ToolBoxID: "toolbox-456",
	}

	err := tool.ValObjCheck()

	assert.NoError(t, err)
}

func TestSkillTool_ValObjCheck_EmptyToolID(t *testing.T) {
	t.Parallel()

	tool := &SkillTool{
		ToolID:    "",
		ToolBoxID: "toolbox-456",
	}

	err := tool.ValObjCheck()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tool_id is required")
}

func TestSkillTool_ValObjCheck_EmptyToolBoxID(t *testing.T) {
	t.Parallel()

	tool := &SkillTool{
		ToolID:    "tool-123",
		ToolBoxID: "",
	}

	err := tool.ValObjCheck()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tool_box_id is required")
}

func TestSkillTool_ValObjCheck_EmptyBoth(t *testing.T) {
	t.Parallel()

	tool := &SkillTool{
		ToolID:    "",
		ToolBoxID: "",
	}

	err := tool.ValObjCheck()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tool_id is required")
}

func TestSkillTool_NewSkillTool(t *testing.T) {
	t.Parallel()

	inputData := json.RawMessage(`{"param": "value"}`)
	strategies := []ResultProcessStrategy{
		{
			Category: Category{ID: "cat1", Name: "Category 1", Description: "Test category"},
			Strategy: Strategy{ID: "str1", Name: "Strategy 1", Description: "Test strategy"},
		},
	}
	tool := &SkillTool{
		ToolID:                          "tool-123",
		ToolBoxID:                       "toolbox-456",
		ToolTimeout:                     30,
		ToolInput:                       inputData,
		Intervention:                    true,
		InterventionConfirmationMessage: "Please confirm",
		ResultProcessStrategies:         strategies,
	}

	assert.Equal(t, "tool-123", tool.ToolID)
	assert.Equal(t, "toolbox-456", tool.ToolBoxID)
	assert.Equal(t, 30, tool.ToolTimeout)
	assert.Equal(t, inputData, tool.ToolInput)
	assert.True(t, tool.Intervention)
	assert.Equal(t, "Please confirm", tool.InterventionConfirmationMessage)
	assert.Len(t, tool.ResultProcessStrategies, 1)
}

func TestSkillTool_Empty(t *testing.T) {
	t.Parallel()

	tool := &SkillTool{}

	assert.Empty(t, tool.ToolID)
	assert.Empty(t, tool.ToolBoxID)
	assert.Equal(t, 0, tool.ToolTimeout)
	assert.False(t, tool.Intervention)
	assert.Empty(t, tool.InterventionConfirmationMessage)
	assert.Nil(t, tool.ResultProcessStrategies)
}

func TestResultProcessStrategy_Fields(t *testing.T) {
	t.Parallel()

	strategy := ResultProcessStrategy{
		Category: Category{
			ID:          "cat-1",
			Name:        "Test Category",
			Description: "Test category description",
		},
		Strategy: Strategy{
			ID:          "str-1",
			Name:        "Test Strategy",
			Description: "Test strategy description",
		},
	}

	assert.Equal(t, "cat-1", strategy.Category.ID)
	assert.Equal(t, "Test Category", strategy.Category.Name)
	assert.Equal(t, "Test category description", strategy.Category.Description)
	assert.Equal(t, "str-1", strategy.Strategy.ID)
	assert.Equal(t, "Test Strategy", strategy.Strategy.Name)
	assert.Equal(t, "Test strategy description", strategy.Strategy.Description)
}

func TestSkillTool_WithEmptyStrategies(t *testing.T) {
	t.Parallel()

	tool := &SkillTool{
		ToolID:                  "tool-123",
		ToolBoxID:               "toolbox-456",
		ResultProcessStrategies: []ResultProcessStrategy{},
	}

	assert.NotNil(t, tool.ResultProcessStrategies)
	assert.Len(t, tool.ResultProcessStrategies, 0)
}

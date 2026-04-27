package agentrespvo

import (
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/chat_enum/chatresenum"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMiddleOutputVarRes(t *testing.T) {
	t.Parallel()

	res := NewMiddleOutputVarRes()

	assert.NotNil(t, res)
	assert.NotNil(t, res.Vars)
	assert.Empty(t, res.Vars)
}

func TestMiddleOutputVarRes_LoadFrom_Simple(t *testing.T) {
	t.Parallel()

	res := NewMiddleOutputVarRes()

	vars := []string{"var1", "var2"}
	valuesMap := map[string]interface{}{
		"var1": "simple string value",
		"var2": 12345,
	}
	interventionMap := map[string][]*Intervention{}

	err := res.LoadFrom(vars, valuesMap, interventionMap)
	require.NoError(t, err)
	assert.Len(t, res.Vars, 2)

	// Check var1
	assert.Equal(t, "var1", res.Vars[0].VarName)
	assert.Equal(t, chatresenum.OutputVarTypeOther, res.Vars[0].Type)
	assert.Equal(t, "simple string value", res.Vars[0].Value)

	// Check var2
	assert.Equal(t, "var2", res.Vars[1].VarName)
	assert.Equal(t, chatresenum.OutputVarTypeOther, res.Vars[1].Type)
	assert.Equal(t, 12345, res.Vars[1].Value)
}

func TestMiddleOutputVarRes_LoadFrom_WithInterventions(t *testing.T) {
	t.Parallel()

	res := NewMiddleOutputVarRes()

	vars := []string{"var1"}
	valuesMap := map[string]interface{}{
		"var1": "test value",
	}
	interventions := []*Intervention{
		{
			ToolName: "test_tool",
			ToolCallInfo: &ToolCallInfo{
				ToolName: "test_tool",
				Args:     map[string]interface{}{"arg1": "value1"},
			},
		},
	}
	interventionMap := map[string][]*Intervention{
		"var1": interventions,
	}

	err := res.LoadFrom(vars, valuesMap, interventionMap)
	require.NoError(t, err)
	assert.Len(t, res.Vars, 1)
	assert.Len(t, res.Vars[0].Interventions, 1)
	assert.Equal(t, "test_tool", res.Vars[0].Interventions[0].ToolName)
}

func TestMiddleOutputVarRes_LoadFrom_EmptyVars(t *testing.T) {
	t.Parallel()

	res := NewMiddleOutputVarRes()

	vars := []string{}
	valuesMap := map[string]interface{}{}
	interventionMap := map[string][]*Intervention{}

	err := res.LoadFrom(vars, valuesMap, interventionMap)
	require.NoError(t, err)
	assert.Empty(t, res.Vars)
}

func TestMiddleOutputVarRes_LoadFrom_SomeVarsMissing(t *testing.T) {
	t.Parallel()

	res := NewMiddleOutputVarRes()

	vars := []string{"var1", "var2", "var3"}
	valuesMap := map[string]interface{}{
		"var1": "value1",
		// var2 is missing
		"var3": "value3",
	}
	interventionMap := map[string][]*Intervention{}

	err := res.LoadFrom(vars, valuesMap, interventionMap)
	require.NoError(t, err)
	assert.Len(t, res.Vars, 2) // Only var1 and var3
}

func TestMiddleOutputVarRes_LoadFrom_WithPromptType(t *testing.T) {
	t.Parallel()

	res := NewMiddleOutputVarRes()

	vars := []string{"promptVar"}
	valuesMap := map[string]interface{}{
		"promptVar": map[string]interface{}{
			"answer": "This is the answer",
			"think":  "This is the thinking",
		},
	}
	interventionMap := map[string][]*Intervention{}

	err := res.LoadFrom(vars, valuesMap, interventionMap)
	require.NoError(t, err)
	assert.Len(t, res.Vars, 1)

	// Check that the prompt type is detected
	assert.Equal(t, "promptVar", res.Vars[0].VarName)
	assert.Equal(t, chatresenum.OutputVarTypePrompt, res.Vars[0].Type)
	assert.Equal(t, "This is the answer", res.Vars[0].Value)
	assert.Equal(t, "This is the thinking", res.Vars[0].Thinking)
}

func TestMiddleOutputVarRes_LoadFrom_MixedVarTypes(t *testing.T) {
	t.Parallel()

	res := NewMiddleOutputVarRes()

	vars := []string{"promptVar", "exploreVar", "otherVar", "missingVar"}
	valuesMap := map[string]interface{}{
		"promptVar": map[string]interface{}{
			"answer": "prompt-answer",
			"think":  "prompt-think",
		},
		"exploreVar": []interface{}{
			map[string]interface{}{
				"agent_name":  "agent-1",
				"answer":      "a1",
				"think":       "t1",
				"status":      "success",
				"interrupted": false,
			},
		},
		"otherVar": "plain-value",
	}
	interventionMap := map[string][]*Intervention{}

	err := res.LoadFrom(vars, valuesMap, interventionMap)
	require.NoError(t, err)
	require.Len(t, res.Vars, 3)

	assert.Equal(t, "promptVar", res.Vars[0].VarName)
	assert.Equal(t, chatresenum.OutputVarTypePrompt, res.Vars[0].Type)
	assert.Equal(t, "prompt-answer", res.Vars[0].Value)
	assert.Equal(t, "prompt-think", res.Vars[0].Thinking)

	assert.Equal(t, "exploreVar", res.Vars[1].VarName)
	assert.Equal(t, chatresenum.OutputVarTypeExplore, res.Vars[1].Type)

	assert.Equal(t, "otherVar", res.Vars[2].VarName)
	assert.Equal(t, chatresenum.OutputVarTypeOther, res.Vars[2].Type)
	assert.Equal(t, "plain-value", res.Vars[2].Value)
}

func TestGetPromptVal_ErrorCases(t *testing.T) {
	t.Parallel()

	t.Run("marshal error", func(t *testing.T) {
		t.Parallel()

		_, _, err := getPromptVal(make(chan int))
		require.Error(t, err)
	})

	t.Run("answer is not string", func(t *testing.T) {
		t.Parallel()

		_, _, err := getPromptVal(map[string]interface{}{
			"answer": 123,
			"think":  "ok",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "answer is not string")
	})

	t.Run("think is not string", func(t *testing.T) {
		t.Parallel()

		_, _, err := getPromptVal(map[string]interface{}{
			"answer": "ok",
			"think":  123,
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "think is not string")
	})
}

func TestMiddleOutputVarRes_LoadFrom_WithExploreType(t *testing.T) {
	t.Parallel()

	res := NewMiddleOutputVarRes()

	vars := []string{"exploreVar"}
	valuesMap := map[string]interface{}{
		"exploreVar": []interface{}{
			map[string]interface{}{
				"agent_name":  "agent1",
				"answer":      "Answer 1",
				"think":       "Think 1",
				"status":      "success",
				"interrupted": false,
			},
			map[string]interface{}{
				"agent_name":  "agent2",
				"answer":      "Answer 2",
				"think":       "Think 2",
				"status":      "success",
				"interrupted": false,
			},
		},
	}
	interventionMap := map[string][]*Intervention{}

	err := res.LoadFrom(vars, valuesMap, interventionMap)
	require.NoError(t, err)
	assert.Len(t, res.Vars, 1)

	// Check that the explore type is detected
	assert.Equal(t, "exploreVar", res.Vars[0].VarName)
	assert.Equal(t, chatresenum.OutputVarTypeExplore, res.Vars[0].Type)
}

func TestMiddleOutputVarRes_ToExploreList_Success(t *testing.T) {
	t.Parallel()

	res := NewMiddleOutputVarRes()

	exploreData := []interface{}{
		map[string]interface{}{
			"agent_name":  "agent1",
			"answer":      "Answer 1",
			"think":       "Think 1",
			"status":      "success",
			"interrupted": false,
		},
	}

	exploreList, err := res.ToExploreList(exploreData)
	require.NoError(t, err)
	assert.Len(t, exploreList, 1)
	assert.Equal(t, "agent1", exploreList[0].AgentName)
}

func TestMiddleOutputVarRes_ToExploreList_InvalidData(t *testing.T) {
	t.Parallel()

	res := NewMiddleOutputVarRes()

	// Invalid data - not an explore list
	invalidData := "not an explore list"

	// CopyUseJSON might not error for invalid data, just return empty list
	exploreList, _ := res.ToExploreList(invalidData)
	// The function might not error, just check the result
	assert.NotNil(t, exploreList)
	assert.Empty(t, exploreList)
}

func TestMiddleOutputVarRes_LoadFrom_PromptWithNonStringAnswer(t *testing.T) {
	t.Parallel()

	res := NewMiddleOutputVarRes()

	vars := []string{"promptVar"}
	valuesMap := map[string]interface{}{
		"promptVar": map[string]interface{}{
			"answer": 12345, // answer is not a string
			"think":  "This is the thinking",
		},
	}
	interventionMap := map[string][]*Intervention{}

	err := res.LoadFrom(vars, valuesMap, interventionMap)
	// Should not error, but should be treated as Other type
	assert.NoError(t, err)
	assert.Len(t, res.Vars, 1)
	// Since getPromptVal fails, it should be treated as Other type
	assert.Equal(t, chatresenum.OutputVarTypeOther, res.Vars[0].Type)
}

func TestMiddleOutputVarRes_LoadFrom_PromptWithNonStringThink(t *testing.T) {
	t.Parallel()

	res := NewMiddleOutputVarRes()

	vars := []string{"promptVar"}
	valuesMap := map[string]interface{}{
		"promptVar": map[string]interface{}{
			"answer": "This is the answer",
			"think":  12345, // think is not a string
		},
	}
	interventionMap := map[string][]*Intervention{}

	err := res.LoadFrom(vars, valuesMap, interventionMap)
	// Should not error, but should be treated as Other type
	assert.NoError(t, err)
	assert.Len(t, res.Vars, 1)
	// Since getPromptVal fails, it should be treated as Other type
	assert.Equal(t, chatresenum.OutputVarTypeOther, res.Vars[0].Type)
}

func TestMiddleOutputVarRes_LoadFrom_WithUnmarshalableValue(t *testing.T) {
	t.Parallel()

	res := NewMiddleOutputVarRes()

	vars := []string{"var1"}
	// Create a value that cannot be marshaled by sonic (a channel)
	unmarshalableValue := make(chan int)
	valuesMap := map[string]interface{}{
		"var1": unmarshalableValue,
	}
	interventionMap := map[string][]*Intervention{}

	err := res.LoadFrom(vars, valuesMap, interventionMap)
	// The value cannot be marshaled, so getVarType returns empty string
	assert.NoError(t, err)
	// The var should still be added but with empty type
	assert.Len(t, res.Vars, 1)
	// Since getVarType fails, it returns empty string
	assert.Equal(t, chatresenum.OutputVarType(""), res.Vars[0].Type)
}

func TestMiddleOutputVarRes_LoadFrom_ValidPrompt(t *testing.T) {
	t.Parallel()

	// Test with valid prompt structure to fully cover getPromptVal
	res := NewMiddleOutputVarRes()

	vars := []string{"promptVar"}
	valuesMap := map[string]interface{}{
		"promptVar": map[string]interface{}{
			"answer": "This is the answer",
			"think":  "This is the thinking",
		},
	}
	interventionMap := map[string][]*Intervention{}

	err := res.LoadFrom(vars, valuesMap, interventionMap)
	assert.NoError(t, err)
	assert.Len(t, res.Vars, 1)
	assert.Equal(t, chatresenum.OutputVarTypePrompt, res.Vars[0].Type)
	assert.Equal(t, "This is the answer", res.Vars[0].Value)
	assert.Equal(t, "This is the thinking", res.Vars[0].Thinking)
}

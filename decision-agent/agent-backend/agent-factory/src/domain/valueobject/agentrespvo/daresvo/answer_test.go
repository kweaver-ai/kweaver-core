package daresvo

import (
	"context"
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/agentconfigvo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/agentrespvo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDataAgentRes_Effective(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		setup    func() *DataAgentRes
		expected bool
	}{
		{
			name: "has final answer",
			setup: func() *DataAgentRes {
				res := &DataAgentRes{
					Answer: agentrespvo.NewAnswerS(),
				}
				res.finalAnswerVarHelper = NewResHelper(res.Answer, &agentconfigvo.OutputVariablesS{
					AnswerVar: "answer",
				}, VarFieldTypeFinalAnswer)
				res.Answer.SetField("answer", "test answer")
				return res
			},
			expected: true,
		},
		{
			name: "no final answer",
			setup: func() *DataAgentRes {
				res := &DataAgentRes{
					Answer: agentrespvo.NewAnswerS(),
				}
				res.finalAnswerVarHelper = NewResHelper(res.Answer, &agentconfigvo.OutputVariablesS{
					AnswerVar: "answer",
				}, VarFieldTypeFinalAnswer)
				return res
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			res := tt.setup()
			assert.Equal(t, tt.expected, res.Effective())
		})
	}
}

func TestDataAgentRes_GetFinalAnswer(t *testing.T) {
	t.Parallel()

	res := &DataAgentRes{
		Answer: agentrespvo.NewAnswerS(),
	}
	res.finalAnswerVarHelper = NewResHelper(res.Answer, &agentconfigvo.OutputVariablesS{
		AnswerVar: "final_answer",
	}, VarFieldTypeFinalAnswer)
	res.Answer.SetField("final_answer", "my answer")

	answer := res.GetFinalAnswer()
	assert.NotNil(t, answer)
	assert.Equal(t, "my answer", answer)
}

func TestDataAgentRes_GetFinalAnswer_Nil(t *testing.T) {
	t.Parallel()

	res := &DataAgentRes{
		Answer: agentrespvo.NewAnswerS(),
	}
	res.finalAnswerVarHelper = NewResHelper(res.Answer, &agentconfigvo.OutputVariablesS{
		AnswerVar: "final_answer",
	}, VarFieldTypeFinalAnswer)
	// Don't set any field

	answer := res.GetFinalAnswer()
	assert.Nil(t, answer)
}

func TestDataAgentRes_GetFinalAnswerJSON_Valid(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	data := []byte(`{
		"answer": {
			"answer": "Test answer"
		}
	}`)

	outputVars := &agentconfigvo.OutputVariablesS{
		AnswerVar: "answer",
	}

	res, err := NewDataAgentRes(ctx, data, outputVars)
	assert.NoError(t, err)

	jsonBytes, err := res.GetFinalAnswerJSON()
	assert.NoError(t, err)
	assert.Contains(t, string(jsonBytes), "Test answer")
}

func TestDataAgentRes_GetFinalAnswerJSON_Empty(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	data := []byte(`{
		"answer": {
			"other_field": "value"
		}
	}`)

	outputVars := &agentconfigvo.OutputVariablesS{
		AnswerVar: "non_existent_field",
	}

	res, err := NewDataAgentRes(ctx, data, outputVars)
	require.NoError(t, err)

	jsonBytes, err := res.GetFinalAnswerJSON()
	assert.NoError(t, err)
	assert.Nil(t, jsonBytes)
}

func TestDataAgentRes_GetExploreAnswerList_EmptyAnswer(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	data := []byte(`{
		"answer": {}
	}`)

	outputVars := &agentconfigvo.OutputVariablesS{
		AnswerVar: "answer",
	}

	res, err := NewDataAgentRes(ctx, data, outputVars)
	require.NoError(t, err)

	answerList, ok := res.GetExploreAnswerList()
	assert.False(t, ok)
	assert.NotNil(t, answerList) // Returns empty slice, not nil
}

func TestDataAgentRes_IsPromptType_EmptyAnswer(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	data := []byte(`{
		"answer": {}
	}`)

	outputVars := &agentconfigvo.OutputVariablesS{
		AnswerVar: "answer",
	}

	res, err := NewDataAgentRes(ctx, data, outputVars)
	require.NoError(t, err)

	answer, ok := res.IsPromptType()
	assert.False(t, ok)
	assert.NotNil(t, answer) // IsPromptType always returns a non-nil AnswerPrompt
}

func TestDataAgentRes_GetFinalAnswerJSON_Error(t *testing.T) {
	t.Parallel()

	res := &DataAgentRes{
		Answer: agentrespvo.NewAnswerS(),
	}
	// Create a helper with invalid config to trigger error
	res.finalAnswerVarHelper = NewResHelper(res.Answer, &agentconfigvo.OutputVariablesS{
		AnswerVar: "answer",
	}, VarFieldTypeFinalAnswer)
	// Set an invalid value that can't be marshaled to JSON properly
	res.Answer.SetField("answer", func() {}) // Functions can't be marshaled to JSON

	jsonBytes, err := res.GetFinalAnswerJSON()
	assert.Error(t, err)
	assert.Nil(t, jsonBytes)
}

package daresvo

import (
	"context"
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/agentconfigvo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDataAgentRes_GetExploreAnswerList_NotExploreType(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	// Not an explore type - answer is a nested object but not an explore array
	data := []byte(`{
		"answer": {
			"answer": "Test answer content"
		}
	}`)
	outputVars := &agentconfigvo.OutputVariablesS{
		AnswerVar: "answer",
	}

	res, err := NewDataAgentRes(ctx, data, outputVars)
	require.NoError(t, err)

	answerList, ok := res.GetExploreAnswerList()
	assert.False(t, ok)
	assert.Empty(t, answerList)
}

func TestDataAgentRes_IsPromptType_NotPromptType(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	// Not a prompt type - answer is a nested object but not a prompt
	data := []byte(`{
		"answer": {
			"answer": "Test answer content"
		}
	}`)
	outputVars := &agentconfigvo.OutputVariablesS{
		AnswerVar: "answer",
	}

	res, err := NewDataAgentRes(ctx, data, outputVars)
	require.NoError(t, err)

	answer, ok := res.IsPromptType()
	assert.False(t, ok)
	// Answer object is created but empty, not nil
	assert.NotNil(t, answer)
}

func TestDataAgentRes_GetExploreAnswerList_ExploreType(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	// Valid explore type - an array with required fields
	// Note: The answer field should contain the array as a nested field under "answer"
	data := []byte(`{
		"answer": {
			"answer": [
				{
					"agent_name": "agent1",
					"answer": "Answer content 1",
					"think": "Think content 1",
					"status": "success",
					"interrupted": false
				},
				{
					"agent_name": "agent2",
					"answer": "Answer content 2",
					"think": "Think content 2",
					"status": "success",
					"interrupted": false
				}
			]
		}
	}`)
	outputVars := &agentconfigvo.OutputVariablesS{
		AnswerVar: "answer",
	}

	res, err := NewDataAgentRes(ctx, data, outputVars)
	require.NoError(t, err)

	answerList, ok := res.GetExploreAnswerList()
	assert.True(t, ok)
	assert.Len(t, answerList, 2)
	assert.Equal(t, "agent1", answerList[0].AgentName)
	assert.Equal(t, "agent2", answerList[1].AgentName)
}

func TestDataAgentRes_IsPromptType_PromptType(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	// Valid prompt type - an object with answer and think fields
	// Note: The answer field should contain the object as a nested field under "answer"
	data := []byte(`{
		"answer": {
			"answer": {
				"answer": "Test answer content",
				"think": "Test think content"
			}
		}
	}`)
	outputVars := &agentconfigvo.OutputVariablesS{
		AnswerVar: "answer",
	}

	res, err := NewDataAgentRes(ctx, data, outputVars)
	require.NoError(t, err)

	answer, ok := res.IsPromptType()
	assert.True(t, ok)
	assert.NotNil(t, answer)
	assert.Equal(t, "Test answer content", answer.Answer)
	assert.Equal(t, "Test think content", answer.Think)
}

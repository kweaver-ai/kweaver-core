package agentrespvo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAnswerExplore_NewAnswerExplore(t *testing.T) {
	t.Parallel()

	explore := &AnswerExplore{
		AgentName:    "Test Agent",
		Answer:       "Test answer",
		Think:        "Test thinking",
		Status:       "completed",
		Interrupted:  false,
		InputMessage: "Test input",
	}

	assert.Equal(t, "Test Agent", explore.AgentName)
	assert.Equal(t, "Test answer", explore.Answer)
	assert.Equal(t, "Test thinking", explore.Think)
	assert.Equal(t, "completed", explore.Status)
	assert.False(t, explore.Interrupted)
	assert.Equal(t, "Test input", explore.InputMessage)
}

func TestAnswerExplore_Empty(t *testing.T) {
	t.Parallel()

	explore := &AnswerExplore{}

	assert.Empty(t, explore.AgentName)
	assert.Nil(t, explore.Answer)
	assert.Empty(t, explore.Think)
	assert.Empty(t, explore.Status)
	assert.False(t, explore.Interrupted)
	assert.Nil(t, explore.InputMessage)
}

func TestAnswerExplore_WithInterrupted(t *testing.T) {
	t.Parallel()

	explore := &AnswerExplore{
		AgentName:   "Interrupted Agent",
		Status:      "interrupted",
		Interrupted: true,
	}

	assert.True(t, explore.Interrupted)
	assert.Equal(t, "interrupted", explore.Status)
}

func TestAnswerExplore_WithNilAnswer(t *testing.T) {
	t.Parallel()

	explore := &AnswerExplore{
		AgentName: "Test Agent",
		Answer:    nil,
	}

	assert.Nil(t, explore.Answer)
	assert.Equal(t, "Test Agent", explore.AgentName)
}

func TestAnswerExplore_WithComplexAnswer(t *testing.T) {
	t.Parallel()

	complexAnswer := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
	}
	explore := &AnswerExplore{
		Answer: complexAnswer,
	}

	assert.NotNil(t, explore.Answer)
	assert.Equal(t, complexAnswer, explore.Answer)
}

package agentrespvo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAnswerPrompt_NewAnswerPrompt(t *testing.T) {
	t.Parallel()

	prompt := &AnswerPrompt{
		Answer: "Test answer",
		Think:  "Test thinking",
	}

	assert.Equal(t, "Test answer", prompt.Answer)
	assert.Equal(t, "Test thinking", prompt.Think)
}

func TestAnswerPrompt_Empty(t *testing.T) {
	t.Parallel()

	prompt := &AnswerPrompt{}

	assert.Empty(t, prompt.Answer)
	assert.Empty(t, prompt.Think)
}

func TestAnswerPrompt_WithOnlyAnswer(t *testing.T) {
	t.Parallel()

	prompt := &AnswerPrompt{
		Answer: "Only answer",
	}

	assert.Equal(t, "Only answer", prompt.Answer)
	assert.Empty(t, prompt.Think)
}

func TestAnswerPrompt_WithOnlyThink(t *testing.T) {
	t.Parallel()

	prompt := &AnswerPrompt{
		Think: "Only thinking",
	}

	assert.Empty(t, prompt.Answer)
	assert.Equal(t, "Only thinking", prompt.Think)
}

func TestAnswerPromptText_NewAnswerPromptText(t *testing.T) {
	t.Parallel()

	text := &AnswerPromptText{
		Text: "Test text content",
	}

	assert.Equal(t, "Test text content", text.Text)
}

func TestAnswerPromptText_Empty(t *testing.T) {
	t.Parallel()

	text := &AnswerPromptText{}

	assert.Empty(t, text.Text)
}

func TestAnswerPromptText_WithEmptyString(t *testing.T) {
	t.Parallel()

	text := &AnswerPromptText{
		Text: "",
	}

	assert.Empty(t, text.Text)
}

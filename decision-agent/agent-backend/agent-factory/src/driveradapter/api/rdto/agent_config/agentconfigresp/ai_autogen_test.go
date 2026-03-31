package agentconfigresp

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAiAutogenRes_StructFields(t *testing.T) {
	t.Parallel()

	resp := AiAutogenRes{
		Result: "AI generated content",
	}

	assert.Equal(t, "AI generated content", resp.Result)
}

func TestAiAutogenRes_Empty(t *testing.T) {
	t.Parallel()

	resp := AiAutogenRes{}

	assert.Empty(t, resp.Result)
}

func TestAiAutogenRes_WithLongContent(t *testing.T) {
	t.Parallel()

	longContent := "This is a very long AI generated content that contains multiple sentences and detailed information."
	resp := AiAutogenRes{
		Result: longContent,
	}

	assert.Equal(t, longContent, resp.Result)
	assert.Contains(t, resp.Result, "multiple sentences")
}

func TestAiAutogenRes_WithSpecialCharacters(t *testing.T) {
	t.Parallel()

	resp := AiAutogenRes{
		Result: "Content with \"quotes\" and 'apostrophes' and 新文字",
	}

	assert.Contains(t, resp.Result, "quotes")
	assert.Contains(t, resp.Result, "新文字")
}

func TestAiAutogenRes_JSONTags(t *testing.T) {
	t.Parallel()

	resp := AiAutogenRes{
		Result: "test result",
	}

	// Marshal to JSON
	data, err := json.Marshal(resp)
	require.NoError(t, err)

	// Unmarshal to map to check JSON tags
	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	assert.Equal(t, "test result", result["result"])
}

func TestPreSetQuestions_Type(t *testing.T) {
	t.Parallel()

	// PreSetQuestions is a slice type
	var questions PreSetQuestions

	assert.Nil(t, questions)
	assert.IsType(t, PreSetQuestions{}, questions)
}

func TestPreSetQuestions_Empty(t *testing.T) {
	t.Parallel()

	questions := PreSetQuestions{}

	assert.Empty(t, questions)
	assert.Len(t, questions, 0)
}

func TestPreSetQuestions_WithItems(t *testing.T) {
	t.Parallel()

	questions := PreSetQuestions{
		"Question 1?",
		"Question 2?",
		"Question 3?",
	}

	assert.Len(t, questions, 3)
	assert.Equal(t, "Question 1?", questions[0])
	assert.Equal(t, "Question 2?", questions[1])
	assert.Equal(t, "Question 3?", questions[2])
}

func TestPreSetQuestions_Append(t *testing.T) {
	t.Parallel()

	questions := PreSetQuestions{}

	questions = append(questions, "New Question 1?")
	questions = append(questions, "New Question 2?")

	assert.Len(t, questions, 2)
	assert.Equal(t, "New Question 1?", questions[0])
	assert.Equal(t, "New Question 2?", questions[1])
}

func TestPreSetQuestions_JSONMarshaling(t *testing.T) {
	t.Parallel()

	questions := PreSetQuestions{
		"What is AI?",
		"How does machine learning work?",
		"Explain deep learning.",
	}

	// Marshal to JSON
	data, err := json.Marshal(questions)
	require.NoError(t, err)

	// Unmarshal back
	var result PreSetQuestions
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	assert.Len(t, result, 3)
	assert.Equal(t, "What is AI?", result[0])
	assert.Equal(t, "How does machine learning work?", result[1])
}

func TestPreSetQuestions_WithSpecialCharacters(t *testing.T) {
	t.Parallel()

	questions := PreSetQuestions{
		"问题一？",
		"Question with \"quotes\"?",
		"Question with 'apostrophies'?",
	}

	assert.Len(t, questions, 3)
	assert.Equal(t, "问题一？", questions[0])
	assert.Contains(t, questions[1], "quotes")
}

func TestPreSetQuestions_EmptyStrings(t *testing.T) {
	t.Parallel()

	questions := PreSetQuestions{
		"",
		"Valid question?",
		"",
	}

	assert.Len(t, questions, 3)
	assert.Empty(t, questions[0])
	assert.NotEmpty(t, questions[1])
	assert.Empty(t, questions[2])
}

func TestPreSetQuestions_LongQuestions(t *testing.T) {
	t.Parallel()

	longQuestion := "This is a very long question that contains multiple words and detailed information about a specific topic?"
	questions := PreSetQuestions{
		longQuestion,
		"Short?",
	}

	assert.Len(t, questions, 2)
	assert.Equal(t, longQuestion, questions[0])
	assert.Equal(t, "Short?", questions[1])
}

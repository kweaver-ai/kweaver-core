package chatresenum

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAnswerType_Constants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, AnswerType("explore"), AnswerTypeExplore)
	assert.Equal(t, AnswerType("prompt"), AnswerTypePrompt)
	assert.Equal(t, AnswerType("other"), AnswerTypeOther)
}

func TestAnswerType_NotEmpty(t *testing.T) {
	t.Parallel()

	assert.NotEmpty(t, AnswerTypeExplore)
	assert.NotEmpty(t, AnswerTypePrompt)
	assert.NotEmpty(t, AnswerTypeOther)
}

func TestAnswerType_AreUnique(t *testing.T) {
	t.Parallel()

	values := []AnswerType{
		AnswerTypeExplore,
		AnswerTypePrompt,
		AnswerTypeOther,
	}

	uniqueValues := make(map[string]bool)

	for _, v := range values {
		strValue := string(v)
		assert.False(t, uniqueValues[strValue], "Duplicate value found: %s", strValue)
		uniqueValues[strValue] = true
	}
}

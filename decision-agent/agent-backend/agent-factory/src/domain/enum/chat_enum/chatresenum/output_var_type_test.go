package chatresenum

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOutputVarType_Constants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, OutputVarType("prompt"), OutputVarTypePrompt)
	assert.Equal(t, OutputVarType("explore"), OutputVarTypeExplore)
	assert.Equal(t, OutputVarType("other"), OutputVarTypeOther)
}

func TestOutputVarType_NotEmpty(t *testing.T) {
	t.Parallel()

	assert.NotEmpty(t, OutputVarTypePrompt)
	assert.NotEmpty(t, OutputVarTypeExplore)
	assert.NotEmpty(t, OutputVarTypeOther)
}

func TestOutputVarType_AreUnique(t *testing.T) {
	t.Parallel()

	values := []OutputVarType{
		OutputVarTypePrompt,
		OutputVarTypeExplore,
		OutputVarTypeOther,
	}

	uniqueValues := make(map[string]bool)

	for _, v := range values {
		strValue := string(v)
		assert.False(t, uniqueValues[strValue], "Duplicate value found: %s", strValue)
		uniqueValues[strValue] = true
	}
}

func TestOutputVarType_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		varType  OutputVarType
		expected string
	}{
		{
			name:     "Prompt type",
			varType:  OutputVarTypePrompt,
			expected: "prompt",
		},
		{
			name:     "Explore type",
			varType:  OutputVarTypeExplore,
			expected: "explore",
		},
		{
			name:     "Other type",
			varType:  OutputVarTypeOther,
			expected: "other",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := string(tt.varType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

package daenum

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAiAutogenFrom_EnumCheck_Valid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		from AiAutogenFrom
	}{
		{"system prompt", AiAutogenFromSystemPrompt},
		{"opening remarks", AiAutogenFromOpeningRemarks},
		{"preset question", AiAutogenFromPreSetQuestion},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.from.EnumCheck()
			assert.NoError(t, err)
		})
	}
}

func TestAiAutogenFrom_EnumCheck_Invalid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		from AiAutogenFrom
	}{
		{"empty", ""},
		{"invalid", "invalid_from"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.from.EnumCheck()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid enum value")
		})
	}
}

func TestAiAutogenFrom_String(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "system_prompt", string(AiAutogenFromSystemPrompt))
	assert.Equal(t, "opening_remarks", string(AiAutogenFromOpeningRemarks))
	assert.Equal(t, "preset_question", string(AiAutogenFromPreSetQuestion))
}

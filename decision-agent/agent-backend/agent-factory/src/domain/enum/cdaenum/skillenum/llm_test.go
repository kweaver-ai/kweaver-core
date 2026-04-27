package skillenum

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLLM_EnumCheck_Valid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		llm  LLM
	}{
		{"inherit main", LLMInheritMain},
		{"self configured", LLMSelfConfiged},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.llm.EnumCheck()
			assert.NoError(t, err)
		})
	}
}

func TestLLM_EnumCheck_Invalid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		llm  LLM
	}{
		{"empty llm", ""},
		{"invalid llm", "invalid_llm"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.llm.EnumCheck()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid skill agent llm")
		})
	}
}

func TestLLM_String(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "inherit_main", string(LLMInheritMain))
	assert.Equal(t, "self_configured", string(LLMSelfConfiged))
}

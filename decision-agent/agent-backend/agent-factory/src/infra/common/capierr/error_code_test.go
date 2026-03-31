package capierr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorCodeConstants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value string
	}{
		{
			name:  "DataAgentConfigLlmRequired",
			value: DataAgentConfigLlmRequired,
		},
		{
			name:  "DataAgentConfigRetrieverDataSourceKnEntryExceedLimitSize",
			value: DataAgentConfigRetrieverDataSourceKnEntryExceedLimitSize,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.NotEmpty(t, tt.value)
			assert.Contains(t, tt.value, "AgentFactory")
		})
	}
}

func TestErrorCodeConstants_Format(t *testing.T) {
	t.Parallel()

	// Verify the error code format follows the pattern
	assert.Contains(t, DataAgentConfigLlmRequired, ".BadRequest.")
	assert.Contains(t, DataAgentConfigRetrieverDataSourceKnEntryExceedLimitSize, ".BadRequest.")
}

func TestErrorCodeConstants_Uniqueness(t *testing.T) {
	t.Parallel()

	// Ensure error codes are unique
	codes := []string{
		DataAgentConfigLlmRequired,
		DataAgentConfigRetrieverDataSourceKnEntryExceedLimitSize,
	}

	uniqueCodes := make(map[string]bool)
	for _, code := range codes {
		uniqueCodes[code] = true
	}

	assert.Equal(t, len(codes), len(uniqueCodes), "Error codes should be unique")
}

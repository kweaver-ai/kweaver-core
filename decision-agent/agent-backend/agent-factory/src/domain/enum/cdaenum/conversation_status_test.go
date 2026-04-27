package cdaenum

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConversationStatus_Constants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, ConversationStatus("processing"), ConvStatusProcessing)
	assert.Equal(t, ConversationStatus("completed"), ConvStatusCompleted)
	assert.Equal(t, ConversationStatus("cancelled"), ConvStatusCancelled)
	assert.Equal(t, ConversationStatus("failed"), ConvStatusFailed)
}

func TestConversationStatus_StringValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		status   ConversationStatus
		expected string
	}{
		{
			name:     "processing status",
			status:   ConvStatusProcessing,
			expected: "processing",
		},
		{
			name:     "completed status",
			status:   ConvStatusCompleted,
			expected: "completed",
		},
		{
			name:     "cancelled status",
			status:   ConvStatusCancelled,
			expected: "cancelled",
		},
		{
			name:     "failed status",
			status:   ConvStatusFailed,
			expected: "failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := string(tt.status)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConversationStatus_AllUnique(t *testing.T) {
	t.Parallel()

	statuses := []ConversationStatus{
		ConvStatusProcessing,
		ConvStatusCompleted,
		ConvStatusCancelled,
		ConvStatusFailed,
	}

	uniqueStatuses := make(map[ConversationStatus]bool)
	for _, status := range statuses {
		assert.False(t, uniqueStatuses[status], "Duplicate status found: %s", status)
		uniqueStatuses[status] = true
	}
}

func TestConversationStatus_NotEmpty(t *testing.T) {
	t.Parallel()

	assert.NotEmpty(t, ConvStatusProcessing)
	assert.NotEmpty(t, ConvStatusCompleted)
	assert.NotEmpty(t, ConvStatusCancelled)
	assert.NotEmpty(t, ConvStatusFailed)
}

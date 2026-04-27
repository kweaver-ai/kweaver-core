package constant

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileCheckStatusConstants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "completed", FileCheckStatusSuccess)
	assert.Equal(t, "failed", FileCheckStatusFailed)
	assert.Equal(t, "processing", FileCheckStatusProcessing)
}

func TestFileCheckStatusSuccess_NotEmpty(t *testing.T) {
	t.Parallel()

	assert.NotEmpty(t, FileCheckStatusSuccess)
}

func TestFileCheckStatusFailed_NotEmpty(t *testing.T) {
	t.Parallel()

	assert.NotEmpty(t, FileCheckStatusFailed)
}

func TestFileCheckStatusProcessing_NotEmpty(t *testing.T) {
	t.Parallel()

	assert.NotEmpty(t, FileCheckStatusProcessing)
}

func TestFileCheckStatusValues_AreUnique(t *testing.T) {
	t.Parallel()

	values := []string{
		FileCheckStatusSuccess,
		FileCheckStatusFailed,
		FileCheckStatusProcessing,
	}

	// Check that all values are unique
	uniqueValues := make(map[string]bool)
	for _, v := range values {
		assert.False(t, uniqueValues[v], "Duplicate value found: %s", v)
		uniqueValues[v] = true
	}
}

package constant

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalMode(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "normal", NormalMode)
}

func TestDeepThinkingMode(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "deep_thinking", DeepThinkingMode)
}

func TestChatModes_AreUnique(t *testing.T) {
	t.Parallel()

	assert.NotEqual(t, NormalMode, DeepThinkingMode)
}

func TestChatModes_NotEmpty(t *testing.T) {
	t.Parallel()

	assert.NotEmpty(t, NormalMode)
	assert.NotEmpty(t, DeepThinkingMode)
}

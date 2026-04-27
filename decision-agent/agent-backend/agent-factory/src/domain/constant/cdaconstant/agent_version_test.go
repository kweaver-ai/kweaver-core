package cdaconstant

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAgentVersionUnpublished(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "v0", AgentVersionUnpublished)
}

func TestAgentVersionLatest(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "latest", AgentVersionLatest)
}

func TestAgentVersionConstants_NotEmpty(t *testing.T) {
	t.Parallel()

	assert.NotEmpty(t, AgentVersionUnpublished)
	assert.NotEmpty(t, AgentVersionLatest)
}

func TestAgentVersionConstants_AreUnique(t *testing.T) {
	t.Parallel()

	assert.NotEqual(t, AgentVersionUnpublished, AgentVersionLatest)
}

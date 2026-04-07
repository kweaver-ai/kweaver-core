package daconstant

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAgentVersion_Constants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "v0", AgentVersionUnpublished)
	assert.Equal(t, "latest", AgentVersionLatest)
}

func TestAgentInoutMaxSize(t *testing.T) {
	t.Parallel()

	assert.Equal(t, 500, AgentInoutMaxSize)
}

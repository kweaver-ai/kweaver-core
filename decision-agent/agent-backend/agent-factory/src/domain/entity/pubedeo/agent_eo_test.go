package pubedeo

import (
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/daconfvalobj"
	"github.com/stretchr/testify/assert"
)

func TestPublishedAgentEo_NewPublishedAgentEo(t *testing.T) {
	t.Parallel()

	config := daconfvalobj.NewConfig()
	eo := &PublishedAgentEo{
		Config:          config,
		PublishedByName: "Publisher",
	}

	assert.NotNil(t, eo)
	assert.NotNil(t, eo.Config)
	assert.Equal(t, "Publisher", eo.PublishedByName)
}

func TestPublishedAgentEo_Empty(t *testing.T) {
	t.Parallel()

	eo := &PublishedAgentEo{}

	assert.NotNil(t, eo)
	assert.Nil(t, eo.Config)
	assert.Empty(t, eo.PublishedByName)
}

func TestPublishedAgentEo_WithNilConfig(t *testing.T) {
	t.Parallel()

	eo := &PublishedAgentEo{
		Config:          nil,
		PublishedByName: "Test Publisher",
	}

	assert.Nil(t, eo.Config)
	assert.Equal(t, "Test Publisher", eo.PublishedByName)
}

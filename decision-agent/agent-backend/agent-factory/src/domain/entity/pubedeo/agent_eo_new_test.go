package pubedeo

import (
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/daconfvalobj"
	"github.com/stretchr/testify/assert"
)

func TestPublishedAgentEo_New(t *testing.T) {
	t.Parallel()

	eo := &PublishedAgentEo{
		PublishedByName: "John Doe",
		Config:          &daconfvalobj.Config{},
	}

	assert.NotNil(t, eo)
	assert.Equal(t, "John Doe", eo.PublishedByName)
	assert.NotNil(t, eo.Config)
}

func TestPublishedAgentEo_WithConfig(t *testing.T) {
	t.Parallel()

	config := &daconfvalobj.Config{
		Metadata: daconfvalobj.ConfigMetadata{},
	}

	eo := &PublishedAgentEo{
		Config: config,
	}

	assert.Equal(t, config, eo.Config)
}

func TestPublishedAgentEo_NilConfig(t *testing.T) {
	t.Parallel()

	eo := &PublishedAgentEo{
		Config: nil,
	}

	assert.NotNil(t, eo)
	assert.Nil(t, eo.Config)
}

func TestPublishedAgentEo_EmptyName(t *testing.T) {
	t.Parallel()

	eo := &PublishedAgentEo{
		PublishedByName: "",
	}

	assert.NotNil(t, eo)
	assert.Empty(t, eo.PublishedByName)
}

func TestPublishedAgentEo_WithSpecialCharactersInName(t *testing.T) {
	t.Parallel()

	eo := &PublishedAgentEo{
		PublishedByName: "张三 🌍",
	}

	assert.Equal(t, "张三 🌍", eo.PublishedByName)
}

func TestPublishedAgentEo_StructFields(t *testing.T) {
	t.Parallel()

	eo := &PublishedAgentEo{
		Config: &daconfvalobj.Config{},
	}

	assert.NotNil(t, eo.Config)
}

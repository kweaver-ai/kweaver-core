package pubedeo

import (
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/stretchr/testify/assert"
)

func TestPublishedTplListEo_NewPublishedTplListEo(t *testing.T) {
	t.Parallel()

	eo := &PublishedTplListEo{
		PublishedTplPo: dapo.PublishedTplPo{
			ID: 123,
		},
		CreatedByName:   "User 1",
		UpdatedByName:   "User 2",
		PublishedByName: "User 3",
	}

	assert.NotNil(t, eo)
	assert.Equal(t, int64(123), eo.ID)
	assert.Equal(t, "User 1", eo.CreatedByName)
	assert.Equal(t, "User 2", eo.UpdatedByName)
	assert.Equal(t, "User 3", eo.PublishedByName)
}

func TestPublishedTplListEo_Empty(t *testing.T) {
	t.Parallel()

	eo := &PublishedTplListEo{}

	assert.NotNil(t, eo)
	assert.Empty(t, eo.CreatedByName)
	assert.Empty(t, eo.UpdatedByName)
	assert.Empty(t, eo.PublishedByName)
}

func TestPublishedTplListEo_WithPartialData(t *testing.T) {
	t.Parallel()

	eo := &PublishedTplListEo{
		CreatedByName: "Creator",
	}

	assert.Equal(t, "Creator", eo.CreatedByName)
	assert.Empty(t, eo.UpdatedByName)
	assert.Empty(t, eo.PublishedByName)
}

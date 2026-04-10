package daconfeo

import (
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/stretchr/testify/assert"
)

func TestDataAgentTplListEo_Fields(t *testing.T) {
	t.Parallel()

	eo := &DataAgentTplListEo{
		DataAgentTplPo: dapo.DataAgentTplPo{
			Name: "Test Template",
		},
		CreatedByName:   "User 1",
		UpdatedByName:   "User 2",
		PublishedByName: "User 3",
	}

	assert.Equal(t, "Test Template", eo.Name)
	assert.Equal(t, "User 1", eo.CreatedByName)
	assert.Equal(t, "User 2", eo.UpdatedByName)
	assert.Equal(t, "User 3", eo.PublishedByName)
}

func TestDataAgentTplListEo_Empty(t *testing.T) {
	t.Parallel()

	eo := &DataAgentTplListEo{}

	assert.Empty(t, eo.Name)
	assert.Empty(t, eo.CreatedByName)
	assert.Empty(t, eo.UpdatedByName)
	assert.Empty(t, eo.PublishedByName)
}

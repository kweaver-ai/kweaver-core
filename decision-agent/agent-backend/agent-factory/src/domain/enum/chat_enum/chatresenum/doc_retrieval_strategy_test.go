package chatresenum

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDocRetrievalStrategyStandard(t *testing.T) {
	t.Parallel()

	assert.Equal(t, DocRetrievalStrategy("standard"), DocRetrievalStrategyStandard)
}

func TestDocRetrievalStrategy_NotEmpty(t *testing.T) {
	t.Parallel()

	assert.NotEmpty(t, DocRetrievalStrategyStandard)
}

func TestDocRetrievalStrategy_String(t *testing.T) {
	t.Parallel()

	strategy := DocRetrievalStrategyStandard
	assert.Equal(t, "standard", string(strategy))
}

func TestDocRetrievalStrategy_CustomValue(t *testing.T) {
	t.Parallel()

	customStrategy := DocRetrievalStrategy("custom")
	assert.Equal(t, "custom", string(customStrategy))
}

package publishvo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewListPublishInfo(t *testing.T) {
	t.Parallel()

	info := NewListPublishInfo()

	assert.NotNil(t, info)
	assert.NotNil(t, info.PublishedToBeStruct)
}

func TestListPublishInfo_MultipleInstances(t *testing.T) {
	t.Parallel()

	info1 := NewListPublishInfo()
	info2 := NewListPublishInfo()

	assert.NotNil(t, info1)
	assert.NotNil(t, info2)
	assert.NotSame(t, info1, info2)
}

func TestListPublishInfo_EmbeddedStruct(t *testing.T) {
	t.Parallel()

	info := NewListPublishInfo()

	// Verify the embedded struct is accessible
	assert.NotNil(t, info.PublishedToBeStruct)
}

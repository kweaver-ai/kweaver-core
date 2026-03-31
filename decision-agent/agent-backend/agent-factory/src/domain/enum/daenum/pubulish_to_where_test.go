package daenum

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPublishToWhere_Constants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, PublishToWhere("custom_space"), PublishToWhereCustomSpace)
	assert.Equal(t, PublishToWhere("square"), PublishToWhereSquare)
}

func TestPublishToWhere_EnumCheck_Valid(t *testing.T) {
	t.Parallel()

	validTypes := []PublishToWhere{
		PublishToWhereCustomSpace,
		PublishToWhereSquare,
	}

	for _, ptw := range validTypes {
		t.Run(string(ptw), func(t *testing.T) {
			t.Parallel()

			err := ptw.EnumCheck()
			assert.NoError(t, err)
		})
	}
}

func TestPublishToWhere_EnumCheck_Invalid(t *testing.T) {
	t.Parallel()

	invalidType := PublishToWhere("invalid_type")
	err := invalidType.EnumCheck()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid publish to where")
}

func TestPublishToWhere_EnumCheck_Empty(t *testing.T) {
	t.Parallel()

	emptyType := PublishToWhere("")
	err := emptyType.EnumCheck()
	assert.Error(t, err)
}

func TestPublishToWhere_AllUnique(t *testing.T) {
	t.Parallel()

	publishTypes := []PublishToWhere{
		PublishToWhereCustomSpace,
		PublishToWhereSquare,
	}

	uniqueTypes := make(map[PublishToWhere]bool)
	for _, ptw := range publishTypes {
		assert.False(t, uniqueTypes[ptw], "Duplicate publish type found: %s", ptw)
		uniqueTypes[ptw] = true
	}
}

func TestPublishToWhere_StringValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		ptw      PublishToWhere
		expected string
	}{
		{
			name:     "custom space type",
			ptw:      PublishToWhereCustomSpace,
			expected: "custom_space",
		},
		{
			name:     "square type",
			ptw:      PublishToWhereSquare,
			expected: "square",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := string(tt.ptw)
			assert.Equal(t, tt.expected, result)
		})
	}
}

package daenum

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAgentCreatedType_Constants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, AgentCreatedType("create"), AgentCreatedTypeCreate)
	assert.Equal(t, AgentCreatedType("copy"), AgentCreatedTypeCopy)
	assert.Equal(t, AgentCreatedType("import"), AgentCreatedTypeImport)
}

func TestAgentCreatedType_EnumCheck_Valid(t *testing.T) {
	t.Parallel()

	validTypes := []AgentCreatedType{
		AgentCreatedTypeCreate,
		AgentCreatedTypeCopy,
		AgentCreatedTypeImport,
	}

	for _, act := range validTypes {
		t.Run(string(act), func(t *testing.T) {
			t.Parallel()

			err := act.EnumCheck()
			assert.NoError(t, err)
		})
	}
}

func TestAgentCreatedType_EnumCheck_Invalid(t *testing.T) {
	t.Parallel()

	invalidType := AgentCreatedType("invalid_type")
	err := invalidType.EnumCheck()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid agent created type")
}

func TestAgentCreatedType_EnumCheck_Empty(t *testing.T) {
	t.Parallel()

	emptyType := AgentCreatedType("")
	err := emptyType.EnumCheck()
	assert.Error(t, err)
}

func TestAgentCreatedType_AllUnique(t *testing.T) {
	t.Parallel()

	createdTypes := []AgentCreatedType{
		AgentCreatedTypeCreate,
		AgentCreatedTypeCopy,
		AgentCreatedTypeImport,
	}

	uniqueTypes := make(map[AgentCreatedType]bool)
	for _, act := range createdTypes {
		assert.False(t, uniqueTypes[act], "Duplicate created type found: %s", act)
		uniqueTypes[act] = true
	}
}

func TestAgentCreatedType_StringValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		act      AgentCreatedType
		expected string
	}{
		{
			name:     "create type",
			act:      AgentCreatedTypeCreate,
			expected: "create",
		},
		{
			name:     "copy type",
			act:      AgentCreatedTypeCopy,
			expected: "copy",
		},
		{
			name:     "import type",
			act:      AgentCreatedTypeImport,
			expected: "import",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := string(tt.act)
			assert.Equal(t, tt.expected, result)
		})
	}
}

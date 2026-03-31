package daenum

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAgentTplCreatedType_Constants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, AgentTplCreatedType("copy_from_agent"), AgentTplCreatedTypeCopyFromAgent)
	assert.Equal(t, AgentTplCreatedType("copy_from_tpl"), AgentTplCreatedTypeCopyFromTpl)
}

func TestAgentTplCreatedType_EnumCheck_Valid(t *testing.T) {
	t.Parallel()

	validTypes := []AgentTplCreatedType{
		AgentTplCreatedTypeCopyFromAgent,
		AgentTplCreatedTypeCopyFromTpl,
	}

	for _, att := range validTypes {
		t.Run(string(att), func(t *testing.T) {
			t.Parallel()

			err := att.EnumCheck()
			assert.NoError(t, err)
		})
	}
}

func TestAgentTplCreatedType_EnumCheck_Invalid(t *testing.T) {
	t.Parallel()

	invalidType := AgentTplCreatedType("invalid_type")
	err := invalidType.EnumCheck()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "无效的AgentTplCreatedType")
}

func TestAgentTplCreatedType_EnumCheck_Empty(t *testing.T) {
	t.Parallel()

	emptyType := AgentTplCreatedType("")
	err := emptyType.EnumCheck()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "无效的AgentTplCreatedType")
}

func TestAgentTplCreatedType_AllUnique(t *testing.T) {
	t.Parallel()

	createdTypes := []AgentTplCreatedType{
		AgentTplCreatedTypeCopyFromAgent,
		AgentTplCreatedTypeCopyFromTpl,
	}

	uniqueTypes := make(map[AgentTplCreatedType]bool)
	for _, att := range createdTypes {
		assert.False(t, uniqueTypes[att], "Duplicate created type found: %s", att)
		uniqueTypes[att] = true
	}
}

func TestAgentTplCreatedType_StringValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		act      AgentTplCreatedType
		expected string
	}{
		{
			name:     "copy from agent type",
			act:      AgentTplCreatedTypeCopyFromAgent,
			expected: "copy_from_agent",
		},
		{
			name:     "copy from tpl type",
			act:      AgentTplCreatedTypeCopyFromTpl,
			expected: "copy_from_tpl",
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

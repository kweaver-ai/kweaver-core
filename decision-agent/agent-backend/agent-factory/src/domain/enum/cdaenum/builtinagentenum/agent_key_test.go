package builtinagentenum

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAgentKey_EnumCheck_Valid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		key  AgentKey
	}{
		{"doc qa", AgentKeyDocQA},
		{"graph qa", AgentKeyGraphQA},
		{"online search", AgentKeyOnlineSearch},
		{"plan", AgentKeyPlan},
		{"simple chat", AgentKeySimpleChat},
		{"summary", AgentKeySummary},
		{"deep search", AgentKeyDeepSearch},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.key.EnumCheck()
			assert.NoError(t, err)
		})
	}
}

func TestAgentKey_EnumCheck_Invalid(t *testing.T) {
	t.Parallel()

	invalidKey := AgentKey("invalid_agent_key")
	err := invalidKey.EnumCheck()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid agent key")
}

func TestAgentKey_String(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "DocQA_Agent", AgentKeyDocQA.String())
	assert.Equal(t, "GraphQA_Agent", AgentKeyGraphQA.String())
	assert.Equal(t, "OnlineSearch_Agent", AgentKeyOnlineSearch.String())
	assert.Equal(t, "Plan_Agent", AgentKeyPlan.String())
	assert.Equal(t, "SimpleChat_Agent", AgentKeySimpleChat.String())
	assert.Equal(t, "Summary_Agent", AgentKeySummary.String())
	assert.Equal(t, "deepsearch", AgentKeyDeepSearch.String())
}

func TestAgentKey_IsDocQA(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		key      AgentKey
		expected bool
	}{
		{"is doc qa", AgentKeyDocQA, true},
		{"is not doc qa", AgentKeyGraphQA, false},
		{"is not doc qa - plan", AgentKeyPlan, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.key.IsDocQA()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAgentKey_IsGraphQA(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		key      AgentKey
		expected bool
	}{
		{"is graph qa", AgentKeyGraphQA, true},
		{"is not graph qa", AgentKeyDocQA, false},
		{"is not graph qa - plan", AgentKeyPlan, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.key.IsGraphQA()
			assert.Equal(t, tt.expected, result)
		})
	}
}

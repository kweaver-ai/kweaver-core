package chat_enum

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChatScenarioType_EnumCheck_Valid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		scenario ChatScenarioType
	}{
		{"adp chat page", ChatScenarioADPChatPage},
		{"adp agent debug", ChatScenarioADPAgentDebug},
		{"adp data flow", ChatScenarioADPDataFlow},
		{"third system", ChatScenarioThirdSystem},
		{"custom", ChatScenarioCustom},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.scenario.EnumCheck()
			assert.NoError(t, err)
		})
	}
}

func TestChatScenarioType_EnumCheck_Invalid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		scenario ChatScenarioType
	}{
		{"empty scenario", ""},
		{"invalid scenario", "invalid_scenario"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.scenario.EnumCheck()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "对话场景类型不合法")
		})
	}
}

func TestChatScenarioType_ToString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		scenario ChatScenarioType
		expected string
	}{
		{"adp chat page", ChatScenarioADPChatPage, "ADP_chat_page"},
		{"adp agent debug", ChatScenarioADPAgentDebug, "ADP_agent_debug"},
		{"adp data flow", ChatScenarioADPDataFlow, "ADP_data_flow"},
		{"third system", ChatScenarioThirdSystem, "third_system"},
		{"custom", ChatScenarioCustom, "custom"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.scenario.ToString()
			assert.Equal(t, tt.expected, result)
		})
	}
}

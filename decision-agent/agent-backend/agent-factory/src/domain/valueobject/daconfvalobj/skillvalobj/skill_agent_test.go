package skillvalobj

import (
	"encoding/json"
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum/skillenum"
	"github.com/stretchr/testify/assert"
)

func TestCurrentPmsCheckStatusT_Constants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, CurrentPmsCheckStatusT("success"), CurrentPmsCheckStatusSuccess)
	assert.Equal(t, CurrentPmsCheckStatusT("failed"), CurrentPmsCheckStatusFailed)
}

func TestSkillAgent_ValObjCheck_Valid(t *testing.T) {
	t.Parallel()

	agent := &SkillAgent{
		AgentKey: "test-agent-key",
	}

	err := agent.ValObjCheck()

	assert.NoError(t, err)
}

func TestSkillAgent_ValObjCheck_EmptyAgentKey(t *testing.T) {
	t.Parallel()

	agent := &SkillAgent{
		AgentKey: "",
	}

	err := agent.ValObjCheck()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "agent_key is required")
}

func TestSkillAgent_NewSkillAgent(t *testing.T) {
	t.Parallel()

	inputData := json.RawMessage(`{"key": "value"}`)
	agent := &SkillAgent{
		AgentKey:                        "agent-key-123",
		AgentVersion:                    "v1.0",
		AgentInput:                      inputData,
		Intervention:                    true,
		InterventionConfirmationMessage: "Please confirm",
		AgentTimeout:                    30,
	}

	assert.Equal(t, "agent-key-123", agent.AgentKey)
	assert.Equal(t, "v1.0", agent.AgentVersion)
	assert.Equal(t, inputData, agent.AgentInput)
	assert.True(t, agent.Intervention)
	assert.Equal(t, "Please confirm", agent.InterventionConfirmationMessage)
	assert.Equal(t, 30, agent.AgentTimeout)
}

func TestSkillAgent_Empty(t *testing.T) {
	t.Parallel()

	agent := &SkillAgent{}

	assert.Empty(t, agent.AgentKey)
	assert.Empty(t, agent.AgentVersion)
	assert.False(t, agent.Intervention)
	assert.Empty(t, agent.InterventionConfirmationMessage)
	assert.Equal(t, 0, agent.AgentTimeout)
}

func TestSkillAgent_WithDataSourceConfig(t *testing.T) {
	t.Parallel()

	dsConfig := &DataSourceConfig{
		Type:            skillenum.Datasource("knowledge"),
		SpecificInherit: skillenum.DatasourceSpecificInherit("none"),
	}
	agent := &SkillAgent{
		AgentKey:         "test-key",
		DataSourceConfig: dsConfig,
	}

	assert.NotNil(t, agent.DataSourceConfig)
	assert.Equal(t, skillenum.Datasource("knowledge"), dsConfig.Type)
}

func TestSkillAgent_WithLLMConfig(t *testing.T) {
	t.Parallel()

	llmConfig := &LLMConfig{
		Type: skillenum.LLM("openai"),
	}
	agent := &SkillAgent{
		AgentKey:  "test-key",
		LlmConfig: llmConfig,
	}

	assert.NotNil(t, agent.LlmConfig)
	assert.Equal(t, skillenum.LLM("openai"), llmConfig.Type)
}

func TestSkillAgent_CurrentPmsCheckStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status CurrentPmsCheckStatusT
	}{
		{"Success status", CurrentPmsCheckStatusSuccess},
		{"Failed status", CurrentPmsCheckStatusFailed},
		{"Empty status", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			agent := &SkillAgent{
				AgentKey:              "test-key",
				CurrentPmsCheckStatus: tt.status,
			}

			assert.Equal(t, tt.status, agent.CurrentPmsCheckStatus)
		})
	}
}

func TestSkillAgent_CurrentIsExistsAndPublished(t *testing.T) {
	t.Parallel()

	agent := &SkillAgent{
		AgentKey:                    "test-key",
		CurrentIsExistsAndPublished: true,
	}

	assert.True(t, agent.CurrentIsExistsAndPublished)
}

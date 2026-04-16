package swagger

import (
	"encoding/json"
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/daconfvalobj"
	"github.com/stretchr/testify/require"
)

func TestAgentConfigConfig_JSONContainsNonDolphinModeConfig(t *testing.T) {
	t.Parallel()

	model := AgentConfigConfig{
		NonDolphinModeConfig: &daconfvalobj.NonDolphinModeConfig{
			DisableHistoryInAConversation: true,
			DisableLLMCache:               true,
		},
	}

	data, err := json.Marshal(model)
	require.NoError(t, err)
	require.Contains(t, string(data), "non_dolphin_mode_config")
	require.Contains(t, string(data), "disable_history_in_a_conversation")
	require.Contains(t, string(data), "disable_llm_cache")
}

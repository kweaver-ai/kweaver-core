package swagger

import (
	"encoding/json"
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/daconfvalobj"
	"github.com/stretchr/testify/require"
)

func TestAgentConfigConfig_JSONContainsModeAndReactConfig(t *testing.T) {
	t.Parallel()

	model := AgentConfigConfig{
		Mode: cdaenum.AgentModeReact,
		ReactConfig: &daconfvalobj.ReactConfig{
			DisableHistoryInAConversation: true,
			DisableLLMCache:               true,
		},
	}

	data, err := json.Marshal(model)
	require.NoError(t, err)

	var payload map[string]any
	err = json.Unmarshal(data, &payload)
	require.NoError(t, err)

	require.Equal(t, "react", payload["mode"])

	config, ok := payload["react_config"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, true, config["disable_history_in_a_conversation"])
	require.Equal(t, true, payload["react_config"] != nil)
}

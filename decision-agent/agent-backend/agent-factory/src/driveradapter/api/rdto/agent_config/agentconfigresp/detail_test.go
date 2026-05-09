package agentconfigresp

import (
	"encoding/json"
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/entity/daconfeo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/daconfvalobj"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDetailRes(t *testing.T) {
	t.Parallel()

	res := NewDetailRes()

	assert.NotNil(t, res)
	assert.Empty(t, res.ID)
	assert.Empty(t, res.Key)
	assert.Empty(t, res.Name)
	assert.Empty(t, res.Profile)
	assert.Empty(t, res.Avatar)
	assert.Empty(t, res.ProductKey)
	assert.Empty(t, res.ProductName)
	assert.Nil(t, res.Config)
	assert.False(t, res.IsPublished)
}

func TestDetailRes_LoadFromEo(t *testing.T) {
	t.Parallel()

	t.Run("load from valid entity", func(t *testing.T) {
		t.Parallel()

		res := NewDetailRes()

		builtInYes := cdaenum.BuiltInYes
		systemAgentYes := cenum.YesNoInt8Yes
		isDefault := true
		eo := &daconfeo.DataAgent{
			DataAgentPo: dapo.DataAgentPo{
				ID:            "agent-123",
				Key:           "test-agent",
				Name:          "Test Agent",
				Profile:       strPtr("Test profile"),
				AvatarType:    cdaenum.AvatarTypeBuiltIn,
				Avatar:        "🤖",
				ProductKey:    "product-key",
				IsBuiltIn:     &builtInYes,
				IsSystemAgent: &systemAgentYes,
				Status:        cdaenum.StatusPublished,
			},
			Config: &daconfvalobj.Config{
				Input:  &daconfvalobj.Input{},
				Output: &daconfvalobj.Output{},
				Llms: []*daconfvalobj.LlmItem{
					{
						IsDefault: isDefault,
						LlmConfig: &daconfvalobj.LlmConfig{
							Name: "gpt-4",
						},
					},
				},
			},
		}

		err := res.LoadFromEo(eo)

		require.NoError(t, err)
		assert.Equal(t, "agent-123", res.ID)
		assert.Equal(t, "test-agent", res.Key)
		assert.Equal(t, "Test Agent", res.Name)
		assert.Equal(t, "Test profile", res.Profile)
		assert.Equal(t, "🤖", res.Avatar)
		assert.Equal(t, "product-key", res.ProductKey)
		assert.Equal(t, cdaenum.AvatarTypeBuiltIn, res.AvatarType)
		assert.Equal(t, cdaenum.BuiltInYes, res.IsBuiltIn)
		assert.NotNil(t, res.IsSystemAgent)
		assert.Equal(t, cenum.YesNoInt8Yes, *res.IsSystemAgent)
		assert.Equal(t, cdaenum.StatusPublished, res.Status)
		assert.NotNil(t, res.Config)
		assert.NotNil(t, res.Config.Llms)
	})

	t.Run("load from entity with nil pointers", func(t *testing.T) {
		t.Parallel()

		res := NewDetailRes()

		eo := &daconfeo.DataAgent{
			DataAgentPo: dapo.DataAgentPo{
				ID:            "agent-456",
				Key:           "minimal-agent",
				Name:          "Minimal Agent",
				Profile:       nil,
				IsBuiltIn:     nil,
				IsSystemAgent: nil,
				Status:        cdaenum.StatusUnpublished,
			},
			Config: &daconfvalobj.Config{
				Input:  &daconfvalobj.Input{},
				Output: &daconfvalobj.Output{},
			},
		}

		err := res.LoadFromEo(eo)

		require.NoError(t, err)
		assert.Equal(t, "agent-456", res.ID)
		assert.Equal(t, "minimal-agent", res.Key)
		assert.Equal(t, "Minimal Agent", res.Name)
		assert.Empty(t, res.Profile)
		assert.Nil(t, res.IsSystemAgent)
		assert.Equal(t, cdaenum.StatusUnpublished, res.Status)
		assert.NotNil(t, res.Config)
		// Note: IsBuiltIn is copied by value, so nil pointers become zero values
	})

	t.Run("load from entity with config", func(t *testing.T) {
		t.Parallel()

		res := NewDetailRes()

		eo := &daconfeo.DataAgent{
			DataAgentPo: dapo.DataAgentPo{
				ID:     "agent-789",
				Key:    "config-agent",
				Name:   "Config Agent",
				Status: cdaenum.StatusPublished,
			},
			Config: &daconfvalobj.Config{
				Input: &daconfvalobj.Input{
					Fields: daconfvalobj.Fields{
						&daconfvalobj.Field{
							Name: "field1",
							Type: cdaenum.InputFieldTypeString,
						},
					},
				},
				Output: &daconfvalobj.Output{},
			},
		}

		err := res.LoadFromEo(eo)

		require.NoError(t, err)
		assert.Equal(t, "agent-789", res.ID)
		assert.NotNil(t, res.Config)
		assert.NotNil(t, res.Config.Input)
		assert.Len(t, res.Config.Input.Fields, 1)
		assert.Equal(t, "field1", res.Config.Input.Fields[0].Name)
	})
}

func TestDetailRes_Fields(t *testing.T) {
	t.Parallel()

	res := &DetailRes{
		ID:          "test-id",
		Key:         "test-key",
		Name:        "Test Name",
		Profile:     "Test Profile",
		AvatarType:  cdaenum.AvatarTypeUserUploaded,
		Avatar:      "avatar.png",
		ProductKey:  "product-1",
		ProductName: "Product 1",
		Status:      cdaenum.StatusPublished,
		IsPublished: true,
	}

	assert.Equal(t, "test-id", res.ID)
	assert.Equal(t, "test-key", res.Key)
	assert.Equal(t, "Test Name", res.Name)
	assert.Equal(t, "Test Profile", res.Profile)
	assert.Equal(t, cdaenum.AvatarTypeUserUploaded, res.AvatarType)
	assert.Equal(t, "avatar.png", res.Avatar)
	assert.Equal(t, "product-1", res.ProductKey)
	assert.Equal(t, "Product 1", res.ProductName)
	assert.Equal(t, cdaenum.StatusPublished, res.Status)
	assert.True(t, res.IsPublished)
}

func TestDetailRes_WithChineseCharacters(t *testing.T) {
	t.Parallel()

	res := NewDetailRes()

	eo := &daconfeo.DataAgent{
		DataAgentPo: dapo.DataAgentPo{
			ID:      "zh-cn-id",
			Key:     "zhongwen-key",
			Name:    "中文智能体",
			Profile: strPtr("中文描述"),
		},
		Config: &daconfvalobj.Config{
			Input:  &daconfvalobj.Input{},
			Output: &daconfvalobj.Output{},
		},
	}

	err := res.LoadFromEo(eo)

	require.NoError(t, err)
	assert.Equal(t, "中文智能体", res.Name)
	assert.Equal(t, "中文描述", res.Profile)
}

func TestDetailRes_LoadFromEo_UsesReactConfigFieldInJSON(t *testing.T) {
	t.Parallel()

	res := NewDetailRes()
	eo := &daconfeo.DataAgent{
		DataAgentPo: dapo.DataAgentPo{
			ID:   "react-agent",
			Key:  "react-agent",
			Name: "React Agent",
		},
		Config: &daconfvalobj.Config{
			Input:  &daconfvalobj.Input{},
			Output: &daconfvalobj.Output{},
			Mode:   cdaenum.AgentModeReact,
			ReactConfig: &daconfvalobj.ReactConfig{
				DisableHistoryInAConversation: true,
				DisableLLMCache:               true,
			},
		},
	}

	err := res.LoadFromEo(eo)
	require.NoError(t, err)

	data, err := json.Marshal(res)
	require.NoError(t, err)

	var payload map[string]any
	err = json.Unmarshal(data, &payload)
	require.NoError(t, err)

	config, ok := payload["config"].(map[string]any)
	require.True(t, ok)
	_, ok = config["react_config"]
	assert.True(t, ok)
}

func TestDetailRes_LoadFromEo_FillsDolphinModeOnResponseCopy(t *testing.T) {
	t.Parallel()

	res := NewDetailRes()
	eo := &daconfeo.DataAgent{
		DataAgentPo: dapo.DataAgentPo{
			ID:   "dolphin-agent",
			Key:  "dolphin-agent",
			Name: "Dolphin Agent",
		},
		Config: &daconfvalobj.Config{
			Input:         &daconfvalobj.Input{},
			Output:        &daconfvalobj.Output{},
			IsDolphinMode: cdaenum.DolphinModeEnabled,
		},
	}

	err := res.LoadFromEo(eo)
	require.NoError(t, err)

	require.NotNil(t, res.Config)
	assert.Equal(t, cdaenum.AgentModeDolphin, res.Config.Mode)
	assert.Equal(t, cdaenum.AgentModeDolphin, res.Config.GetMode())
	assert.Empty(t, eo.Config.Mode)
	assert.NotSame(t, eo.Config, res.Config)
}

func TestDetailRes_LoadFromEo_FillsDefaultModeOnResponseCopy(t *testing.T) {
	t.Parallel()

	res := NewDetailRes()
	eo := &daconfeo.DataAgent{
		DataAgentPo: dapo.DataAgentPo{
			ID:   "default-agent",
			Key:  "default-agent",
			Name: "Default Agent",
		},
		Config: &daconfvalobj.Config{
			Input:         &daconfvalobj.Input{},
			Output:        &daconfvalobj.Output{},
			IsDolphinMode: cdaenum.DolphinModeDisabled,
		},
	}

	err := res.LoadFromEo(eo)
	require.NoError(t, err)

	require.NotNil(t, res.Config)
	assert.Equal(t, cdaenum.AgentModeDefault, res.Config.Mode)
	assert.Equal(t, cdaenum.AgentModeDefault, res.Config.GetMode())
	assert.Empty(t, eo.Config.Mode)
	assert.NotSame(t, eo.Config, res.Config)
}

// Helper function
func strPtr(s string) *string {
	return &s
}

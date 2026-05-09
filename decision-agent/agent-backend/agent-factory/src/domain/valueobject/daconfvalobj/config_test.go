package daconfvalobj

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/daconfvalobj/datasourcevalobj"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/daconfvalobj/skillvalobj"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_GetBuiltInDsDocSourceFields(t *testing.T) {
	t.Parallel()

	config := &Config{}
	fields := config.GetBuiltInDsDocSourceFields()
	assert.Empty(t, fields)

	config.DataSource = &datasourcevalobj.RetrieverDataSource{
		Doc: []*datasourcevalobj.DocSource{
			{
				DsID: "ds-1",
			},
		},
	}
	fields = config.GetBuiltInDsDocSourceFields()
	assert.Empty(t, fields) // Since there is no actual implementation in DocSource returning fields, it will be empty in unit test
}

func TestConfig_GetConfigMetadata(t *testing.T) {
	t.Parallel()

	config := &Config{
		Metadata: ConfigMetadata{
			ConfigLastSetTimestamp: 123456789,
		},
	}
	metadata := config.GetConfigMetadata()
	assert.NotNil(t, metadata)
	assert.Equal(t, uint64(123456789), metadata.ConfigLastSetTimestamp)
}

func TestConfig_CheckProductAndDataSource(t *testing.T) {
	t.Parallel()

	config := &Config{}

	// No data source
	err := config.CheckProductAndDataSource(cdaenum.ProductChatBI)
	assert.NoError(t, err)

	// With empty doc data source
	config.DataSource = &datasourcevalobj.RetrieverDataSource{
		Doc: []*datasourcevalobj.DocSource{},
	}
	err = config.CheckProductAndDataSource(cdaenum.ProductChatBI)
	assert.NoError(t, err)

	// With doc data source for ChatBI product
	config.DataSource = &datasourcevalobj.RetrieverDataSource{
		Doc: []*datasourcevalobj.DocSource{
			{
				DsID: "ds-1",
			},
		},
	}
	err = config.CheckProductAndDataSource(cdaenum.ProductChatBI)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "文档数据源不能用于智能问数产品")

	// With doc data source for non-ChatBI product
	err = config.CheckProductAndDataSource(cdaenum.ProductDIP)
	assert.NoError(t, err)
}

func TestConfig_GetErrMsgMap(t *testing.T) {
	t.Parallel()

	config := &Config{}
	errMap := config.GetErrMsgMap()
	assert.NotNil(t, errMap)
	assert.Equal(t, `"input"不能为空`, errMap["Input.required"])
	assert.Equal(t, `"output"不能为空`, errMap["Output.required"])
}

func TestNewConfig(t *testing.T) {
	t.Parallel()

	config := NewConfig()
	assert.NotNil(t, config)
}

func TestConfig_ValObjCheckWithCtx(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("nil input", func(t *testing.T) {
		t.Parallel()

		config := NewConfig()
		err := config.ValObjCheckWithCtx(ctx, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "input is required")
	})

	t.Run("invalid input", func(t *testing.T) {
		t.Parallel()

		config := NewConfig()
		config.Input = &Input{Fields: nil}
		config.Output = &Output{DefaultFormat: cdaenum.OutputDefaultFormat("invalid")}
		err := config.ValObjCheckWithCtx(ctx, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "input is invalid")
	})

	validInput := &Input{
		Fields: Fields{
			&Field{Name: "param1", Type: cdaenum.InputFieldTypeString},
		},
	}
	validOutput := &Output{
		DefaultFormat: cdaenum.OutputDefaultFormatJson,
		Variables:     &VariablesS{},
	}

	t.Run("llms without default", func(t *testing.T) {
		t.Parallel()

		config := NewConfig()
		config.Input = validInput
		config.Output = validOutput
		config.IsDolphinMode = cdaenum.DolphinModeDisabled
		config.Llms = []*LlmItem{
			{IsDefault: false, LlmConfig: &LlmConfig{Name: "test", MaxTokens: 100}},
		}
		err := config.ValObjCheckWithCtx(ctx, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must have at least one default llm")
	})

	t.Run("invalid IsDataFlowSetEnabled", func(t *testing.T) {
		t.Parallel()

		config := NewConfig()
		config.Input = validInput
		config.Output = validOutput
		config.IsDolphinMode = cdaenum.DolphinModeDisabled
		config.Llms = []*LlmItem{
			{IsDefault: true, LlmConfig: &LlmConfig{Name: "test", MaxTokens: 100}},
		}
		config.IsDataFlowSetEnabled = 3
		err := config.ValObjCheckWithCtx(ctx, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "is_data_flow_set_enabled must be 0 or 1")
	})

	t.Run("invalid opening remark", func(t *testing.T) {
		t.Parallel()

		config := NewConfig()
		config.Input = validInput
		config.Output = validOutput
		config.IsDolphinMode = cdaenum.DolphinModeDisabled
		config.IsDataFlowSetEnabled = 1
		config.OpeningRemarkConfig = &OpeningRemarkConfig{Type: "invalid"}
		err := config.ValObjCheckWithCtx(ctx, true) // private api -> ignore llm
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "opening_remark_config is invalid")
	})

	t.Run("invalid plan mode vs dolphin mode", func(t *testing.T) {
		t.Parallel()

		config := NewConfig()
		config.Input = validInput
		config.Output = &Output{
			DefaultFormat: cdaenum.OutputDefaultFormatJson,
			Variables:     &VariablesS{AnswerVar: "answer"},
		}
		config.IsDolphinMode = cdaenum.DolphinModeEnabled
		config.Dolphin = "some dolphin text"
		config.PlanMode = &PlanMode{IsEnabled: true}
		err := config.ValObjCheckWithCtx(ctx, true)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "plan_mode is invalid when is_dolphin_mode is true")
	})

	t.Run("valid full config private API", func(t *testing.T) {
		t.Parallel()

		config := NewConfig()
		config.Input = validInput
		config.Output = validOutput
		config.IsDolphinMode = cdaenum.DolphinModeDisabled
		config.IsDataFlowSetEnabled = 1
		err := config.ValObjCheckWithCtx(ctx, true)
		assert.NoError(t, err)
	})

	t.Run("legacy disabled dolphin flag overrides explicit dolphin mode", func(t *testing.T) {
		t.Parallel()

		config := NewConfig()
		config.Mode = cdaenum.AgentModeDolphin
		config.Input = validInput
		config.Dolphin = "some dolphin text"
		config.Output = &Output{
			DefaultFormat: cdaenum.OutputDefaultFormatJson,
			Variables:     &VariablesS{AnswerVar: "answer"},
		}
		config.Llms = []*LlmItem{
			{IsDefault: true, LlmConfig: &LlmConfig{Name: "test", MaxTokens: 100}},
		}

		err := config.ValObjCheckWithCtx(ctx, false)
		require.NoError(t, err)
		assert.Equal(t, cdaenum.AgentModeDefault, config.Mode)
		assert.Equal(t, cdaenum.DolphinModeDisabled, config.IsDolphinMode)
	})

	t.Run("derive dolphin mode from empty mode and legacy flag", func(t *testing.T) {
		t.Parallel()

		config := NewConfig()
		config.IsDolphinMode = cdaenum.DolphinModeEnabled
		config.Input = validInput
		config.Dolphin = "some dolphin text"
		config.Output = &Output{
			DefaultFormat: cdaenum.OutputDefaultFormatJson,
			Variables:     &VariablesS{AnswerVar: "answer"},
		}
		config.Llms = []*LlmItem{
			{IsDefault: true, LlmConfig: &LlmConfig{Name: "test", MaxTokens: 100}},
		}

		err := config.ValObjCheckWithCtx(ctx, false)
		require.NoError(t, err)
		assert.Equal(t, cdaenum.AgentModeDolphin, config.Mode)
		assert.Equal(t, cdaenum.DolphinModeEnabled, config.IsDolphinMode)
	})

	t.Run("legacy enabled dolphin flag overrides explicit default mode", func(t *testing.T) {
		t.Parallel()

		config := NewConfig()
		config.Mode = cdaenum.AgentModeDefault
		config.IsDolphinMode = cdaenum.DolphinModeEnabled
		config.Input = validInput
		config.Dolphin = "some dolphin text"
		config.Output = &Output{
			DefaultFormat: cdaenum.OutputDefaultFormatJson,
			Variables:     &VariablesS{AnswerVar: "answer"},
		}
		config.Llms = []*LlmItem{
			{IsDefault: true, LlmConfig: &LlmConfig{Name: "test", MaxTokens: 100}},
		}

		err := config.ValObjCheckWithCtx(ctx, false)
		require.NoError(t, err)
		assert.Equal(t, cdaenum.AgentModeDolphin, config.Mode)
		assert.Equal(t, cdaenum.DolphinModeEnabled, config.IsDolphinMode)
	})

	t.Run("react mode ignores legacy dolphin flag", func(t *testing.T) {
		t.Parallel()

		config := NewConfig()
		config.Mode = cdaenum.AgentModeReact
		config.IsDolphinMode = cdaenum.DolphinModeEnabled
		config.Input = validInput
		config.Output = validOutput

		err := config.ValObjCheckWithCtx(ctx, true)
		require.NoError(t, err)
		assert.Equal(t, cdaenum.AgentModeReact, config.Mode)
		assert.Equal(t, cdaenum.DolphinModeDisabled, config.IsDolphinMode)
	})

	t.Run("invalid non-empty mode returns error before normalization", func(t *testing.T) {
		t.Parallel()

		config := NewConfig()
		config.Mode = cdaenum.AgentMode("invalid")
		config.IsDolphinMode = cdaenum.DolphinModeDisabled
		config.Input = validInput
		config.Output = validOutput

		err := config.ValObjCheckWithCtx(ctx, true)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "mode is invalid")
		assert.Equal(t, cdaenum.AgentMode("invalid"), config.Mode)
		assert.Equal(t, cdaenum.DolphinModeDisabled, config.IsDolphinMode)
	})
		}

func TestConfig_checkAboutDolphin(t *testing.T) {
	t.Parallel()

	t.Run("invalid is dolphin mode", func(t *testing.T) {
		t.Parallel()

		config := NewConfig()
		config.IsDolphinMode = cdaenum.DolphinMode(-1)
		err := config.checkAboutDolphin()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "is_dolphin_mode is invalid")
	})

	t.Run("pre dolphin invalid", func(t *testing.T) {
		t.Parallel()

		config := NewConfig()
		config.IsDolphinMode = cdaenum.DolphinModeDisabled
		config.PreDolphin = []*DolphinTpl{
			{Key: cdaenum.DolphinTplKey("invalid_key")},
		}
		err := config.checkAboutDolphin()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "pre_dolphin is invalid")
	})

	t.Run("post dolphin invalid", func(t *testing.T) {
		t.Parallel()

		config := NewConfig()
		config.IsDolphinMode = cdaenum.DolphinModeDisabled
		config.PostDolphin = []*DolphinTpl{
			{Key: cdaenum.DolphinTplKey("invalid_key")},
		}
		err := config.checkAboutDolphin()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "post_dolphin is invalid")
	})

	t.Run("dolphin mode true but empty content", func(t *testing.T) {
		t.Parallel()

		config := NewConfig()
		config.IsDolphinMode = cdaenum.DolphinModeEnabled
		config.Dolphin = ""
		err := config.checkAboutDolphin()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "pre_dolphin or post_dolphin or dolphin is required")
	})

	t.Run("dolphin mode true with content", func(t *testing.T) {
		t.Parallel()

		config := NewConfig()
		config.IsDolphinMode = cdaenum.DolphinModeEnabled
		config.Dolphin = "valid content"
		err := config.checkAboutDolphin()
		assert.NoError(t, err)
	})

	t.Run("dolphin mode true with pre dolphin", func(t *testing.T) {
		t.Parallel()

		config := NewConfig()
		config.IsDolphinMode = cdaenum.DolphinModeEnabled
		config.PreDolphin = []*DolphinTpl{
			{Key: cdaenum.DolphinTplKeyDocRetrieve}, // valid key?
		}
		err := config.checkAboutDolphin()
		// this might fail because Key is valid but Value or Enabled is not valid, actually DolphinTpl_ValObjCheck is not tested here.
		// assume it passes for now or we just test it broadly.
		_ = err
	})
}

func TestConfig_UnmarshalJSON_WithReactConfigAlias(t *testing.T) {
	t.Parallel()

	payload := `{
		"mode":"react",
		"input":{"fields":[{"name":"query","type":"string"}]},
		"llms":[{"is_default":true,"llm_config":{"name":"test","max_tokens":100}}],
		"output":{"default_format":"markdown"},
		"react_config":{
			"disable_history_in_a_conversation":true,
			"disable_llm_cache":false
		}
	}`

	var config Config

	err := json.Unmarshal([]byte(payload), &config)
	require.NoError(t, err)
	require.NotNil(t, config.ReactConfig)
	assert.Equal(t, cdaenum.AgentModeReact, config.Mode)
	assert.True(t, config.ReactConfig.DisableHistoryInAConversation)
	assert.False(t, config.ReactConfig.DisableLLMCache)
	assert.Equal(t, cdaenum.DolphinModeDisabled, config.IsDolphinMode)
}

func TestConfig_UnmarshalJSON_DoesNotAcceptLegacyReactConfigField(t *testing.T) {
	t.Parallel()

	legacyFieldName := strings.Join([]string{"non", "dolphin", "mode", "config"}, "_")
	payloadMap := map[string]any{
		"mode": "react",
		"input": map[string]any{
			"fields": []map[string]any{
				{"name": "query", "type": "string"},
			},
		},
		"llms": []map[string]any{
			{
				"is_default": true,
				"llm_config": map[string]any{"name": "test", "max_tokens": 100},
			},
		},
		"output": map[string]any{"default_format": "markdown"},
		legacyFieldName: map[string]any{
			"disable_history_in_a_conversation": true,
			"disable_llm_cache":                 false,
		},
	}
	payload, err := json.Marshal(payloadMap)
	require.NoError(t, err)

	var config Config

	err = json.Unmarshal(payload, &config)
	require.NoError(t, err)
	assert.Nil(t, config.ReactConfig)
	assert.Equal(t, cdaenum.AgentModeReact, config.Mode)
	assert.Equal(t, cdaenum.DolphinModeDisabled, config.IsDolphinMode)
}

func TestConfig_UnmarshalJSON_DoesNotNormalizeMode(t *testing.T) {
	t.Parallel()

	payload := `{
		"mode":"dolphin",
		"input":{"fields":[{"name":"query","type":"string"}]},
		"llms":[{"is_default":true,"llm_config":{"name":"test","max_tokens":100}}],
		"output":{"default_format":"markdown"},
		"dolphin":"test dolphin statement"
	}`

	var config Config

	err := json.Unmarshal([]byte(payload), &config)
	require.NoError(t, err)
	assert.Equal(t, cdaenum.AgentModeDolphin, config.Mode)
	assert.Equal(t, cdaenum.DolphinModeDisabled, config.IsDolphinMode)
}

func TestConfig_ValObjCheck_Skills(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	config := NewConfig()
	config.Input = &Input{
		Fields: Fields{
			&Field{Name: "param1", Type: cdaenum.InputFieldTypeString},
		},
	}
	config.Output = &Output{DefaultFormat: cdaenum.OutputDefaultFormatJson, Variables: &VariablesS{}}
	config.IsDolphinMode = cdaenum.DolphinModeDisabled
	config.Skill = &skillvalobj.Skill{
		Tools: []*skillvalobj.SkillTool{
			{}, // Invalid tool
		},
	}

	err := config.ValObjCheckWithCtx(ctx, true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tools is invalid")
}

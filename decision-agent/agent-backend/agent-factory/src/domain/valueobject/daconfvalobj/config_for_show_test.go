package daconfvalobj

import (
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/daconfvalobj/datasourcevalobj"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/daconfvalobj/skillvalobj"
	"github.com/stretchr/testify/assert"
)

func TestConfigForShow_New(t *testing.T) {
	t.Parallel()

	config := &ConfigForShow{
		Input: &Input{
			Fields: []*Field{},
		},
		Output: &Output{},
	}

	assert.NotNil(t, config)
	assert.NotNil(t, config.Input)
	assert.NotNil(t, config.Output)
}

func TestConfigForShow_WithSystemPrompt(t *testing.T) {
	t.Parallel()

	config := &ConfigForShow{
		Input:        &Input{Fields: []*Field{}},
		SystemPrompt: "You are a helpful assistant",
		Output:       &Output{},
	}

	assert.Equal(t, "You are a helpful assistant", config.SystemPrompt)
}

func TestConfigForShow_WithDolphin(t *testing.T) {
	t.Parallel()

	config := &ConfigForShow{
		Input:   &Input{Fields: []*Field{}},
		Dolphin: "dolphin prompt template",
		Output:  &Output{},
	}

	assert.Equal(t, "dolphin prompt template", config.Dolphin)
}

func TestConfigForShow_WithDolphinMode(t *testing.T) {
	t.Parallel()

	config := &ConfigForShow{
		Input:         &Input{Fields: []*Field{}},
		IsDolphinMode: cdaenum.DolphinModeEnabled,
		Output:        &Output{},
	}

	assert.Equal(t, cdaenum.DolphinModeEnabled, config.IsDolphinMode)
}

func TestConfigForShow_WithDataSource(t *testing.T) {
	t.Parallel()

	ds := &datasourcevalobj.RetrieverDataSource{
		Doc: []*datasourcevalobj.DocSource{},
	}

	config := &ConfigForShow{
		Input:      &Input{Fields: []*Field{}},
		DataSource: ds,
		Output:     &Output{},
	}

	assert.NotNil(t, config.DataSource)
	assert.NotNil(t, config.Input)
}

func TestConfigForShow_WithSkill(t *testing.T) {
	t.Parallel()

	skill := &skillvalobj.Skill{
		Agents: []*skillvalobj.SkillAgent{},
		Tools:  []*skillvalobj.SkillTool{},
	}

	config := &ConfigForShow{
		Input:  &Input{Fields: []*Field{}},
		Skill:  skill,
		Output: &Output{},
	}

	assert.NotNil(t, config.Skill)
}

func TestConfigForShow_WithLlms(t *testing.T) {
	t.Parallel()

	llms := []*LlmItem{
		{
			LlmConfig: &LlmConfig{},
		},
	}

	config := &ConfigForShow{
		Input:  &Input{Fields: []*Field{}},
		Llms:   llms,
		Output: &Output{},
	}

	assert.Len(t, config.Llms, 1)
}

func TestConfigForShow_WithOpeningRemarkConfig(t *testing.T) {
	t.Parallel()

	config := &ConfigForShow{
		Input:               &Input{Fields: []*Field{}},
		OpeningRemarkConfig: &OpeningRemarkConfig{},
		Output:              &Output{},
	}

	assert.NotNil(t, config.OpeningRemarkConfig)
}

func TestConfigForShow_WithPresetQuestions(t *testing.T) {
	t.Parallel()

	questions := []*PresetQuestion{
		{
			Question: "What is this?",
		},
	}

	config := &ConfigForShow{
		Input:           &Input{Fields: []*Field{}},
		PresetQuestions: questions,
		Output:          &Output{},
	}

	assert.Len(t, config.PresetQuestions, 1)
	assert.Equal(t, "What is this?", config.PresetQuestions[0].Question)
}

func TestConfigForShow_AllFields(t *testing.T) {
	t.Parallel()

	config := &ConfigForShow{
		Input:                &Input{Fields: []*Field{}},
		SystemPrompt:         "system prompt",
		Dolphin:              "dolphin template",
		IsDolphinMode:        cdaenum.DolphinModeEnabled,
		DataSource:           &datasourcevalobj.RetrieverDataSource{},
		Skill:                &skillvalobj.Skill{},
		Llms:                 []*LlmItem{},
		IsDataFlowSetEnabled: 1,
		OpeningRemarkConfig:  &OpeningRemarkConfig{},
		PresetQuestions:      []*PresetQuestion{},
		Output:               &Output{},
	}

	assert.NotNil(t, config.Input)
	assert.NotNil(t, config.DataSource)
	assert.NotNil(t, config.Skill)
	assert.NotNil(t, config.Output)
	assert.Equal(t, 1, config.IsDataFlowSetEnabled)
}

package daconfvalobj

import (
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/stretchr/testify/assert"
)

func TestLlmItem_ValObjCheck_Valid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		item *LlmItem
	}{
		{
			name: "valid llm item with default",
			item: &LlmItem{
				IsDefault: true,
				LlmConfig: &LlmConfig{
					ID:        "llm-123",
					Name:      "Test LLM",
					ModelType: cdaenum.ModelTypeLlm,
					MaxTokens: 1000,
				},
			},
		},
		{
			name: "valid llm item without default",
			item: &LlmItem{
				IsDefault: false,
				LlmConfig: &LlmConfig{
					ID:        "llm-456",
					Name:      "Another LLM",
					ModelType: cdaenum.ModelTypeLlm,
					MaxTokens: 500,
				},
			},
		},
		{
			name: "valid llm item with zero max tokens (should be set to 500)",
			item: &LlmItem{
				IsDefault: true,
				LlmConfig: &LlmConfig{
					ID:        "llm-789",
					Name:      "Test LLM",
					ModelType: cdaenum.ModelTypeLlm,
					MaxTokens: 0,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.item.ValObjCheck()
			assert.NoError(t, err)
		})
	}
}

func TestLlmItem_ValObjCheck_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		item        *LlmItem
		expectedErr string
	}{
		{
			name: "nil llm config",
			item: &LlmItem{
				LlmConfig: nil,
			},
			expectedErr: "llm_config is required",
		},
		{
			name: "invalid llm config - empty name",
			item: &LlmItem{
				LlmConfig: &LlmConfig{
					ID:        "llm-123",
					Name:      "",
					ModelType: cdaenum.ModelTypeLlm,
					MaxTokens: 100,
				},
			},
			expectedErr: "llm_config is invalid",
		},
		{
			name: "invalid llm config - invalid model type",
			item: &LlmItem{
				LlmConfig: &LlmConfig{
					ID:        "llm-456",
					Name:      "Test",
					ModelType: "invalid_type",
					MaxTokens: 100,
				},
			},
			expectedErr: "llm_config is invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.item.ValObjCheck()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestLlmItem_ValObjCheck_Nil(t *testing.T) {
	t.Parallel()

	var item *LlmItem
	// Nil pointer will panic, so we test for that
	assert.Panics(t, func() {
		item.ValObjCheck() //nolint:errcheck
	})
}

func TestLlmItem_Fields(t *testing.T) {
	t.Parallel()

	config := &LlmConfig{
		ID:               "llm-789",
		Name:             "Test LLM Config",
		ModelType:        cdaenum.ModelTypeLlm,
		Temperature:      0.7,
		MaxTokens:        1000,
		TopP:             0.9,
		FrequencyPenalty: 0.5,
	}

	item := &LlmItem{
		IsDefault: true,
		LlmConfig: config,
	}

	assert.True(t, item.IsDefault)
	assert.Equal(t, config, item.LlmConfig)
	assert.Equal(t, "llm-789", item.LlmConfig.ID)
	assert.Equal(t, "Test LLM Config", item.LlmConfig.Name)
}

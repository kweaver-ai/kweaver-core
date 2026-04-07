package daconfvalobj

import (
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/stretchr/testify/assert"
)

func TestLlmConfig_ValObjCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		config  *LlmConfig
		wantErr bool
	}{
		{
			name: "完整配置",
			config: &LlmConfig{
				ID:               "model-123",
				Name:             "gpt-4",
				ModelType:        cdaenum.ModelTypeLlm,
				Temperature:      0.7,
				TopP:             0.9,
				TopK:             40,
				FrequencyPenalty: 0.0,
				PresencePenalty:  0.0,
				MaxTokens:        2048,
			},
			wantErr: false,
		},
		{
			name: "MaxTokens为0时设置默认值",
			config: &LlmConfig{
				ID:        "model-456",
				Name:      "gpt-3.5",
				ModelType: cdaenum.ModelTypeRlm,
				MaxTokens: 0,
			},
			wantErr: false,
		},
		{
			name: "Temperature超出范围",
			config: &LlmConfig{
				ID:          "model-789",
				Name:        "gpt-4",
				ModelType:   cdaenum.ModelTypeLlm,
				Temperature: 3.0,
				MaxTokens:   2048,
			},
			wantErr: true,
		},
		{
			name: "TopP超出范围",
			config: &LlmConfig{
				ID:        "model-101",
				Name:      "gpt-4",
				ModelType: cdaenum.ModelTypeLlm,
				TopP:      2.0,
				MaxTokens: 2048,
			},
			wantErr: true,
		},
		{
			name: "MaxTokens为负数",
			config: &LlmConfig{
				ID:        "model-202",
				Name:      "gpt-4",
				ModelType: cdaenum.ModelTypeLlm,
				MaxTokens: -1,
			},
			wantErr: true,
		},
		{
			name: "ModelType为空时设置默认值",
			config: &LlmConfig{
				ID:        "model-303",
				Name:      "gpt-4",
				ModelType: "",
				MaxTokens: 2048,
			},
			wantErr: false,
		},
		{
			name: "无效的ModelType",
			config: &LlmConfig{
				ID:        "model-404",
				Name:      "gpt-4",
				ModelType: cdaenum.ModelType("invalid"),
				MaxTokens: 2048,
			},
			wantErr: true,
		},
		{
			name: "Name为空",
			config: &LlmConfig{
				ID:        "model-505",
				Name:      "",
				ModelType: cdaenum.ModelTypeLlm,
				MaxTokens: 2048,
			},
			wantErr: true,
		},
		{
			name: "边界值测试-最小参数",
			config: &LlmConfig{
				ID:               "model-606",
				Name:             "gpt-4",
				ModelType:        cdaenum.ModelTypeLlm,
				Temperature:      0.0,
				TopP:             0.0,
				TopK:             0,
				FrequencyPenalty: -2.0,
				PresencePenalty:  -2.0,
				MaxTokens:        1,
			},
			wantErr: false,
		},
		{
			name: "边界值测试-最大参数",
			config: &LlmConfig{
				ID:               "model-707",
				Name:             "gpt-4",
				ModelType:        cdaenum.ModelTypeLlm,
				Temperature:      2.0,
				TopP:             1.0,
				TopK:             100,
				FrequencyPenalty: 2.0,
				PresencePenalty:  2.0,
				MaxTokens:        128000,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.config.ValObjCheck()
			if tt.wantErr {
				assert.Error(t, err, "expected error")
			} else {
				assert.NoError(t, err, "expected no error")

				if tt.config.MaxTokens == 0 {
					assert.Equal(t, 500, tt.config.MaxTokens, "MaxTokens should be set to default 500")
				}

				if tt.config.ModelType == "" {
					assert.Equal(t, cdaenum.ModelTypeLlm, tt.config.ModelType, "ModelType should be set to default ModelTypeLlm")
				}
			}
		})
	}
}

func TestLlmConfig_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		config  *LlmConfig
		wantErr bool
	}{
		{
			name: "有效配置",
			config: &LlmConfig{
				Name:      "gpt-4",
				MaxTokens: 2048,
			},
			wantErr: false,
		},
		{
			name: "Name缺失",
			config: &LlmConfig{
				MaxTokens: 2048,
			},
			wantErr: true,
		},
		{
			name: "MaxTokens缺失",
			config: &LlmConfig{
				Name: "gpt-4",
			},
			wantErr: true,
		},
		{
			name: "所有参数有效",
			config: &LlmConfig{
				Name:             "gpt-4",
				ModelType:        cdaenum.ModelTypeLlm,
				Temperature:      0.7,
				TopP:             0.9,
				TopK:             40,
				FrequencyPenalty: 0.0,
				PresencePenalty:  0.0,
				MaxTokens:        2048,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err, "expected error")
			} else {
				assert.NoError(t, err, "expected no error")
			}
		})
	}
}

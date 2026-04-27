package daconfvalobj

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRewrite_ValObjCheck(t *testing.T) {
	t.Parallel()

	validLlmConfig := &LlmConfig{
		Name:      "gpt-4",
		MaxTokens: 2048,
	}

	tests := []struct {
		name    string
		rewrite *Rewrite
		wantErr bool
	}{
		{
			name: "启用且配置完整",
			rewrite: &Rewrite{
				Enable:    func() *bool { b := true; return &b }(),
				LlmConfig: validLlmConfig,
			},
			wantErr: false,
		},
		{
			name: "禁用",
			rewrite: &Rewrite{
				Enable:    func() *bool { b := false; return &b }(),
				LlmConfig: validLlmConfig,
			},
			wantErr: false,
		},
		{
			name: "Enable为空",
			rewrite: &Rewrite{
				LlmConfig: validLlmConfig,
			},
			wantErr: true,
		},
		{
			name: "启用但LlmConfig为空",
			rewrite: &Rewrite{
				Enable:    func() *bool { b := true; return &b }(),
				LlmConfig: nil,
			},
			wantErr: true,
		},
		{
			name: "启用但LlmConfig无效",
			rewrite: &Rewrite{
				Enable: func() *bool { b := true; return &b }(),
				LlmConfig: &LlmConfig{
					// Missing required Name field
					MaxTokens: 2048,
				},
			},
			wantErr: true,
		},
		{
			name: "禁用且LlmConfig为空",
			rewrite: &Rewrite{
				Enable:    func() *bool { b := false; return &b }(),
				LlmConfig: nil,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.rewrite.ValObjCheck()
			if tt.wantErr {
				assert.Error(t, err, "expected error")
			} else {
				assert.NoError(t, err, "expected no error")
			}
		})
	}
}

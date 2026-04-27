package cdaenum

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToolInputParamValType_EnumCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		t       ToolInputParamValType
		wantErr bool
	}{
		{
			name:    "固定值",
			t:       ToolInputParamValTypeFixedValue,
			wantErr: false,
		},
		{
			name:    "引用变量",
			t:       ToolInputParamValTypeVar,
			wantErr: false,
		},
		{
			name:    "选择模型",
			t:       ToolInputParamValTypeModel,
			wantErr: false,
		},
		{
			name:    "自动生成",
			t:       ToolInputParamValTypeAuto,
			wantErr: false,
		},
		{
			name:    "无效类型",
			t:       ToolInputParamValType("invalid"),
			wantErr: true,
		},
		{
			name:    "空字符串",
			t:       ToolInputParamValType(""),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.t.EnumCheck()
			if tt.wantErr {
				assert.Error(t, err, "expected error")
			} else {
				assert.NoError(t, err, "expected no error")
			}
		})
	}
}

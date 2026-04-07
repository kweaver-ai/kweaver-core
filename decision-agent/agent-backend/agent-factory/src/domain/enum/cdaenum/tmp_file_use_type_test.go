package cdaenum

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTmpFileUseType_EnumCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		t       TmpFileUseType
		wantErr bool
	}{
		{
			name:    "直接上传",
			t:       TmpFileUseTypeUpload,
			wantErr: false,
		},
		{
			name:    "从临时区选择",
			t:       TmpFileUseTypeSelectFromTempZone,
			wantErr: false,
		},
		{
			name:    "无效类型",
			t:       TmpFileUseType("invalid"),
			wantErr: true,
		},
		{
			name:    "空字符串",
			t:       TmpFileUseType(""),
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

package cdaenum

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSupportDataType_EnumCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		c       SupportDataTypes
		wantErr bool
	}{
		{
			name:    "文件类型",
			c:       SupportDataTypes{"file"},
			wantErr: false,
		},
		{
			name:    "空列表",
			c:       SupportDataTypes{},
			wantErr: true,
		},
		{
			name:    "无效类型",
			c:       SupportDataTypes{"invalid"},
			wantErr: true,
		},
		{
			name:    "包含文件和无效类型",
			c:       SupportDataTypes{"file", "invalid"},
			wantErr: true,
		},
		{
			name:    "包含重复的有效类型",
			c:       SupportDataTypes{"file", "file"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.c.EnumCheck()
			if tt.wantErr {
				assert.Error(t, err, "expected error")
			} else {
				assert.NoError(t, err, "expected no error")
			}
		})
	}
}

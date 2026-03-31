package cdaenum

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAvatarType_EnumCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		t       AvatarType
		wantErr bool
	}{
		{
			name:    "内置头像",
			t:       AvatarTypeBuiltIn,
			wantErr: false,
		},
		{
			name:    "用户上传头像",
			t:       AvatarTypeUserUploaded,
			wantErr: false,
		},
		{
			name:    "AI生成头像",
			t:       AvatarTypeAIGenerated,
			wantErr: false,
		},
		{
			name:    "负数",
			t:       AvatarType(-1),
			wantErr: true,
		},
		{
			name:    "大于最大值",
			t:       AvatarType(4),
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

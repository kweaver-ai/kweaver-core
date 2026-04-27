package cdaenum

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBitUnit_EnumCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		b       BitUnit
		wantErr bool
	}{
		{
			name:    "KB",
			b:       KB,
			wantErr: false,
		},
		{
			name:    "MB",
			b:       MB,
			wantErr: false,
		},
		{
			name:    "GB",
			b:       GB,
			wantErr: false,
		},
		{
			name:    "无效单位",
			b:       BitUnit("TB"),
			wantErr: true,
		},
		{
			name:    "空字符串",
			b:       BitUnit(""),
			wantErr: true,
		},
		{
			name:    "其他单位",
			b:       BitUnit("byte"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.b.EnumCheck()
			if tt.wantErr {
				assert.Error(t, err, "expected error")
			} else {
				assert.NoError(t, err, "expected no error")
			}
		})
	}
}

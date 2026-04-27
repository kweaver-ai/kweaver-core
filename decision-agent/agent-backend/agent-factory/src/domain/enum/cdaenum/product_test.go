package cdaenum

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProduct_EnumCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		p       Product
		wantErr bool
	}{
		{
			name:    "DIP",
			p:       ProductDIP,
			wantErr: false,
		},
		{
			name:    "AnyShare",
			p:       ProductAnyShare,
			wantErr: false,
		},
		{
			name:    "ChatBI",
			p:       ProductChatBI,
			wantErr: false,
		},
		{
			name:    "无效产品",
			p:       Product("invalid"),
			wantErr: true,
		},
		{
			name:    "空字符串",
			p:       Product(""),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.p.EnumCheck()
			if tt.wantErr {
				assert.Error(t, err, "expected error")
			} else {
				assert.NoError(t, err, "expected no error")
			}
		})
	}
}

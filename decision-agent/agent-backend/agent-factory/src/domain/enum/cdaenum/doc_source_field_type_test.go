package cdaenum

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDocSourceFieldType_EnumCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		ft      DocSourceFieldType
		wantErr bool
	}{
		{
			name:    "valid folder type",
			ft:      DocSourceFieldTypeFolder,
			wantErr: false,
		},
		{
			name:    "valid file type",
			ft:      DocSourceFieldTypeFile,
			wantErr: false,
		},
		{
			name:    "invalid type",
			ft:      DocSourceFieldType("invalid"),
			wantErr: true,
		},
		{
			name:    "empty type",
			ft:      DocSourceFieldType(""),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.ft.EnumCheck()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

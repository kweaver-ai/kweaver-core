package cdaenum

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTempZoneStatus_Values(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status TempZoneStatus
		value  int
	}{
		{
			name:   "disabled value",
			status: TempZoneDisabled,
			value:  0,
		},
		{
			name:   "enabled value",
			status: TempZoneEnabled,
			value:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.value, int(tt.status))
		})
	}
}

func TestTempZoneStatus_IsEnabled(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status TempZoneStatus
		want   bool
	}{
		{
			name:   "disabled is not enabled",
			status: TempZoneDisabled,
			want:   false,
		},
		{
			name:   "enabled is enabled",
			status: TempZoneEnabled,
			want:   true,
		},
		{
			name:   "custom value 2 is not enabled",
			status: TempZoneStatus(2),
			want:   false,
		},
		{
			name:   "negative value is not enabled",
			status: TempZoneStatus(-1),
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.status == TempZoneEnabled
			assert.Equal(t, tt.want, got)
		})
	}
}

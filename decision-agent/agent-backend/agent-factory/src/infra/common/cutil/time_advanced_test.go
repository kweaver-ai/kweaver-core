package cutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseTime_Advanced(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		timeStr string
		wantErr bool
	}{
		{
			name:    "valid time - zero values",
			timeStr: "00:00:00",
			wantErr: false,
		},
		{
			name:    "valid time - max values",
			timeStr: "23:59:59",
			wantErr: false,
		},
		{
			name:    "valid time - single digits",
			timeStr: "1:2:3",
			wantErr: false,
		},
		{
			name:    "missing minutes",
			timeStr: "12:",
			wantErr: true,
		},
		{
			name:    "missing seconds",
			timeStr: "12:30",
			wantErr: true,
		},
		{
			name:    "only seconds",
			timeStr: ":30",
			wantErr: true,
		},
		{
			name:    "too many colons",
			timeStr: "12:30:45:00",
			wantErr: true,
		},
		{
			name:    "invalid format with letters",
			timeStr: "ab:cd:ef",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			hour, min, sec, err := ParseTime(tt.timeStr)

			if tt.wantErr {
				assert.Error(t, err, "ParseTime should return error")
			} else {
				assert.NoError(t, err, "ParseTime should not return error")
				assert.GreaterOrEqual(t, hour, 0, "hour should be >= 0")
				assert.GreaterOrEqual(t, min, 0, "min should be >= 0")
				assert.GreaterOrEqual(t, sec, 0, "sec should be >= 0")
				assert.LessOrEqual(t, hour, 23, "hour should be <= 23")
				assert.LessOrEqual(t, min, 59, "min should be <= 59")
				assert.LessOrEqual(t, sec, 59, "sec should be <= 59")
			}
		})
	}
}

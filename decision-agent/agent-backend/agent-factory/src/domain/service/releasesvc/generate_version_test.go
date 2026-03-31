package releasesvc

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateAgentVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		oldVersion  string
		newVersion  string
		wantErr     bool
		errContains string
	}{
		{
			name:       "increment version with v prefix",
			oldVersion: "v1",
			newVersion: "v2",
			wantErr:    false,
		},
		{
			name:       "increment version without v prefix",
			oldVersion: "10",
			newVersion: "v11",
			wantErr:    false,
		},
		{
			name:       "large version number",
			oldVersion: "v100",
			newVersion: "v101",
			wantErr:    false,
		},
		{
			name:        "invalid version format - not a number",
			oldVersion:  "v1.0",
			newVersion:  "",
			wantErr:     true,
			errContains: "invalid version format",
		},
		{
			name:        "invalid version format - letters",
			oldVersion:  "vabc",
			newVersion:  "",
			wantErr:     true,
			errContains: "invalid version format",
		},
		{
			name:        "empty version - error",
			oldVersion:  "",
			newVersion:  "",
			wantErr:     true,
			errContains: "invalid version format",
		},
		{
			name:       "zero version",
			oldVersion: "v0",
			newVersion: "v1",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := &releaseSvc{}
			result, err := svc.generateAgentVersion(tt.oldVersion)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Empty(t, result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.newVersion, result)
			}
		})
	}
}

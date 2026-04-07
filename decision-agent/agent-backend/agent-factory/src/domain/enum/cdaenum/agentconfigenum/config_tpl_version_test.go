package agentconfigenum

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigTplVersionT_EnumCheck_Valid(t *testing.T) {
	t.Parallel()

	err := ConfigTplVersionV1.EnumCheck()
	assert.NoError(t, err)
}

func TestConfigTplVersionT_EnumCheck_Invalid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		version ConfigTplVersionT
	}{
		{
			name:    "empty version",
			version: "",
		},
		{
			name:    "v2 version",
			version: "v2",
		},
		{
			name:    "invalid version",
			version: "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.version.EnumCheck()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid config version")
		})
	}
}

func TestConfigTplVersionT_String(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "v1", string(ConfigTplVersionV1))
}

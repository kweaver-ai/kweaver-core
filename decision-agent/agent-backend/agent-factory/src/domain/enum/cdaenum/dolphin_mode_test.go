package cdaenum

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDolphinMode_Constants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, DolphinMode(0), DolphinModeDisabled)
	assert.Equal(t, DolphinMode(1), DolphinModeEnabled)
}

func TestDolphinMode_EnumCheck_Valid(t *testing.T) {
	t.Parallel()

	validModes := []DolphinMode{
		DolphinModeDisabled,
		DolphinModeEnabled,
	}

	for _, mode := range validModes {
		t.Run("", func(t *testing.T) {
			t.Parallel()

			err := mode.EnumCheck()
			assert.NoError(t, err)
		})
	}
}

func TestDolphinMode_EnumCheck_Invalid(t *testing.T) {
	t.Parallel()

	invalidModes := []DolphinMode{
		-1,
		2,
		100,
	}

	for _, mode := range invalidModes {
		t.Run("", func(t *testing.T) {
			t.Parallel()

			err := mode.EnumCheck()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "dolphin模式不合法")
		})
	}
}

func TestDolphinMode_Bool(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		mode     DolphinMode
		expected bool
	}{
		{
			name:     "disabled mode",
			mode:     DolphinModeDisabled,
			expected: false,
		},
		{
			name:     "enabled mode",
			mode:     DolphinModeEnabled,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.mode.Bool()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDolphinMode_AllUnique(t *testing.T) {
	t.Parallel()

	modes := []DolphinMode{
		DolphinModeDisabled,
		DolphinModeEnabled,
	}

	uniqueModes := make(map[DolphinMode]bool)
	for _, mode := range modes {
		assert.False(t, uniqueModes[mode], "Duplicate mode found: %d", mode)
		uniqueModes[mode] = true
	}
}

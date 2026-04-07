package cconstant

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPmsAllFlag(t *testing.T) {
	t.Parallel()

	t.Run("PmsAllFlag is an asterisk", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "*", PmsAllFlag)
	})
}

func TestIsContainsPmsAllFlag(t *testing.T) {
	t.Parallel()

	t.Run("returns true when slice contains PmsAllFlag", func(t *testing.T) {
		t.Parallel()

		slice := []string{"read", "write", "*"}
		result := IsContainsPmsAllFlag(slice)
		assert.True(t, result)
	})

	t.Run("returns false when slice does not contain PmsAllFlag", func(t *testing.T) {
		t.Parallel()

		slice := []string{"read", "write"}
		result := IsContainsPmsAllFlag(slice)
		assert.False(t, result)
	})

	t.Run("returns false for empty slice", func(t *testing.T) {
		t.Parallel()

		slice := []string{}
		result := IsContainsPmsAllFlag(slice)
		assert.False(t, result)
	})
}

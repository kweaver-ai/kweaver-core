package cconstant

import (
	"testing"
)

func TestRequestConstants(t *testing.T) {
	t.Parallel()

	t.Run("NameMaxLength constant", func(t *testing.T) {
		t.Parallel()

		if NameMaxLength != 50 {
			t.Errorf("Expected NameMaxLength to be 50, got %d", NameMaxLength)
		}
	})

	t.Run("ProfileMaxLength constant", func(t *testing.T) {
		t.Parallel()

		if ProfileMaxLength != 100 {
			t.Errorf("Expected ProfileMaxLength to be 100, got %d", ProfileMaxLength)
		}
	})
}

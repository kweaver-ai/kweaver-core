package cconstant

import (
	"testing"
)

func TestUniqueIDFlag(t *testing.T) {
	t.Parallel()

	t.Run("UniqueIDFlag type exists", func(t *testing.T) {
		t.Parallel()
		// Test that we can create variables of UniqueIDFlag type
		var flag UniqueIDFlag
		_ = flag
	})

	t.Run("UniqueIDFlagDB constant", func(t *testing.T) {
		t.Parallel()

		expected := UniqueIDFlag(1)
		if UniqueIDFlagDB != expected {
			t.Errorf("Expected UniqueIDFlagDB to be %d, got %d", expected, UniqueIDFlagDB)
		}
	})

	t.Run("UniqueIDFlagRedisDlm constant", func(t *testing.T) {
		t.Parallel()

		expected := UniqueIDFlag(2)
		if UniqueIDFlagRedisDlm != expected {
			t.Errorf("Expected UniqueIDFlagRedisDlm to be %d, got %d", expected, UniqueIDFlagRedisDlm)
		}
	})

	t.Run("flags are unique", func(t *testing.T) {
		t.Parallel()

		if UniqueIDFlagDB == UniqueIDFlagRedisDlm {
			t.Error("UniqueIDFlagDB and UniqueIDFlagRedisDlm should be different")
		}
	})

	t.Run("assign flag to variable", func(t *testing.T) {
		t.Parallel()

		var flag UniqueIDFlag = UniqueIDFlagDB
		if flag != UniqueIDFlagDB {
			t.Errorf("Expected flag to be UniqueIDFlagDB, got %d", flag)
		}
	})
}

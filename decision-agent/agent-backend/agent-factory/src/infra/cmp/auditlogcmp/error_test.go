package auditlogcmp

import (
	"errors"
	"testing"
)

func TestErrAuditLogIsNil(t *testing.T) {
	t.Parallel()

	t.Run("error value", func(t *testing.T) {
		t.Parallel()

		if ErrAuditLogIsNil == nil {
			t.Error("Expected ErrAuditLogIsNil to be defined")
		}

		expectedMsg := "audit log is nil"
		if ErrAuditLogIsNil.Error() != expectedMsg {
			t.Errorf("Expected error message '%s', got '%s'", expectedMsg, ErrAuditLogIsNil.Error())
		}
	})

	t.Run("error is errors.New type", func(t *testing.T) {
		t.Parallel()
		// Verify it's a proper error
		var err error = ErrAuditLogIsNil
		if err == nil {
			t.Error("Expected ErrAuditLogIsNil to be a non-nil error")
		}
	})

	t.Run("can be compared with errors.Is", func(t *testing.T) {
		t.Parallel()

		customErr := errors.New("audit log is nil")
		if !errors.Is(customErr, ErrAuditLogIsNil) {
			// This should actually be false since they're different instances
			// but let's verify the error works correctly
			t.Log("Custom error is not the same as ErrAuditLogIsNil (expected)")
		}

		// Now test with the actual error
		testErr := ErrAuditLogIsNil
		if !errors.Is(testErr, ErrAuditLogIsNil) {
			t.Error("Expected errors.Is to return true for same error")
		}
	})
}

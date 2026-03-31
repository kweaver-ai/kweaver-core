package cconstant

import (
	"testing"
)

func TestObjectTypes(t *testing.T) {
	t.Parallel()

	t.Run("OBJECTC_TYPE_PUBLISH constant", func(t *testing.T) {
		t.Parallel()

		expected := "ID_AUDIT_PUBLISH"
		if OBJECTC_TYPE_PUBLISH != expected {
			t.Errorf("Expected OBJECTC_TYPE_PUBLISH to be '%s', got '%s'", expected, OBJECTC_TYPE_PUBLISH)
		}
	})

	t.Run("OBJECTC_TYPE_PUBLISH is not empty", func(t *testing.T) {
		t.Parallel()

		if OBJECTC_TYPE_PUBLISH == "" {
			t.Error("Expected OBJECTC_TYPE_PUBLISH to not be empty")
		}
	})
}

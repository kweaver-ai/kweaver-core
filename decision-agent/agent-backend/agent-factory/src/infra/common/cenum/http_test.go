package cenum

import (
	"testing"
)

func TestHTTPConstants(t *testing.T) {
	t.Parallel()

	t.Run("HTTPHct constant", func(t *testing.T) {
		t.Parallel()

		if HTTPHct != "Content-Type" {
			t.Errorf("Expected HTTPHct to be 'Content-Type', got '%s'", HTTPHct)
		}
	})

	t.Run("HTTPHctJSON constant", func(t *testing.T) {
		t.Parallel()

		if HTTPHctJSON != "application/json" {
			t.Errorf("Expected HTTPHctJSON to be 'application/json', got '%s'", HTTPHctJSON)
		}
	})

	t.Run("HTTPHctXML constant", func(t *testing.T) {
		t.Parallel()

		if HTTPHctXML != "application/xml" {
			t.Errorf("Expected HTTPHctXML to be 'application/xml', got '%s'", HTTPHctXML)
		}
	})

	t.Run("HTTPHctForm constant", func(t *testing.T) {
		t.Parallel()

		if HTTPHctForm != "application/x-www-form-urlencoded" {
			t.Errorf("Expected HTTPHctForm to be 'application/x-www-form-urlencoded', got '%s'", HTTPHctForm)
		}
	})
}

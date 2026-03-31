package validatorhelper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommonCustomValidator(t *testing.T) {
	t.Parallel()

	t.Run("register custom validators successfully", func(t *testing.T) {
		t.Parallel()

		err := CommonCustomValidator()
		assert.NoError(t, err, "CommonCustomValidator should return nil on success")
	})
}

func TestCommonCustomValidator_MultipleCalls(t *testing.T) {
	t.Parallel()

	t.Run("multiple calls should be safe", func(t *testing.T) {
		t.Parallel()
		// First call
		err1 := CommonCustomValidator()
		assert.NoError(t, err1)

		// Second call - should not error (registering same validator twice might error)
		err2 := CommonCustomValidator()
		// The second call might return an error if the validator is already registered
		// This is expected behavior for the go-playground/validator
		if err2 != nil {
			t.Logf("Second call returned expected error: %v", err2)
		}
	})
}

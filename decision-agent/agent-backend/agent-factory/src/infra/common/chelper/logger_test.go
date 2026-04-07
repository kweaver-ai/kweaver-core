package chelper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetStdoutLogger(t *testing.T) {
	t.Parallel()

	t.Run("returns non-nil logger", func(t *testing.T) {
		t.Parallel()

		logger := GetStdoutLogger()
		assert.NotNil(t, logger)
	})

	t.Run("returns same logger on multiple calls", func(t *testing.T) {
		t.Parallel()

		logger1 := GetStdoutLogger()
		logger2 := GetStdoutLogger()

		assert.Same(t, logger1, logger2)
	})
}

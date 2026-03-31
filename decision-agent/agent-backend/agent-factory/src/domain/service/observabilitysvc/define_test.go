package observabilitysvc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewObservabilitySvc(t *testing.T) {
	t.Parallel()

	t.Run("creates service with all dependencies", func(t *testing.T) {
		t.Parallel()

		dto := &NewObservabilitySvcDto{
			Logger:    nil,
			Uniquery:  nil,
			SquareSvc: nil,
		}

		svc := NewObservabilitySvc(dto)

		assert.NotNil(t, svc)
		assert.IsType(t, &observabilitySvc{}, svc)
	})

	t.Run("creates service with minimal dependencies", func(t *testing.T) {
		t.Parallel()

		dto := &NewObservabilitySvcDto{
			Logger:    nil,
			Uniquery:  nil,
			SquareSvc: nil,
		}

		svc := NewObservabilitySvc(dto)

		assert.NotNil(t, svc)
	})
}

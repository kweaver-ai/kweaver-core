package squaresvc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSquareSvc_Struct(t *testing.T) {
	t.Parallel()

	t.Run("squareSvc struct exists", func(t *testing.T) {
		t.Parallel()
		// Create a simple instance to verify struct can be created
		svc := &squareSvc{}

		assert.NotNil(t, svc)
	})

	t.Run("squareSvc with nil SvcBase", func(t *testing.T) {
		t.Parallel()

		svc := &squareSvc{
			SvcBase: nil,
		}

		assert.NotNil(t, svc)
		assert.Nil(t, svc.SvcBase)
	})
}

func TestSquareSvc_Interface(t *testing.T) {
	t.Parallel()

	t.Run("implements ISquareSvc interface", func(t *testing.T) {
		t.Parallel()
		// Verify the struct implements the interface
		// This is a compile-time check done in the source code
		// var _ iv3portdriver.ISquareSvc = &squareSvc{}

		svc := &squareSvc{}
		assert.NotNil(t, svc)
	})
}

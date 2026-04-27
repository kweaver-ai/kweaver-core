package crest

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetRestHttpErr_NilError(t *testing.T) {
	t.Parallel()

	httpError, ok := GetRestHttpErr(nil)

	assert.Nil(t, httpError)
	assert.False(t, ok)
}

func TestGetRestHttpErr_GenericError(t *testing.T) {
	t.Parallel()

	genericErr := errors.New("generic error")

	httpError, ok := GetRestHttpErr(genericErr)

	assert.Nil(t, httpError)
	assert.False(t, ok)
}

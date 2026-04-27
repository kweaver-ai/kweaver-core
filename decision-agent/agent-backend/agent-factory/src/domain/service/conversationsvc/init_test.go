package conversationsvc

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/stretchr/testify/assert"
)

func TestIsSessionNotFoundError(t *testing.T) {
	t.Parallel()

	svc := &conversationSvc{}
	ctx := context.Background()

	t.Run("returns true for 404 HTTPError", func(t *testing.T) {
		t.Parallel()

		err := rest.NewHTTPError(ctx, http.StatusNotFound, rest.PublicError_NotFound)
		result := svc.isSessionNotFoundError(err)
		assert.True(t, result)
	})

	t.Run("returns false for 500 HTTPError", func(t *testing.T) {
		t.Parallel()

		err := rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError)
		result := svc.isSessionNotFoundError(err)
		assert.False(t, result)
	})

	t.Run("returns false for non-HTTPError", func(t *testing.T) {
		t.Parallel()

		err := errors.New("some error")
		result := svc.isSessionNotFoundError(err)
		assert.False(t, result)
	})

	t.Run("returns false for nil error", func(t *testing.T) {
		t.Parallel()

		result := svc.isSessionNotFoundError(nil)
		assert.False(t, result)
	})
}

func TestIsSessionAlreadyExistsError(t *testing.T) {
	t.Parallel()

	svc := &conversationSvc{}
	ctx := context.Background()

	t.Run("returns true for 409 HTTPError", func(t *testing.T) {
		t.Parallel()

		err := rest.NewHTTPError(ctx, http.StatusConflict, rest.PublicError_Conflict)
		result := svc.isSessionAlreadyExistsError(err)
		assert.True(t, result)
	})

	t.Run("returns true for error with 'already exists' message", func(t *testing.T) {
		t.Parallel()

		err := errors.New("session already exists")
		result := svc.isSessionAlreadyExistsError(err)
		assert.True(t, result)
	})

	t.Run("returns false for other HTTPError", func(t *testing.T) {
		t.Parallel()

		err := rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest)
		result := svc.isSessionAlreadyExistsError(err)
		assert.False(t, result)
	})

	t.Run("returns false for non-HTTPError without 'already exists'", func(t *testing.T) {
		t.Parallel()

		err := errors.New("some other error")
		result := svc.isSessionAlreadyExistsError(err)
		assert.False(t, result)
	})

	t.Run("case sensitive check for 'already exists'", func(t *testing.T) {
		t.Parallel()

		err := errors.New("Session ALREADY EXISTS error")
		result := svc.isSessionAlreadyExistsError(err)
		assert.False(t, result) // Implementation is case-sensitive, only checks lowercase "already exists"
	})
}

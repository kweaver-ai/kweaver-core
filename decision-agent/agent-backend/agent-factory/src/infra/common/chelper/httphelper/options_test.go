package httphelper

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHTTPClient_ReturnsNonNil(t *testing.T) {
	t.Parallel()

	c := NewHTTPClient()
	require.NotNil(t, c)
}

func TestWithToken(t *testing.T) {
	t.Parallel()

	t.Run("empty token is ignored", func(t *testing.T) {
		t.Parallel()

		c := newTestHTTPClient()
		WithToken("")(c)
		assert.Equal(t, "", c.token)
	})

	t.Run("sets token without Bearer prefix", func(t *testing.T) {
		t.Parallel()

		c := newTestHTTPClient()
		WithToken("mytoken")(c)
		assert.Equal(t, "mytoken", c.token)
	})

	t.Run("strips Bearer prefix", func(t *testing.T) {
		t.Parallel()

		c := newTestHTTPClient()
		WithToken("Bearer mytoken")(c)
		assert.Equal(t, "mytoken", c.token)
	})
}

func TestWithHeader(t *testing.T) {
	t.Parallel()

	c := newTestHTTPClient()
	require.NotNil(t, c.client)
	WithHeader("X-Custom", "value")(c)
	assert.NotNil(t, c.client)
}

func TestWithHeaders(t *testing.T) {
	t.Parallel()

	c := newTestHTTPClient()
	require.NotNil(t, c.client)
	WithHeaders(map[string]string{
		"X-A": "a",
		"X-B": "b",
	})(c)
	assert.NotNil(t, c.client)
}

func TestGetClientFromOptions(t *testing.T) {
	t.Parallel()

	c := newTestHTTPClient()
	gc := c.GetClient()
	assert.NotNil(t, gc)
}

func newTestHTTPClient() *httpClient {
	return NewHTTPClient().(*httpClient)
}

package httphelper

import (
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHTTPClient(t *testing.T) {
	t.Parallel()

	client := NewHTTPClient()

	assert.NotNil(t, client, "NewHTTPClient should return a non-nil client")
	assert.Implements(t, (*icmp.IHttpClient)(nil), client, "Should implement IHttpClient interface")
}

func TestNewHTTPClient_WithNoOptions(t *testing.T) {
	t.Parallel()

	c := NewHTTPClient()
	httpClient, ok := c.(*httpClient)
	require.True(t, ok, "Should be able to cast to *httpClient")

	assert.NotNil(t, httpClient.client, "Internal gclient should be initialized")
	assert.Empty(t, httpClient.token, "Token should be empty by default")
}

func TestNewHTTPClient_WithToken(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		token string
	}{
		{
			name:  "with plain token",
			token: "test-token-123",
		},
		{
			name:  "with Bearer prefix",
			token: "Bearer test-token-123",
		},
		{
			name:  "empty token",
			token: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := NewHTTPClient(WithToken(tt.token))
			httpClient, ok := c.(*httpClient)
			require.True(t, ok)

			// Token should be stripped of "Bearer " prefix
			expectedToken := tt.token
			if len(expectedToken) > 7 && expectedToken[:7] == "Bearer " {
				expectedToken = expectedToken[7:]
			}

			if expectedToken == "" {
				expectedToken = ""
			}

			assert.Equal(t, expectedToken, httpClient.token)
		})
	}
}

func TestWithToken_BearerPrefixRemoval(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       string
		expectedTok string
	}{
		{
			name:        "token without Bearer prefix",
			input:       "my-token",
			expectedTok: "my-token",
		},
		{
			name:        "token with Bearer prefix",
			input:       "Bearer my-token",
			expectedTok: "my-token",
		},
		{
			name:        "empty string",
			input:       "",
			expectedTok: "",
		},
		{
			name:        "multiple Bearer prefixes",
			input:       "Bearer Bearer my-token",
			expectedTok: "Bearer my-token",
		},
		{
			name:        "Bearer with space after",
			input:       "Bearer  my-token",
			expectedTok: " my-token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := NewHTTPClient(WithToken(tt.input))
			httpClient, ok := c.(*httpClient)
			require.True(t, ok)

			assert.Equal(t, tt.expectedTok, httpClient.token)
		})
	}
}

func TestNewHTTPClient_WithHeader(t *testing.T) {
	t.Parallel()

	c := NewHTTPClient(WithHeader("X-Custom-Header", "custom-value"))
	httpClient, ok := c.(*httpClient)
	require.True(t, ok)

	assert.NotNil(t, httpClient.client)
	// Verify the client is properly configured
	assert.NotNil(t, httpClient.client.Client)
}

func TestNewHTTPClient_WithHeaders(t *testing.T) {
	t.Parallel()

	headers := map[string]string{
		"X-Header-1": "value1",
		"X-Header-2": "value2",
		"X-Header-3": "value3",
	}

	c := NewHTTPClient(WithHeaders(headers))
	httpClient, ok := c.(*httpClient)
	require.True(t, ok)

	assert.NotNil(t, httpClient.client)
	assert.NotNil(t, httpClient.client.Client)
}

func TestNewHTTPClient_WithEmptyHeaders(t *testing.T) {
	t.Parallel()

	c := NewHTTPClient(WithHeaders(map[string]string{}))
	httpClient, ok := c.(*httpClient)
	require.True(t, ok)

	assert.NotNil(t, httpClient.client)
}

func TestNewHTTPClient_MultipleOptions(t *testing.T) {
	t.Parallel()

	c := NewHTTPClient(
		WithToken("test-token"),
		WithHeader("X-Custom-1", "value1"),
		WithHeader("X-Custom-2", "value2"),
		WithHeaders(map[string]string{
			"X-Grouped-1": "group1",
			"X-Grouped-2": "group2",
		}),
	)
	httpClient, ok := c.(*httpClient)
	require.True(t, ok)

	assert.NotNil(t, httpClient.client)
	assert.Equal(t, "test-token", httpClient.token)
}

func TestNewHTTPClient_WithHeaderOverrides(t *testing.T) {
	t.Parallel()

	c := NewHTTPClient(
		WithHeader("X-Test", "value1"),
		WithHeader("X-Test", "value2"),
	)
	httpClient, ok := c.(*httpClient)
	require.True(t, ok)

	// Headers should be set (last value wins)
	assert.NotNil(t, httpClient.client)
}

func TestHttpClient_GetClient(t *testing.T) {
	t.Parallel()

	c := NewHTTPClient()
	httpClient, ok := c.(*httpClient)
	require.True(t, ok)

	gclient := httpClient.GetClient()
	assert.NotNil(t, gclient)
	assert.Same(t, httpClient.client, gclient, "GetClient should return the same gclient instance")
}

func TestHttpClient_ImplementsIHttpClient(t *testing.T) {
	t.Parallel()

	var _ icmp.IHttpClient = NewHTTPClient()

	assert.NotNil(t, NewHTTPClient())
}

func TestHttpClient_SetContentType(t *testing.T) {
	t.Parallel()

	c := NewHTTPClient()
	httpClient, ok := c.(*httpClient)
	require.True(t, ok)

	// setContentType is a private method, but we can test its effect
	// through the public API
	httpClient.setContentType("application/json")

	// Verify the client is still properly configured
	assert.NotNil(t, httpClient.client)
}

func TestOptionFunction(t *testing.T) {
	t.Parallel()

	// Test that Option function type works correctly
	var opt Option = func(c *httpClient) {
		c.token = "test-token-from-func"
	}

	c := &httpClient{}
	opt(c)

	assert.Equal(t, "test-token-from-func", c.token)
}

func TestNewHTTPClient_ChainedOptions(t *testing.T) {
	t.Parallel()

	c := NewHTTPClient(
		WithToken("token1"),
		WithToken("token2"), // This will override
	)
	httpClient, ok := c.(*httpClient)
	require.True(t, ok)

	// Last WithToken should win, but note that WithToken with empty string is a no-op
	// So token2 should be set
	assert.Equal(t, "token2", httpClient.token)
}

func TestWithToken_EmptyString(t *testing.T) {
	t.Parallel()

	c := NewHTTPClient(WithToken(""))
	httpClient, ok := c.(*httpClient)
	require.True(t, ok)

	assert.Empty(t, httpClient.token, "Empty token should remain empty")
}

func TestWithToken_WhitespaceOnly(t *testing.T) {
	t.Parallel()

	c := NewHTTPClient(WithToken("   "))
	httpClient, ok := c.(*httpClient)
	require.True(t, ok)

	assert.Equal(t, "   ", httpClient.token, "Whitespace token should be preserved")
}

func TestDetailMap(t *testing.T) {
	t.Parallel()

	// Test that DetailMap type alias works correctly
	detail := DetailMap{
		"field1": "value1",
		"field2": 123,
		"field3": true,
	}

	assert.Equal(t, "value1", detail["field1"])
	assert.Equal(t, 123, detail["field2"])
	assert.Equal(t, true, detail["field3"])
}

func TestCommonResp_Structure(t *testing.T) {
	t.Parallel()

	resp := CommonResp{
		Code:        400,
		Cause:       "test cause",
		Message:     "test message",
		Description: "test description",
		Solution:    "test solution",
		Detail:      DetailMap{"key": "value"},
		Debug:       "debug info",
	}

	assert.Equal(t, 400, resp.Code)
	assert.Equal(t, "test cause", resp.Cause)
	assert.Equal(t, "test message", resp.Message)
	assert.Equal(t, "test description", resp.Description)
	assert.Equal(t, "test solution", resp.Solution)
	assert.Equal(t, "value", resp.Detail["key"])
	assert.Equal(t, "debug info", resp.Debug)
}

func TestCommonRespError_TypeAlias(t *testing.T) {
	t.Parallel()

	resp := CommonResp{
		Code:    500,
		Message: "error",
	}

	// CommonRespError is an alias for CommonResp
	var errResp *CommonRespError = (*CommonRespError)(&resp)

	assert.Equal(t, 500, errResp.Code)
	assert.Equal(t, "error", errResp.Message)
}

func TestNewHTTPClient_PreservesExistingTransport(t *testing.T) {
	t.Parallel()

	// Test that the HTTP client is properly initialized with OpenTelemetry transport
	c := NewHTTPClient()
	httpClient, ok := c.(*httpClient)
	require.True(t, ok)

	assert.NotNil(t, httpClient.client)
	// The client should have a transport set (with otelhttp wrapping)
	assert.NotNil(t, httpClient.client.Client.Transport)
}

func TestHttpClient_TransportConfiguration(t *testing.T) {
	t.Parallel()

	c := NewHTTPClient()
	httpClient, ok := c.(*httpClient)
	require.True(t, ok)

	// Verify the underlying http.Client has proper configuration
	underlyingClient := httpClient.client.Client
	assert.NotNil(t, underlyingClient)
	assert.NotNil(t, underlyingClient.Transport)
}

func TestWithHeaders_NilMap(t *testing.T) {
	t.Parallel()

	// This test verifies that WithHeaders handles nil gracefully
	c := NewHTTPClient(WithHeaders(nil))
	assert.NotNil(t, c)
}

func TestWithToken_VariousFormats(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		token    string
		expected string
	}{
		{
			name:     "simple token",
			token:    "abc123",
			expected: "abc123",
		},
		{
			name:     "token with Bearer prefix",
			token:    "Bearer abc123",
			expected: "abc123",
		},
		{
			name:     "token with multiple Bearer",
			token:    "Bearer Bearer abc123",
			expected: "Bearer abc123",
		},
		{
			name:     "Bearer only",
			token:    "Bearer ",
			expected: "",
		},
		{
			name:     "Bearer with spaces",
			token:    "Bearer    abc123",
			expected: "   abc123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := NewHTTPClient(WithToken(tt.token))
			httpClient, ok := c.(*httpClient)
			require.True(t, ok)
			assert.Equal(t, tt.expected, httpClient.token)
		})
	}
}

func TestNewHTTPClient_UnderlyingClientNotNil(t *testing.T) {
	t.Parallel()

	c := NewHTTPClient()
	httpClient, ok := c.(*httpClient)
	require.True(t, ok)

	// The underlying gclient should not be nil
	assert.NotNil(t, httpClient.client)
	// The underlying http.Client should not be nil
	assert.NotNil(t, httpClient.client.Client)
}

func TestHttpClient_GetClient_ReturnsSameInstance(t *testing.T) {
	t.Parallel()

	c := NewHTTPClient()
	httpClient, ok := c.(*httpClient)
	require.True(t, ok)

	gclient1 := httpClient.GetClient()
	gclient2 := httpClient.GetClient()

	assert.Same(t, gclient1, gclient2)
}

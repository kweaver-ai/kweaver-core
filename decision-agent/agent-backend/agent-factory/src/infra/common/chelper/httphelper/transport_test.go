package httphelper

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetDefaultTp(t *testing.T) {
	t.Parallel()

	tp := GetDefaultTp()

	assert.NotNil(t, tp, "GetDefaultTp should return a non-nil Transport")
	assert.NotNil(t, tp.DialContext, "DialContext should be set")
	assert.True(t, tp.ForceAttemptHTTP2, "ForceAttemptHTTP2 should be true")
	assert.Equal(t, 512, tp.MaxIdleConns, "MaxIdleConns should be 512")
	assert.Equal(t, 100, tp.MaxIdleConnsPerHost, "MaxIdleConnsPerHost should be 100")
	assert.Equal(t, 90*time.Second, tp.IdleConnTimeout, "IdleConnTimeout should be 90 seconds")
	assert.Equal(t, 10*time.Second, tp.TLSHandshakeTimeout, "TLSHandshakeTimeout should be 10 seconds")
	assert.Equal(t, 1*time.Second, tp.ExpectContinueTimeout, "ExpectContinueTimeout should be 1 second")
}

func TestGetDefaultTp_DialContextTimeouts(t *testing.T) {
	t.Parallel()

	tp := GetDefaultTp()

	// Note: We can't directly access the Dialer's timeouts without reflection
	// but we can verify the Transport is properly configured
	assert.NotNil(t, tp, "Transport should be initialized")
}

func TestGetClient(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		timeout time.Duration
	}{
		{
			name:    "zero timeout",
			timeout: 0,
		},
		{
			name:    "5 second timeout",
			timeout: 5 * time.Second,
		},
		{
			name:    "default timeout",
			timeout: DefaultTimeout,
		},
		{
			name:    "custom timeout",
			timeout: 30 * time.Second,
		},
		{
			name:    "millisecond timeout",
			timeout: 500 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client := GetClient(tt.timeout)

			assert.NotNil(t, client, "GetClient should return a non-nil Client")
			assert.NotNil(t, client.Transport, "Client Transport should be set")
			assert.Equal(t, tt.timeout, client.Timeout, "Client Timeout should match the provided timeout")
		})
	}
}

func TestGetClient_TransportType(t *testing.T) {
	t.Parallel()

	client := GetClient(5 * time.Second)

	assert.NotNil(t, client.Transport)
	// The transport should be an *http.Transport
	_, ok := client.Transport.(*http.Transport)
	assert.True(t, ok, "Transport should be *http.Transport")
}

func TestGetDefaultStdClient(t *testing.T) {
	t.Parallel()

	client := GetDefaultStdClient()

	assert.NotNil(t, client, "GetDefaultStdClient should return a non-nil Client")
	assert.NotNil(t, client.Transport, "Default client Transport should be set")
}

func TestGetDefaultStdClient_SameInstance(t *testing.T) {
	t.Parallel()

	client1 := GetDefaultStdClient()
	client2 := GetDefaultStdClient()

	// Should return the same instance (initialized in init())
	assert.Same(t, client1, client2, "GetDefaultStdClient should return the same instance")
}

func TestGetDefaultClient(t *testing.T) {
	t.Parallel()

	client := GetDefaultClient()

	assert.NotNil(t, client, "GetDefaultClient should return a non-nil Client")
	assert.NotNil(t, client.Transport, "Default client Transport should be set")
}

func TestGetDefaultClient_SameAsStdClient(t *testing.T) {
	t.Parallel()

	stdClient := GetDefaultStdClient()
	defaultClient := GetDefaultClient()

	assert.Same(t, stdClient, defaultClient, "GetDefaultClient should return the same instance as GetDefaultStdClient")
}

func TestGetClient_MultipleCalls(t *testing.T) {
	t.Parallel()

	client1 := GetClient(5 * time.Second)
	client2 := GetClient(10 * time.Second)

	assert.NotNil(t, client1)
	assert.NotNil(t, client2)

	// Different calls should return different instances
	assert.NotSame(t, client1, client2, "GetClient should return different instances for different calls")

	// Each should have its own timeout
	assert.Equal(t, 5*time.Second, client1.Timeout)
	assert.Equal(t, 10*time.Second, client2.Timeout)
}

func TestGetClient_TransportConfiguration(t *testing.T) {
	t.Parallel()

	timeout := 15 * time.Second
	client := GetClient(timeout)

	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatal("Transport should be *http.Transport")
	}

	assert.Equal(t, 512, transport.MaxIdleConns)
	assert.Equal(t, 100, transport.MaxIdleConnsPerHost)
	assert.Equal(t, 90*time.Second, transport.IdleConnTimeout)
}

func TestTransport_Constants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, 5, RetryInterval, "RetryInterval should be 5")
}

func TestGetDefaultTp_HasForceAttemptHTTP2(t *testing.T) {
	t.Parallel()

	tp := GetDefaultTp()

	// Verify HTTP/2 is enabled
	assert.True(t, tp.ForceAttemptHTTP2, "ForceAttemptHTTP2 should be enabled")
}

func TestGetClient_ZeroTimeoutBehavior(t *testing.T) {
	t.Parallel()

	// A zero timeout means no timeout (infinite wait)
	client := GetClient(0)

	assert.NotNil(t, client)
	assert.Equal(t, time.Duration(0), client.Timeout)
}

func TestGetClient_NegativeTimeout(t *testing.T) {
	t.Parallel()

	// Test with negative timeout (should still work, though not recommended)
	client := GetClient(-1 * time.Second)

	assert.NotNil(t, client)
	// Negative timeout is effectively 0 in Go's http.Client
	assert.Equal(t, time.Duration(-1)*time.Second, client.Timeout)
}

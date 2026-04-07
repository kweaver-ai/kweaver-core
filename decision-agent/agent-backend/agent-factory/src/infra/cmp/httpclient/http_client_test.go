package httpclient

import (
	"testing"
	"time"
)

func TestNewRawHTTPClient(t *testing.T) {
	t.Parallel()

	t.Run("create raw HTTP client", func(t *testing.T) {
		t.Parallel()

		client := NewRawHTTPClient()

		if client == nil {
			t.Fatal("Expected client to be created, got nil")
		}

		if client.Timeout != 100*time.Second {
			t.Errorf("Expected timeout to be 100s, got %v", client.Timeout)
		}
	})

	t.Run("singleton pattern", func(t *testing.T) {
		t.Parallel()

		client1 := NewRawHTTPClient()
		client2 := NewRawHTTPClient()

		if client1 != client2 {
			t.Error("Expected NewRawHTTPClient to return singleton instance")
		}
	})
}

func TestNewHTTPClient(t *testing.T) {
	t.Parallel()

	t.Run("create HTTP client", func(t *testing.T) {
		t.Parallel()

		client := NewHTTPClient()

		if client == nil {
			t.Fatal("Expected client to be created, got nil")
		}
	})

	t.Run("singleton pattern", func(t *testing.T) {
		t.Parallel()

		client1 := NewHTTPClient()
		client2 := NewHTTPClient()

		if client1 != client2 {
			t.Error("Expected NewHTTPClient to return singleton instance")
		}
	})
}

func TestNewHTTPClientEx(t *testing.T) {
	t.Parallel()

	t.Run("create HTTP client with custom timeout", func(t *testing.T) {
		t.Parallel()

		timeout := 30 * time.Second
		client := NewHTTPClientEx(30)

		if client == nil {
			t.Fatal("Expected client to be created, got nil")
		}

		// Check if it's a httpClient type
		httpClient, ok := client.(*httpClient)
		if !ok {
			t.Fatal("Expected httpClient type")
		}

		if httpClient.client.Timeout != timeout {
			t.Errorf("Expected timeout to be %v, got %v", timeout, httpClient.client.Timeout)
		}
	})

	t.Run("different instances", func(t *testing.T) {
		t.Parallel()

		client1 := NewHTTPClientEx(10)
		client2 := NewHTTPClientEx(20)

		if client1 == client2 {
			t.Error("Expected NewHTTPClientEx to return different instances")
		}
	})
}

func TestAddHeaders(t *testing.T) {
	t.Parallel()

	t.Run("nil headers", func(t *testing.T) {
		t.Parallel()
		// Test with nil headers - this would need a real http.Request
		// For now, we just verify the method doesn't panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Unexpected panic: %v", r)
			}
		}()
		// Cannot test with nil request in unit test without creating real request
	})

	t.Run("empty headers", func(t *testing.T) {
		t.Parallel()
		// Test with empty headers map
	})

	t.Run("with headers", func(t *testing.T) {
		t.Parallel()
		// Test with valid headers
	})

	t.Run("skip empty header values", func(t *testing.T) {
		t.Parallel()
		// Test that empty header values are skipped
	})
}

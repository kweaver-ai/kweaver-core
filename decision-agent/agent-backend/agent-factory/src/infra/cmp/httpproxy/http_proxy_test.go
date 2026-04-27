package httpproxy

import (
	"net/http"
	"testing"
	"time"
)

// Test types for generic proxy
type TestRequest struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

type TestResponse struct {
	Result  string `json:"result"`
	Status  int    `json:"status"`
	Message string `json:"message"`
}

func TestNewJSONPostProxy(t *testing.T) {
	t.Parallel()

	t.Run("create proxy", func(t *testing.T) {
		t.Parallel()

		targetURL := "http://example.com/api"
		proxy := NewJSONPostProxy[TestRequest, TestResponse](targetURL)

		if proxy == nil {
			t.Fatal("Expected proxy to be created, got nil")
		}

		if proxy.TargetURL != targetURL {
			t.Errorf("Expected TargetURL to be '%s', got '%s'", targetURL, proxy.TargetURL)
		}

		if proxy.Client == nil {
			t.Error("Expected Client to be initialized")
		}

		if proxy.Client.Timeout != 10*time.Second {
			t.Errorf("Expected timeout to be 10s, got %v", proxy.Client.Timeout)
		}

		if proxy.Token != "" {
			t.Error("Expected Token to be empty")
		}
	})

	t.Run("empty target URL", func(t *testing.T) {
		t.Parallel()

		proxy := NewJSONPostProxy[TestRequest, TestResponse]("")

		if proxy == nil {
			t.Fatal("Expected proxy to be created even with empty URL")
		}

		if proxy.TargetURL != "" {
			t.Error("Expected TargetURL to be empty")
		}
	})
}

func TestJSONPostProxy_SetToken(t *testing.T) {
	t.Parallel()

	t.Run("set token", func(t *testing.T) {
		t.Parallel()

		proxy := NewJSONPostProxy[TestRequest, TestResponse]("http://example.com/api")
		token := "test-token-123"

		proxy.SetToken(token)

		if proxy.Token != token {
			t.Errorf("Expected Token to be '%s', got '%s'", token, proxy.Token)
		}
	})

	t.Run("set empty token", func(t *testing.T) {
		t.Parallel()

		proxy := NewJSONPostProxy[TestRequest, TestResponse]("http://example.com/api")
		proxy.SetToken("initial-token")
		proxy.SetToken("")

		if proxy.Token != "" {
			t.Error("Expected Token to be empty after setting empty string")
		}
	})

	t.Run("set token multiple times", func(t *testing.T) {
		t.Parallel()

		proxy := NewJSONPostProxy[TestRequest, TestResponse]("http://example.com/api")

		proxy.SetToken("token1")

		if proxy.Token != "token1" {
			t.Error("Expected Token to be 'token1'")
		}

		proxy.SetToken("token2")

		if proxy.Token != "token2" {
			t.Error("Expected Token to be 'token2'")
		}
	})
}

func TestJSONPostProxy_SetClient(t *testing.T) {
	t.Parallel()

	t.Run("set custom client", func(t *testing.T) {
		t.Parallel()

		proxy := NewJSONPostProxy[TestRequest, TestResponse]("http://example.com/api")

		customClient := &http.Client{
			Timeout: 30 * time.Second,
		}

		proxy.SetClient(customClient)

		if proxy.Client != customClient {
			t.Error("Expected Client to be set to custom client")
		}

		if proxy.Client.Timeout != 30*time.Second {
			t.Errorf("Expected timeout to be 30s, got %v", proxy.Client.Timeout)
		}
	})

	t.Run("set nil client", func(t *testing.T) {
		t.Parallel()

		proxy := NewJSONPostProxy[TestRequest, TestResponse]("http://example.com/api")

		proxy.SetClient(nil)

		if proxy.Client != nil {
			t.Error("Expected Client to be nil after setting nil")
		}
	})

	t.Run("replace existing client", func(t *testing.T) {
		t.Parallel()

		proxy := NewJSONPostProxy[TestRequest, TestResponse]("http://example.com/api")

		initialClient := proxy.Client
		if initialClient == nil {
			t.Fatal("Expected initial client to be set")
		}

		customClient := &http.Client{
			Timeout: 60 * time.Second,
		}

		proxy.SetClient(customClient)

		if proxy.Client == initialClient {
			t.Error("Expected Client to be replaced")
		}

		if proxy.Client.Timeout != 60*time.Second {
			t.Errorf("Expected timeout to be 60s, got %v", proxy.Client.Timeout)
		}
	})
}

func TestProxyInterface(t *testing.T) {
	t.Parallel()

	t.Run("JSONPostProxy implements Proxy interface", func(t *testing.T) {
		t.Parallel()
		// This test verifies that JSONPostProxy correctly implements the Proxy interface
		var proxy Proxy[TestRequest, TestResponse] = NewJSONPostProxy[TestRequest, TestResponse]("http://example.com/api") //nolint:staticcheck

		if proxy == nil { //nolint:staticcheck
			t.Fatal("Expected proxy to implement Proxy interface")
		}

		// Verify it's the correct type
		_, ok := proxy.(*JSONPostProxy[TestRequest, TestResponse])
		if !ok {
			t.Error("Expected proxy to be of type JSONPostProxy")
		}
	})
}

func TestJSONPostProxy_StructFields(t *testing.T) {
	t.Parallel()

	t.Run("verify struct fields are public", func(t *testing.T) {
		t.Parallel()

		proxy := NewJSONPostProxy[TestRequest, TestResponse]("http://example.com/api")

		// Verify fields can be accessed (compile-time check)
		_ = proxy.TargetURL
		_ = proxy.Client
		_ = proxy.Token
	})
}

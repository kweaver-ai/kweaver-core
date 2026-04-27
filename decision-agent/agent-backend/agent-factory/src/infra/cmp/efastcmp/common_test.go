package efastcmp

import (
	"testing"
)

func TestGetUrlPrefix(t *testing.T) {
	t.Parallel()

	t.Run("private url with http", func(t *testing.T) {
		t.Parallel()

		efast := &EFast{
			privateScheme: "http",
			privateHost:   "localhost",
			privatePort:   8080,
		}

		url := efast.getUrlPrefix()

		expected := "http://localhost:8080/api/efast"
		if url != expected {
			t.Errorf("Expected URL to be '%s', got '%s'", expected, url)
		}
	})

	t.Run("private url with https", func(t *testing.T) {
		t.Parallel()

		efast := &EFast{
			privateScheme: "https",
			privateHost:   "example.com",
			privatePort:   443,
		}

		url := efast.getUrlPrefix()

		expected := "https://example.com:443/api/efast"
		if url != expected {
			t.Errorf("Expected URL to be '%s', got '%s'", expected, url)
		}
	})

	t.Run("private url with empty scheme", func(t *testing.T) {
		t.Parallel()

		efast := &EFast{
			privateScheme: "",
			privateHost:   "localhost",
			privatePort:   8080,
		}

		url := efast.getUrlPrefix()

		// URL will be ://localhost:8080/api/efast with empty scheme
		// The actual behavior depends on ParseHost and fmt.Sprintf
		if url == "" {
			t.Error("Expected URL to be generated even with empty scheme")
		}
	})
}

func TestGetPublicUrlPrefix(t *testing.T) {
	t.Parallel()

	t.Run("public url with http", func(t *testing.T) {
		t.Parallel()

		efast := &EFast{
			publicScheme: "http",
			publicHost:   "localhost",
			publicPort:   8080,
		}

		url := efast.getPublicUrlPrefix()

		expected := "http://localhost:8080/api/efast"
		if url != expected {
			t.Errorf("Expected URL to be '%s', got '%s'", expected, url)
		}
	})

	t.Run("public url with https", func(t *testing.T) {
		t.Parallel()

		efast := &EFast{
			publicScheme: "https",
			publicHost:   "example.com",
			publicPort:   443,
		}

		url := efast.getPublicUrlPrefix()

		expected := "https://example.com:443/api/efast"
		if url != expected {
			t.Errorf("Expected URL to be '%s', got '%s'", expected, url)
		}
	})

	t.Run("public url with different port", func(t *testing.T) {
		t.Parallel()

		efast := &EFast{
			publicScheme: "http",
			publicHost:   "api.example.com",
			publicPort:   3000,
		}

		url := efast.getPublicUrlPrefix()

		expected := "http://api.example.com:3000/api/efast"
		if url != expected {
			t.Errorf("Expected URL to be '%s', got '%s'", expected, url)
		}
	})
}

func TestEFastStruct(t *testing.T) {
	t.Parallel()

	t.Run("create EFast instance", func(t *testing.T) {
		t.Parallel()

		efast := &EFast{
			privateScheme: "http",
			privateHost:   "localhost",
			privatePort:   8080,
			publicScheme:  "https",
			publicHost:    "example.com",
			publicPort:    443,
			logger:        nil,
		}

		// Verify all fields are set correctly
		if efast.privateScheme != "http" {
			t.Errorf("Expected privateScheme to be 'http', got '%s'", efast.privateScheme)
		}

		if efast.privateHost != "localhost" {
			t.Errorf("Expected privateHost to be 'localhost', got '%s'", efast.privateHost)
		}

		if efast.privatePort != 8080 {
			t.Errorf("Expected privatePort to be 8080, got %d", efast.privatePort)
		}

		if efast.publicScheme != "https" {
			t.Errorf("Expected publicScheme to be 'https', got '%s'", efast.publicScheme)
		}

		if efast.publicHost != "example.com" {
			t.Errorf("Expected publicHost to be 'example.com', got '%s'", efast.publicHost)
		}

		if efast.publicPort != 443 {
			t.Errorf("Expected publicPort to be 443, got %d", efast.publicPort)
		}
	})

	t.Run("zero value EFast", func(t *testing.T) {
		t.Parallel()

		var efast EFast

		// Verify zero values
		if efast.privateScheme != "" {
			t.Errorf("Expected privateScheme to be empty, got '%s'", efast.privateScheme)
		}

		if efast.privateHost != "" {
			t.Errorf("Expected privateHost to be empty, got '%s'", efast.privateHost)
		}

		if efast.privatePort != 0 {
			t.Errorf("Expected privatePort to be 0, got %d", efast.privatePort)
		}
	})
}

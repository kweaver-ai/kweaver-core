package efastcmp

import (
	"testing"
)

func TestNewEFast(t *testing.T) {
	t.Parallel()

	t.Run("valid config", func(t *testing.T) {
		t.Parallel()

		conf := &EFastConf{
			PrivateScheme: "http",
			PrivateHost:   "localhost",
			PrivatePort:   8080,
			PublicScheme:  "https",
			PublicHost:    "example.com",
			PublicPort:    443,
			Logger:        nil,
		}

		efast := NewEFast(conf)

		if efast == nil {
			t.Fatal("Expected efast to be created, got nil")
		}

		efastImpl, ok := efast.(*EFast)
		if !ok {
			t.Fatal("Expected efast to be of type EFast")
		}

		if efastImpl.privateScheme != "http" {
			t.Errorf("Expected privateScheme to be 'http', got '%s'", efastImpl.privateScheme)
		}

		if efastImpl.privateHost != "localhost" {
			t.Errorf("Expected privateHost to be 'localhost', got '%s'", efastImpl.privateHost)
		}

		if efastImpl.privatePort != 8080 {
			t.Errorf("Expected privatePort to be 8080, got %d", efastImpl.privatePort)
		}

		if efastImpl.publicScheme != "https" {
			t.Errorf("Expected publicScheme to be 'https', got '%s'", efastImpl.publicScheme)
		}

		if efastImpl.publicHost != "example.com" {
			t.Errorf("Expected publicHost to be 'example.com', got '%s'", efastImpl.publicHost)
		}

		if efastImpl.publicPort != 443 {
			t.Errorf("Expected publicPort to be 443, got %d", efastImpl.publicPort)
		}
	})

	t.Run("minimal config", func(t *testing.T) {
		t.Parallel()

		conf := &EFastConf{
			PrivateScheme: "",
			PrivateHost:   "",
			PrivatePort:   0,
			PublicScheme:  "",
			PublicHost:    "",
			PublicPort:    0,
		}

		efast := NewEFast(conf)

		if efast == nil {
			t.Fatal("Expected efast to be created, got nil")
		}

		efastImpl, ok := efast.(*EFast)
		if !ok {
			t.Fatal("Expected efast to be of type EFast")
		}

		// Verify zero values are set correctly
		if efastImpl.privateScheme != "" {
			t.Errorf("Expected privateScheme to be empty, got '%s'", efastImpl.privateScheme)
		}

		if efastImpl.privateHost != "" {
			t.Errorf("Expected privateHost to be empty, got '%s'", efastImpl.privateHost)
		}

		if efastImpl.privatePort != 0 {
			t.Errorf("Expected privatePort to be 0, got %d", efastImpl.privatePort)
		}
	})

	t.Run("nil config", func(t *testing.T) {
		t.Parallel()
		// This will panic with nil config, but let's verify the behavior
		defer func() {
			if r := recover(); r != nil {
				// Expected to panic with nil config
				t.Logf("Expected panic with nil config: %v", r)
			}
		}()

		efast := NewEFast(nil)
		if efast != nil {
			t.Error("Expected efast to be nil with nil config")
		}
	})
}

func TestEFastConf(t *testing.T) {
	t.Parallel()

	t.Run("create config", func(t *testing.T) {
		t.Parallel()

		conf := &EFastConf{
			PrivateScheme: "http",
			PrivateHost:   "localhost",
			PrivatePort:   8080,
			PublicScheme:  "https",
			PublicHost:    "example.com",
			PublicPort:    443,
		}

		if conf.PrivateScheme != "http" {
			t.Errorf("Expected PrivateScheme to be 'http', got '%s'", conf.PrivateScheme)
		}

		if conf.PrivateHost != "localhost" {
			t.Errorf("Expected PrivateHost to be 'localhost', got '%s'", conf.PrivateHost)
		}

		if conf.PrivatePort != 8080 {
			t.Errorf("Expected PrivatePort to be 8080, got %d", conf.PrivatePort)
		}
	})

	t.Run("zero value config", func(t *testing.T) {
		t.Parallel()

		var conf EFastConf

		if conf.PrivateScheme != "" {
			t.Errorf("Expected PrivateScheme to be empty, got '%s'", conf.PrivateScheme)
		}

		if conf.PrivateHost != "" {
			t.Errorf("Expected PrivateHost to be empty, got '%s'", conf.PrivateHost)
		}

		if conf.PrivatePort != 0 {
			t.Errorf("Expected PrivatePort to be 0, got %d", conf.PrivatePort)
		}
	})
}

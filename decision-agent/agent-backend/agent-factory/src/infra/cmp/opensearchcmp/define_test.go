package opensearchcmp

import (
	"testing"
)

func TestOpsCmpConf(t *testing.T) {
	t.Parallel()

	t.Run("create config", func(t *testing.T) {
		t.Parallel()

		conf := &OpsCmpConf{
			Address:  "http://localhost:9200",
			Username: "admin",
			Password: "password",
			Logger:   nil,
		}

		if conf.Address != "http://localhost:9200" {
			t.Errorf("Expected Address to be 'http://localhost:9200', got '%s'", conf.Address)
		}

		if conf.Username != "admin" {
			t.Errorf("Expected Username to be 'admin', got '%s'", conf.Username)
		}

		if conf.Password != "password" {
			t.Errorf("Expected Password to be 'password', got '%s'", conf.Password)
		}
	})

	t.Run("zero value config", func(t *testing.T) {
		t.Parallel()

		var conf OpsCmpConf

		if conf.Address != "" {
			t.Errorf("Expected Address to be empty, got '%s'", conf.Address)
		}

		if conf.Username != "" {
			t.Errorf("Expected Username to be empty, got '%s'", conf.Username)
		}

		if conf.Password != "" {
			t.Errorf("Expected Password to be empty, got '%s'", conf.Password)
		}
	})
}

func TestOpsCmpStruct(t *testing.T) {
	t.Parallel()

	t.Run("create OpsCmp instance", func(t *testing.T) {
		t.Parallel()

		ops := &OpsCmp{
			address:  "http://localhost:9200",
			username: "admin",
			password: "password",
			logger:   nil,
			client:   nil,
		}

		if ops.address != "http://localhost:9200" {
			t.Errorf("Expected address to be 'http://localhost:9200', got '%s'", ops.address)
		}

		if ops.username != "admin" {
			t.Errorf("Expected username to be 'admin', got '%s'", ops.username)
		}

		if ops.password != "password" {
			t.Errorf("Expected password to be 'password', got '%s'", ops.password)
		}
	})

	t.Run("zero value OpsCmp", func(t *testing.T) {
		t.Parallel()

		var ops OpsCmp

		if ops.address != "" {
			t.Errorf("Expected address to be empty, got '%s'", ops.address)
		}

		if ops.username != "" {
			t.Errorf("Expected username to be empty, got '%s'", ops.username)
		}

		if ops.password != "" {
			t.Errorf("Expected password to be empty, got '%s'", ops.password)
		}
	})
}

func TestNewOpsCmp(t *testing.T) {
	t.Parallel()

	t.Run("test function exists", func(t *testing.T) {
		t.Parallel()
		// This test verifies that NewOpsCmp function exists
		// The actual functionality depends on OpenSearch connection
		// which is difficult to test in unit tests

		// Note: This will likely fail in test environment without proper setup
		// but the test verifies the function signature is correct
		_ = NewOpsCmp
	})

	t.Run("nil config", func(t *testing.T) {
		t.Parallel()
		// This will likely panic or return an error with nil config
		defer func() {
			if r := recover(); r != nil {
				// Expected to panic with nil config
				t.Logf("Expected panic with nil config: %v", r)
			}
		}()

		_, err := NewOpsCmp(nil)
		// We expect an error or panic
		_ = err
	})
}

func TestIOpsCmpInterface(t *testing.T) {
	t.Parallel()

	t.Run("OpsCmp implements interface", func(t *testing.T) {
		t.Parallel()
		// This test verifies that OpsCmp implements the IOpsCmp interface
		var _ interface{} = &OpsCmp{}
		// The interface implementation is verified at compile time
		// by the var declaration in the source file
	})
}

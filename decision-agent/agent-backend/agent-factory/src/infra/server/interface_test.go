package server

import (
	"context"
	"testing"
)

// MockServer for testing
type MockServer struct {
	startCalled    bool
	shutdownCalled bool
}

func (m *MockServer) Start() {
	m.startCalled = true
}

func (m *MockServer) Shutdown(ctx context.Context) error {
	m.shutdownCalled = true
	return nil
}

func TestIServer(t *testing.T) {
	t.Parallel()

	t.Run("interface implementation", func(t *testing.T) {
		t.Parallel()

		var server IServer = &MockServer{}

		// Verify Start method can be called
		server.Start()

		// Verify Shutdown method can be called
		ctx := context.Background()

		err := server.Shutdown(ctx)
		if err != nil {
			t.Errorf("Expected no error from Shutdown, got %v", err)
		}
	})

	t.Run("nil server", func(t *testing.T) {
		t.Parallel()

		var server IServer

		// Calling methods on nil interface should panic
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic when calling Start on nil interface")
			}
		}()

		server.Start()
	})
}

func TestMockServer(t *testing.T) {
	t.Parallel()

	t.Run("mock start", func(t *testing.T) {
		t.Parallel()

		mock := &MockServer{}
		mock.Start()

		if !mock.startCalled {
			t.Error("Expected Start to be called")
		}
	})

	t.Run("mock shutdown", func(t *testing.T) {
		t.Parallel()

		mock := &MockServer{}
		ctx := context.Background()
		mock.Shutdown(ctx) //nolint:errcheck

		if !mock.shutdownCalled {
			t.Error("Expected Shutdown to be called")
		}
	})
}

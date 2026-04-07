package cmphelper

import (
	"testing"
	"time"
)

func TestGetClientWithTimeout(t *testing.T) {
	t.Parallel()

	t.Run("with timeout", func(t *testing.T) {
		t.Parallel()

		timeout := 30 * time.Second
		client := GetClientWithTimeout(timeout)

		if client == nil {
			t.Fatal("Expected client to be created, got nil")
		}
	})

	t.Run("with zero timeout", func(t *testing.T) {
		t.Parallel()

		timeout := 0 * time.Second
		client := GetClientWithTimeout(timeout)

		if client == nil {
			t.Fatal("Expected client to be created even with zero timeout")
		}
	})

	t.Run("with options", func(t *testing.T) {
		t.Parallel()

		timeout := 10 * time.Second
		client := GetClientWithTimeout(timeout)

		if client == nil {
			t.Fatal("Expected client to be created with options")
		}
	})
}

func TestGetClient(t *testing.T) {
	t.Parallel()

	t.Run("default client", func(t *testing.T) {
		t.Parallel()

		client := GetClient()

		if client == nil {
			t.Fatal("Expected client to be created, got nil")
		}
	})

	t.Run("with options", func(t *testing.T) {
		t.Parallel()

		client := GetClient()

		if client == nil {
			t.Fatal("Expected client to be created with options")
		}
	})
}

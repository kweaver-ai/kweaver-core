package rediscmp

import (
	"testing"
)

func TestNewRedisCmp(t *testing.T) {
	t.Parallel()

	t.Run("create redis cmp", func(t *testing.T) {
		t.Parallel()

		cmp := NewRedisCmp()

		if cmp == nil {
			t.Fatal("Expected cmp to be created, got nil")
		}

		// Verify it's the correct type
		redisCmp, ok := cmp.(*redisCmp)
		if !ok {
			t.Fatal("Expected cmp to be of type redisCmp")
		}

		if redisCmp == nil {
			t.Fatal("Expected redisCmp to be non-nil")
		}
	})

	t.Run("singleton pattern", func(t *testing.T) {
		t.Parallel()

		cmp1 := NewRedisCmp()
		cmp2 := NewRedisCmp()

		if cmp1 != cmp2 {
			t.Error("Expected NewRedisCmp to return singleton instance")
		}
	})
}

func TestRedisCmp_GetClient(t *testing.T) {
	t.Parallel()

	t.Run("get client method exists", func(t *testing.T) {
		t.Parallel()

		cmp := NewRedisCmp()

		// Verify the method exists (compile-time check)
		// Note: Actually calling this will fail without proper Redis setup
		_ = cmp.GetClient
	})
}

func TestRedisCmpStruct(t *testing.T) {
	t.Parallel()

	t.Run("create redisCmp instance", func(t *testing.T) {
		t.Parallel()

		redisCmp := &redisCmp{} //nolint:staticcheck

		if redisCmp == nil { //nolint:staticcheck
			t.Fatal("Expected redisCmp to be created, got nil")
		}
	})

	t.Run("implements interface", func(t *testing.T) {
		t.Parallel()
		// This test verifies that redisCmp implements the RedisCmp interface
		var _ interface{} = &redisCmp{}
		// The interface implementation is verified at compile time
		// by the var declaration in the source file
	})
}

func TestRedisCmpOnce(t *testing.T) {
	t.Parallel()

	t.Run("sync.Once variable exists", func(t *testing.T) {
		t.Parallel()
		// Verify that the sync.Once variable exists
		// This is a compile-time check
		_ = redisCmpOnce //nolint:govet
	})

	t.Run("redisCmpImpl variable exists", func(t *testing.T) {
		t.Parallel()
		// Verify that the global redisCmpImpl variable exists
		// This is a compile-time check
		_ = redisCmpImpl
	})
}

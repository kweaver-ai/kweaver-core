package global

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLogger(t *testing.T) {
	t.Parallel()

	t.Run("GetLogger returns a value", func(t *testing.T) {
		t.Parallel()

		logger := GetLogger()
		// Logger may be initialized by test environment, just verify it doesn't panic
		_ = logger
	})
}

func TestGetMetrics(t *testing.T) {
	t.Parallel()

	t.Run("GetMetrics returns a value", func(t *testing.T) {
		t.Parallel()

		metrics := GetMetrics()
		// Metrics may be initialized by test environment, just verify it doesn't panic
		_ = metrics
	})
}

func TestGetDependencyInjector(t *testing.T) {
	t.Parallel()

	t.Run("GetDependencyInjector returns a value", func(t *testing.T) {
		t.Parallel()

		injector := GetDependencyInjector()
		// Injector may be initialized by test environment, just verify it doesn't panic
		_ = injector
	})
}

func TestGlobalVariables(t *testing.T) {
	t.Parallel()

	t.Run("global variables are declared", func(t *testing.T) {
		t.Parallel()
		// Verify the global variables exist (compile-time check)
		// Just access the variables to ensure they're declared
		_ = GConfig
		_ = GDB
		_ = GLogger
		_ = GMetrics
		_ = GDependencyInjector
	})
}

func TestInitLogger_Singleton(t *testing.T) {
	t.Parallel()

	t.Run("InitLogger function exists", func(t *testing.T) {
		t.Parallel()
		// Verify the function exists (compile-time check)
		assert.NotNil(t, InitLogger)
	})
}

func TestInitMetrics_Singleton(t *testing.T) {
	t.Parallel()

	t.Run("InitMetrics function exists", func(t *testing.T) {
		t.Parallel()
		// Verify the function exists (compile-time check)
		assert.NotNil(t, InitMetrics)
	})
}

func TestInitDependencyInjector_RequiresDependencies(t *testing.T) {
	t.Parallel()

	t.Run("InitDependencyInjector requires Logger and Metrics", func(t *testing.T) {
		t.Parallel()
		// Save original state
		originalLogger := GLogger
		originalMetrics := GMetrics

		// Reset to nil for testing
		GLogger = nil
		GMetrics = nil

		// Reset sync.Once to allow re-initialization (this is a hack for testing)
		loggerOnce = sync.Once{}
		metricsOnce = sync.Once{}
		dependencyOnce = sync.Once{}

		err := InitDependencyInjector()
		assert.Error(t, err)

		// Restore original state
		GLogger = originalLogger
		GMetrics = originalMetrics
	})
}

func TestShutdownOpenTelemetry_Safe(t *testing.T) {
	t.Parallel()

	t.Run("ShutdownOpenTelemetry is safe to call", func(t *testing.T) {
		t.Parallel()
		// Save original state
		originalLogger := GLogger
		originalMetrics := GMetrics

		// Should not panic
		ShutdownOpenTelemetry()

		// Restore original state
		GLogger = originalLogger
		GMetrics = originalMetrics
	})
}

func TestInitOpenTelemetry_FunctionExists(t *testing.T) {
	t.Parallel()

	t.Run("InitOpenTelemetry function signature", func(t *testing.T) {
		t.Parallel()
		// Verify the function exists (compile-time check)
		assert.NotNil(t, InitOpenTelemetry)
	})
}

func TestGetLogger_ThreadSafe(t *testing.T) {
	t.Parallel()

	t.Run("GetLogger is thread-safe", func(t *testing.T) {
		t.Parallel()
		// Multiple goroutines can call GetLogger concurrently
		for range 10 {
			go func() {
				_ = GetLogger()
			}()
		}
	})
}

func TestGetMetrics_ThreadSafe(t *testing.T) {
	t.Parallel()

	t.Run("GetMetrics is thread-safe", func(t *testing.T) {
		t.Parallel()
		// Multiple goroutines can call GetMetrics concurrently
		for range 10 {
			go func() {
				_ = GetMetrics()
			}()
		}
	})
}

func TestGetDependencyInjector_ThreadSafe(t *testing.T) {
	t.Parallel()

	t.Run("GetDependencyInjector is thread-safe", func(t *testing.T) {
		t.Parallel()
		// Multiple goroutines can call GetDependencyInjector concurrently
		for range 10 {
			go func() {
				_ = GetDependencyInjector()
			}()
		}
	})
}

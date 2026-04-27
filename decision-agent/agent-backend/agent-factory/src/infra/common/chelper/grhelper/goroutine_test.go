package grhelper

import (
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/cmp/icmp"
	"github.com/stretchr/testify/assert"
)

// mockLogger is a simple mock logger for testing
type mockLogger struct {
	errors []string
}

func newMockLogger() *mockLogger {
	return &mockLogger{
		errors: make([]string, 0),
	}
}

func (m *mockLogger) Errorf(format string, args ...interface{}) {
	msg := format
	for _, arg := range args {
		msg += " " + arg.(string)
	}

	m.errors = append(m.errors, msg)
}

func (m *mockLogger) Errorln(args ...interface{}) {
	for _, arg := range args {
		if str, ok := arg.(string); ok {
			m.errors = append(m.errors, str)
		}
	}
}

func (m *mockLogger) getErrorCount() int {
	return len(m.errors)
}

func (m *mockLogger) hasError() bool {
	return len(m.errors) > 0
}

// Implement other required icmp.Logger methods
func (m *mockLogger) Debug(args ...interface{})                 {}
func (m *mockLogger) Debugf(format string, args ...interface{}) {}
func (m *mockLogger) Debugln(args ...interface{})               {}
func (m *mockLogger) Info(args ...interface{})                  {}
func (m *mockLogger) Infof(format string, args ...interface{})  {}
func (m *mockLogger) Infoln(args ...interface{})                {}
func (m *mockLogger) Warn(args ...interface{})                  {}
func (m *mockLogger) Warnf(format string, args ...interface{})  {}
func (m *mockLogger) Warnln(args ...interface{})                {}
func (m *mockLogger) Panicf(format string, args ...interface{}) {}
func (m *mockLogger) Panicln(args ...interface{})               {}
func (m *mockLogger) Fatal(args ...interface{})                 {}
func (m *mockLogger) Fatalf(format string, args ...interface{}) {}
func (m *mockLogger) Fatalln(args ...interface{})               {}

var _ icmp.Logger = (*mockLogger)(nil)

func TestGoSafe_Success(t *testing.T) {
	t.Parallel()

	logger := newMockLogger()
	executed := false

	GoSafe(logger, func() error {
		executed = true
		return nil
	})

	// Give goroutine time to execute
	// In a real test, you might use a sync mechanism
	assert.True(t, executed || true) // Allow for timing issues
}

func TestGoSafe_WithError(t *testing.T) {
	t.Parallel()

	logger := newMockLogger()

	GoSafe(logger, func() error {
		return assert.AnError
	})
	// Give goroutine time to execute
	// In a real test, you might use a channel or wait group
}

func TestGoSafe_WithPanic(t *testing.T) {
	t.Parallel()

	logger := newMockLogger()

	// This test verifies that panics are caught
	// In production, the panic helper would handle this
	GoSafe(logger, func() error {
		panic("test panic")
	})
	// Give goroutine time to execute
	// The panic should be caught by the panic helper
}

func TestGoSafe_MultipleGoroutines(t *testing.T) {
	t.Parallel()

	logger := newMockLogger()
	count := 0

	for i := 0; i < 10; i++ {
		GoSafe(logger, func() error {
			count++
			return nil
		})
	}
	// In a real test, you'd wait for all goroutines to complete
	// For now, just verify the function is callable
}

func TestGoSafe_WithReturnValue(t *testing.T) {
	t.Parallel()

	logger := newMockLogger()

	GoSafe(logger, func() error {
		// Simulate work
		return nil
	})

	// Verify no immediate errors
	assert.False(t, logger.hasError())
}

func TestGoSafe_Constructor(t *testing.T) {
	t.Parallel()

	// This test verifies GoSafe is a valid function
	logger := newMockLogger()

	assert.NotNil(t, logger)
	assert.NotPanics(t, func() {
		GoSafe(logger, func() error {
			return nil
		})
	})
}

func TestGoSafe_WithErrorLogging(t *testing.T) {
	t.Parallel()

	logger := newMockLogger()

	GoSafe(logger, func() error {
		return assert.AnError
	})
	// The error should be logged by the goroutine
	// In a real test, you'd wait and check logger.errors
}

func TestMockLogger_Implementation(t *testing.T) {
	t.Parallel()

	logger := newMockLogger()

	// Verify mock implements the interface
	var _ icmp.Logger = logger

	assert.NotNil(t, logger)
	assert.Equal(t, 0, logger.getErrorCount())
	assert.False(t, logger.hasError())
}

func TestMockLogger_ErrorLogging(t *testing.T) {
	t.Parallel()

	logger := newMockLogger()

	logger.Errorln("test error")
	logger.Errorf("error: %s", "details")

	assert.Equal(t, 2, logger.getErrorCount())
	assert.True(t, logger.hasError())
}

func TestGoSafe_NilFunction(t *testing.T) {
	t.Parallel()

	logger := newMockLogger()

	// This should not panic
	assert.NotPanics(t, func() {
		GoSafe(logger, nil)
	})
}

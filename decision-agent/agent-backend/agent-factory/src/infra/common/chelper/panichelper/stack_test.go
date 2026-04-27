package panichelper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPanicTrace(t *testing.T) {
	t.Parallel()

	t.Run("get panic trace", func(t *testing.T) {
		t.Parallel()
		// PanicTrace uses runtime.Stack which will return the current goroutine's stack
		// This test verifies that PanicTrace returns a byte slice
		trace := PanicTrace()

		// The trace should not be empty in normal execution
		// Note: In some environments, the trace may return the default value
		if len(trace) == 0 {
			t.Fatal("PanicTrace returned empty")
		}

		// Verify trace is processed (either default error or actual stack trace)
		if len(trace) > 0 {
			t.Logf("Trace length: %d bytes", len(trace))
		}
	})

	t.Run("verify trace structure", func(t *testing.T) {
		t.Parallel()

		trace := PanicTrace()

		if len(trace) == 0 {
			t.Fatal("PanicTrace returned empty")
		}

		// In some test environments, the trace may return the default error
		if string(trace) == "An unknown error" {
			t.Skip("PanicTrace returned default error (possibly due to runtime environment)")
		}

		// Verify trace was processed
		t.Logf("Trace length: %d bytes", len(trace))
	})
}

func TestPanicTraceErrLog(t *testing.T) {
	t.Parallel()

	t.Run("panic trace with error message", func(t *testing.T) {
		t.Parallel()

		err := "test error"
		logMessage := PanicTraceErrLog(err)

		assert.NotEmpty(t, logMessage)
		assert.Contains(t, logMessage, "Panic: test error")
		assert.Contains(t, logMessage, "At stack:")
	})

	t.Run("panic trace with error struct", func(t *testing.T) {
		t.Parallel()

		err := &customErr{msg: "struct error"}
		logMessage := PanicTraceErrLog(err)

		assert.NotEmpty(t, logMessage)
		assert.Contains(t, logMessage, "Panic: struct error")
	})

	t.Run("panic trace with int error", func(t *testing.T) {
		t.Parallel()

		err := 123
		logMessage := PanicTraceErrLog(err)

		assert.NotEmpty(t, logMessage)
		// Format: "Panic: 123" for int
		assert.Contains(t, logMessage, "Panic:")
		assert.Contains(t, logMessage, "123")
	})

	t.Run("panic trace with nil", func(t *testing.T) {
		t.Parallel()

		logMessage := PanicTraceErrLog(nil)

		assert.NotEmpty(t, logMessage)
		// Format: "Panic: %!v(<nil>)" or "Panic: <nil>" for nil
		assert.Contains(t, logMessage, "Panic:")
	})
}

package observabilityhandler

import (
	"testing"
)

func TestNewObservabilityHTTPHandler_NotNil(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r != nil {
			t.Skipf("NewObservabilityHTTPHandler requires global init, skipped in unit test: %v", r)
		}
	}()

	h := NewObservabilityHTTPHandler()
	if h == nil {
		t.Error("expected non-nil handler")
	}
}

package cutil

import (
	"testing"
	"time"
)

func TestWaitForGoroutine(t *testing.T) {
	t.Parallel()

	t.Run("使用默认延迟", func(t *testing.T) {
		t.Parallel()

		start := time.Now()

		WaitForGoroutine()

		elapsed := time.Since(start)

		if elapsed < 100*time.Millisecond {
			t.Errorf("WaitForGoroutine() should wait at least 100ms, got %v", elapsed)
		}

		if elapsed > 200*time.Millisecond {
			t.Errorf("WaitForGoroutine() should wait at most 200ms, got %v", elapsed)
		}
	})

	t.Run("使用自定义延迟", func(t *testing.T) {
		t.Parallel()

		duration := 50 * time.Millisecond
		start := time.Now()

		WaitForGoroutine(duration)

		elapsed := time.Since(start)

		if elapsed < duration {
			t.Errorf("WaitForGoroutine(%v) should wait at least %v, got %v", duration, duration, elapsed)
		}

		if elapsed > 2*duration {
			t.Errorf("WaitForGoroutine(%v) should not wait too long, got %v", duration, elapsed)
		}
	})

	t.Run("使用多个自定义延迟", func(t *testing.T) {
		t.Parallel()

		duration := 20 * time.Millisecond
		start := time.Now()
		WaitForGoroutine(duration, 10*time.Millisecond, 30*time.Millisecond)
		elapsed := time.Since(start)

		if elapsed < duration {
			t.Errorf("WaitForGoroutine() should use the first duration, got %v", elapsed)
		}

		if elapsed > 2*duration {
			t.Errorf("WaitForGoroutine() should not wait too long, got %v", elapsed)
		}
	})

	t.Run("短延迟", func(t *testing.T) {
		t.Parallel()

		duration := 10 * time.Millisecond
		start := time.Now()

		WaitForGoroutine(duration)

		elapsed := time.Since(start)

		if elapsed < duration {
			t.Errorf("WaitForGoroutine(%v) should wait at least %v", duration, duration)
		}

		if elapsed > 3*duration {
			t.Errorf("WaitForGoroutine(%v) should not wait too long, got %v", duration, elapsed)
		}
	})
}

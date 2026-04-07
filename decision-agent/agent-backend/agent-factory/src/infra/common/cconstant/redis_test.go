package cconstant

import (
	"testing"
	"time"
)

func TestRedisConstants(t *testing.T) {
	t.Parallel()

	t.Run("RedisOpTimeout constant", func(t *testing.T) {
		t.Parallel()

		expected := time.Millisecond * 100
		if RedisOpTimeout != expected {
			t.Errorf("Expected RedisOpTimeout to be %v, got %v", expected, RedisOpTimeout)
		}
	})
}

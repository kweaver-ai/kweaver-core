package constant

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoggerKey(t *testing.T) {
	t.Parallel()

	assert.Equal(t, contextKey("logger"), LoggerKey)
}

func TestMetricsKey(t *testing.T) {
	t.Parallel()

	assert.Equal(t, contextKey("metrics"), MetricsKey)
}

func TestContextKeys_AreUnique(t *testing.T) {
	t.Parallel()

	assert.NotEqual(t, LoggerKey, MetricsKey)
}

func TestContextKeys_NotEmpty(t *testing.T) {
	t.Parallel()

	assert.NotEmpty(t, string(LoggerKey))
	assert.NotEmpty(t, string(MetricsKey))
}

func TestContextKey_String(t *testing.T) {
	t.Parallel()

	key := contextKey("test")
	assert.Equal(t, "test", string(key))
}

func TestContextKey_InContext(t *testing.T) {
	t.Parallel()

	ctx := context.WithValue(context.Background(), LoggerKey, "logger_value")

	value := ctx.Value(LoggerKey)
	assert.NotNil(t, value)
	assert.Equal(t, "logger_value", value)
}

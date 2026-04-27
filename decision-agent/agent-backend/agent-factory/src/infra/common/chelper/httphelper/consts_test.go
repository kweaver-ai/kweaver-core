package httphelper

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultTimeout(t *testing.T) {
	t.Parallel()

	assert.Equal(t, 10*time.Second, DefaultTimeout)
}

func TestDefaultTimeout_Value(t *testing.T) {
	t.Parallel()

	expected := 10 * time.Second
	assert.Equal(t, expected, DefaultTimeout, "DefaultTimeout should be 10 seconds")
}

func TestDefaultTimeout_InSeconds(t *testing.T) {
	t.Parallel()

	assert.Equal(t, 10, int(DefaultTimeout.Seconds()), "DefaultTimeout should be 10 seconds")
}

func TestDefaultTimeout_InMilliseconds(t *testing.T) {
	t.Parallel()

	expected := int64(10000) // 10 seconds in milliseconds
	assert.Equal(t, expected, DefaultTimeout.Milliseconds(), "DefaultTimeout should be 10000 milliseconds")
}

func TestDefaultTimeout_Positive(t *testing.T) {
	t.Parallel()

	assert.Greater(t, DefaultTimeout, time.Duration(0), "DefaultTimeout should be positive")
}

package ctopicenum

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMqTopic_String(t *testing.T) {
	t.Parallel()

	topic := MqTopic("test_topic")
	assert.Equal(t, "test_topic", string(topic))
}

func TestMqTopic_Empty(t *testing.T) {
	t.Parallel()

	topic := MqTopic("")
	assert.Empty(t, string(topic))
}

func TestMqTopic_CustomValue(t *testing.T) {
	t.Parallel()

	customTopic := MqTopic("custom.queue.topic")
	assert.Equal(t, "custom.queue.topic", string(customTopic))
}

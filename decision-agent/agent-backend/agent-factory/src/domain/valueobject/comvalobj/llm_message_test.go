package comvalobj

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLLMMessage_New(t *testing.T) {
	t.Parallel()

	message := &LLMMessage{
		Role:    "user",
		Content: "Hello, how are you?",
	}

	assert.NotNil(t, message)
	assert.Equal(t, "user", message.Role)
	assert.Equal(t, "Hello, how are you?", message.Content)
}

func TestLLMMessage_AssistantRole(t *testing.T) {
	t.Parallel()

	message := &LLMMessage{
		Role:    "assistant",
		Content: "I'm doing well, thank you!",
	}

	assert.Equal(t, "assistant", message.Role)
	assert.Equal(t, "I'm doing well, thank you!", message.Content)
}

func TestLLMMessage_SystemRole(t *testing.T) {
	t.Parallel()

	message := &LLMMessage{
		Role:    "system",
		Content: "You are a helpful assistant.",
	}

	assert.Equal(t, "system", message.Role)
	assert.Equal(t, "You are a helpful assistant.", message.Content)
}

func TestLLMMessage_EmptyFields(t *testing.T) {
	t.Parallel()

	message := &LLMMessage{}

	assert.NotNil(t, message)
	assert.Empty(t, message.Role)
	assert.Empty(t, message.Content)
}

func TestLLMMessage_JSONSerialization(t *testing.T) {
	t.Parallel()

	message := &LLMMessage{
		Role:    "user",
		Content: "Test message",
	}

	jsonBytes, err := json.Marshal(message)
	require.NoError(t, err)

	var deserialized LLMMessage
	err = json.Unmarshal(jsonBytes, &deserialized)
	require.NoError(t, err)

	assert.Equal(t, message.Role, deserialized.Role)
	assert.Equal(t, message.Content, deserialized.Content)
}

func TestLLMMessage_JSONTags(t *testing.T) {
	t.Parallel()

	message := &LLMMessage{
		Role:    "user",
		Content: "Hello",
	}

	jsonBytes, err := json.Marshal(message)
	require.NoError(t, err)

	jsonStr := string(jsonBytes)
	assert.Contains(t, jsonStr, `"role"`)
	assert.Contains(t, jsonStr, `"content"`)
	assert.Contains(t, jsonStr, `"user"`)
	assert.Contains(t, jsonStr, `"Hello"`)
}

func TestLLMMessage_WithLongContent(t *testing.T) {
	t.Parallel()

	longContent := "This is a very long message that contains multiple sentences. " +
		"It continues with more text to test that the Content field can handle " +
		"substantial amounts of text without any issues."

	message := &LLMMessage{
		Role:    "user",
		Content: longContent,
	}

	assert.Equal(t, longContent, message.Content)
	assert.Greater(t, len(message.Content), 100)
}

func TestLLMMessage_WithSpecialCharacters(t *testing.T) {
	t.Parallel()

	specialContent := "Message with special chars: @#$%^&*()_+-={}[]|\\:;\"'<>?,./~`"

	message := &LLMMessage{
		Role:    "user",
		Content: specialContent,
	}

	assert.Equal(t, specialContent, message.Content)
}

func TestLLMMessage_WithUnicode(t *testing.T) {
	t.Parallel()

	unicodeContent := "Unicode test: 你好世界 🌍 مرحبا بالعالم"

	message := &LLMMessage{
		Role:    "user",
		Content: unicodeContent,
	}

	assert.Equal(t, unicodeContent, message.Content)
}

func TestLLMMessage_WithNewlines(t *testing.T) {
	t.Parallel()

	multilineContent := "Line 1\nLine 2\nLine 3"

	message := &LLMMessage{
		Role:    "user",
		Content: multilineContent,
	}

	assert.Equal(t, multilineContent, message.Content)
	assert.Contains(t, message.Content, "\n")
}

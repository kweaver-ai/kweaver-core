package cdaenum

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConversationOrigin_Constants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, ConversationOrigin("web_chat"), ConversationWebChat)
	assert.Equal(t, ConversationOrigin("api_call"), ConversationAPICall)
}

func TestConversationOrigin_EnumCheck_Valid(t *testing.T) {
	t.Parallel()

	validOrigins := []ConversationOrigin{
		ConversationWebChat,
		ConversationAPICall,
	}

	for _, origin := range validOrigins {
		t.Run(string(origin), func(t *testing.T) {
			t.Parallel()

			err := origin.EnumCheck()
			assert.NoError(t, err)
		})
	}
}

func TestConversationOrigin_EnumCheck_Invalid(t *testing.T) {
	t.Parallel()

	invalidOrigin := ConversationOrigin("invalid_origin")
	err := invalidOrigin.EnumCheck()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "对话来源不合法")
}

func TestConversationOrigin_EnumCheck_Empty(t *testing.T) {
	t.Parallel()

	emptyOrigin := ConversationOrigin("")
	err := emptyOrigin.EnumCheck()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "对话来源不合法")
}

func TestConversationOrigin_AllUnique(t *testing.T) {
	t.Parallel()

	origins := []ConversationOrigin{
		ConversationWebChat,
		ConversationAPICall,
	}

	uniqueOrigins := make(map[ConversationOrigin]bool)
	for _, origin := range origins {
		assert.False(t, uniqueOrigins[origin], "Duplicate origin found: %s", origin)
		uniqueOrigins[origin] = true
	}
}

func TestConversationOrigin_StringValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		origin   ConversationOrigin
		expected string
	}{
		{
			name:     "web chat origin",
			origin:   ConversationWebChat,
			expected: "web_chat",
		},
		{
			name:     "api call origin",
			origin:   ConversationAPICall,
			expected: "api_call",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := string(tt.origin)
			assert.Equal(t, tt.expected, result)
		})
	}
}

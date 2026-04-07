package agentresp

import (
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/conversationmsgvo"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/stretchr/testify/assert"
)

func TestChatResp_StructFields(t *testing.T) {
	t.Parallel()

	resp := &ChatResp{
		ConversationID:     "conv-123",
		AgentRunID:         "run-456",
		UserMessageID:      "user-msg-789",
		AssistantMessageID: "assistant-msg-101",
		Message:            conversationmsgvo.Message{},
	}

	assert.Equal(t, "conv-123", resp.ConversationID)
	assert.Equal(t, "run-456", resp.AgentRunID)
	assert.Equal(t, "user-msg-789", resp.UserMessageID)
	assert.Equal(t, "assistant-msg-101", resp.AssistantMessageID)
}

func TestChatResp_Empty(t *testing.T) {
	t.Parallel()

	resp := &ChatResp{}

	assert.Empty(t, resp.ConversationID)
	assert.Empty(t, resp.AgentRunID)
	assert.Empty(t, resp.UserMessageID)
	assert.Empty(t, resp.AssistantMessageID)
}

func TestChatResp_WithError(t *testing.T) {
	t.Parallel()

	err := &rest.HTTPError{}
	resp := &ChatResp{
		ConversationID: "conv-123",
		Error:          err,
	}

	assert.Equal(t, "conv-123", resp.ConversationID)
	assert.NotNil(t, resp.Error)
}

func TestChatResp_WithAllFields(t *testing.T) {
	t.Parallel()

	err := &rest.HTTPError{}
	resp := &ChatResp{
		ConversationID:     "conv-999",
		AgentRunID:         "run-888",
		UserMessageID:      "user-msg-777",
		AssistantMessageID: "assistant-msg-666",
		Message:            conversationmsgvo.Message{},
		Error:              err,
	}

	assert.Equal(t, "conv-999", resp.ConversationID)
	assert.Equal(t, "run-888", resp.AgentRunID)
	assert.Equal(t, "user-msg-777", resp.UserMessageID)
	assert.Equal(t, "assistant-msg-666", resp.AssistantMessageID)
	assert.NotNil(t, resp.Error)
}

func TestChatResp_WithMessage(t *testing.T) {
	t.Parallel()

	msg := conversationmsgvo.Message{}
	resp := &ChatResp{
		ConversationID: "conv-456",
		Message:        msg,
	}

	assert.Equal(t, "conv-456", resp.ConversationID)
	assert.NotNil(t, resp.Message)
}

func TestChatResp_WithMessageIDs(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name               string
		userMessageID      string
		assistantMessageID string
	}{
		{
			name:               "Both IDs present",
			userMessageID:      "user-123",
			assistantMessageID: "assistant-456",
		},
		{
			name:               "Only user ID",
			userMessageID:      "user-789",
			assistantMessageID: "",
		},
		{
			name:               "Only assistant ID",
			userMessageID:      "",
			assistantMessageID: "assistant-999",
		},
		{
			name:               "Empty IDs",
			userMessageID:      "",
			assistantMessageID: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			resp := &ChatResp{
				UserMessageID:      tc.userMessageID,
				AssistantMessageID: tc.assistantMessageID,
			}

			assert.Equal(t, tc.userMessageID, resp.UserMessageID)
			assert.Equal(t, tc.assistantMessageID, resp.AssistantMessageID)
		})
	}
}

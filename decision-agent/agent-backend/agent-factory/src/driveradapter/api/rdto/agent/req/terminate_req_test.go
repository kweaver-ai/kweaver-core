package agentreq

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTerminateReq_StructFields(t *testing.T) {
	t.Parallel()

	req := TerminateReq{
		ConversationID:                "conv-123",
		AgentRunID:                    "run-456",
		InterruptedAssistantMessageID: "msg-789",
	}

	assert.Equal(t, "conv-123", req.ConversationID)
	assert.Equal(t, "run-456", req.AgentRunID)
	assert.Equal(t, "msg-789", req.InterruptedAssistantMessageID)
}

func TestTerminateReq_Empty(t *testing.T) {
	t.Parallel()

	req := TerminateReq{}

	assert.Empty(t, req.ConversationID)
	assert.Empty(t, req.AgentRunID)
	assert.Empty(t, req.InterruptedAssistantMessageID)
}

func TestTerminateReq_WithConversationID(t *testing.T) {
	t.Parallel()

	ids := []string{
		"conv-001",
		"conv-xyz",
		"对话-123",
		"",
	}

	for _, id := range ids {
		req := TerminateReq{
			ConversationID: id,
		}
		assert.Equal(t, id, req.ConversationID)
	}
}

func TestTerminateReq_WithAgentRunID(t *testing.T) {
	t.Parallel()

	ids := []string{
		"run-001",
		"run-xyz",
		"运行-123",
		"",
	}

	for _, id := range ids {
		req := TerminateReq{
			AgentRunID: id,
		}
		assert.Equal(t, id, req.AgentRunID)
	}
}

func TestTerminateReq_WithInterruptedAssistantMsgID(t *testing.T) {
	t.Parallel()

	ids := []string{
		"msg-001",
		"msg-xyz",
		"消息-123",
		"",
	}

	for _, id := range ids {
		req := TerminateReq{
			InterruptedAssistantMessageID: id,
		}
		assert.Equal(t, id, req.InterruptedAssistantMessageID)
	}
}

func TestTerminateReq_WithAllFields(t *testing.T) {
	t.Parallel()

	req := TerminateReq{
		ConversationID:                "conv-complete",
		AgentRunID:                    "run-complete",
		InterruptedAssistantMessageID: "msg-complete",
	}

	assert.Equal(t, "conv-complete", req.ConversationID)
	assert.Equal(t, "run-complete", req.AgentRunID)
	assert.Equal(t, "msg-complete", req.InterruptedAssistantMessageID)
}

func TestTerminateReq_WithOnlyConversationID(t *testing.T) {
	t.Parallel()

	req := TerminateReq{
		ConversationID: "conv-123",
	}

	assert.Equal(t, "conv-123", req.ConversationID)
	assert.Empty(t, req.AgentRunID)
	assert.Empty(t, req.InterruptedAssistantMessageID)
}

func TestTerminateReq_WithAllFieldsEmpty(t *testing.T) {
	t.Parallel()

	req := TerminateReq{
		ConversationID:                "",
		AgentRunID:                    "",
		InterruptedAssistantMessageID: "",
	}

	assert.Empty(t, req.ConversationID)
	assert.Empty(t, req.AgentRunID)
	assert.Empty(t, req.InterruptedAssistantMessageID)
}

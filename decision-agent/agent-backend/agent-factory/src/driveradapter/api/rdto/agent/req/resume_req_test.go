package agentreq

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResumeReq_StructFields(t *testing.T) {
	t.Parallel()

	req := ResumeReq{
		ConversationID: "conv-123",
	}

	assert.Equal(t, "conv-123", req.ConversationID)
}

func TestResumeReq_Empty(t *testing.T) {
	t.Parallel()

	req := ResumeReq{}

	assert.Empty(t, req.ConversationID)
}

func TestResumeReq_WithDifferentIDs(t *testing.T) {
	t.Parallel()

	ids := []string{
		"conv-001",
		"conv-xyz",
		"对话-123",
		"",
	}

	for _, id := range ids {
		req := ResumeReq{
			ConversationID: id,
		}
		assert.Equal(t, id, req.ConversationID)
	}
}

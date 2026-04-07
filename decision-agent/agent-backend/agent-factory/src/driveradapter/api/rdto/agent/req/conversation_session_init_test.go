package agentreq

import (
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
	"github.com/stretchr/testify/assert"
)

func TestConversationSessionInitReq_Check(t *testing.T) {
	t.Parallel()

	t.Run("with valid request", func(t *testing.T) {
		t.Parallel()

		req := &ConversationSessionInitReq{
			ConversationID: "conv-123",
			AgentID:        "agent-456",
			AgentVersion:   "v1.0.0",
		}

		err := req.Check()
		assert.NoError(t, err)
	})

	t.Run("with empty conversation_id", func(t *testing.T) {
		t.Parallel()

		req := &ConversationSessionInitReq{
			ConversationID: "",
		}

		err := req.Check()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "conversation_id is empty")
	})

	t.Run("with agent_id but empty agent_version", func(t *testing.T) {
		t.Parallel()

		req := &ConversationSessionInitReq{
			ConversationID: "conv-123",
			AgentID:        "agent-456",
			AgentVersion:   "",
		}

		err := req.Check()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "agent_version cannot be empty")
	})

	t.Run("without agent_id", func(t *testing.T) {
		t.Parallel()

		req := &ConversationSessionInitReq{
			ConversationID: "conv-123",
			AgentID:        "",
			AgentVersion:   "",
		}

		err := req.Check()
		assert.NoError(t, err)
	})

	t.Run("with only conversation_id", func(t *testing.T) {
		t.Parallel()

		req := &ConversationSessionInitReq{
			ConversationID: "conv-789",
		}

		err := req.Check()
		assert.NoError(t, err)
	})
}

func TestConversationSessionInitReq_StructFields(t *testing.T) {
	t.Parallel()

	req := &ConversationSessionInitReq{
		ConversationID:        "conv-123",
		ConversationSessionID: "session-456",
		AgentID:               "agent-789",
		AgentVersion:          "v1.0.0",
		UserID:                "user-101",
		XAccountID:            "account-202",
		XAccountType:          cenum.AccountTypeUser,
		XBusinessDomainID:     "domain-303",
	}

	assert.Equal(t, "conv-123", req.ConversationID)
	assert.Equal(t, "session-456", req.ConversationSessionID)
	assert.Equal(t, "agent-789", req.AgentID)
	assert.Equal(t, "v1.0.0", req.AgentVersion)
	assert.Equal(t, "user-101", req.UserID)
	assert.Equal(t, "account-202", req.XAccountID)
	assert.Equal(t, cenum.AccountTypeUser, req.XAccountType)
	assert.Equal(t, "domain-303", req.XBusinessDomainID)
}

func TestConversationSessionInitReq_Empty(t *testing.T) {
	t.Parallel()

	req := &ConversationSessionInitReq{}

	assert.Empty(t, req.ConversationID)
	assert.Empty(t, req.ConversationSessionID)
	assert.Empty(t, req.AgentID)
	assert.Empty(t, req.AgentVersion)
	assert.Empty(t, req.UserID)
	assert.Empty(t, req.XAccountID)
	assert.Empty(t, req.XBusinessDomainID)
}

func TestConversationSessionInitReq_WithSessionID(t *testing.T) {
	t.Parallel()

	sessionIDs := []string{
		"session-001",
		"session-abc-123",
		"test-session",
		"",
	}

	for _, sessionID := range sessionIDs {
		req := &ConversationSessionInitReq{
			ConversationID:        "conv-123",
			ConversationSessionID: sessionID,
		}
		assert.Equal(t, sessionID, req.ConversationSessionID)
	}
}

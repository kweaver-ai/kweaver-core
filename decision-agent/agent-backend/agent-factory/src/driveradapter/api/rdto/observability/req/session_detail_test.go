package observabilityreq

import (
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
	"github.com/stretchr/testify/assert"
)

func TestSessionDetailReq_StructFields(t *testing.T) {
	t.Parallel()

	accountType := cenum.AccountTypeApp
	req := SessionDetailReq{
		AgentID:        "agent-123",
		AgentVersion:   "v1.0.0",
		ConversationID: "conv-456",
		SessionID:      "session-789",
		StartTime:      1700000000000,
		EndTime:        1700003600000,
		XAccountID:     "user-123",
		XAccountType:   accountType,
	}

	assert.Equal(t, "agent-123", req.AgentID)
	assert.Equal(t, "v1.0.0", req.AgentVersion)
	assert.Equal(t, "conv-456", req.ConversationID)
	assert.Equal(t, "session-789", req.SessionID)
	assert.Equal(t, int64(1700000000000), req.StartTime)
	assert.Equal(t, int64(1700003600000), req.EndTime)
	assert.Equal(t, "user-123", req.XAccountID)
	assert.Equal(t, cenum.AccountTypeApp, req.XAccountType)
}

func TestSessionDetailReq_Empty(t *testing.T) {
	t.Parallel()

	req := SessionDetailReq{}

	assert.Empty(t, req.AgentID)
	assert.Empty(t, req.AgentVersion)
	assert.Empty(t, req.ConversationID)
	assert.Empty(t, req.SessionID)
	assert.Zero(t, req.StartTime)
	assert.Zero(t, req.EndTime)
	assert.Empty(t, req.XAccountID)
	assert.Zero(t, req.XAccountType)
}

func TestSessionDetailReq_WithAllFields(t *testing.T) {
	t.Parallel()

	accountType := cenum.AccountTypeAnonymous
	req := SessionDetailReq{
		AgentID:        "complete-agent",
		AgentVersion:   "v2.0.0",
		ConversationID: "complete-conv",
		SessionID:      "complete-session",
		StartTime:      1700000000000,
		EndTime:        1700010000000,
		XAccountID:     "complete-user",
		XAccountType:   accountType,
	}

	assert.Equal(t, "complete-agent", req.AgentID)
	assert.Equal(t, "v2.0.0", req.AgentVersion)
	assert.Equal(t, "complete-conv", req.ConversationID)
	assert.Equal(t, "complete-session", req.SessionID)
}

package observabilityreq

import (
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cenum"
	"github.com/stretchr/testify/assert"
)

func TestRunDetailReq_StructFields(t *testing.T) {
	t.Parallel()

	accountType := cenum.AccountTypeUser
	req := RunDetailReq{
		AgentID:        "agent-123",
		AgentVersion:   "v1.0.0",
		ConversationID: "conv-456",
		SessionID:      "session-789",
		RunID:          "run-001",
		StartTime:      1700000000000,
		EndTime:        1700000005000,
		XAccountID:     "user-123",
		XAccountType:   accountType,
	}

	assert.Equal(t, "agent-123", req.AgentID)
	assert.Equal(t, "v1.0.0", req.AgentVersion)
	assert.Equal(t, "conv-456", req.ConversationID)
	assert.Equal(t, "session-789", req.SessionID)
	assert.Equal(t, "run-001", req.RunID)
	assert.Equal(t, int64(1700000000000), req.StartTime)
	assert.Equal(t, int64(1700000005000), req.EndTime)
	assert.Equal(t, "user-123", req.XAccountID)
	assert.Equal(t, cenum.AccountTypeUser, req.XAccountType)
}

func TestRunDetailReq_Empty(t *testing.T) {
	t.Parallel()

	req := RunDetailReq{}

	assert.Empty(t, req.AgentID)
	assert.Empty(t, req.AgentVersion)
	assert.Empty(t, req.ConversationID)
	assert.Empty(t, req.SessionID)
	assert.Empty(t, req.RunID)
	assert.Zero(t, req.StartTime)
	assert.Zero(t, req.EndTime)
	assert.Empty(t, req.XAccountID)
	assert.Zero(t, req.XAccountType)
}

func TestRunDetailReq_WithAgentID(t *testing.T) {
	t.Parallel()

	agentIDs := []string{
		"agent-001",
		"agent-test",
		"智能体-123",
		"",
	}

	for _, agentID := range agentIDs {
		req := RunDetailReq{AgentID: agentID}
		assert.Equal(t, agentID, req.AgentID)
	}
}

func TestRunDetailReq_WithRunID(t *testing.T) {
	t.Parallel()

	runIDs := []string{
		"run-001",
		"run-test-123",
		"",
	}

	for _, runID := range runIDs {
		req := RunDetailReq{RunID: runID}
		assert.Equal(t, runID, req.RunID)
	}
}

func TestRunDetailReq_WithTimeRange(t *testing.T) {
	t.Parallel()

	req := RunDetailReq{
		StartTime: 1700000000000, // 2023-11-15
		EndTime:   1700003600000, // 2023-11-15 + 1 hour
	}

	assert.Equal(t, int64(1700000000000), req.StartTime)
	assert.Equal(t, int64(1700003600000), req.EndTime)
	assert.Greater(t, req.EndTime, req.StartTime)
}

func TestRunDetailReq_WithAllFields(t *testing.T) {
	t.Parallel()

	accountType := cenum.AccountTypeApp
	req := RunDetailReq{
		AgentID:        "complete-agent",
		AgentVersion:   "v2.5.0",
		ConversationID: "complete-conv",
		SessionID:      "complete-session",
		RunID:          "complete-run",
		StartTime:      1700000000000,
		EndTime:        1700100000000,
		XAccountID:     "complete-user",
		XAccountType:   accountType,
	}

	assert.Equal(t, "complete-agent", req.AgentID)
	assert.Equal(t, "v2.5.0", req.AgentVersion)
	assert.Equal(t, "complete-conv", req.ConversationID)
	assert.Equal(t, "complete-session", req.SessionID)
	assert.Equal(t, "complete-run", req.RunID)
}

func TestRunDetailReq_WithDifferentAccountTypes(t *testing.T) {
	t.Parallel()

	accountTypes := []cenum.AccountType{
		cenum.AccountTypeUser,
		cenum.AccountTypeApp,
		cenum.AccountTypeAnonymous,
	}

	for _, accountType := range accountTypes {
		req := RunDetailReq{XAccountType: accountType}
		assert.Equal(t, accountType, req.XAccountType)
	}
}

func TestRunDetailReq_WithZeroTime(t *testing.T) {
	t.Parallel()

	req := RunDetailReq{
		StartTime: 0,
		EndTime:   0,
	}

	assert.Zero(t, req.StartTime)
	assert.Zero(t, req.EndTime)
}

func TestRunDetailReq_WithNegativeTime(t *testing.T) {
	t.Parallel()

	req := RunDetailReq{
		StartTime: -1000,
		EndTime:   -500,
	}

	assert.Equal(t, int64(-1000), req.StartTime)
	assert.Equal(t, int64(-500), req.EndTime)
}

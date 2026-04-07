package observabilityreq

import (
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
	"github.com/stretchr/testify/assert"
)

func TestSessionListReq_StructFields(t *testing.T) {
	t.Parallel()

	req := &SessionListReq{
		AgentID:        "agent-123",
		AgentVersion:   "v1.0.0",
		ConversationID: "conv-456",
		StartTime:      1234567890,
		EndTime:        1234567999,
		Size:           20,
		Page:           1,
		XAccountID:     "account-789",
		XAccountType:   cenum.AccountTypeUser,
	}

	assert.Equal(t, "agent-123", req.AgentID)
	assert.Equal(t, "v1.0.0", req.AgentVersion)
	assert.Equal(t, "conv-456", req.ConversationID)
	assert.Equal(t, int64(1234567890), req.StartTime)
	assert.Equal(t, int64(1234567999), req.EndTime)
	assert.Equal(t, 20, req.Size)
	assert.Equal(t, 1, req.Page)
	assert.Equal(t, "account-789", req.XAccountID)
	assert.Equal(t, cenum.AccountTypeUser, req.XAccountType)
}

func TestSessionListReq_Empty(t *testing.T) {
	t.Parallel()

	req := &SessionListReq{}

	assert.Empty(t, req.AgentID)
	assert.Empty(t, req.AgentVersion)
	assert.Empty(t, req.ConversationID)
	assert.Equal(t, int64(0), req.StartTime)
	assert.Equal(t, int64(0), req.EndTime)
	assert.Equal(t, 0, req.Size)
	assert.Equal(t, 0, req.Page)
	assert.Empty(t, req.XAccountID)
}

func TestSessionListReq_WithPagination(t *testing.T) {
	t.Parallel()

	req := &SessionListReq{
		Size: 50,
		Page: 2,
	}

	assert.Equal(t, 50, req.Size)
	assert.Equal(t, 2, req.Page)
}

func TestSessionListReq_WithTimeRange(t *testing.T) {
	t.Parallel()

	req := &SessionListReq{
		StartTime: 1000000000,
		EndTime:   2000000000,
	}

	assert.Equal(t, int64(1000000000), req.StartTime)
	assert.Equal(t, int64(2000000000), req.EndTime)
}

func TestSessionListReq_WithConversationID(t *testing.T) {
	t.Parallel()

	convIDs := []string{
		"conv-001",
		"conv-abc-123",
		"test-conversation",
		"",
	}

	for _, convID := range convIDs {
		req := &SessionListReq{ConversationID: convID}
		assert.Equal(t, convID, req.ConversationID)
	}
}

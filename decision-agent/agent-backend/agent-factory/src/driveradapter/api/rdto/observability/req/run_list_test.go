package observabilityreq

import (
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cenum"
	"github.com/stretchr/testify/assert"
)

func TestRunListReq_StructFields(t *testing.T) {
	t.Parallel()

	accountType := cenum.AccountTypeUser
	req := RunListReq{
		AgentID:        "agent-123",
		AgentVersion:   "v1.0.0",
		ConversationID: "conv-456",
		SessionID:      "session-789",
		StartTime:      1700000000000,
		EndTime:        1700000005000,
		Page:           1,
		Size:           10,
		XAccountID:     "user-123",
		XAccountType:   accountType,
	}

	assert.Equal(t, "agent-123", req.AgentID)
	assert.Equal(t, "v1.0.0", req.AgentVersion)
	assert.Equal(t, "conv-456", req.ConversationID)
	assert.Equal(t, "session-789", req.SessionID)
	assert.Equal(t, int64(1700000000000), req.StartTime)
	assert.Equal(t, int64(1700000005000), req.EndTime)
	assert.Equal(t, 1, req.Page)
	assert.Equal(t, 10, req.Size)
	assert.Equal(t, "user-123", req.XAccountID)
	assert.Equal(t, cenum.AccountTypeUser, req.XAccountType)
}

func TestRunListReq_Empty(t *testing.T) {
	t.Parallel()

	req := RunListReq{}

	assert.Empty(t, req.AgentID)
	assert.Empty(t, req.AgentVersion)
	assert.Empty(t, req.ConversationID)
	assert.Empty(t, req.SessionID)
	assert.Zero(t, req.StartTime)
	assert.Zero(t, req.EndTime)
	assert.Zero(t, req.Page)
	assert.Zero(t, req.Size)
	assert.Empty(t, req.XAccountID)
	assert.Zero(t, req.XAccountType)
}

func TestRunListReq_WithPagination(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		page int
		size int
	}{
		{"page 1 size 10", 1, 10},
		{"page 2 size 20", 2, 20},
		{"page 5 size 50", 5, 50},
		{"page 1 size 100", 1, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := RunListReq{
				Page: tt.page,
				Size: tt.size,
			}
			assert.Equal(t, tt.page, req.Page)
			assert.Equal(t, tt.size, req.Size)
		})
	}
}

func TestRunListReq_WithZeroPagination(t *testing.T) {
	t.Parallel()

	req := RunListReq{
		Page: 0,
		Size: 0,
	}

	assert.Zero(t, req.Page)
	assert.Zero(t, req.Size)
}

func TestRunListReq_WithLargePageNumber(t *testing.T) {
	t.Parallel()

	req := RunListReq{
		Page: 1000,
		Size: 100,
	}

	assert.Equal(t, 1000, req.Page)
	assert.Equal(t, 100, req.Size)
}

func TestRunListReq_WithNegativePage(t *testing.T) {
	t.Parallel()

	req := RunListReq{
		Page: -1,
		Size: 10,
	}

	assert.Equal(t, -1, req.Page)
	assert.Equal(t, 10, req.Size)
}

func TestRunListReq_WithAllFields(t *testing.T) {
	t.Parallel()

	accountType := cenum.AccountTypeAnonymous
	req := RunListReq{
		AgentID:        "complete-agent",
		AgentVersion:   "v2.5.0",
		ConversationID: "complete-conv",
		SessionID:      "complete-session",
		StartTime:      1700000000000,
		EndTime:        1700010000000,
		Page:           3,
		Size:           25,
		XAccountID:     "complete-user",
		XAccountType:   accountType,
	}

	assert.Equal(t, "complete-agent", req.AgentID)
	assert.Equal(t, 3, req.Page)
	assert.Equal(t, 25, req.Size)
	assert.Equal(t, cenum.AccountTypeAnonymous, req.XAccountType)
}

func TestRunListReq_DefaultPageAndSize(t *testing.T) {
	t.Parallel()

	req := RunListReq{
		AgentID: "test-agent",
	}

	// Default values should be 0 for Page and Size
	assert.Zero(t, req.Page)
	assert.Zero(t, req.Size)
}

func TestRunListReq_WithTimeRange(t *testing.T) {
	t.Parallel()

	req := RunListReq{
		StartTime: 1609459200000, // 2021-01-01
		EndTime:   1609545600000, // 2021-01-02
		Page:      1,
		Size:      10,
	}

	assert.Equal(t, int64(1609459200000), req.StartTime)
	assert.Equal(t, int64(1609545600000), req.EndTime)
	assert.Greater(t, req.EndTime, req.StartTime)
}

package observabilityreq

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConversationListReq_StructFields(t *testing.T) {
	t.Parallel()

	req := &ConversationListReq{
		AgentID:      "agent-123",
		AgentVersion: "v1.0.0",
		Title:        "Test Conversation",
		Size:         20,
		Page:         1,
		StartTime:    1234567890,
		EndTime:      1234567999,
	}

	assert.Equal(t, "agent-123", req.AgentID)
	assert.Equal(t, "v1.0.0", req.AgentVersion)
	assert.Equal(t, "Test Conversation", req.Title)
	assert.Equal(t, 20, req.Size)
	assert.Equal(t, 1, req.Page)
	assert.Equal(t, int64(1234567890), req.StartTime)
	assert.Equal(t, int64(1234567999), req.EndTime)
}

func TestConversationListReq_Empty(t *testing.T) {
	t.Parallel()

	req := &ConversationListReq{}

	assert.Empty(t, req.AgentID)
	assert.Empty(t, req.AgentVersion)
	assert.Empty(t, req.Title)
	assert.Equal(t, 0, req.Size)
	assert.Equal(t, 0, req.Page)
	assert.Equal(t, int64(0), req.StartTime)
	assert.Equal(t, int64(0), req.EndTime)
}

func TestConversationListReq_WithPagination(t *testing.T) {
	t.Parallel()

	req := &ConversationListReq{
		Size: 50,
		Page: 2,
	}

	assert.Equal(t, 50, req.Size)
	assert.Equal(t, 2, req.Page)
}

func TestConversationListReq_WithTimeRange(t *testing.T) {
	t.Parallel()

	req := &ConversationListReq{
		StartTime: 1000000000,
		EndTime:   2000000000,
	}

	assert.Equal(t, int64(1000000000), req.StartTime)
	assert.Equal(t, int64(2000000000), req.EndTime)
}

func TestConversationListReq_WithDifferentPageSizes(t *testing.T) {
	t.Parallel()

	pageSizes := []int{0, 10, 20, 50, 100}

	for _, size := range pageSizes {
		req := &ConversationListReq{Size: size}
		assert.Equal(t, size, req.Size)
	}
}

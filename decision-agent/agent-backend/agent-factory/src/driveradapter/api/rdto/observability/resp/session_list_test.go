package observabilityresp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSessionListItem_StructFields(t *testing.T) {
	t.Parallel()

	item := SessionListItem{
		SessionID:       "session-123",
		StartTime:       1234567890,
		EndTime:         1234567999,
		SessionDuration: 109,
	}

	assert.Equal(t, "session-123", item.SessionID)
	assert.Equal(t, int64(1234567890), item.StartTime)
	assert.Equal(t, int64(1234567999), item.EndTime)
	assert.Equal(t, 109, item.SessionDuration)
}

func TestSessionListItem_Empty(t *testing.T) {
	t.Parallel()

	item := SessionListItem{}

	assert.Empty(t, item.SessionID)
	assert.Equal(t, int64(0), item.StartTime)
	assert.Equal(t, int64(0), item.EndTime)
	assert.Equal(t, 0, item.SessionDuration)
}

func TestSessionListItem_WithSessionID(t *testing.T) {
	t.Parallel()

	sessionIDs := []string{
		"session-001",
		"test-session",
		"SESSION-ABC-123",
		"",
	}

	for _, sessionID := range sessionIDs {
		item := SessionListItem{SessionID: sessionID}
		assert.Equal(t, sessionID, item.SessionID)
	}
}

func TestSessionListItem_WithTimeRange(t *testing.T) {
	t.Parallel()

	item := SessionListItem{
		StartTime: 1000000000,
		EndTime:   2000000000,
	}

	assert.Equal(t, int64(1000000000), item.StartTime)
	assert.Equal(t, int64(2000000000), item.EndTime)
}

func TestSessionListItem_WithDuration(t *testing.T) {
	t.Parallel()

	durations := []int{0, 100, 1000, 5000, 10000}

	for _, duration := range durations {
		item := SessionListItem{
			SessionDuration: duration,
		}
		assert.Equal(t, duration, item.SessionDuration)
	}
}

func TestSessionListResp_StructFields(t *testing.T) {
	t.Parallel()

	entries := []SessionListItem{
		{SessionID: "session-1"},
		{SessionID: "session-2"},
	}

	resp := SessionListResp{
		Entries:    entries,
		TotalCount: 10,
	}

	assert.Len(t, resp.Entries, 2)
	assert.Equal(t, 10, resp.TotalCount)
	assert.Equal(t, "session-1", resp.Entries[0].SessionID)
	assert.Equal(t, "session-2", resp.Entries[1].SessionID)
}

func TestSessionListResp_Empty(t *testing.T) {
	t.Parallel()

	resp := SessionListResp{}

	assert.Nil(t, resp.Entries)
	assert.Equal(t, 0, resp.TotalCount)
}

func TestSessionListResp_WithEmptyEntries(t *testing.T) {
	t.Parallel()

	resp := SessionListResp{
		Entries:    []SessionListItem{},
		TotalCount: 0,
	}

	assert.NotNil(t, resp.Entries)
	assert.Len(t, resp.Entries, 0)
	assert.Equal(t, 0, resp.TotalCount)
}

func TestSessionListResp_WithSingleEntry(t *testing.T) {
	t.Parallel()

	entries := []SessionListItem{
		{
			SessionID: "session-123",
			StartTime: 1234567890,
			EndTime:   1234567999,
		},
	}

	resp := SessionListResp{
		Entries:    entries,
		TotalCount: 1,
	}

	assert.Len(t, resp.Entries, 1)
	assert.Equal(t, 1, resp.TotalCount)
	assert.Equal(t, "session-123", resp.Entries[0].SessionID)
}

func TestSessionListResp_WithTotalCount(t *testing.T) {
	t.Parallel()

	totalCounts := []int{0, 1, 10, 100, 1000}

	for _, totalCount := range totalCounts {
		resp := SessionListResp{
			TotalCount: totalCount,
		}
		assert.Equal(t, totalCount, resp.TotalCount)
	}
}

func TestSessionListResp_WithMultipleEntries(t *testing.T) {
	t.Parallel()

	entries := []SessionListItem{
		{SessionID: "session-1", StartTime: 1000, EndTime: 2000},
		{SessionID: "session-2", StartTime: 3000, EndTime: 4000},
		{SessionID: "session-3", StartTime: 5000, EndTime: 6000},
	}

	resp := SessionListResp{
		Entries:    entries,
		TotalCount: 3,
	}

	assert.Len(t, resp.Entries, 3)
	assert.Equal(t, 3, resp.TotalCount)
	assert.Equal(t, int64(1000), resp.Entries[0].StartTime)
	assert.Equal(t, int64(6000), resp.Entries[2].EndTime)
}

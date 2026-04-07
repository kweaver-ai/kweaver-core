package observabilityresp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunDetail_StructFields(t *testing.T) {
	t.Parallel()

	detail := RunDetail{
		RunID:               "run-123",
		AgentID:             "agent-456",
		AgentVersion:        "v1.0.0",
		ConversationID:      "conv-789",
		SessionID:           "session-001",
		UserID:              "user-123",
		CallType:            "chat",
		StartTime:           1700000000000,
		EndTime:             1700000005000,
		TTFT:                100,
		TotalTime:           5000,
		TotalTokens:         1000,
		InputMessage:        "Hello, AI!",
		ToolCallCount:       2,
		ToolCallFailedCount: 0,
		Status:              "Success",
	}

	assert.Equal(t, "run-123", detail.RunID)
	assert.Equal(t, "agent-456", detail.AgentID)
	assert.Equal(t, "v1.0.0", detail.AgentVersion)
	assert.Equal(t, "conv-789", detail.ConversationID)
	assert.Equal(t, "session-001", detail.SessionID)
	assert.Equal(t, "user-123", detail.UserID)
	assert.Equal(t, "chat", detail.CallType)
	assert.Equal(t, int64(1700000000000), detail.StartTime)
	assert.Equal(t, int64(1700000005000), detail.EndTime)
	assert.Equal(t, 100, detail.TTFT)
	assert.Equal(t, 5000, detail.TotalTime)
	assert.Equal(t, int64(1000), detail.TotalTokens)
	assert.Equal(t, "Hello, AI!", detail.InputMessage)
	assert.Equal(t, 2, detail.ToolCallCount)
	assert.Equal(t, 0, detail.ToolCallFailedCount)
	assert.Equal(t, "Success", detail.Status)
}

func TestRunDetail_Empty(t *testing.T) {
	t.Parallel()

	detail := RunDetail{}

	assert.Empty(t, detail.RunID)
	assert.Empty(t, detail.AgentID)
	assert.Empty(t, detail.AgentVersion)
	assert.Empty(t, detail.ConversationID)
	assert.Empty(t, detail.SessionID)
	assert.Empty(t, detail.UserID)
	assert.Empty(t, detail.CallType)
	assert.Zero(t, detail.StartTime)
	assert.Zero(t, detail.EndTime)
	assert.Zero(t, detail.TTFT)
	assert.Zero(t, detail.TotalTime)
	assert.Zero(t, detail.TotalTokens)
	assert.Empty(t, detail.InputMessage)
	assert.Zero(t, detail.ToolCallCount)
	assert.Zero(t, detail.ToolCallFailedCount)
	assert.Empty(t, detail.Status)
}

func TestRunListItem_StructFields(t *testing.T) {
	t.Parallel()

	item := RunListItem{
		AgentID:      "agent-123",
		RunID:        "run-456",
		InputMessage: "Test input",
		Status:       "Success",
		TotalTime:    5000,
	}

	assert.Equal(t, "agent-123", item.AgentID)
	assert.Equal(t, "run-456", item.RunID)
	assert.Equal(t, "Test input", item.InputMessage)
	assert.Equal(t, "Success", item.Status)
	assert.Equal(t, 5000, item.TotalTime)
}

func TestRunListItem_Empty(t *testing.T) {
	t.Parallel()

	item := RunListItem{}

	assert.Empty(t, item.AgentID)
	assert.Empty(t, item.RunID)
	assert.Empty(t, item.InputMessage)
	assert.Empty(t, item.Status)
	assert.Zero(t, item.TotalTime)
}

func TestRunListResp_StructFields(t *testing.T) {
	t.Parallel()

	resp := RunListResp{
		Entries: []RunListItem{
			{
				AgentID:      "agent-1",
				RunID:        "run-1",
				InputMessage: "Input 1",
				Status:       "Success",
				TotalTime:    1000,
			},
			{
				AgentID:      "agent-2",
				RunID:        "run-2",
				InputMessage: "Input 2",
				Status:       "Failed",
				TotalTime:    2000,
			},
		},
		TotalCount: 100,
	}

	assert.Len(t, resp.Entries, 2)
	assert.Equal(t, 100, resp.TotalCount)
	assert.Equal(t, "agent-1", resp.Entries[0].AgentID)
	assert.Equal(t, "run-1", resp.Entries[0].RunID)
	assert.Equal(t, "Success", resp.Entries[0].Status)
	assert.Equal(t, "agent-2", resp.Entries[1].AgentID)
	assert.Equal(t, "run-2", resp.Entries[1].RunID)
	assert.Equal(t, "Failed", resp.Entries[1].Status)
}

func TestRunListResp_Empty(t *testing.T) {
	t.Parallel()

	resp := RunListResp{}

	assert.Nil(t, resp.Entries)
	assert.Zero(t, resp.TotalCount)
}

func TestRunListResp_WithEmptyEntries(t *testing.T) {
	t.Parallel()

	resp := RunListResp{
		Entries:    []RunListItem{},
		TotalCount: 0,
	}

	assert.NotNil(t, resp.Entries)
	assert.Len(t, resp.Entries, 0)
	assert.Zero(t, resp.TotalCount)
}

func TestRunListItem_WithDifferentStatuses(t *testing.T) {
	t.Parallel()

	statuses := []string{
		"Success",
		"Failed",
		"Running",
		"",
	}

	for _, status := range statuses {
		item := RunListItem{Status: status}
		assert.Equal(t, status, item.Status)
	}
}

func TestRunListItem_WithTotalTimes(t *testing.T) {
	t.Parallel()

	times := []int{
		0,
		100,
		1000,
		5000,
		10000,
	}

	for _, totalTime := range times {
		item := RunListItem{TotalTime: totalTime}
		assert.Equal(t, totalTime, item.TotalTime)
	}
}

func TestRunDetail_WithFailedStatus(t *testing.T) {
	t.Parallel()

	detail := RunDetail{
		RunID:         "failed-run",
		Status:        "Failed",
		TotalTime:     10000,
		ToolCallCount: 5,
	}

	assert.Equal(t, "failed-run", detail.RunID)
	assert.Equal(t, "Failed", detail.Status)
	assert.Equal(t, 10000, detail.TotalTime)
	assert.Equal(t, 5, detail.ToolCallCount)
}

func TestRunListResp_WithSingleEntry(t *testing.T) {
	t.Parallel()

	resp := RunListResp{
		Entries: []RunListItem{
			{
				AgentID:      "single-agent",
				RunID:        "single-run",
				InputMessage: "Single input",
				Status:       "Success",
				TotalTime:    1000,
			},
		},
		TotalCount: 1,
	}

	assert.Len(t, resp.Entries, 1)
	assert.Equal(t, 1, resp.TotalCount)
	assert.Equal(t, "single-agent", resp.Entries[0].AgentID)
	assert.Equal(t, "single-run", resp.Entries[0].RunID)
}

func TestRunListResp_WithLargeTotalCount(t *testing.T) {
	t.Parallel()

	resp := RunListResp{
		Entries:    []RunListItem{},
		TotalCount: 1000000,
	}

	assert.NotNil(t, resp.Entries)
	assert.Equal(t, 1000000, resp.TotalCount)
}

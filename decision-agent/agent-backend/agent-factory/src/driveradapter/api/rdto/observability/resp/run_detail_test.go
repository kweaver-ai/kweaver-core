package observabilityresp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunDetailResp_StructFields(t *testing.T) {
	t.Parallel()

	resp := RunDetailResp{
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

	assert.Equal(t, "run-123", resp.RunID)
	assert.Equal(t, "agent-456", resp.AgentID)
	assert.Equal(t, "v1.0.0", resp.AgentVersion)
	assert.Equal(t, "conv-789", resp.ConversationID)
	assert.Equal(t, "session-001", resp.SessionID)
	assert.Equal(t, "user-123", resp.UserID)
	assert.Equal(t, "chat", resp.CallType)
	assert.Equal(t, int64(1700000000000), resp.StartTime)
	assert.Equal(t, int64(1700000005000), resp.EndTime)
	assert.Equal(t, 100, resp.TTFT)
	assert.Equal(t, int64(5000), resp.TotalTime)
	assert.Equal(t, int64(1000), resp.TotalTokens)
	assert.Equal(t, "Hello, AI!", resp.InputMessage)
	assert.Equal(t, 2, resp.ToolCallCount)
	assert.Equal(t, 0, resp.ToolCallFailedCount)
	assert.Equal(t, "Success", resp.Status)
}

func TestRunDetailResp_Empty(t *testing.T) {
	t.Parallel()

	resp := RunDetailResp{}

	assert.Empty(t, resp.RunID)
	assert.Empty(t, resp.AgentID)
	assert.Empty(t, resp.AgentVersion)
	assert.Empty(t, resp.ConversationID)
	assert.Empty(t, resp.SessionID)
	assert.Empty(t, resp.UserID)
	assert.Empty(t, resp.CallType)
	assert.Zero(t, resp.StartTime)
	assert.Zero(t, resp.EndTime)
	assert.Zero(t, resp.TTFT)
	assert.Zero(t, resp.TotalTime)
	assert.Zero(t, resp.TotalTokens)
	assert.Empty(t, resp.InputMessage)
	assert.Zero(t, resp.ToolCallCount)
	assert.Zero(t, resp.ToolCallFailedCount)
	assert.Empty(t, resp.Status)
}

func TestRunDetailResp_WithDifferentCallTypes(t *testing.T) {
	t.Parallel()

	callTypes := []string{
		"chat",
		"debugchat",
		"apichat",
	}

	for _, callType := range callTypes {
		resp := RunDetailResp{CallType: callType}
		assert.Equal(t, callType, resp.CallType)
	}
}

func TestRunDetailResp_WithStatus(t *testing.T) {
	t.Parallel()

	statuses := []string{
		"Success",
		"Failed",
		"Running",
		"",
	}

	for _, status := range statuses {
		resp := RunDetailResp{Status: status}
		assert.Equal(t, status, resp.Status)
	}
}

func TestRunDetailResp_WithFailedToolCalls(t *testing.T) {
	t.Parallel()

	resp := RunDetailResp{
		ToolCallCount:       5,
		ToolCallFailedCount: 2,
	}

	assert.Equal(t, 5, resp.ToolCallCount)
	assert.Equal(t, 2, resp.ToolCallFailedCount)
}

func TestRunDetailResp_WithAllFields(t *testing.T) {
	t.Parallel()

	resp := RunDetailResp{
		RunID:               "complete-run",
		AgentID:             "complete-agent",
		AgentVersion:        "v2.5.0",
		ConversationID:      "complete-conv",
		SessionID:           "complete-session",
		UserID:              "complete-user",
		CallType:            "apichat",
		StartTime:           1700000000000,
		EndTime:             1700000100000,
		TTFT:                200,
		TotalTime:           100000,
		TotalTokens:         5000,
		InputMessage:        "Complete message",
		ToolCallCount:       10,
		ToolCallFailedCount: 1,
		Status:              "Success",
	}

	assert.Equal(t, "complete-run", resp.RunID)
	assert.Equal(t, "complete-agent", resp.AgentID)
	assert.Equal(t, "v2.5.0", resp.AgentVersion)
	assert.Equal(t, 10, resp.ToolCallCount)
	assert.Equal(t, 1, resp.ToolCallFailedCount)
}

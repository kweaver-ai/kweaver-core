package v2agentexecutordto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestV2AgentCallResp_StructFields(t *testing.T) {
	t.Parallel()

	interruptHandle := &InterruptHandle{
		FrameID:       "frame-123",
		SnapshotID:    "snapshot-456",
		ResumeToken:   "token-789",
		InterruptType: "tool_call",
		CurrentBlock:  1,
		RestartBlock:  false,
	}

	interruptData := &InterruptData{
		ToolName:        "test_tool",
		ToolDescription: "Test tool description",
	}

	interruptInfo := &ToolInterruptInfo{
		Handle: interruptHandle,
		Data:   interruptData,
	}

	resp := &V2AgentCallResp{
		Answer:        "test answer",
		Status:        "completed",
		AgentRunID:    "run-123",
		InterruptInfo: interruptInfo,
	}

	assert.Equal(t, "test answer", resp.Answer)
	assert.Equal(t, "completed", resp.Status)
	assert.Equal(t, "run-123", resp.AgentRunID)
	assert.NotNil(t, resp.InterruptInfo)
	assert.NotNil(t, resp.InterruptInfo.Handle)
	assert.NotNil(t, resp.InterruptInfo.Data)
	assert.Equal(t, "test_tool", resp.InterruptInfo.Data.ToolName)
}

func TestV2AgentCallResp_Empty(t *testing.T) {
	t.Parallel()

	resp := &V2AgentCallResp{}

	assert.Nil(t, resp.Answer)
	assert.Empty(t, resp.Status)
	assert.Empty(t, resp.AgentRunID)
	assert.Nil(t, resp.InterruptInfo)
}

func TestV2AgentCallResp_WithMapAnswer(t *testing.T) {
	t.Parallel()

	answer := map[string]interface{}{
		"response": "test response",
		"tokens":   100,
	}

	resp := &V2AgentCallResp{
		Answer: answer,
		Status: "success",
	}

	assert.NotNil(t, resp.Answer)
	assert.Equal(t, "success", resp.Status)
}

func TestV2AgentDebugResp_StructFields(t *testing.T) {
	t.Parallel()

	resp := &V2AgentDebugResp{
		Answer:     "debug answer",
		Status:     "debugging",
		AgentRunID: "debug-run-456",
	}

	assert.Equal(t, "debug answer", resp.Answer)
	assert.Equal(t, "debugging", resp.Status)
	assert.Equal(t, "debug-run-456", resp.AgentRunID)
}

func TestV2AgentDebugResp_WithInterruptInfo(t *testing.T) {
	t.Parallel()

	interruptHandle := &InterruptHandle{
		FrameID:    "frame-abc",
		SnapshotID: "snapshot-def",
	}

	interruptData := &InterruptData{
		ToolName:        "interrupt_tool",
		ToolDescription: "Tool needs confirmation",
		ToolArgs: []ToolArg{
			{Key: "arg1", Value: "val1", Type: "string"},
		},
	}

	interruptInfo := &ToolInterruptInfo{
		Handle: interruptHandle,
		Data:   interruptData,
	}

	resp := &V2AgentDebugResp{
		Answer:        nil,
		Status:        "interrupted",
		InterruptInfo: interruptInfo,
	}

	assert.Nil(t, resp.Answer)
	assert.Equal(t, "interrupted", resp.Status)
	assert.NotNil(t, resp.InterruptInfo)
	assert.NotNil(t, resp.InterruptInfo.Handle)
	assert.NotNil(t, resp.InterruptInfo.Data)
	assert.Equal(t, "interrupt_tool", resp.InterruptInfo.Data.ToolName)
}

func TestV2AgentCallResp_WithNilInterruptInfo(t *testing.T) {
	t.Parallel()

	resp := &V2AgentCallResp{
		Answer:        "answer without interrupt",
		Status:        "completed",
		AgentRunID:    "run-789",
		InterruptInfo: nil,
	}

	assert.Equal(t, "answer without interrupt", resp.Answer)
	assert.Nil(t, resp.InterruptInfo)
}

func TestToolInterruptInfo_WithOnlyHandle(t *testing.T) {
	t.Parallel()

	interruptHandle := &InterruptHandle{
		FrameID:     "frame-only",
		SnapshotID:  "snapshot-only",
		ResumeToken: "token-only",
	}

	interruptInfo := &ToolInterruptInfo{
		Handle: interruptHandle,
		Data:   nil,
	}

	assert.NotNil(t, interruptInfo.Handle)
	assert.Nil(t, interruptInfo.Data)
	assert.Equal(t, "frame-only", interruptInfo.Handle.FrameID)
}

func TestToolInterruptInfo_WithOnlyData(t *testing.T) {
	t.Parallel()

	interruptData := &InterruptData{
		ToolName:        "data_only_tool",
		ToolDescription: "Description",
		ToolArgs:        []ToolArg{},
	}

	interruptInfo := &ToolInterruptInfo{
		Handle: nil,
		Data:   interruptData,
	}

	assert.Nil(t, interruptInfo.Handle)
	assert.NotNil(t, interruptInfo.Data)
	assert.Equal(t, "data_only_tool", interruptInfo.Data.ToolName)
}

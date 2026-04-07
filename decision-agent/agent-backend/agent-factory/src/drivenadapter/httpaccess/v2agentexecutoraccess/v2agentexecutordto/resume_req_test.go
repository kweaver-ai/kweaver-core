package v2agentexecutordto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAgentResumeReq_StructFields(t *testing.T) {
	t.Parallel()

	resumeHandle := &InterruptHandle{
		FrameID:       "frame-123",
		SnapshotID:    "snapshot-456",
		ResumeToken:   "token-789",
		InterruptType: "tool_call",
		CurrentBlock:  1,
		RestartBlock:  false,
	}

	resumeInfo := &AgentResumeInfo{
		ResumeHandle: resumeHandle,
		Action:       "confirm",
		ModifiedArgs: []ModifiedArg{{Key: "arg1", Value: "val1"}},
		Data:         &InterruptData{},
	}

	req := &AgentResumeReq{
		AgentRunID: "run-123",
		ResumeInfo: resumeInfo,
	}

	assert.Equal(t, "run-123", req.AgentRunID)
	assert.NotNil(t, req.ResumeInfo)
	assert.Equal(t, "confirm", req.ResumeInfo.Action)
}

func TestAgentResumeReq_Empty(t *testing.T) {
	t.Parallel()

	req := &AgentResumeReq{}

	assert.Empty(t, req.AgentRunID)
	assert.Nil(t, req.ResumeInfo)
}

func TestAgentResumeInfo_StructFields(t *testing.T) {
	t.Parallel()

	interruptHandle := &InterruptHandle{
		FrameID: "frame-abc",
	}

	info := &AgentResumeInfo{
		ResumeHandle: interruptHandle,
		Action:       "skip",
		ModifiedArgs: []ModifiedArg{
			{Key: "param1", Value: "value1"},
			{Key: "param2", Value: 123},
		},
	}

	assert.Equal(t, "skip", info.Action)
	assert.NotNil(t, info.ResumeHandle)
	assert.Len(t, info.ModifiedArgs, 2)
	assert.Equal(t, "param1", info.ModifiedArgs[0].Key)
	assert.Equal(t, "value1", info.ModifiedArgs[0].Value)
}

func TestAgentResumeInfo_WithActionConfirm(t *testing.T) {
	t.Parallel()

	info := &AgentResumeInfo{
		Action: "confirm",
	}

	assert.Equal(t, "confirm", info.Action)
}

func TestAgentResumeInfo_WithActionSkip(t *testing.T) {
	t.Parallel()

	info := &AgentResumeInfo{
		Action: "skip",
	}

	assert.Equal(t, "skip", info.Action)
}

func TestModifiedArg_StructFields(t *testing.T) {
	t.Parallel()

	arg := ModifiedArg{
		Key:   "test_key",
		Value: "test_value",
	}

	assert.Equal(t, "test_key", arg.Key)
	assert.Equal(t, "test_value", arg.Value)
}

func TestModifiedArg_WithNumberValue(t *testing.T) {
	t.Parallel()

	arg := ModifiedArg{
		Key:   "count",
		Value: 42,
	}

	assert.Equal(t, "count", arg.Key)
	assert.Equal(t, 42, arg.Value)
}

func TestModifiedArg_WithObjectValue(t *testing.T) {
	t.Parallel()

	obj := map[string]interface{}{
		"nested": "value",
		"number": 123,
	}

	arg := ModifiedArg{
		Key:   "config",
		Value: obj,
	}

	assert.Equal(t, "config", arg.Key)
	assert.NotNil(t, arg.Value)
}

func TestAgentResumeReq_WithNilResumeInfo(t *testing.T) {
	t.Parallel()

	req := &AgentResumeReq{
		AgentRunID: "run-456",
		ResumeInfo: nil,
	}

	assert.Equal(t, "run-456", req.AgentRunID)
	assert.Nil(t, req.ResumeInfo)
}

func TestAgentResumeInfo_EmptyModifiedArgs(t *testing.T) {
	t.Parallel()

	info := &AgentResumeInfo{
		Action:       "confirm",
		ModifiedArgs: []ModifiedArg{},
	}

	assert.Equal(t, "confirm", info.Action)
	assert.NotNil(t, info.ModifiedArgs)
	assert.Len(t, info.ModifiedArgs, 0)
}

func TestAgentResumeInfo_WithNilData(t *testing.T) {
	t.Parallel()

	info := &AgentResumeInfo{
		Action: "skip",
		Data:   nil,
	}

	assert.Equal(t, "skip", info.Action)
	assert.Nil(t, info.Data)
}

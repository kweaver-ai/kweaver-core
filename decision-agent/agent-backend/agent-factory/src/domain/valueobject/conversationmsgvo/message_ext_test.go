package conversationmsgvo

import (
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/agentresperr"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/httpaccess/v2agentexecutoraccess/v2agentexecutordto"
	"github.com/stretchr/testify/assert"
)

func TestMessageExt_StructFields(t *testing.T) {
	t.Parallel()

	interruptInfo := &v2agentexecutordto.ToolInterruptInfo{
		Handle: &v2agentexecutordto.InterruptHandle{
			FrameID: "frame-123",
		},
		Data: &v2agentexecutordto.InterruptData{
			ToolName: "test_tool",
		},
	}

	ext := &MessageExt{
		InterruptInfo:  interruptInfo,
		RelatedQueries: []string{"query1", "query2"},
		TotalTime:      1.5,
		TotalTokens:    100,
		TTFT:           500,
		AgentRunID:     "run-456",
	}

	assert.NotNil(t, ext.InterruptInfo)
	assert.Len(t, ext.RelatedQueries, 2)
	assert.Equal(t, 1.5, ext.TotalTime)
	assert.Equal(t, int64(100), ext.TotalTokens)
	assert.Equal(t, int64(500), ext.TTFT)
	assert.Equal(t, "run-456", ext.AgentRunID)
}

func TestMessageExt_IsInterrupted_WithInterruptInfo(t *testing.T) {
	t.Parallel()

	interruptInfo := &v2agentexecutordto.ToolInterruptInfo{
		Handle: &v2agentexecutordto.InterruptHandle{
			FrameID: "frame-abc",
		},
	}

	ext := &MessageExt{
		InterruptInfo: interruptInfo,
	}

	assert.True(t, ext.IsInterrupted())
}

func TestMessageExt_IsInterrupted_WithoutInterruptInfo(t *testing.T) {
	t.Parallel()

	ext := &MessageExt{}

	assert.False(t, ext.IsInterrupted())
}

func TestMessageExt_WithError(t *testing.T) {
	t.Parallel()

	respErr := &agentresperr.RespError{
		Type:  agentresperr.RespErrorTypeAgentFactory,
		Error: "Test error message",
	}

	ext := &MessageExt{
		Error: respErr,
	}

	assert.NotNil(t, ext.Error)
	assert.Equal(t, agentresperr.RespErrorTypeAgentFactory, ext.Error.Type)
}

func TestMessageExt_Empty(t *testing.T) {
	t.Parallel()

	ext := &MessageExt{}

	assert.Nil(t, ext.InterruptInfo)
	assert.Nil(t, ext.RelatedQueries)
	assert.Zero(t, ext.TotalTime)
	assert.Zero(t, ext.TotalTokens)
	assert.Zero(t, ext.TTFT)
	assert.Empty(t, ext.AgentRunID)
	assert.Nil(t, ext.Error)
}

func TestMessageExt_WithEmptyRelatedQueries(t *testing.T) {
	t.Parallel()

	ext := &MessageExt{
		RelatedQueries: []string{},
	}

	assert.NotNil(t, ext.RelatedQueries)
	assert.Len(t, ext.RelatedQueries, 0)
}

func TestMessageExt_AllFields(t *testing.T) {
	t.Parallel()

	interruptInfo := &v2agentexecutordto.ToolInterruptInfo{
		Handle: &v2agentexecutordto.InterruptHandle{
			FrameID:       "frame-xyz",
			SnapshotID:    "snapshot-xyz",
			ResumeToken:   "token-xyz",
			InterruptType: "tool_call",
			CurrentBlock:  1,
			RestartBlock:  false,
		},
		Data: &v2agentexecutordto.InterruptData{
			ToolName:        "confirmation_tool",
			ToolDescription: "Needs user confirmation",
			InterruptConfig: &v2agentexecutordto.InterruptConfig{
				RequiresConfirmation: true,
				ConfirmationMessage:  "Please confirm",
			},
		},
	}

	respErr := &agentresperr.RespError{
		Type:  agentresperr.RespErrorTypeAgentExecutor,
		Error: map[string]string{"code": "ERR_TIMEOUT", "message": "Request timeout"},
	}

	ext := &MessageExt{
		InterruptInfo:  interruptInfo,
		RelatedQueries: []string{"What is AI?", "How does ML work?"},
		TotalTime:      2.5,
		TotalTokens:    500,
		TTFT:           1000,
		AgentRunID:     "run-789",
		Error:          respErr,
	}

	assert.True(t, ext.IsInterrupted())
	assert.Len(t, ext.RelatedQueries, 2)
	assert.Equal(t, 2.5, ext.TotalTime)
	assert.Equal(t, int64(500), ext.TotalTokens)
	assert.Equal(t, int64(1000), ext.TTFT)
	assert.Equal(t, "run-789", ext.AgentRunID)
	assert.NotNil(t, ext.Error)
	assert.Equal(t, agentresperr.RespErrorTypeAgentExecutor, ext.Error.Type)
}

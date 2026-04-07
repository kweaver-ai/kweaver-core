package v2agentexecutordto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAgentTerminateReq_StructFields(t *testing.T) {
	t.Parallel()

	req := &AgentTerminateReq{
		AgentRunID: "run-123",
	}

	assert.Equal(t, "run-123", req.AgentRunID)
}

func TestAgentTerminateReq_Empty(t *testing.T) {
	t.Parallel()

	req := &AgentTerminateReq{}

	assert.Empty(t, req.AgentRunID)
}

func TestAgentTerminateReq_WithID(t *testing.T) {
	t.Parallel()

	req := &AgentTerminateReq{
		AgentRunID: "agent-run-abc-xyz",
	}

	assert.Equal(t, "agent-run-abc-xyz", req.AgentRunID)
}

func TestAgentTerminateReq_LongID(t *testing.T) {
	t.Parallel()

	longID := "very-long-agent-run-id-with-many-characters"

	req := &AgentTerminateReq{
		AgentRunID: longID,
	}

	assert.Equal(t, longID, req.AgentRunID)
	assert.Len(t, req.AgentRunID, len(longID))
}

package cpmsreq

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckAgentRunReq_StructFields(t *testing.T) {
	t.Parallel()

	req := &CheckAgentRunReq{
		AgentID:       "agent-123",
		CustomSpaceID: "space-456",
		UserID:        "user-789",
		AppAccountID:  "app-101",
	}

	assert.Equal(t, "agent-123", req.AgentID)
	assert.Equal(t, "space-456", req.CustomSpaceID)
	assert.Equal(t, "user-789", req.UserID)
	assert.Equal(t, "app-101", req.AppAccountID)
}

func TestCheckAgentRunReq_GetErrMsgMap(t *testing.T) {
	t.Parallel()

	req := &CheckAgentRunReq{}
	errMap := req.GetErrMsgMap()

	assert.NotNil(t, errMap)
	assert.Len(t, errMap, 1)
	assert.Contains(t, errMap, "AgentID.required")
	assert.Equal(t, `"agent_id"不能为空`, errMap["AgentID.required"])
}

func TestCheckAgentRunReq_EmptyValues(t *testing.T) {
	t.Parallel()

	req := &CheckAgentRunReq{}

	assert.Empty(t, req.AgentID)
	assert.Empty(t, req.CustomSpaceID)
	assert.Empty(t, req.UserID)
	assert.Empty(t, req.AppAccountID)
}

func TestCheckAgentRunReq_WithOnlyRequiredFields(t *testing.T) {
	t.Parallel()

	req := &CheckAgentRunReq{
		AgentID: "agent-123",
	}

	assert.Equal(t, "agent-123", req.AgentID)
	assert.Empty(t, req.CustomSpaceID)
	assert.Empty(t, req.UserID)
	assert.Empty(t, req.AppAccountID)
}

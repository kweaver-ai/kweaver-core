package sessionreq

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSessionManageActionType_Constants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, SessionManageActionType("get_info_or_create"), SessionManageActionGetInfoOrCreate)
	assert.Equal(t, SessionManageActionType("recover_lifetime_or_create"), SessionManageActionRecoverLifetimeOrCreate)
}

func TestSessionManageActionType_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		action   SessionManageActionType
		expected string
	}{
		{
			name:     "get_info_or_create action",
			action:   SessionManageActionGetInfoOrCreate,
			expected: "get_info_or_create",
		},
		{
			name:     "recover_lifetime_or_create action",
			action:   SessionManageActionRecoverLifetimeOrCreate,
			expected: "recover_lifetime_or_create",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := string(tt.action)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestManageReq_StructFields(t *testing.T) {
	t.Parallel()

	req := ManageReq{
		ConversationID: "conv-123",
		Action:         SessionManageActionGetInfoOrCreate,
		AgentID:        "agent-456",
		AgentVersion:   "1.0.0",
	}

	assert.Equal(t, "conv-123", req.ConversationID)
	assert.Equal(t, SessionManageActionGetInfoOrCreate, req.Action)
	assert.Equal(t, "agent-456", req.AgentID)
	assert.Equal(t, "1.0.0", req.AgentVersion)
}

func TestManageReq_Empty(t *testing.T) {
	t.Parallel()

	req := ManageReq{}

	assert.Empty(t, req.ConversationID)
	assert.Empty(t, req.Action)
	assert.Empty(t, req.AgentID)
	assert.Empty(t, req.AgentVersion)
}

func TestManageReq_GetErrMsgMap(t *testing.T) {
	t.Parallel()

	req := ManageReq{}

	errMsgMap := req.GetErrMsgMap()

	assert.NotNil(t, errMsgMap)
	assert.Equal(t, `"action"不能为空`, errMsgMap["Action.required"])
	assert.Equal(t, `"agent_id"不能为空`, errMsgMap["AgentID.required"])
	assert.Equal(t, `"agent_version"不能为空`, errMsgMap["AgentVersion.required"])
}

func TestManageReq_ReqCheck(t *testing.T) {
	t.Parallel()

	req := ManageReq{
		Action:       SessionManageActionGetInfoOrCreate,
		AgentID:      "agent-123",
		AgentVersion: "1.0.0",
	}

	err := req.ReqCheck()

	assert.NoError(t, err)
}

func TestManageReq_ReqCheck_Empty(t *testing.T) {
	t.Parallel()

	req := ManageReq{}

	err := req.ReqCheck()

	assert.NoError(t, err)
}

func TestManageReq_WithGetInfoOrCreateAction(t *testing.T) {
	t.Parallel()

	req := ManageReq{
		Action:       SessionManageActionGetInfoOrCreate,
		AgentID:      "agent-123",
		AgentVersion: "1.0.0",
	}

	assert.Equal(t, SessionManageActionGetInfoOrCreate, req.Action)
	err := req.ReqCheck()
	assert.NoError(t, err)
}

func TestManageReq_WithRecoverLifetimeOrCreateAction(t *testing.T) {
	t.Parallel()

	req := ManageReq{
		Action:       SessionManageActionRecoverLifetimeOrCreate,
		AgentID:      "agent-456",
		AgentVersion: "2.0.0",
	}

	assert.Equal(t, SessionManageActionRecoverLifetimeOrCreate, req.Action)
	err := req.ReqCheck()
	assert.NoError(t, err)
}

func TestManageReq_WithDifferentAgentVersions(t *testing.T) {
	t.Parallel()

	versions := []string{
		"1.0.0",
		"2.1.3",
		"3.0.0-alpha",
		"latest",
	}

	for _, version := range versions {
		req := ManageReq{
			Action:       SessionManageActionGetInfoOrCreate,
			AgentID:      "agent-123",
			AgentVersion: version,
		}
		assert.Equal(t, version, req.AgentVersion)
	}
}

func TestManageReq_WithConversationID(t *testing.T) {
	t.Parallel()

	convIDs := []string{
		"conv-001",
		"conv-xyz",
		"会话-123",
		"",
	}

	for _, convID := range convIDs {
		req := ManageReq{
			ConversationID: convID,
			Action:         SessionManageActionGetInfoOrCreate,
			AgentID:        "agent-123",
			AgentVersion:   "1.0.0",
		}
		assert.Equal(t, convID, req.ConversationID)
	}
}

func TestManageReq_WithAgentID(t *testing.T) {
	t.Parallel()

	agentIDs := []string{
		"agent-001",
		"agent-xyz",
		"智能体-123",
	}

	for _, agentID := range agentIDs {
		req := ManageReq{
			Action:       SessionManageActionGetInfoOrCreate,
			AgentID:      agentID,
			AgentVersion: "1.0.0",
		}
		assert.Equal(t, agentID, req.AgentID)
	}
}

func TestManageReq_WithAllFields(t *testing.T) {
	t.Parallel()

	req := ManageReq{
		ConversationID: "conv-complete",
		Action:         SessionManageActionRecoverLifetimeOrCreate,
		AgentID:        "agent-complete",
		AgentVersion:   "2.5.0",
	}

	assert.Equal(t, "conv-complete", req.ConversationID)
	assert.Equal(t, SessionManageActionRecoverLifetimeOrCreate, req.Action)
	assert.Equal(t, "agent-complete", req.AgentID)
	assert.Equal(t, "2.5.0", req.AgentVersion)
}

func TestManageReq_WithCustomAction(t *testing.T) {
	t.Parallel()

	customAction := SessionManageActionType("custom_action")
	req := ManageReq{
		Action:       customAction,
		AgentID:      "agent-123",
		AgentVersion: "1.0.0",
	}

	assert.Equal(t, customAction, req.Action)
}

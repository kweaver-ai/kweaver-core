package releasereq

import (
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/daenum"
	"github.com/stretchr/testify/assert"
)

func TestNewPublishReq(t *testing.T) {
	t.Parallel()

	req := NewPublishReq()

	assert.NotNil(t, req)
	assert.NotNil(t, req.UpdatePublishInfoReq)
	assert.IsType(t, &PublishReq{}, req)
}

func TestPublishReq_StructFields(t *testing.T) {
	t.Parallel()

	req := &PublishReq{
		UserID:               "user-123",
		AgentID:              "agent-456",
		BusinessDomainID:     "domain-789",
		IsInternalAPI:        true,
		UpdatePublishInfoReq: &UpdatePublishInfoReq{},
	}
	req.Description = "Test description"

	assert.Equal(t, "user-123", req.UserID)
	assert.Equal(t, "agent-456", req.AgentID)
	assert.Equal(t, "domain-789", req.BusinessDomainID)
	assert.True(t, req.IsInternalAPI)
	assert.Equal(t, "Test description", req.Description)
}

func TestPublishReq_Empty(t *testing.T) {
	t.Parallel()

	req := &PublishReq{}

	assert.Empty(t, req.UserID)
	assert.Empty(t, req.AgentID)
	assert.Empty(t, req.BusinessDomainID)
	assert.False(t, req.IsInternalAPI)
}

func TestPublishReq_GetErrMsgMap(t *testing.T) {
	t.Parallel()

	req := &PublishReq{}

	errMsgMap := req.GetErrMsgMap()

	assert.NotNil(t, errMsgMap)
	assert.Equal(t, `"Agent ID 不能为空`, errMsgMap["AgentID.required"])
}

func TestPublishReq_ReqCheck_Valid(t *testing.T) {
	t.Parallel()

	req := &PublishReq{
		AgentID:              "agent-123",
		UpdatePublishInfoReq: &UpdatePublishInfoReq{},
	}
	req.CategoryIDs = []string{"cat-1"}
	req.PublishToWhere = []daenum.PublishToWhere{
		daenum.PublishToWhereSquare,
	}
	req.PublishToBes = []cdaenum.PublishToBe{
		cdaenum.PublishToBeAPIAgent,
	}

	err := req.ReqCheck()

	assert.NoError(t, err)
}

func TestPublishReq_ReqCheck_EmptyAgentID(t *testing.T) {
	t.Parallel()

	req := &PublishReq{
		AgentID:              "",
		UpdatePublishInfoReq: &UpdatePublishInfoReq{},
	}
	req.PublishToWhere = []daenum.PublishToWhere{
		daenum.PublishToWhereSquare,
	}

	err := req.ReqCheck()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "agent_id is required")
}

func TestPublishReq_ReqCheck_InvalidPublishInfo(t *testing.T) {
	t.Parallel()

	req := &PublishReq{
		AgentID:              "agent-123",
		UpdatePublishInfoReq: &UpdatePublishInfoReq{},
	}
	req.CategoryIDs = []string{"cat-1"}
	req.PublishToBes = []cdaenum.PublishToBe{
		cdaenum.PublishToBe("invalid"),
	}

	err := req.ReqCheck()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "publish_to_bes is invalid")
}

func TestPublishReq_WithAllFields(t *testing.T) {
	t.Parallel()

	req := &PublishReq{
		UserID:               "user-001",
		AgentID:              "agent-002",
		BusinessDomainID:     "domain-003",
		IsInternalAPI:        false,
		UpdatePublishInfoReq: &UpdatePublishInfoReq{},
	}
	req.CategoryIDs = []string{"cat-1", "cat-2"}
	req.Description = "Full description"
	req.PublishToWhere = []daenum.PublishToWhere{
		daenum.PublishToWhereSquare,
	}
	req.PublishToBes = []cdaenum.PublishToBe{
		cdaenum.PublishToBeAPIAgent,
		cdaenum.PublishToBeWebSDKAgent,
	}

	err := req.ReqCheck()

	assert.NoError(t, err)
	assert.Equal(t, "user-001", req.UserID)
	assert.Equal(t, "agent-002", req.AgentID)
	assert.Len(t, req.CategoryIDs, 2)
	assert.Len(t, req.PublishToBes, 2)
}

func TestPublishReq_IsInternalAPI(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		isInternalAPI bool
	}{
		{
			name:          "internal API",
			isInternalAPI: true,
		},
		{
			name:          "external API",
			isInternalAPI: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := &PublishReq{
				AgentID:       "agent-123",
				IsInternalAPI: tt.isInternalAPI,
			}

			assert.Equal(t, tt.isInternalAPI, req.IsInternalAPI)
		})
	}
}

func TestPublishReq_WithBusinessDomainID(t *testing.T) {
	t.Parallel()

	domainIDs := []string{
		"domain-001",
		"domain-xyz",
		"domain-中文",
		"",
	}

	for _, domainID := range domainIDs {
		req := &PublishReq{
			AgentID:          "agent-123",
			BusinessDomainID: domainID,
		}

		assert.Equal(t, domainID, req.BusinessDomainID)
	}
}

func TestPublishReq_WithUserID(t *testing.T) {
	t.Parallel()

	userIDs := []string{
		"user-001",
		"user-admin",
		"user-中文-user",
	}

	for _, userID := range userIDs {
		req := &PublishReq{
			AgentID: "agent-123",
			UserID:  userID,
		}

		assert.Equal(t, userID, req.UserID)
	}
}

func TestPublishReq_WithInvalidAgentID(t *testing.T) {
	t.Parallel()

	invalidAgentIDs := []string{
		"",
	}

	for _, agentID := range invalidAgentIDs {
		req := &PublishReq{
			AgentID: agentID,
		}

		err := req.ReqCheck()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "agent_id is required")
	}
}

func TestPublishReq_WithValidAgentID(t *testing.T) {
	t.Parallel()

	validAgentIDs := []string{
		"agent-001",
		"agent-xyz",
		"agent-中文",
		"12345",
		"agent-with-dashes-and_underscores",
	}

	for _, agentID := range validAgentIDs {
		req := &PublishReq{
			AgentID:              agentID,
			UpdatePublishInfoReq: &UpdatePublishInfoReq{},
		}
		req.CategoryIDs = []string{"cat-1"}
		req.PublishToWhere = []daenum.PublishToWhere{
			daenum.PublishToWhereSquare,
		}

		err := req.ReqCheck()

		assert.NoError(t, err)
	}
}

func TestPublishReq_EmptyUpdatePublishInfoReq(t *testing.T) {
	t.Parallel()

	req := NewPublishReq()
	req.AgentID = "agent-123"

	err := req.ReqCheck()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "category_ids is required")
}

func TestPublishReq_ReqCheck_WithNilUpdatePublishInfoReq(t *testing.T) {
	t.Parallel()

	req := &PublishReq{AgentID: "agent-123"}

	assert.NotPanics(t, func() {
		err := req.ReqCheck()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "category_ids is required")
	})
}

func TestPublishReq_ReqCheck_DefaultPublishToWhere(t *testing.T) {
	t.Parallel()

	req := NewPublishReq()
	req.AgentID = "agent-123"
	req.CategoryIDs = []string{"cat-1"}

	err := req.ReqCheck()

	assert.NoError(t, err)
	assert.Equal(t, []daenum.PublishToWhere{daenum.PublishToWhereSquare}, req.PublishToWhere)
	assert.Empty(t, req.PublishToBes)
}

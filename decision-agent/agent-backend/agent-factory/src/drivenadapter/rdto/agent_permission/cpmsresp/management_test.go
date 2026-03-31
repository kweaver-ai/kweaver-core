package cpmsresp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAgentPermission_StructFields(t *testing.T) {
	t.Parallel()

	perm := &AgentPermission{
		Publish:                  true,
		Unpublish:                true,
		UnpublishOtherUserAgent:  true,
		PublishToBeSkillAgent:    true,
		PublishToBeWebSdkAgent:   true,
		PublishToBeApiAgent:      true,
		PublishToBeDataFlowAgent: true,
		CreateSystemAgent:        true,
		MgntBuiltInAgent:         true,
		SeeTrajectoryAnalysis:    true,
	}

	assert.True(t, perm.Publish)
	assert.True(t, perm.Unpublish)
	assert.True(t, perm.UnpublishOtherUserAgent)
	assert.True(t, perm.PublishToBeSkillAgent)
	assert.True(t, perm.PublishToBeWebSdkAgent)
	assert.True(t, perm.PublishToBeApiAgent)
	assert.True(t, perm.PublishToBeDataFlowAgent)
	assert.True(t, perm.CreateSystemAgent)
	assert.True(t, perm.MgntBuiltInAgent)
	assert.True(t, perm.SeeTrajectoryAnalysis)
}

func TestAgentTplPermission_StructFields(t *testing.T) {
	t.Parallel()

	perm := &AgentTplPermission{
		Publish:                    true,
		Unpublish:                  true,
		UnpublishOtherUserAgentTpl: true,
	}

	assert.True(t, perm.Publish)
	assert.True(t, perm.Unpublish)
	assert.True(t, perm.UnpublishOtherUserAgentTpl)
}

func TestUserStatusResp_StructFields(t *testing.T) {
	t.Parallel()

	resp := &UserStatusResp{
		Agent:    AgentPermission{Publish: true},
		AgentTpl: AgentTplPermission{Publish: true},
	}

	assert.True(t, resp.Agent.Publish)
	assert.True(t, resp.AgentTpl.Publish)
}

func TestNewUserStatusResp(t *testing.T) {
	t.Parallel()

	resp := NewUserStatusResp()

	assert.NotNil(t, resp)
	assert.False(t, resp.Agent.Publish)
	assert.False(t, resp.AgentTpl.Publish)
}

func TestNewUserStatusRespAllAllowed(t *testing.T) {
	t.Parallel()

	resp := NewUserStatusRespAllAllowed()

	assert.NotNil(t, resp)
	assert.True(t, resp.Agent.Publish)
	assert.True(t, resp.Agent.Unpublish)
	assert.True(t, resp.Agent.UnpublishOtherUserAgent)
	assert.True(t, resp.Agent.PublishToBeSkillAgent)
	assert.True(t, resp.Agent.PublishToBeWebSdkAgent)
	assert.True(t, resp.Agent.PublishToBeApiAgent)
	assert.True(t, resp.Agent.PublishToBeDataFlowAgent)
	assert.True(t, resp.Agent.CreateSystemAgent)
	assert.True(t, resp.Agent.MgntBuiltInAgent)
	assert.True(t, resp.Agent.SeeTrajectoryAnalysis)
	assert.True(t, resp.AgentTpl.Publish)
	assert.True(t, resp.AgentTpl.Unpublish)
	assert.True(t, resp.AgentTpl.UnpublishOtherUserAgentTpl)
}

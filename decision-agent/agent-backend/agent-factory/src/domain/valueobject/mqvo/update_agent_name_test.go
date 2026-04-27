package mqvo

import (
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
	"github.com/stretchr/testify/assert"
)

func TestNewUpdateAgentNameMqMsg(t *testing.T) {
	t.Parallel()

	tests := []struct {
		id        string
		agentName string
	}{
		{
			id:        "agent_123",
			agentName: "Test Agent",
		},
		{
			id:        "",
			agentName: "Agent Name",
		},
		{
			id:        "agent_456",
			agentName: "",
		},
		{
			id:        "",
			agentName: "",
		},
		{
			id:        "agent_@#$%",
			agentName: "Agent @#$%",
		},
		{
			id:        "agent_中文",
			agentName: "智能代理",
		},
		{
			id:        "agent_very_long_id_with_many_characters_1234567890",
			agentName: "This is a very long agent name that contains many words and should still work fine as a test case",
		},
	}

	for _, tt := range tests {
		t.Run(tt.id+"_"+tt.agentName, func(t *testing.T) {
			t.Parallel()

			msg := NewUpdateAgentNameMqMsg(tt.id, tt.agentName)

			assert.NotNil(t, msg)
			assert.Equal(t, tt.id, msg.ID)
			assert.Equal(t, tt.agentName, msg.Name)
			assert.Equal(t, cdaenum.ResourceTypeDataAgent, msg.Type)
		})
	}
}

func TestUpdateAgentNameMqMsg_NewInstance(t *testing.T) {
	t.Parallel()

	msg := &UpdateAgentNameMqMsg{}

	assert.NotNil(t, msg)
	assert.Equal(t, "", msg.ID)
	assert.Equal(t, "", msg.Name)
	assert.Equal(t, cdaenum.ResourceType(""), msg.Type)
}

func TestUpdateAgentNameMqMsg_WithTplType(t *testing.T) {
	t.Parallel()

	msg := &UpdateAgentNameMqMsg{
		ID:   "test_id",
		Type: cdaenum.ResourceTypeDataAgentTpl,
		Name: "Test Name",
	}

	assert.Equal(t, "test_id", msg.ID)
	assert.Equal(t, cdaenum.ResourceTypeDataAgentTpl, msg.Type)
	assert.Equal(t, "Test Name", msg.Name)
}

func TestUpdateAgentNameMqMsg_TypeConstant(t *testing.T) {
	t.Parallel()

	msg := NewUpdateAgentNameMqMsg("id", "name")

	assert.Equal(t, cdaenum.ResourceTypeDataAgent, msg.Type)
}

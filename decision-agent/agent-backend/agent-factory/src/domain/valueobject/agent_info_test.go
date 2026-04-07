package valueobject

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAgentInfo_New(t *testing.T) {
	t.Parallel()

	agentInfo := &AgentInfo{
		AgentID:      "agent-123",
		AgentName:    "Test Agent",
		AgentStatus:  "active",
		AgentVersion: "1.0.0",
	}

	assert.NotNil(t, agentInfo)
	assert.Equal(t, "agent-123", agentInfo.AgentID)
	assert.Equal(t, "Test Agent", agentInfo.AgentName)
	assert.Equal(t, "active", agentInfo.AgentStatus)
	assert.Equal(t, "1.0.0", agentInfo.AgentVersion)
}

func TestAgentInfo_Empty(t *testing.T) {
	t.Parallel()

	agentInfo := &AgentInfo{}

	assert.NotNil(t, agentInfo)
	assert.Empty(t, agentInfo.AgentID)
	assert.Empty(t, agentInfo.AgentName)
	assert.Empty(t, agentInfo.AgentStatus)
	assert.Empty(t, agentInfo.AgentVersion)
}

func TestAgentInfo_JSONSerialization(t *testing.T) {
	t.Parallel()

	agentInfo := &AgentInfo{
		AgentID:      "agent-456",
		AgentName:    "JSON Agent",
		AgentStatus:  "inactive",
		AgentVersion: "2.0.0",
	}

	// Serialize to JSON
	jsonBytes, err := json.Marshal(agentInfo)
	assert.NoError(t, err)

	// Deserialize from JSON
	var deserialized AgentInfo
	err = json.Unmarshal(jsonBytes, &deserialized)
	assert.NoError(t, err)

	assert.Equal(t, agentInfo.AgentID, deserialized.AgentID)
	assert.Equal(t, agentInfo.AgentName, deserialized.AgentName)
	assert.Equal(t, agentInfo.AgentStatus, deserialized.AgentStatus)
	assert.Equal(t, agentInfo.AgentVersion, deserialized.AgentVersion)
}

func TestAgentInfo_JSONTags(t *testing.T) {
	t.Parallel()

	agentInfo := &AgentInfo{
		AgentID:      "test-id",
		AgentName:    "test-name",
		AgentStatus:  "test-status",
		AgentVersion: "test-version",
	}

	jsonBytes, err := json.Marshal(agentInfo)
	assert.NoError(t, err)

	jsonStr := string(jsonBytes)
	assert.Contains(t, jsonStr, `"agent_id"`)
	assert.Contains(t, jsonStr, `"agent_name"`)
	assert.Contains(t, jsonStr, `"agent_status"`)
	assert.Contains(t, jsonStr, `"agent_version"`)
}

func TestAgentInfo_WithEmptyFields(t *testing.T) {
	t.Parallel()

	agentInfo := &AgentInfo{
		AgentID:      "id-only",
		AgentName:    "",
		AgentStatus:  "",
		AgentVersion: "",
	}

	assert.Equal(t, "id-only", agentInfo.AgentID)
	assert.Empty(t, agentInfo.AgentName)
	assert.Empty(t, agentInfo.AgentStatus)
	assert.Empty(t, agentInfo.AgentVersion)
}

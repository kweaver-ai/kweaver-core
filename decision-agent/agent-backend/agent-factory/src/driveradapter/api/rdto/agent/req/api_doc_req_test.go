package agentreq

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAPIDocReq_StructFields(t *testing.T) {
	t.Parallel()

	req := GetAPIDocReq{
		AppKey:       "app-123",
		AgentID:      "agent-456",
		AgentVersion: "v1.0.0",
	}

	assert.Equal(t, "app-123", req.AppKey)
	assert.Equal(t, "agent-456", req.AgentID)
	assert.Equal(t, "v1.0.0", req.AgentVersion)
}

func TestGetAPIDocReq_Empty(t *testing.T) {
	t.Parallel()

	req := GetAPIDocReq{}

	assert.Empty(t, req.AppKey)
	assert.Empty(t, req.AgentID)
	assert.Empty(t, req.AgentVersion)
}

func TestGetAPIDocReq_WithAppKey(t *testing.T) {
	t.Parallel()

	appKeys := []string{
		"app-001",
		"test-app",
		"APP-ABC-123",
		"",
	}

	for _, appKey := range appKeys {
		req := GetAPIDocReq{AppKey: appKey}
		assert.Equal(t, appKey, req.AppKey)
	}
}

func TestGetAPIDocReq_WithAgentID(t *testing.T) {
	t.Parallel()

	agentIDs := []string{
		"agent-001",
		"test-agent",
		"AGENT-ABC-123",
		"",
	}

	for _, agentID := range agentIDs {
		req := GetAPIDocReq{AgentID: agentID}
		assert.Equal(t, agentID, req.AgentID)
	}
}

func TestGetAPIDocReq_WithAgentVersion(t *testing.T) {
	t.Parallel()

	versions := []string{
		"v1.0.0",
		"v2.1.3",
		"latest",
		"",
	}

	for _, version := range versions {
		req := GetAPIDocReq{AgentVersion: version}
		assert.Equal(t, version, req.AgentVersion)
	}
}

func TestGetAPIDocReq_WithAllFields(t *testing.T) {
	t.Parallel()

	req := GetAPIDocReq{
		AppKey:       "my-app",
		AgentID:      "my-agent",
		AgentVersion: "v3.0.0",
	}

	assert.Equal(t, "my-app", req.AppKey)
	assert.Equal(t, "my-agent", req.AgentID)
	assert.Equal(t, "v3.0.0", req.AgentVersion)
}

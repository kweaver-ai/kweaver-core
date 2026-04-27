package skillvalobj

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSkillMCP_ValObjCheck_Valid(t *testing.T) {
	t.Parallel()

	mcp := &SkillMCP{
		MCPServerID: "mcp-server-123",
	}

	err := mcp.ValObjCheck()

	assert.NoError(t, err)
}

func TestSkillMCP_ValObjCheck_EmptyServerID(t *testing.T) {
	t.Parallel()

	mcp := &SkillMCP{
		MCPServerID: "",
	}

	err := mcp.ValObjCheck()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mcp_server_id is required")
}

func TestSkillMCP_NewSkillMCP(t *testing.T) {
	t.Parallel()

	mcp := &SkillMCP{
		MCPServerID: "test-mcp-server",
	}

	assert.Equal(t, "test-mcp-server", mcp.MCPServerID)
}

func TestSkillMCP_Empty(t *testing.T) {
	t.Parallel()

	mcp := &SkillMCP{}

	assert.Empty(t, mcp.MCPServerID)
}

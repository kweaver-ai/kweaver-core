package comvalobj

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDataAgentUniqFlag_New(t *testing.T) {
	t.Parallel()

	flag := NewDataAgentUniqFlag("agent-123", "1.0.0")

	assert.NotNil(t, flag)
	assert.Equal(t, "agent-123", flag.AgentID)
	assert.Equal(t, "1.0.0", flag.AgentVersion)
}

func TestDataAgentUniqFlag_ValObjCheck_Valid(t *testing.T) {
	t.Parallel()

	flag := NewDataAgentUniqFlag("agent-456", "2.0.0")

	err := flag.ValObjCheck()
	assert.NoError(t, err)
}

func TestDataAgentUniqFlag_ValObjCheck_EmptyAgentID(t *testing.T) {
	t.Parallel()

	flag := NewDataAgentUniqFlag("", "1.0.0")

	err := flag.ValObjCheck()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "agent_id is required")
}

func TestDataAgentUniqFlag_ValObjCheck_EmptyAgentVersion(t *testing.T) {
	t.Parallel()

	flag := NewDataAgentUniqFlag("agent-789", "")

	err := flag.ValObjCheck()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "agent_version is required")
}

func TestDataAgentUniqFlag_ValObjCheck_BothEmpty(t *testing.T) {
	t.Parallel()

	flag := NewDataAgentUniqFlag("", "")

	err := flag.ValObjCheck()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "agent_id is required")
}

func TestDataAgentUniqFlag_IsUnpublish_True(t *testing.T) {
	t.Parallel()

	flag := NewDataAgentUniqFlag("agent-999", "v0")

	result := flag.IsUnpublish()
	assert.True(t, result)
}

func TestDataAgentUniqFlag_IsUnpublish_False(t *testing.T) {
	t.Parallel()

	flag := NewDataAgentUniqFlag("agent-888", "1.0.0")

	result := flag.IsUnpublish()
	assert.False(t, result)
}

func TestDataAgentUniqFlag_EmptyFields(t *testing.T) {
	t.Parallel()

	flag := &DataAgentUniqFlag{}

	assert.NotNil(t, flag)
	assert.Empty(t, flag.AgentID)
	assert.Empty(t, flag.AgentVersion)
}

func TestDataAgentUniqFlag_WithSpecialCharacters(t *testing.T) {
	t.Parallel()

	flag := NewDataAgentUniqFlag("agent-中文-123", "版本-1.0.0")

	assert.Equal(t, "agent-中文-123", flag.AgentID)
	assert.Equal(t, "版本-1.0.0", flag.AgentVersion)
}

func TestDataAgentUniqFlag_LongStrings(t *testing.T) {
	t.Parallel()

	longID := "agent-" + string(make([]byte, 1000))
	longVersion := string(make([]byte, 1000))

	flag := NewDataAgentUniqFlag(longID, longVersion)

	assert.Equal(t, longID, flag.AgentID)
	assert.Equal(t, longVersion, flag.AgentVersion)
}

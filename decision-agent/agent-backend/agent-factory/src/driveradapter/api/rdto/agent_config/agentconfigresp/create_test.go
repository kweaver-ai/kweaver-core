package agentconfigresp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateRes_StructFields(t *testing.T) {
	t.Parallel()

	resp := CreateRes{
		ID:      "agent-123",
		Version: "v1.0.0",
	}

	assert.Equal(t, "agent-123", resp.ID)
	assert.Equal(t, "v1.0.0", resp.Version)
}

func TestCreateRes_Empty(t *testing.T) {
	t.Parallel()

	resp := CreateRes{}

	assert.Empty(t, resp.ID)
	assert.Empty(t, resp.Version)
}

func TestCreateRes_WithID(t *testing.T) {
	t.Parallel()

	ids := []string{
		"agent-001",
		"new-agent-123",
		"created-agent",
		"",
	}

	for _, id := range ids {
		resp := CreateRes{ID: id}
		assert.Equal(t, id, resp.ID)
	}
}

func TestCreateRes_WithVersion(t *testing.T) {
	t.Parallel()

	versions := []string{
		"v1.0.0",
		"v2.5.3",
		"unpublished",
		"v0.0.1",
		"",
	}

	for _, version := range versions {
		resp := CreateRes{Version: version}
		assert.Equal(t, version, resp.Version)
	}
}

func TestCreateRes_AllFields(t *testing.T) {
	t.Parallel()

	resp := CreateRes{
		ID:      "complete-agent-id",
		Version: "v3.0.0",
	}

	assert.Equal(t, "complete-agent-id", resp.ID)
	assert.Equal(t, "v3.0.0", resp.Version)
}

func TestCreateRes_WithChineseID(t *testing.T) {
	t.Parallel()

	resp := CreateRes{
		ID: "智能体-123",
	}

	assert.Equal(t, "智能体-123", resp.ID)
}

func TestCreateRes_WithUnpublishedVersion(t *testing.T) {
	t.Parallel()

	resp := CreateRes{
		ID:      "new-agent",
		Version: "unpublished",
	}

	assert.Equal(t, "new-agent", resp.ID)
	assert.Equal(t, "unpublished", resp.Version)
}

func TestCreateRes_EmptyVersion(t *testing.T) {
	t.Parallel()

	resp := CreateRes{
		ID: "agent-no-version",
	}

	assert.Empty(t, resp.Version)
	assert.Equal(t, "agent-no-version", resp.ID)
}

func TestCreateRes_NilPointer(t *testing.T) {
	t.Parallel()

	var resp *CreateRes

	assert.Nil(t, resp)
}

func TestCreateRes_NewInstance(t *testing.T) {
	t.Parallel()

	resp := &CreateRes{}

	assert.NotNil(t, resp)
	assert.Empty(t, resp.ID)
	assert.Empty(t, resp.Version)
}

func TestCreateRes_WithSpecialCharacters(t *testing.T) {
	t.Parallel()

	resp := CreateRes{
		ID: "agent-with-special_chars-@#$",
	}

	assert.Equal(t, "agent-with-special_chars-@#$", resp.ID)
}

func TestCreateRes_LongVersion(t *testing.T) {
	t.Parallel()

	longVersion := "v100.200.300-beta+build12345"
	resp := CreateRes{
		Version: longVersion,
	}

	assert.Equal(t, longVersion, resp.Version)
}

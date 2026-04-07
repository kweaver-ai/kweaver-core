package agentconfigresp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCopyResp_StructFields(t *testing.T) {
	t.Parallel()

	resp := CopyResp{
		ID:      "agent-123",
		Name:    "Copied Agent",
		Key:     "copied-agent-key",
		Version: "v1.0.0",
	}

	assert.Equal(t, "agent-123", resp.ID)
	assert.Equal(t, "Copied Agent", resp.Name)
	assert.Equal(t, "copied-agent-key", resp.Key)
	assert.Equal(t, "v1.0.0", resp.Version)
}

func TestCopyResp_Empty(t *testing.T) {
	t.Parallel()

	resp := CopyResp{}

	assert.Empty(t, resp.ID)
	assert.Empty(t, resp.Name)
	assert.Empty(t, resp.Key)
	assert.Empty(t, resp.Version)
}

func TestCopyResp_WithID(t *testing.T) {
	t.Parallel()

	ids := []string{
		"agent-001",
		"agent-copy-123",
		"复制-智能体",
		"",
	}

	for _, id := range ids {
		resp := CopyResp{ID: id}
		assert.Equal(t, id, resp.ID)
	}
}

func TestCopyResp_WithVersion(t *testing.T) {
	t.Parallel()

	versions := []string{
		"v1.0.0",
		"v2.5.3",
		"unpublished",
		"",
	}

	for _, version := range versions {
		resp := CopyResp{Version: version}
		assert.Equal(t, version, resp.Version)
	}
}

func TestCopy2TplResp_StructFields(t *testing.T) {
	t.Parallel()

	resp := Copy2TplResp{
		ID:   12345,
		Name: "Copied Template",
		Key:  "copied-tpl-key",
	}

	assert.Equal(t, int64(12345), resp.ID)
	assert.Equal(t, "Copied Template", resp.Name)
	assert.Equal(t, "copied-tpl-key", resp.Key)
}

func TestCopy2TplResp_Empty(t *testing.T) {
	t.Parallel()

	resp := Copy2TplResp{}

	assert.Zero(t, resp.ID)
	assert.Empty(t, resp.Name)
	assert.Empty(t, resp.Key)
}

func TestCopy2TplResp_WithID(t *testing.T) {
	t.Parallel()

	ids := []int64{
		0,
		1,
		123,
		456,
		999999,
	}

	for _, id := range ids {
		resp := Copy2TplResp{ID: id}
		assert.Equal(t, id, resp.ID)
	}
}

func TestCopy2TplResp_WithChineseName(t *testing.T) {
	t.Parallel()

	resp := Copy2TplResp{
		Name: "复制的模板名称",
	}

	assert.Equal(t, "复制的模板名称", resp.Name)
}

func TestCopyResp_AllFields(t *testing.T) {
	t.Parallel()

	resp := CopyResp{
		ID:      "complete-agent-id",
		Name:    "Complete Copied Agent",
		Key:     "complete-key",
		Version: "v2.0.0",
	}

	assert.Equal(t, "complete-agent-id", resp.ID)
	assert.Equal(t, "Complete Copied Agent", resp.Name)
	assert.Equal(t, "complete-key", resp.Key)
	assert.Equal(t, "v2.0.0", resp.Version)
}

func TestCopy2TplResp_AllFields(t *testing.T) {
	t.Parallel()

	resp := Copy2TplResp{
		ID:   789,
		Name: "Complete Template",
		Key:  "template-key-789",
	}

	assert.Equal(t, int64(789), resp.ID)
	assert.Equal(t, "Complete Template", resp.Name)
	assert.Equal(t, "template-key-789", resp.Key)
}

func TestCopyResp_WithEmptyVersion(t *testing.T) {
	t.Parallel()

	resp := CopyResp{
		ID:   "agent-empty-version",
		Name: "Agent Without Version",
		Key:  "no-version-key",
	}

	assert.Empty(t, resp.Version)
	assert.Equal(t, "agent-empty-version", resp.ID)
}

func TestCopy2TplResp_WithZeroID(t *testing.T) {
	t.Parallel()

	resp := Copy2TplResp{
		ID:   0,
		Name: "Template With Zero ID",
		Key:  "zero-id-key",
	}

	assert.Zero(t, resp.ID)
	assert.Equal(t, "Template With Zero ID", resp.Name)
}

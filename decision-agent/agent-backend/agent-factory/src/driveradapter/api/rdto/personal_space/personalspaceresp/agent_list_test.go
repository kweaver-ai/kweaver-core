package personalspaceresp

import (
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/entity/daconfeo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================== PAListPaginationMarker ====================

func TestNewPAListPaginationMarker(t *testing.T) {
	t.Parallel()

	m := NewPAListPaginationMarker()
	assert.NotNil(t, m)
	assert.Equal(t, int64(0), m.UpdatedAt)
	assert.Equal(t, "", m.LastAgentID)
}

func TestPAListPaginationMarker_ToString(t *testing.T) {
	t.Parallel()

	m := &PAListPaginationMarker{
		UpdatedAt:   1700000000,
		LastAgentID: "agent-123",
	}

	str, err := m.ToString()
	require.NoError(t, err)
	assert.NotEmpty(t, str)
}

func TestPAListPaginationMarker_LoadFromStr_Success(t *testing.T) {
	t.Parallel()

	// 先 ToString 再 LoadFromStr
	original := &PAListPaginationMarker{
		UpdatedAt:   1700000000,
		LastAgentID: "agent-123",
	}
	str, err := original.ToString()
	require.NoError(t, err)

	loaded := NewPAListPaginationMarker()
	err = loaded.LoadFromStr(str)
	require.NoError(t, err)

	assert.Equal(t, original.UpdatedAt, loaded.UpdatedAt)
	assert.Equal(t, original.LastAgentID, loaded.LastAgentID)
}

func TestPAListPaginationMarker_LoadFromStr_Empty(t *testing.T) {
	t.Parallel()

	m := NewPAListPaginationMarker()
	err := m.LoadFromStr("")
	assert.NoError(t, err)
}

func TestPAListPaginationMarker_LoadFromStr_InvalidBase64(t *testing.T) {
	t.Parallel()

	m := NewPAListPaginationMarker()
	err := m.LoadFromStr("not-valid-base64!!!")
	assert.Error(t, err)
}

func TestPAListPaginationMarker_LoadFromStr_InvalidJSON(t *testing.T) {
	t.Parallel()

	m := NewPAListPaginationMarker()
	// 有效 base64 但无效 JSON
	err := m.LoadFromStr("bm90LWpzb24=") // "not-json" in base64
	assert.Error(t, err)
}

// ==================== AgentListResp ====================

func TestNewAgentListResp(t *testing.T) {
	t.Parallel()

	resp := NewAgentListResp()
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Entries)
	assert.Empty(t, resp.Entries)
}

func TestNewAgentListItem(t *testing.T) {
	t.Parallel()

	item := NewAgentListItem()
	assert.NotNil(t, item)
	assert.NotNil(t, item.PublishInfo)
}

func TestAgentListResp_LoadFromEos_Empty(t *testing.T) {
	t.Parallel()

	resp := NewAgentListResp()
	err := resp.LoadFromEos(nil, nil)
	assert.NoError(t, err)
	assert.Empty(t, resp.Entries)
}

func TestAgentListResp_LoadFromEos_WithData(t *testing.T) {
	t.Parallel()

	resp := NewAgentListResp()
	eos := []*daconfeo.DataAgent{
		{DataAgentPo: dapo.DataAgentPo{ID: "agent-1", Name: "Test Agent"}},
		{DataAgentPo: dapo.DataAgentPo{ID: "agent-2", Name: "Test Agent 2"}},
	}

	err := resp.LoadFromEos(eos, nil)
	assert.NoError(t, err)
	assert.Len(t, resp.Entries, 2)
	assert.Equal(t, "agent-1", resp.Entries[0].ID)
	assert.Equal(t, "Test Agent", resp.Entries[0].Name)
}

func TestAgentListResp_LoadFromEos_IsLastPage(t *testing.T) {
	t.Parallel()

	resp := NewAgentListResp()
	resp.IsLastPage = true

	eos := []*daconfeo.DataAgent{
		{DataAgentPo: dapo.DataAgentPo{ID: "agent-1"}},
	}

	err := resp.LoadFromEos(eos, nil)
	assert.NoError(t, err)
	assert.Len(t, resp.Entries, 1)
	// IsLastPage=true 时不生成 marker
	assert.Empty(t, resp.PaginationMarkerStr)
}

func TestAgentListResp_GenMarkerStr_NotLastPage(t *testing.T) {
	t.Parallel()

	resp := NewAgentListResp()
	resp.IsLastPage = false

	eos := []*daconfeo.DataAgent{
		{DataAgentPo: dapo.DataAgentPo{ID: "agent-1"}},
	}

	err := resp.LoadFromEos(eos, nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.PaginationMarkerStr)

	// Verify marker can be loaded back
	marker := NewPAListPaginationMarker()
	err = marker.LoadFromStr(resp.PaginationMarkerStr)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), marker.UpdatedAt)
	assert.Equal(t, "agent-1", marker.LastAgentID)
}

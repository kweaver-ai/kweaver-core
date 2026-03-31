package personalspaceresp

import (
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/entity/daconfeo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/daenum"
	"github.com/stretchr/testify/assert"
)

func TestNewAgentTplListResp(t *testing.T) {
	t.Parallel()

	resp := NewAgentTplListResp()

	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Entries)
	assert.Empty(t, resp.Entries)
}

func TestAgentTplListItem_StructFields(t *testing.T) {
	t.Parallel()

	item := &AgentTplListItem{
		ID:                  123,
		Key:                 "tpl-key",
		IsBuiltIn:           1,
		Name:                "Test Template",
		Profile:             "Test profile",
		Status:              cdaenum.StatusPublished,
		AgentTplCreatedType: daenum.AgentTplCreatedTypeCopyFromAgent,
		CreatedBy:           "user-1",
		CreatedByName:       "User One",
		CreatedAt:           1234567890,
		UpdatedBy:           "user-2",
		UpdatedByName:       "User Two",
		UpdatedAt:           1234567891,
		PublishedAt:         1234567892,
		PublishedBy:         "user-3",
		PublishedByName:     "User Three",
	}

	assert.Equal(t, int64(123), item.ID)
	assert.Equal(t, "tpl-key", item.Key)
	assert.Equal(t, 1, item.IsBuiltIn)
	assert.Equal(t, "Test Template", item.Name)
	assert.Equal(t, cdaenum.StatusPublished, item.Status)
	assert.Equal(t, daenum.AgentTplCreatedTypeCopyFromAgent, item.AgentTplCreatedType)
	assert.Equal(t, "user-1", item.CreatedBy)
	assert.Equal(t, "User One", item.CreatedByName)
	assert.Equal(t, int64(1234567890), item.CreatedAt)
}

func TestAgentTplListResp_LoadFromEos_Empty(t *testing.T) {
	t.Parallel()

	resp := NewAgentTplListResp()
	eos := []*daconfeo.DataAgentTplListEo{}

	err := resp.LoadFromEos(eos)

	assert.NoError(t, err)
	assert.Empty(t, resp.Entries)
	assert.Empty(t, resp.PaginationMarkerStr)
}

func TestAgentTplListResp_LoadFromEos_Single(t *testing.T) {
	t.Parallel()

	resp := NewAgentTplListResp()

	eo := &daconfeo.DataAgentTplListEo{}
	// Note: The actual copying requires the EO to have fields populated,
	// but this test verifies the method exists and can be called

	err := resp.LoadFromEos([]*daconfeo.DataAgentTplListEo{eo})

	assert.NoError(t, err)
	assert.Len(t, resp.Entries, 1)
}

func TestAgentTplListResp_LoadFromEos_Multiple(t *testing.T) {
	t.Parallel()

	resp := NewAgentTplListResp()

	eos := []*daconfeo.DataAgentTplListEo{
		{},
		{},
		{},
	}

	err := resp.LoadFromEos(eos)

	assert.NoError(t, err)
	assert.Len(t, resp.Entries, 3)
}

func TestAgentTplListResp_genMarkerStr_EmptyEntries(t *testing.T) {
	t.Parallel()

	resp := &AgentTplListResp{
		Entries: []*AgentTplListItem{},
	}

	markerStr, err := resp.genMarkerStr()

	assert.NoError(t, err)
	assert.Empty(t, markerStr)
}

func TestAgentTplListResp_genMarkerStr_IsLastPage(t *testing.T) {
	t.Parallel()

	resp := &AgentTplListResp{
		Entries: []*AgentTplListItem{
			{ID: 1, UpdatedAt: 100},
		},
		IsLastPage: true,
	}

	markerStr, err := resp.genMarkerStr()

	assert.NoError(t, err)
	assert.Empty(t, markerStr)
}

func TestAgentTplListResp_StructFields(t *testing.T) {
	t.Parallel()

	marker := &PTplListPaginationMarker{}
	resp := &AgentTplListResp{
		Entries:             []*AgentTplListItem{},
		PaginationMarkerStr: "marker-string",
		Marker:              marker,
		IsLastPage:          true,
	}

	assert.NotNil(t, resp.Entries)
	assert.Equal(t, "marker-string", resp.PaginationMarkerStr)
	assert.Equal(t, marker, resp.Marker)
	assert.True(t, resp.IsLastPage)
}

func TestAgentTplListItem_Empty(t *testing.T) {
	t.Parallel()

	item := &AgentTplListItem{}

	assert.Zero(t, item.ID)
	assert.Empty(t, item.Key)
	assert.Zero(t, item.IsBuiltIn)
	assert.Empty(t, item.Name)
	assert.Empty(t, item.Profile)
}

package pubedresp

import (
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/entity/pubedeo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/stretchr/testify/assert"
)

func TestNewPAListPaginationMarker(t *testing.T) {
	t.Parallel()

	marker := NewPAListPaginationMarker()
	assert.NotNil(t, marker)
	assert.Zero(t, marker.PublishedAt)
	assert.Empty(t, marker.LastReleaseID)
}

func TestPAListPaginationMarker_ToString(t *testing.T) {
	t.Parallel()

	t.Run("valid marker", func(t *testing.T) {
		t.Parallel()

		marker := &PAListPaginationMarker{
			PublishedAt:   1234567890,
			LastReleaseID: "release-123",
		}

		str, err := marker.ToString()
		assert.NoError(t, err)
		assert.NotEmpty(t, str)
	})

	t.Run("empty marker", func(t *testing.T) {
		t.Parallel()

		marker := &PAListPaginationMarker{}

		str, err := marker.ToString()
		assert.NoError(t, err)
		assert.NotEmpty(t, str)
	})
}

func TestPAListPaginationMarker_LoadFromStr(t *testing.T) {
	t.Parallel()

	t.Run("empty string", func(t *testing.T) {
		t.Parallel()

		marker := &PAListPaginationMarker{}
		err := marker.LoadFromStr("")
		assert.NoError(t, err)
	})

	t.Run("invalid base64", func(t *testing.T) {
		t.Parallel()

		marker := &PAListPaginationMarker{}
		err := marker.LoadFromStr("invalid base64!")
		assert.Error(t, err)
	})

	t.Run("invalid json", func(t *testing.T) {
		t.Parallel()

		marker := &PAListPaginationMarker{}
		err := marker.LoadFromStr("aGVsbG8=") // "hello" in base64
		assert.Error(t, err)
	})

	t.Run("round trip", func(t *testing.T) {
		t.Parallel()

		original := &PAListPaginationMarker{
			PublishedAt:   9876543210,
			LastReleaseID: "release-xyz",
		}

		str, err := original.ToString()
		assert.NoError(t, err)

		restored := &PAListPaginationMarker{}
		err = restored.LoadFromStr(str)
		assert.NoError(t, err)
		assert.Equal(t, original.PublishedAt, restored.PublishedAt)
		assert.Equal(t, original.LastReleaseID, restored.LastReleaseID)
	})
}

func TestPAListPaginationMarker_LoadFromPos(t *testing.T) {
	t.Parallel()

	t.Run("empty pos", func(t *testing.T) {
		t.Parallel()

		marker := &PAListPaginationMarker{}
		marker.LoadFromPos([]*dapo.PublishedJoinPo{})
		assert.Zero(t, marker.PublishedAt)
		assert.Empty(t, marker.LastReleaseID)
	})

	t.Run("single pos", func(t *testing.T) {
		t.Parallel()

		pos := []*dapo.PublishedJoinPo{
			{
				ReleasePartPo: dapo.ReleasePartPo{
					ReleaseID:   "release-1",
					PublishedAt: 1234567890,
				},
			},
		}

		marker := &PAListPaginationMarker{}
		marker.LoadFromPos(pos)
		assert.Equal(t, int64(1234567890), marker.PublishedAt)
		assert.Equal(t, "release-1", marker.LastReleaseID)
	})

	t.Run("multiple pos", func(t *testing.T) {
		t.Parallel()

		pos := []*dapo.PublishedJoinPo{
			{
				ReleasePartPo: dapo.ReleasePartPo{
					ReleaseID:   "release-1",
					PublishedAt: 1000000,
				},
			},
			{
				ReleasePartPo: dapo.ReleasePartPo{
					ReleaseID:   "release-2",
					PublishedAt: 2000000,
				},
			},
		}

		marker := &PAListPaginationMarker{}
		marker.LoadFromPos(pos)
		assert.Equal(t, int64(2000000), marker.PublishedAt)
		assert.Equal(t, "release-2", marker.LastReleaseID)
	})
}

func TestNewPAListItemResp(t *testing.T) {
	t.Parallel()

	item := NewPAListItemResp()
	assert.NotNil(t, item)
	assert.NotNil(t, item.PublishInfo)
	assert.Empty(t, item.ID)
	assert.Empty(t, item.Version)
}

func TestNewPAListResp(t *testing.T) {
	t.Parallel()

	resp := NewPAListResp()
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Entries)
	assert.Empty(t, resp.Entries)
	assert.Empty(t, resp.PaginationMarkerStr)
	assert.False(t, resp.IsLastPage)
}

func TestPubedAgentListResp_LoadFromEos(t *testing.T) {
	t.Parallel()

	t.Run("empty eos", func(t *testing.T) {
		t.Parallel()

		resp := NewPAListResp()
		eos := []*pubedeo.PublishedAgentEo{}
		agentID2BdIDMap := map[string]string{}

		err := resp.LoadFromEos(eos, agentID2BdIDMap)
		assert.NoError(t, err)
		assert.Empty(t, resp.Entries)
		assert.Empty(t, resp.PaginationMarkerStr)
	})

	t.Run("with single eo", func(t *testing.T) {
		t.Parallel()

		resp := NewPAListResp()
		eo := &pubedeo.PublishedAgentEo{}
		agentID2BdIDMap := map[string]string{}

		err := resp.LoadFromEos([]*pubedeo.PublishedAgentEo{eo}, agentID2BdIDMap)
		assert.NoError(t, err)
		assert.Len(t, resp.Entries, 1)
	})

	t.Run("is last page", func(t *testing.T) {
		t.Parallel()

		resp := NewPAListResp()
		resp.IsLastPage = true
		eo := &pubedeo.PublishedAgentEo{}
		agentID2BdIDMap := map[string]string{}

		err := resp.LoadFromEos([]*pubedeo.PublishedAgentEo{eo}, agentID2BdIDMap)
		assert.NoError(t, err)
		assert.Empty(t, resp.PaginationMarkerStr)
	})
}

func TestPubedAgentListResp_genMarkerStr(t *testing.T) {
	t.Parallel()

	t.Run("empty entries", func(t *testing.T) {
		t.Parallel()

		resp := &PubedAgentListResp{
			Entries: []*PAListItemResp{},
		}

		markerStr, err := resp.genMarkerStr()
		assert.NoError(t, err)
		assert.Empty(t, markerStr)
	})

	t.Run("is last page", func(t *testing.T) {
		t.Parallel()

		resp := &PubedAgentListResp{
			Entries: []*PAListItemResp{
				{ReleaseID: "release-1", PublishedAt: 1000000},
			},
			IsLastPage: true,
		}

		markerStr, err := resp.genMarkerStr()
		assert.NoError(t, err)
		assert.Empty(t, markerStr)
	})

	t.Run("valid entries", func(t *testing.T) {
		t.Parallel()

		resp := &PubedAgentListResp{
			Entries: []*PAListItemResp{
				{ReleaseID: "release-1", PublishedAt: 1000000},
				{ReleaseID: "release-2", PublishedAt: 2000000},
			},
		}

		markerStr, err := resp.genMarkerStr()
		assert.NoError(t, err)
		assert.NotEmpty(t, markerStr)
	})
}

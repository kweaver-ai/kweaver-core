package pubedresp

import (
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/entity/pubedeo"
	"github.com/stretchr/testify/assert"
)

func TestNewPTplListPaginationMarker(t *testing.T) {
	t.Parallel()

	marker := NewPTplListPaginationMarker()
	assert.NotNil(t, marker)
	assert.Zero(t, marker.LastPubedTplID)
}

func TestPTplListPaginationMarker_ToString(t *testing.T) {
	t.Parallel()

	t.Run("valid marker", func(t *testing.T) {
		t.Parallel()

		marker := &PTplListPaginationMarker{
			LastPubedTplID: 12345,
		}

		str, err := marker.ToString()
		assert.NoError(t, err)
		assert.NotEmpty(t, str)
	})

	t.Run("empty marker", func(t *testing.T) {
		t.Parallel()

		marker := &PTplListPaginationMarker{}

		str, err := marker.ToString()
		assert.NoError(t, err)
		assert.NotEmpty(t, str)
	})
}

func TestPTplListPaginationMarker_LoadFromStr(t *testing.T) {
	t.Parallel()

	t.Run("empty string", func(t *testing.T) {
		t.Parallel()

		marker := &PTplListPaginationMarker{}
		err := marker.LoadFromStr("")
		assert.NoError(t, err)
	})

	t.Run("invalid base64", func(t *testing.T) {
		t.Parallel()

		marker := &PTplListPaginationMarker{}
		err := marker.LoadFromStr("invalid base64!")
		assert.Error(t, err)
	})

	t.Run("invalid json", func(t *testing.T) {
		t.Parallel()

		marker := &PTplListPaginationMarker{}
		err := marker.LoadFromStr("aGVsbG8=") // "hello" in base64
		assert.Error(t, err)
	})

	t.Run("round trip", func(t *testing.T) {
		t.Parallel()

		original := &PTplListPaginationMarker{
			LastPubedTplID: 999,
		}

		str, err := original.ToString()
		assert.NoError(t, err)

		restored := &PTplListPaginationMarker{}
		err = restored.LoadFromStr(str)
		assert.NoError(t, err)
		assert.Equal(t, original.LastPubedTplID, restored.LastPubedTplID)
	})
}

func TestNewPublishedAgentTplListResp(t *testing.T) {
	t.Parallel()

	resp := NewPublishedAgentTplListResp()
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Entries)
	assert.Empty(t, resp.Entries)
	assert.True(t, resp.IsLastPage)
	assert.Empty(t, resp.PaginationMarkerStr)
}

func TestPublishedAgentTplListResp_LoadFromEos(t *testing.T) {
	t.Parallel()

	t.Run("empty eos", func(t *testing.T) {
		t.Parallel()

		resp := NewPublishedAgentTplListResp()
		eos := []*pubedeo.PublishedTplListEo{}

		err := resp.LoadFromEos(eos)
		assert.NoError(t, err)
		assert.Empty(t, resp.Entries)
		assert.Empty(t, resp.PaginationMarkerStr)
	})

	t.Run("with single eo", func(t *testing.T) {
		t.Parallel()

		resp := NewPublishedAgentTplListResp()
		eo := &pubedeo.PublishedTplListEo{}

		err := resp.LoadFromEos([]*pubedeo.PublishedTplListEo{eo})
		assert.NoError(t, err)
		assert.Len(t, resp.Entries, 1)
	})

	t.Run("is last page", func(t *testing.T) {
		t.Parallel()

		resp := NewPublishedAgentTplListResp()
		resp.IsLastPage = true
		eo := &pubedeo.PublishedTplListEo{}

		err := resp.LoadFromEos([]*pubedeo.PublishedTplListEo{eo})
		assert.NoError(t, err)
		assert.Empty(t, resp.PaginationMarkerStr)
	})

	t.Run("with multiple eos", func(t *testing.T) {
		t.Parallel()

		resp := NewPublishedAgentTplListResp()
		eos := []*pubedeo.PublishedTplListEo{
			{},
			{},
		}

		err := resp.LoadFromEos(eos)
		assert.NoError(t, err)
		assert.Len(t, resp.Entries, 2)
	})
}

func TestPublishedAgentTplListResp_genMarkerStr(t *testing.T) {
	t.Parallel()

	t.Run("empty entries", func(t *testing.T) {
		t.Parallel()

		resp := &PublishedAgentTplListResp{
			Entries: []*PubedTplListItemResp{},
		}

		markerStr, err := resp.genMarkerStr()
		assert.NoError(t, err)
		assert.Empty(t, markerStr)
	})

	t.Run("is last page", func(t *testing.T) {
		t.Parallel()

		resp := &PublishedAgentTplListResp{
			Entries: []*PubedTplListItemResp{
				{ID: 123},
			},
			IsLastPage: true,
		}

		markerStr, err := resp.genMarkerStr()
		assert.NoError(t, err)
		assert.Empty(t, markerStr)
	})

	t.Run("valid entries", func(t *testing.T) {
		t.Parallel()

		resp := &PublishedAgentTplListResp{
			Entries: []*PubedTplListItemResp{
				{ID: 100},
				{ID: 200},
			},
		}

		markerStr, err := resp.genMarkerStr()
		assert.NoError(t, err)
		assert.NotEmpty(t, markerStr)
	})
}

func TestPubedTplListItemResp_Fields(t *testing.T) {
	t.Parallel()

	item := &PubedTplListItemResp{
		ID:        123,
		TplID:     456,
		Key:       "tpl-key",
		IsBuiltIn: 1,
		Name:      "Test Template",
		Profile:   "Test profile",
	}

	assert.Equal(t, int64(123), item.ID)
	assert.Equal(t, int64(456), item.TplID)
	assert.Equal(t, "tpl-key", item.Key)
	assert.Equal(t, 1, item.IsBuiltIn)
	assert.Equal(t, "Test Template", item.Name)
}

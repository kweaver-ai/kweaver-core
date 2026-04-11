package agentconfigresp

import (
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewListForBenchmarkResp(t *testing.T) {
	t.Parallel()

	resp := NewListForBenchmarkResp()

	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Entries)
	assert.Empty(t, resp.Entries)
	assert.Equal(t, int64(0), resp.Total)
}

func TestListForBenchmarkResp_LoadFromListForBenchmark(t *testing.T) {
	t.Parallel()

	t.Run("load from valid PO list", func(t *testing.T) {
		t.Parallel()

		resp := NewListForBenchmarkResp()

		pos := []*dapo.ListForBenchmarkPo{
			{
				ID:        "agent-1",
				Key:       "agent-key-1",
				Name:      "Agent 1",
				Version:   "v1.0",
				Status:    cdaenum.StatusPublished,
				UpdatedAt: 1234567890,
			},
			{
				ID:        "agent-2",
				Key:       "agent-key-2",
				Name:      "Agent 2",
				Version:   "v2.0",
				Status:    cdaenum.StatusUnpublished,
				UpdatedAt: 1234567891,
			},
		}

		err := resp.LoadFromListForBenchmark(pos)

		require.NoError(t, err)
		assert.Len(t, resp.Entries, 2)
		assert.Equal(t, "agent-1", resp.Entries[0].ID)
		assert.Equal(t, "Agent 1", resp.Entries[0].Name)
		assert.Equal(t, "agent-key-1", resp.Entries[0].Key)
		assert.Equal(t, "agent-2", resp.Entries[1].ID)
		assert.Equal(t, "Agent 2", resp.Entries[1].Name)
		assert.Equal(t, "agent-key-2", resp.Entries[1].Key)
	})

	t.Run("load from empty PO list", func(t *testing.T) {
		t.Parallel()

		resp := NewListForBenchmarkResp()

		pos := []*dapo.ListForBenchmarkPo{}

		err := resp.LoadFromListForBenchmark(pos)

		require.NoError(t, err)
		assert.Len(t, resp.Entries, 0)
	})

	t.Run("load from PO with minimal fields", func(t *testing.T) {
		t.Parallel()

		resp := NewListForBenchmarkResp()

		pos := []*dapo.ListForBenchmarkPo{
			{
				ID:        "agent-3",
				Key:       "agent-key-3",
				Name:      "Agent 3",
				Status:    cdaenum.StatusPublished,
				UpdatedAt: 1234567892,
			},
		}

		err := resp.LoadFromListForBenchmark(pos)

		require.NoError(t, err)
		assert.Len(t, resp.Entries, 1)
		assert.Equal(t, "agent-3", resp.Entries[0].ID)
		assert.Equal(t, "Agent 3", resp.Entries[0].Name)
	})

	t.Run("load from PO with all fields", func(t *testing.T) {
		t.Parallel()

		resp := NewListForBenchmarkResp()

		pos := []*dapo.ListForBenchmarkPo{
			{
				ID:        "agent-full",
				Key:       "full-agent-key",
				Name:      "Full Agent",
				Version:   "v3.0",
				Status:    cdaenum.StatusPublished,
				UpdatedAt: 1234567893,
			},
		}

		err := resp.LoadFromListForBenchmark(pos)

		require.NoError(t, err)
		assert.Len(t, resp.Entries, 1)
		assert.Equal(t, "agent-full", resp.Entries[0].ID)
		assert.Equal(t, "Full Agent", resp.Entries[0].Name)
		assert.Equal(t, "full-agent-key", resp.Entries[0].Key)
		assert.Equal(t, "v3.0", resp.Entries[0].Version)
		assert.Equal(t, cdaenum.StatusPublished, resp.Entries[0].Status)
		assert.Equal(t, int64(1234567893), resp.Entries[0].UpdatedAt)
	})
}

func TestListForBenchmarkResp_Fields(t *testing.T) {
	t.Parallel()

	resp := &ListForBenchmarkResp{
		Entries: []*ListForBenchmarkItem{
			{
				ListForBenchmarkPo: dapo.ListForBenchmarkPo{
					ID:        "test-agent",
					Key:       "test-key",
					Name:      "Test Agent",
					Status:    cdaenum.StatusPublished,
					UpdatedAt: 1234567890,
				},
			},
		},
		Total: 100,
	}

	assert.Len(t, resp.Entries, 1)
	assert.Equal(t, "test-agent", resp.Entries[0].ID)
	assert.Equal(t, "Test Agent", resp.Entries[0].Name)
	assert.Equal(t, int64(100), resp.Total)
}

func TestListForBenchmarkResp_Empty(t *testing.T) {
	t.Parallel()

	resp := &ListForBenchmarkResp{}

	assert.Nil(t, resp.Entries)
	assert.Equal(t, int64(0), resp.Total)
}

func TestListForBenchmarkItem_StructFields(t *testing.T) {
	t.Parallel()

	item := &ListForBenchmarkItem{
		ListForBenchmarkPo: dapo.ListForBenchmarkPo{
			ID:        "item-agent",
			Key:       "item-key",
			Name:      "Item Agent",
			Version:   "v1.0",
			Status:    cdaenum.StatusPublished,
			UpdatedAt: 1234567891,
		},
	}

	assert.Equal(t, "item-agent", item.ID)
	assert.Equal(t, "Item Agent", item.Name)
	assert.Equal(t, "item-key", item.Key)
	assert.Equal(t, "v1.0", item.Version)
	assert.Equal(t, cdaenum.StatusPublished, item.Status)
}

package pubedresp

import (
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/entity/pubedeo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/daconfvalobj"
	"github.com/stretchr/testify/assert"
)

func TestNewPublishedAgentInfoListItem(t *testing.T) {
	t.Parallel()

	item := NewPublishedAgentInfoListItem()
	assert.NotNil(t, item)
	assert.NotNil(t, item.PublishInfo)
	assert.Empty(t, item.ID)
	assert.Nil(t, item.Config)
}

func TestPublishedAgentInfoListItem_HlConfig(t *testing.T) {
	t.Parallel()

	t.Run("with input field", func(t *testing.T) {
		t.Parallel()

		item := &PublishedAgentInfoListItem{
			Config: &daconfvalobj.Config{
				Input:  &daconfvalobj.Input{},
				Output: &daconfvalobj.Output{},
			},
		}

		item.HlConfig([]string{"input"})

		assert.NotNil(t, item.Config)
		assert.NotNil(t, item.Config.Input)
		assert.Nil(t, item.Config.Output)
	})

	t.Run("with no fields", func(t *testing.T) {
		t.Parallel()

		item := &PublishedAgentInfoListItem{
			Config: &daconfvalobj.Config{
				Input:  &daconfvalobj.Input{},
				Output: &daconfvalobj.Output{},
			},
		}

		item.HlConfig([]string{})

		assert.NotNil(t, item.Config)
		assert.Nil(t, item.Config.Input)
		assert.Nil(t, item.Config.Output)
	})

	t.Run("nil config", func(t *testing.T) {
		t.Parallel()

		item := &PublishedAgentInfoListItem{
			Config: nil,
		}

		assert.Panics(t, func() {
			item.HlConfig([]string{"input"})
		})
	})
}

func TestNewPublishedAgentInfoListResp(t *testing.T) {
	t.Parallel()

	resp := NewPublishedAgentInfoListResp()
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Entries)
	assert.Empty(t, resp.Entries)
}

func TestPAInfoListResp_LoadFromEos(t *testing.T) {
	t.Parallel()

	t.Run("empty eos", func(t *testing.T) {
		t.Parallel()

		resp := NewPublishedAgentInfoListResp()
		eos := []*pubedeo.PublishedAgentEo{}

		err := resp.LoadFromEos(eos, []string{})
		assert.NoError(t, err)
		assert.Empty(t, resp.Entries)
	})

	t.Run("with single eo", func(t *testing.T) {
		t.Parallel()

		resp := NewPublishedAgentInfoListResp()
		eo := &pubedeo.PublishedAgentEo{}

		err := resp.LoadFromEos([]*pubedeo.PublishedAgentEo{eo}, []string{})
		assert.NoError(t, err)
		assert.Len(t, resp.Entries, 1)
	})

	t.Run("with config fields - expects panic due to nil config", func(t *testing.T) {
		t.Parallel()

		resp := NewPublishedAgentInfoListResp()
		eo := &pubedeo.PublishedAgentEo{}

		assert.Panics(t, func() {
			resp.LoadFromEos([]*pubedeo.PublishedAgentEo{eo}, []string{"input"}) //nolint:errcheck
		})
	})
}

func TestPublishedAgentInfoListItem_Fields(t *testing.T) {
	t.Parallel()

	item := &PublishedAgentInfoListItem{
		ID:            "agent-123",
		Version:       "v1.0",
		Key:           "test-agent",
		IsBuiltIn:     1,
		IsSystemAgent: 0,
		Name:          "Test Agent",
		Profile:       "Test profile",
		AvatarType:    cdaenum.AvatarTypeBuiltIn,
		Avatar:        "🤖",
		PublishedAt:   1234567890,
		PublishedBy:   "user-1",
	}

	assert.Equal(t, "agent-123", item.ID)
	assert.Equal(t, "v1.0", item.Version)
	assert.Equal(t, "test-agent", item.Key)
	assert.Equal(t, 1, item.IsBuiltIn)
	assert.Equal(t, "Test Agent", item.Name)
	assert.Equal(t, cdaenum.AvatarTypeBuiltIn, item.AvatarType)
}

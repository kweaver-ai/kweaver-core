package publishedp2e

import (
	"context"
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPublishedAgent_Simple(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	po := &dapo.PublishedJoinPo{
		DataAgentPo: dapo.DataAgentPo{
			ID:         "agent1",
			Name:       "Test Agent",
			Key:        "test-agent",
			ProductKey: "product1",
		},
		ReleasePartPo: dapo.ReleasePartPo{
			ReleaseID:   "release1",
			Version:     "1.0.0",
			PublishDesc: "Test release",
			PublishedBy: "user1",
		},
	}

	eo, err := PublishedAgent(ctx, po, false)
	require.NoError(t, err)
	assert.NotNil(t, eo)
	assert.Equal(t, "agent1", eo.ID)
	assert.Equal(t, "Test Agent", eo.Name)
	assert.Equal(t, "test-agent", eo.Key)
	assert.Equal(t, "release1", eo.ReleaseID)
	assert.Equal(t, "1.0.0", eo.Version)
	assert.Equal(t, "user1", eo.PublishedBy)
}

func TestPublishedAgent_WithConfig(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	configJSON := `{"profile":"test profile"}`
	po := &dapo.PublishedJoinPo{
		DataAgentPo: dapo.DataAgentPo{
			ID:     "agent1",
			Name:   "Test Agent",
			Config: configJSON,
		},
		ReleasePartPo: dapo.ReleasePartPo{
			ReleaseID: "release1",
		},
	}

	eo, err := PublishedAgent(ctx, po, true)
	require.NoError(t, err)
	assert.NotNil(t, eo)
	assert.Equal(t, "agent1", eo.ID)
	assert.NotNil(t, eo.Config)
}

func TestPublishedAgent_InvalidConfig(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	invalidJSON := `{invalid json`
	po := &dapo.PublishedJoinPo{
		DataAgentPo: dapo.DataAgentPo{
			ID:     "agent1",
			Name:   "Test Agent",
			Config: invalidJSON,
		},
		ReleasePartPo: dapo.ReleasePartPo{
			ReleaseID: "release1",
		},
	}

	eo, err := PublishedAgent(ctx, po, true)
	// The function returns an error but may still return a non-nil EO
	// Just check that the error is properly wrapped
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "PublishedAgent unmarshal config error")

	_ = eo // EO may be non-nil even on error
}

func TestPublishedAgent_WithEmptyConfig(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	po := &dapo.PublishedJoinPo{
		DataAgentPo: dapo.DataAgentPo{
			ID:     "agent1",
			Name:   "Test Agent",
			Config: "",
		},
		ReleasePartPo: dapo.ReleasePartPo{
			ReleaseID: "release1",
		},
	}

	eo, err := PublishedAgent(ctx, po, true)
	require.NoError(t, err)
	assert.NotNil(t, eo)
	assert.Equal(t, "agent1", eo.ID)
}

func TestPublishedAgent_WithIsPmsCtrl(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	po := &dapo.PublishedJoinPo{
		DataAgentPo: dapo.DataAgentPo{
			ID:   "agent1",
			Name: "Test Agent",
		},
		ReleasePartPo: dapo.ReleasePartPo{
			ReleaseID: "release1",
			IsPmsCtrl: 1,
		},
	}

	eo, err := PublishedAgent(ctx, po, false)
	require.NoError(t, err)
	assert.NotNil(t, eo)
	assert.True(t, eo.IsPmsCtrlBool())
}

func TestPublishedAgent_NoUnmarshalConfig(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	configJSON := `{"profile":"test profile"}`
	po := &dapo.PublishedJoinPo{
		DataAgentPo: dapo.DataAgentPo{
			ID:     "agent1",
			Name:   "Test Agent",
			Config: configJSON,
		},
		ReleasePartPo: dapo.ReleasePartPo{
			ReleaseID: "release1",
		},
	}

	eo, err := PublishedAgent(ctx, po, false)
	require.NoError(t, err)
	assert.NotNil(t, eo)
	// Config should not be unmarshaled
	// The Config field will be empty or default
}

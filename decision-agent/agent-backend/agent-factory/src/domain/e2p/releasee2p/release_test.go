package releasee2p

import (
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/entity/releaseeo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReleaseE2P(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		entity  *releaseeo.ReleaseEO
		checkPO func(t *testing.T, po *dapo.ReleasePO)
	}{
		{
			name: "valid release entity",
			entity: &releaseeo.ReleaseEO{
				ID:           "release-1",
				AgentID:      "agent-1",
				AgentConfig:  "{\"key\":\"value\"}",
				AgentVersion: "v1.0.0",
				AgentDesc:    "Test Agent",
				PublishToBes: []cdaenum.PublishToBe{
					cdaenum.PublishToBeAPIAgent,
					cdaenum.PublishToBeWebSDKAgent,
				},
			},
			checkPO: func(t *testing.T, po *dapo.ReleasePO) {
				assert.Equal(t, "release-1", po.ID)
				assert.Equal(t, "agent-1", po.AgentID)
				assert.Equal(t, "{\"key\":\"value\"}", po.AgentConfig)
				assert.Equal(t, "v1.0.0", po.AgentVersion)
				assert.Equal(t, "Test Agent", po.AgentDesc)
				// Check that PublishToBes is set
				assert.NotNil(t, po.IsAPIAgent)
				assert.Equal(t, 1, *po.IsAPIAgent)
				assert.NotNil(t, po.IsWebSDKAgent)
				assert.Equal(t, 1, *po.IsWebSDKAgent)
			},
		},
		{
			name: "release without publish targets",
			entity: &releaseeo.ReleaseEO{
				ID:           "release-2",
				AgentID:      "agent-2",
				AgentConfig:  "{}",
				AgentVersion: "v2.0.0",
				AgentDesc:    "Test Agent 2",
				PublishToBes: []cdaenum.PublishToBe{},
			},
			checkPO: func(t *testing.T, po *dapo.ReleasePO) {
				assert.Equal(t, "release-2", po.ID)
				assert.Equal(t, "agent-2", po.AgentID)
				assert.Equal(t, "v2.0.0", po.AgentVersion)
				// All publish flags should be nil or 0
				assert.Nil(t, po.IsAPIAgent)
				assert.Nil(t, po.IsWebSDKAgent)
			},
		},
		{
			name: "release with nil publish targets",
			entity: &releaseeo.ReleaseEO{
				ID:           "release-3",
				AgentID:      "agent-3",
				AgentConfig:  "{}",
				AgentVersion: "v3.0.0",
				PublishToBes: nil,
			},
			checkPO: func(t *testing.T, po *dapo.ReleasePO) {
				assert.Equal(t, "release-3", po.ID)
				assert.Equal(t, "agent-3", po.AgentID)
			},
		},
		{
			name: "release with skill agent",
			entity: &releaseeo.ReleaseEO{
				ID:           "release-4",
				AgentID:      "agent-4",
				AgentConfig:  "{}",
				AgentVersion: "v4.0.0",
				PublishToBes: []cdaenum.PublishToBe{
					cdaenum.PublishToBeSkillAgent,
				},
			},
			checkPO: func(t *testing.T, po *dapo.ReleasePO) {
				assert.Equal(t, "release-4", po.ID)
				assert.NotNil(t, po.IsSkillAgent)
				assert.Equal(t, 1, *po.IsSkillAgent)
			},
		},
		{
			name: "release with data flow agent",
			entity: &releaseeo.ReleaseEO{
				ID:           "release-5",
				AgentID:      "agent-5",
				AgentConfig:  "{}",
				AgentVersion: "v5.0.0",
				PublishToBes: []cdaenum.PublishToBe{
					cdaenum.PublishToBeDataFlowAgent,
				},
			},
			checkPO: func(t *testing.T, po *dapo.ReleasePO) {
				assert.Equal(t, "release-5", po.ID)
				assert.NotNil(t, po.IsDataFlowAgent)
				assert.Equal(t, 1, *po.IsDataFlowAgent)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			po := ReleaseE2P(tt.entity)
			require.NotNil(t, po)

			if tt.checkPO != nil {
				tt.checkPO(t, po)
			}
		})
	}
}

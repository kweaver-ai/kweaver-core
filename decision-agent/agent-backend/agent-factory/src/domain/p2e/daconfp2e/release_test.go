package daconfp2e

import (
	"context"
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/entity/releaseeo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReleaseDAConfEoSimple(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	isPmsCtrl := 1

	tests := []struct {
		name    string
		po      *dapo.ReleasePO
		wantErr bool
		checkEo func(t *testing.T, eo *releaseeo.ReleaseDAConfWrapperEO)
	}{
		{
			name: "valid release with agent config",
			po: &dapo.ReleasePO{
				ID:      "1",
				AgentID: "agent-1",
				AgentConfig: `{
					"id": "1",
					"key": "test-agent",
					"name": "Test Agent",
					"config": "{\"input\":{\"fields\":[{\"name\":\"field1\",\"type\":\"text\"}]}}"
				}`,
			},
			wantErr: false,
			checkEo: func(t *testing.T, eo *releaseeo.ReleaseDAConfWrapperEO) {
				assert.NotNil(t, eo)
				assert.Equal(t, "1", eo.ID)
				assert.Equal(t, "agent-1", eo.AgentID)
				assert.NotNil(t, eo.Config)
				assert.NotNil(t, eo.Config.Input)
			},
		},
		{
			name: "valid release with empty agent config",
			po: &dapo.ReleasePO{
				ID:          "1",
				AgentID:     "agent-1",
				AgentConfig: "",
			},
			wantErr: false,
			checkEo: func(t *testing.T, eo *releaseeo.ReleaseDAConfWrapperEO) {
				assert.NotNil(t, eo)
				assert.NotNil(t, eo.Config)
			},
		},
		{
			name: "valid release with all fields",
			po: &dapo.ReleasePO{
				ID:           "1",
				AgentID:      "agent-1",
				AgentVersion: "1.0",
				AgentDesc:    "Test description",
				IsPmsCtrl:    &isPmsCtrl,
				AgentConfig:  `{"id":"1","key":"test-agent","name":"Test Agent","config":"{\"input\":{\"fields\":[{\"name\":\"question\",\"type\":\"text\"}]}}"}`,
			},
			wantErr: false,
			checkEo: func(t *testing.T, eo *releaseeo.ReleaseDAConfWrapperEO) {
				assert.NotNil(t, eo)
				assert.Equal(t, "1", eo.ID)
				assert.Equal(t, "agent-1", eo.AgentID)
				assert.Equal(t, "1.0", eo.AgentVersion)
				assert.Equal(t, "Test description", eo.AgentDesc)
				assert.Equal(t, 1, eo.IsPmsCtrl)
				assert.NotNil(t, eo.Config)
				assert.NotNil(t, eo.Config.Input)
			},
		},
		{
			name: "invalid agent config json - first level",
			po: &dapo.ReleasePO{
				ID:          "1",
				AgentID:     "agent-1",
				AgentConfig: `{invalid json}`,
			},
			wantErr: true,
		},
		{
			name: "invalid config json - second level",
			po: &dapo.ReleasePO{
				ID:          "1",
				AgentID:     "agent-1",
				AgentConfig: `{"id":"1","key":"test-agent","name":"Test Agent","config":"{invalid json}"}`,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			eo, err := ReleaseDAConfEoSimple(ctx, tt.po)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				if tt.checkEo != nil {
					tt.checkEo(t, eo)
				}
			}
		})
	}
}

func TestReleaseDAConfEoSimple_ReleasePOFields(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	isAPIAgent := 1
	isPmsCtrl := 1

	releasePo := &dapo.ReleasePO{
		ID:           "release-1",
		AgentID:      "agent-1",
		AgentVersion: "1.0.0",
		AgentDesc:    "Test agent description",
		IsAPIAgent:   &isAPIAgent,
		IsPmsCtrl:    &isPmsCtrl,
		AgentConfig: `{
			"id": "1",
			"key": "test-agent",
			"name": "Test Agent",
			"config": "{\"input\":{}}"
		}`,
	}

	eo, err := ReleaseDAConfEoSimple(ctx, releasePo)
	require.NoError(t, err)

	assert.NotNil(t, eo)
	assert.Equal(t, "release-1", eo.ID)
	assert.Equal(t, "agent-1", eo.AgentID)
	assert.Equal(t, "1.0.0", eo.AgentVersion)
	assert.Equal(t, "Test agent description", eo.AgentDesc)
	assert.Equal(t, 1, eo.IsPmsCtrl)
}

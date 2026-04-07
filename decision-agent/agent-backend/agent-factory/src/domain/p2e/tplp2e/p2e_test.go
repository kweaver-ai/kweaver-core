package tplp2e

import (
	"context"
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/entity/daconfeo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/daenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgentTplListEo(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	builtInYes := cdaenum.BuiltInYes
	publishedAt := int64(1000000)
	publishedBy := "publisher"

	tests := []struct {
		name    string
		po      *dapo.DataAgentTplPo
		wantErr bool
		checkEo func(t *testing.T, eo *daconfeo.DataAgentTplListEo)
	}{
		{
			name: "valid template po",
			po: &dapo.DataAgentTplPo{
				ID:         1,
				Name:       "Test Template",
				Key:        "test-template",
				ProductKey: "test-product",
				IsBuiltIn:  &builtInYes,
				CreatedBy:  "user-1",
				UpdatedBy:  "user-2",
				Status:     cdaenum.StatusPublished,
			},
			wantErr: false,
			checkEo: func(t *testing.T, eo *daconfeo.DataAgentTplListEo) {
				assert.NotNil(t, eo)
				assert.Equal(t, int64(1), eo.ID)
				assert.Equal(t, "Test Template", eo.Name)
				assert.Equal(t, "test-template", eo.Key)
				assert.Equal(t, "test-product", eo.ProductKey)
				assert.Equal(t, "user-1", eo.CreatedBy)
				assert.Equal(t, "user-2", eo.UpdatedBy)
			},
		},
		{
			name: "minimal template po",
			po: &dapo.DataAgentTplPo{
				ID:   2,
				Name: "Minimal Template",
				Key:  "minimal-template",
			},
			wantErr: false,
			checkEo: func(t *testing.T, eo *daconfeo.DataAgentTplListEo) {
				assert.NotNil(t, eo)
				assert.Equal(t, int64(2), eo.ID)
				assert.Equal(t, "Minimal Template", eo.Name)
			},
		},
		{
			name: "template with all fields",
			po: &dapo.DataAgentTplPo{
				ID:          3,
				Name:        "Full Template",
				Key:         "full-template",
				Profile:     strPtr("Full description"),
				ProductKey:  "product-key",
				Avatar:      "📝",
				AvatarType:  cdaenum.AvatarTypeBuiltIn,
				IsBuiltIn:   &builtInYes,
				CreatedBy:   "creator",
				UpdatedBy:   "updater",
				PublishedBy: &publishedBy,
				PublishedAt: &publishedAt,
				Status:      cdaenum.StatusPublished,
				CreatedType: daenum.AgentTplCreatedTypeCopyFromAgent,
				CreateFrom:  "from-test",
			},
			wantErr: false,
			checkEo: func(t *testing.T, eo *daconfeo.DataAgentTplListEo) {
				assert.NotNil(t, eo)
				assert.Equal(t, int64(3), eo.ID)
				assert.Equal(t, "Full Template", eo.Name)
				assert.Equal(t, "📝", eo.Avatar)
				assert.Equal(t, "creator", eo.CreatedBy)
				assert.Equal(t, "updater", eo.UpdatedBy)
				assert.NotNil(t, eo.PublishedBy)
				assert.Equal(t, "publisher", *eo.PublishedBy)
				assert.Equal(t, int64(1000000), *eo.PublishedAt)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			eo, err := AgentTplListEo(ctx, tt.po)
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

func TestAgentTplListEo_EmptyPO(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	po := &dapo.DataAgentTplPo{}

	eo, err := AgentTplListEo(ctx, po)
	require.NoError(t, err)
	assert.NotNil(t, eo)
}

// Helper functions
func strPtr(s string) *string {
	return &s
}

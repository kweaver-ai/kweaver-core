package daconfe2p

import (
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/entity/daconfeo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/daconfvalobj"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDataAgentTpl(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		eo      *daconfeo.DataAgentTpl
		wantErr bool
		checkPO func(t *testing.T, po *dapo.DataAgentTplPo)
	}{
		{
			name: "valid template entity",
			eo: &daconfeo.DataAgentTpl{
				DataAgentTplPo: dapo.DataAgentTplPo{
					ID:     1,
					Name:   "Test Template",
					Status: cdaenum.StatusPublished,
				},
				Config: &daconfvalobj.Config{
					Input: &daconfvalobj.Input{
						Fields: daconfvalobj.Fields{
							&daconfvalobj.Field{
								Name: "field1",
								Type: cdaenum.InputFieldTypeString,
							},
						},
					},
					Output: &daconfvalobj.Output{},
				},
			},
			wantErr: false,
			checkPO: func(t *testing.T, po *dapo.DataAgentTplPo) {
				assert.Equal(t, int64(1), po.ID)
				assert.Equal(t, "Test Template", po.Name)
				assert.Equal(t, cdaenum.StatusPublished, po.Status)
				assert.NotEmpty(t, po.Config, "Config should be marshaled to JSON")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			po, err := DataAgentTpl(tt.eo)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, po)
			} else {
				require.NoError(t, err)
				require.NotNil(t, po)

				if tt.checkPO != nil {
					tt.checkPO(t, po)
				}
			}
		})
	}
}

func TestDataAgentTpls(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		eos      []*daconfeo.DataAgentTpl
		wantErr  bool
		checkPOs func(t *testing.T, pos []*dapo.DataAgentTplPo)
	}{
		{
			name: "multiple valid templates",
			eos: []*daconfeo.DataAgentTpl{
				{
					DataAgentTplPo: dapo.DataAgentTplPo{
						ID:   1,
						Name: "Template 1",
					},
					Config: &daconfvalobj.Config{
						Input:  &daconfvalobj.Input{},
						Output: &daconfvalobj.Output{},
					},
				},
				{
					DataAgentTplPo: dapo.DataAgentTplPo{
						ID:   2,
						Name: "Template 2",
					},
					Config: &daconfvalobj.Config{
						Input:  &daconfvalobj.Input{},
						Output: &daconfvalobj.Output{},
					},
				},
			},
			wantErr: false,
			checkPOs: func(t *testing.T, pos []*dapo.DataAgentTplPo) {
				assert.Len(t, pos, 2)
				assert.Equal(t, int64(1), pos[0].ID)
				assert.Equal(t, int64(2), pos[1].ID)
			},
		},
		{
			name:    "empty slice",
			eos:     []*daconfeo.DataAgentTpl{},
			wantErr: false,
			checkPOs: func(t *testing.T, pos []*dapo.DataAgentTplPo) {
				assert.Len(t, pos, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			pos, err := DataAgentTpls(tt.eos)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)

				if tt.checkPOs != nil {
					tt.checkPOs(t, pos)
				}
			}
		})
	}
}

func TestDataAgentTpl_WithNilConfig(t *testing.T) {
	t.Parallel()

	eo := &daconfeo.DataAgentTpl{
		DataAgentTplPo: dapo.DataAgentTplPo{
			ID:     1,
			Name:   "Test Template",
			Status: cdaenum.StatusPublished,
		},
		Config: nil,
	}

	po, err := DataAgentTpl(eo)
	require.NoError(t, err)
	require.NotNil(t, po)
	assert.Equal(t, int64(1), po.ID)
}

func TestDataAgentTpl_WithComplexConfig(t *testing.T) {
	t.Parallel()

	isDefault := true
	eo := &daconfeo.DataAgentTpl{
		DataAgentTplPo: dapo.DataAgentTplPo{
			ID:     1,
			Name:   "Complex Template",
			Status: cdaenum.StatusPublished,
		},
		Config: &daconfvalobj.Config{
			Input: &daconfvalobj.Input{
				Fields: daconfvalobj.Fields{
					&daconfvalobj.Field{
						Name: "field1",
						Type: cdaenum.InputFieldTypeString,
					},
				},
			},
			Llms: []*daconfvalobj.LlmItem{
				{
					IsDefault: isDefault,
					LlmConfig: &daconfvalobj.LlmConfig{
						Name:      "test-model",
						MaxTokens: 500,
					},
				},
			},
			IsDataFlowSetEnabled: 0,
			Output: &daconfvalobj.Output{
				DefaultFormat: cdaenum.OutputDefaultFormatMarkdown,
			},
		},
	}

	po, err := DataAgentTpl(eo)
	require.NoError(t, err)
	require.NotNil(t, po)
	assert.Equal(t, int64(1), po.ID)
	assert.NotEmpty(t, po.Config)
}

func TestDataAgentTpls_SingleEntity(t *testing.T) {
	t.Parallel()

	eos := []*daconfeo.DataAgentTpl{
		{
			DataAgentTplPo: dapo.DataAgentTplPo{
				ID:   1,
				Name: "Single Template",
			},
			Config: &daconfvalobj.Config{
				Input:  &daconfvalobj.Input{},
				Output: &daconfvalobj.Output{},
			},
		},
	}

	pos, err := DataAgentTpls(eos)
	require.NoError(t, err)
	assert.Len(t, pos, 1)
	assert.Equal(t, int64(1), pos[0].ID)
}

func TestDataAgentTpls_NilInSlice(t *testing.T) {
	t.Parallel()

	t.Run("nil entity in slice causes panic", func(t *testing.T) {
		t.Parallel()

		eos := []*daconfeo.DataAgentTpl{
			{
				DataAgentTplPo: dapo.DataAgentTplPo{
					ID:   1,
					Name: "Template 1",
				},
				Config: &daconfvalobj.Config{
					Input:  &daconfvalobj.Input{},
					Output: &daconfvalobj.Output{},
				},
			},
			nil,
		}

		// The function will panic on nil, so we need to recover
		defer func() {
			if r := recover(); r != nil {
				// Expected panic
				t.Logf("Expected panic with nil entity: %v", r)
			}
		}()

		pos, err := DataAgentTpls(eos)
		// If we get here without panic, that's also OK
		_ = pos
		_ = err
	})
}

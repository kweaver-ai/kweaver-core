package daconfp2e

import (
	"context"
	"os"
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/locale"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/entity/daconfeo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/daenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/cmp/umcmp/umtypes"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper/cenvhelper"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/ihttpaccess/iumacc/httpaccmock"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestMain(m *testing.M) {
	// Setup environment for tests
	os.Setenv("SERVICE_NAME", "AGENT_FACTORY")
	// Note: Do NOT set AGENT_FACTORY_LOCAL_DEV=true
	// We only test non-local-dev (production) mode to avoid environment variable race conditions
	os.Setenv("I18N_MODE_UT", "true")

	// Re-init cenvhelper so SERVICE_NAME takes effect
	cenvhelper.InitEnvForTest()

	// Initialize locale (only once)
	locale.Register()

	// Run tests
	code := m.Run()
	os.Exit(code)
}

func TestDataAgent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	builtInYes := cdaenum.BuiltInYes

	tests := []struct {
		name    string
		po      *dapo.DataAgentPo
		wantErr bool
		checkEo func(t *testing.T, eo *daconfeo.DataAgent)
	}{
		{
			name: "valid po with config",
			po: &dapo.DataAgentPo{
				ID:         "1",
				Key:        "test-agent",
				Name:       "Test Agent",
				ProductKey: "test-product",
				Config:     `{"input":{"fields":[{"name":"field1","type":"text"}]}}`,
			},
			wantErr: false,
			checkEo: func(t *testing.T, eo *daconfeo.DataAgent) {
				assert.NotNil(t, eo)
				assert.Equal(t, "1", eo.ID)
				assert.Equal(t, "test-agent", eo.Key)
				assert.Equal(t, "Test Agent", eo.Name)
				assert.NotNil(t, eo.Config)
				assert.NotNil(t, eo.Config.Input)
			},
		},
		{
			name: "valid po with empty config",
			po: &dapo.DataAgentPo{
				ID:         "1",
				Key:        "test-agent",
				Name:       "Test Agent",
				ProductKey: "test-product",
				Config:     "",
			},
			wantErr: false,
			checkEo: func(t *testing.T, eo *daconfeo.DataAgent) {
				assert.NotNil(t, eo)
				assert.NotNil(t, eo.Config)
			},
		},
		{
			name: "valid po with full config",
			po: &dapo.DataAgentPo{
				ID:         "1",
				Key:        "test-agent",
				Name:       "Test Agent",
				ProductKey: "test-product",
				CreatedBy:  "user-1",
				UpdatedBy:  "user-2",
				Config: `{
					"input": {
						"fields": [
							{"name": "question", "type": "text"}
						]
					},
					"output": {
						"output_1": {
							"name": "answer",
							"type": "text"
						}
					},
					"llms": [],
					"memory": {"enabled": false}
				}`,
			},
			wantErr: false,
			checkEo: func(t *testing.T, eo *daconfeo.DataAgent) {
				assert.NotNil(t, eo)
				assert.Equal(t, "user-1", eo.CreatedBy)
				assert.Equal(t, "user-2", eo.UpdatedBy)
				assert.NotNil(t, eo.Config)
				assert.NotNil(t, eo.Config.Input)
				assert.NotNil(t, eo.Config.Output)
			},
		},
		{
			name: "valid po with all fields",
			po: &dapo.DataAgentPo{
				ID:          "1",
				Key:         "test-agent",
				Name:        "Test Agent",
				Profile:     strPtr("test profile"),
				ProductKey:  "test-product",
				AvatarType:  cdaenum.AvatarTypeBuiltIn,
				Avatar:      "🤖",
				Status:      cdaenum.StatusPublished,
				IsBuiltIn:   &builtInYes,
				CreatedAt:   100,
				UpdatedAt:   200,
				CreatedBy:   "user-1",
				UpdatedBy:   "user-2",
				Config:      `{"input":{"fields":[{"name":"field1","type":"text"}]}}`,
				CreatedType: daenum.AgentCreatedTypeCreate,
				CreateFrom:  "from-test",
			},
			wantErr: false,
			checkEo: func(t *testing.T, eo *daconfeo.DataAgent) {
				assert.NotNil(t, eo)
				assert.Equal(t, "test-agent", eo.Key)
				assert.Equal(t, "Test Agent", eo.Name)
				assert.Equal(t, "test profile", *eo.Profile)
				assert.Equal(t, cdaenum.AvatarTypeBuiltIn, eo.AvatarType)
				assert.Equal(t, "🤖", eo.Avatar)
				assert.Equal(t, cdaenum.StatusPublished, eo.Status)
				assert.True(t, eo.IsBuiltInBool())
			},
		},
		{
			name: "invalid config json",
			po: &dapo.DataAgentPo{
				ID:     "1",
				Key:    "test-agent",
				Name:   "Test Agent",
				Config: `{invalid json}`,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			eo, err := DataAgent(ctx, tt.po)
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

func TestDataAgentSimple(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name    string
		po      *dapo.DataAgentPo
		wantErr bool
		checkEo func(t *testing.T, eo *daconfeo.DataAgent)
	}{
		{
			name: "valid po with config",
			po: &dapo.DataAgentPo{
				ID:     "1",
				Key:    "test-agent",
				Name:   "Test Agent",
				Config: `{"input":{"fields":[{"name":"field1","type":"text"}]}}`,
			},
			wantErr: false,
			checkEo: func(t *testing.T, eo *daconfeo.DataAgent) {
				assert.NotNil(t, eo)
				assert.Equal(t, "1", eo.ID)
				assert.Equal(t, "test-agent", eo.Key)
				assert.NotNil(t, eo.Config)
				assert.NotNil(t, eo.Config.Input)
			},
		},
		{
			name: "valid po with empty config",
			po: &dapo.DataAgentPo{
				ID:     "1",
				Key:    "test-agent",
				Name:   "Test Agent",
				Config: "",
			},
			wantErr: false,
			checkEo: func(t *testing.T, eo *daconfeo.DataAgent) {
				assert.NotNil(t, eo)
				assert.NotNil(t, eo.Config)
			},
		},
		{
			name: "invalid config json",
			po: &dapo.DataAgentPo{
				ID:     "1",
				Key:    "test-agent",
				Name:   "Test Agent",
				Config: `{invalid json}`,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			eo, err := DataAgentSimple(ctx, tt.po)
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

func TestDataAgent_Equals_DataAgentSimple(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create a valid PO
	po := &dapo.DataAgentPo{
		ID:     "1",
		Key:    "test-agent",
		Name:   "Test Agent",
		Config: `{"input":{"fields":[{"name":"question","type":"text"}]}}`,
	}

	// Test that both functions produce the same result
	eo1, err1 := DataAgent(ctx, po)
	eo2, err2 := DataAgentSimple(ctx, po)

	require.NoError(t, err1)
	require.NoError(t, err2)

	assert.Equal(t, eo1.ID, eo2.ID)
	assert.Equal(t, eo1.Key, eo2.Key)
	assert.Equal(t, eo1.Name, eo2.Name)
}

// Helper function
func strPtr(s string) *string {
	return &s
}

func TestDataAgents_NonLocalDevMode(t *testing.T) {
	ctx := context.WithValue(context.Background(), cenum.VisitLangCtxKey.String(), rest.SimplifiedChinese) //nolint:staticcheck

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)
	mockUmHttp := httpaccmock.NewMockUmHttpAcc(ctrl)

	pos := []*dapo.DataAgentPo{
		{ID: "1", Key: "test-agent", Name: "Test Agent", ProductKey: "test-product", CreatedBy: "user1", UpdatedBy: "user2"},
	}

	// Expect GetOsnNames to be called in non-local dev mode
	osnInfoMap := umtypes.NewOsnInfoMapS()
	osnInfoMap.UserNameMap["user1"] = "Real User 1"
	osnInfoMap.UserNameMap["user2"] = "Real User 2"
	mockUmHttp.EXPECT().GetOsnNames(ctx, gomock.Any()).Return(osnInfoMap, nil)

	productMap := map[string]string{"test-product": "Test Product"}
	mockProductRepo.EXPECT().GetByNameMapByKeys(ctx, []string{"test-product"}).Return(productMap, nil)

	eos, err := DataAgents(ctx, pos, mockProductRepo, mockUmHttp)

	assert.NoError(t, err)
	assert.NotNil(t, eos)
	assert.Len(t, eos, 1)
	// In non-local dev mode, real user names should be used
	assert.Equal(t, "Real User 1", eos[0].CreatedByName)
	assert.Equal(t, "Real User 2", eos[0].UpdatedByName)
	assert.Equal(t, "Test Product", eos[0].ProductName)
}

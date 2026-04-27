package publishedp2e

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/locale"
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

func TestPublishedAgent_WithoutUnmarshalConfig(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	po := &dapo.PublishedJoinPo{
		ReleasePartPo: dapo.ReleasePartPo{
			ReleaseID:   "release-1",
			PublishedBy: "user1",
			Version:     "1.0",
		},
		DataAgentPo: dapo.DataAgentPo{
			ID:   "agent1",
			Name: "Test Agent",
			Key:  "test-agent",
		},
	}

	eo, err := PublishedAgent(ctx, po, false)

	require.NoError(t, err)
	assert.NotNil(t, eo)
	assert.Equal(t, "release-1", eo.ReleaseID)
	assert.Equal(t, "user1", eo.PublishedBy)
	assert.Equal(t, "agent1", eo.ID)
	assert.Equal(t, "Test Agent", eo.Name)
	// Config is nil when isUnmarshalConfig is false and po.Config is empty
	assert.Nil(t, eo.Config)
}

func TestPublishedAgent_WithUnmarshalConfig_NoConfig(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	po := &dapo.PublishedJoinPo{
		ReleasePartPo: dapo.ReleasePartPo{
			ReleaseID:   "release-1",
			PublishedBy: "user1",
			Version:     "1.0",
			// No AgentConfig provided
		},
		DataAgentPo: dapo.DataAgentPo{
			ID:   "agent1",
			Name: "Test Agent",
			Key:  "test-agent",
		},
	}

	eo, err := PublishedAgent(ctx, po, true)

	require.NoError(t, err)
	assert.NotNil(t, eo)
	// Config remains nil when there's no config to unmarshal
	assert.Nil(t, eo.Config)
}

func TestPublishedAgent_WithProfile(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	profile := "test profile"
	po := &dapo.PublishedJoinPo{
		ReleasePartPo: dapo.ReleasePartPo{
			ReleaseID:   "release-1",
			PublishedBy: "user1",
			Version:     "1.0",
		},
		DataAgentPo: dapo.DataAgentPo{
			ID:      "agent1",
			Name:    "Test Agent",
			Key:     "test-agent",
			Profile: &profile,
		},
	}

	eo, err := PublishedAgent(ctx, po, false)

	require.NoError(t, err)
	assert.NotNil(t, eo)
	assert.NotNil(t, eo.Profile)
	assert.Equal(t, "test profile", *eo.Profile)
}

func TestPublishedTplListEo_Simple(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	po := &dapo.PublishedTplPo{
		ID:          1,
		Key:         "test-tpl",
		Name:        "Test Template",
		PublishedBy: "user1",
	}

	eo, err := PublishedTplListEo(ctx, po)

	require.NoError(t, err)
	assert.NotNil(t, eo)
	assert.Equal(t, int64(1), eo.ID)
	assert.Equal(t, "test-tpl", eo.Key)
	assert.Equal(t, "Test Template", eo.Name)
	assert.Equal(t, "user1", eo.PublishedBy)
}

func TestPublishedTpl_Simple(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)

	po := &dapo.PublishedTplPo{
		ID:          1,
		Key:         "test-tpl",
		Name:        "Test Template",
		ProductKey:  "test-product",
		PublishedBy: "user1",
	}

	// Return an empty product po to simulate not found
	mockProductRepo.EXPECT().GetByKey(ctx, "test-product").Return(&dapo.ProductPo{}, nil)

	eo, err := PublishedTpl(ctx, po, mockProductRepo)

	require.NoError(t, err)
	assert.NotNil(t, eo)
	assert.Equal(t, int64(1), eo.ID)
	assert.Equal(t, "test-tpl", eo.Key)
	assert.Equal(t, "Test Template", eo.Name)
	assert.NotNil(t, eo.Config)
}

func TestPublishedTpl_WithProduct(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)

	po := &dapo.PublishedTplPo{
		ID:          1,
		Key:         "test-tpl",
		Name:        "Test Template",
		ProductKey:  "test-product",
		PublishedBy: "user1",
	}

	productPo := &dapo.ProductPo{
		Key:  "test-product",
		Name: "Test Product",
	}

	mockProductRepo.EXPECT().GetByKey(ctx, "test-product").Return(productPo, nil)

	eo, err := PublishedTpl(ctx, po, mockProductRepo)

	require.NoError(t, err)
	assert.NotNil(t, eo)
	assert.Equal(t, "Test Product", eo.ProductName)
}

func TestPublishedTpl_WithConfig(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)

	// Use a minimal valid config JSON
	configJSON := `{"input":{"fields":[]},"output":{}}`
	po := &dapo.PublishedTplPo{
		ID:         1,
		Key:        "test-tpl",
		Name:       "Test Template",
		ProductKey: "test-product",
		Config:     configJSON,
	}

	// Return an empty product po
	mockProductRepo.EXPECT().GetByKey(ctx, "test-product").Return(&dapo.ProductPo{}, nil)

	eo, err := PublishedTpl(ctx, po, mockProductRepo)

	require.NoError(t, err)
	assert.NotNil(t, eo)
	assert.NotNil(t, eo.Config)
	assert.NotNil(t, eo.Config.Input)
}

func TestPublishedTpl_InvalidConfig(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)

	// Use completely invalid JSON that cannot be parsed
	po := &dapo.PublishedTplPo{
		ID:         1,
		Key:        "test-tpl",
		Name:       "Test Template",
		ProductKey: "test-product",
		Config:     `not json at all`,
	}

	// Config is unmarshaled BEFORE product lookup, so we don't need to expect GetByKey call
	_, err := PublishedTpl(ctx, po, mockProductRepo)

	// The function returns an error for invalid config
	assert.Error(t, err)
}

func TestPublishedTpl_NoProductKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)

	po := &dapo.PublishedTplPo{
		ID:         1,
		Key:        "test-tpl",
		Name:       "Test Template",
		ProductKey: "", // Empty product key
	}

	// Should not call GetByKey when ProductKey is empty
	eo, err := PublishedTpl(ctx, po, mockProductRepo)

	require.NoError(t, err)
	assert.NotNil(t, eo)
	assert.Equal(t, int64(1), eo.ID)
}

func TestPublishedAgents_NonLocalDevMode(t *testing.T) {
	ctx := context.WithValue(context.Background(), cenum.VisitLangCtxKey.String(), rest.SimplifiedChinese) //nolint:staticcheck // SA1029

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUmHttp := httpaccmock.NewMockUmHttpAcc(ctrl)

	pos := []*dapo.PublishedJoinPo{
		{
			ReleasePartPo: dapo.ReleasePartPo{
				ReleaseID:   "release-1",
				PublishedBy: "user1",
				Version:     "1.0",
			},
			DataAgentPo: dapo.DataAgentPo{
				ID:   "agent1",
				Name: "Test Agent",
				Key:  "test-agent",
			},
		},
	}

	// Expect GetOsnNames to be called in non-local dev mode
	osnInfoMap := umtypes.NewOsnInfoMapS()
	osnInfoMap.UserNameMap["user1"] = "Real User 1"
	mockUmHttp.EXPECT().GetOsnNames(ctx, gomock.Any()).Return(osnInfoMap, nil)

	eos, err := PublishedAgents(ctx, pos, mockUmHttp, false)

	assert.NoError(t, err)
	assert.NotNil(t, eos)
	assert.Len(t, eos, 1)
	// In non-local dev mode, real user names should be used
	assert.Equal(t, "Real User 1", eos[0].PublishedByName)
}

func TestPublishedTpl_ProductRepoError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)

	po := &dapo.PublishedTplPo{
		ID:         1,
		Key:        "test-tpl",
		Name:       "Test Template",
		ProductKey: "test-product",
	}

	// Simulate a non-not-found error from productRepo
	mockProductRepo.EXPECT().GetByKey(ctx, "test-product").Return(nil, errors.New("database error"))

	_, err := PublishedTpl(ctx, po, mockProductRepo)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "get product name error")
}

func TestPublishedTplListEos_NonLocalDevMode(t *testing.T) {
	ctx := context.WithValue(context.Background(), cenum.VisitLangCtxKey.String(), rest.SimplifiedChinese) //nolint:staticcheck // SA1029

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUmHttp := httpaccmock.NewMockUmHttpAcc(ctrl)

	pos := []*dapo.PublishedTplPo{
		{ID: 1, Key: "test-tpl", Name: "Test Template", PublishedBy: "user1"},
	}

	// Expect GetOsnNames to be called in non-local dev mode
	osnInfoMap := umtypes.NewOsnInfoMapS()
	osnInfoMap.UserNameMap["user1"] = "Real User 1"
	mockUmHttp.EXPECT().GetOsnNames(ctx, gomock.Any()).Return(osnInfoMap, nil)

	eos, err := PublishedTplListEos(ctx, pos, mockUmHttp)

	assert.NoError(t, err)
	assert.NotNil(t, eos)
	assert.Len(t, eos, 1)
	// In non-local dev mode, real user names should be used
	assert.Equal(t, "Real User 1", eos[0].PublishedByName)
}

func TestPublishedAgents_NonLocalDevModeError(t *testing.T) {
	ctx := context.WithValue(context.Background(), cenum.VisitLangCtxKey.String(), rest.SimplifiedChinese) //nolint:staticcheck // SA1029

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUmHttp := httpaccmock.NewMockUmHttpAcc(ctrl)

	pos := []*dapo.PublishedJoinPo{
		{
			ReleasePartPo: dapo.ReleasePartPo{
				ReleaseID:   "release-1",
				PublishedBy: "user1",
				Version:     "1.0",
			},
			DataAgentPo: dapo.DataAgentPo{
				ID:   "agent1",
				Name: "Test Agent",
				Key:  "test-agent",
			},
		},
	}

	// Expect GetOsnNames to return an error
	mockUmHttp.EXPECT().GetOsnNames(ctx, gomock.Any()).Return(nil, errors.New("network error"))

	eos, err := PublishedAgents(ctx, pos, mockUmHttp, false)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "network error")
	// eos may be non-nil but empty slice on error
	assert.Empty(t, eos)
}

func TestPublishedTplListEos_NonLocalDevModeError(t *testing.T) {
	ctx := context.WithValue(context.Background(), cenum.VisitLangCtxKey.String(), rest.SimplifiedChinese) //nolint:staticcheck // SA1029

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUmHttp := httpaccmock.NewMockUmHttpAcc(ctrl)

	pos := []*dapo.PublishedTplPo{
		{ID: 1, Key: "test-tpl", Name: "Test Template", PublishedBy: "user1"},
	}

	// Expect GetOsnNames to return an error
	mockUmHttp.EXPECT().GetOsnNames(ctx, gomock.Any()).Return(nil, errors.New("network error"))

	eos, err := PublishedTplListEos(ctx, pos, mockUmHttp)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "network error")
	// eos may be non-nil but empty slice on error
	assert.Empty(t, eos)
}

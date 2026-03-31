package tplp2e

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/locale"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/umcmp/umtypes"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/cenvhelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/ihttpaccess/iumacc/httpaccmock"
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

func TestDataAgentTpl_Simple(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)

	po := &dapo.DataAgentTplPo{
		ID:         1,
		Name:       "Test Template",
		Key:        "test-template",
		ProductKey: "test-product",
	}

	// Return an empty product po to simulate not found
	productPo := &dapo.ProductPo{}
	mockProductRepo.EXPECT().GetByKey(ctx, "test-product").Return(productPo, nil)

	eo, err := DataAgentTpl(ctx, po, mockProductRepo)
	require.NoError(t, err)
	assert.NotNil(t, eo)
	assert.Equal(t, int64(1), eo.ID)
	assert.Equal(t, "Test Template", eo.Name)
	assert.Equal(t, "test-template", eo.Key)
}

func TestDataAgentTpl_WithProduct(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)

	po := &dapo.DataAgentTplPo{
		ID:         1,
		Name:       "Test Template",
		Key:        "test-template",
		ProductKey: "test-product",
	}

	productPo := &dapo.ProductPo{
		Key:  "test-product",
		Name: "Test Product",
	}

	mockProductRepo.EXPECT().GetByKey(ctx, "test-product").Return(productPo, nil)

	eo, err := DataAgentTpl(ctx, po, mockProductRepo)
	require.NoError(t, err)
	assert.NotNil(t, eo)
	assert.Equal(t, "Test Product", eo.ProductName)
}

func TestDataAgentTpl_WithConfig(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)

	configJSON := `{"profile":"test profile"}`
	po := &dapo.DataAgentTplPo{
		ID:         1,
		Name:       "Test Template",
		Key:        "test-template",
		ProductKey: "test-product",
		Config:     configJSON,
	}

	// Return an empty product po
	productPo := &dapo.ProductPo{}
	mockProductRepo.EXPECT().GetByKey(ctx, "test-product").Return(productPo, nil)

	eo, err := DataAgentTpl(ctx, po, mockProductRepo)
	require.NoError(t, err)
	assert.NotNil(t, eo)
	assert.NotNil(t, eo.Config)
}

func TestDataAgentTpl_InvalidConfig(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)

	invalidJSON := `{invalid json`
	po := &dapo.DataAgentTplPo{
		ID:         1,
		Name:       "Test Template",
		Key:        "test-template",
		ProductKey: "test-product",
		Config:     invalidJSON,
	}

	// Config is unmarshaled BEFORE product lookup, so we don't need to expect GetByKey call
	// because the function will return early due to invalid JSON

	eo, err := DataAgentTpl(ctx, po, mockProductRepo)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "DataAgentTpl unmarshal config error")

	_ = eo // EO may be non-nil even on error
}

func TestDataAgentTpl_ProductRepoError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)

	po := &dapo.DataAgentTplPo{
		ID:         1,
		Name:       "Test Template",
		Key:        "test-template",
		ProductKey: "test-product",
	}

	// Return a non-nil error that's not "sql not found"
	mockProductRepo.EXPECT().GetByKey(ctx, "test-product").Return(nil, errors.New("database connection failed"))

	eo, err := DataAgentTpl(ctx, po, mockProductRepo)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "get product name error")

	_ = eo // EO may be non-nil even on error
}

func TestDataAgentTpl_NoProductKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)

	po := &dapo.DataAgentTplPo{
		ID:         1,
		Name:       "Test Template",
		Key:        "test-template",
		ProductKey: "", // Empty product key
	}

	// Should not call GetByKey when ProductKey is empty
	eo, err := DataAgentTpl(ctx, po, mockProductRepo)
	require.NoError(t, err)
	assert.NotNil(t, eo)
	assert.Equal(t, int64(1), eo.ID)
}

func TestAgentTplListEos_NonLocalDevMode(t *testing.T) {
	ctx := context.WithValue(context.Background(), cenum.VisitLangCtxKey.String(), rest.SimplifiedChinese) //nolint:staticcheck // SA1029

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUmHttp := httpaccmock.NewMockUmHttpAcc(ctrl)

	pos := []*dapo.DataAgentTplPo{
		{ID: 1, Name: "Template 1", Key: "tpl-1", CreatedBy: "user1", UpdatedBy: "user2"},
	}

	// Expect GetOsnNames to be called in non-local dev mode
	osnInfoMap := umtypes.NewOsnInfoMapS()
	osnInfoMap.UserNameMap["user1"] = "Real User 1"
	osnInfoMap.UserNameMap["user2"] = "Real User 2"
	mockUmHttp.EXPECT().GetOsnNames(ctx, gomock.Any()).Return(osnInfoMap, nil)

	eos, err := AgentTplListEos(ctx, pos, mockUmHttp)

	assert.NoError(t, err)
	assert.NotNil(t, eos)
	assert.Len(t, eos, 1)
	// In non-local dev mode, real user names should be used
	assert.Equal(t, "Real User 1", eos[0].CreatedByName)
	assert.Equal(t, "Real User 2", eos[0].UpdatedByName)
}

func TestAgentTplListEos_NonLocalDevModeError(t *testing.T) {
	ctx := context.WithValue(context.Background(), cenum.VisitLangCtxKey.String(), rest.SimplifiedChinese) //nolint:staticcheck // SA1029

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUmHttp := httpaccmock.NewMockUmHttpAcc(ctrl)

	pos := []*dapo.DataAgentTplPo{
		{ID: 1, Name: "Template 1", Key: "tpl-1", CreatedBy: "user1", UpdatedBy: "user2"},
	}

	// Expect GetOsnNames to return an error
	mockUmHttp.EXPECT().GetOsnNames(ctx, gomock.Any()).Return(nil, errors.New("network error"))

	eos, err := AgentTplListEos(ctx, pos, mockUmHttp)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "network error")
	// eos may be non-nil but empty slice on error
	assert.Empty(t, eos)
}

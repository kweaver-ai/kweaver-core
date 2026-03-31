package personalspacep2e

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

func TestAgentsListForPersonalSpace_Simple(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	po := &dapo.DataAgentPo{
		ID:         "agent1",
		Name:       "Test Agent",
		Key:        "test-agent",
		ProductKey: "product1",
		CreatedBy:  "user1",
		UpdatedBy:  "user2",
	}

	eo, err := AgentsListForPersonalSpace(ctx, po)
	require.NoError(t, err)
	assert.NotNil(t, eo)
	assert.Equal(t, "agent1", eo.ID)
	assert.Equal(t, "Test Agent", eo.Name)
	assert.Equal(t, "test-agent", eo.Key)
	assert.Equal(t, "product1", eo.ProductKey)
	assert.Equal(t, "user1", eo.CreatedBy)
	assert.Equal(t, "user2", eo.UpdatedBy)
}

func TestAgentsListForPersonalSpace_WithProfile(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	profile := "test profile"
	po := &dapo.DataAgentPo{
		ID:        "agent1",
		Name:      "Test Agent",
		Profile:   &profile,
		CreatedBy: "user1",
		UpdatedBy: "user1",
	}

	eo, err := AgentsListForPersonalSpace(ctx, po)
	require.NoError(t, err)
	assert.NotNil(t, eo)
	assert.Equal(t, "agent1", eo.ID)
	assert.NotNil(t, eo.Profile)
	assert.Equal(t, "test profile", *eo.Profile)
}

func TestAgentsListForPersonalSpace_EmptyPo(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	po := &dapo.DataAgentPo{}

	eo, err := AgentsListForPersonalSpace(ctx, po)
	require.NoError(t, err)
	assert.NotNil(t, eo)
}

func TestAgentsListForPersonalSpaces_NonLocalDevMode(t *testing.T) {
	ctx := context.WithValue(context.Background(), cenum.VisitLangCtxKey.String(), rest.SimplifiedChinese) //nolint:staticcheck

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUmHttp := httpaccmock.NewMockUmHttpAcc(ctrl)

	pos := []*dapo.DataAgentPo{
		{ID: "agent1", Name: "Test Agent", Key: "test-agent", ProductKey: "product1", CreatedBy: "user1", UpdatedBy: "user2"},
	}

	// Expect GetOsnNames to be called in non-local dev mode
	osnInfoMap := umtypes.NewOsnInfoMapS()
	osnInfoMap.UserNameMap["user1"] = "Real User 1"
	osnInfoMap.UserNameMap["user2"] = "Real User 2"
	mockUmHttp.EXPECT().GetOsnNames(ctx, gomock.Any()).Return(osnInfoMap, nil)

	eos, err := AgentsListForPersonalSpaces(ctx, pos, mockUmHttp)

	assert.NoError(t, err)
	assert.NotNil(t, eos)
	assert.Len(t, eos, 1)
	// In non-local dev mode, real user names should be used
	assert.Equal(t, "Real User 1", eos[0].CreatedByName)
	assert.Equal(t, "Real User 2", eos[0].UpdatedByName)
}

func TestAgentsListForPersonalSpaces_NonLocalDevModeError(t *testing.T) {
	ctx := context.WithValue(context.Background(), cenum.VisitLangCtxKey.String(), rest.SimplifiedChinese) //nolint:staticcheck

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUmHttp := httpaccmock.NewMockUmHttpAcc(ctrl)

	pos := []*dapo.DataAgentPo{
		{ID: "agent1", Name: "Test Agent", Key: "test-agent", ProductKey: "product1", CreatedBy: "user1", UpdatedBy: "user2"},
	}

	// Expect GetOsnNames to return an error
	mockUmHttp.EXPECT().GetOsnNames(ctx, gomock.Any()).Return(nil, errors.New("network error"))

	eos, err := AgentsListForPersonalSpaces(ctx, pos, mockUmHttp)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "network error")
	// eos may be non-nil but empty slice on error
	assert.Empty(t, eos)
}

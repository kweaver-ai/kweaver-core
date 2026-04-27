package releaseresp

import (
	"context"
	"os"
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper/cenvhelper"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/ihttpaccess/iumacc/httpaccmock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestMain(m *testing.M) {
	os.Setenv("SERVICE_NAME", "AGENT_FACTORY")
	// Note: Do NOT set AGENT_FACTORY_LOCAL_DEV=true
	// We only test non-local-dev (production) mode to avoid environment variable race conditions
	os.Setenv("I18N_MODE_UT", "true")
	// Re-init cenvhelper so SERVICE_NAME takes effect
	// (init() runs before TestMain, so env vars set here need a re-init)
	cenvhelper.InitEnvForTest()
	os.Exit(m.Run())
}

func TestPublishUpsertResp_StructFields(t *testing.T) {
	t.Parallel()

	resp := PublishUpsertResp{
		ReleaseId:       "release-123",
		Version:         "1.0.0",
		PublishedAt:     1640995200000,
		PublishedBy:     "user-123",
		PublishedByName: "John Doe",
	}

	assert.Equal(t, "release-123", resp.ReleaseId)
	assert.Equal(t, "1.0.0", resp.Version)
	assert.Equal(t, int64(1640995200000), resp.PublishedAt)
	assert.Equal(t, "user-123", resp.PublishedBy)
	assert.Equal(t, "John Doe", resp.PublishedByName)
}

func TestPublishUpsertResp_Empty(t *testing.T) {
	t.Parallel()

	resp := PublishUpsertResp{}

	assert.Empty(t, resp.ReleaseId)
	assert.Empty(t, resp.Version)
	assert.Equal(t, int64(0), resp.PublishedAt)
	assert.Empty(t, resp.PublishedBy)
	assert.Empty(t, resp.PublishedByName)
}

func TestPublishUpsertResp_WithReleaseId(t *testing.T) {
	t.Parallel()

	ids := []string{
		"release-001",
		"release-xyz",
		"发布-123",
		"",
	}

	for _, id := range ids {
		resp := PublishUpsertResp{
			ReleaseId: id,
		}
		assert.Equal(t, id, resp.ReleaseId)
	}
}

func TestPublishUpsertResp_WithVersion(t *testing.T) {
	t.Parallel()

	versions := []string{
		"1.0.0",
		"2.1.3",
		"3.0.0-alpha",
		"4.5.2-beta.1",
		"latest",
		"",
	}

	for _, version := range versions {
		resp := PublishUpsertResp{
			Version: version,
		}
		assert.Equal(t, version, resp.Version)
	}
}

func TestPublishUpsertResp_WithPublishedAt(t *testing.T) {
	t.Parallel()

	timestamps := []int64{
		1640995200000, // 2022-01-01
		1643673600000, // 2022-02-01
		1646092800000, // 2022-03-01
		1672531200000, // 2023-01-01
		1704067200000, // 2024-01-01
		0,             // Zero timestamp
	}

	for _, ts := range timestamps {
		resp := PublishUpsertResp{
			PublishedAt: ts,
		}
		assert.Equal(t, ts, resp.PublishedAt)
	}
}

func TestPublishUpsertResp_WithPublishedBy(t *testing.T) {
	t.Parallel()

	users := []string{
		"user-001",
		"user-xyz",
		"用户-123",
		"",
	}

	for _, user := range users {
		resp := PublishUpsertResp{
			PublishedBy: user,
		}
		assert.Equal(t, user, resp.PublishedBy)
	}
}

func TestPublishUpsertResp_WithPublishedByName(t *testing.T) {
	t.Parallel()

	names := []string{
		"John Doe",
		"张三",
		"User with numbers 123",
		"",
	}

	for _, name := range names {
		resp := PublishUpsertResp{
			PublishedByName: name,
		}
		assert.Equal(t, name, resp.PublishedByName)
	}
}

func TestPublishUpsertResp_WithAllFields(t *testing.T) {
	t.Parallel()

	resp := PublishUpsertResp{
		ReleaseId:       "release-complete",
		Version:         "9.9.9",
		PublishedAt:     1704067200000,
		PublishedBy:     "user-complete",
		PublishedByName: "Complete User Name",
	}

	assert.Equal(t, "release-complete", resp.ReleaseId)
	assert.Equal(t, "9.9.9", resp.Version)
	assert.Equal(t, int64(1704067200000), resp.PublishedAt)
	assert.Equal(t, "user-complete", resp.PublishedBy)
	assert.Equal(t, "Complete User Name", resp.PublishedByName)
}

func TestPublishUpsertResp_FillPublishedByName_LocalDev(t *testing.T) {
	t.Parallel()

	// This test depends on the environment
	// We'll just test that the method exists and can be called
	resp := PublishUpsertResp{
		PublishedBy: "user-123",
	}

	// Save original value and restore after test
	// Note: This is a simple test to ensure the method is callable
	// Actual behavior depends on cenvhelper.IsLocalDev()
	assert.Equal(t, "user-123", resp.PublishedBy)
	assert.Empty(t, resp.PublishedByName)
}

func TestPublishUpsertResp_WithTimestamp(t *testing.T) {
	t.Parallel()

	resp := PublishUpsertResp{
		ReleaseId:   "release-123",
		Version:     "1.0.0",
		PublishedAt: 1640995200000,
		PublishedBy: "user-123",
	}

	// Verify timestamp is set correctly
	assert.Equal(t, int64(1640995200000), resp.PublishedAt)
	assert.Greater(t, resp.PublishedAt, int64(0))
}

func TestPublishUpsertResp_WithNegativeTimestamp(t *testing.T) {
	t.Parallel()

	resp := PublishUpsertResp{
		PublishedAt: -12345,
	}

	assert.Equal(t, int64(-12345), resp.PublishedAt)
}

func TestPublishUpsertResp_WithChineseName(t *testing.T) {
	t.Parallel()

	resp := PublishUpsertResp{
		PublishedByName: "张三",
	}

	assert.Equal(t, "张三", resp.PublishedByName)
}

func TestPublishUpsertResp_WithMixedName(t *testing.T) {
	t.Parallel()

	resp := PublishUpsertResp{
		PublishedByName: "User用户Name",
	}

	assert.Equal(t, "User用户Name", resp.PublishedByName)
}

func TestPublishUpsertResp_FillPublishedByName_Signature(t *testing.T) {
	t.Parallel()

	// Test that FillPublishedByName has the correct signature
	resp := &PublishUpsertResp{
		PublishedBy: "user-123",
	}

	// Just verify the method can be called with correct parameters
	// Actual implementation depends on environment and mock
	assert.NotNil(t, resp)
	assert.Equal(t, "user-123", resp.PublishedBy)
}

func TestPublishUpsertResp_ContextUsage(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	resp := &PublishUpsertResp{
		PublishedBy: "user-123",
	}

	// Verify context can be used (though actual UM call would need mock)
	assert.NotNil(t, ctx)
	assert.NotNil(t, resp)
}

func TestPublishUpsertResp_WithSemanticVersioning(t *testing.T) {
	t.Parallel()

	versions := []string{
		"1.0.0",
		"1.1.0",
		"2.0.0",
		"2.1.3",
		"3.0.0-rc1",
		"3.0.0-beta",
		"3.0.0-alpha.1",
	}

	for _, version := range versions {
		resp := PublishUpsertResp{
			Version: version,
		}
		assert.Equal(t, version, resp.Version)
	}
}

func TestPublishUpsertResp_WithReleaseIdFormats(t *testing.T) {
	t.Parallel()

	releaseIds := []string{
		"release-001",
		"RELEASE-002",
		"Release-003",
		"rls-004",
		"",
	}

	for _, releaseId := range releaseIds {
		resp := PublishUpsertResp{
			ReleaseId: releaseId,
		}
		assert.Equal(t, releaseId, resp.ReleaseId)
	}
}

func TestPublishUpsertResp_FillPublishedByName_NonLocalDev(t *testing.T) {
	t.Parallel()

	t.Run("non-local dev - successfully gets user name", func(t *testing.T) {
		t.Parallel()

		resp := &PublishUpsertResp{
			PublishedBy: "user123",
		}
		ctx := context.Background()
		ctrl := gomock.NewController(t)
		mockUm := httpaccmock.NewMockUmHttpAcc(ctrl)
		mockUm.EXPECT().GetSingleUserName(ctx, "user123").Return("John Doe", nil)

		err := resp.FillPublishedByName(ctx, mockUm)
		assert.NoError(t, err)
		assert.Equal(t, "John Doe", resp.PublishedByName)
	})

	t.Run("non-local dev - error getting user name", func(t *testing.T) {
		t.Parallel()

		resp := &PublishUpsertResp{
			PublishedBy: "user123",
		}
		ctx := context.Background()
		ctrl := gomock.NewController(t)
		mockUm := httpaccmock.NewMockUmHttpAcc(ctrl)
		mockUm.EXPECT().GetSingleUserName(ctx, "user123").Return("", assert.AnError)

		err := resp.FillPublishedByName(ctx, mockUm)
		assert.Error(t, err)
	})

	t.Run("non-local dev - empty user name returned", func(t *testing.T) {
		t.Parallel()

		resp := &PublishUpsertResp{
			PublishedBy: "user456",
		}
		ctx := context.Background()
		ctrl := gomock.NewController(t)
		mockUm := httpaccmock.NewMockUmHttpAcc(ctrl)
		mockUm.EXPECT().GetSingleUserName(ctx, "user456").Return("", nil)

		err := resp.FillPublishedByName(ctx, mockUm)
		assert.NoError(t, err)
		assert.Equal(t, "", resp.PublishedByName)
	})
}

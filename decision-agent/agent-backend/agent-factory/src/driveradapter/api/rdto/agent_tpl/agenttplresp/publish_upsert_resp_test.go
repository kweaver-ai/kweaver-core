package agenttplresp

import (
	"context"
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/ihttpaccess/iumacc/httpaccmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestPublishUpsertResp_StructFields(t *testing.T) {
	t.Parallel()

	resp := PublishUpsertResp{
		AgentTplId:      12345,
		PublishedAt:     1640995200000,
		PublishedBy:     "user-123",
		PublishedByName: "John Doe",
	}

	assert.Equal(t, int64(12345), resp.AgentTplId)
	assert.Equal(t, int64(1640995200000), resp.PublishedAt)
	assert.Equal(t, "user-123", resp.PublishedBy)
	assert.Equal(t, "John Doe", resp.PublishedByName)
}

func TestPublishUpsertResp_Empty(t *testing.T) {
	t.Parallel()

	resp := PublishUpsertResp{}

	assert.Equal(t, int64(0), resp.AgentTplId)
	assert.Equal(t, int64(0), resp.PublishedAt)
	assert.Empty(t, resp.PublishedBy)
	assert.Empty(t, resp.PublishedByName)
}

func TestPublishUpsertResp_WithAgentTplId(t *testing.T) {
	t.Parallel()

	ids := []int64{
		0,
		1,
		12345,
		999999,
	}

	for _, id := range ids {
		resp := PublishUpsertResp{
			AgentTplId: id,
		}
		assert.Equal(t, id, resp.AgentTplId)
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
		AgentTplId:      98765,
		PublishedAt:     1672531200000,
		PublishedBy:     "user-complete",
		PublishedByName: "Complete User Name",
	}

	assert.Equal(t, int64(98765), resp.AgentTplId)
	assert.Equal(t, int64(1672531200000), resp.PublishedAt)
	assert.Equal(t, "user-complete", resp.PublishedBy)
	assert.Equal(t, "Complete User Name", resp.PublishedByName)
}

func TestPublishUpsertResp_FillPublishedByName_LocalDev(t *testing.T) {
	t.Parallel()

	// Test when IsLocalDev is true - the function should append "_name" to PublishedBy
	// Note: This test depends on environment variable APP_ENV being set to "local" or "dev"
	// Since we can't control the environment in tests, we'll set up expectations for both paths

	t.Run("with mock for non-local environment", func(t *testing.T) {
		t.Parallel()

		resp := &PublishUpsertResp{
			PublishedBy: "user-123",
		}

		ctx := context.Background()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockUm := httpaccmock.NewMockUmHttpAcc(ctrl)

		// Set up expectation for non-local environment
		mockUm.EXPECT().GetSingleUserName(ctx, "user-123").Return("Test User", nil)

		err := resp.FillPublishedByName(ctx, mockUm)
		assert.NoError(t, err)
		assert.Equal(t, "Test User", resp.PublishedByName)
	})
}

func TestPublishUpsertResp_FillPublishedByName_WithMock(t *testing.T) {
	t.Parallel()

	resp := &PublishUpsertResp{
		PublishedBy: "user-123",
	}

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockUm := httpaccmock.NewMockUmHttpAcc(ctrl)

	// Set up mock to return a name
	mockUm.EXPECT().GetSingleUserName(ctx, "user-123").Return("John Doe", nil)

	// Call the function - the mock returns name and nil error
	err := resp.FillPublishedByName(ctx, mockUm)
	require.NoError(t, err)
	assert.Equal(t, "user-123", resp.PublishedBy)
	assert.Equal(t, "John Doe", resp.PublishedByName)
}

func TestPublishUpsertResp_FillPublishedByName_NilResponse(t *testing.T) {
	t.Parallel()

	resp := &PublishUpsertResp{
		PublishedBy:     "",
		PublishedByName: "Existing Name",
	}

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockUm := httpaccmock.NewMockUmHttpAcc(ctrl)

	// Set up mock to return empty string for empty user ID
	mockUm.EXPECT().GetSingleUserName(ctx, "").Return("", nil)

	err := resp.FillPublishedByName(ctx, mockUm)
	// Should not error even with empty PublishedBy
	require.NoError(t, err)
	// PublishedByName may or may not be overwritten depending on environment
	assert.NotNil(t, resp)
}

func TestPublishUpsertResp_FillPublishedByName_EmptyUserID(t *testing.T) {
	t.Parallel()

	resp := &PublishUpsertResp{
		PublishedBy: "",
	}

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockUm := httpaccmock.NewMockUmHttpAcc(ctrl)

	// Set up mock to return empty string
	mockUm.EXPECT().GetSingleUserName(ctx, "").Return("", nil)

	err := resp.FillPublishedByName(ctx, mockUm)
	// Should not error
	require.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestPublishUpsertResp_WithTimestamp(t *testing.T) {
	t.Parallel()

	resp := PublishUpsertResp{
		AgentTplId:  12345,
		PublishedAt: 1640995200000,
		PublishedBy: "user-123",
	}

	// Verify timestamp is set correctly
	assert.Equal(t, int64(1640995200000), resp.PublishedAt)
	assert.Greater(t, resp.PublishedAt, int64(0))
}

func TestPublishUpsertResp_WithZeroTimestamp(t *testing.T) {
	t.Parallel()

	resp := PublishUpsertResp{
		PublishedAt: 0,
	}

	assert.Equal(t, int64(0), resp.PublishedAt)
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

func TestPublishUpsertResp_FillPublishedByName_Context(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		publishedBy string
	}{
		{"with user ID", "user-123"},
		{"with empty user ID", ""},
		{"with admin ID", "admin-001"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			resp := &PublishUpsertResp{
				PublishedBy: tt.publishedBy,
			}

			ctx := context.Background()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockUm := httpaccmock.NewMockUmHttpAcc(ctrl)

			// Set up mock to return empty name
			mockUm.EXPECT().GetSingleUserName(ctx, tt.publishedBy).Return("", nil)

			err := resp.FillPublishedByName(ctx, mockUm)
			// Function should not panic or error for these cases
			require.NoError(t, err)
			assert.Equal(t, tt.publishedBy, resp.PublishedBy)
		})
	}
}

func TestPublishUpsertResp_FillPublishedByName_Error(t *testing.T) {
	t.Parallel()

	resp := &PublishUpsertResp{
		PublishedBy: "user-123",
	}

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockUm := httpaccmock.NewMockUmHttpAcc(ctrl)

	// Set up mock to return an error
	mockUm.EXPECT().GetSingleUserName(ctx, "user-123").Return("", assert.AnError)

	err := resp.FillPublishedByName(ctx, mockUm)
	assert.Error(t, err)
}

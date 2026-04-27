package squaresvc

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestCheckAndGetID(t *testing.T) {
	t.Parallel()

	t.Run("agent exists by ID - returns same ID", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
		mockAgentRepo.EXPECT().ExistsByID(gomock.Any(), "agent-123").Return(true, nil)

		svc := &squareSvc{
			SvcBase:       service.NewSvcBase(),
			agentConfRepo: mockAgentRepo,
		}

		newAgentID, err := svc.CheckAndGetID(context.Background(), "agent-123")

		assert.NoError(t, err)
		assert.Equal(t, "agent-123", newAgentID)
	})

	t.Run("agent exists by key - returns agent ID from key lookup", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
		mockAgentRepo.EXPECT().ExistsByID(gomock.Any(), "agent-key").Return(false, nil)

		agentPo := &dapo.DataAgentPo{
			ID:  "agent-456",
			Key: "agent-key",
		}
		mockAgentRepo.EXPECT().GetByKey(gomock.Any(), "agent-key").Return(agentPo, nil)

		svc := &squareSvc{
			SvcBase:       service.NewSvcBase(),
			agentConfRepo: mockAgentRepo,
		}

		newAgentID, err := svc.CheckAndGetID(context.Background(), "agent-key")

		assert.NoError(t, err)
		assert.Equal(t, "agent-456", newAgentID)
	})

	t.Run("agent not found - returns 404 error", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
		mockAgentRepo.EXPECT().ExistsByID(gomock.Any(), "non-existent").Return(false, nil)
		mockAgentRepo.EXPECT().GetByKey(gomock.Any(), "non-existent").Return(nil, sql.ErrNoRows)

		svc := &squareSvc{
			SvcBase:       service.NewSvcBase(),
			agentConfRepo: mockAgentRepo,
		}

		newAgentID, err := svc.CheckAndGetID(context.Background(), "non-existent")

		assert.Error(t, err)
		assert.Empty(t, newAgentID)
		assert.Contains(t, err.Error(), "agent not found")
	})

	t.Run("repository error on ExistsByID - returns error", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
		expectedErr := errors.New("database error")
		mockAgentRepo.EXPECT().ExistsByID(gomock.Any(), "agent-123").Return(false, expectedErr)

		svc := &squareSvc{
			SvcBase:       service.NewSvcBase(),
			agentConfRepo: mockAgentRepo,
		}

		newAgentID, err := svc.CheckAndGetID(context.Background(), "agent-123")

		assert.Error(t, err)
		assert.Empty(t, newAgentID)
		assert.Contains(t, err.Error(), "svc.agentConfRepo.ExistsByID")
	})

	t.Run("repository error on GetByKey - returns error", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
		mockAgentRepo.EXPECT().ExistsByID(gomock.Any(), "agent-key").Return(false, nil)

		expectedErr := errors.New("database connection failed")
		mockAgentRepo.EXPECT().GetByKey(gomock.Any(), "agent-key").Return(nil, expectedErr)

		svc := &squareSvc{
			SvcBase:       service.NewSvcBase(),
			agentConfRepo: mockAgentRepo,
		}

		newAgentID, err := svc.CheckAndGetID(context.Background(), "agent-key")

		assert.Error(t, err)
		assert.Empty(t, newAgentID)
	})
}

func TestGetLatestVersion(t *testing.T) {
	t.Parallel()

	t.Run("release exists - returns release version", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockReleaseHistoryRepo := idbaccessmock.NewMockIReleaseHistoryRepo(ctrl)

		releasePo := &dapo.ReleasePO{
			ID:           "release-123",
			AgentID:      "agent-123",
			AgentVersion: "v1.0.0",
		}

		svc := &squareSvc{
			SvcBase:            service.NewSvcBase(),
			releaseHistoryRepo: mockReleaseHistoryRepo,
		}

		latestVersion, err := svc.getLatestVersion(context.Background(), "agent-123", releasePo)

		assert.NoError(t, err)
		assert.Equal(t, "v1.0.0", latestVersion)
	})

	t.Run("release nil - gets from history", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockReleaseHistoryRepo := idbaccessmock.NewMockIReleaseHistoryRepo(ctrl)

		historyPo := &dapo.ReleaseHistoryPO{
			ID:           "history-123",
			AgentID:      "agent-123",
			AgentVersion: "v0.9.0",
		}
		mockReleaseHistoryRepo.EXPECT().GetLatestVersionByAgentID(gomock.Any(), "agent-123").Return(historyPo, nil)

		svc := &squareSvc{
			SvcBase:            service.NewSvcBase(),
			releaseHistoryRepo: mockReleaseHistoryRepo,
		}

		latestVersion, err := svc.getLatestVersion(context.Background(), "agent-123", nil)

		assert.NoError(t, err)
		assert.Equal(t, "v0.9.0", latestVersion)
	})

	t.Run("no release or history - returns unpublished", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockReleaseHistoryRepo := idbaccessmock.NewMockIReleaseHistoryRepo(ctrl)
		mockReleaseHistoryRepo.EXPECT().GetLatestVersionByAgentID(gomock.Any(), "agent-123").Return(nil, nil)

		svc := &squareSvc{
			SvcBase:            service.NewSvcBase(),
			releaseHistoryRepo: mockReleaseHistoryRepo,
		}

		latestVersion, err := svc.getLatestVersion(context.Background(), "agent-123", nil)

		assert.NoError(t, err)
		assert.Equal(t, "v0", latestVersion)
	})

	t.Run("history repo error - returns error", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockReleaseHistoryRepo := idbaccessmock.NewMockIReleaseHistoryRepo(ctrl)
		expectedErr := errors.New("history database error")
		mockReleaseHistoryRepo.EXPECT().GetLatestVersionByAgentID(gomock.Any(), "agent-123").Return(nil, expectedErr)

		svc := &squareSvc{
			SvcBase:            service.NewSvcBase(),
			releaseHistoryRepo: mockReleaseHistoryRepo,
		}

		latestVersion, err := svc.getLatestVersion(context.Background(), "agent-123", nil)

		assert.Error(t, err)
		assert.Empty(t, latestVersion)
		assert.Contains(t, err.Error(), "[squareSvc.getLatestVersion]")
	})
}

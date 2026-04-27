package squaresvc

import (
	"context"
	"errors"
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/constant/daconstant"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/square/squarereq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestSquareSvc_GetAgentInfoByIDOrKey(t *testing.T) {
	t.Parallel()

	t.Run("agent exists by ID", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAgentConfRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
		mockReleaseHistoryRepo := idbaccessmock.NewMockIReleaseHistoryRepo(ctrl)

		svc := &squareSvc{
			SvcBase:            &service.SvcBase{Logger: noopSquareLogger{}},
			agentConfRepo:      mockAgentConfRepo,
			releaseHistoryRepo: mockReleaseHistoryRepo,
		}

		mockAgentConfRepo.EXPECT().ExistsByID(gomock.Any(), "agent-123").Return(true, nil)
		mockAgentConfRepo.EXPECT().GetByID(gomock.Any(), "agent-123").Return(&dapo.DataAgentPo{
			ID:     "agent-123",
			Key:    "agent-key",
			Name:   "Agent 123",
			Config: "{}",
		}, nil)
		mockReleaseHistoryRepo.EXPECT().GetLatestVersionByAgentID(gomock.Any(), "agent-123").
			Return(&dapo.ReleaseHistoryPO{AgentVersion: "v1"}, nil)

		req := &squarereq.AgentInfoReq{
			AgentID:      "agent-123",
			AgentVersion: daconstant.AgentVersionUnpublished,
		}
		resp, err := svc.GetAgentInfoByIDOrKey(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "agent-123", req.AgentID)
		assert.Equal(t, "agent-123", resp.ID)
	})

	t.Run("agent key resolved before loading details", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAgentConfRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
		mockReleaseHistoryRepo := idbaccessmock.NewMockIReleaseHistoryRepo(ctrl)

		svc := &squareSvc{
			SvcBase:            &service.SvcBase{Logger: noopSquareLogger{}},
			agentConfRepo:      mockAgentConfRepo,
			releaseHistoryRepo: mockReleaseHistoryRepo,
		}

		mockAgentConfRepo.EXPECT().ExistsByID(gomock.Any(), "agent-key").Return(false, nil)
		mockAgentConfRepo.EXPECT().GetByKey(gomock.Any(), "agent-key").Return(&dapo.DataAgentPo{
			ID:  "agent-456",
			Key: "agent-key",
		}, nil)
		mockAgentConfRepo.EXPECT().GetByID(gomock.Any(), "agent-456").Return(&dapo.DataAgentPo{
			ID:     "agent-456",
			Key:    "agent-key",
			Name:   "Agent 456",
			Config: "{}",
		}, nil)
		mockReleaseHistoryRepo.EXPECT().GetLatestVersionByAgentID(gomock.Any(), "agent-456").
			Return(&dapo.ReleaseHistoryPO{AgentVersion: "v9"}, nil)

		req := &squarereq.AgentInfoReq{
			AgentID:      "agent-key",
			AgentVersion: daconstant.AgentVersionUnpublished,
		}
		resp, err := svc.GetAgentInfoByIDOrKey(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "agent-456", req.AgentID)
		assert.Equal(t, "agent-456", resp.ID)
	})

	t.Run("check and get id error returns early", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAgentConfRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)

		svc := &squareSvc{
			SvcBase:       &service.SvcBase{Logger: noopSquareLogger{}},
			agentConfRepo: mockAgentConfRepo,
		}

		mockAgentConfRepo.EXPECT().ExistsByID(gomock.Any(), "agent-error").Return(false, errors.New("repo down"))

		req := &squarereq.AgentInfoReq{
			AgentID:      "agent-error",
			AgentVersion: daconstant.AgentVersionUnpublished,
		}
		resp, err := svc.GetAgentInfoByIDOrKey(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, "agent-error", req.AgentID)
	})

	t.Run("get agent info error is propagated", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAgentConfRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)

		svc := &squareSvc{
			SvcBase:       &service.SvcBase{Logger: noopSquareLogger{}},
			agentConfRepo: mockAgentConfRepo,
		}

		mockAgentConfRepo.EXPECT().ExistsByID(gomock.Any(), "agent-123").Return(true, nil)
		mockAgentConfRepo.EXPECT().GetByID(gomock.Any(), "agent-123").Return(nil, errors.New("db error"))

		req := &squarereq.AgentInfoReq{
			AgentID:      "agent-123",
			AgentVersion: daconstant.AgentVersionUnpublished,
		}
		resp, err := svc.GetAgentInfoByIDOrKey(context.Background(), req)

		assert.Error(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "agent-123", req.AgentID)
		assert.Contains(t, err.Error(), "[squareSvc.GetAgentInfo]")
	})
}

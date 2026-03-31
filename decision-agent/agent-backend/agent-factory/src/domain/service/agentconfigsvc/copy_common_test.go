package v3agentconfigsvc

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/dbaccess/pubedagentdbacc/padbret"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestGetPublishedAgentPo(t *testing.T) {
	t.Parallel()

	t.Run("returns error when repo fails", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := idbaccessmock.NewMockIPubedAgentRepo(ctrl)
		svc := &dataAgentConfigSvc{
			SvcBase:        service.NewSvcBase(),
			pubedAgentRepo: mockRepo,
		}

		ctx := context.Background()

		mockRepo.EXPECT().GetPubedPoMapByXx(gomock.Any(), gomock.Any()).Return(nil, errors.New("repo error"))

		po, err := svc.getPublishedAgentPo(ctx, "agent1")

		assert.Error(t, err)
		assert.Nil(t, po)
	})

	t.Run("returns error when agent not found in result", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := idbaccessmock.NewMockIPubedAgentRepo(ctrl)
		svc := &dataAgentConfigSvc{
			SvcBase:        service.NewSvcBase(),
			pubedAgentRepo: mockRepo,
		}

		ctx := context.Background()
		// Return a result with an empty map (agent not found)
		ret := &padbret.GetPaPoMapByXxRet{
			JoinPosID2PoMap:  make(map[string]*dapo.PublishedJoinPo),
			JoinPosKey2PoMap: make(map[string]*dapo.PublishedJoinPo),
		}
		mockRepo.EXPECT().GetPubedPoMapByXx(gomock.Any(), gomock.Any()).Return(ret, nil)

		po, err := svc.getPublishedAgentPo(ctx, "agent1")

		assert.Error(t, err)
		assert.Nil(t, po)
		assert.Contains(t, err.Error(), "此已发布的Agent不存在")
	})

	t.Run("returns agent po when found", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := idbaccessmock.NewMockIPubedAgentRepo(ctrl)
		svc := &dataAgentConfigSvc{
			SvcBase:        service.NewSvcBase(),
			pubedAgentRepo: mockRepo,
		}

		ctx := context.Background()
		// Return a result with the agent in the map
		expectedPo := &dapo.DataAgentPo{
			ID: "agent1",
		}
		joinPo := &dapo.PublishedJoinPo{
			DataAgentPo: *expectedPo,
		}
		ret := &padbret.GetPaPoMapByXxRet{
			JoinPosID2PoMap: map[string]*dapo.PublishedJoinPo{
				"agent1": joinPo,
			},
			JoinPosKey2PoMap: make(map[string]*dapo.PublishedJoinPo),
		}
		mockRepo.EXPECT().GetPubedPoMapByXx(gomock.Any(), gomock.Any()).Return(ret, nil)

		po, err := svc.getPublishedAgentPo(ctx, "agent1")

		assert.NoError(t, err)
		assert.NotNil(t, po)
		assert.Equal(t, "agent1", po.ID)
	})
}

func TestGetAgentPoForCopy(t *testing.T) {
	t.Parallel()

	t.Run("returns error when repo not found", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
		svc := &dataAgentConfigSvc{
			SvcBase:       service.NewSvcBase(),
			agentConfRepo: mockRepo,
		}

		ctx := context.Background()
		mockRepo.EXPECT().GetByID(ctx, "agent1").Return(nil, sql.ErrNoRows)

		po, err := svc.getAgentPoForCopy(ctx, "agent1")

		assert.Error(t, err)
		assert.Nil(t, po)
	})

	t.Run("returns agent po when found", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
		svc := &dataAgentConfigSvc{
			SvcBase:       service.NewSvcBase(),
			agentConfRepo: mockRepo,
		}

		ctx := context.Background()
		expectedPo := &dapo.DataAgentPo{
			ID: "agent1",
		}
		mockRepo.EXPECT().GetByID(ctx, "agent1").Return(expectedPo, nil)

		po, err := svc.getAgentPoForCopy(ctx, "agent1")

		assert.NoError(t, err)
		assert.Equal(t, expectedPo, po)
	})
}

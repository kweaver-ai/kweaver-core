package agentinoutsvc

import (
	"context"
	"errors"
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/agent_inout/agentinoutresp"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestUpsertCheckRepeatAgentKey(t *testing.T) {
	t.Parallel()

	t.Run("no existing agents - returns empty map", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
		mockAgentRepo.EXPECT().GetByKeys(gomock.Any(), gomock.Any()).Return([]*dapo.DataAgentPo{}, nil)

		exportData := &agentinoutresp.ExportResp{
			Agents: []*agentinoutresp.ExportAgentItem{
				{DataAgentPo: &dapo.DataAgentPo{Key: "agent-1", Name: "Agent 1"}},
			},
		}
		resp := agentinoutresp.NewImportResp()

		svc := &agentInOutSvc{
			SvcBase:       service.NewSvcBase(),
			agentConfRepo: mockAgentRepo,
		}

		existingMap, err := svc.upsertCheckRepeatAgentKey(context.Background(), exportData, "user-123", resp)

		assert.NoError(t, err)
		assert.NotNil(t, existingMap)
		assert.Empty(t, existingMap)
		assert.False(t, resp.HasFail())
	})

	t.Run("agent owned by current user - added to map", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)

		existingAgent := &dapo.DataAgentPo{
			ID:        "existing-id",
			Key:       "agent-1",
			Name:      "Existing Agent",
			CreatedBy: "user-123",
		}
		mockAgentRepo.EXPECT().GetByKeys(gomock.Any(), []string{"agent-1"}).Return([]*dapo.DataAgentPo{existingAgent}, nil)

		exportData := &agentinoutresp.ExportResp{
			Agents: []*agentinoutresp.ExportAgentItem{
				{DataAgentPo: &dapo.DataAgentPo{Key: "agent-1", Name: "Agent 1"}},
			},
		}
		resp := agentinoutresp.NewImportResp()

		svc := &agentInOutSvc{
			SvcBase:       service.NewSvcBase(),
			agentConfRepo: mockAgentRepo,
		}

		existingMap, err := svc.upsertCheckRepeatAgentKey(context.Background(), exportData, "user-123", resp)

		assert.NoError(t, err)
		assert.NotNil(t, existingMap)
		assert.Len(t, existingMap, 1)
		assert.Equal(t, existingAgent, existingMap["agent-1"])
		assert.False(t, resp.HasFail())
	})

	t.Run("agent owned by different user - marked as conflict", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)

		existingAgent := &dapo.DataAgentPo{
			ID:        "existing-id",
			Key:       "agent-1",
			Name:      "Existing Agent",
			CreatedBy: "other-user",
		}
		mockAgentRepo.EXPECT().GetByKeys(gomock.Any(), []string{"agent-1"}).Return([]*dapo.DataAgentPo{existingAgent}, nil)

		exportData := &agentinoutresp.ExportResp{
			Agents: []*agentinoutresp.ExportAgentItem{
				{DataAgentPo: &dapo.DataAgentPo{Key: "agent-1", Name: "Agent 1"}},
			},
		}
		resp := agentinoutresp.NewImportResp()

		svc := &agentInOutSvc{
			SvcBase:       service.NewSvcBase(),
			agentConfRepo: mockAgentRepo,
		}

		existingMap, err := svc.upsertCheckRepeatAgentKey(context.Background(), exportData, "user-123", resp)

		assert.NoError(t, err)
		assert.NotNil(t, existingMap)
		assert.Empty(t, existingMap)
		assert.True(t, resp.HasFail())
		assert.NotEmpty(t, resp.AgentKeyConflict)
	})

	t.Run("multiple agents - mixed ownership", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)

		agents := []*dapo.DataAgentPo{
			{ID: "id-1", Key: "agent-1", Name: "Agent 1", CreatedBy: "user-123"},
			{ID: "id-2", Key: "agent-2", Name: "Agent 2", CreatedBy: "other-user"},
			{ID: "id-3", Key: "agent-3", Name: "Agent 3", CreatedBy: "user-123"},
		}
		mockAgentRepo.EXPECT().GetByKeys(gomock.Any(), []string{"agent-1", "agent-2", "agent-3"}).Return(agents, nil)

		exportData := &agentinoutresp.ExportResp{
			Agents: []*agentinoutresp.ExportAgentItem{
				{DataAgentPo: &dapo.DataAgentPo{Key: "agent-1", Name: "Agent 1"}},
				{DataAgentPo: &dapo.DataAgentPo{Key: "agent-2", Name: "Agent 2"}},
				{DataAgentPo: &dapo.DataAgentPo{Key: "agent-3", Name: "Agent 3"}},
			},
		}
		resp := agentinoutresp.NewImportResp()

		svc := &agentInOutSvc{
			SvcBase:       service.NewSvcBase(),
			agentConfRepo: mockAgentRepo,
		}

		existingMap, err := svc.upsertCheckRepeatAgentKey(context.Background(), exportData, "user-123", resp)

		assert.NoError(t, err)
		assert.NotNil(t, existingMap)
		assert.Len(t, existingMap, 2)
		assert.Equal(t, agents[0], existingMap["agent-1"])
		assert.Equal(t, agents[2], existingMap["agent-3"])
		assert.True(t, resp.HasFail())
	})

	t.Run("repository error - returns error", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
		expectedErr := errors.New("database error")
		mockAgentRepo.EXPECT().GetByKeys(gomock.Any(), gomock.Any()).Return(nil, expectedErr)

		exportData := &agentinoutresp.ExportResp{
			Agents: []*agentinoutresp.ExportAgentItem{
				{DataAgentPo: &dapo.DataAgentPo{Key: "agent-1", Name: "Agent 1"}},
			},
		}
		resp := agentinoutresp.NewImportResp()

		svc := &agentInOutSvc{
			SvcBase:       service.NewSvcBase(),
			agentConfRepo: mockAgentRepo,
		}

		existingMap, err := svc.upsertCheckRepeatAgentKey(context.Background(), exportData, "user-123", resp)

		assert.Error(t, err)
		assert.Nil(t, existingMap)
	})
}

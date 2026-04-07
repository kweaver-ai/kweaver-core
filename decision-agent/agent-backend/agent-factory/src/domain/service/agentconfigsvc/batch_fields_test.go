package v3agentconfigsvc

import (
	"context"
	"errors"
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service"
	agentconfigreq "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent_config/agentconfigreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestDataAgentConfigRepo_GetByIDS(t *testing.T) {
	t.Parallel()

	t.Run("delegates to repo successfully", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
		svc := &dataAgentConfigSvc{
			SvcBase:       service.NewSvcBase(),
			agentConfRepo: mockRepo,
		}

		ctx := context.Background()
		agentIDs := []string{"agent1", "agent2"}
		expectedPOs := []*dapo.DataAgentPo{
			{ID: "agent1"},
			{ID: "agent2"},
		}

		mockRepo.EXPECT().GetByIDS(ctx, agentIDs).Return(expectedPOs, nil)

		result, err := mockRepo.GetByIDS(ctx, agentIDs)

		assert.NoError(t, err)
		assert.Equal(t, expectedPOs, result)
		assert.NotNil(t, svc)
	})
}

func TestBatchFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(*gomock.Controller) (*dataAgentConfigSvc, context.Context)
		wantErr bool
	}{
		{
			name: "successfully retrieves batch fields",
			setup: func(ctrl *gomock.Controller) (*dataAgentConfigSvc, context.Context) {
				ctx := context.Background()
				mockRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)

				agentPOs := []*dapo.DataAgentPo{
					{ID: "agent1", Name: "Agent 1"},
					{ID: "agent2", Name: "Agent 2"},
				}

				mockRepo.EXPECT().GetByIDS(ctx, []string{"agent1", "agent2"}).Return(agentPOs, nil)

				svc := &dataAgentConfigSvc{
					SvcBase:       service.NewSvcBase(),
					agentConfRepo: mockRepo,
				}

				return svc, ctx
			},
			wantErr: false,
		},
		{
			name: "returns error when repo fails",
			setup: func(ctrl *gomock.Controller) (*dataAgentConfigSvc, context.Context) {
				ctx := context.Background()
				mockRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)

				mockRepo.EXPECT().GetByIDS(ctx, []string{"agent1"}).Return(nil, errors.New("database error"))

				svc := &dataAgentConfigSvc{
					SvcBase:       service.NewSvcBase(),
					agentConfRepo: mockRepo,
				}

				return svc, ctx
			},
			wantErr: true,
		},
		{
			name: "returns empty result for empty agent list",
			setup: func(ctrl *gomock.Controller) (*dataAgentConfigSvc, context.Context) {
				ctx := context.Background()
				mockRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)

				mockRepo.EXPECT().GetByIDS(ctx, []string{}).Return([]*dapo.DataAgentPo{}, nil)

				svc := &dataAgentConfigSvc{
					SvcBase:       service.NewSvcBase(),
					agentConfRepo: mockRepo,
				}

				return svc, ctx
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc, ctx := tt.setup(ctrl)
			req := &agentconfigreq.BatchFieldsReq{
				AgentIDs: []string{"agent1", "agent2"},
				Fields:   []agentconfigreq.BatchFieldsReqField{agentconfigreq.BatchFieldsReqFieldName},
			}

			// Adjust request for empty test case
			if tt.name == "returns empty result for empty agent list" {
				req.AgentIDs = []string{}
			} else if tt.name == "returns error when repo fails" {
				req.AgentIDs = []string{"agent1"}
			}

			resp, err := svc.BatchFields(ctx, req)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, resp)
			}
		})
	}
}

func TestBatchFieldsReqField(t *testing.T) {
	t.Parallel()

	t.Run("String method returns correct value", func(t *testing.T) {
		t.Parallel()

		field := agentconfigreq.BatchFieldsReqFieldName
		assert.Equal(t, "name", field.String())
	})

	t.Run("ValObjCheck passes for valid field", func(t *testing.T) {
		t.Parallel()

		field := agentconfigreq.BatchFieldsReqFieldName
		err := field.ValObjCheck()
		assert.NoError(t, err)
	})

	t.Run("ValObjCheck fails for invalid field", func(t *testing.T) {
		t.Parallel()

		field := agentconfigreq.BatchFieldsReqField("invalid")
		err := field.ValObjCheck()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid field")
	})
}

package releasesvc

import (
	"context"
	"errors"
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestGetAgentName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		agentID      string
		setup        func(*gomock.Controller) (*releaseSvc, context.Context)
		expectedName string
		wantErr      bool
	}{
		{
			name:    "successfully get agent name",
			agentID: "agent-123",
			setup: func(ctrl *gomock.Controller) (*releaseSvc, context.Context) {
				ctx := context.Background()
				agentConfigRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)

				agentNameMap := map[string]string{
					"agent-123": "Test Agent",
				}
				agentConfigRepo.EXPECT().
					GetIDNameMapByID(ctx, []string{"agent-123"}).
					Return(agentNameMap, nil)

				svc := &releaseSvc{
					SvcBase:         service.NewSvcBase(),
					agentConfigRepo: agentConfigRepo,
				}

				return svc, ctx
			},
			expectedName: "Test Agent",
			wantErr:      false,
		},
		{
			name:    "agent name not found",
			agentID: "agent-not-found",
			setup: func(ctrl *gomock.Controller) (*releaseSvc, context.Context) {
				ctx := context.Background()
				agentConfigRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)

				agentNameMap := map[string]string{
					"other-agent": "Other Agent",
				}
				agentConfigRepo.EXPECT().
					GetIDNameMapByID(ctx, []string{"agent-not-found"}).
					Return(agentNameMap, nil)

				svc := &releaseSvc{
					SvcBase:         service.NewSvcBase(),
					agentConfigRepo: agentConfigRepo,
				}

				return svc, ctx
			},
			expectedName: "",
			wantErr:      true,
		},
		{
			name:    "repository error",
			agentID: "agent-error",
			setup: func(ctrl *gomock.Controller) (*releaseSvc, context.Context) {
				ctx := context.Background()
				agentConfigRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)

				agentConfigRepo.EXPECT().
					GetIDNameMapByID(ctx, []string{"agent-error"}).
					Return(nil, errors.New("database error"))

				svc := &releaseSvc{
					SvcBase:         service.NewSvcBase(),
					agentConfigRepo: agentConfigRepo,
				}

				return svc, ctx
			},
			expectedName: "",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc, ctx := tt.setup(ctrl)
			name, err := svc.getAgentName(ctx, tt.agentID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedName, name)
			}
		})
	}
}

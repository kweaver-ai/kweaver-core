package personalspacesvc

import (
	"context"
	"errors"
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdapmsenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driver/iv3portdriver/v3portdrivermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestIsHasBuiltInAgentMgmtPermission(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(*gomock.Controller) (*PersonalSpaceService, context.Context)
		want    bool
		wantErr bool
	}{
		{
			name: "has built-in agent management permission",
			setup: func(ctrl *gomock.Controller) (*PersonalSpaceService, context.Context) {
				ctx := context.Background()
				pmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

				pmsSvc.EXPECT().
					GetSingleMgmtPermission(ctx, cdaenum.ResourceTypeDataAgent, cdapmsenum.AgentBuiltInAgentMgmt).
					Return(true, nil)

				svc := &PersonalSpaceService{
					SvcBase: service.NewSvcBase(),
					pmsSvc:  pmsSvc,
				}

				return svc, ctx
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "no built-in agent management permission",
			setup: func(ctrl *gomock.Controller) (*PersonalSpaceService, context.Context) {
				ctx := context.Background()
				pmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

				pmsSvc.EXPECT().
					GetSingleMgmtPermission(ctx, cdaenum.ResourceTypeDataAgent, cdapmsenum.AgentBuiltInAgentMgmt).
					Return(false, nil)

				svc := &PersonalSpaceService{
					SvcBase: service.NewSvcBase(),
					pmsSvc:  pmsSvc,
				}

				return svc, ctx
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "permission service error",
			setup: func(ctrl *gomock.Controller) (*PersonalSpaceService, context.Context) {
				ctx := context.Background()
				pmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

				pmsSvc.EXPECT().
					GetSingleMgmtPermission(ctx, cdaenum.ResourceTypeDataAgent, cdapmsenum.AgentBuiltInAgentMgmt).
					Return(false, errors.New("permission error"))

				svc := &PersonalSpaceService{
					SvcBase: service.NewSvcBase(),
					pmsSvc:  pmsSvc,
				}

				return svc, ctx
			},
			want:    false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc, ctx := tt.setup(ctrl)
			result, err := svc.isHasBuiltInAgentMgmtPermission(ctx)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}
		})
	}
}

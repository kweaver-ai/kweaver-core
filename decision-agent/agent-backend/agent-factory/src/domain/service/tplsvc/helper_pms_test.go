package tplsvc

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

func TestIsHasPublishPermission(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(*gomock.Controller) (*dataAgentTplSvc, context.Context)
		want    bool
		wantErr bool
	}{
		{
			name: "has publish permission",
			setup: func(ctrl *gomock.Controller) (*dataAgentTplSvc, context.Context) {
				ctx := context.Background()
				pmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

				pmsSvc.EXPECT().
					GetSingleMgmtPermission(ctx, cdaenum.ResourceTypeDataAgentTpl, cdapmsenum.AgentTplPublish).
					Return(true, nil)

				svc := &dataAgentTplSvc{
					SvcBase: service.NewSvcBase(),
					pmsSvc:  pmsSvc,
				}

				return svc, ctx
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "no publish permission",
			setup: func(ctrl *gomock.Controller) (*dataAgentTplSvc, context.Context) {
				ctx := context.Background()
				pmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

				pmsSvc.EXPECT().
					GetSingleMgmtPermission(ctx, cdaenum.ResourceTypeDataAgentTpl, cdapmsenum.AgentTplPublish).
					Return(false, nil)

				svc := &dataAgentTplSvc{
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
			setup: func(ctrl *gomock.Controller) (*dataAgentTplSvc, context.Context) {
				ctx := context.Background()
				pmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

				pmsSvc.EXPECT().
					GetSingleMgmtPermission(ctx, cdaenum.ResourceTypeDataAgentTpl, cdapmsenum.AgentTplPublish).
					Return(false, errors.New("permission error"))

				svc := &dataAgentTplSvc{
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
			result, err := svc.isHasPublishPermission(ctx)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}
		})
	}
}

func TestIsHasUnPublishPermission(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(*gomock.Controller) (*dataAgentTplSvc, context.Context)
		want    bool
		wantErr bool
	}{
		{
			name: "has unpublish permission",
			setup: func(ctrl *gomock.Controller) (*dataAgentTplSvc, context.Context) {
				ctx := context.Background()
				pmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

				pmsSvc.EXPECT().
					GetSingleMgmtPermission(ctx, cdaenum.ResourceTypeDataAgentTpl, cdapmsenum.AgentTplUnpublish).
					Return(true, nil)

				svc := &dataAgentTplSvc{
					SvcBase: service.NewSvcBase(),
					pmsSvc:  pmsSvc,
				}

				return svc, ctx
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "no unpublish permission",
			setup: func(ctrl *gomock.Controller) (*dataAgentTplSvc, context.Context) {
				ctx := context.Background()
				pmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

				pmsSvc.EXPECT().
					GetSingleMgmtPermission(ctx, cdaenum.ResourceTypeDataAgentTpl, cdapmsenum.AgentTplUnpublish).
					Return(false, nil)

				svc := &dataAgentTplSvc{
					SvcBase: service.NewSvcBase(),
					pmsSvc:  pmsSvc,
				}

				return svc, ctx
			},
			want:    false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc, ctx := tt.setup(ctrl)
			result, err := svc.isHasUnPublishPermission(ctx)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}
		})
	}
}

func TestIsHasUnpublishOtherUserAgentTplPermission(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(*gomock.Controller) (*dataAgentTplSvc, context.Context)
		want    bool
		wantErr bool
	}{
		{
			name: "has unpublish other user agent tpl permission",
			setup: func(ctrl *gomock.Controller) (*dataAgentTplSvc, context.Context) {
				ctx := context.Background()
				pmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

				pmsSvc.EXPECT().
					GetSingleMgmtPermission(ctx, cdaenum.ResourceTypeDataAgentTpl, cdapmsenum.AgentTplUnpublishOtherUserAgentTpl).
					Return(true, nil)

				svc := &dataAgentTplSvc{
					SvcBase: service.NewSvcBase(),
					pmsSvc:  pmsSvc,
				}

				return svc, ctx
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "no unpublish other user agent tpl permission",
			setup: func(ctrl *gomock.Controller) (*dataAgentTplSvc, context.Context) {
				ctx := context.Background()
				pmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

				pmsSvc.EXPECT().
					GetSingleMgmtPermission(ctx, cdaenum.ResourceTypeDataAgentTpl, cdapmsenum.AgentTplUnpublishOtherUserAgentTpl).
					Return(false, nil)

				svc := &dataAgentTplSvc{
					SvcBase: service.NewSvcBase(),
					pmsSvc:  pmsSvc,
				}

				return svc, ctx
			},
			want:    false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc, ctx := tt.setup(ctrl)
			result, err := svc.isHasUnpublishOtherUserAgentTplPermission(ctx)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}
		})
	}
}

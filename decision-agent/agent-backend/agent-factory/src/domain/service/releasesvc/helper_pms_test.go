package releasesvc

import (
	"context"
	"errors"
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdapmsenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driver/iv3portdriver/v3portdrivermock"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestIsHasUnpublishOtherUserAgentPermission(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(*gomock.Controller) (*releaseSvc, context.Context)
		want    bool
		wantErr bool
	}{
		{
			name: "has unpublish other user agent permission",
			setup: func(ctrl *gomock.Controller) (*releaseSvc, context.Context) {
				ctx := context.Background()
				pmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

				pmsSvc.EXPECT().
					GetSingleMgmtPermission(ctx, cdaenum.ResourceTypeDataAgent, cdapmsenum.AgentUnpublishOtherUserAgent).
					Return(true, nil)

				svc := &releaseSvc{
					SvcBase: service.NewSvcBase(),
					pmsSvc:  pmsSvc,
				}

				return svc, ctx
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "no unpublish other user agent permission",
			setup: func(ctrl *gomock.Controller) (*releaseSvc, context.Context) {
				ctx := context.Background()
				pmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

				pmsSvc.EXPECT().
					GetSingleMgmtPermission(ctx, cdaenum.ResourceTypeDataAgent, cdapmsenum.AgentUnpublishOtherUserAgent).
					Return(false, nil)

				svc := &releaseSvc{
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
			setup: func(ctrl *gomock.Controller) (*releaseSvc, context.Context) {
				ctx := context.Background()
				pmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

				pmsSvc.EXPECT().
					GetSingleMgmtPermission(ctx, cdaenum.ResourceTypeDataAgent, cdapmsenum.AgentUnpublishOtherUserAgent).
					Return(false, errors.New("permission error"))

				svc := &releaseSvc{
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
			result, err := svc.isHasUnpublishOtherUserAgentPermission(ctx)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}
		})
	}
}

func TestIsHasPublishPermission(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(*gomock.Controller) (*releaseSvc, context.Context, *dapo.DataAgentPo)
		want    bool
		wantErr bool
	}{
		{
			name: "owner has publish permission",
			setup: func(ctrl *gomock.Controller) (*releaseSvc, context.Context, *dapo.DataAgentPo) {
				ctx := context.Background()
				pmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

				builtIn := cdaenum.BuiltInNo
				po := &dapo.DataAgentPo{
					CreatedBy: "user-123",
					IsBuiltIn: &builtIn,
				}

				visitor := &rest.Visitor{
					ID: "user-123",
				}
				ctx = context.WithValue(ctx, cenum.VisitUserInfoCtxKey.String(), visitor) //nolint:staticcheck // SA1029

				pmsSvc.EXPECT().
					GetSingleMgmtPermission(ctx, cdaenum.ResourceTypeDataAgent, cdapmsenum.AgentPublish).
					Return(true, nil)

				svc := &releaseSvc{
					SvcBase: service.NewSvcBase(),
					pmsSvc:  pmsSvc,
				}

				return svc, ctx, po
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "no publish permission",
			setup: func(ctrl *gomock.Controller) (*releaseSvc, context.Context, *dapo.DataAgentPo) {
				ctx := context.Background()
				pmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

				builtIn := cdaenum.BuiltInNo
				po := &dapo.DataAgentPo{
					CreatedBy: "user-123",
					IsBuiltIn: &builtIn,
				}

				visitor := &rest.Visitor{
					ID: "user-123",
				}
				ctx = context.WithValue(ctx, cenum.VisitUserInfoCtxKey.String(), visitor) //nolint:staticcheck // SA1029

				pmsSvc.EXPECT().
					GetSingleMgmtPermission(ctx, cdaenum.ResourceTypeDataAgent, cdapmsenum.AgentPublish).
					Return(false, nil)

				svc := &releaseSvc{
					SvcBase: service.NewSvcBase(),
					pmsSvc:  pmsSvc,
				}

				return svc, ctx, po
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "permission service error",
			setup: func(ctrl *gomock.Controller) (*releaseSvc, context.Context, *dapo.DataAgentPo) {
				ctx := context.Background()
				pmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

				builtIn := cdaenum.BuiltInNo
				po := &dapo.DataAgentPo{
					CreatedBy: "user-123",
					IsBuiltIn: &builtIn,
				}

				visitor := &rest.Visitor{
					ID: "user-123",
				}
				ctx = context.WithValue(ctx, cenum.VisitUserInfoCtxKey.String(), visitor) //nolint:staticcheck // SA1029

				pmsSvc.EXPECT().
					GetSingleMgmtPermission(ctx, cdaenum.ResourceTypeDataAgent, cdapmsenum.AgentPublish).
					Return(false, errors.New("permission error"))

				svc := &releaseSvc{
					SvcBase: service.NewSvcBase(),
					pmsSvc:  pmsSvc,
				}

				return svc, ctx, po
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

			svc, ctx, po := tt.setup(ctrl)
			result, err := svc.isHasPublishPermission(ctx, po)

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
		setup   func(*gomock.Controller) (*releaseSvc, context.Context, *dapo.DataAgentPo)
		want    bool
		wantErr bool
	}{
		{
			name: "owner has unpublish permission",
			setup: func(ctrl *gomock.Controller) (*releaseSvc, context.Context, *dapo.DataAgentPo) {
				ctx := context.Background()
				pmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

				builtIn := cdaenum.BuiltInNo
				po := &dapo.DataAgentPo{
					CreatedBy: "user-123",
					IsBuiltIn: &builtIn,
				}

				visitor := &rest.Visitor{
					ID: "user-123",
				}
				ctx = context.WithValue(ctx, cenum.VisitUserInfoCtxKey.String(), visitor) //nolint:staticcheck // SA1029

				pmsSvc.EXPECT().
					GetSingleMgmtPermission(ctx, cdaenum.ResourceTypeDataAgent, cdapmsenum.AgentUnpublish).
					Return(true, nil)

				svc := &releaseSvc{
					SvcBase: service.NewSvcBase(),
					pmsSvc:  pmsSvc,
				}

				return svc, ctx, po
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "no unpublish permission",
			setup: func(ctrl *gomock.Controller) (*releaseSvc, context.Context, *dapo.DataAgentPo) {
				ctx := context.Background()
				pmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

				builtIn := cdaenum.BuiltInNo
				po := &dapo.DataAgentPo{
					CreatedBy: "user-123",
					IsBuiltIn: &builtIn,
				}

				visitor := &rest.Visitor{
					ID: "user-123",
				}
				ctx = context.WithValue(ctx, cenum.VisitUserInfoCtxKey.String(), visitor) //nolint:staticcheck // SA1029

				pmsSvc.EXPECT().
					GetSingleMgmtPermission(ctx, cdaenum.ResourceTypeDataAgent, cdapmsenum.AgentUnpublish).
					Return(false, nil)

				svc := &releaseSvc{
					SvcBase: service.NewSvcBase(),
					pmsSvc:  pmsSvc,
				}

				return svc, ctx, po
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "permission service error",
			setup: func(ctrl *gomock.Controller) (*releaseSvc, context.Context, *dapo.DataAgentPo) {
				ctx := context.Background()
				pmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

				builtIn := cdaenum.BuiltInNo
				po := &dapo.DataAgentPo{
					CreatedBy: "user-123",
					IsBuiltIn: &builtIn,
				}

				visitor := &rest.Visitor{
					ID: "user-123",
				}
				ctx = context.WithValue(ctx, cenum.VisitUserInfoCtxKey.String(), visitor) //nolint:staticcheck // SA1029

				pmsSvc.EXPECT().
					GetSingleMgmtPermission(ctx, cdaenum.ResourceTypeDataAgent, cdapmsenum.AgentUnpublish).
					Return(false, errors.New("permission error"))

				svc := &releaseSvc{
					SvcBase: service.NewSvcBase(),
					pmsSvc:  pmsSvc,
				}

				return svc, ctx, po
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

			svc, ctx, po := tt.setup(ctrl)
			result, err := svc.isHasUnPublishPermission(ctx, po)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}
		})
	}
}

func TestIsHasPubOrUnPubPms(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setup      func(*gomock.Controller) (*releaseSvc, context.Context, *dapo.DataAgentPo)
		operator   cdapmsenum.Operator
		want       bool
		wantErr    bool
		errMessage string
	}{
		{
			name: "invalid operator",
			setup: func(ctrl *gomock.Controller) (*releaseSvc, context.Context, *dapo.DataAgentPo) {
				ctx := context.Background()
				pmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

				builtIn := cdaenum.BuiltInNo
				po := &dapo.DataAgentPo{
					CreatedBy: "user-123",
					IsBuiltIn: &builtIn,
				}

				svc := &releaseSvc{
					SvcBase: service.NewSvcBase(),
					pmsSvc:  pmsSvc,
				}

				return svc, ctx, po
			},
			operator:   cdapmsenum.AgentUse, // Invalid operator for publish/unpublish
			want:       false,
			wantErr:    true,
			errMessage: "[isHasPubOrUnPubPms]: invalid operator",
		},
		{
			name: "owner has publish permission",
			setup: func(ctrl *gomock.Controller) (*releaseSvc, context.Context, *dapo.DataAgentPo) {
				ctx := context.Background()
				pmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

				builtIn := cdaenum.BuiltInNo
				po := &dapo.DataAgentPo{
					CreatedBy: "user-123",
					IsBuiltIn: &builtIn,
				}

				visitor := &rest.Visitor{
					ID: "user-123",
				}
				ctx = context.WithValue(ctx, cenum.VisitUserInfoCtxKey.String(), visitor) //nolint:staticcheck // SA1029

				pmsSvc.EXPECT().
					GetSingleMgmtPermission(ctx, cdaenum.ResourceTypeDataAgent, cdapmsenum.AgentPublish).
					Return(true, nil)

				svc := &releaseSvc{
					SvcBase: service.NewSvcBase(),
					pmsSvc:  pmsSvc,
				}

				return svc, ctx, po
			},
			operator: cdapmsenum.AgentPublish,
			want:     true,
			wantErr:  false,
		},
		{
			name: "non-owner with built-in agent has built-in mgmt permission",
			setup: func(ctrl *gomock.Controller) (*releaseSvc, context.Context, *dapo.DataAgentPo) {
				ctx := context.Background()
				pmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

				builtIn := cdaenum.BuiltInYes
				po := &dapo.DataAgentPo{
					CreatedBy: "user-456", // Different from visitor
					IsBuiltIn: &builtIn,
				}

				visitor := &rest.Visitor{
					ID: "user-123",
				}
				ctx = context.WithValue(ctx, cenum.VisitUserInfoCtxKey.String(), visitor) //nolint:staticcheck // SA1029

				pmsSvc.EXPECT().
					GetSingleMgmtPermission(ctx, cdaenum.ResourceTypeDataAgent, cdapmsenum.AgentPublish).
					Return(false, nil) // No publish permission

				pmsSvc.EXPECT().
					GetSingleMgmtPermission(ctx, cdaenum.ResourceTypeDataAgent, cdapmsenum.AgentBuiltInAgentMgmt).
					Return(true, nil) // Has built-in agent mgmt permission

				svc := &releaseSvc{
					SvcBase: service.NewSvcBase(),
					pmsSvc:  pmsSvc,
				}

				return svc, ctx, po
			},
			operator: cdapmsenum.AgentPublish,
			want:     true,
			wantErr:  false,
		},
		{
			name: "non-owner without built-in agent has no permission",
			setup: func(ctrl *gomock.Controller) (*releaseSvc, context.Context, *dapo.DataAgentPo) {
				ctx := context.Background()
				pmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

				builtIn := cdaenum.BuiltInNo
				po := &dapo.DataAgentPo{
					CreatedBy: "user-456", // Different from visitor
					IsBuiltIn: &builtIn,
				}

				visitor := &rest.Visitor{
					ID: "user-123",
				}
				ctx = context.WithValue(ctx, cenum.VisitUserInfoCtxKey.String(), visitor) //nolint:staticcheck // SA1029

				pmsSvc.EXPECT().
					GetSingleMgmtPermission(ctx, cdaenum.ResourceTypeDataAgent, cdapmsenum.AgentPublish).
					Return(false, nil) // No publish permission

				svc := &releaseSvc{
					SvcBase: service.NewSvcBase(),
					pmsSvc:  pmsSvc,
				}

				return svc, ctx, po
			},
			operator: cdapmsenum.AgentPublish,
			want:     false,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc, ctx, po := tt.setup(ctrl)
			result, err := svc.isHasPubOrUnPubPms(ctx, po, tt.operator)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMessage)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}
		})
	}
}

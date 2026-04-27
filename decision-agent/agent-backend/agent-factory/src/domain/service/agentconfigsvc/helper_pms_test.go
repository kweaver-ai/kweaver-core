package v3agentconfigsvc

import (
	"context"
	"errors"
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdapmsenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/ihttpaccess/iauthzacc/authzaccmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driver/iv3portdriver/v3portdrivermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestIsHasTplPublishPermission(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(*gomock.Controller) (*dataAgentConfigSvc, context.Context)
		want    bool
		wantErr bool
	}{
		{
			name: "has template publish permission",
			setup: func(ctrl *gomock.Controller) (*dataAgentConfigSvc, context.Context) {
				ctx := context.Background()
				pmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

				pmsSvc.EXPECT().
					GetSingleMgmtPermission(ctx, cdaenum.ResourceTypeDataAgentTpl, cdapmsenum.AgentTplPublish).
					Return(true, nil)

				svc := &dataAgentConfigSvc{
					SvcBase: service.NewSvcBase(),
					pmsSvc:  pmsSvc,
				}

				return svc, ctx
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "no template publish permission",
			setup: func(ctrl *gomock.Controller) (*dataAgentConfigSvc, context.Context) {
				ctx := context.Background()
				pmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

				pmsSvc.EXPECT().
					GetSingleMgmtPermission(ctx, cdaenum.ResourceTypeDataAgentTpl, cdapmsenum.AgentTplPublish).
					Return(false, nil)

				svc := &dataAgentConfigSvc{
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
			setup: func(ctrl *gomock.Controller) (*dataAgentConfigSvc, context.Context) {
				ctx := context.Background()
				pmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

				pmsSvc.EXPECT().
					GetSingleMgmtPermission(ctx, cdaenum.ResourceTypeDataAgentTpl, cdapmsenum.AgentTplPublish).
					Return(false, errors.New("permission error"))

				svc := &dataAgentConfigSvc{
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
			result, err := svc.isHasTplPublishPermission(ctx)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}
		})
	}
}

func TestIsHasBuiltInAgentMgmtPermission(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(*gomock.Controller) (*dataAgentConfigSvc, context.Context)
		want    bool
		wantErr bool
	}{
		{
			name: "has built-in agent mgmt permission",
			setup: func(ctrl *gomock.Controller) (*dataAgentConfigSvc, context.Context) {
				ctx := context.Background()
				pmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

				pmsSvc.EXPECT().
					GetSingleMgmtPermission(ctx, cdaenum.ResourceTypeDataAgent, cdapmsenum.AgentBuiltInAgentMgmt).
					Return(true, nil)

				svc := &dataAgentConfigSvc{
					SvcBase: service.NewSvcBase(),
					pmsSvc:  pmsSvc,
				}

				return svc, ctx
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "no built-in agent mgmt permission",
			setup: func(ctrl *gomock.Controller) (*dataAgentConfigSvc, context.Context) {
				ctx := context.Background()
				pmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

				pmsSvc.EXPECT().
					GetSingleMgmtPermission(ctx, cdaenum.ResourceTypeDataAgent, cdapmsenum.AgentBuiltInAgentMgmt).
					Return(false, nil)

				svc := &dataAgentConfigSvc{
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

func TestIsHasSystemAgentCreatePermission(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(*gomock.Controller) (*dataAgentConfigSvc, context.Context)
		want    bool
		wantErr bool
	}{
		{
			name: "has system agent create permission",
			setup: func(ctrl *gomock.Controller) (*dataAgentConfigSvc, context.Context) {
				ctx := context.Background()
				pmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

				pmsSvc.EXPECT().
					GetSingleMgmtPermission(ctx, cdaenum.ResourceTypeDataAgent, cdapmsenum.AgentCreateSystemAgent).
					Return(true, nil)

				svc := &dataAgentConfigSvc{
					SvcBase: service.NewSvcBase(),
					pmsSvc:  pmsSvc,
				}

				return svc, ctx
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "no system agent create permission",
			setup: func(ctrl *gomock.Controller) (*dataAgentConfigSvc, context.Context) {
				ctx := context.Background()
				pmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

				pmsSvc.EXPECT().
					GetSingleMgmtPermission(ctx, cdaenum.ResourceTypeDataAgent, cdapmsenum.AgentCreateSystemAgent).
					Return(false, nil)

				svc := &dataAgentConfigSvc{
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
			result, err := svc.isHasSystemAgentCreatePermission(ctx)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}
		})
	}
}

func TestIsOwnerOrHasBuiltInAgentMgmtPermission(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(*gomock.Controller) (*dataAgentConfigSvc, context.Context, *dapo.DataAgentPo)
		uid     string
		wantErr bool
	}{
		{
			name: "is owner - should pass",
			setup: func(ctrl *gomock.Controller) (*dataAgentConfigSvc, context.Context, *dapo.DataAgentPo) {
				ctx := context.Background()
				pmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

				builtIn := cdaenum.BuiltInNo
				po := &dapo.DataAgentPo{
					CreatedBy: "user-123",
					IsBuiltIn: &builtIn,
				}

				svc := &dataAgentConfigSvc{
					SvcBase: service.NewSvcBase(),
					pmsSvc:  pmsSvc,
				}

				return svc, ctx, po
			},
			uid:     "user-123",
			wantErr: false,
		},
		{
			name: "not owner, not built-in - should error",
			setup: func(ctrl *gomock.Controller) (*dataAgentConfigSvc, context.Context, *dapo.DataAgentPo) {
				ctx := context.Background()
				pmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

				builtIn := cdaenum.BuiltInNo
				po := &dapo.DataAgentPo{
					CreatedBy: "user-456",
					IsBuiltIn: &builtIn,
				}

				svc := &dataAgentConfigSvc{
					SvcBase: service.NewSvcBase(),
					pmsSvc:  pmsSvc,
				}

				return svc, ctx, po
			},
			uid:     "user-123",
			wantErr: true,
		},
		{
			name: "not owner, is built-in, has mgmt permission - should pass",
			setup: func(ctrl *gomock.Controller) (*dataAgentConfigSvc, context.Context, *dapo.DataAgentPo) {
				ctx := context.Background()
				pmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

				builtIn := cdaenum.BuiltInYes
				po := &dapo.DataAgentPo{
					CreatedBy: "user-456",
					IsBuiltIn: &builtIn,
				}

				pmsSvc.EXPECT().
					GetSingleMgmtPermission(ctx, cdaenum.ResourceTypeDataAgent, cdapmsenum.AgentBuiltInAgentMgmt).
					Return(true, nil)

				svc := &dataAgentConfigSvc{
					SvcBase: service.NewSvcBase(),
					pmsSvc:  pmsSvc,
				}

				return svc, ctx, po
			},
			uid:     "user-123",
			wantErr: false,
		},
		{
			name: "not owner, is built-in, no mgmt permission - should error",
			setup: func(ctrl *gomock.Controller) (*dataAgentConfigSvc, context.Context, *dapo.DataAgentPo) {
				ctx := context.Background()
				pmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

				builtIn := cdaenum.BuiltInYes
				po := &dapo.DataAgentPo{
					CreatedBy: "user-456",
					IsBuiltIn: &builtIn,
				}

				pmsSvc.EXPECT().
					GetSingleMgmtPermission(ctx, cdaenum.ResourceTypeDataAgent, cdapmsenum.AgentBuiltInAgentMgmt).
					Return(false, nil)

				svc := &dataAgentConfigSvc{
					SvcBase: service.NewSvcBase(),
					pmsSvc:  pmsSvc,
				}

				return svc, ctx, po
			},
			uid:     "user-123",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc, ctx, po := tt.setup(ctrl)
			err := svc.isOwnerOrHasBuiltInAgentMgmtPermission(ctx, po, tt.uid)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCheckUseAgentPms(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(*gomock.Controller) (*dataAgentConfigSvc, context.Context, map[string]*dapo.PublishedJoinPo, string)
		want    map[string]struct{}
		wantErr bool
	}{
		{
			name: "all agents have permission - no pms control",
			setup: func(ctrl *gomock.Controller) (*dataAgentConfigSvc, context.Context, map[string]*dapo.PublishedJoinPo, string) {
				ctx := context.Background()
				authZHttp := authzaccmock.NewMockAuthZHttpAcc(ctrl)

				// No pms ctrl agents, but FilterCanUseAgentIDMap is still called with empty slice (or nil)
				authZHttp.EXPECT().
					FilterCanUseAgentIDMap(ctx, "user-123", gomock.Any()).
					Return(map[string]struct{}{}, nil)

				svc := &dataAgentConfigSvc{
					SvcBase:   service.NewSvcBase(),
					authZHttp: authZHttp,
				}

				noPmsCtrl := 0
				publishedMap := map[string]*dapo.PublishedJoinPo{
					"agent-1": {
						DataAgentPo: dapo.DataAgentPo{
							Key: "agent-1",
							ID:  "id-1",
						},
						ReleasePartPo: dapo.ReleasePartPo{
							IsPmsCtrl: noPmsCtrl,
						},
					},
					"agent-2": {
						DataAgentPo: dapo.DataAgentPo{
							Key: "agent-2",
							ID:  "id-2",
						},
						ReleasePartPo: dapo.ReleasePartPo{
							IsPmsCtrl: noPmsCtrl,
						},
					},
				}

				return svc, ctx, publishedMap, "user-123"
			},
			want: map[string]struct{}{
				"agent-1": {},
				"agent-2": {},
			},
			wantErr: false,
		},
		{
			name: "filter agents with pms control",
			setup: func(ctrl *gomock.Controller) (*dataAgentConfigSvc, context.Context, map[string]*dapo.PublishedJoinPo, string) {
				ctx := context.Background()
				authZHttp := authzaccmock.NewMockAuthZHttpAcc(ctrl)

				hasPmsCtrl := 1
				noPmsCtrl := 0
				publishedMap := map[string]*dapo.PublishedJoinPo{
					"agent-1": {
						DataAgentPo: dapo.DataAgentPo{
							Key: "agent-1",
							ID:  "id-1",
						},
						ReleasePartPo: dapo.ReleasePartPo{
							IsPmsCtrl: hasPmsCtrl,
						},
					},
					"agent-2": {
						DataAgentPo: dapo.DataAgentPo{
							Key: "agent-2",
							ID:  "id-2",
						},
						ReleasePartPo: dapo.ReleasePartPo{
							IsPmsCtrl: hasPmsCtrl,
						},
					},
					"agent-3": {
						DataAgentPo: dapo.DataAgentPo{
							Key: "agent-3",
							ID:  "id-3",
						},
						ReleasePartPo: dapo.ReleasePartPo{
							IsPmsCtrl: noPmsCtrl,
						},
					},
				}

				// Only agent-1 has permission, agent-2 doesn't, agent-3 has no pms ctrl
				// Use gomock.Any() for agentIds since map iteration order is not guaranteed
				authZHttp.EXPECT().
					FilterCanUseAgentIDMap(ctx, "user-123", gomock.Any()).
					DoAndReturn(func(_ context.Context, _ string, ids []string) (map[string]struct{}, error) {
						// Verify the IDs contain both id-1 and id-2 (order may vary)
						idMap := make(map[string]struct{})
						for _, id := range ids {
							idMap[id] = struct{}{}
						}
						if _, ok := idMap["id-1"]; !ok {
							panic("Expected id-1 to be in agentIds")
						}
						if _, ok := idMap["id-2"]; !ok {
							panic("Expected id-2 to be in agentIds")
						}
						return map[string]struct{}{
							"id-1": {},
						}, nil
					})

				svc := &dataAgentConfigSvc{
					SvcBase:   service.NewSvcBase(),
					authZHttp: authZHttp,
				}

				return svc, ctx, publishedMap, "user-123"
			},
			want: map[string]struct{}{
				"agent-1": {},
				"agent-3": {},
			},
			wantErr: false,
		},
		{
			name: "all filtered agents have no permission",
			setup: func(ctrl *gomock.Controller) (*dataAgentConfigSvc, context.Context, map[string]*dapo.PublishedJoinPo, string) {
				ctx := context.Background()
				authZHttp := authzaccmock.NewMockAuthZHttpAcc(ctrl)

				hasPmsCtrl := 1
				publishedMap := map[string]*dapo.PublishedJoinPo{
					"agent-1": {
						DataAgentPo: dapo.DataAgentPo{
							Key: "agent-1",
							ID:  "id-1",
						},
						ReleasePartPo: dapo.ReleasePartPo{
							IsPmsCtrl: hasPmsCtrl,
						},
					},
				}

				authZHttp.EXPECT().
					FilterCanUseAgentIDMap(ctx, "user-123", []string{"id-1"}).
					Return(map[string]struct{}{}, nil)

				svc := &dataAgentConfigSvc{
					SvcBase:   service.NewSvcBase(),
					authZHttp: authZHttp,
				}

				return svc, ctx, publishedMap, "user-123"
			},
			want:    map[string]struct{}{},
			wantErr: false,
		},
		{
			name: "authorization http error",
			setup: func(ctrl *gomock.Controller) (*dataAgentConfigSvc, context.Context, map[string]*dapo.PublishedJoinPo, string) {
				ctx := context.Background()
				authZHttp := authzaccmock.NewMockAuthZHttpAcc(ctrl)

				hasPmsCtrl := 1
				publishedMap := map[string]*dapo.PublishedJoinPo{
					"agent-1": {
						DataAgentPo: dapo.DataAgentPo{
							Key: "agent-1",
							ID:  "id-1",
						},
						ReleasePartPo: dapo.ReleasePartPo{
							IsPmsCtrl: hasPmsCtrl,
						},
					},
				}

				authZHttp.EXPECT().
					FilterCanUseAgentIDMap(ctx, "user-123", []string{"id-1"}).
					Return(nil, errors.New("authorization error"))

				svc := &dataAgentConfigSvc{
					SvcBase:   service.NewSvcBase(),
					authZHttp: authZHttp,
				}

				return svc, ctx, publishedMap, "user-123"
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "empty published map",
			setup: func(ctrl *gomock.Controller) (*dataAgentConfigSvc, context.Context, map[string]*dapo.PublishedJoinPo, string) {
				ctx := context.Background()
				authZHttp := authzaccmock.NewMockAuthZHttpAcc(ctrl)

				publishedMap := map[string]*dapo.PublishedJoinPo{}

				// Expect FilterCanUseAgentIDMap to be called with empty slice (or nil)
				authZHttp.EXPECT().
					FilterCanUseAgentIDMap(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(map[string]struct{}{}, nil)

				svc := &dataAgentConfigSvc{
					SvcBase:   service.NewSvcBase(),
					authZHttp: authZHttp,
				}

				return svc, ctx, publishedMap, "user-123"
			},
			want:    map[string]struct{}{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc, ctx, publishedMap, uid := tt.setup(ctrl)
			result, err := svc.checkUseAgentPms(ctx, publishedMap, uid)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}
		})
	}
}

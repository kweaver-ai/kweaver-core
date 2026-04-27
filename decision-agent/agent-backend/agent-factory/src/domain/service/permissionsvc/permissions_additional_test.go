package permissionsvc

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/conf"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdapmsenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/httpaccess/authzhttp/authzhttpres"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/rdto/agent_permission/cpmsreq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/global"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/ihttpaccess/iauthzacc/authzaccmock"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/stretchr/testify/assert"
)

func setPermissionDisablePmsCheck(t *testing.T, disable bool) {
	t.Helper()

	oldCfg := global.GConfig
	global.GConfig = &conf.Config{
		SwitchFields: conf.NewSwitchFields(),
	}
	global.GConfig.SwitchFields.DisablePmsCheck = disable

	t.Cleanup(func() {
		global.GConfig = oldCfg
	})
}

func TestPermissionSvc_GetPolicyOfAgentUse(t *testing.T) {
	t.Parallel()

	t.Run("list policy error", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAuthz := authzaccmock.NewMockAuthZHttpAcc(ctrl)
		svc := &permissionSvc{authZHttp: mockAuthz}
		mockAuthz.EXPECT().ListPolicyAll(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil, errors.New("authz failed"))

		res, err := svc.GetPolicyOfAgentUse(context.Background(), "a1")
		assert.Error(t, err)
		assert.Nil(t, res)
		assert.Contains(t, err.Error(), "获取智能体[a1]使用权限策略失败")
	})

	t.Run("nil result", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAuthz := authzaccmock.NewMockAuthZHttpAcc(ctrl)
		svc := &permissionSvc{authZHttp: mockAuthz}
		mockAuthz.EXPECT().ListPolicyAll(gomock.Any(), gomock.Any(), "token-1").
			Return(nil, nil)

		ctx := context.WithValue(context.Background(), cenum.VisitUserInfoCtxKey.String(), &rest.Visitor{ //nolint:staticcheck // SA1029
			TokenID: "token-1",
		})
		res, err := svc.GetPolicyOfAgentUse(ctx, "a1")
		assert.Error(t, err)
		assert.Nil(t, res)
		assert.Contains(t, err.Error(), "获取智能体使用权限策略返回空结果")
	})

	t.Run("filter expires error", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAuthz := authzaccmock.NewMockAuthZHttpAcc(ctrl)
		svc := &permissionSvc{authZHttp: mockAuthz}
		mockAuthz.EXPECT().ListPolicyAll(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(&authzhttpres.ListPolicyRes{
				Entries: []*authzhttpres.PolicyEntry{
					{
						ExpiresAt: "bad-time",
						Operation: &authzhttpres.PolicyOperation{
							Allow: []*authzhttpres.PolicyOperationItem{{ID: cdapmsenum.AgentUse}},
						},
					},
				},
			}, nil)

		res, err := svc.GetPolicyOfAgentUse(context.Background(), "a1")
		assert.Error(t, err)
		assert.NotNil(t, res)
		assert.Contains(t, err.Error(), "FilterByExpiresAt")
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAuthz := authzaccmock.NewMockAuthZHttpAcc(ctrl)
		svc := &permissionSvc{authZHttp: mockAuthz}

		future := time.Now().Add(time.Hour).Format(time.RFC3339)
		mockAuthz.EXPECT().ListPolicyAll(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(&authzhttpres.ListPolicyRes{
				Entries: []*authzhttpres.PolicyEntry{
					{
						ID:        "keep",
						ExpiresAt: future,
						Operation: &authzhttpres.PolicyOperation{
							Allow: []*authzhttpres.PolicyOperationItem{{ID: cdapmsenum.AgentUse}},
						},
					},
					{
						ID:        "drop-by-op",
						ExpiresAt: future,
						Operation: &authzhttpres.PolicyOperation{
							Allow: []*authzhttpres.PolicyOperationItem{{ID: cdapmsenum.AgentPublish}},
						},
					},
				},
			}, nil)

		res, err := svc.GetPolicyOfAgentUse(context.Background(), "a1")
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Len(t, res.Entries, 1)
		assert.Equal(t, "keep", res.Entries[0].ID)
	})
}

func TestPermissionSvc_CheckUsePermission_AdditionalBranches(t *testing.T) {
	// 不使用 t.Parallel() - 此测试修改全局配置
	t.Run("disable pms check", func(t *testing.T) {
		// 不使用 t.Parallel() - 修改全局状态
		setPermissionDisablePmsCheck(t, true)

		svc := &permissionSvc{}
		resp, err := svc.CheckUsePermission(context.Background(), &cpmsreq.CheckAgentRunReq{
			AgentID: "a1",
		})
		assert.NoError(t, err)
		assert.True(t, resp.IsAllowed)
	})

	t.Run("owner can use", func(t *testing.T) {
		// 不使用 t.Parallel() - 修改全局状态
		setPermissionDisablePmsCheck(t, false)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
		mockReleaseRepo := idbaccessmock.NewMockIReleaseRepo(ctrl)
		mockAuthz := authzaccmock.NewMockAuthZHttpAcc(ctrl)
		svc := &permissionSvc{
			agentConfRepo: mockAgentRepo,
			releaseRepo:   mockReleaseRepo,
			authZHttp:     mockAuthz,
		}

		mockAgentRepo.EXPECT().GetByID(gomock.Any(), "a1").
			Return(&dapo.DataAgentPo{ID: "a1", CreatedBy: "u1"}, nil)

		resp, err := svc.CheckUsePermission(context.Background(), &cpmsreq.CheckAgentRunReq{
			AgentID: "a1",
			UserID:  "u1",
		})
		assert.NoError(t, err)
		assert.True(t, resp.IsAllowed)
	})

	t.Run("release repo error", func(t *testing.T) {
		// 不使用 t.Parallel() - 修改全局状态
		setPermissionDisablePmsCheck(t, false)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
		mockReleaseRepo := idbaccessmock.NewMockIReleaseRepo(ctrl)
		mockAuthz := authzaccmock.NewMockAuthZHttpAcc(ctrl)
		svc := &permissionSvc{
			agentConfRepo: mockAgentRepo,
			releaseRepo:   mockReleaseRepo,
			authZHttp:     mockAuthz,
		}

		mockAgentRepo.EXPECT().GetByID(gomock.Any(), "a1").
			Return(&dapo.DataAgentPo{ID: "a1", CreatedBy: "owner"}, nil)
		mockReleaseRepo.EXPECT().GetByAgentID(gomock.Any(), "a1").
			Return(nil, errors.New("release db error"))

		resp, err := svc.CheckUsePermission(context.Background(), &cpmsreq.CheckAgentRunReq{
			AgentID: "a1",
			UserID:  "u1",
		})
		assert.Error(t, err)
		assert.NotNil(t, resp)
		assert.False(t, resp.IsAllowed)
	})

	t.Run("release exists and not pms ctrl", func(t *testing.T) {
		// 不使用 t.Parallel() - 修改全局状态
		setPermissionDisablePmsCheck(t, false)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
		mockReleaseRepo := idbaccessmock.NewMockIReleaseRepo(ctrl)
		mockAuthz := authzaccmock.NewMockAuthZHttpAcc(ctrl)
		svc := &permissionSvc{
			agentConfRepo: mockAgentRepo,
			releaseRepo:   mockReleaseRepo,
			authZHttp:     mockAuthz,
		}

		mockAgentRepo.EXPECT().GetByID(gomock.Any(), "a1").
			Return(&dapo.DataAgentPo{ID: "a1", CreatedBy: "owner"}, nil)

		isPmsCtrl := 0
		mockReleaseRepo.EXPECT().GetByAgentID(gomock.Any(), "a1").
			Return(&dapo.ReleasePO{IsPmsCtrl: &isPmsCtrl}, nil)

		resp, err := svc.CheckUsePermission(context.Background(), &cpmsreq.CheckAgentRunReq{
			AgentID: "a1",
			UserID:  "u1",
		})
		assert.NoError(t, err)
		assert.True(t, resp.IsAllowed)
	})
}

package releasesvc

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/cconf"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/pmsvo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/cmp/umcmp/umtypes"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cglobal"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/ihttpaccess/iauthzacc/authzaccmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/ihttpaccess/iumacc/httpaccmock"
	"github.com/stretchr/testify/assert"
)

func setReleaseSvcTestConfig(t *testing.T) {
	t.Helper()

	oldCfg := cglobal.GConfig
	cglobal.GConfig = cconf.BaseDefConfig()

	t.Cleanup(func() {
		cglobal.GConfig = oldCfg
	})
}

func TestReleaseSvc_handlePmsCtrl_MoreBranches(t *testing.T) {
	// 不使用 t.Parallel(): setReleaseSvcTestConfig 修改全局 cglobal.GConfig
	setReleaseSvcTestConfig(t)

	t.Run("delete permissions failed", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPermRepo := idbaccessmock.NewMockIReleasePermissionRepo(ctrl)
		svc := &releaseSvc{
			SvcBase:               service.NewSvcBase(),
			releasePermissionRepo: mockPermRepo,
		}

		mockPermRepo.EXPECT().DelByReleaseID(gomock.Any(), nil, "r1").Return(errors.New("del failed"))

		err := svc.handlePmsCtrl(context.Background(), nil, "r1", "a1", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "delete permissions failed")
	})

	t.Run("get agent name failed", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPermRepo := idbaccessmock.NewMockIReleasePermissionRepo(ctrl)
		mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
		svc := &releaseSvc{
			SvcBase:               service.NewSvcBase(),
			releasePermissionRepo: mockPermRepo,
			agentConfigRepo:       mockAgentRepo,
		}

		mockPermRepo.EXPECT().DelByReleaseID(gomock.Any(), nil, "r1").Return(nil)
		mockAgentRepo.EXPECT().GetIDNameMapByID(gomock.Any(), []string{"a1"}).Return(nil, errors.New("db failed"))

		err := svc.handlePmsCtrl(context.Background(), nil, "r1", "a1", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "get agent name failed")
	})

	t.Run("remove use pms failed", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPermRepo := idbaccessmock.NewMockIReleasePermissionRepo(ctrl)
		mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
		mockAuthz := authzaccmock.NewMockAuthZHttpAcc(ctrl)
		svc := &releaseSvc{
			SvcBase:               service.NewSvcBase(),
			releasePermissionRepo: mockPermRepo,
			agentConfigRepo:       mockAgentRepo,
			authZHttp:             mockAuthz,
		}

		mockPermRepo.EXPECT().DelByReleaseID(gomock.Any(), nil, "r1").Return(nil)
		mockAgentRepo.EXPECT().GetIDNameMapByID(gomock.Any(), []string{"a1"}).Return(map[string]string{"a1": "agent-1"}, nil)
		mockAuthz.EXPECT().DeleteAgentPolicy(gomock.Any(), "a1").Return(errors.New("authz failed"))

		err := svc.handlePmsCtrl(context.Background(), nil, "r1", "a1", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "remove use pms failed")
	})

	t.Run("success without pms control", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPermRepo := idbaccessmock.NewMockIReleasePermissionRepo(ctrl)
		mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
		mockAuthz := authzaccmock.NewMockAuthZHttpAcc(ctrl)
		svc := &releaseSvc{
			SvcBase:               service.NewSvcBase(),
			releasePermissionRepo: mockPermRepo,
			agentConfigRepo:       mockAgentRepo,
			authZHttp:             mockAuthz,
		}

		mockPermRepo.EXPECT().DelByReleaseID(gomock.Any(), nil, "r1").Return(nil)
		mockAgentRepo.EXPECT().GetIDNameMapByID(gomock.Any(), []string{"a1"}).Return(map[string]string{"a1": "agent-1"}, nil)
		mockAuthz.EXPECT().DeleteAgentPolicy(gomock.Any(), "a1").Return(nil)

		err := svc.handlePmsCtrl(context.Background(), nil, "r1", "a1", nil)
		assert.NoError(t, err)
	})
}

func TestReleaseSvc_handlePmsCtrlRange_And_genPmsControlResp(t *testing.T) {
	// 不使用 t.Parallel(): setReleaseSvcTestConfig 修改全局 cglobal.GConfig
	setReleaseSvcTestConfig(t)

	t.Run("handlePmsCtrlRange role batch create failed", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPermRepo := idbaccessmock.NewMockIReleasePermissionRepo(ctrl)
		svc := &releaseSvc{
			SvcBase:               service.NewSvcBase(),
			releasePermissionRepo: mockPermRepo,
		}
		pms := &pmsvo.PmsControlObjS{RoleIDs: []string{"r1"}}

		mockPermRepo.EXPECT().BatchCreate(gomock.Any(), nil, gomock.Any()).Return(errors.New("role insert failed"))

		err := svc.handlePmsCtrlRange(context.Background(), pms, "rel1", "agent1", nil, "agent-name")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "batch create role permissions failed")
	})

	t.Run("handlePmsCtrlRange success", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPermRepo := idbaccessmock.NewMockIReleasePermissionRepo(ctrl)
		mockAuthz := authzaccmock.NewMockAuthZHttpAcc(ctrl)
		svc := &releaseSvc{
			SvcBase:               service.NewSvcBase(),
			releasePermissionRepo: mockPermRepo,
			authZHttp:             mockAuthz,
		}
		pms := &pmsvo.PmsControlObjS{
			RoleIDs:       []string{"r1"},
			UserIDs:       []string{"u1"},
			UserGroupIDs:  []string{"g1"},
			DepartmentIDs: []string{"d1"},
			AppAccountIDs: []string{"app1"},
		}

		mockPermRepo.EXPECT().BatchCreate(gomock.Any(), nil, gomock.Any()).Return(nil).Times(5)
		mockAuthz.EXPECT().GrantAgentUsePmsForAccessors(gomock.Any(), gomock.Any(), "agent1", "agent-name").Return(nil)

		err := svc.handlePmsCtrlRange(context.Background(), pms, "rel1", "agent1", nil, "agent-name")
		assert.NoError(t, err)
	})

	t.Run("genPmsControlResp get osn failed", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUm := httpaccmock.NewMockUmHttpAcc(ctrl)
		svc := &releaseSvc{
			SvcBase: service.NewSvcBase(),
			umHttp:  mockUm,
		}
		pos := []*dapo.ReleasePermissionPO{{ObjectType: cenum.PmsTargetObjTypeUser, ObjectId: "u1"}}

		mockUm.EXPECT().GetOsnNames(gomock.Any(), gomock.Any()).Return(nil, errors.New("um failed"))

		resp, err := svc.genPmsControlResp(context.Background(), pos)
		assert.Error(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("genPmsControlResp success with unknown user fallback", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockUm := httpaccmock.NewMockUmHttpAcc(ctrl)
		svc := &releaseSvc{
			SvcBase: service.NewSvcBase(),
			umHttp:  mockUm,
		}
		pos := []*dapo.ReleasePermissionPO{
			{ObjectType: cenum.PmsTargetObjTypeRole, ObjectId: "role-1"},
			{ObjectType: cenum.PmsTargetObjTypeUser, ObjectId: "user-1"},
			{ObjectType: cenum.PmsTargetObjTypeUserGroup, ObjectId: "group-1"},
			{ObjectType: cenum.PmsTargetObjTypeDep, ObjectId: "dep-1"},
			{ObjectType: cenum.PmsTargetObjTypeAppAccount, ObjectId: "app-1"},
		}

		osn := umtypes.NewOsnInfoMapS()
		osn.GroupNameMap["group-1"] = "group one"
		osn.DepartmentNameMap["dep-1"] = "dep one"
		osn.AppNameMap["app-1"] = "app one"
		mockUm.EXPECT().GetOsnNames(gomock.Any(), gomock.Any()).Return(osn, nil)

		resp, err := svc.genPmsControlResp(context.Background(), pos)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Roles, 1)
		assert.Len(t, resp.Users, 1)
		assert.Len(t, resp.UserGroups, 1)
		assert.Len(t, resp.Departments, 1)
		assert.Len(t, resp.AppAccounts, 1)
		assert.NotEmpty(t, resp.Users[0].Username)
	})
}

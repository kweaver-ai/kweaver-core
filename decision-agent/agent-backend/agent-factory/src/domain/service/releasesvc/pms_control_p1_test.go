package releasesvc

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/pmsvo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/authzhttp/authzhttpres"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/ihttpaccess/iauthzacc/authzaccmock"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driver/iv3portdriver/v3portdrivermock"
	"github.com/stretchr/testify/assert"
)

// ---- handlePmsCtrl: pmsControl != nil 路径 ----

func TestReleaseSvc_HandlePmsCtrl_WithPmsControl_RangeError(t *testing.T) {
	t.Parallel()

	setReleaseSvcTestConfig(t)

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
	// handlePmsCtrlRange: role batch create fail
	mockPermRepo.EXPECT().BatchCreate(gomock.Any(), nil, gomock.Any()).Return(errors.New("role insert fail"))

	pms := &pmsvo.PmsControlObjS{RoleIDs: []string{"role-1"}}
	err := svc.handlePmsCtrl(context.Background(), pms, "r1", "a1", nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "handlePmsCtrlRange failed")
}

func TestReleaseSvc_HandlePmsCtrl_WithPmsControl_Success(t *testing.T) {
	t.Parallel()

	setReleaseSvcTestConfig(t)

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
	mockPermRepo.EXPECT().BatchCreate(gomock.Any(), nil, gomock.Any()).Return(nil).Times(5)
	mockAuthz.EXPECT().GrantAgentUsePmsForAccessors(gomock.Any(), gomock.Any(), "a1", "agent-1").Return(nil)

	pms := &pmsvo.PmsControlObjS{
		RoleIDs:       []string{"r1"},
		UserIDs:       []string{"u1"},
		UserGroupIDs:  []string{"g1"},
		DepartmentIDs: []string{"d1"},
		AppAccountIDs: []string{"app1"},
	}
	err := svc.handlePmsCtrl(context.Background(), pms, "r1", "a1", nil)

	assert.NoError(t, err)
}

// ---- handlePmsCtrlRange: 各类型 BatchCreate 失败分支 ----

func TestReleaseSvc_HandlePmsCtrlRange_UserBatchCreateError(t *testing.T) {
	t.Parallel()

	setReleaseSvcTestConfig(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPermRepo := idbaccessmock.NewMockIReleasePermissionRepo(ctrl)
	svc := &releaseSvc{
		SvcBase:               service.NewSvcBase(),
		releasePermissionRepo: mockPermRepo,
	}

	pms := &pmsvo.PmsControlObjS{
		RoleIDs: []string{},
		UserIDs: []string{"u1"},
	}

	// roles batch create ok, users batch create fail
	mockPermRepo.EXPECT().BatchCreate(gomock.Any(), nil, gomock.Any()).Return(nil).Times(1)
	mockPermRepo.EXPECT().BatchCreate(gomock.Any(), nil, gomock.Any()).Return(errors.New("user insert fail")).Times(1)

	err := svc.handlePmsCtrlRange(context.Background(), pms, "rel1", "agent1", nil, "agent-name")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "batch create user permissions failed")
}

func TestReleaseSvc_HandlePmsCtrlRange_UserGroupBatchCreateError(t *testing.T) {
	t.Parallel()

	setReleaseSvcTestConfig(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPermRepo := idbaccessmock.NewMockIReleasePermissionRepo(ctrl)
	svc := &releaseSvc{
		SvcBase:               service.NewSvcBase(),
		releasePermissionRepo: mockPermRepo,
	}

	pms := &pmsvo.PmsControlObjS{
		RoleIDs:      []string{},
		UserIDs:      []string{},
		UserGroupIDs: []string{"g1"},
	}

	mockPermRepo.EXPECT().BatchCreate(gomock.Any(), nil, gomock.Any()).Return(nil).Times(2)
	mockPermRepo.EXPECT().BatchCreate(gomock.Any(), nil, gomock.Any()).Return(errors.New("group insert fail")).Times(1)

	err := svc.handlePmsCtrlRange(context.Background(), pms, "rel1", "agent1", nil, "agent-name")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "batch create user group permissions failed")
}

func TestReleaseSvc_HandlePmsCtrlRange_DepartmentBatchCreateError(t *testing.T) {
	t.Parallel()

	setReleaseSvcTestConfig(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPermRepo := idbaccessmock.NewMockIReleasePermissionRepo(ctrl)
	svc := &releaseSvc{
		SvcBase:               service.NewSvcBase(),
		releasePermissionRepo: mockPermRepo,
	}

	pms := &pmsvo.PmsControlObjS{
		DepartmentIDs: []string{"d1"},
	}

	mockPermRepo.EXPECT().BatchCreate(gomock.Any(), nil, gomock.Any()).Return(nil).Times(3)
	mockPermRepo.EXPECT().BatchCreate(gomock.Any(), nil, gomock.Any()).Return(errors.New("dept insert fail")).Times(1)

	err := svc.handlePmsCtrlRange(context.Background(), pms, "rel1", "agent1", nil, "agent-name")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "batch create department permissions failed")
}

func TestReleaseSvc_HandlePmsCtrlRange_AppAccountBatchCreateError(t *testing.T) {
	t.Parallel()

	setReleaseSvcTestConfig(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPermRepo := idbaccessmock.NewMockIReleasePermissionRepo(ctrl)
	svc := &releaseSvc{
		SvcBase:               service.NewSvcBase(),
		releasePermissionRepo: mockPermRepo,
	}

	pms := &pmsvo.PmsControlObjS{
		AppAccountIDs: []string{"app1"},
	}

	mockPermRepo.EXPECT().BatchCreate(gomock.Any(), nil, gomock.Any()).Return(nil).Times(4)
	mockPermRepo.EXPECT().BatchCreate(gomock.Any(), nil, gomock.Any()).Return(errors.New("app insert fail")).Times(1)

	err := svc.handlePmsCtrlRange(context.Background(), pms, "rel1", "agent1", nil, "agent-name")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "batch create app account permissions failed")
}

func TestReleaseSvc_HandlePmsCtrlRange_GrantUsePmsError(t *testing.T) {
	t.Parallel()

	setReleaseSvcTestConfig(t)

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
		UserIDs: []string{"u1"},
	}

	mockPermRepo.EXPECT().BatchCreate(gomock.Any(), nil, gomock.Any()).Return(nil).Times(5)
	mockAuthz.EXPECT().GrantAgentUsePmsForAccessors(gomock.Any(), gomock.Any(), "agent1", "agent-name").Return(errors.New("grant fail"))

	err := svc.handlePmsCtrlRange(context.Background(), pms, "rel1", "agent1", nil, "agent-name")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "grant use pms failed")
}

// ---- genPmsControlRespFromPolicy ----

func TestReleaseSvc_GenPmsControlRespFromPolicy_Error(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)
	svc := &releaseSvc{
		SvcBase: service.NewSvcBase(),
		pmsSvc:  mockPmsSvc,
	}

	mockPmsSvc.EXPECT().GetPolicyOfAgentUse(gomock.Any(), "agent-1").Return(nil, errors.New("pms error"))

	resp, err := svc.genPmsControlRespFromPolicy(context.Background(), "agent-1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "获取智能体")
	assert.NotNil(t, resp)
}

func TestReleaseSvc_GenPmsControlRespFromPolicy_EmptyEntries(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)
	svc := &releaseSvc{
		SvcBase: service.NewSvcBase(),
		pmsSvc:  mockPmsSvc,
	}

	mockPmsSvc.EXPECT().GetPolicyOfAgentUse(gomock.Any(), "agent-1").Return(&authzhttpres.ListPolicyRes{
		Entries: nil,
	}, nil)

	resp, err := svc.genPmsControlRespFromPolicy(context.Background(), "agent-1")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Empty(t, resp.Roles)
	assert.Empty(t, resp.Users)
}

func TestReleaseSvc_GenPmsControlRespFromPolicy_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)
	svc := &releaseSvc{
		SvcBase: service.NewSvcBase(),
		pmsSvc:  mockPmsSvc,
	}

	entries := []*authzhttpres.PolicyEntry{
		{Accessor: &authzhttpres.PolicyAccessor{ID: "role-1", Type: cenum.PmsTargetObjTypeRole, Name: "Role One"}},
		{Accessor: &authzhttpres.PolicyAccessor{ID: "user-1", Type: cenum.PmsTargetObjTypeUser, Name: "User One"}},
		{Accessor: &authzhttpres.PolicyAccessor{ID: "group-1", Type: cenum.PmsTargetObjTypeUserGroup, Name: "Group One"}},
		{Accessor: &authzhttpres.PolicyAccessor{ID: "dep-1", Type: cenum.PmsTargetObjTypeDep, Name: "Dep One"}},
		{Accessor: &authzhttpres.PolicyAccessor{ID: "app-1", Type: cenum.PmsTargetObjTypeAppAccount, Name: "App One"}},
		nil,
		{Accessor: nil},
	}

	mockPmsSvc.EXPECT().GetPolicyOfAgentUse(gomock.Any(), "agent-1").Return(&authzhttpres.ListPolicyRes{
		Entries: entries,
	}, nil)

	resp, err := svc.genPmsControlRespFromPolicy(context.Background(), "agent-1")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Roles, 1)
	assert.Len(t, resp.Users, 1)
	assert.Len(t, resp.UserGroups, 1)
	assert.Len(t, resp.Departments, 1)
	assert.Len(t, resp.AppAccounts, 1)
	assert.Equal(t, "role-1", resp.Roles[0].RoleID)
}

package permissionsvc

import (
	"context"
	"errors"
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/conf"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/rdto/agent_permission/cpmsreq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/global"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/ihttpaccess/iauthzacc/authzaccmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/ihttpaccess/iumacc/httpaccmock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func init() {
	// Initialize global config for tests
	if global.GConfig == nil {
		global.GConfig = &conf.Config{
			SwitchFields: conf.NewSwitchFields(),
		}
	}
}

func TestPermissionSvc_CheckUsePermission_PanicsWithoutAgentConfRepo(t *testing.T) {
	t.Parallel()

	svc := &permissionSvc{
		SvcBase: service.NewSvcBase(),
	}

	ctx := context.Background()
	req := &cpmsreq.CheckAgentRunReq{
		AgentID: "agent-123",
		UserID:  "user-123", // Provide UserID to bypass the empty check
	}

	assert.Panics(t, func() {
		_, _ = svc.CheckUsePermission(ctx, req)
	})
}

func TestPermissionSvc_CheckUsePermission_AgentNotFound(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentConfRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockReleaseRepo := idbaccessmock.NewMockIReleaseRepo(ctrl)
	mockAuthZHttp := authzaccmock.NewMockAuthZHttpAcc(ctrl)
	mockUmHttp := httpaccmock.NewMockUmHttpAcc(ctrl)

	svc := &permissionSvc{
		SvcBase:       service.NewSvcBase(),
		agentConfRepo: mockAgentConfRepo,
		releaseRepo:   mockReleaseRepo,
		authZHttp:     mockAuthZHttp,
		umHttp:        mockUmHttp,
	}

	ctx := context.Background()
	req := &cpmsreq.CheckAgentRunReq{
		AgentID: "non-existent-agent",
		UserID:  "user-123",
	}

	notFoundErr := errors.New("sql: no rows in result set")
	mockAgentConfRepo.EXPECT().GetByID(gomock.Any(), "non-existent-agent").Return(nil, notFoundErr)

	resp, err := svc.CheckUsePermission(ctx, req)

	assert.Error(t, err)
	assert.NotNil(t, resp)
}

func TestPermissionSvc_CheckUsePermission_GetByIDError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentConfRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockReleaseRepo := idbaccessmock.NewMockIReleaseRepo(ctrl)
	mockAuthZHttp := authzaccmock.NewMockAuthZHttpAcc(ctrl)
	mockUmHttp := httpaccmock.NewMockUmHttpAcc(ctrl)

	svc := &permissionSvc{
		SvcBase:       service.NewSvcBase(),
		agentConfRepo: mockAgentConfRepo,
		releaseRepo:   mockReleaseRepo,
		authZHttp:     mockAuthZHttp,
		umHttp:        mockUmHttp,
	}

	ctx := context.Background()
	req := &cpmsreq.CheckAgentRunReq{
		AgentID: "agent-123",
		UserID:  "user-123",
	}

	dbErr := errors.New("database connection failed")
	mockAgentConfRepo.EXPECT().GetByID(gomock.Any(), "agent-123").Return(nil, dbErr)

	resp, err := svc.CheckUsePermission(ctx, req)

	assert.Error(t, err)
	assert.NotNil(t, resp)
}

func TestPermissionSvc_CheckUsePermission_NoUserOrAppAccount(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentConfRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockReleaseRepo := idbaccessmock.NewMockIReleaseRepo(ctrl)
	mockAuthZHttp := authzaccmock.NewMockAuthZHttpAcc(ctrl)
	mockUmHttp := httpaccmock.NewMockUmHttpAcc(ctrl)

	svc := &permissionSvc{
		SvcBase:       service.NewSvcBase(),
		agentConfRepo: mockAgentConfRepo,
		releaseRepo:   mockReleaseRepo,
		authZHttp:     mockAuthZHttp,
		umHttp:        mockUmHttp,
	}

	ctx := context.Background()
	req := &cpmsreq.CheckAgentRunReq{
		AgentID:      "agent-123",
		UserID:       "",
		AppAccountID: "",
	}

	resp, err := svc.CheckUsePermission(ctx, req)

	assert.Error(t, err)
	assert.NotNil(t, resp)
	assert.Contains(t, err.Error(), "user id or app account id cannot be all empty")
}

func TestPermissionSvc_CheckUsePermission_PoNil(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentConfRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockReleaseRepo := idbaccessmock.NewMockIReleaseRepo(ctrl)
	mockAuthZHttp := authzaccmock.NewMockAuthZHttpAcc(ctrl)
	mockUmHttp := httpaccmock.NewMockUmHttpAcc(ctrl)

	svc := &permissionSvc{
		SvcBase:       service.NewSvcBase(),
		agentConfRepo: mockAgentConfRepo,
		releaseRepo:   mockReleaseRepo,
		authZHttp:     mockAuthZHttp,
		umHttp:        mockUmHttp,
	}

	ctx := context.Background()
	req := &cpmsreq.CheckAgentRunReq{
		AgentID: "agent-123",
		UserID:  "user-123",
	}

	// Return nil PO - this will panic in the conversion
	mockAgentConfRepo.EXPECT().GetByID(gomock.Any(), "agent-123").Return(nil, nil)

	assert.Panics(t, func() {
		_, _ = svc.CheckUsePermission(ctx, req)
	})
}

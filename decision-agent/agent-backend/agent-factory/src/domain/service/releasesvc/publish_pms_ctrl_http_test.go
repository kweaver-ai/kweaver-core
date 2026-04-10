package releasesvc

import (
	"context"
	"errors"
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/ihttpaccess/iauthzacc/authzaccmock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestReleaseSvc_RemoveUsePmsByHTTPAcc_DeleteAgentPolicyError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthZHttp := authzaccmock.NewMockAuthZHttpAcc(ctrl)
	mockReleaseRepo := idbaccessmock.NewMockIReleaseRepo(ctrl)
	mockReleaseHistoryRepo := idbaccessmock.NewMockIReleaseHistoryRepo(ctrl)
	mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)

	svc := &releaseSvc{
		SvcBase:            service.NewSvcBase(),
		authZHttp:          mockAuthZHttp,
		releaseRepo:        mockReleaseRepo,
		releaseHistoryRepo: mockReleaseHistoryRepo,
		agentConfigRepo:    mockAgentRepo,
	}

	ctx := context.Background()
	agentID := "agent-123"
	httpErr := errors.New("http request failed")

	mockAuthZHttp.EXPECT().DeleteAgentPolicy(ctx, agentID).Return(httpErr)

	err := svc.removeUsePmsByHTTPAcc(ctx, agentID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "remove use pms failed")
}

func TestReleaseSvc_RemoveUsePmsByHTTPAcc_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthZHttp := authzaccmock.NewMockAuthZHttpAcc(ctrl)
	mockReleaseRepo := idbaccessmock.NewMockIReleaseRepo(ctrl)
	mockReleaseHistoryRepo := idbaccessmock.NewMockIReleaseHistoryRepo(ctrl)
	mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)

	svc := &releaseSvc{
		SvcBase:            service.NewSvcBase(),
		authZHttp:          mockAuthZHttp,
		releaseRepo:        mockReleaseRepo,
		releaseHistoryRepo: mockReleaseHistoryRepo,
		agentConfigRepo:    mockAgentRepo,
	}

	ctx := context.Background()
	agentID := "agent-123"

	mockAuthZHttp.EXPECT().DeleteAgentPolicy(ctx, agentID).Return(nil)

	err := svc.removeUsePmsByHTTPAcc(ctx, agentID)

	assert.NoError(t, err)
}

func TestReleaseSvc_GrantUsePms_EmptyAccessors(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthZHttp := authzaccmock.NewMockAuthZHttpAcc(ctrl)
	mockReleaseRepo := idbaccessmock.NewMockIReleaseRepo(ctrl)
	mockReleaseHistoryRepo := idbaccessmock.NewMockIReleaseHistoryRepo(ctrl)
	mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)

	svc := &releaseSvc{
		SvcBase:            service.NewSvcBase(),
		authZHttp:          mockAuthZHttp,
		releaseRepo:        mockReleaseRepo,
		releaseHistoryRepo: mockReleaseHistoryRepo,
		agentConfigRepo:    mockAgentRepo,
	}

	ctx := context.Background()
	agentID := "agent-123"
	agentName := "Test Agent"
	pmsMap := map[cenum.PmsTargetObjType][]string{}

	err := svc.grantUsePms(ctx, agentID, agentName, pmsMap)

	assert.NoError(t, err)
}

func TestReleaseSvc_GrantUsePms_GrantAgentUsePmsForAccessorsError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthZHttp := authzaccmock.NewMockAuthZHttpAcc(ctrl)
	mockReleaseRepo := idbaccessmock.NewMockIReleaseRepo(ctrl)
	mockReleaseHistoryRepo := idbaccessmock.NewMockIReleaseHistoryRepo(ctrl)
	mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)

	svc := &releaseSvc{
		SvcBase:            service.NewSvcBase(),
		authZHttp:          mockAuthZHttp,
		releaseRepo:        mockReleaseRepo,
		releaseHistoryRepo: mockReleaseHistoryRepo,
		agentConfigRepo:    mockAgentRepo,
	}

	ctx := context.Background()
	agentID := "agent-123"
	agentName := "Test Agent"
	pmsMap := map[cenum.PmsTargetObjType][]string{
		cenum.PmsTargetObjTypeUser: {"user-1", "user-2"},
	}
	httpErr := errors.New("http request failed")

	mockAuthZHttp.EXPECT().GrantAgentUsePmsForAccessors(gomock.Any(), gomock.Any(), agentID, agentName).Return(httpErr)

	err := svc.grantUsePms(ctx, agentID, agentName, pmsMap)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "grant use pms failed")
}

func TestReleaseSvc_GrantUsePms_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthZHttp := authzaccmock.NewMockAuthZHttpAcc(ctrl)
	mockReleaseRepo := idbaccessmock.NewMockIReleaseRepo(ctrl)
	mockReleaseHistoryRepo := idbaccessmock.NewMockIReleaseHistoryRepo(ctrl)
	mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)

	svc := &releaseSvc{
		SvcBase:            service.NewSvcBase(),
		authZHttp:          mockAuthZHttp,
		releaseRepo:        mockReleaseRepo,
		releaseHistoryRepo: mockReleaseHistoryRepo,
		agentConfigRepo:    mockAgentRepo,
	}

	ctx := context.Background()
	agentID := "agent-123"
	agentName := "Test Agent"
	pmsMap := map[cenum.PmsTargetObjType][]string{
		cenum.PmsTargetObjTypeUser: {"user-1", "user-2"},
	}

	mockAuthZHttp.EXPECT().GrantAgentUsePmsForAccessors(gomock.Any(), gomock.Any(), agentID, agentName).Return(nil)

	err := svc.grantUsePms(ctx, agentID, agentName, pmsMap)

	assert.NoError(t, err)
}

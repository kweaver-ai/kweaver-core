package releasesvc

import (
	"context"
	"errors"
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/release/releasereq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/ihttpaccess/iumacc/httpaccmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driver/iv3portdriver/v3portdrivermock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestReleaseSvc_Publish_PanicsWithoutAgentConfigRepo(t *testing.T) {
	t.Parallel()

	svc := &releaseSvc{}
	// All repos are nil

	ctx := context.Background()
	req := &releasereq.PublishReq{
		AgentID:              "agent-123",
		UpdatePublishInfoReq: &releasereq.UpdatePublishInfoReq{},
		IsInternalAPI:        false,
	}

	assert.Panics(t, func() {
		_, _, _ = svc.Publish(ctx, req)
	})
}

func TestReleaseSvc_Publish_AgentNotFound(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentConfigRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockUmHttp := httpaccmock.NewMockUmHttpAcc(ctrl)
	mockPmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

	svc := &releaseSvc{
		SvcBase:         service.NewSvcBase(),
		agentConfigRepo: mockAgentConfigRepo,
		umHttp:          mockUmHttp,
		pmsSvc:          mockPmsSvc,
	}

	ctx := context.Background()
	req := &releasereq.PublishReq{
		AgentID:              "non-existent-agent",
		UpdatePublishInfoReq: &releasereq.UpdatePublishInfoReq{},
		IsInternalAPI:        false,
	}

	mockAgentConfigRepo.EXPECT().GetByID(gomock.Any(), "non-existent-agent").Return(nil, errors.New("record not found"))

	resp, auditLogInfo, err := svc.Publish(ctx, req)

	assert.Error(t, err)
	assert.NotNil(t, resp) // Response is created by defer function
	assert.Empty(t, auditLogInfo.ID)
	assert.Empty(t, auditLogInfo.Name)
}

func TestReleaseSvc_Publish_GetByIDError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentConfigRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockUmHttp := httpaccmock.NewMockUmHttpAcc(ctrl)
	mockPmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

	svc := &releaseSvc{
		SvcBase:         service.NewSvcBase(),
		agentConfigRepo: mockAgentConfigRepo,
		umHttp:          mockUmHttp,
		pmsSvc:          mockPmsSvc,
	}

	ctx := context.Background()
	req := &releasereq.PublishReq{
		AgentID:              "agent-123",
		UpdatePublishInfoReq: &releasereq.UpdatePublishInfoReq{},
		IsInternalAPI:        false,
	}

	dbErr := errors.New("database connection failed")
	mockAgentConfigRepo.EXPECT().GetByID(gomock.Any(), "agent-123").Return(nil, dbErr)

	resp, auditLogInfo, err := svc.Publish(ctx, req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "get agent config by id failed")
	assert.NotNil(t, resp) // Response is created by defer function
	assert.Empty(t, auditLogInfo.ID)
	assert.Empty(t, auditLogInfo.Name)
}

func TestReleaseSvc_Publish_AgentFound(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentConfigRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockUmHttp := httpaccmock.NewMockUmHttpAcc(ctrl)
	mockPmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

	svc := &releaseSvc{
		SvcBase:         service.NewSvcBase(),
		agentConfigRepo: mockAgentConfigRepo,
		umHttp:          mockUmHttp,
		pmsSvc:          mockPmsSvc,
	}

	ctx := context.Background()
	req := &releasereq.PublishReq{
		AgentID:              "agent-123",
		UpdatePublishInfoReq: &releasereq.UpdatePublishInfoReq{},
		IsInternalAPI:        false,
	}

	agentPo := &dapo.DataAgentPo{
		ID:   "agent-123",
		Name: "Test Agent",
	}

	mockAgentConfigRepo.EXPECT().GetByID(gomock.Any(), "agent-123").Return(agentPo, nil)
	mockPmsSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)

	// This will panic at some point due to other nil repos, but we can test the initial logic
	assert.Panics(t, func() {
		_, _, _ = svc.Publish(ctx, req)
	})
}

func TestReleaseSvc_Publish_InternalAPI(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentConfigRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockUmHttp := httpaccmock.NewMockUmHttpAcc(ctrl)
	mockPmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

	svc := &releaseSvc{
		SvcBase:         service.NewSvcBase(),
		agentConfigRepo: mockAgentConfigRepo,
		umHttp:          mockUmHttp,
		pmsSvc:          mockPmsSvc,
	}

	ctx := context.Background()
	req := &releasereq.PublishReq{
		AgentID:              "agent-123",
		UpdatePublishInfoReq: &releasereq.UpdatePublishInfoReq{},
		IsInternalAPI:        true, // Internal API, skip permission check
	}

	agentPo := &dapo.DataAgentPo{
		ID:   "agent-123",
		Name: "Test Agent",
	}

	mockAgentConfigRepo.EXPECT().GetByID(gomock.Any(), "agent-123").Return(agentPo, nil)
	// No permission check expected for internal API

	// This will panic at some point due to other nil repos
	assert.Panics(t, func() {
		_, _, _ = svc.Publish(ctx, req)
	})
}

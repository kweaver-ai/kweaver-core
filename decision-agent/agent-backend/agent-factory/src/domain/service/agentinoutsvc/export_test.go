package agentinoutsvc

import (
	"context"
	"errors"
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/agent_inout/agentinoutreq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driver/iv3portdriver/v3portdrivermock"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestAgentInOutSvc_Export_NoUserIDReturnsError(t *testing.T) {
	t.Parallel()

	svc := &agentInOutSvc{
		SvcBase: service.NewSvcBase(),
	}

	ctx := context.Background()
	req := &agentinoutreq.ExportReq{
		AgentIDs: []string{"agent-1"},
	}

	resp, filename, err := svc.Export(ctx, req)

	assert.Error(t, err)
	assert.NotNil(t, resp)
	assert.Empty(t, filename)
	assert.Contains(t, err.Error(), "无法获取用户ID")
}

func TestAgentInOutSvc_Export_PanicsWithoutAgentConfRepo(t *testing.T) {
	t.Parallel()

	svc := &agentInOutSvc{
		SvcBase: service.NewSvcBase(),
	}

	ctx := context.Background()
	userID := "user-123"
	req := &agentinoutreq.ExportReq{
		AgentIDs: []string{"agent-1"},
	}

	// Set up context with user ID
	visitor := &rest.Visitor{
		ID: userID,
	}
	ctx = context.WithValue(ctx, cenum.VisitUserInfoCtxKey.String(), visitor) //nolint:staticcheck // SA1029

	assert.Panics(t, func() {
		_, _, _ = svc.Export(ctx, req)
	})
}

func TestAgentInOutSvc_Export_GetByIDsAndCreatedByError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentConfRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockPmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

	svc := &agentInOutSvc{
		SvcBase:       service.NewSvcBase(),
		agentConfRepo: mockAgentConfRepo,
		pmsSvc:        mockPmsSvc,
	}

	ctx := context.Background()
	userID := "user-123"
	req := &agentinoutreq.ExportReq{
		AgentIDs: []string{"agent-1"},
	}

	// Set up context with user ID
	visitor := &rest.Visitor{
		ID: userID,
	}
	ctx = context.WithValue(ctx, cenum.VisitUserInfoCtxKey.String(), visitor) //nolint:staticcheck // SA1029

	dbErr := errors.New("database connection failed")
	mockAgentConfRepo.EXPECT().GetByIDsAndCreatedBy(gomock.Any(), req.AgentIDs, userID).Return(nil, dbErr)

	resp, filename, err := svc.Export(ctx, req)

	assert.Error(t, err)
	assert.NotNil(t, resp)
	assert.Empty(t, filename)
}

func TestAgentInOutSvc_Export_AgentNotFound(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentConfRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockPmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

	svc := &agentInOutSvc{
		SvcBase:       service.NewSvcBase(),
		agentConfRepo: mockAgentConfRepo,
		pmsSvc:        mockPmsSvc,
	}

	ctx := context.Background()
	userID := "user-123"
	req := &agentinoutreq.ExportReq{
		AgentIDs: []string{"agent-1"},
	}

	// Set up context with user ID
	visitor := &rest.Visitor{
		ID: userID,
	}
	ctx = context.WithValue(ctx, cenum.VisitUserInfoCtxKey.String(), visitor) //nolint:staticcheck // SA1029

	// Return empty list - agent not found
	mockAgentConfRepo.EXPECT().GetByIDsAndCreatedBy(gomock.Any(), req.AgentIDs, userID).Return([]*dapo.DataAgentPo{}, nil)

	resp, filename, err := svc.Export(ctx, req)

	assert.Error(t, err)
	assert.NotNil(t, resp)
	assert.Empty(t, filename)
	assert.Contains(t, err.Error(), "not_found_agents")
}

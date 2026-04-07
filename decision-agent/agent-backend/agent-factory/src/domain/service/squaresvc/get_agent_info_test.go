package squaresvc

import (
	"context"
	"errors"
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/square/squarereq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/ihttpaccess/iumacc/httpaccmock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestSquareSvc_GetAgentInfo_PanicsWithoutAgentConfRepo(t *testing.T) {
	t.Parallel()

	svc := &squareSvc{
		SvcBase: service.NewSvcBase(),
	}

	ctx := context.Background()
	req := &squarereq.AgentInfoReq{
		AgentID: "agent-123",
	}

	assert.Panics(t, func() {
		_, _ = svc.GetAgentInfo(ctx, req)
	})
}

func TestSquareSvc_GetAgentInfo_AgentNotFound(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentConfRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockUmHttp := httpaccmock.NewMockUmHttpAcc(ctrl)

	svc := &squareSvc{
		SvcBase:       service.NewSvcBase(),
		agentConfRepo: mockAgentConfRepo,
		umHttp:        mockUmHttp,
	}

	ctx := context.Background()
	req := &squarereq.AgentInfoReq{
		AgentID: "non-existent-agent",
	}

	notFoundErr := errors.New("sql: no rows in result set")
	mockAgentConfRepo.EXPECT().GetByID(gomock.Any(), "non-existent-agent").Return(nil, notFoundErr)

	resp, err := svc.GetAgentInfo(ctx, req)

	assert.Error(t, err)
	assert.NotNil(t, resp) // Response is created at the beginning
	assert.Contains(t, err.Error(), "sql: no rows in result set")
}

func TestSquareSvc_GetAgentInfo_GetByIDError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentConfRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockUmHttp := httpaccmock.NewMockUmHttpAcc(ctrl)

	svc := &squareSvc{
		SvcBase:       service.NewSvcBase(),
		agentConfRepo: mockAgentConfRepo,
		umHttp:        mockUmHttp,
	}

	ctx := context.Background()
	req := &squarereq.AgentInfoReq{
		AgentID: "agent-123",
	}

	dbErr := errors.New("database connection failed")
	mockAgentConfRepo.EXPECT().GetByID(gomock.Any(), "agent-123").Return(nil, dbErr)

	resp, err := svc.GetAgentInfo(ctx, req)

	assert.Error(t, err)
	assert.NotNil(t, resp) // Response is created at the beginning
	assert.Contains(t, err.Error(), "[squareSvc.GetAgentInfo]")
}

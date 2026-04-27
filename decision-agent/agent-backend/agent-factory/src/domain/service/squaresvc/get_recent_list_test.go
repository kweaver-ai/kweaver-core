package squaresvc

import (
	"context"
	"errors"
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/common"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/square/squarereq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestSquareSvc_GetRecentAgentList_PanicsWithoutReleaseRepo(t *testing.T) {
	t.Parallel()

	svc := &squareSvc{
		SvcBase: service.NewSvcBase(),
		// releaseRepo is nil
	}

	ctx := context.Background()
	req := squarereq.AgentSquareRecentAgentReq{
		PageSize: common.PageSize{
			Page: 1,
			Size: 10,
		},
	}

	assert.Panics(t, func() {
		_, _ = svc.GetRecentAgentList(ctx, req)
	})
}

func TestSquareSvc_GetRecentAgentList_DatabaseError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReleaseRepo := idbaccessmock.NewMockIReleaseRepo(ctrl)

	svc := &squareSvc{
		SvcBase:     service.NewSvcBase(),
		releaseRepo: mockReleaseRepo,
	}

	ctx := context.Background()
	req := squarereq.AgentSquareRecentAgentReq{
		PageSize: common.PageSize{
			Page: 1,
			Size: 10,
		},
	}

	dbErr := errors.New("database connection failed")
	mockReleaseRepo.EXPECT().ListRecentAgentForMarket(gomock.Any(), req).Return(nil, dbErr)

	result, err := svc.GetRecentAgentList(ctx, req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "list recent agent for square failed")
	assert.Nil(t, result)
}

func TestSquareSvc_GetRecentAgentList_EmptyList(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReleaseRepo := idbaccessmock.NewMockIReleaseRepo(ctrl)

	svc := &squareSvc{
		SvcBase:     service.NewSvcBase(),
		releaseRepo: mockReleaseRepo,
	}

	ctx := context.Background()
	req := squarereq.AgentSquareRecentAgentReq{
		PageSize: common.PageSize{
			Page: 1,
			Size: 10,
		},
	}

	// Return empty list
	mockReleaseRepo.EXPECT().ListRecentAgentForMarket(gomock.Any(), req).Return([]*dapo.RecentVisitAgentPO{}, nil)

	result, err := svc.GetRecentAgentList(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result)
}

func TestSquareSvc_GetRecentAgentList_PageOutOfRange(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReleaseRepo := idbaccessmock.NewMockIReleaseRepo(ctrl)

	svc := &squareSvc{
		SvcBase:     service.NewSvcBase(),
		releaseRepo: mockReleaseRepo,
	}

	ctx := context.Background()
	req := squarereq.AgentSquareRecentAgentReq{
		PageSize: common.PageSize{
			Page: 10,
			Size: 10,
		},
	}

	// Return some data, but not enough for page 10
	po := &dapo.RecentVisitAgentPO{}
	po.ID = "1"
	pos := []*dapo.RecentVisitAgentPO{po}
	mockReleaseRepo.EXPECT().ListRecentAgentForMarket(gomock.Any(), req).Return(pos, nil)

	result, err := svc.GetRecentAgentList(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result)
}

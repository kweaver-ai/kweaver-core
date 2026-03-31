package squaresvc

import (
	"context"
	"errors"
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestSquareSvc_IsAgentExists_Exists(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentConfRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)

	svc := &squareSvc{
		SvcBase:       service.NewSvcBase(),
		agentConfRepo: mockAgentConfRepo,
	}

	ctx := context.Background()
	agentID := "agent-123"

	mockAgentConfRepo.EXPECT().ExistsByID(gomock.Any(), agentID).Return(true, nil)

	exists, err := svc.IsAgentExists(ctx, agentID)

	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestSquareSvc_IsAgentExists_NotExists(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentConfRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)

	svc := &squareSvc{
		SvcBase:       service.NewSvcBase(),
		agentConfRepo: mockAgentConfRepo,
	}

	ctx := context.Background()
	agentID := "agent-999"

	mockAgentConfRepo.EXPECT().ExistsByID(gomock.Any(), agentID).Return(false, nil)

	exists, err := svc.IsAgentExists(ctx, agentID)

	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestSquareSvc_IsAgentExists_RepoError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentConfRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)

	svc := &squareSvc{
		SvcBase:       service.NewSvcBase(),
		agentConfRepo: mockAgentConfRepo,
	}

	ctx := context.Background()
	agentID := "agent-123"
	dbErr := errors.New("database connection failed")

	mockAgentConfRepo.EXPECT().ExistsByID(gomock.Any(), agentID).Return(false, dbErr)

	exists, err := svc.IsAgentExists(ctx, agentID)

	assert.Error(t, err)
	assert.False(t, exists)
	assert.Contains(t, err.Error(), "IsAgentExists")
}

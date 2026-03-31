package releasesvc

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driver/iv3portdriver/v3portdrivermock"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// Helper function to create context with user ID
func createUnpublishCtx(userID string) context.Context {
	visitor := &rest.Visitor{
		ID: userID,
	}

	return context.WithValue(context.Background(), cenum.VisitUserInfoCtxKey.String(), visitor) //nolint:staticcheck // SA1029
}

func TestUnPublish_AgentNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentConfigRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)

	svc := &releaseSvc{
		SvcBase:         service.NewSvcBase(),
		agentConfigRepo: mockAgentConfigRepo,
	}

	ctx := context.Background()
	agentID := "nonexistent-agent"

	mockAgentConfigRepo.EXPECT().GetByID(ctx, agentID).Return(nil, sql.ErrNoRows)

	auditLog, err := svc.UnPublish(ctx, agentID)

	assert.Error(t, err)
	assert.Empty(t, auditLog.ID)
	assert.Contains(t, err.Error(), "agent not found")
}

func TestUnPublish_NotOwner_NoPermission(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentConfigRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockPermissionSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

	agentID := "agent-123"
	ownerID := "user-456"
	currentUserID := "user-123"

	agentPo := &dapo.DataAgentPo{
		ID:        agentID,
		Name:      "Test Agent",
		CreatedBy: ownerID,
	}

	mockAgentConfigRepo.EXPECT().GetByID(gomock.Any(), agentID).Return(agentPo, nil)
	mockPermissionSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil)

	svc := &releaseSvc{
		SvcBase:         service.NewSvcBase(),
		agentConfigRepo: mockAgentConfigRepo,
		pmsSvc:          mockPermissionSvc,
	}

	ctx := createUnpublishCtx(currentUserID)
	auditLog, err := svc.UnPublish(ctx, agentID)

	assert.Error(t, err)
	// The error message comes from the permission check which happens before the owner check
	assert.NotEmpty(t, auditLog.ID)
	assert.Contains(t, err.Error(), "do not have unpublish permission")
}

func TestUnPublish_NoReleaseRecord(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentConfigRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockReleaseRepo := idbaccessmock.NewMockIReleaseRepo(ctrl)
	mockPermissionSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

	agentID := "agent-123"
	currentUserID := "user-123"

	agentPo := &dapo.DataAgentPo{
		ID:        agentID,
		Name:      "Test Agent",
		CreatedBy: currentUserID,
	}

	mockAgentConfigRepo.EXPECT().GetByID(gomock.Any(), agentID).Return(agentPo, nil)
	mockPermissionSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
	mockReleaseRepo.EXPECT().GetByAgentID(gomock.Any(), agentID).Return(nil, nil)

	svc := &releaseSvc{
		SvcBase:         service.NewSvcBase(),
		agentConfigRepo: mockAgentConfigRepo,
		releaseRepo:     mockReleaseRepo,
		pmsSvc:          mockPermissionSvc,
	}

	ctx := createUnpublishCtx(currentUserID)
	auditLog, err := svc.UnPublish(ctx, agentID)

	assert.NoError(t, err)
	assert.Equal(t, agentID, auditLog.ID)
	assert.Equal(t, "Test Agent", auditLog.Name)
}

func TestUnPublish_RepositoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentConfigRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)

	svc := &releaseSvc{
		SvcBase:         service.NewSvcBase(),
		agentConfigRepo: mockAgentConfigRepo,
	}

	ctx := context.Background()
	agentID := "agent-123"

	expectedErr := errors.New("database error")
	mockAgentConfigRepo.EXPECT().GetByID(ctx, agentID).Return(nil, expectedErr)

	auditLog, err := svc.UnPublish(ctx, agentID)

	assert.Error(t, err)
	assert.Empty(t, auditLog.ID)
	assert.Contains(t, err.Error(), "get agent config by id failed")
}

func TestUnPublish_BeginTxError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentConfigRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockReleaseRepo := idbaccessmock.NewMockIReleaseRepo(ctrl)
	mockPermissionSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

	agentID := "agent-123"
	currentUserID := "user-123"
	agentPo := &dapo.DataAgentPo{ID: agentID, Name: "Test Agent", CreatedBy: currentUserID}
	releasePo := &dapo.ReleasePO{ID: "release-123", AgentID: agentID}

	mockAgentConfigRepo.EXPECT().GetByID(gomock.Any(), agentID).Return(agentPo, nil)
	mockPermissionSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
	mockReleaseRepo.EXPECT().GetByAgentID(gomock.Any(), agentID).Return(releasePo, nil)
	mockReleaseRepo.EXPECT().BeginTx(gomock.Any()).Return(nil, errors.New("begin tx failed"))

	svc := &releaseSvc{
		SvcBase:         service.NewSvcBase(),
		agentConfigRepo: mockAgentConfigRepo,
		releaseRepo:     mockReleaseRepo,
		pmsSvc:          mockPermissionSvc,
	}

	ctx := createUnpublishCtx(currentUserID)
	_, err := svc.UnPublish(ctx, agentID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "begin transaction failed")
}

func TestUnPublish_DeleteByAgentIDError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tx, sqlMock, done := newReleaseTx(t)
	defer done()
	sqlMock.ExpectRollback()

	mockAgentConfigRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockReleaseRepo := idbaccessmock.NewMockIReleaseRepo(ctrl)
	mockPermissionSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

	agentID := "agent-123"
	currentUserID := "user-123"
	agentPo := &dapo.DataAgentPo{ID: agentID, Name: "Test Agent", CreatedBy: currentUserID}
	releasePo := &dapo.ReleasePO{ID: "release-123", AgentID: agentID}

	mockAgentConfigRepo.EXPECT().GetByID(gomock.Any(), agentID).Return(agentPo, nil)
	mockPermissionSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
	mockReleaseRepo.EXPECT().GetByAgentID(gomock.Any(), agentID).Return(releasePo, nil)
	mockReleaseRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
	mockReleaseRepo.EXPECT().DeleteByAgentID(gomock.Any(), tx, agentID).Return(errors.New("delete failed"))

	svc := &releaseSvc{
		SvcBase:         service.NewSvcBase(),
		agentConfigRepo: mockAgentConfigRepo,
		releaseRepo:     mockReleaseRepo,
		pmsSvc:          mockPermissionSvc,
	}

	ctx := createUnpublishCtx(currentUserID)
	_, err := svc.UnPublish(ctx, agentID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "delete release by agent id failed")
}

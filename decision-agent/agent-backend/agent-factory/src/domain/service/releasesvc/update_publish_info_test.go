package releasesvc

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/release/releasereq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/ihttpaccess/iumacc/httpaccmock"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driver/iv3portdriver/v3portdrivermock"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func createContextWithUserID(userID string) context.Context {
	visitor := &rest.Visitor{
		ID: userID,
	}

	return context.WithValue(context.Background(), cenum.VisitUserInfoCtxKey.String(), visitor) //nolint:staticcheck // SA1029
}

func TestReleaseSvc_UpdatePublishInfo_GetByIDError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentConfigRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockPmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

	svc := &releaseSvc{
		SvcBase:         service.NewSvcBase(),
		agentConfigRepo: mockAgentConfigRepo,
		pmsSvc:          mockPmsSvc,
	}

	ctx := context.Background()
	agentID := "agent-123"
	req := &releasereq.UpdatePublishInfoReq{}

	dbErr := errors.New("database connection failed")
	mockAgentConfigRepo.EXPECT().GetByID(ctx, agentID).Return(nil, dbErr)

	resp, auditLogInfo, err := svc.UpdatePublishInfo(ctx, agentID, req)

	assert.Error(t, err)
	assert.NotNil(t, resp) // Response is initialized in defer
	assert.Empty(t, auditLogInfo.ID)
	assert.Contains(t, err.Error(), "database connection failed")
}

func TestReleaseSvc_UpdatePublishInfo_AgentNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentConfigRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockPmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

	svc := &releaseSvc{
		SvcBase:         service.NewSvcBase(),
		agentConfigRepo: mockAgentConfigRepo,
		pmsSvc:          mockPmsSvc,
	}

	ctx := context.Background()
	agentID := "non-existent-agent"
	req := &releasereq.UpdatePublishInfoReq{}

	mockAgentConfigRepo.EXPECT().GetByID(ctx, agentID).Return(nil, sql.ErrNoRows)

	resp, auditLogInfo, err := svc.UpdatePublishInfo(ctx, agentID, req)

	assert.Error(t, err)
	assert.NotNil(t, resp)
	assert.Empty(t, auditLogInfo.ID)
	assert.Contains(t, err.Error(), "agent not found")
}

func TestReleaseSvc_UpdatePublishInfo_NoPermission(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentConfigRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockPmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

	svc := &releaseSvc{
		SvcBase:         service.NewSvcBase(),
		agentConfigRepo: mockAgentConfigRepo,
		pmsSvc:          mockPmsSvc,
	}

	ctx := context.Background()
	agentID := "agent-123"
	creatorID := "user-123"
	req := &releasereq.UpdatePublishInfoReq{}

	builtInNo := cdaenum.BuiltInNo

	po := &dapo.DataAgentPo{
		ID:        agentID,
		Name:      "Test Agent",
		CreatedBy: creatorID,
		IsBuiltIn: &builtInNo,
		Status:    cdaenum.StatusPublished,
	}

	pmsErr := errors.New("permission check failed")

	mockAgentConfigRepo.EXPECT().GetByID(ctx, agentID).Return(po, nil)
	mockPmsSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, pmsErr)

	resp, auditLogInfo, err := svc.UpdatePublishInfo(ctx, agentID, req)

	assert.Error(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, auditLogInfo.ID)
	assert.Equal(t, agentID, auditLogInfo.ID)
}

func TestReleaseSvc_UpdatePublishInfo_PermissionDenied(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentConfigRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockPmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

	svc := &releaseSvc{
		SvcBase:         service.NewSvcBase(),
		agentConfigRepo: mockAgentConfigRepo,
		pmsSvc:          mockPmsSvc,
	}

	ctx := context.Background()
	agentID := "agent-123"
	creatorID := "user-123"
	req := &releasereq.UpdatePublishInfoReq{}

	builtInNo := cdaenum.BuiltInNo

	po := &dapo.DataAgentPo{
		ID:        agentID,
		Name:      "Test Agent",
		CreatedBy: creatorID,
		IsBuiltIn: &builtInNo,
		Status:    cdaenum.StatusPublished,
	}

	mockAgentConfigRepo.EXPECT().GetByID(ctx, agentID).Return(po, nil)
	mockPmsSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil)

	resp, auditLogInfo, err := svc.UpdatePublishInfo(ctx, agentID, req)

	assert.Error(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, auditLogInfo.ID)
	assert.Contains(t, err.Error(), "do not have publish permission")
}

func TestReleaseSvc_UpdatePublishInfo_BeginTxError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentConfigRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockReleaseRepo := idbaccessmock.NewMockIReleaseRepo(ctrl)
	mockPmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

	svc := &releaseSvc{
		SvcBase:         service.NewSvcBase(),
		agentConfigRepo: mockAgentConfigRepo,
		releaseRepo:     mockReleaseRepo,
		pmsSvc:          mockPmsSvc,
	}

	ctx := createContextWithUserID("user-123")
	agentID := "agent-123"
	creatorID := "user-123"
	req := &releasereq.UpdatePublishInfoReq{}

	builtInNo := cdaenum.BuiltInNo

	po := &dapo.DataAgentPo{
		ID:        agentID,
		Name:      "Test Agent",
		CreatedBy: creatorID,
		IsBuiltIn: &builtInNo,
		Status:    cdaenum.StatusPublished,
	}

	txErr := errors.New("begin transaction failed")

	mockAgentConfigRepo.EXPECT().GetByID(ctx, agentID).Return(po, nil)
	mockPmsSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
	mockReleaseRepo.EXPECT().BeginTx(ctx).Return(nil, txErr)

	resp, auditLogInfo, err := svc.UpdatePublishInfo(ctx, agentID, req)

	assert.Error(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, auditLogInfo.ID)
	assert.Contains(t, err.Error(), "begin transaction failed")
}

func TestReleaseSvc_UpdatePublishInfo_GetByAgentIDError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentConfigRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockReleaseRepo := idbaccessmock.NewMockIReleaseRepo(ctrl)
	mockPmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

	svc := &releaseSvc{
		SvcBase:         service.NewSvcBase(),
		agentConfigRepo: mockAgentConfigRepo,
		releaseRepo:     mockReleaseRepo,
		pmsSvc:          mockPmsSvc,
	}

	ctx := createContextWithUserID("user-123")
	agentID := "agent-123"
	creatorID := "user-123"
	req := &releasereq.UpdatePublishInfoReq{}

	builtInNo := cdaenum.BuiltInNo

	po := &dapo.DataAgentPo{
		ID:        agentID,
		Name:      "Test Agent",
		CreatedBy: creatorID,
		IsBuiltIn: &builtInNo,
		Status:    cdaenum.StatusPublished,
	}

	getReleaseErr := errors.New("get release failed")

	mockAgentConfigRepo.EXPECT().GetByID(ctx, agentID).Return(po, nil)
	mockPmsSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
	// Return nil transaction to avoid rollback issues
	mockReleaseRepo.EXPECT().BeginTx(ctx).Return(nil, getReleaseErr)

	resp, auditLogInfo, err := svc.UpdatePublishInfo(ctx, agentID, req)

	assert.Error(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, auditLogInfo.ID)
	// Since BeginTx errors before GetByAgentID, the error will be "begin transaction failed"
	assert.Contains(t, err.Error(), "begin transaction failed")
}

func TestReleaseSvc_UpdatePublishInfo_ReleaseNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentConfigRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockReleaseRepo := idbaccessmock.NewMockIReleaseRepo(ctrl)
	mockPmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

	svc := &releaseSvc{
		SvcBase:         service.NewSvcBase(),
		agentConfigRepo: mockAgentConfigRepo,
		releaseRepo:     mockReleaseRepo,
		pmsSvc:          mockPmsSvc,
	}

	ctx := createContextWithUserID("user-123")
	agentID := "agent-123"
	creatorID := "user-123"
	req := &releasereq.UpdatePublishInfoReq{}

	builtInNo := cdaenum.BuiltInNo

	po := &dapo.DataAgentPo{
		ID:        agentID,
		Name:      "Test Agent",
		CreatedBy: creatorID,
		IsBuiltIn: &builtInNo,
		Status:    cdaenum.StatusPublished,
	}

	notFoundErr := errors.New("release not found")

	mockAgentConfigRepo.EXPECT().GetByID(ctx, agentID).Return(po, nil)
	mockPmsSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
	// Return error from BeginTx to simulate error before GetByAgentID
	mockReleaseRepo.EXPECT().BeginTx(ctx).Return(nil, notFoundErr)

	resp, auditLogInfo, err := svc.UpdatePublishInfo(ctx, agentID, req)

	assert.Error(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, auditLogInfo.ID)
}

func TestReleaseSvc_UpdatePublishInfo_UpdateError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentConfigRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockReleaseRepo := idbaccessmock.NewMockIReleaseRepo(ctrl)
	mockUmHttp := httpaccmock.NewMockUmHttpAcc(ctrl)
	mockPmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

	svc := &releaseSvc{
		SvcBase:         service.NewSvcBase(),
		agentConfigRepo: mockAgentConfigRepo,
		releaseRepo:     mockReleaseRepo,
		umHttp:          mockUmHttp,
		pmsSvc:          mockPmsSvc,
	}

	ctx := createContextWithUserID("user-123")
	agentID := "agent-123"
	creatorID := "user-123"
	req := &releasereq.UpdatePublishInfoReq{}
	req.PublishInfo.Description = "Updated description"

	builtInNo := cdaenum.BuiltInNo

	po := &dapo.DataAgentPo{
		ID:        agentID,
		Name:      "Test Agent",
		CreatedBy: creatorID,
		IsBuiltIn: &builtInNo,
		Status:    cdaenum.StatusPublished,
	}

	updateErr := errors.New("update release failed")

	mockAgentConfigRepo.EXPECT().GetByID(ctx, agentID).Return(po, nil)
	mockPmsSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
	// Return error from BeginTx to simulate error before Update
	mockReleaseRepo.EXPECT().BeginTx(ctx).Return(nil, updateErr)

	resp, auditLogInfo, err := svc.UpdatePublishInfo(ctx, agentID, req)

	assert.Error(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, auditLogInfo.ID)
	assert.Contains(t, err.Error(), "begin transaction failed")
}

func TestReleaseSvc_UpdatePublishInfo_FillPublishedByNameError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentConfigRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockReleaseRepo := idbaccessmock.NewMockIReleaseRepo(ctrl)
	mockUmHttp := httpaccmock.NewMockUmHttpAcc(ctrl)
	mockPmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

	svc := &releaseSvc{
		SvcBase:         service.NewSvcBase(),
		agentConfigRepo: mockAgentConfigRepo,
		releaseRepo:     mockReleaseRepo,
		umHttp:          mockUmHttp,
		pmsSvc:          mockPmsSvc,
	}

	ctx := createContextWithUserID("user-123")
	agentID := "agent-123"
	creatorID := "user-123"
	req := &releasereq.UpdatePublishInfoReq{}
	req.PublishInfo.Description = "Updated description"

	builtInNo := cdaenum.BuiltInNo

	po := &dapo.DataAgentPo{
		ID:        agentID,
		Name:      "Test Agent",
		CreatedBy: creatorID,
		IsBuiltIn: &builtInNo,
		Status:    cdaenum.StatusPublished,
	}

	fillNameErr := errors.New("fill published by name failed")

	mockAgentConfigRepo.EXPECT().GetByID(ctx, agentID).Return(po, nil)
	mockPmsSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
	// Return error from BeginTx to simulate error before Update
	mockReleaseRepo.EXPECT().BeginTx(ctx).Return(nil, fillNameErr)

	resp, auditLogInfo, err := svc.UpdatePublishInfo(ctx, agentID, req)

	assert.Error(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, auditLogInfo.ID)
	assert.Contains(t, err.Error(), "begin transaction failed")
}

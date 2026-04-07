package tplsvc

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent_tpl/agenttplreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp/cmpmock"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driver/iv3portdriver/v3portdrivermock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestDataAgentTplSvc_UpdatePublishInfo_PanicsWithoutAgentTplRepo(t *testing.T) {
	t.Parallel()

	svc := &dataAgentTplSvc{
		SvcBase: service.NewSvcBase(),
	}

	ctx := context.Background()
	tplID := int64(123)
	req := &agenttplreq.UpdatePublishInfoReq{}

	assert.Panics(t, func() {
		_, _, _ = svc.UpdatePublishInfo(ctx, req, tplID)
	})
}

func TestDataAgentTplSvc_UpdatePublishInfo_PermissionDenied(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)
	mockPmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

	svc := &dataAgentTplSvc{
		SvcBase:      service.NewSvcBase(),
		agentTplRepo: mockAgentTplRepo,
		logger:       mockLogger,
		pmsSvc:       mockPmsSvc,
	}

	ctx := context.Background()
	tplID := int64(123)
	req := &agenttplreq.UpdatePublishInfoReq{}

	// Permission denied
	mockPmsSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil)

	resp, auditLogInfo, err := svc.UpdatePublishInfo(ctx, req, tplID)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Empty(t, auditLogInfo.ID)
	assert.Contains(t, err.Error(), "do not have publish permission")
}

func TestDataAgentTplSvc_UpdatePublishInfo_TemplateNotFound(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)
	mockPmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

	svc := &dataAgentTplSvc{
		SvcBase:      service.NewSvcBase(),
		agentTplRepo: mockAgentTplRepo,
		logger:       mockLogger,
		pmsSvc:       mockPmsSvc,
	}

	ctx := context.Background()
	tplID := int64(999)
	req := &agenttplreq.UpdatePublishInfoReq{}

	// Permission granted but template not found
	mockPmsSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
	mockAgentTplRepo.EXPECT().GetByID(gomock.Any(), tplID).Return(nil, sql.ErrNoRows)

	resp, auditLogInfo, err := svc.UpdatePublishInfo(ctx, req, tplID)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Empty(t, auditLogInfo.ID)
	assert.Contains(t, err.Error(), "模板不存在")
}

func TestDataAgentTplSvc_UpdatePublishInfo_GetByIDError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)
	mockPmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

	svc := &dataAgentTplSvc{
		SvcBase:      service.NewSvcBase(),
		agentTplRepo: mockAgentTplRepo,
		logger:       mockLogger,
		pmsSvc:       mockPmsSvc,
	}

	ctx := context.Background()
	tplID := int64(123)
	req := &agenttplreq.UpdatePublishInfoReq{}

	dbErr := errors.New("database connection failed")

	// Permission granted but GetByID fails
	mockPmsSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
	mockAgentTplRepo.EXPECT().GetByID(gomock.Any(), tplID).Return(nil, dbErr)

	resp, auditLogInfo, err := svc.UpdatePublishInfo(ctx, req, tplID)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Empty(t, auditLogInfo.ID)
}

func TestDataAgentTplSvc_UpdatePublishInfo_PublishedTplNotFound(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
	mockPublishedTplRepo := idbaccessmock.NewMockIPublishedTplRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)
	mockPmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

	svc := &dataAgentTplSvc{
		SvcBase:          service.NewSvcBase(),
		agentTplRepo:     mockAgentTplRepo,
		publishedTplRepo: mockPublishedTplRepo,
		logger:           mockLogger,
		pmsSvc:           mockPmsSvc,
	}

	ctx := context.Background()
	tplID := int64(123)
	req := &agenttplreq.UpdatePublishInfoReq{}

	po := &dapo.DataAgentTplPo{
		ID:   123,
		Name: "Test Template",
	}

	// Permission granted, template found, but published template not found
	mockPmsSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
	mockAgentTplRepo.EXPECT().GetByID(gomock.Any(), tplID).Return(po, nil)
	mockPublishedTplRepo.EXPECT().GetByTplID(gomock.Any(), tplID).Return(nil, sql.ErrNoRows)

	resp, _, err := svc.UpdatePublishInfo(ctx, req, tplID)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "此已发布模板不存在")
}

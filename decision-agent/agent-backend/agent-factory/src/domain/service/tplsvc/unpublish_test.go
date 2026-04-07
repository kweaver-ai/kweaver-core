package tplsvc

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp/cmpmock"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driver/iv3portdriver/v3portdrivermock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestDataAgentTplSvc_Unpublish_PanicsWithoutAgentTplRepo(t *testing.T) {
	t.Parallel()

	svc := &dataAgentTplSvc{
		SvcBase: service.NewSvcBase(),
	}

	ctx := context.Background()
	tplID := int64(123)

	assert.Panics(t, func() {
		_, _ = svc.Unpublish(ctx, tplID)
	})
}

func TestDataAgentTplSvc_Unpublish_PermissionDenied(t *testing.T) {
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

	// Permission denied
	mockPmsSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil)

	auditLogInfo, err := svc.Unpublish(ctx, tplID)

	assert.Error(t, err)
	assert.Empty(t, auditLogInfo.ID)
	assert.Contains(t, err.Error(), "do not have unpublish permission")
}

func TestDataAgentTplSvc_Unpublish_TemplateNotFound(t *testing.T) {
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

	// Permission granted but template not found
	mockPmsSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
	mockAgentTplRepo.EXPECT().GetByID(gomock.Any(), tplID).Return(nil, sql.ErrNoRows)

	auditLogInfo, err := svc.Unpublish(ctx, tplID)

	assert.Error(t, err)
	assert.Empty(t, auditLogInfo.ID)
	assert.Contains(t, err.Error(), "模板不存在")
}

func TestDataAgentTplSvc_Unpublish_GetByIDError(t *testing.T) {
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

	dbErr := errors.New("database connection failed")

	// Permission granted but GetByID fails
	mockPmsSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
	mockAgentTplRepo.EXPECT().GetByID(gomock.Any(), tplID).Return(nil, dbErr)

	auditLogInfo, err := svc.Unpublish(ctx, tplID)

	assert.Error(t, err)
	assert.Empty(t, auditLogInfo.ID)
}

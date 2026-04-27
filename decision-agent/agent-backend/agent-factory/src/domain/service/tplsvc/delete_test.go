package tplsvc

import (
	"context"
	"errors"
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/cmp/icmp/cmpmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestDataAgentTplSvc_Delete_ZeroIDReturnsError(t *testing.T) {
	t.Parallel()

	svc := &dataAgentTplSvc{
		SvcBase: service.NewSvcBase(),
	}

	ctx := context.Background()
	id := int64(0)
	uid := "user-123"

	auditLogInfo, err := svc.Delete(ctx, id, uid, false)

	assert.Error(t, err)
	assert.Empty(t, auditLogInfo.ID)
	assert.Empty(t, auditLogInfo.Name)
	assert.Contains(t, err.Error(), "模板ID不能为空")
}

func TestDataAgentTplSvc_Delete_PanicsWithoutAgentTplRepo(t *testing.T) {
	t.Parallel()

	svc := &dataAgentTplSvc{
		SvcBase: service.NewSvcBase(),
	}

	ctx := context.Background()
	id := int64(123)
	uid := "user-123"

	assert.Panics(t, func() {
		_, _ = svc.Delete(ctx, id, uid, false)
	})
}

func TestDataAgentTplSvc_Delete_TemplateNotExists(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentTplSvc{
		SvcBase:      service.NewSvcBase(),
		agentTplRepo: mockAgentTplRepo,
		logger:       mockLogger,
	}

	ctx := context.Background()
	id := int64(999)
	uid := "user-123"

	mockAgentTplRepo.EXPECT().ExistsByID(gomock.Any(), id).Return(false, nil)

	auditLogInfo, err := svc.Delete(ctx, id, uid, false)

	assert.Error(t, err)
	assert.Empty(t, auditLogInfo.ID)
	assert.Contains(t, err.Error(), "模板不存在")
}

func TestDataAgentTplSvc_Delete_ExistsByIDError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentTplSvc{
		SvcBase:      service.NewSvcBase(),
		agentTplRepo: mockAgentTplRepo,
		logger:       mockLogger,
	}

	ctx := context.Background()
	id := int64(123)
	uid := "user-123"

	dbErr := errors.New("database connection failed")
	mockAgentTplRepo.EXPECT().ExistsByID(gomock.Any(), id).Return(false, dbErr)

	auditLogInfo, err := svc.Delete(ctx, id, uid, false)

	assert.Error(t, err)
	assert.Empty(t, auditLogInfo.ID)
}

func TestDataAgentTplSvc_Delete_GetByIDError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentTplSvc{
		SvcBase:      service.NewSvcBase(),
		agentTplRepo: mockAgentTplRepo,
		logger:       mockLogger,
	}

	ctx := context.Background()
	id := int64(123)
	uid := "user-123"

	// Exists returns true but GetByID fails
	mockAgentTplRepo.EXPECT().ExistsByID(gomock.Any(), id).Return(true, nil)

	dbErr := errors.New("database connection failed")
	mockAgentTplRepo.EXPECT().GetByID(gomock.Any(), id).Return(nil, dbErr)

	auditLogInfo, err := svc.Delete(ctx, id, uid, false)

	assert.Error(t, err)
	assert.Empty(t, auditLogInfo.ID)
}

func TestDataAgentTplSvc_Delete_TemplatePublished(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentTplSvc{
		SvcBase:      service.NewSvcBase(),
		agentTplRepo: mockAgentTplRepo,
		logger:       mockLogger,
	}

	ctx := context.Background()
	id := int64(123)
	uid := "user-123"

	po := &dapo.DataAgentTplPo{
		ID:     123,
		Name:   "Published Template",
		Status: cdaenum.StatusPublished,
	}

	mockAgentTplRepo.EXPECT().ExistsByID(gomock.Any(), id).Return(true, nil)
	mockAgentTplRepo.EXPECT().GetByID(gomock.Any(), id).Return(po, nil)

	auditLogInfo, err := svc.Delete(ctx, id, uid, false)

	assert.Error(t, err)
	assert.NotEmpty(t, auditLogInfo.ID)
	assert.Contains(t, err.Error(), "模板已发布，无法删除")
}

func TestDataAgentTplSvc_Delete_NotOwner(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentTplSvc{
		SvcBase:      service.NewSvcBase(),
		agentTplRepo: mockAgentTplRepo,
		logger:       mockLogger,
	}

	ctx := context.Background()
	id := int64(123)
	uid := "user-123"
	creatorID := "other-user"

	builtInNo := cdaenum.BuiltInNo

	po := &dapo.DataAgentTplPo{
		ID:        123,
		Name:      "Test Template",
		Status:    cdaenum.StatusUnpublished,
		CreatedBy: creatorID,
		IsBuiltIn: &builtInNo,
	}

	mockAgentTplRepo.EXPECT().ExistsByID(gomock.Any(), id).Return(true, nil)
	mockAgentTplRepo.EXPECT().GetByID(gomock.Any(), id).Return(po, nil)

	auditLogInfo, err := svc.Delete(ctx, id, uid, false)

	assert.Error(t, err)
	assert.NotEmpty(t, auditLogInfo.ID)
	assert.Contains(t, err.Error(), "无权限删除，非创建人")
}

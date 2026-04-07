package v3agentconfigsvc

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp/cmpmock"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/ihttpaccess/ibizdomainacc/bizdomainaccmock"
	"github.com/stretchr/testify/assert"
)

func TestDataAgentConfigSvc_Delete_PanicsWithoutAgentConfRepo(t *testing.T) {
	svc := &dataAgentConfigSvc{
		SvcBase: service.NewSvcBase(),
	}

	ctx := context.Background()
	id := "agent-123"
	uid := "user-123"

	assert.Panics(t, func() {
		_, _ = svc.Delete(ctx, id, uid, false)
	})
}

func TestDataAgentConfigSvc_Delete_AgentNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentConfRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:       service.NewSvcBase(),
		agentConfRepo: mockAgentConfRepo,
		logger:        mockLogger,
	}

	ctx := context.Background()
	id := "non-existent-agent"
	uid := "user-123"

	mockAgentConfRepo.EXPECT().GetByID(gomock.Any(), id).Return(nil, sql.ErrNoRows)

	auditLogInfo, err := svc.Delete(ctx, id, uid, false)

	assert.Error(t, err)
	assert.Empty(t, auditLogInfo.ID)
	assert.Contains(t, err.Error(), "数据智能体配置不存在")
}

func TestDataAgentConfigSvc_Delete_GetByIDError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentConfRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:       service.NewSvcBase(),
		agentConfRepo: mockAgentConfRepo,
		logger:        mockLogger,
	}

	ctx := context.Background()
	id := "agent-123"
	uid := "user-123"

	dbErr := errors.New("database connection failed")
	mockAgentConfRepo.EXPECT().GetByID(gomock.Any(), id).Return(nil, dbErr)

	auditLogInfo, err := svc.Delete(ctx, id, uid, false)

	assert.Error(t, err)
	assert.Empty(t, auditLogInfo.ID)
}

func TestDataAgentConfigSvc_Delete_AgentPublished(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentConfRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:       service.NewSvcBase(),
		agentConfRepo: mockAgentConfRepo,
		logger:        mockLogger,
	}

	ctx := context.Background()
	id := "agent-123"
	uid := "user-123"

	po := &dapo.DataAgentPo{
		ID:     id,
		Name:   "Published Agent",
		Status: cdaenum.StatusPublished,
	}

	mockAgentConfRepo.EXPECT().GetByID(gomock.Any(), id).Return(po, nil)

	auditLogInfo, err := svc.Delete(ctx, id, uid, false)

	assert.Error(t, err)
	assert.NotEmpty(t, auditLogInfo.ID)
	assert.Contains(t, err.Error(), "数据智能体配置已发布，无法删除")
}

func TestDataAgentConfigSvc_Delete_NotOwner(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentConfRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:       service.NewSvcBase(),
		agentConfRepo: mockAgentConfRepo,
		logger:        mockLogger,
	}

	ctx := context.Background()
	id := "agent-123"
	uid := "user-123"
	creatorID := "other-user"

	builtInNo := cdaenum.BuiltInNo

	po := &dapo.DataAgentPo{
		ID:        id,
		Name:      "Test Agent",
		Status:    cdaenum.StatusUnpublished,
		CreatedBy: creatorID,
		IsBuiltIn: &builtInNo,
	}

	mockAgentConfRepo.EXPECT().GetByID(gomock.Any(), id).Return(po, nil)

	auditLogInfo, err := svc.Delete(ctx, id, uid, false)

	assert.Error(t, err)
	assert.NotEmpty(t, auditLogInfo.ID)
	assert.Contains(t, err.Error(), "访问者不是创建者，无法删除")
}

func TestDataAgentConfigSvc_Delete_BeginTxError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentConfRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockBdAgentRelRepo := idbaccessmock.NewMockIBizDomainAgentRelRepo(ctrl)
	mockBizDomainHttp := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:        service.NewSvcBase(),
		agentConfRepo:  mockAgentConfRepo,
		bdAgentRelRepo: mockBdAgentRelRepo,
		bizDomainHttp:  mockBizDomainHttp,
		logger:         mockLogger,
	}

	ctx := context.Background()
	id := "agent-123"
	uid := "user-123"

	builtInNo := cdaenum.BuiltInNo

	po := &dapo.DataAgentPo{
		ID:        id,
		Name:      "Test Agent",
		Status:    cdaenum.StatusUnpublished,
		CreatedBy: uid,
		IsBuiltIn: &builtInNo,
	}

	txErr := errors.New("transaction begin failed")

	mockAgentConfRepo.EXPECT().GetByID(gomock.Any(), id).Return(po, nil)
	mockAgentConfRepo.EXPECT().BeginTx(gomock.Any()).Return(nil, txErr)
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

	auditLogInfo, err := svc.Delete(ctx, id, uid, false)

	assert.Error(t, err)
	assert.NotEmpty(t, auditLogInfo.ID)
}

func TestDataAgentConfigSvc_Delete_BuiltInAgent_CannotDelete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentConfRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:       service.NewSvcBase(),
		agentConfRepo: mockAgentConfRepo,
		logger:        mockLogger,
	}

	builtInYes := cdaenum.BuiltInYes
	po := &dapo.DataAgentPo{
		ID:        "built-in-agent",
		Name:      "BuiltIn",
		Status:    cdaenum.StatusUnpublished,
		CreatedBy: "user-1", // same as uid → passes owner check
		IsBuiltIn: &builtInYes,
	}

	mockAgentConfRepo.EXPECT().GetByID(gomock.Any(), "built-in-agent").Return(po, nil)

	_, err := svc.Delete(context.Background(), "built-in-agent", "user-1", false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "内置数据智能体不可删除")
}

func TestDataAgentConfigSvc_Delete_PrivateAPI_BeginTxError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentConfRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:       service.NewSvcBase(),
		agentConfRepo: mockAgentConfRepo,
		logger:        mockLogger,
	}

	builtInNo := cdaenum.BuiltInNo
	po := &dapo.DataAgentPo{
		ID:        "agent-private",
		Name:      "Agent",
		Status:    cdaenum.StatusUnpublished,
		CreatedBy: "other-user",
		IsBuiltIn: &builtInNo,
	}

	// isPrivate=true → 跳过权限检查，直接进入 BeginTx
	mockAgentConfRepo.EXPECT().GetByID(gomock.Any(), "agent-private").Return(po, nil)
	mockAgentConfRepo.EXPECT().BeginTx(gomock.Any()).Return(nil, errors.New("tx err"))
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

	_, err := svc.Delete(context.Background(), "agent-private", "user-1", true)
	assert.Error(t, err)
}

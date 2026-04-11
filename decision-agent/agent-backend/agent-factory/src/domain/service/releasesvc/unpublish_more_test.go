package releasesvc

import (
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/ihttpaccess/iauthzacc/authzaccmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driver/iv3portdriver/v3portdrivermock"
	"github.com/stretchr/testify/assert"
)

// TestUnPublish_NotOwner_isHasOtherPmsError 非创建者，检查二次权限时出错
func TestUnPublish_NotOwner_isHasOtherPmsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentConfigRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockPermissionSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

	agentID := "agent-123"
	ownerID := "owner-user"
	currentUserID := "current-user"

	// 需要内置 Agent 才能让 isHasPubOrUnPubPms 继续走第二次权限检查
	// isHasPubOrUnPubPms 流程: hasPubPms=true, isOwner=false, isBuiltIn=true
	// → 调用第二次 GetSingleMgmtPermission (BuiltInAgentMgmt) → true
	// → isHasUnPublishPermission 返回 true
	// → 进入 agentCfgPo.CreatedBy != userID 分支
	// → 调用 isHasUnpublishOtherUserAgentPermission → error
	builtInYes := cdaenum.BuiltInYes
	agentPo := &dapo.DataAgentPo{
		ID:        agentID,
		Name:      "Test Agent",
		CreatedBy: ownerID, // 非当前用户创建
		IsBuiltIn: &builtInYes,
	}

	mockAgentConfigRepo.EXPECT().GetByID(gomock.Any(), agentID).Return(agentPo, nil)
	// isHasPubOrUnPubPms: 1st GetSingleMgmtPermission (unpublish基本权限) → true
	// isHasPubOrUnPubPms: 2nd GetSingleMgmtPermission (内置Agent权限) → true
	// isHasUnpublishOtherUserAgentPermission: 3rd GetSingleMgmtPermission → error
	gomock.InOrder(
		mockPermissionSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil),
		mockPermissionSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil),
		mockPermissionSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, errors.New("pms check error")),
	)

	svc := &releaseSvc{
		SvcBase:         service.NewSvcBase(),
		agentConfigRepo: mockAgentConfigRepo,
		pmsSvc:          mockPermissionSvc,
	}

	ctx := createUnpublishCtx(currentUserID)
	_, err := svc.UnPublish(ctx, agentID)
	assert.Error(t, err)
}

// TestUnPublish_NotOwner_NoOtherPms 非创建者 且无跨用户取消发布权限
func TestUnPublish_NotOwner_NoOtherPms(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentConfigRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockPermissionSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

	agentID := "agent-123"
	ownerID := "owner-user"
	currentUserID := "current-user"

	builtInYes := cdaenum.BuiltInYes
	agentPo := &dapo.DataAgentPo{
		ID:        agentID,
		Name:      "Test Agent",
		CreatedBy: ownerID,
		IsBuiltIn: &builtInYes,
	}

	mockAgentConfigRepo.EXPECT().GetByID(gomock.Any(), agentID).Return(agentPo, nil)
	gomock.InOrder(
		mockPermissionSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil),
		mockPermissionSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil),
		mockPermissionSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil),
	)

	svc := &releaseSvc{
		SvcBase:         service.NewSvcBase(),
		agentConfigRepo: mockAgentConfigRepo,
		pmsSvc:          mockPermissionSvc,
	}

	ctx := createUnpublishCtx(currentUserID)
	_, err := svc.UnPublish(ctx, agentID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "无取消发布的权限")
}

// TestUnPublish_NotOwner_WithPermission 非创建者 但有跨用户取消发布权限，后续 GetByAgentID 失败
func TestUnPublish_NotOwner_WithPermission_GetReleaseError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentConfigRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockReleaseRepo := idbaccessmock.NewMockIReleaseRepo(ctrl)
	mockPermissionSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

	agentID := "agent-123"
	ownerID := "owner-user"
	currentUserID := "current-user"

	builtInYes := cdaenum.BuiltInYes
	agentPo := &dapo.DataAgentPo{
		ID:        agentID,
		Name:      "Test Agent",
		CreatedBy: ownerID,
		IsBuiltIn: &builtInYes,
	}

	mockAgentConfigRepo.EXPECT().GetByID(gomock.Any(), agentID).Return(agentPo, nil)
	gomock.InOrder(
		mockPermissionSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil),
		mockPermissionSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil),
		mockPermissionSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil),
	)
	mockReleaseRepo.EXPECT().GetByAgentID(gomock.Any(), agentID).Return(nil, errors.New("db error getting release"))

	svc := &releaseSvc{
		SvcBase:         service.NewSvcBase(),
		agentConfigRepo: mockAgentConfigRepo,
		releaseRepo:     mockReleaseRepo,
		pmsSvc:          mockPermissionSvc,
	}

	ctx := createUnpublishCtx(currentUserID)
	_, err := svc.UnPublish(ctx, agentID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "get release by agent id failed")
}

// TestUnPublish_DelCategoryRelError 成功开启事务但删除分类关联失败
func TestUnPublish_DelCategoryRelError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tx, sqlMock, done := newReleaseTx(t)
	defer done()
	sqlMock.ExpectRollback()

	mockAgentConfigRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockReleaseRepo := idbaccessmock.NewMockIReleaseRepo(ctrl)
	mockPermissionSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)
	mockCategoryRel := idbaccessmock.NewMockIReleaseCategoryRelRepo(ctrl)

	agentID := "agent-123"
	currentUserID := "user-123"
	agentPo := &dapo.DataAgentPo{ID: agentID, Name: "Test Agent", CreatedBy: currentUserID}
	releasePo := &dapo.ReleasePO{ID: "release-123", AgentID: agentID}

	mockAgentConfigRepo.EXPECT().GetByID(gomock.Any(), agentID).Return(agentPo, nil)
	mockPermissionSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
	mockReleaseRepo.EXPECT().GetByAgentID(gomock.Any(), agentID).Return(releasePo, nil)
	mockReleaseRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
	mockReleaseRepo.EXPECT().DeleteByAgentID(gomock.Any(), tx, agentID).Return(nil)
	mockCategoryRel.EXPECT().DelByReleaseID(gomock.Any(), tx, "release-123").Return(errors.New("category del failed"))

	svc := &releaseSvc{
		SvcBase:                service.NewSvcBase(),
		agentConfigRepo:        mockAgentConfigRepo,
		releaseRepo:            mockReleaseRepo,
		pmsSvc:                 mockPermissionSvc,
		releaseCategoryRelRepo: mockCategoryRel,
	}

	ctx := createUnpublishCtx(currentUserID)
	_, err := svc.UnPublish(ctx, agentID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "delete release category rel by release id failed")
}

// TestUnPublish_DelPermissionError 删除权限关联失败
func TestUnPublish_DelPermissionError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tx, sqlMock, done := newReleaseTx(t)
	defer done()
	sqlMock.ExpectRollback()

	mockAgentConfigRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockReleaseRepo := idbaccessmock.NewMockIReleaseRepo(ctrl)
	mockPermissionSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)
	mockCategoryRel := idbaccessmock.NewMockIReleaseCategoryRelRepo(ctrl)
	mockPermRepo := idbaccessmock.NewMockIReleasePermissionRepo(ctrl)

	agentID := "agent-123"
	currentUserID := "user-123"
	agentPo := &dapo.DataAgentPo{ID: agentID, Name: "Test Agent", CreatedBy: currentUserID}
	releasePo := &dapo.ReleasePO{ID: "release-123", AgentID: agentID}

	mockAgentConfigRepo.EXPECT().GetByID(gomock.Any(), agentID).Return(agentPo, nil)
	mockPermissionSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
	mockReleaseRepo.EXPECT().GetByAgentID(gomock.Any(), agentID).Return(releasePo, nil)
	mockReleaseRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
	mockReleaseRepo.EXPECT().DeleteByAgentID(gomock.Any(), tx, agentID).Return(nil)
	mockCategoryRel.EXPECT().DelByReleaseID(gomock.Any(), tx, "release-123").Return(nil)
	mockPermRepo.EXPECT().DelByReleaseID(gomock.Any(), tx, "release-123").Return(errors.New("perm del failed"))

	svc := &releaseSvc{
		SvcBase:                service.NewSvcBase(),
		agentConfigRepo:        mockAgentConfigRepo,
		releaseRepo:            mockReleaseRepo,
		pmsSvc:                 mockPermissionSvc,
		releaseCategoryRelRepo: mockCategoryRel,
		releasePermissionRepo:  mockPermRepo,
	}

	ctx := createUnpublishCtx(currentUserID)
	_, err := svc.UnPublish(ctx, agentID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "delete release permission by release id failed")
}

// TestUnPublish_UpdateStatusError 更新 Agent 状态失败
func TestUnPublish_UpdateStatusError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tx, sqlMock, done := newReleaseTx(t)
	defer done()
	sqlMock.ExpectRollback()

	mockAgentConfigRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockReleaseRepo := idbaccessmock.NewMockIReleaseRepo(ctrl)
	mockPermissionSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)
	mockCategoryRel := idbaccessmock.NewMockIReleaseCategoryRelRepo(ctrl)
	mockPermRepo := idbaccessmock.NewMockIReleasePermissionRepo(ctrl)

	agentID := "agent-123"
	currentUserID := "user-123"
	agentPo := &dapo.DataAgentPo{ID: agentID, Name: "Test Agent", CreatedBy: currentUserID}
	releasePo := &dapo.ReleasePO{ID: "release-123", AgentID: agentID}

	mockAgentConfigRepo.EXPECT().GetByID(gomock.Any(), agentID).Return(agentPo, nil)
	mockPermissionSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
	mockReleaseRepo.EXPECT().GetByAgentID(gomock.Any(), agentID).Return(releasePo, nil)
	mockReleaseRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
	mockReleaseRepo.EXPECT().DeleteByAgentID(gomock.Any(), tx, agentID).Return(nil)
	mockCategoryRel.EXPECT().DelByReleaseID(gomock.Any(), tx, "release-123").Return(nil)
	mockPermRepo.EXPECT().DelByReleaseID(gomock.Any(), tx, "release-123").Return(nil)
	mockAgentConfigRepo.EXPECT().UpdateStatus(gomock.Any(), tx, cdaenum.StatusUnpublished, agentID, "").Return(errors.New("update status failed"))

	svc := &releaseSvc{
		SvcBase:                service.NewSvcBase(),
		agentConfigRepo:        mockAgentConfigRepo,
		releaseRepo:            mockReleaseRepo,
		pmsSvc:                 mockPermissionSvc,
		releaseCategoryRelRepo: mockCategoryRel,
		releasePermissionRepo:  mockPermRepo,
	}

	ctx := createUnpublishCtx(currentUserID)
	_, err := svc.UnPublish(ctx, agentID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update agent status to unpublished failed")
}

// TestUnPublish_RemoveUsePmsError 调用 removeUsePmsByHTTPAcc 失败
func TestUnPublish_RemoveUsePmsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tx, sqlMock, done := newReleaseTx(t)
	defer done()
	// 因为 removeUsePmsByHTTPAcc 失败不会触发 tx.Rollback (错误在tx外部)，所以 defer TxRollback 会调用
	sqlMock.ExpectRollback()

	mockAgentConfigRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockReleaseRepo := idbaccessmock.NewMockIReleaseRepo(ctrl)
	mockPermissionSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)
	mockCategoryRel := idbaccessmock.NewMockIReleaseCategoryRelRepo(ctrl)
	mockPermRepo := idbaccessmock.NewMockIReleasePermissionRepo(ctrl)
	mockAuthz := authzaccmock.NewMockAuthZHttpAcc(ctrl)

	agentID := "agent-123"
	currentUserID := "user-123"
	agentPo := &dapo.DataAgentPo{ID: agentID, Name: "Test Agent", CreatedBy: currentUserID}
	releasePo := &dapo.ReleasePO{ID: "release-123", AgentID: agentID}

	mockAgentConfigRepo.EXPECT().GetByID(gomock.Any(), agentID).Return(agentPo, nil)
	mockPermissionSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
	mockReleaseRepo.EXPECT().GetByAgentID(gomock.Any(), agentID).Return(releasePo, nil)
	mockReleaseRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
	mockReleaseRepo.EXPECT().DeleteByAgentID(gomock.Any(), tx, agentID).Return(nil)
	mockCategoryRel.EXPECT().DelByReleaseID(gomock.Any(), tx, "release-123").Return(nil)
	mockPermRepo.EXPECT().DelByReleaseID(gomock.Any(), tx, "release-123").Return(nil)
	mockAgentConfigRepo.EXPECT().UpdateStatus(gomock.Any(), tx, cdaenum.StatusUnpublished, agentID, "").Return(nil)
	mockAuthz.EXPECT().DeleteAgentPolicy(gomock.Any(), agentID).Return(errors.New("authz delete failed"))

	svc := &releaseSvc{
		SvcBase:                service.NewSvcBase(),
		agentConfigRepo:        mockAgentConfigRepo,
		releaseRepo:            mockReleaseRepo,
		pmsSvc:                 mockPermissionSvc,
		releaseCategoryRelRepo: mockCategoryRel,
		releasePermissionRepo:  mockPermRepo,
		authZHttp:              mockAuthz,
	}

	ctx := createUnpublishCtx(currentUserID)
	_, err := svc.UnPublish(ctx, agentID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "remove use pms failed")
}

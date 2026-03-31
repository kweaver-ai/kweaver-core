package tplsvc

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp/cmpmock"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driver/iv3portdriver/v3portdrivermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================== Unpublish — deeper paths ====================

func TestUnpublish_PubedTplNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
	mockPubTplRepo := idbaccessmock.NewMockIPublishedTplRepo(ctrl)
	mockPmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentTplSvc{
		SvcBase:          service.NewSvcBase(),
		agentTplRepo:     mockAgentTplRepo,
		publishedTplRepo: mockPubTplRepo,
		pmsSvc:           mockPmsSvc,
		logger:           mockLogger,
	}

	tplPo := &dapo.DataAgentTplPo{ID: 1, Name: "Tpl", CreatedBy: "u1"}

	mockPmsSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
	mockAgentTplRepo.EXPECT().GetByID(gomock.Any(), int64(1)).Return(tplPo, nil)
	mockPubTplRepo.EXPECT().GetByTplID(gomock.Any(), int64(1)).Return(nil, sql.ErrNoRows)

	_, err := svc.Unpublish(context.Background(), 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "此已发布模板不存在")
}

func TestUnpublish_PubedTplGetError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
	mockPubTplRepo := idbaccessmock.NewMockIPublishedTplRepo(ctrl)
	mockPmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentTplSvc{
		SvcBase:          service.NewSvcBase(),
		agentTplRepo:     mockAgentTplRepo,
		publishedTplRepo: mockPubTplRepo,
		pmsSvc:           mockPmsSvc,
		logger:           mockLogger,
	}

	tplPo := &dapo.DataAgentTplPo{ID: 1, Name: "Tpl", CreatedBy: "u1"}

	mockPmsSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
	mockAgentTplRepo.EXPECT().GetByID(gomock.Any(), int64(1)).Return(tplPo, nil)
	mockPubTplRepo.EXPECT().GetByTplID(gomock.Any(), int64(1)).Return(nil, errors.New("db err"))

	_, err := svc.Unpublish(context.Background(), 1)
	assert.Error(t, err)
}

func TestUnpublish_NotOwner_NoPms(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
	mockPubTplRepo := idbaccessmock.NewMockIPublishedTplRepo(ctrl)
	mockPmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentTplSvc{
		SvcBase:          service.NewSvcBase(),
		agentTplRepo:     mockAgentTplRepo,
		publishedTplRepo: mockPubTplRepo,
		pmsSvc:           mockPmsSvc,
		logger:           mockLogger,
	}

	tplPo := &dapo.DataAgentTplPo{ID: 1, Name: "Tpl", CreatedBy: "other-user"}
	pubedPo := &dapo.PublishedTplPo{ID: 100, TplID: 1}
	// unpublish pms OK
	mockPmsSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
	mockAgentTplRepo.EXPECT().GetByID(gomock.Any(), int64(1)).Return(tplPo, nil)
	mockPubTplRepo.EXPECT().GetByTplID(gomock.Any(), int64(1)).Return(pubedPo, nil)
	// not owner → check 'unpublish other user tpl pms' → denied
	mockPmsSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil)

	_, err := svc.Unpublish(context.Background(), 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "无取消发布的权限")
}

func TestUnpublish_NotOwner_PmsCheckError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
	mockPubTplRepo := idbaccessmock.NewMockIPublishedTplRepo(ctrl)
	mockPmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentTplSvc{
		SvcBase:          service.NewSvcBase(),
		agentTplRepo:     mockAgentTplRepo,
		publishedTplRepo: mockPubTplRepo,
		pmsSvc:           mockPmsSvc,
		logger:           mockLogger,
	}

	tplPo := &dapo.DataAgentTplPo{ID: 1, Name: "Tpl", CreatedBy: "other-user"}
	pubedPo := &dapo.PublishedTplPo{ID: 100, TplID: 1}

	mockPmsSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
	mockAgentTplRepo.EXPECT().GetByID(gomock.Any(), int64(1)).Return(tplPo, nil)
	mockPubTplRepo.EXPECT().GetByTplID(gomock.Any(), int64(1)).Return(pubedPo, nil)
	mockPmsSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, errors.New("pms err"))

	_, err := svc.Unpublish(context.Background(), 1)
	assert.Error(t, err)
}

func TestUnpublish_DelCategoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db, sqlMk, err := sqlmock.New()
	require.NoError(t, err)

	defer db.Close()
	sqlMk.ExpectBegin()
	sqlMk.ExpectRollback()

	tx, err := db.Begin()
	require.NoError(t, err)

	mockAgentTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
	mockPubTplRepo := idbaccessmock.NewMockIPublishedTplRepo(ctrl)
	mockPmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentTplSvc{
		SvcBase:          &service.SvcBase{Logger: mockLogger},
		agentTplRepo:     mockAgentTplRepo,
		publishedTplRepo: mockPubTplRepo,
		pmsSvc:           mockPmsSvc,
	}

	// owner match
	tplPo := &dapo.DataAgentTplPo{ID: 1, Name: "Tpl", CreatedBy: ""}
	pubedPo := &dapo.PublishedTplPo{ID: 100, TplID: 1}

	mockPmsSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
	mockAgentTplRepo.EXPECT().GetByID(gomock.Any(), int64(1)).Return(tplPo, nil)
	mockPubTplRepo.EXPECT().GetByTplID(gomock.Any(), int64(1)).Return(pubedPo, nil)
	mockAgentTplRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
	mockPubTplRepo.EXPECT().DelCategoryAssocByTplID(gomock.Any(), tx, int64(1)).Return(errors.New("del cat err"))
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

	_, err = svc.Unpublish(context.Background(), 1)
	assert.Error(t, err)
	require.NoError(t, sqlMk.ExpectationsWereMet())
}

func TestUnpublish_UpdateStatusError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db, sqlMk, err := sqlmock.New()
	require.NoError(t, err)

	defer db.Close()
	sqlMk.ExpectBegin()
	sqlMk.ExpectRollback()

	tx, err := db.Begin()
	require.NoError(t, err)

	mockAgentTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
	mockPubTplRepo := idbaccessmock.NewMockIPublishedTplRepo(ctrl)
	mockPmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentTplSvc{
		SvcBase:          &service.SvcBase{Logger: mockLogger},
		agentTplRepo:     mockAgentTplRepo,
		publishedTplRepo: mockPubTplRepo,
		pmsSvc:           mockPmsSvc,
	}

	tplPo := &dapo.DataAgentTplPo{ID: 1, Name: "Tpl", CreatedBy: ""}
	pubedPo := &dapo.PublishedTplPo{ID: 100, TplID: 1}

	mockPmsSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
	mockAgentTplRepo.EXPECT().GetByID(gomock.Any(), int64(1)).Return(tplPo, nil)
	mockPubTplRepo.EXPECT().GetByTplID(gomock.Any(), int64(1)).Return(pubedPo, nil)
	mockAgentTplRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
	mockPubTplRepo.EXPECT().DelCategoryAssocByTplID(gomock.Any(), tx, int64(1)).Return(nil)
	mockAgentTplRepo.EXPECT().UpdateStatus(gomock.Any(), tx, gomock.Any(), int64(1), gomock.Any(), gomock.Any()).Return(errors.New("status err"))
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

	_, err = svc.Unpublish(context.Background(), 1)
	assert.Error(t, err)
	require.NoError(t, sqlMk.ExpectationsWereMet())
}

func TestUnpublish_DeletePubedError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db, sqlMk, err := sqlmock.New()
	require.NoError(t, err)

	defer db.Close()
	sqlMk.ExpectBegin()
	sqlMk.ExpectRollback()

	tx, err := db.Begin()
	require.NoError(t, err)

	mockAgentTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
	mockPubTplRepo := idbaccessmock.NewMockIPublishedTplRepo(ctrl)
	mockPmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentTplSvc{
		SvcBase:          &service.SvcBase{Logger: mockLogger},
		agentTplRepo:     mockAgentTplRepo,
		publishedTplRepo: mockPubTplRepo,
		pmsSvc:           mockPmsSvc,
	}

	tplPo := &dapo.DataAgentTplPo{ID: 1, Name: "Tpl", CreatedBy: ""}
	pubedPo := &dapo.PublishedTplPo{ID: 100, TplID: 1}

	mockPmsSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
	mockAgentTplRepo.EXPECT().GetByID(gomock.Any(), int64(1)).Return(tplPo, nil)
	mockPubTplRepo.EXPECT().GetByTplID(gomock.Any(), int64(1)).Return(pubedPo, nil)
	mockAgentTplRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
	mockPubTplRepo.EXPECT().DelCategoryAssocByTplID(gomock.Any(), tx, int64(1)).Return(nil)
	mockAgentTplRepo.EXPECT().UpdateStatus(gomock.Any(), tx, gomock.Any(), int64(1), gomock.Any(), gomock.Any()).Return(nil)
	mockPubTplRepo.EXPECT().Delete(gomock.Any(), tx, int64(100)).Return(errors.New("del err"))
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

	_, err = svc.Unpublish(context.Background(), 1)
	assert.Error(t, err)
	require.NoError(t, sqlMk.ExpectationsWereMet())
}

func TestUnpublish_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db, sqlMk, err := sqlmock.New()
	require.NoError(t, err)

	defer db.Close()
	sqlMk.ExpectBegin()
	sqlMk.ExpectCommit()

	tx, err := db.Begin()
	require.NoError(t, err)

	mockAgentTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
	mockPubTplRepo := idbaccessmock.NewMockIPublishedTplRepo(ctrl)
	mockPmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentTplSvc{
		SvcBase:          &service.SvcBase{Logger: mockLogger},
		agentTplRepo:     mockAgentTplRepo,
		publishedTplRepo: mockPubTplRepo,
		pmsSvc:           mockPmsSvc,
	}

	tplPo := &dapo.DataAgentTplPo{ID: 1, Name: "Tpl", CreatedBy: ""}
	pubedPo := &dapo.PublishedTplPo{ID: 100, TplID: 1}

	mockPmsSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
	mockAgentTplRepo.EXPECT().GetByID(gomock.Any(), int64(1)).Return(tplPo, nil)
	mockPubTplRepo.EXPECT().GetByTplID(gomock.Any(), int64(1)).Return(pubedPo, nil)
	mockAgentTplRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
	mockPubTplRepo.EXPECT().DelCategoryAssocByTplID(gomock.Any(), tx, int64(1)).Return(nil)
	mockAgentTplRepo.EXPECT().UpdateStatus(gomock.Any(), tx, gomock.Any(), int64(1), gomock.Any(), gomock.Any()).Return(nil)
	mockPubTplRepo.EXPECT().Delete(gomock.Any(), tx, int64(100)).Return(nil)

	auditLog, err := svc.Unpublish(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, "1", auditLog.ID)
	assert.Equal(t, "Tpl", auditLog.Name)
	require.NoError(t, sqlMk.ExpectationsWereMet())
}

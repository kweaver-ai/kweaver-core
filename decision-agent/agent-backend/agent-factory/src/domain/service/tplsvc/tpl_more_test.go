package tplsvc

import (
	"context"
	"errors"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/agent_tpl/agenttplreq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/ihttpaccess/ibizdomainacc/bizdomainaccmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driver/iv3portdriver/v3portdrivermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- UpdatePublishInfo additional tests ---

func TestDataAgentTplSvc_UpdatePublishInfo_GetByTplIDError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
	mockPubedRepo := idbaccessmock.NewMockIPublishedTplRepo(ctrl)
	mockPms := v3portdrivermock.NewMockIPermissionSvc(ctrl)

	svc := &dataAgentTplSvc{
		SvcBase:          &service.SvcBase{Logger: noopTplLogger{}},
		agentTplRepo:     mockTplRepo,
		publishedTplRepo: mockPubedRepo,
		pmsSvc:           mockPms,
	}

	po := &dapo.DataAgentTplPo{ID: 1, Name: "tpl1", CreatedBy: "u1"}

	mockPms.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
	mockTplRepo.EXPECT().GetByID(gomock.Any(), int64(1)).Return(po, nil)
	mockPubedRepo.EXPECT().GetByTplID(gomock.Any(), int64(1)).Return(nil, errors.New("db error"))

	_, _, err := svc.UpdatePublishInfo(context.Background(), &agenttplreq.UpdatePublishInfoReq{}, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "get published template by tpl id")
}

func TestDataAgentTplSvc_UpdatePublishInfo_NotOwner(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
	mockPubedRepo := idbaccessmock.NewMockIPublishedTplRepo(ctrl)
	mockPms := v3portdrivermock.NewMockIPermissionSvc(ctrl)

	svc := &dataAgentTplSvc{
		SvcBase:          &service.SvcBase{Logger: noopTplLogger{}},
		agentTplRepo:     mockTplRepo,
		publishedTplRepo: mockPubedRepo,
		pmsSvc:           mockPms,
	}

	// Creator is u2, current user is u1 → should fail with 403
	ctx := createTplCtxWithUserID("u1")
	po := &dapo.DataAgentTplPo{ID: 5, Name: "tpl5", CreatedBy: "u2"}

	mockPms.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
	mockTplRepo.EXPECT().GetByID(gomock.Any(), int64(5)).Return(po, nil)
	mockPubedRepo.EXPECT().GetByTplID(gomock.Any(), int64(5)).Return(&dapo.PublishedTplPo{ID: 50, TplID: 5}, nil)

	_, _, err := svc.UpdatePublishInfo(ctx, &agenttplreq.UpdatePublishInfoReq{}, 5)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "非创建人")
}

func TestDataAgentTplSvc_UpdatePublishInfo_BeginTxError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
	mockPubedRepo := idbaccessmock.NewMockIPublishedTplRepo(ctrl)
	mockPms := v3portdrivermock.NewMockIPermissionSvc(ctrl)

	svc := &dataAgentTplSvc{
		SvcBase:          &service.SvcBase{Logger: noopTplLogger{}},
		agentTplRepo:     mockTplRepo,
		publishedTplRepo: mockPubedRepo,
		pmsSvc:           mockPms,
	}

	ctx := createTplCtxWithUserID("u1")
	po := &dapo.DataAgentTplPo{ID: 6, Name: "tpl6", CreatedBy: "u1"}

	mockPms.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
	mockTplRepo.EXPECT().GetByID(gomock.Any(), int64(6)).Return(po, nil)
	mockPubedRepo.EXPECT().GetByTplID(gomock.Any(), int64(6)).Return(&dapo.PublishedTplPo{ID: 60, TplID: 6}, nil)
	mockTplRepo.EXPECT().BeginTx(gomock.Any()).Return(nil, errors.New("begin tx failed"))

	_, _, err := svc.UpdatePublishInfo(ctx, &agenttplreq.UpdatePublishInfoReq{}, 6)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "begin transaction")
}

// --- Delete additional tests ---

func TestDataAgentTplSvc_Delete_BuiltIn_CannotDelete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)

	svc := &dataAgentTplSvc{
		SvcBase:      &service.SvcBase{Logger: noopTplLogger{}},
		agentTplRepo: mockTplRepo,
	}

	builtInYes := cdaenum.BuiltInYes
	po := &dapo.DataAgentTplPo{
		ID:        10,
		Name:      "BuiltIn Tpl",
		Status:    cdaenum.StatusUnpublished,
		CreatedBy: "u1",
		IsBuiltIn: &builtInYes,
	}

	mockTplRepo.EXPECT().ExistsByID(gomock.Any(), int64(10)).Return(true, nil)
	mockTplRepo.EXPECT().GetByID(gomock.Any(), int64(10)).Return(po, nil)

	_, err := svc.Delete(context.Background(), 10, "u1", false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "内置数据智能体模板不可删除")
}

func TestDataAgentTplSvc_Delete_BeginTxError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)

	builtInNo := cdaenum.BuiltInNo
	svc := &dataAgentTplSvc{
		SvcBase:      &service.SvcBase{Logger: noopTplLogger{}},
		agentTplRepo: mockTplRepo,
	}

	po := &dapo.DataAgentTplPo{
		ID:        11,
		Name:      "Tpl11",
		Status:    cdaenum.StatusUnpublished,
		CreatedBy: "u1",
		IsBuiltIn: &builtInNo,
	}

	mockTplRepo.EXPECT().ExistsByID(gomock.Any(), int64(11)).Return(true, nil)
	mockTplRepo.EXPECT().GetByID(gomock.Any(), int64(11)).Return(po, nil)
	mockTplRepo.EXPECT().BeginTx(gomock.Any()).Return(nil, errors.New("tx failed"))

	_, err := svc.Delete(context.Background(), 11, "u1", false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "begin transaction")
}

func TestDataAgentTplSvc_Delete_RepoDeleteError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db, sqlMk, err := sqlmock.New()
	require.NoError(t, err)

	defer db.Close()
	sqlMk.ExpectBegin()
	sqlMk.ExpectRollback()

	tx, err := db.Begin()
	require.NoError(t, err)

	mockTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
	mockPubedRepo := idbaccessmock.NewMockIPublishedTplRepo(ctrl)

	builtInNo := cdaenum.BuiltInNo
	svc := &dataAgentTplSvc{
		SvcBase:          &service.SvcBase{Logger: noopTplLogger{}},
		agentTplRepo:     mockTplRepo,
		publishedTplRepo: mockPubedRepo,
	}

	po := &dapo.DataAgentTplPo{ID: 12, Name: "Tpl12", Status: cdaenum.StatusUnpublished, CreatedBy: "u1", IsBuiltIn: &builtInNo}

	mockTplRepo.EXPECT().ExistsByID(gomock.Any(), int64(12)).Return(true, nil)
	mockTplRepo.EXPECT().GetByID(gomock.Any(), int64(12)).Return(po, nil)
	mockTplRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
	mockTplRepo.EXPECT().Delete(gomock.Any(), tx, int64(12)).Return(errors.New("delete failed"))

	_, err = svc.Delete(context.Background(), 12, "u1", false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "delete template")
	require.NoError(t, sqlMk.ExpectationsWereMet())
}

func TestDataAgentTplSvc_Delete_DeletePublishedError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db, sqlMk, err := sqlmock.New()
	require.NoError(t, err)

	defer db.Close()
	sqlMk.ExpectBegin()
	sqlMk.ExpectRollback()

	tx, err := db.Begin()
	require.NoError(t, err)

	mockTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
	mockPubedRepo := idbaccessmock.NewMockIPublishedTplRepo(ctrl)

	builtInNo := cdaenum.BuiltInNo
	svc := &dataAgentTplSvc{
		SvcBase:          &service.SvcBase{Logger: noopTplLogger{}},
		agentTplRepo:     mockTplRepo,
		publishedTplRepo: mockPubedRepo,
	}

	po := &dapo.DataAgentTplPo{ID: 13, Name: "Tpl13", Status: cdaenum.StatusUnpublished, CreatedBy: "u1", IsBuiltIn: &builtInNo}

	mockTplRepo.EXPECT().ExistsByID(gomock.Any(), int64(13)).Return(true, nil)
	mockTplRepo.EXPECT().GetByID(gomock.Any(), int64(13)).Return(po, nil)
	mockTplRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
	mockTplRepo.EXPECT().Delete(gomock.Any(), tx, int64(13)).Return(nil)
	mockPubedRepo.EXPECT().DeleteByTplID(gomock.Any(), tx, int64(13)).Return(errors.New("delete published failed"))

	_, err = svc.Delete(context.Background(), 13, "u1", false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "delete published template")
	require.NoError(t, sqlMk.ExpectationsWereMet())
}

func TestDataAgentTplSvc_Delete_BdRelDeleteError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db, sqlMk, err := sqlmock.New()
	require.NoError(t, err)

	defer db.Close()
	sqlMk.ExpectBegin()
	sqlMk.ExpectRollback()

	tx, err := db.Begin()
	require.NoError(t, err)

	mockTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
	mockPubedRepo := idbaccessmock.NewMockIPublishedTplRepo(ctrl)
	mockBdRelRepo := idbaccessmock.NewMockIBizDomainAgentTplRelRepo(ctrl)

	builtInNo := cdaenum.BuiltInNo
	svc := &dataAgentTplSvc{
		SvcBase:           &service.SvcBase{Logger: noopTplLogger{}},
		agentTplRepo:      mockTplRepo,
		publishedTplRepo:  mockPubedRepo,
		bdAgentTplRelRepo: mockBdRelRepo,
	}

	ctx := context.WithValue(context.Background(), cenum.BizDomainIDCtxKey.String(), "bd-99") //nolint:staticcheck
	po := &dapo.DataAgentTplPo{ID: 14, Name: "Tpl14", Status: cdaenum.StatusUnpublished, CreatedBy: "u1", IsBuiltIn: &builtInNo}

	mockTplRepo.EXPECT().ExistsByID(gomock.Any(), int64(14)).Return(true, nil)
	mockTplRepo.EXPECT().GetByID(gomock.Any(), int64(14)).Return(po, nil)
	mockTplRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
	mockTplRepo.EXPECT().Delete(gomock.Any(), tx, int64(14)).Return(nil)
	mockPubedRepo.EXPECT().DeleteByTplID(gomock.Any(), tx, int64(14)).Return(nil)
	mockBdRelRepo.EXPECT().DeleteByAgentTplID(gomock.Any(), tx, int64(14)).Return(errors.New("bd rel delete failed"))

	_, err = svc.Delete(ctx, 14, "u1", false)
	assert.Error(t, err)
	require.NoError(t, sqlMk.ExpectationsWereMet())
}

func TestDataAgentTplSvc_Delete_DisassociateError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db, sqlMk, err := sqlmock.New()
	require.NoError(t, err)

	defer db.Close()
	sqlMk.ExpectBegin()
	sqlMk.ExpectRollback()

	tx, err := db.Begin()
	require.NoError(t, err)

	mockTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
	mockPubedRepo := idbaccessmock.NewMockIPublishedTplRepo(ctrl)
	mockBdRelRepo := idbaccessmock.NewMockIBizDomainAgentTplRelRepo(ctrl)
	mockBizDomain := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)

	builtInNo := cdaenum.BuiltInNo
	svc := &dataAgentTplSvc{
		SvcBase:           &service.SvcBase{Logger: noopTplLogger{}},
		agentTplRepo:      mockTplRepo,
		publishedTplRepo:  mockPubedRepo,
		bdAgentTplRelRepo: mockBdRelRepo,
		bizDomainHttp:     mockBizDomain,
	}

	ctx := context.WithValue(context.Background(), cenum.BizDomainIDCtxKey.String(), "bd-99") //nolint:staticcheck // SA1029
	po := &dapo.DataAgentTplPo{ID: 15, Name: "Tpl15", Status: cdaenum.StatusUnpublished, CreatedBy: "u1", IsBuiltIn: &builtInNo}

	mockTplRepo.EXPECT().ExistsByID(gomock.Any(), int64(15)).Return(true, nil)
	mockTplRepo.EXPECT().GetByID(gomock.Any(), int64(15)).Return(po, nil)
	mockTplRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
	mockTplRepo.EXPECT().Delete(gomock.Any(), tx, int64(15)).Return(nil)
	mockPubedRepo.EXPECT().DeleteByTplID(gomock.Any(), tx, int64(15)).Return(nil)
	mockBdRelRepo.EXPECT().DeleteByAgentTplID(gomock.Any(), tx, int64(15)).Return(nil)
	mockBizDomain.EXPECT().DisassociateResource(gomock.Any(), gomock.Any()).Return(errors.New("disassociate failed"))

	_, err = svc.Delete(ctx, 15, "u1", false)
	assert.Error(t, err)
	require.NoError(t, sqlMk.ExpectationsWereMet())
}

// Delete with isPrivate=true (skip owner check) should still go to BeginTx
func TestDataAgentTplSvc_Delete_IsPrivate_SkipsOwnerCheck(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)

	builtInNo := cdaenum.BuiltInNo
	svc := &dataAgentTplSvc{
		SvcBase:      &service.SvcBase{Logger: noopTplLogger{}},
		agentTplRepo: mockTplRepo,
	}

	// CreatedBy differs from uid, but isPrivate=true
	po := &dapo.DataAgentTplPo{ID: 16, Name: "PrivateTpl", Status: cdaenum.StatusUnpublished, CreatedBy: "other-user", IsBuiltIn: &builtInNo}

	mockTplRepo.EXPECT().ExistsByID(gomock.Any(), int64(16)).Return(true, nil)
	mockTplRepo.EXPECT().GetByID(gomock.Any(), int64(16)).Return(po, nil)
	mockTplRepo.EXPECT().BeginTx(gomock.Any()).Return(nil, errors.New("tx err"))

	_, err := svc.Delete(context.Background(), 16, "u1", true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "begin transaction")
}

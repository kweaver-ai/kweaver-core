package tplsvc

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/agent_tpl/agenttplreq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/cmp/icmp/cmpmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driver/iv3portdriver/v3portdrivermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================== Publish — deeper paths ====================

func TestPublish_TplNotFound(t *testing.T) {
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
	mockPmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentTplSvc{
		SvcBase:      &service.SvcBase{Logger: mockLogger},
		agentTplRepo: mockAgentTplRepo,
		pmsSvc:       mockPmsSvc,
	}

	mockPmsSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
	mockAgentTplRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
	mockAgentTplRepo.EXPECT().GetByIDWithTx(gomock.Any(), tx, int64(1)).Return(nil, sql.ErrNoRows)

	_, _, err = svc.Publish(context.Background(), nil, &agenttplreq.PublishReq{}, 1, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "模板不存在")
	require.NoError(t, sqlMk.ExpectationsWereMet())
}

func TestPublish_GetByIDWithTxError(t *testing.T) {
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
	mockPmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentTplSvc{
		SvcBase:      &service.SvcBase{Logger: mockLogger},
		agentTplRepo: mockAgentTplRepo,
		pmsSvc:       mockPmsSvc,
	}

	mockPmsSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
	mockAgentTplRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
	mockAgentTplRepo.EXPECT().GetByIDWithTx(gomock.Any(), tx, int64(1)).Return(nil, errors.New("db err"))

	_, _, err = svc.Publish(context.Background(), nil, &agenttplreq.PublishReq{}, 1, false)
	assert.Error(t, err)
	require.NoError(t, sqlMk.ExpectationsWereMet())
}

func TestPublish_NotOwner(t *testing.T) {
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
	mockPmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentTplSvc{
		SvcBase:      &service.SvcBase{Logger: mockLogger},
		agentTplRepo: mockAgentTplRepo,
		pmsSvc:       mockPmsSvc,
	}

	tplPo := &dapo.DataAgentTplPo{ID: 1, Name: "Tpl", CreatedBy: "other-user"}

	mockPmsSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
	mockAgentTplRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
	mockAgentTplRepo.EXPECT().GetByIDWithTx(gomock.Any(), tx, int64(1)).Return(tplPo, nil)

	_, _, err = svc.Publish(context.Background(), nil, &agenttplreq.PublishReq{}, 1, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "无权限发布，非创建人")
	require.NoError(t, sqlMk.ExpectationsWereMet())
}

func TestPublish_GenPublishedPoDeleteError(t *testing.T) {
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

	// owner match (CreatedBy = "" matches GetUserIDFromCtx(Background()) = "")
	tplPo := &dapo.DataAgentTplPo{ID: 1, Name: "Tpl", CreatedBy: "", Config: "{}"}

	mockPmsSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
	mockAgentTplRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
	mockAgentTplRepo.EXPECT().GetByIDWithTx(gomock.Any(), tx, int64(1)).Return(tplPo, nil)
	mockPubTplRepo.EXPECT().DeleteByTplID(gomock.Any(), tx, int64(1)).Return(errors.New("del err"))

	_, _, err = svc.Publish(context.Background(), nil, &agenttplreq.PublishReq{}, 1, false)
	assert.Error(t, err)
	require.NoError(t, sqlMk.ExpectationsWereMet())
}

func TestPublish_GenPublishedPoCreateError(t *testing.T) {
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

	tplPo := &dapo.DataAgentTplPo{ID: 1, Name: "Tpl", CreatedBy: "", Config: "{}"}

	mockPmsSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
	mockAgentTplRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
	mockAgentTplRepo.EXPECT().GetByIDWithTx(gomock.Any(), tx, int64(1)).Return(tplPo, nil)
	mockPubTplRepo.EXPECT().DeleteByTplID(gomock.Any(), tx, int64(1)).Return(nil)
	mockPubTplRepo.EXPECT().Create(gomock.Any(), tx, gomock.Any()).Return(int64(0), errors.New("create err"))

	_, _, err = svc.Publish(context.Background(), nil, &agenttplreq.PublishReq{}, 1, false)
	assert.Error(t, err)
	require.NoError(t, sqlMk.ExpectationsWereMet())
}

func TestPublish_HandleCategoryError(t *testing.T) {
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
	mockCategoryRepo := idbaccessmock.NewMockICategoryRepo(ctrl)
	mockPmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentTplSvc{
		SvcBase:          &service.SvcBase{Logger: mockLogger},
		agentTplRepo:     mockAgentTplRepo,
		publishedTplRepo: mockPubTplRepo,
		categoryRepo:     mockCategoryRepo,
		pmsSvc:           mockPmsSvc,
	}

	tplPo := &dapo.DataAgentTplPo{ID: 1, Name: "Tpl", CreatedBy: "", Config: "{}"}

	mockPmsSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
	mockAgentTplRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
	mockAgentTplRepo.EXPECT().GetByIDWithTx(gomock.Any(), tx, int64(1)).Return(tplPo, nil)
	mockPubTplRepo.EXPECT().DeleteByTplID(gomock.Any(), tx, int64(1)).Return(nil)
	mockPubTplRepo.EXPECT().Create(gomock.Any(), tx, gomock.Any()).Return(int64(200), nil)
	// handleCategory → categoryRepo.GetIDNameMap fails
	mockCategoryRepo.EXPECT().GetIDNameMap(gomock.Any(), gomock.Any()).Return(nil, errors.New("cat err"))

	req := &agenttplreq.PublishReq{UpdatePublishInfoReq: &agenttplreq.UpdatePublishInfoReq{CategoryIDs: []string{"cat-1"}}}
	_, _, err = svc.Publish(context.Background(), nil, req, 1, false)
	assert.Error(t, err)
	require.NoError(t, sqlMk.ExpectationsWereMet())
}

func TestPublish_HandleCategoryNotFound(t *testing.T) {
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
	mockCategoryRepo := idbaccessmock.NewMockICategoryRepo(ctrl)
	mockPmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentTplSvc{
		SvcBase:          &service.SvcBase{Logger: mockLogger},
		agentTplRepo:     mockAgentTplRepo,
		publishedTplRepo: mockPubTplRepo,
		categoryRepo:     mockCategoryRepo,
		pmsSvc:           mockPmsSvc,
	}

	tplPo := &dapo.DataAgentTplPo{ID: 1, Name: "Tpl", CreatedBy: "", Config: "{}"}

	mockPmsSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
	mockAgentTplRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
	mockAgentTplRepo.EXPECT().GetByIDWithTx(gomock.Any(), tx, int64(1)).Return(tplPo, nil)
	mockPubTplRepo.EXPECT().DeleteByTplID(gomock.Any(), tx, int64(1)).Return(nil)
	mockPubTplRepo.EXPECT().Create(gomock.Any(), tx, gomock.Any()).Return(int64(200), nil)
	// handleCategory → category exists but name is empty (not found)
	mockCategoryRepo.EXPECT().GetIDNameMap(gomock.Any(), gomock.Any()).Return(map[string]string{}, nil)

	req := &agenttplreq.PublishReq{UpdatePublishInfoReq: &agenttplreq.UpdatePublishInfoReq{CategoryIDs: []string{"cat-1"}}}
	_, _, err = svc.Publish(context.Background(), nil, req, 1, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "分类[cat-1]不存在")
	require.NoError(t, sqlMk.ExpectationsWereMet())
}

func TestPublish_UpdateStatusError(t *testing.T) {
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

	tplPo := &dapo.DataAgentTplPo{ID: 1, Name: "Tpl", CreatedBy: "", Config: "{}"}

	mockPmsSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
	mockAgentTplRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
	mockAgentTplRepo.EXPECT().GetByIDWithTx(gomock.Any(), tx, int64(1)).Return(tplPo, nil)
	mockPubTplRepo.EXPECT().DeleteByTplID(gomock.Any(), tx, int64(1)).Return(nil)
	mockPubTplRepo.EXPECT().Create(gomock.Any(), tx, gomock.Any()).Return(int64(200), nil)
	// no categories → skip handleCategory
	mockAgentTplRepo.EXPECT().UpdateStatus(gomock.Any(), tx, gomock.Any(), int64(1), gomock.Any(), gomock.Any()).Return(errors.New("status err"))

	req := agenttplreq.NewPublishReq()
	_, _, err = svc.Publish(context.Background(), nil, req, 1, false)
	assert.Error(t, err)
	require.NoError(t, sqlMk.ExpectationsWereMet())
}

func TestPublish_FromCopy2TplAndPublish_SkipsPmsCheck(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db, sqlMk, err := sqlmock.New()
	require.NoError(t, err)

	defer db.Close()
	sqlMk.ExpectBegin()

	tx, err := db.Begin()
	require.NoError(t, err)

	mockAgentTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
	mockPubTplRepo := idbaccessmock.NewMockIPublishedTplRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentTplSvc{
		SvcBase:          &service.SvcBase{Logger: mockLogger},
		agentTplRepo:     mockAgentTplRepo,
		publishedTplRepo: mockPubTplRepo,
		// No pmsSvc → would panic if pms check is called
	}

	tplPo := &dapo.DataAgentTplPo{ID: 1, Name: "Tpl", CreatedBy: "", Config: "{}"}

	// isFromCopy2TplAndPublish=true → skip pms + skip BeginTx (pass tx in)
	mockAgentTplRepo.EXPECT().GetByIDWithTx(gomock.Any(), tx, int64(1)).Return(tplPo, nil)
	mockPubTplRepo.EXPECT().DeleteByTplID(gomock.Any(), tx, int64(1)).Return(errors.New("del err"))

	_, _, err = svc.Publish(context.Background(), tx, &agenttplreq.PublishReq{}, 1, true)
	assert.Error(t, err)
	require.NoError(t, sqlMk.ExpectationsWereMet())
}

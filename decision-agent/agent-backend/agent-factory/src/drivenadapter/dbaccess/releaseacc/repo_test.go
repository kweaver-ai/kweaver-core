package releaseacc

import (
	"context"
	"errors"
	"regexp"
	"sync"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/dbaccess"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/global"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/proton-rds-sdk-go/sqlx"
)

type releaseTestLogger struct{}

func (releaseTestLogger) Infof(string, ...interface{})  {}
func (releaseTestLogger) Infoln(...interface{})         {}
func (releaseTestLogger) Debugf(string, ...interface{}) {}
func (releaseTestLogger) Debugln(...interface{})        {}
func (releaseTestLogger) Errorf(string, ...interface{}) {}
func (releaseTestLogger) Errorln(...interface{})        {}
func (releaseTestLogger) Warnf(string, ...interface{})  {}
func (releaseTestLogger) Warnln(...interface{})         {}
func (releaseTestLogger) Panicf(string, ...interface{}) {}
func (releaseTestLogger) Panicln(...interface{})        {}
func (releaseTestLogger) Fatalf(string, ...interface{}) {}
func (releaseTestLogger) Fatalln(...interface{})        {}

func newReleaseRepoWithMock(t *testing.T) (*releaseRepo, *sqlx.DB, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlx.New()
	require.NoError(t, err)

	repo := &releaseRepo{
		db:             db,
		logger:         releaseTestLogger{},
		IDBAccBaseRepo: dbaccess.NewDBAccBase(),
	}

	return repo, db, mock
}

func newReleaseHistoryRepoWithMock(t *testing.T) (*releaseHistoryRepo, *sqlx.DB, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlx.New()
	require.NoError(t, err)

	repo := &releaseHistoryRepo{
		db:             db,
		logger:         releaseTestLogger{},
		IDBAccBaseRepo: dbaccess.NewDBAccBase(),
	}

	return repo, db, mock
}

func newReleasePermissionRepoWithMock(t *testing.T) (*releasePermissionRepo, *sqlx.DB, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlx.New()
	require.NoError(t, err)

	repo := &releasePermissionRepo{
		db:             db,
		logger:         releaseTestLogger{},
		IDBAccBaseRepo: dbaccess.NewDBAccBase(),
	}

	return repo, db, mock
}

func newReleaseCategoryRelRepoWithMock(t *testing.T) (*releaseCategoryRelRepo, *sqlx.DB, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlx.New()
	require.NoError(t, err)

	repo := &releaseCategoryRelRepo{
		db:             db,
		logger:         releaseTestLogger{},
		IDBAccBaseRepo: dbaccess.NewDBAccBase(),
	}

	return repo, db, mock
}

// --- releaseRepo Singleton ---

func TestNewReleaseRepo_Singleton(t *testing.T) {
	t.Parallel()

	oldOnce := releaseRepoOnce //nolint:govet
	oldImpl := releaseRepoImpl
	oldGDB := global.GDB

	t.Cleanup(func() {
		releaseRepoOnce = oldOnce //nolint:govet
		releaseRepoImpl = oldImpl
		global.GDB = oldGDB
	})

	db, _, err := sqlx.New()
	require.NoError(t, err)

	global.GDB = db
	releaseRepoOnce = sync.Once{}
	releaseRepoImpl = nil

	r1 := NewReleaseRepo()
	r2 := NewReleaseRepo()

	assert.NotNil(t, r1)
	assert.Same(t, r1, r2)
}

// --- releaseRepo.Create ---

func TestReleaseRepo_Create_Happy(t *testing.T) {
	t.Parallel()

	repo, db, mock := newReleaseRepoWithMock(t)
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta("insert into t_data_agent_release")).
		WillReturnResult(sqlmock.NewResult(1, 1))

	po := &dapo.ReleasePO{AgentID: "agent-1", AgentName: "test", CreateBy: "u1", UpdateBy: "u1"}
	_, err := repo.Create(context.Background(), nil, po)
	require.NoError(t, err)
	assert.NotEmpty(t, po.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReleaseRepo_Create_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newReleaseRepoWithMock(t)
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta("insert into t_data_agent_release")).
		WillReturnError(errors.New("insert failed"))

	po := &dapo.ReleasePO{AgentID: "agent-1"}
	_, err := repo.Create(context.Background(), nil, po)
	assert.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- releaseRepo.GetByAgentID ---

func mockReleaseRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"f_id", "f_agent_id", "f_agent_name", "f_agent_config", "f_agent_version", "f_agent_desc",
		"f_is_api_agent", "f_is_web_sdk_agent", "f_is_skill_agent", "f_is_data_flow_agent",
		"f_is_to_custom_space", "f_is_to_square", "f_is_pms_ctrl",
		"f_create_time", "f_update_time", "f_create_by", "f_update_by",
	}).AddRow(
		"rel-1", "agent-1", "name-1", `{}`, "v1", "desc",
		1, 0, 0, 0,
		0, 0, 0,
		int64(1), int64(1), "u1", "u1",
	)
}

func TestReleaseRepo_GetByAgentID_Happy(t *testing.T) {
	t.Parallel()

	repo, db, mock := newReleaseRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`select .* from t_data_agent_release where f_agent_id = \?`).
		WithArgs("agent-1").
		WillReturnRows(mockReleaseRows())

	po, err := repo.GetByAgentID(context.Background(), "agent-1")
	require.NoError(t, err)
	assert.Equal(t, "rel-1", po.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReleaseRepo_GetByAgentID_QueryError(t *testing.T) {
	t.Parallel()

	repo, db, mock := newReleaseRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`select .* from t_data_agent_release where f_agent_id = \?`).
		WithArgs("agent-1").
		WillReturnError(errors.New("query failed"))

	_, err := repo.GetByAgentID(context.Background(), "agent-1")
	assert.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- releaseRepo.Update ---

func TestReleaseRepo_Update_Happy(t *testing.T) {
	t.Parallel()

	repo, db, mock := newReleaseRepoWithMock(t)
	defer db.Close()

	mock.ExpectExec(`update t_data_agent_release set .* where f_id = \?`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	po := &dapo.ReleasePO{ID: "rel-1", AgentID: "agent-1", AgentName: "updated"}
	err := repo.Update(context.Background(), nil, po)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReleaseRepo_Update_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newReleaseRepoWithMock(t)
	defer db.Close()

	mock.ExpectExec(`update t_data_agent_release set .* where f_id = \?`).
		WillReturnError(errors.New("update failed"))

	po := &dapo.ReleasePO{ID: "rel-1"}
	err := repo.Update(context.Background(), nil, po)
	assert.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- releaseRepo.DeleteByAgentID ---

func TestReleaseRepo_DeleteByAgentID_Happy(t *testing.T) {
	t.Parallel()

	repo, db, mock := newReleaseRepoWithMock(t)
	defer db.Close()

	mock.ExpectExec(`delete from t_data_agent_release where f_agent_id = \?`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.DeleteByAgentID(context.Background(), nil, "agent-1")
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReleaseRepo_DeleteByAgentID_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newReleaseRepoWithMock(t)
	defer db.Close()

	mock.ExpectExec(`delete from t_data_agent_release where f_agent_id = \?`).
		WillReturnError(errors.New("delete failed"))

	err := repo.DeleteByAgentID(context.Background(), nil, "agent-1")
	assert.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- releaseHistoryRepo Singleton ---

func TestNewReleaseHistoryRepo_Singleton(t *testing.T) {
	t.Parallel()

	oldOnce := releaseHistoryRepoOnce //nolint:govet
	oldImpl := releaseHistoryRepoImpl
	oldGDB := global.GDB

	t.Cleanup(func() {
		releaseHistoryRepoOnce = oldOnce //nolint:govet
		releaseHistoryRepoImpl = oldImpl
		global.GDB = oldGDB
	})

	db, _, err := sqlx.New()
	require.NoError(t, err)

	global.GDB = db
	releaseHistoryRepoOnce = sync.Once{}
	releaseHistoryRepoImpl = nil

	r1 := NewReleaseHistoryRepo()
	r2 := NewReleaseHistoryRepo()

	assert.NotNil(t, r1)
	assert.Same(t, r1, r2)
}

// --- releaseHistoryRepo.Create ---

func TestReleaseHistoryRepo_Create_Happy(t *testing.T) {
	t.Parallel()

	repo, db, mock := newReleaseHistoryRepoWithMock(t)
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta("insert into t_data_agent_release_history")).
		WillReturnResult(sqlmock.NewResult(1, 1))

	po := &dapo.ReleaseHistoryPO{AgentID: "agent-1", AgentVersion: "v1", CreateBy: "u1"}
	_, err := repo.Create(context.Background(), nil, po)
	require.NoError(t, err)
	assert.NotEmpty(t, po.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReleaseHistoryRepo_Create_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newReleaseHistoryRepoWithMock(t)
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta("insert into t_data_agent_release_history")).
		WillReturnError(errors.New("insert failed"))

	po := &dapo.ReleaseHistoryPO{AgentID: "agent-1"}
	_, err := repo.Create(context.Background(), nil, po)
	assert.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- releaseHistoryRepo.GetByAgentIdVersion ---

func mockReleaseHistoryRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"f_id", "f_agent_id", "f_agent_config", "f_agent_version", "f_agent_desc",
		"f_create_time", "f_update_time", "f_create_by", "f_update_by",
	}).AddRow(
		"hist-1", "agent-1", `{}`, "v1", "desc",
		int64(1), int64(1), "u1", "u1",
	)
}

func TestReleaseHistoryRepo_GetByAgentIdVersion_Happy(t *testing.T) {
	t.Parallel()

	repo, db, mock := newReleaseHistoryRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`select .* from t_data_agent_release_history where f_agent_id = \? and f_agent_version = \?`).
		WithArgs("agent-1", "v1").
		WillReturnRows(mockReleaseHistoryRows())

	po, err := repo.GetByAgentIdVersion(context.Background(), "agent-1", "v1")
	require.NoError(t, err)
	assert.Equal(t, "hist-1", po.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReleaseHistoryRepo_GetByAgentIdVersion_QueryError(t *testing.T) {
	t.Parallel()

	repo, db, mock := newReleaseHistoryRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`select .* from t_data_agent_release_history where f_agent_id = \? and f_agent_version = \?`).
		WithArgs("agent-1", "v1").
		WillReturnError(errors.New("query failed"))

	_, err := repo.GetByAgentIdVersion(context.Background(), "agent-1", "v1")
	assert.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- releaseHistoryRepo.GetLatestVersionByAgentID ---

func TestReleaseHistoryRepo_GetLatestVersionByAgentID_Happy(t *testing.T) {
	t.Parallel()

	repo, db, mock := newReleaseHistoryRepoWithMock(t)
	defer db.Close()

	rows := sqlmock.NewRows([]string{
		"f_id", "f_agent_id", "f_agent_config", "f_agent_version", "f_agent_desc",
		"f_create_time", "f_update_time", "f_create_by", "f_update_by",
	}).
		AddRow("hist-1", "agent-1", `{}`, "v2", "desc", int64(1), int64(1), "u1", "u1").
		AddRow("hist-2", "agent-1", `{}`, "v1", "desc", int64(1), int64(1), "u1", "u1")

	mock.ExpectQuery(`select .* from t_data_agent_release_history where f_agent_id = \? order by f_create_time DESC`).
		WithArgs("agent-1").
		WillReturnRows(rows)

	rt, err := repo.GetLatestVersionByAgentID(context.Background(), "agent-1")
	require.NoError(t, err)
	assert.Equal(t, "v2", rt.AgentVersion)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReleaseHistoryRepo_GetLatestVersionByAgentID_Empty(t *testing.T) {
	t.Parallel()

	repo, db, mock := newReleaseHistoryRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`select .* from t_data_agent_release_history where f_agent_id = \? order by f_create_time DESC`).
		WithArgs("agent-1").
		WillReturnRows(sqlmock.NewRows([]string{"f_id", "f_agent_id", "f_agent_config", "f_agent_version", "f_agent_desc", "f_create_time", "f_update_time", "f_create_by", "f_update_by"}))

	rt, err := repo.GetLatestVersionByAgentID(context.Background(), "agent-1")
	require.NoError(t, err)
	assert.Nil(t, rt)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReleaseHistoryRepo_GetLatestVersionByAgentID_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newReleaseHistoryRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`select .* from t_data_agent_release_history where f_agent_id = \? order by f_create_time DESC`).
		WithArgs("agent-1").
		WillReturnError(errors.New("query failed"))

	_, err := repo.GetLatestVersionByAgentID(context.Background(), "agent-1")
	assert.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- releasePermissionRepo Singleton ---

func TestNewReleasePermissionRepo_Singleton(t *testing.T) {
	t.Parallel()

	oldOnce := releasePermissionRepoOnce //nolint:govet
	oldImpl := releasePermissionRepoImpl
	oldGDB := global.GDB

	t.Cleanup(func() {
		releasePermissionRepoOnce = oldOnce //nolint:govet
		releasePermissionRepoImpl = oldImpl
		global.GDB = oldGDB
	})

	db, _, err := sqlx.New()
	require.NoError(t, err)

	global.GDB = db
	releasePermissionRepoOnce = sync.Once{}
	releasePermissionRepoImpl = nil

	r1 := NewReleasePermissionRepo()
	r2 := NewReleasePermissionRepo()

	assert.NotNil(t, r1)
	assert.Same(t, r1, r2)
}

// --- releasePermissionRepo.Create ---

func TestReleasePermissionRepo_Create_Happy(t *testing.T) {
	t.Parallel()

	repo, db, mock := newReleasePermissionRepoWithMock(t)
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta("insert into t_data_agent_release_permission")).
		WillReturnResult(sqlmock.NewResult(1, 1))

	po := &dapo.ReleasePermissionPO{ReleaseId: "rel-1", ObjectId: "obj-1"}
	err := repo.Create(context.Background(), nil, po)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReleasePermissionRepo_Create_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newReleasePermissionRepoWithMock(t)
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta("insert into t_data_agent_release_permission")).
		WillReturnError(errors.New("insert failed"))

	po := &dapo.ReleasePermissionPO{ReleaseId: "rel-1", ObjectId: "obj-1"}
	err := repo.Create(context.Background(), nil, po)
	assert.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReleasePermissionRepo_BatchCreate_Empty(t *testing.T) {
	t.Parallel()

	repo, db, _ := newReleasePermissionRepoWithMock(t)
	defer db.Close()

	err := repo.BatchCreate(context.Background(), nil, []*dapo.ReleasePermissionPO{})
	require.NoError(t, err)
}

func TestReleasePermissionRepo_BatchCreate_Happy(t *testing.T) {
	t.Parallel()

	repo, db, mock := newReleasePermissionRepoWithMock(t)
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta("insert into t_data_agent_release_permission")).
		WillReturnResult(sqlmock.NewResult(2, 2))

	pos := []*dapo.ReleasePermissionPO{
		{ReleaseId: "rel-1", ObjectId: "obj-1"},
		{ReleaseId: "rel-1", ObjectId: "obj-2"},
	}
	err := repo.BatchCreate(context.Background(), nil, pos)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- releaseCategoryRelRepo Singleton ---

func TestNewReleaseCategoryRelRepo_Singleton(t *testing.T) {
	t.Parallel()

	oldOnce := releaseCategoryRelRepoOnce //nolint:govet
	oldImpl := releaseCategoryRelRepoImpl
	oldGDB := global.GDB

	t.Cleanup(func() {
		releaseCategoryRelRepoOnce = oldOnce //nolint:govet
		releaseCategoryRelRepoImpl = oldImpl
		global.GDB = oldGDB
	})

	db, _, err := sqlx.New()
	require.NoError(t, err)

	global.GDB = db
	releaseCategoryRelRepoOnce = sync.Once{}
	releaseCategoryRelRepoImpl = nil

	r1 := NewReleaseCategoryRelRepo()
	r2 := NewReleaseCategoryRelRepo()

	assert.NotNil(t, r1)
	assert.Same(t, r1, r2)
}

// --- releaseCategoryRelRepo.Create ---

func TestReleaseCategoryRelRepo_Create_Happy(t *testing.T) {
	t.Parallel()

	repo, db, mock := newReleaseCategoryRelRepoWithMock(t)
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta("insert into t_data_agent_release_category_rel")).
		WillReturnResult(sqlmock.NewResult(1, 1))

	po := &dapo.ReleaseCategoryRelPO{ReleaseID: "rel-1", CategoryID: "cat-1"}
	err := repo.Create(context.Background(), nil, po)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReleaseCategoryRelRepo_Create_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newReleaseCategoryRelRepoWithMock(t)
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta("insert into t_data_agent_release_category_rel")).
		WillReturnError(errors.New("insert failed"))

	po := &dapo.ReleaseCategoryRelPO{ReleaseID: "rel-1", CategoryID: "cat-1"}
	err := repo.Create(context.Background(), nil, po)
	assert.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReleaseCategoryRelRepo_BatchCreate_Empty(t *testing.T) {
	t.Parallel()

	repo, db, _ := newReleaseCategoryRelRepoWithMock(t)
	defer db.Close()

	err := repo.BatchCreate(context.Background(), nil, []*dapo.ReleaseCategoryRelPO{})
	require.NoError(t, err)
}

func TestReleaseCategoryRelRepo_BatchCreate_Happy(t *testing.T) {
	t.Parallel()

	repo, db, mock := newReleaseCategoryRelRepoWithMock(t)
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta("insert into t_data_agent_release_category_rel")).
		WillReturnResult(sqlmock.NewResult(2, 2))

	pos := []*dapo.ReleaseCategoryRelPO{
		{ReleaseID: "rel-1", CategoryID: "cat-1"},
		{ReleaseID: "rel-1", CategoryID: "cat-2"},
	}
	err := repo.BatchCreate(context.Background(), nil, pos)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

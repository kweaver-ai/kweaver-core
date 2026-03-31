package categoryacc

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/global"
	"github.com/kweaver-ai/proton-rds-sdk-go/sqlx"
)

type catTestLogger struct{}

func (catTestLogger) Infof(string, ...interface{})  {}
func (catTestLogger) Infoln(...interface{})         {}
func (catTestLogger) Debugf(string, ...interface{}) {}
func (catTestLogger) Debugln(...interface{})        {}
func (catTestLogger) Errorf(string, ...interface{}) {}
func (catTestLogger) Errorln(...interface{})        {}
func (catTestLogger) Warnf(string, ...interface{})  {}
func (catTestLogger) Warnln(...interface{})         {}
func (catTestLogger) Panicf(string, ...interface{}) {}
func (catTestLogger) Panicln(...interface{})        {}
func (catTestLogger) Fatalf(string, ...interface{}) {}
func (catTestLogger) Fatalln(...interface{})        {}

func newCatRepoWithMock(t *testing.T) (*categoryRepo, *sqlx.DB, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlx.New()
	require.NoError(t, err)

	repo := &categoryRepo{
		db:       db,
		logger:   catTestLogger{},
		RepoBase: drivenadapter.NewRepoBase(),
	}

	return repo, db, mock
}

// ==================== NewCategoryRepo ====================

func TestNewCategoryRepo_Singleton(t *testing.T) {
	oldOnce := categoryRepoOnce //nolint:govet
	oldImpl := categoryRepoImpl
	oldGDB := global.GDB

	t.Cleanup(func() {
		categoryRepoOnce = oldOnce //nolint:govet
		categoryRepoImpl = oldImpl
		global.GDB = oldGDB
	})

	db, _, err := sqlx.New()
	require.NoError(t, err)

	global.GDB = db
	categoryRepoOnce = sync.Once{}
	categoryRepoImpl = nil

	r1 := NewCategoryRepo()
	r2 := NewCategoryRepo()

	assert.NotNil(t, r1)
	assert.Same(t, r1, r2)
}

// ==================== GetByReleaseId ====================

func TestGetByReleaseId_ReturnsNil(t *testing.T) {
	t.Parallel()

	repo, db, _ := newCatRepoWithMock(t)
	defer db.Close()

	rt, err := repo.GetByReleaseId(context.Background(), "release-1")
	assert.NoError(t, err)
	assert.Nil(t, rt)
}

// ==================== List ====================

func TestList_Happy(t *testing.T) {
	t.Parallel()

	repo, db, mock := newCatRepoWithMock(t)
	defer db.Close()

	rows := sqlmock.NewRows([]string{
		"f_id", "f_name", "f_description",
		"f_create_time", "f_update_time", "f_create_by", "f_update_by",
	}).AddRow("cat-1", "分类1", "描述", int64(1), int64(1), "u1", "u1")

	mock.ExpectQuery(`(?i)SELECT .* FROM t_data_agent_release_category`).
		WillReturnRows(rows)

	rt, err := repo.List(context.Background(), nil)
	require.NoError(t, err)
	assert.Len(t, rt, 1)
	assert.Equal(t, "cat-1", rt[0].ID)
	assert.Equal(t, "分类1", rt[0].Name)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestList_QueryError(t *testing.T) {
	t.Parallel()

	repo, db, mock := newCatRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)SELECT .* FROM t_data_agent_release_category`).
		WillReturnError(errors.New("db error"))

	_, err := repo.List(context.Background(), nil)
	assert.Error(t, err)
}

func TestList_EmptyResult(t *testing.T) {
	t.Parallel()

	repo, db, mock := newCatRepoWithMock(t)
	defer db.Close()

	rows := sqlmock.NewRows([]string{
		"f_id", "f_name", "f_description",
		"f_create_time", "f_update_time", "f_create_by", "f_update_by",
	})

	mock.ExpectQuery(`(?i)SELECT .* FROM t_data_agent_release_category`).
		WillReturnRows(rows)

	rt, err := repo.List(context.Background(), nil)
	require.NoError(t, err)
	assert.Empty(t, rt)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ==================== GetIDNameMap ====================

func TestGetIDNameMap_Happy(t *testing.T) {
	t.Parallel()

	repo, db, mock := newCatRepoWithMock(t)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"f_id", "f_name"}).
		AddRow("cat-1", "分类1").
		AddRow("cat-2", "分类2")

	mock.ExpectQuery(`(?i)SELECT .* FROM t_data_agent_release_category`).
		WillReturnRows(rows)

	m, err := repo.GetIDNameMap(context.Background(), []string{"cat-1", "cat-2"})
	require.NoError(t, err)
	assert.Len(t, m, 2)
	assert.Equal(t, "分类1", m["cat-1"])
	assert.Equal(t, "分类2", m["cat-2"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetIDNameMap_QueryError(t *testing.T) {
	t.Parallel()

	repo, db, mock := newCatRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)SELECT .* FROM t_data_agent_release_category`).
		WillReturnError(errors.New("db error"))

	_, err := repo.GetIDNameMap(context.Background(), []string{"cat-1"})
	assert.Error(t, err)
}

func TestGetIDNameMap_Empty(t *testing.T) {
	t.Parallel()

	repo, db, mock := newCatRepoWithMock(t)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"f_id", "f_name"})

	mock.ExpectQuery(`(?i)SELECT .* FROM t_data_agent_release_category`).
		WillReturnRows(rows)

	m, err := repo.GetIDNameMap(context.Background(), []string{"cat-1"})
	require.NoError(t, err)
	assert.Len(t, m, 0)
	require.NoError(t, mock.ExpectationsWereMet())
}

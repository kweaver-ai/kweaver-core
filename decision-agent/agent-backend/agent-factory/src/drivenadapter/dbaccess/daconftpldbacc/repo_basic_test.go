package daconftpldbacc

import (
	"context"
	"errors"
	"regexp"
	"sync"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/dbaccess"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/global"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/kweaver-ai/proton-rds-sdk-go/sqlx"
)

type tplTestLogger struct{}

func (tplTestLogger) Infof(string, ...interface{})  {}
func (tplTestLogger) Infoln(...interface{})         {}
func (tplTestLogger) Debugf(string, ...interface{}) {}
func (tplTestLogger) Debugln(...interface{})        {}
func (tplTestLogger) Errorf(string, ...interface{}) {}
func (tplTestLogger) Errorln(...interface{})        {}
func (tplTestLogger) Warnf(string, ...interface{})  {}
func (tplTestLogger) Warnln(...interface{})         {}
func (tplTestLogger) Panicf(string, ...interface{}) {}
func (tplTestLogger) Panicln(...interface{})        {}
func (tplTestLogger) Fatalf(string, ...interface{}) {}
func (tplTestLogger) Fatalln(...interface{})        {}

func newTplRepoWithMock(t *testing.T) (*DAConfigTplRepo, *sqlx.DB, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlx.New()
	require.NoError(t, err)

	repo := &DAConfigTplRepo{
		db:             db,
		logger:         tplTestLogger{},
		IDBAccBaseRepo: dbaccess.NewDBAccBase(),
	}

	return repo, db, mock
}

func tplColumns() []string {
	return []string{
		"f_id",
		"f_name",
		"f_key",
		"f_product_key",
		"f_profile",
		"f_avatar_type",
		"f_avatar",
		"f_status",
		"f_is_built_in",
		"f_created_at",
		"f_updated_at",
		"f_created_by",
		"f_updated_by",
		"f_deleted_at",
		"f_deleted_by",
		"f_config",
		"f_created_type",
		"f_published_at",
		"f_published_by",
		"f_create_from",
	}
}

func tplRows() *sqlmock.Rows {
	return sqlmock.NewRows(tplColumns()).AddRow(
		int64(1),
		"tpl-1",
		"tpl-key-1",
		"pk",
		nil,
		0,
		"",
		string(cdaenum.StatusUnpublished),
		nil,
		int64(1),
		int64(1),
		"u1",
		"u1",
		int64(0),
		"",
		"{}",
		"",
		nil,
		nil,
		"",
	)
}

func tplUserCtx(uid string) context.Context {
	visitor := &rest.Visitor{ID: uid}
	return context.WithValue(context.Background(), cenum.VisitUserInfoCtxKey.String(), visitor) //nolint:staticcheck // SA1029
}

func TestNewDataAgentTplRepo_Singleton(t *testing.T) {
	t.Parallel()

	oldOnce := agentTplRepoOnce //nolint:govet
	oldImpl := agentTplRepoImpl
	oldGDB := global.GDB

	t.Cleanup(func() {
		agentTplRepoOnce = oldOnce //nolint:govet
		agentTplRepoImpl = oldImpl
		global.GDB = oldGDB
	})

	db, _, err := sqlx.New()
	require.NoError(t, err)

	global.GDB = db
	agentTplRepoOnce = sync.Once{}
	agentTplRepoImpl = nil

	r1 := NewDataAgentTplRepo()
	r2 := NewDataAgentTplRepo()

	assert.NotNil(t, r1)
	assert.Same(t, r1, r2)
}

func TestDAConfigTplRepo_CreateAndUpdate(t *testing.T) {
	t.Parallel()

	repo, db, mock := newTplRepoWithMock(t)
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta("insert into t_data_agent_config_tpl")).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`update t_data_agent_config_tpl set .* where f_id = \?`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	po := &dapo.DataAgentTplPo{ID: 1, Name: "tpl-1", Key: "k1"}
	err := repo.Create(context.Background(), nil, po)
	require.NoError(t, err)
	err = repo.Update(context.Background(), nil, po)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDAConfigTplRepo_ExistsMethods(t *testing.T) {
	t.Parallel()

	t.Run("exists true", func(t *testing.T) {
		t.Parallel()

		repo, db, mock := newTplRepoWithMock(t)
		defer db.Close()

		mock.ExpectQuery(regexp.QuoteMeta("select count(*) from t_data_agent_config_tpl where f_deleted_at = ? and f_name = ?")).
			WithArgs(0, "n1").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
		mock.ExpectQuery(regexp.QuoteMeta("select count(*) from t_data_agent_config_tpl where f_deleted_at = ? and f_key = ?")).
			WithArgs(0, "k1").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
		mock.ExpectQuery(regexp.QuoteMeta("select count(*) from t_data_agent_config_tpl where f_deleted_at = ? and f_id = ?")).
			WithArgs(0, int64(1)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		e1, err := repo.ExistsByName(context.Background(), "n1")
		require.NoError(t, err)
		assert.True(t, e1)

		e2, err := repo.ExistsByKey(context.Background(), "k1")
		require.NoError(t, err)
		assert.True(t, e2)

		e3, err := repo.ExistsByID(context.Background(), 1)
		require.NoError(t, err)
		assert.True(t, e3)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("exclude id branches", func(t *testing.T) {
		t.Parallel()

		repo, db, mock := newTplRepoWithMock(t)
		defer db.Close()

		mock.ExpectQuery(regexp.QuoteMeta("select count(*) from t_data_agent_config_tpl where f_deleted_at = ? and f_name = ? and f_id <> ?")).
			WithArgs(0, "n1", int64(1)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
		mock.ExpectQuery(regexp.QuoteMeta("select count(*) from t_data_agent_config_tpl where f_deleted_at = ? and f_key = ? and f_id <> ?")).
			WithArgs(0, "k1", int64(1)).
			WillReturnError(errors.New("count failed"))

		e1, err := repo.ExistsByNameExcludeID(context.Background(), "n1", 1)
		require.NoError(t, err)
		assert.False(t, e1)

		_, err = repo.ExistsByKeyExcludeID(context.Background(), "k1", 1)
		assert.Error(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDAConfigTplRepo_GetAllIDsAndGetByMethods(t *testing.T) {
	t.Parallel()

	repo, db, mock := newTplRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta("select f_id from t_data_agent_config_tpl where f_deleted_at = ?")).
		WithArgs(0).
		WillReturnRows(sqlmock.NewRows([]string{"f_id"}).AddRow(int64(1)).AddRow(int64(2)))
	mock.ExpectQuery(`select .* from t_data_agent_config_tpl where f_deleted_at = \? and f_id = \?`).
		WithArgs(0, int64(1)).
		WillReturnRows(tplRows())
	mock.ExpectQuery(`select .* from t_data_agent_config_tpl where f_deleted_at = \? and f_key = \?`).
		WithArgs(0, "k1").
		WillReturnRows(tplRows())
	mock.ExpectQuery(`select .* from t_data_agent_config_tpl where f_deleted_at = \? and f_key in \(\?,\?\)`).
		WithArgs(0, "k1", "k2").
		WillReturnRows(tplRows())
	mock.ExpectQuery(`select .* from t_data_agent_config_tpl where f_deleted_at = \? and f_id in \(\?,\?\)`).
		WithArgs(0, int64(1), int64(2)).
		WillReturnRows(tplRows())

	ids, err := repo.GetAllIDs(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []int64{1, 2}, ids)

	po1, err := repo.GetByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, int64(1), po1.ID)

	po2, err := repo.GetByKey(context.Background(), "k1")
	require.NoError(t, err)
	assert.Equal(t, int64(1), po2.ID)

	posByKey, err := repo.GetByKeys(context.Background(), []string{"k1", "k2"})
	require.NoError(t, err)
	assert.Len(t, posByKey, 1)

	posByID, err := repo.GetByIDS(context.Background(), []int64{1, 2})
	require.NoError(t, err)
	assert.Len(t, posByID, 1)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDAConfigTplRepo_GetByWithTxAndCategoryAndMap(t *testing.T) {
	t.Parallel()

	t.Run("tx and category success", func(t *testing.T) {
		t.Parallel()

		repo, db, mock := newTplRepoWithMock(t)
		defer db.Close()

		mock.ExpectBegin()

		tx, err := db.Begin()
		require.NoError(t, err)

		mock.ExpectQuery(`select .* from t_data_agent_config_tpl where f_deleted_at = \? and f_id = \?`).
			WithArgs(0, int64(1)).
			WillReturnRows(tplRows())
		mock.ExpectQuery(`select .* from t_data_agent_config_tpl where f_deleted_at = \? and f_key = \?`).
			WithArgs(0, "k1").
			WillReturnRows(tplRows())
		mock.ExpectQuery(`select .* from t_data_agent_config_tpl where f_deleted_at = \? and f_category_id = \?`).
			WithArgs(0, "c1").
			WillReturnRows(tplRows())
		mock.ExpectQuery(`select .* from t_data_agent_config_tpl where f_deleted_at = \? and f_id in \(\?,\?\)`).
			WithArgs(0, int64(1), int64(2)).
			WillReturnRows(tplRows())
		mock.ExpectRollback()

		_, err = repo.GetByIDWithTx(context.Background(), tx, 1)
		require.NoError(t, err)
		_, err = repo.GetByKeyWithTx(context.Background(), tx, "k1")
		require.NoError(t, err)
		pos, err := repo.GetByCategoryID(context.Background(), "c1")
		require.NoError(t, err)
		assert.Len(t, pos, 1)

		m, err := repo.GetMapByIDs(context.Background(), []int64{1, 2})
		require.NoError(t, err)
		assert.Len(t, m, 1)
		require.NoError(t, tx.Rollback())
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("map empty and tx error branch", func(t *testing.T) {
		t.Parallel()

		repo, db, mock := newTplRepoWithMock(t)
		defer db.Close()

		m, err := repo.GetMapByIDs(context.Background(), nil)
		require.NoError(t, err)
		assert.Empty(t, m)

		mock.ExpectBegin()

		tx, err := db.Begin()
		require.NoError(t, err)
		mock.ExpectQuery(`select .* from t_data_agent_config_tpl where f_deleted_at = \? and f_id = \?`).
			WithArgs(0, int64(1)).
			WillReturnError(errors.New("query failed"))
		mock.ExpectRollback()

		_, err = repo.GetByIDWithTx(context.Background(), tx, 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "[GetByIDWithTx]")
		require.NoError(t, tx.Rollback())
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDAConfigTplRepo_UpdateStatusAndDelete(t *testing.T) {
	t.Parallel()

	t.Run("update status branches", func(t *testing.T) {
		t.Parallel()

		repo, db, mock := newTplRepoWithMock(t)
		defer db.Close()

		mock.ExpectExec(`update t_data_agent_config_tpl set .* where f_id = \?`).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec(`update t_data_agent_config_tpl set .* where f_id = \?`).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.UpdateStatus(context.Background(), nil, cdaenum.StatusPublished, 1, "u1", 100)
		require.NoError(t, err)
		err = repo.UpdateStatus(context.Background(), nil, cdaenum.StatusUnpublished, 1, "u1", 100)
		require.NoError(t, err)

		err = repo.UpdateStatus(context.Background(), nil, cdaenum.Status("x"), 1, "u1", 100)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid status")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("delete uid empty and success", func(t *testing.T) {
		t.Parallel()

		repo, db, mock := newTplRepoWithMock(t)
		defer db.Close()

		err := repo.Delete(context.Background(), nil, 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "uid is empty")

		mock.ExpectExec(`update t_data_agent_config_tpl set .* where f_id = \?`).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err = repo.Delete(tplUserCtx("u1"), nil, 1)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

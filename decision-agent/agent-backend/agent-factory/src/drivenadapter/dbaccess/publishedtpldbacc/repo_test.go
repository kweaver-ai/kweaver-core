package publishedtpldbacc

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
	"github.com/kweaver-ai/proton-rds-sdk-go/sqlx"
)

type pubedTplTestLogger struct{}

func (pubedTplTestLogger) Infof(string, ...interface{})  {}
func (pubedTplTestLogger) Infoln(...interface{})         {}
func (pubedTplTestLogger) Debugf(string, ...interface{}) {}
func (pubedTplTestLogger) Debugln(...interface{})        {}
func (pubedTplTestLogger) Errorf(string, ...interface{}) {}
func (pubedTplTestLogger) Errorln(...interface{})        {}
func (pubedTplTestLogger) Warnf(string, ...interface{})  {}
func (pubedTplTestLogger) Warnln(...interface{})         {}
func (pubedTplTestLogger) Panicf(string, ...interface{}) {}
func (pubedTplTestLogger) Panicln(...interface{})        {}
func (pubedTplTestLogger) Fatalf(string, ...interface{}) {}
func (pubedTplTestLogger) Fatalln(...interface{})        {}

func newPubedTplRepoWithMock(t *testing.T) (*PubedTplRepo, *sqlx.DB, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlx.New()
	require.NoError(t, err)

	repo := &PubedTplRepo{
		db:             db,
		logger:         pubedTplTestLogger{},
		IDBAccBaseRepo: dbaccess.NewDBAccBase(),
	}

	return repo, db, mock
}

func mockPubedTplRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"f_id", "f_name", "f_key", "f_product_key", "f_profile",
		"f_avatar_type", "f_avatar", "f_is_built_in", "f_config",
		"f_published_at", "f_published_by", "f_tpl_id",
	}).AddRow(
		int64(1), "tpl-name", "tpl-key", "prod-key", nil,
		int(1), "avatar.png", nil, `{}`,
		int64(1), "u1", int64(100),
	)
}

func TestNewPublishedTplRepo_Singleton(t *testing.T) {
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

	r1 := NewPublishedTplRepo()
	r2 := NewPublishedTplRepo()

	assert.NotNil(t, r1)
	assert.Same(t, r1, r2)
}

func TestPubedTplRepo_GetByID_Happy(t *testing.T) {
	t.Parallel()

	repo, db, mock := newPubedTplRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`select .* from t_data_agent_config_tpl_published where f_id = \?`).
		WithArgs(int64(1)).
		WillReturnRows(mockPubedTplRows())

	po, err := repo.GetByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, int64(1), po.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPubedTplRepo_GetByID_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newPubedTplRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`select .* from t_data_agent_config_tpl_published where f_id = \?`).
		WithArgs(int64(1)).
		WillReturnError(errors.New("query failed"))

	_, err := repo.GetByID(context.Background(), 1)
	assert.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPubedTplRepo_GetByKey_Happy(t *testing.T) {
	t.Parallel()

	repo, db, mock := newPubedTplRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`select .* from t_data_agent_config_tpl_published where f_key = \?`).
		WithArgs("tpl-key").
		WillReturnRows(mockPubedTplRows())

	po, err := repo.GetByKey(context.Background(), "tpl-key")
	require.NoError(t, err)
	assert.Equal(t, "tpl-key", po.Key)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPubedTplRepo_GetByKey_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newPubedTplRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`select .* from t_data_agent_config_tpl_published where f_key = \?`).
		WithArgs("tpl-key").
		WillReturnError(errors.New("query failed"))

	_, err := repo.GetByKey(context.Background(), "tpl-key")
	assert.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPubedTplRepo_GetByTplID_Happy(t *testing.T) {
	t.Parallel()

	repo, db, mock := newPubedTplRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`select .* from t_data_agent_config_tpl_published where f_tpl_id = \?`).
		WithArgs(int64(100)).
		WillReturnRows(mockPubedTplRows())

	po, err := repo.GetByTplID(context.Background(), 100)
	require.NoError(t, err)
	assert.Equal(t, int64(100), po.TplID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPubedTplRepo_Delete_Happy(t *testing.T) {
	t.Parallel()

	repo, db, mock := newPubedTplRepoWithMock(t)
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta("delete from t_data_agent_config_tpl_published where f_id = ?")).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete(context.Background(), nil, 1)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPubedTplRepo_Delete_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newPubedTplRepoWithMock(t)
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta("delete from t_data_agent_config_tpl_published where f_id = ?")).
		WillReturnError(errors.New("delete failed"))

	err := repo.Delete(context.Background(), nil, 1)
	assert.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPubedTplRepo_DeleteByTplID_Happy(t *testing.T) {
	t.Parallel()

	repo, db, mock := newPubedTplRepoWithMock(t)
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta("delete from t_data_agent_config_tpl_published where f_tpl_id = ?")).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.DeleteByTplID(context.Background(), nil, 100)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPubedTplRepo_ExistsByKey_True(t *testing.T) {
	t.Parallel()

	repo, db, mock := newPubedTplRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta("select count(*) from t_data_agent_config_tpl_published where f_key = ?")).
		WithArgs("tpl-key").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	exists, err := repo.ExistsByKey(context.Background(), "tpl-key")
	require.NoError(t, err)
	assert.True(t, exists)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPubedTplRepo_ExistsByKey_False(t *testing.T) {
	t.Parallel()

	repo, db, mock := newPubedTplRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta("select count(*) from t_data_agent_config_tpl_published where f_key = ?")).
		WithArgs("tpl-key").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	exists, err := repo.ExistsByKey(context.Background(), "tpl-key")
	require.NoError(t, err)
	assert.False(t, exists)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPubedTplRepo_ExistsByKey_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newPubedTplRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta("select count(*) from t_data_agent_config_tpl_published where f_key = ?")).
		WithArgs("tpl-key").
		WillReturnError(errors.New("query failed"))

	_, err := repo.ExistsByKey(context.Background(), "tpl-key")
	assert.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPubedTplRepo_ExistsByID_True(t *testing.T) {
	t.Parallel()

	repo, db, mock := newPubedTplRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta("select count(*) from t_data_agent_config_tpl_published where f_id = ?")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	exists, err := repo.ExistsByID(context.Background(), 1)
	require.NoError(t, err)
	assert.True(t, exists)
	require.NoError(t, mock.ExpectationsWereMet())
}
